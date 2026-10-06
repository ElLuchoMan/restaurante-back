package routers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	beego "github.com/beego/beego/v2/server/web"
)

func serve(method, path string) *httptest.ResponseRecorder {
	r, _ := http.NewRequest(method, path, strings.NewReader("{}"))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	return w
}

func TestProtectedWriteFiltersRequireToken(t *testing.T) {
	paths := []string{
		"/restaurante/v1/productos",
		"/restaurante/v1/categorias",
		"/restaurante/v1/subcategorias",
	}
	for _, p := range paths {
		for _, m := range []string{httpMethodPost, httpMethodPut, httpMethodDelete} {
			w := serve(m, p)
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("%s %s: esperado 401, obtenido %d", m, p, w.Code)
			}
		}
	}
}

func TestProtectedFiltersSkipTokenForOtherMethods(t *testing.T) {
	// OPTIONS no es POST/PUT/DELETE: el filtro no valida token.
	for _, p := range []string{"/restaurante/v1/productos", "/restaurante/v1/categorias", "/restaurante/v1/subcategorias"} {
		w := serve(http.MethodOptions, p)
		if w.Code == http.StatusUnauthorized {
			t.Fatalf("OPTIONS %s no deberia exigir token", p)
		}
	}
}

func TestNamespaceBeforeValidateTokenRejectsWithoutToken(t *testing.T) {
	w := serve(http.MethodGet, "/restaurante/v1/pedidos")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("esperado 401, obtenido %d", w.Code)
	}
}

func TestCheckoutRouteRequiresToken(t *testing.T) {
	w := serve(http.MethodPost, "/restaurante/v1/pedidos/checkout")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("esperado 401, obtenido %d", w.Code)
	}
}

func TestRegisterSwaggerAssets(t *testing.T) {
	registerSwaggerAssets(filepath.Join(t.TempDir(), "no-existe"))

	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "subdir"), 0o755); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		"asset-test.js":     "console.log('ok')",
		"index.html":        "<html></html>",
		"asset-test.js.map": "{}",
		"doc-test.json":     "{}",
	}
	for n, c := range files {
		if err := os.WriteFile(filepath.Join(dir, n), []byte(c), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	registerSwaggerAssets(dir)

	if w := serve(http.MethodGet, "/swagger/asset-test.js"); w.Code != http.StatusOK || w.Body.String() != "console.log('ok')" {
		t.Fatalf("asset no servido: %d %q", w.Code, w.Body.String())
	}
}

// Los endpoints PATCH de push exigen token y el preflight CORS debe permitir PATCH.
func TestPushPatchRoutesRequireTokenAndCORSAllowsPatch(t *testing.T) {
	for _, p := range []string{"/restaurante/v1/push/dispositivos/visto", "/restaurante/v1/push/dispositivos/topics"} {
		if w := serve(http.MethodPatch, p); w.Code != http.StatusUnauthorized {
			t.Fatalf("PATCH %s: esperado 401, obtenido %d", p, w.Code)
		}
	}

	r, _ := http.NewRequest(http.MethodOptions, "/restaurante/v1/push/dispositivos/visto", nil)
	r.Header.Set("Origin", "http://localhost:4200")
	r.Header.Set("Access-Control-Request-Method", "PATCH")
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	if !strings.Contains(w.Header().Get("Access-Control-Allow-Methods"), "PATCH") {
		t.Fatalf("el preflight no permite PATCH: %v", w.Header())
	}
}
