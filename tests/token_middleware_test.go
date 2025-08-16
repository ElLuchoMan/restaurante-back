package test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	beego "github.com/beego/beego/v2/server/web"
	jwt "github.com/dgrijalva/jwt-go"
	. "github.com/smartystreets/goconvey/convey"
)

// TestProtectedEndpointWithoutToken ensures that endpoints protected by the token middleware return 401 when no token is provided.
func TestProtectedEndpointWithoutToken(t *testing.T) {
	r, _ := http.NewRequest("GET", "/restaurante/v1/clientes", nil)
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	Convey("GET /restaurante/v1/clientes without token should be unauthorized", t, func() {
		So(w.Code, ShouldEqual, http.StatusUnauthorized)
	})
}

// TestOptionsBypassesToken verifies that OPTIONS requests bypass token validation to allow CORS preflight.
func TestOptionsBypassesToken(t *testing.T) {
	r, _ := http.NewRequest("OPTIONS", "/restaurante/v1/clientes", nil)
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	Convey("OPTIONS request should bypass token validation", t, func() {
		So(w.Code, ShouldNotEqual, http.StatusUnauthorized)
		So(w.Code, ShouldEqual, http.StatusMethodNotAllowed)
	})
}

// TestPublicPostWithoutToken confirms that public POST endpoints can be accessed without a token and return a validation error instead.
func TestPublicPostWithoutToken(t *testing.T) {
	rBody := strings.NewReader("{")
	r, _ := http.NewRequest("POST", "/restaurante/v1/clientes", rBody)
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	Convey("POST /restaurante/v1/clientes without token should not return unauthorized", t, func() {
		So(w.Code, ShouldNotEqual, http.StatusUnauthorized)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
	})
}

// generateValidToken creates a JWT for testing purposes using the same secret as the application.
func generateValidToken() string {
	claims := jwt.MapClaims{
		"documento": 1,
		"rol":       "Cliente",
		"nombre":    "Test User",
		"exp":       time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(os.Getenv("cocina-de-maria")))
	return tokenString
}

// TestProtectedEndpointWithInvalidToken verifies that an invalid token is rejected with 401.
func TestProtectedEndpointWithInvalidToken(t *testing.T) {
	r, _ := http.NewRequest("GET", "/restaurante/v1/clientes", nil)
	r.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	Convey("GET /restaurante/v1/clientes with invalid token should be unauthorized", t, func() {
		So(w.Code, ShouldEqual, http.StatusUnauthorized)
	})
}

// TestProtectedEndpointWithValidToken ensures that a valid token allows access past the middleware.
func TestProtectedEndpointWithValidToken(t *testing.T) {
	token := generateValidToken()
	r, _ := http.NewRequest("GET", "/restaurante/v1/clientes", nil)
	r.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	Convey("GET /restaurante/v1/clientes with valid token should not be unauthorized", t, func() {
		So(w.Code, ShouldNotEqual, http.StatusUnauthorized)
	})
}
