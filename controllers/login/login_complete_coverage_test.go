package login

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
	"github.com/golang-jwt/jwt/v5"
)

// TestRefreshToken_WithoutBearerPrefix cubre el caso de token sin prefijo "Bearer"
func TestRefreshToken_WithoutBearerPrefix(t *testing.T) {
	// Generar un refresh token válido
	now := time.Now()
	refreshClaims := &RefreshClaims{
		Documento: 123456,
		Rol:       "Cliente",
		Nombre:    "Test User",
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(30 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	// Enviar request SIN prefijo "Bearer " (solo el token)
	r := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	r.Header.Set("Authorization", tokenString) // SIN "Bearer "
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)

	c := &LoginController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.RefreshToken()

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp models.ApiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Code != http.StatusOK {
		t.Errorf("Expected response code 200, got %d", resp.Code)
	}
}

// TestValidateToken_WithoutBearerPrefix cubre el caso de token sin prefijo "Bearer" en ValidateToken
func TestValidateToken_WithoutBearerPrefix(t *testing.T) {
	// Generar un access token válido
	now := time.Now()
	claims := &Claims{
		Documento: 123456,
		Rol:       "Admin",
		Nombre:    "Admin User",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(120 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	// Request protegido SIN prefijo "Bearer "
	r := httptest.NewRequest(http.MethodGet, "/restaurante/v1/trabajadores", nil)
	r.Header.Set("Authorization", tokenString) // SIN "Bearer "
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)

	ValidateToken(ctx)

	// ValidateToken no debe devolver error si el token es válido, incluso sin Bearer
	if w.Code == http.StatusUnauthorized {
		t.Fatalf("Expected ValidateToken to accept token without Bearer prefix, got 401")
	}
}

// TestRefreshToken_TokenInvalidNotValid cubre el caso de token.Valid == false
func TestRefreshToken_TokenInvalidNotValid(t *testing.T) {
	// Token expirado (para que token.Valid sea false)
	now := time.Now()
	refreshClaims := &RefreshClaims{
		Documento: 123456,
		Rol:       "Cliente",
		Nombre:    "Test User",
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(-1 * time.Hour)), // Expirado
			IssuedAt:  jwt.NewNumericDate(now.Add(-2 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	r := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	r.Header.Set("Authorization", "Bearer "+tokenString)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)

	c := &LoginController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.RefreshToken()

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("Expected status 401 for expired token, got %d", w.Code)
	}
}

// TestValidateToken_InvalidTokenNotValid cubre el caso de token.Valid == false en ValidateToken
func TestValidateToken_InvalidTokenNotValid(t *testing.T) {
	// Token expirado
	now := time.Now()
	claims := &Claims{
		Documento: 123456,
		Rol:       "Cliente",
		Nombre:    "Test User",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(-1 * time.Hour)), // Expirado
			IssuedAt:  jwt.NewNumericDate(now.Add(-2 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	r := httptest.NewRequest(http.MethodGet, "/restaurante/v1/trabajadores", nil)
	r.Header.Set("Authorization", "Bearer "+tokenString)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)

	ValidateToken(ctx)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("Expected status 401 for expired token, got %d", w.Code)
	}
}

// TestGenerateTokens_RefreshTokenSignError cubre el error al firmar el refresh token
func TestGenerateTokens_RefreshTokenSignError(t *testing.T) {
	origSecret := jwtSecret
	origMethod := signingMethod
	defer func() {
		jwtSecret = origSecret
		signingMethod = origMethod
	}()

	jwtSecret = []byte("valid-secret")

	// Mock que falla solo en el segundo SignedString (refresh token)
	callCount := 0
	signingMethod = &customFailingMethod{
		failOnCall: 2, // Fallar en la segunda firma (refresh token)
		callCount:  &callCount,
	}

	_, _, err := generateTokens(123456, "Cliente", "Test")
	if err == nil {
		t.Fatal("Expected error when signing refresh token fails")
	}

	if err.Error() != "error al generar refresh token: key is invalid" {
		t.Errorf("Expected refresh token error, got: %v", err)
	}
}

// customFailingMethod simula un signing method que falla en una llamada específica
type customFailingMethod struct {
	failOnCall int
	callCount  *int
}

func (m *customFailingMethod) Verify(signingString string, signature []byte, key interface{}) error {
	return nil
}

func (m *customFailingMethod) Sign(signingString string, key interface{}) ([]byte, error) {
	*m.callCount++
	if *m.callCount == m.failOnCall {
		return nil, jwt.ErrInvalidKey
	}
	// Usar el método real para otras llamadas
	return jwt.SigningMethodHS256.Sign(signingString, key)
}

func (m *customFailingMethod) Alg() string {
	return "HS256"
}

// TestLogin_RateLimitAfterReset cubre la línea 138 (después de reset)
func TestLogin_RateLimitAfterReset(t *testing.T) {
	origRL := loginRL
	origMax := loginMaxReq
	loginRL = newRateLimiter()
	loginMaxReq = 2
	defer func() {
		loginRL = origRL
		loginMaxReq = origMax
	}()

	loginReq := models.LoginRequest{
		Documento: 123456,
		Password:  "test",
	}
	body, _ := json.Marshal(loginReq)

	// Primera request
	r1 := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	r1.RemoteAddr = "192.168.1.100:1234"
	w1 := httptest.NewRecorder()
	ctx1 := context.NewContext()
	ctx1.Reset(w1, r1)
	ctx1.Input.RequestBody = body

	c1 := &LoginController{}
	c1.Ctx = ctx1
	c1.Data = make(map[interface{}]interface{})

	// Mock ORM para que falle (no nos importa el resultado del login)
	origNewOrm := newOrm
	defer func() { newOrm = origNewOrm }()
	newOrm = func() orm.Ormer {
		return &mockLoginOrmer{ReadFunc: func(v interface{}, cols ...string) error {
			return orm.ErrNoRows
		}}
	}

	c1.Login()

	// Segunda request (debería pasar, aún tenemos 1 más)
	r2 := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	r2.RemoteAddr = "192.168.1.100:1234"
	w2 := httptest.NewRecorder()
	ctx2 := context.NewContext()
	ctx2.Reset(w2, r2)
	ctx2.Input.RequestBody = body

	c2 := &LoginController{}
	c2.Ctx = ctx2
	c2.Data = make(map[interface{}]interface{})

	c2.Login()

	// Tercera request (debería ser bloqueada)
	r3 := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	r3.RemoteAddr = "192.168.1.100:1234"
	w3 := httptest.NewRecorder()
	ctx3 := context.NewContext()
	ctx3.Reset(w3, r3)
	ctx3.Input.RequestBody = body

	c3 := &LoginController{}
	c3.Ctx = ctx3
	c3.Data = make(map[interface{}]interface{})

	c3.Login()

	if w3.Code != http.StatusTooManyRequests {
		t.Fatalf("Expected status 429 on third request, got %d", w3.Code)
	}

	// Simular reset del tiempo
	ip := clientIP(r3)
	entry := loginRL.m[ip]
	entry.reset = time.Now().Add(-1 * time.Minute) // Expirar el reset

	// Cuarta request (después del reset, debería permitir)
	r4 := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	r4.RemoteAddr = "192.168.1.100:1234"
	w4 := httptest.NewRecorder()
	ctx4 := context.NewContext()
	ctx4.Reset(w4, r4)
	ctx4.Input.RequestBody = body

	c4 := &LoginController{}
	c4.Ctx = ctx4
	c4.Data = make(map[interface{}]interface{})

	c4.Login()

	// No debería ser rate limited (línea 136-138)
	if w4.Code == http.StatusTooManyRequests {
		t.Fatal("Expected request to be allowed after reset")
	}
}

// TestValidateToken_DevSwaggerBypass cubre las líneas 431-439
func TestValidateToken_DevSwaggerBypass(t *testing.T) {
	origRunMode := web.BConfig.RunMode
	web.BConfig.RunMode = "dev"
	defer func() { web.BConfig.RunMode = origRunMode }()

	// Test con Referer a swagger
	r1 := httptest.NewRequest(http.MethodGet, "/restaurante/v1/trabajadores", nil)
	r1.Header.Set("Referer", "http://localhost:8080/swagger/index.html")
	w1 := httptest.NewRecorder()
	ctx1 := context.NewContext()
	ctx1.Reset(w1, r1)

	ValidateToken(ctx1)

	// Debería permitir sin token en dev con referer swagger
	if w1.Code == http.StatusUnauthorized {
		t.Fatal("Expected swagger referer to bypass auth in dev mode")
	}

	// Test con path que empieza con /swagger/
	r2 := httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil)
	w2 := httptest.NewRecorder()
	ctx2 := context.NewContext()
	ctx2.Reset(w2, r2)

	ValidateToken(ctx2)

	// Debería permitir sin token para paths de swagger en dev
	if w2.Code == http.StatusUnauthorized {
		t.Fatal("Expected /swagger/ path to bypass auth in dev mode")
	}
}
