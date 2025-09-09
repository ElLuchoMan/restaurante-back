package controllers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
	"github.com/golang-jwt/jwt/v5"
)

func TestLoadJWTSecret_FromEnv(t *testing.T) {
	t.Setenv("JWT_SECRET", "abc123")
	b := loadJWTSecret()
	if string(b) != "abc123" {
		t.Fatalf("esperaba abc123, obtuve %q", string(b))
	}
}

func TestLoadJWTSecret_DevFallback(t *testing.T) {
	// Asegurar modo no prod
	orig := web.BConfig.RunMode
	web.BConfig.RunMode = "dev"
	t.Cleanup(func() { web.BConfig.RunMode = orig })
	os.Unsetenv("JWT_SECRET")
	b := loadJWTSecret()
	if len(b) == 0 {
		t.Fatal("se esperaba secreto no vacío en dev/test")
	}
}

func TestLoadJWTSecret_ProdPanicsWithoutEnv(t *testing.T) {
	orig := web.BConfig.RunMode
	web.BConfig.RunMode = "prod"
	t.Cleanup(func() { web.BConfig.RunMode = orig })
	os.Unsetenv("JWT_SECRET")
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("se esperaba pánico cuando JWT_SECRET no está configurado en prod")
		}
	}()
	_ = loadJWTSecret()
}

func TestGetEnvIntDefault(t *testing.T) {
	if v := getEnvIntDefault("NO_SET", 7); v != 7 {
		t.Fatalf("esperaba 7, got %d", v)
	}
	t.Setenv("X_POS", "5")
	if v := getEnvIntDefault("X_POS", 7); v != 5 {
		t.Fatalf("esperaba 5, got %d", v)
	}
	t.Setenv("X_ZERO", "0")
	if v := getEnvIntDefault("X_ZERO", 7); v != 7 {
		t.Fatalf("esperaba default 7 para 0, got %d", v)
	}
	t.Setenv("X_NEG", "-1")
	if v := getEnvIntDefault("X_NEG", 7); v != 7 {
		t.Fatalf("esperaba default 7 para -1, got %d", v)
	}
	t.Setenv("X_BAD", "bad")
	if v := getEnvIntDefault("X_BAD", 7); v != 7 {
		t.Fatalf("esperaba default 7 para invalido, got %d", v)
	}
}

func TestClientIP(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "127.0.0.1:1234"
	if ip := clientIP(r); ip != "127.0.0.1" && ip != "127.0.0.1:1234" {
		t.Fatalf("ip inesperada: %s", ip)
	}

	r2 := httptest.NewRequest(http.MethodGet, "/", nil)
	r2.Header.Set("X-Forwarded-For", "203.0.113.9, 70.41.3.18, 150.172.238.178")
	r2.RemoteAddr = "10.0.0.1:2222"
	if ip := clientIP(r2); ip != "203.0.113.9" {
		t.Fatalf("esperaba 203.0.113.9, got %s", ip)
	}
}

func TestAllowLogin_RateLimitAndReset(t *testing.T) {
	// Aislar estado global
	origRL := loginRL
	origMax := loginMaxReq
	loginRL = newRateLimiter()
	loginMaxReq = 2
	t.Cleanup(func() { loginRL = origRL; loginMaxReq = origMax })

	r := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader("{}"))
	r.RemoteAddr = "198.51.100.10:4444"
	if !allowLogin(r) { t.Fatal("primera debería permitir") }
	if !allowLogin(r) { t.Fatal("segunda debería permitir") }
	if allowLogin(r) { t.Fatal("tercera debería bloquear") }

	// Simular ventana expirada
	ip := clientIP(r)
	loginRL.m[ip].reset = loginRL.m[ip].reset.Add(-2 * loginWindow)
	if !allowLogin(r) { t.Fatal("después de reset debería permitir") }
}

func TestGenerateJWT_WritesTokenAndClaims(t *testing.T) {
	// Forzar secreto conocido para validar token
	orig := jwtSecret
	jwtSecret = []byte("testsecret123")
	t.Cleanup(func() { jwtSecret = orig })

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/login", nil)
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := &LoginController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	generateJWT(c, 42, "Admin", "Juan Perez")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "\"token\"") || !strings.Contains(body, "Juan Perez") {
		t.Fatalf("respuesta no contiene token/nombre: %s", body)
	}
	// Extraer token de la respuesta simple (sin map JSON estricto para mantener el test simple)
	start := strings.Index(body, "\"token\":\"")
	if start < 0 { t.Fatalf("no se encontró token en body: %s", body) }
	start += len("\"token\":\"")
	end := strings.Index(body[start:], "\"")
	if end < 0 { t.Fatalf("no se cerró token en body") }
	tokenStr := body[start : start+end]

	claims := &Claims{}
	tok, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) { return jwtSecret, nil }, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil || !tok.Valid {
		t.Fatalf("token inválido: %v", err)
	}
	if claims.Documento != 42 || claims.Rol != "Admin" || claims.Nombre != "Juan Perez" {
		t.Fatalf("claims inesperados: %+v", claims)
	}
}
