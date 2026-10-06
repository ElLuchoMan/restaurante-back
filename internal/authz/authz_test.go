package authz

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"restaurante/controllers/login"
	"restaurante/models"

	"github.com/beego/beego/v2/server/web"
	beecontext "github.com/beego/beego/v2/server/web/context"
	"github.com/golang-jwt/jwt/v5"
)

func token(t *testing.T, doc int64, rol string) string {
	t.Helper()
	tk := jwt.NewWithClaims(jwt.SigningMethodHS256, login.Claims{
		Documento:        doc,
		Rol:              rol,
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))},
	})
	s, err := tk.SignedString(login.GetJWTSecret())
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func controller(auth string) (*web.Controller, *httptest.ResponseRecorder) {
	r := httptest.NewRequest(http.MethodGet, "/x", nil)
	if auth != "" {
		r.Header.Set("Authorization", auth)
	}
	w := httptest.NewRecorder()
	ctx := beecontext.NewContext()
	ctx.Reset(w, r)
	return &web.Controller{Ctx: ctx, Data: map[interface{}]interface{}{}}, w
}

func TestRequireAuth(t *testing.T) {
	c, w := controller("")
	if cl, ok := RequireAuth(c); ok || cl != nil || w.Code != http.StatusUnauthorized {
		t.Fatalf("sin token debe ser 401: %v %v %d", cl, ok, w.Code)
	}
	c, w = controller("Bearer basura")
	if _, ok := RequireAuth(c); ok || w.Code != http.StatusUnauthorized {
		t.Fatalf("token inválido debe ser 401: %d", w.Code)
	}
	c, _ = controller("Bearer " + token(t, 7, login.RolCliente))
	if cl, ok := RequireAuth(c); !ok || cl.Documento != 7 {
		t.Fatalf("token válido debe pasar: %v %v", cl, ok)
	}
}

func TestRequireAdmin(t *testing.T) {
	cases := []struct {
		name, auth string
		ok         bool
		status     int
	}{
		{"sin token", "", false, http.StatusUnauthorized},
		{"cliente", "Bearer " + token(t, 7, login.RolCliente), false, http.StatusForbidden},
		{"mesero", "Bearer " + token(t, 3, "Mesero"), false, http.StatusForbidden},
		{"administrador", "Bearer " + token(t, 1, string(models.RolAdministrador)), true, http.StatusOK},
	}
	for _, tc := range cases {
		c, w := controller(tc.auth)
		cl, ok := RequireAdmin(c)
		if ok != tc.ok || (!ok && (cl != nil || w.Code != tc.status)) || (ok && cl == nil) {
			t.Fatalf("%s: ok=%v código=%d", tc.name, ok, w.Code)
		}
	}
}

func TestResolveCliente(t *testing.T) {
	cliente := &login.Claims{Documento: 7, Rol: login.RolCliente}
	staff := &login.Claims{Documento: 3, Rol: "Mesero"}
	admin := &login.Claims{Documento: 1, Rol: string(models.RolAdministrador)}
	cases := []struct {
		name   string
		claims *login.Claims
		body   int64
		want   int64
		status int
	}{
		{"cliente sin body", cliente, 0, 7, http.StatusOK},
		{"cliente con su propio id", cliente, 7, 7, http.StatusOK},
		{"cliente con otro id", cliente, 8, 0, http.StatusForbidden},
		{"cliente con id negativo", cliente, -7, 0, http.StatusForbidden},
		{"token sin documento", &login.Claims{Rol: login.RolCliente}, 0, 0, http.StatusForbidden},
		{"trabajador indica cliente", staff, 8, 8, http.StatusOK},
		{"admin indica cliente", admin, 8, 8, http.StatusOK},
		{"trabajador sin cliente", staff, 0, 0, http.StatusBadRequest},
		{"admin con cliente negativo", admin, -1, 0, http.StatusBadRequest},
	}
	for _, tc := range cases {
		c, w := controller("")
		got, ok := ResolveCliente(c, tc.claims, tc.body)
		if tc.status == http.StatusOK {
			if !ok || got != tc.want {
				t.Fatalf("%s: got=%d ok=%v", tc.name, got, ok)
			}
			continue
		}
		if ok || got != 0 || w.Code != tc.status {
			t.Fatalf("%s: ok=%v código=%d (esperaba %d)", tc.name, ok, w.Code, tc.status)
		}
	}
}

func TestRequireStaff(t *testing.T) {
	cases := []struct {
		name, auth string
		ok         bool
		status     int
	}{
		{"sin token", "", false, http.StatusUnauthorized},
		{"cliente", "Bearer " + token(t, 7, login.RolCliente), false, http.StatusForbidden},
		{"mesero", "Bearer " + token(t, 3, "Mesero"), true, http.StatusOK},
		{"administrador", "Bearer " + token(t, 1, string(models.RolAdministrador)), true, http.StatusOK},
	}
	for _, tc := range cases {
		c, w := controller(tc.auth)
		cl, ok := RequireStaff(c)
		if ok != tc.ok || (!ok && (cl != nil || w.Code != tc.status)) || (ok && cl == nil) {
			t.Fatalf("%s: ok=%v código=%d", tc.name, ok, w.Code)
		}
	}
}

func TestEsDuenio(t *testing.T) {
	cases := []struct {
		name   string
		claims *login.Claims
		doc    int64
		want   bool
	}{
		{"personal", &login.Claims{Documento: 3, Rol: "Mesero"}, 99, true},
		{"cliente dueño", &login.Claims{Documento: 7, Rol: login.RolCliente}, 7, true},
		{"cliente ajeno", &login.Claims{Documento: 7, Rol: login.RolCliente}, 8, false},
		{"cliente sin documento", &login.Claims{Documento: 0, Rol: login.RolCliente}, 0, false},
	}
	for _, tc := range cases {
		if got := EsDuenio(tc.claims, tc.doc); got != tc.want {
			t.Fatalf("%s: %v", tc.name, got)
		}
	}
}
