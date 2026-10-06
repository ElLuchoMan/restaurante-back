package routers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	loginc "restaurante/controllers/login"

	beego "github.com/beego/beego/v2/server/web"
	"github.com/golang-jwt/jwt/v5"
)

func tokenConRol(t *testing.T, rol, tipo string) string {
	t.Helper()
	claims := &loginc.Claims{
		Documento: 1, Rol: rol, Nombre: "Prueba", TokenType: tipo,
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))},
	}
	s, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(loginc.GetJWTSecret())
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func serveAuth(method, path, token string) *httptest.ResponseRecorder {
	r, _ := http.NewRequest(method, path, strings.NewReader("{}"))
	r.Header.Set("Content-Type", "application/json")
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	return w
}

// /trabajadores solo es accesible por el rol Administrador en todos sus métodos.
func TestTrabajadoresSoloAdministrador(t *testing.T) {
	rutas := []struct{ method, path string }{
		{http.MethodGet, "/restaurante/v1/trabajadores?rol=Gerente"},
		{http.MethodGet, "/restaurante/v1/trabajadores/search"},
		{http.MethodPost, "/restaurante/v1/trabajadores"},
		{http.MethodPut, "/restaurante/v1/trabajadores"},
		{http.MethodDelete, "/restaurante/v1/trabajadores"},
	}
	for _, r := range rutas {
		if w := serveAuth(r.method, r.path, ""); w.Code != http.StatusUnauthorized {
			t.Fatalf("%s %s sin token: esperado 401, obtenido %d", r.method, r.path, w.Code)
		}
		if w := serveAuth(r.method, r.path, "no-es-un-jwt"); w.Code != http.StatusUnauthorized {
			t.Fatalf("%s %s token inválido: esperado 401, obtenido %d", r.method, r.path, w.Code)
		}
		if w := serveAuth(r.method, r.path, tokenConRol(t, "Administrador", "refresh")); w.Code != http.StatusUnauthorized {
			t.Fatalf("%s %s refresh token: esperado 401, obtenido %d", r.method, r.path, w.Code)
		}
		for _, rol := range []string{"Mesero", "Cocinero", "Domiciliario", "Oficios_varios", "Cliente"} {
			w := serveAuth(r.method, r.path, tokenConRol(t, rol, ""))
			if w.Code != http.StatusForbidden || !strings.Contains(w.Body.String(), `"code":403`) {
				t.Fatalf("%s %s rol %s: esperado 403, obtenido %d %s", r.method, r.path, rol, w.Code, w.Body.String())
			}
		}
		// un administrador pasa el filtro: llega al controlador, que valida (400) antes de tocar la BD
		w := serveAuth(r.method, r.path, tokenConRol(t, "Administrador", ""))
		if w.Code != http.StatusBadRequest {
			t.Fatalf("%s %s administrador: esperado 400 del controlador, obtenido %d %s", r.method, r.path, w.Code, w.Body.String())
		}
	}
}
