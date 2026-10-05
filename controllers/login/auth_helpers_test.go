package login

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/beego/beego/v2/server/web/context"
)

func ctxWithAuth(method, auth string) (*context.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, "/x", nil)
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	w := httptest.NewRecorder()
	c := context.NewContext()
	c.Reset(w, req)
	return c, w
}

func TestClaimsFromContext(t *testing.T) {
	c, _ := ctxWithAuth(http.MethodGet, "")
	if ClaimsFromContext(c) != nil {
		t.Fatal("sin token debe ser nil")
	}
	c, _ = ctxWithAuth(http.MethodGet, "Bearer basura")
	if ClaimsFromContext(c) != nil {
		t.Fatal("token inválido debe ser nil")
	}
	c, _ = ctxWithAuth(http.MethodGet, "Bearer "+accessToken(t, "Mesero"))
	cl := ClaimsFromContext(c)
	if cl == nil || !cl.IsStaff() || cl.IsAdmin() {
		t.Fatalf("claims mesero inesperados: %+v", cl)
	}
	c, _ = ctxWithAuth(http.MethodGet, accessToken(t, "Administrador"))
	if cl = ClaimsFromContext(c); cl == nil || !cl.IsAdmin() {
		t.Fatal("admin esperado (sin prefijo Bearer)")
	}
	c, _ = ctxWithAuth(http.MethodGet, "Bearer "+accessToken(t, RolCliente))
	if cl = ClaimsFromContext(c); cl.IsStaff() || cl.IsAdmin() {
		t.Fatal("cliente no es staff ni admin")
	}
	var nilClaims *Claims
	if nilClaims.IsStaff() || nilClaims.IsAdmin() {
		t.Fatal("nil no es staff ni admin")
	}
	if (&Claims{}).IsStaff() {
		t.Fatal("rol vacío no es staff")
	}
}

func TestValidateStaff(t *testing.T) {
	cases := []struct {
		name, method, auth string
		want               int
	}{
		{"options", http.MethodOptions, "", http.StatusOK},
		{"sin token", http.MethodGet, "", http.StatusUnauthorized},
		{"cliente", http.MethodGet, "Bearer " + accessToken(t, RolCliente), http.StatusForbidden},
		{"mesero", http.MethodGet, "Bearer " + accessToken(t, "Mesero"), 0},
	}
	for _, tc := range cases {
		c, w := ctxWithAuth(tc.method, tc.auth)
		ValidateStaff(c)
		got := w.Code
		if tc.want == 0 {
			got = c.Output.Status
			if got != 0 && got != http.StatusOK {
				t.Fatalf("%s: no debía rechazar, status %d", tc.name, got)
			}
			continue
		}
		if tc.name == "options" {
			if c.Output.Status != tc.want {
				t.Fatalf("options: %d", c.Output.Status)
			}
			continue
		}
		if got != tc.want {
			t.Fatalf("%s: esperaba %d, obtuve %d", tc.name, tc.want, got)
		}
	}
}
