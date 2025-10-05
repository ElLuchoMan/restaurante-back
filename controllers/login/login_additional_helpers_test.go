package login

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TestGetJWTSecret verifica que GetJWTSecret retorne el secreto configurado
func TestGetJWTSecret(t *testing.T) {
	secret := GetJWTSecret()
	if len(secret) == 0 {
		t.Fatal("Expected non-empty JWT secret")
	}
}

// TestParseTokenClaims_ValidToken verifica que ParseTokenClaims funcione con un token válido
func TestParseTokenClaims_ValidToken(t *testing.T) {
	// Generar un token válido
	now := time.Now()
	claims := &Claims{
		Documento: 12345,
		Rol:       "Administrador",
		Nombre:    "Test User",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(120 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}

	// Parsear el token
	parsedClaims, err := ParseTokenClaims(tokenString)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if parsedClaims.Documento != 12345 {
		t.Errorf("Expected documento 12345, got %d", parsedClaims.Documento)
	}

	if parsedClaims.Rol != "Administrador" {
		t.Errorf("Expected rol Administrador, got %s", parsedClaims.Rol)
	}

	if parsedClaims.Nombre != "Test User" {
		t.Errorf("Expected nombre 'Test User', got %s", parsedClaims.Nombre)
	}
}

// TestParseTokenClaims_InvalidToken verifica que ParseTokenClaims falle con un token inválido
func TestParseTokenClaims_InvalidToken(t *testing.T) {
	_, err := ParseTokenClaims("invalid.token.here")
	if err == nil {
		t.Fatal("Expected error for invalid token, got nil")
	}
}

// TestParseTokenClaims_ExpiredToken verifica que ParseTokenClaims falle con un token expirado
func TestParseTokenClaims_ExpiredToken(t *testing.T) {
	// Generar un token expirado
	now := time.Now()
	claims := &Claims{
		Documento: 12345,
		Rol:       "Administrador",
		Nombre:    "Test User",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(-1 * time.Hour)), // Expirado hace 1 hora
			IssuedAt:  jwt.NewNumericDate(now.Add(-2 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}

	// Parsear el token expirado
	_, err = ParseTokenClaims(tokenString)
	if err == nil {
		t.Fatal("Expected error for expired token, got nil")
	}
}

// TestParseTokenClaims_EmptyToken verifica que ParseTokenClaims falle con un token vacío
func TestParseTokenClaims_EmptyToken(t *testing.T) {
	_, err := ParseTokenClaims("")
	if err == nil {
		t.Fatal("Expected error for empty token, got nil")
	}
}
