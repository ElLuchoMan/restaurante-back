package login

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"restaurante/internal/authguard"
	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
)

// resetGuards deja limitadores y contador por documento limpios durante el test.
func resetGuards(t *testing.T) {
	t.Helper()
	oRL, oRR, oDF := loginRL, refreshRL, docFailures
	oMax, oRMax := loginMaxReq, refreshMaxReq
	loginRL, refreshRL = newRateLimiter(), newRateLimiter()
	docFailures = authguard.NewFailures(authguard.Config{
		MaxFailures: 5, Window: 15 * time.Minute, BaseWait: 30 * time.Second,
		MaxWait: 15 * time.Minute, MaxEntries: 100,
	})
	t.Cleanup(func() {
		loginRL, refreshRL, docFailures = oRL, oRR, oDF
		loginMaxReq, refreshMaxReq = oMax, oRMax
	})
}

func doLogin(t *testing.T, ip string, doc int64, pass string) *httptest.ResponseRecorder {
	t.Helper()
	body := `{"documento":` + strconv.FormatInt(doc, 10) + `,"password":"` + pass + `"}`
	r := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
	r.RemoteAddr = ip + ":1000"
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := LoginController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Login()
	return w
}

func mockUser(t *testing.T, existe bool) {
	t.Helper()
	orig, origCmp := newOrm, compareHashAndPassword
	t.Cleanup(func() { newOrm, compareHashAndPassword = orig, origCmp })
	newOrm = func() orm.Ormer {
		return &mockLoginOrmer{ReadFunc: func(v interface{}, cols ...string) error {
			if tr, ok := v.(*models.Trabajador); ok && existe {
				tr.PASSWORD = "hash-ok"
				return nil
			}
			return orm.ErrNoRows
		}}
	}
	compareHashAndPassword = func(h, p []byte) error {
		if string(h) == "hash-ok" && string(p) == "good" {
			return nil
		}
		return errors.New("mismatch")
	}
}

// Documento inexistente y contraseña incorrecta: misma respuesta, y en ambos se
// ejecuta una comparación bcrypt (el hash ficticio es un bcrypt válido).
func TestLoginSinEnumeracionDeUsuarios(t *testing.T) {
	resetGuards(t)
	var hashes []string
	for _, existe := range []bool{true, false} {
		mockUser(t, existe)
		inner := compareHashAndPassword
		compareHashAndPassword = func(h, p []byte) error {
			hashes = append(hashes, string(h))
			return inner(h, p)
		}
		w := doLogin(t, "198.51.100.1", 500, "bad")
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("existe=%v: esperaba 401, got %d", existe, w.Code)
		}
		var resp models.ApiResponse
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.Message != "Credenciales inválidas" || resp.Cause != "" {
			t.Fatalf("existe=%v: respuesta distinta: %+v", existe, resp)
		}
	}
	if len(hashes) != 2 || hashes[0] != "hash-ok" || hashes[1] != dummyPasswordHash {
		t.Fatalf("comparaciones bcrypt inesperadas: %v", hashes)
	}
}

func TestDummyPasswordHashEsBcryptValido(t *testing.T) {
	err := compareHashAndPassword([]byte(dummyPasswordHash), []byte("x"))
	if err == nil || strings.Contains(err.Error(), "hashedSecret too short") || strings.Contains(err.Error(), "version") {
		t.Fatalf("el hash ficticio debe ser un bcrypt válido que no coincide: %v", err)
	}
}

func TestLoginBloqueoPorDocumentoYReinicio(t *testing.T) {
	resetGuards(t)
	loginMaxReq = 1000
	mockUser(t, true)

	for i := 0; i < 5; i++ {
		if w := doLogin(t, "198.51.100.2", 77, "bad"); w.Code != http.StatusUnauthorized {
			t.Fatalf("intento %d: esperaba 401, got %d", i+1, w.Code)
		}
	}
	// 6º intento, aun con la contraseña correcta y desde otra IP: 429 + Retry-After.
	w := doLogin(t, "198.51.100.3", 77, "good")
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("esperaba 429, got %d", w.Code)
	}
	if ra, err := strconv.Atoi(w.Header().Get("Retry-After")); err != nil || ra < 1 || ra > 30 {
		t.Fatalf("Retry-After inválido: %q", w.Header().Get("Retry-After"))
	}
	// otro documento no se ve afectado
	if w := doLogin(t, "198.51.100.3", 78, "bad"); w.Code != http.StatusUnauthorized {
		t.Fatalf("otro documento: esperaba 401, got %d", w.Code)
	}

	// Un login correcto (antes del bloqueo) reinicia el contador.
	for i := 0; i < 4; i++ {
		doLogin(t, "198.51.100.4", 88, "bad")
	}
	if w := doLogin(t, "198.51.100.4", 88, "good"); w.Code != http.StatusOK {
		t.Fatalf("login correcto: esperaba 200, got %d", w.Code)
	}
	for i := 0; i < 4; i++ {
		if w := doLogin(t, "198.51.100.4", 88, "bad"); w.Code != http.StatusUnauthorized {
			t.Fatalf("tras reinicio, intento %d: esperaba 401, got %d", i+1, w.Code)
		}
	}
}

func TestLoginLimitePorIPConRetryAfter(t *testing.T) {
	resetGuards(t)
	loginMaxReq = 1
	mockUser(t, false)
	doLogin(t, "198.51.100.5", 1, "x")
	w := doLogin(t, "198.51.100.5", 1, "x")
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("esperaba 429, got %d", w.Code)
	}
	if ra, err := strconv.Atoi(w.Header().Get("Retry-After")); err != nil || ra < 1 || ra > 60 {
		t.Fatalf("Retry-After inválido: %q", w.Header().Get("Retry-After"))
	}
}

func TestLoginNoSePuedeFalsearIPConXForwardedFor(t *testing.T) {
	resetGuards(t)
	loginMaxReq = 1
	mockUser(t, false)
	send := func(spoof string) int {
		body := `{"documento":1,"password":"x"}`
		r := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
		r.RemoteAddr = "10.0.0.1:1"
		r.Header.Set("X-Forwarded-For", spoof+", 203.0.113.50, 104.16.0.1")
		w := httptest.NewRecorder()
		ctx := context.NewContext()
		ctx.Reset(w, r)
		ctx.Input.RequestBody = []byte(body)
		c := LoginController{}
		c.Ctx = ctx
		c.Data = make(map[interface{}]interface{})
		c.Login()
		return w.Code
	}
	send("1.1.1.1")
	if code := send("2.2.2.2"); code != http.StatusTooManyRequests {
		t.Fatalf("rotar el primer valor de XFF no debe evadir el límite: %d", code)
	}
}

func doRefresh(ip string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	r.RemoteAddr = ip + ":1000"
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := LoginController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.RefreshToken()
	return w
}

func TestRefreshLimitePorIP(t *testing.T) {
	resetGuards(t)
	refreshMaxReq = 2
	for i := 0; i < 2; i++ {
		if w := doRefresh("198.51.100.6"); w.Code != http.StatusBadRequest {
			t.Fatalf("intento %d: esperaba 400 (sin token), got %d", i+1, w.Code)
		}
	}
	w := doRefresh("198.51.100.6")
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("esperaba 429, got %d", w.Code)
	}
	if ra, err := strconv.Atoi(w.Header().Get("Retry-After")); err != nil || ra < 1 {
		t.Fatalf("Retry-After inválido: %q", w.Header().Get("Retry-After"))
	}
	if w := doRefresh("198.51.100.7"); w.Code != http.StatusBadRequest {
		t.Fatalf("otra IP no debe verse afectada: %d", w.Code)
	}
}
