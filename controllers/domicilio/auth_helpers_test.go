package domicilio

import (
	"testing"
	"time"

	"restaurante/controllers/login"

	"github.com/golang-jwt/jwt/v5"
)

const (
	rolAdmin    = "Administrador"
	rolMesero   = "Mesero"
	rolDomi     = "Domiciliario"
	rolClienteT = login.RolCliente
)

// authHeader es el Authorization que newCtx envía: por defecto un Administrador,
// para que los tests de lógica no dependan de la autorización.
var authHeader = tokenDe(rolAdmin, 1)

// tokenDe devuelve un header Authorization Bearer firmado para rol y documento.
func tokenDe(rol string, documento int64) string {
	claims := login.Claims{
		Documento: documento, Rol: rol, Nombre: "Prueba",
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))},
	}
	tok, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(login.GetJWTSecret())
	if err != nil {
		panic(err)
	}
	return "Bearer " + tok
}

// como hace que las peticiones del test lleguen con el rol y documento dados.
func como(t *testing.T, rol string, documento int64) {
	t.Helper()
	prev := authHeader
	authHeader = tokenDe(rol, documento)
	t.Cleanup(func() { authHeader = prev })
}

// sinToken hace que las peticiones del test lleguen sin Authorization.
func sinToken(t *testing.T) {
	t.Helper()
	prev := authHeader
	authHeader = ""
	t.Cleanup(func() { authHeader = prev })
}
