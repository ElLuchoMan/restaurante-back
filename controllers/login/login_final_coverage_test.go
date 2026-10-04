package login

import (
	"net/http/httptest"
	"os"
	"testing"

	"github.com/beego/beego/v2/server/web"
	"github.com/golang-jwt/jwt/v5"
)

func TestLoadJWTSecret_ProductionWithoutSecret(t *testing.T) {

	origSecret := os.Getenv("JWT_SECRET")
	origMode := web.BConfig.RunMode
	defer func() {
		os.Setenv("JWT_SECRET", origSecret)
		web.BConfig.RunMode = origMode
	}()

	os.Unsetenv("JWT_SECRET")
	web.BConfig.RunMode = "prod"

	defer func() {
		if r := recover(); r != nil {

			if r != "JWT_SECRET no configurado" {
				t.Errorf("Expected panic 'JWT_SECRET no configurado', got %v", r)
			}
		}
	}()

	secret := loadJWTSecret()
	if secret == nil {
		t.Error("Expected secret to be loaded")
	}
}

func TestLoadJWTSecret_DevModeWithRandomGeneration(t *testing.T) {

	origSecret := os.Getenv("JWT_SECRET")
	origMode := web.BConfig.RunMode
	defer func() {
		os.Setenv("JWT_SECRET", origSecret)
		web.BConfig.RunMode = origMode
	}()

	os.Unsetenv("JWT_SECRET")
	web.BConfig.RunMode = "dev"

	secret := loadJWTSecret()
	if len(secret) == 0 {
		t.Error("Expected secret to be generated")
	}

	if string(secret) != "dev-insecure-default" && len(secret) != 32 {
		t.Errorf("Expected either fallback or 32-byte random secret, got %d bytes", len(secret))
	}
}

func TestGenerateTokens_EmptyJWTSecret(t *testing.T) {

	origSecret := jwtSecret
	defer func() { jwtSecret = origSecret }()

	jwtSecret = []byte{}

	_, _, err := generateTokens(123456, "Cliente", "Test User")
	if err == nil {
		t.Error("Expected error when JWT secret is empty")
	}

	if err != nil && err.Error() != "secreto JWT no configurado" {
		t.Errorf("Expected specific error message, got: %v", err)
	}
}

func TestGenerateTokens_SigningError(t *testing.T) {

	origSecret := jwtSecret
	origMethod := signingMethod
	defer func() {
		jwtSecret = origSecret
		signingMethod = origMethod
	}()

	jwtSecret = []byte("valid-secret-key")
	signingMethod = &invalidSigningMethod{}

	_, _, err := generateTokens(123456, "Cliente", "Test User")
	if err == nil {
		t.Error("Expected error with invalid signing method")
	}
}

func TestClientIP_RemoteAddrWithoutPort(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		xForwarded string
		expected   string
	}{
		{
			name:       "IP with port",
			remoteAddr: "192.168.1.100:54321",
			xForwarded: "",
			expected:   "192.168.1.100",
		},
		{
			name:       "IP without port",
			remoteAddr: "192.168.1.100",
			xForwarded: "",
			expected:   "192.168.1.100",
		},
		{
			name:       "X-Forwarded-For present",
			remoteAddr: "10.0.0.1:8080",
			xForwarded: "203.0.113.1, 198.51.100.1",
			expected:   "203.0.113.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.xForwarded != "" {
				req.Header.Set("X-Forwarded-For", tt.xForwarded)
			}
			result := clientIP(req)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestAllowLogin_RateLimitReset(t *testing.T) {

	rlMutex.Lock()
	loginRL = newRateLimiter()
	rlMutex.Unlock()

	req := httptest.NewRequest("POST", "/login", nil)
	req.RemoteAddr = "10.0.0.1:12345"

	for i := 0; i < loginMaxReq; i++ {
		if !allowLogin(req) {
			t.Fatalf("Expected request %d to be allowed", i+1)
		}
	}

	if allowLogin(req) {
		t.Error("Expected request to be rate limited")
	}
}

type invalidSigningMethod struct{}

func (m *invalidSigningMethod) Verify(signingString string, signature []byte, key interface{}) error {
	return jwt.ErrSignatureInvalid
}

func (m *invalidSigningMethod) Sign(signingString string, key interface{}) ([]byte, error) {
	return nil, jwt.ErrInvalidKey
}

func (m *invalidSigningMethod) Alg() string {
	return "INVALID"
}
