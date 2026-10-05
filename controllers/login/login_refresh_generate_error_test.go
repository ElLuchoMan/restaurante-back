package login

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"restaurante/models"

	"github.com/beego/beego/v2/server/web/context"
	"github.com/golang-jwt/jwt/v5"
)

func TestRefreshToken_GenerateTokensError(t *testing.T) {
	origSecret := jwtSecret
	origMethod := signingMethod
	jwtSecret = []byte("secret-refresh-error")
	t.Cleanup(func() {
		jwtSecret = origSecret
		signingMethod = origMethod
	})

	claims := &RefreshClaims{
		Documento: 99,
		Rol:       "CLIENTE",
		Nombre:    "Tester",
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwtSecret)
	if err != nil {
		t.Fatal(err)
	}

	// El token se firma antes con HS256 real; al renovar, la firma falla.
	signingMethod = badSignMethod{}

	r := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	r.Header.Set("Authorization", "Bearer "+tokenString)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := &LoginController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.RefreshToken()

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("esperado 500, obtenido %d", w.Code)
	}
	var resp models.ApiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Message != "Error al generar nuevos tokens" || resp.Cause == "" {
		t.Fatalf("respuesta inesperada: %+v", resp)
	}
}
