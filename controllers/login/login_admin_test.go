package login

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
	"github.com/golang-jwt/jwt/v5"
)

func accessToken(t *testing.T, rol string) string {
	t.Helper()
	acc, _, err := generateTokens(7, rol, "Nombre Prueba")
	if err != nil {
		t.Fatalf("generateTokens: %v", err)
	}
	return acc
}

func refreshTokenStr(t *testing.T) string {
	t.Helper()
	_, ref, err := generateTokens(7, "Administrador", "Nombre Prueba")
	if err != nil {
		t.Fatalf("generateTokens: %v", err)
	}
	return ref
}

func runFilter(f func(*context.Context), method, path, auth string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(method, path, nil)
	if auth != "" {
		r.Header.Set("Authorization", auth)
	}
	ctx := context.NewContext()
	ctx.Reset(w, r)
	f(ctx)
	return w
}

func decodeResp(t *testing.T, w *httptest.ResponseRecorder) models.ApiResponse {
	t.Helper()
	var resp models.ApiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("respuesta no es JSON: %v (%s)", err, w.Body.String())
	}
	return resp
}

func TestValidateAdmin(t *testing.T) {
	const path = "/restaurante/v1/trabajadores"
	cases := []struct {
		name   string
		method string
		auth   string
		want   int
	}{
		{"sin token", http.MethodGet, "", http.StatusUnauthorized},
		{"token invalido", http.MethodPost, "Bearer basura", http.StatusUnauthorized},
		{"refresh token no sirve", http.MethodPut, "Bearer " + refreshTokenStr(t), http.StatusUnauthorized},
		{"rol mesero", http.MethodDelete, "Bearer " + accessToken(t, "Mesero"), http.StatusForbidden},
		{"rol cliente", http.MethodGet, "Bearer " + accessToken(t, "Cliente"), http.StatusForbidden},
		{"admin con Bearer", http.MethodPost, "Bearer " + accessToken(t, "Administrador"), http.StatusOK},
		{"admin sin prefijo Bearer", http.MethodGet, accessToken(t, "Administrador"), http.StatusOK},
		{"OPTIONS pasa", http.MethodOptions, "", http.StatusOK},
	}
	for _, c := range cases {
		w := runFilter(ValidateAdmin, c.method, path, c.auth)
		if w.Code != c.want {
			t.Errorf("%s: esperaba %d, obtuve %d (%s)", c.name, c.want, w.Code, w.Body.String())
		}
		if c.want >= 400 {
			if resp := decodeResp(t, w); resp.Code != c.want {
				t.Errorf("%s: code del cuerpo %d != status %d", c.name, resp.Code, c.want)
			}
		}
	}
}

func TestValidateAdminNoSwaggerBypassInDev(t *testing.T) {
	orig := web.BConfig.RunMode
	web.BConfig.RunMode = "dev"
	t.Cleanup(func() { web.BConfig.RunMode = orig })

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/restaurante/v1/trabajadores", nil)
	r.Header.Set("Referer", "http://localhost/swagger/index.html")
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ValidateAdmin(ctx)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("el referer de swagger no debe saltar la validación de admin: %d", w.Code)
	}
}

func TestValidateTokenRejectsRefreshToken(t *testing.T) {
	w := runFilter(ValidateToken, http.MethodGet, "/restaurante/v1/pedidos", "Bearer "+refreshTokenStr(t))
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("un refresh token no debe valer como access token: %d", w.Code)
	}
	if _, err := ParseTokenClaims(refreshTokenStr(t)); err == nil {
		t.Fatal("ParseTokenClaims debe rechazar refresh tokens")
	}
}

func TestValidateTokenAcceptsAccessToken(t *testing.T) {
	w := runFilter(ValidateToken, http.MethodGet, "/restaurante/v1/pedidos", "Bearer "+accessToken(t, "Mesero"))
	if w.Code != http.StatusOK {
		t.Fatalf("esperaba 200, obtuve %d", w.Code)
	}
}

func TestValidateTokenOptionsPasses(t *testing.T) {
	w := runFilter(ValidateToken, http.MethodOptions, "/restaurante/v1/pedidos", "")
	if w.Code != http.StatusOK {
		t.Fatalf("OPTIONS debe pasar: %d", w.Code)
	}
}

func TestPublicRoutes(t *testing.T) {
	cases := []struct {
		method, path string
		public       bool
	}{
		{http.MethodGet, "/restaurante/v1/productos", true},
		{http.MethodGet, "/restaurante/v1/cambios_horario/actual", true},
		{http.MethodGet, "/restaurante/v1/ofertas/activas", true},
		{http.MethodPost, "/restaurante/v1/clientes", true},
		{http.MethodPost, "/restaurante/v1/reservas", true},
		{http.MethodGet, "/restaurante/v1/reservas/consulta", true},
		// los listados y búsquedas de reservas/contactos con datos personales ya no son públicos
		{http.MethodGet, "/restaurante/v1/reservas", false},
		{http.MethodGet, "/restaurante/v1/reservas/search", false},
		{http.MethodGet, "/restaurante/v1/reservas/parameter", false},
		{http.MethodGet, "/restaurante/v1/reservas/cliente", false},
		{http.MethodGet, "/restaurante/v1/reservas/documento", false},
		{http.MethodGet, "/restaurante/v1/reserva_contacto", false},
		{http.MethodGet, "/restaurante/v1/reserva_contacto/search", false},
		// POST/PUT/DELETE de productos u otras rutas ya no son públicas
		{http.MethodPost, "/restaurante/v1/productos", false},
		{http.MethodPost, "/restaurante/v1/restaurantes", false},
		{http.MethodPut, "/restaurante/v1/productos", false},
		{http.MethodDelete, "/restaurante/v1/reservas", false},
		{http.MethodGet, "/restaurante/v1/clientes", false},
		{http.MethodGet, "/restaurante/v1/trabajadores", false},
	}
	for _, c := range cases {
		if got := isPublicRoute(c.method, c.path); got != c.public {
			t.Errorf("%s %s: público=%v, esperaba %v", c.method, c.path, got, c.public)
		}
	}

	// la query string no debe romper la coincidencia de ruta pública
	w := runFilter(ValidateToken, http.MethodGet, "/restaurante/v1/cambios_horario/actual?x=1", "")
	if w.Code != http.StatusOK {
		t.Fatalf("ruta pública con query debe pasar: %d", w.Code)
	}
}

func TestExpiresInMatchesAccessTokenTTL(t *testing.T) {
	jwtSecret = []byte("testsecret")
	acc, ref, err := generateTokens(1, "Mesero", "N")
	if err != nil {
		t.Fatal(err)
	}
	resp := newAuthResponse(acc, ref, "N")
	if resp.ExpiresIn != "7200" || accessTokenTTL != 120*time.Minute {
		t.Fatalf("expires_in=%q ttl=%v", resp.ExpiresIn, accessTokenTTL)
	}
	claims, err := ParseTokenClaims(acc)
	if err != nil {
		t.Fatal(err)
	}
	ttl := claims.ExpiresAt.Sub(claims.IssuedAt.Time)
	if ttl != accessTokenTTL {
		t.Fatalf("duración real %v != %v", ttl, accessTokenTTL)
	}
	if resp.Token != resp.AccessToken || resp.TokenType != "Bearer" {
		t.Fatalf("respuesta inesperada: %+v", resp)
	}
}

func newLoginCtrl(body string) (*LoginController, *httptest.ResponseRecorder) {
	r := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
	r.RemoteAddr = "203.0.113.77:1"
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := &LoginController{}
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}
	return c, w
}

func withOrm(t *testing.T, read func(v interface{}) error) {
	t.Helper()
	orig := newOrm
	newOrm = func() orm.Ormer {
		return &mockLoginOrmer{ReadFunc: func(v interface{}, _ ...string) error { return read(v) }}
	}
	t.Cleanup(func() { newOrm = orig })
}

func TestLoginMissingFields(t *testing.T) {
	for _, body := range []string{`{}`, `{"documento":5}`, `{"password":"x"}`} {
		c, w := newLoginCtrl(body)
		c.Login()
		if w.Code != http.StatusBadRequest {
			t.Errorf("%s: esperaba 400, obtuve %d", body, w.Code)
		}
		if resp := decodeResp(t, w); resp.Code != http.StatusBadRequest {
			t.Errorf("%s: code %d", body, resp.Code)
		}
	}
}

func TestLoginDBErrors(t *testing.T) {
	dbErr := errors.New("db caída")

	// error al leer trabajador
	withOrm(t, func(v interface{}) error { return dbErr })
	c, w := newLoginCtrl(`{"documento":5,"password":"x"}`)
	c.Login()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("trabajador db error: esperaba 500, obtuve %d", w.Code)
	}

	// trabajador no existe y falla la lectura de cliente
	withOrm(t, func(v interface{}) error {
		if _, ok := v.(*models.Trabajador); ok {
			return orm.ErrNoRows
		}
		return dbErr
	})
	c, w = newLoginCtrl(`{"documento":5,"password":"x"}`)
	c.Login()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("cliente db error: esperaba 500, obtuve %d", w.Code)
	}
	if resp := decodeResp(t, w); resp.Code != 500 || resp.Cause == "" {
		t.Fatalf("respuesta inesperada: %+v", resp)
	}
}

func TestLoginResponseShapeMatchesAuthResponse(t *testing.T) {
	withOrm(t, func(v interface{}) error {
		if tr, ok := v.(*models.Trabajador); ok {
			tr.NOMBRE, tr.APELLIDO, tr.ROL, tr.PASSWORD = "Ana", "Ruiz", models.RolAdministrador, "hash"
			return nil
		}
		return orm.ErrNoRows
	})
	orig := compareHashAndPassword
	compareHashAndPassword = func([]byte, []byte) error { return nil }
	t.Cleanup(func() { compareHashAndPassword = orig })

	c, w := newLoginCtrl(`{"documento":5,"password":"x"}`)
	c.Login()
	if w.Code != http.StatusOK {
		t.Fatalf("esperaba 200, obtuve %d: %s", w.Code, w.Body.String())
	}
	var envelope struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	dec := json.NewDecoder(bytes.NewReader(envelope.Data))
	dec.DisallowUnknownFields()
	var auth models.AuthResponse
	if err := dec.Decode(&auth); err != nil {
		t.Fatalf("la respuesta real no coincide con models.AuthResponse: %v", err)
	}
	if auth.Nombre != "Ana Ruiz" || auth.ExpiresIn != "7200" || auth.Token == "" || auth.RefreshToken == "" {
		t.Fatalf("auth inesperado: %+v", auth)
	}
	claims := &Claims{}
	if _, err := jwt.ParseWithClaims(auth.Token, claims, func(*jwt.Token) (interface{}, error) { return jwtSecret, nil }); err != nil {
		t.Fatal(err)
	}
	if claims.Rol != "Administrador" || claims.TokenType != "" {
		t.Fatalf("claims inesperados: %+v", claims)
	}
}

func TestRefreshResponseShape(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	r.Header.Set("Authorization", "Bearer "+refreshTokenStr(t))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := &LoginController{}
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}
	c.RefreshToken()
	if w.Code != http.StatusOK {
		t.Fatalf("esperaba 200, obtuve %d", w.Code)
	}
	var envelope struct {
		Data json.RawMessage `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &envelope)
	dec := json.NewDecoder(bytes.NewReader(envelope.Data))
	dec.DisallowUnknownFields()
	var auth models.AuthResponse
	if err := dec.Decode(&auth); err != nil || auth.ExpiresIn != "7200" {
		t.Fatalf("respuesta de refresh inesperada: %v %+v", err, auth)
	}
}
