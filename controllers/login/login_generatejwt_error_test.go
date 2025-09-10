package login

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/beego/beego/v2/server/web/context"
	"github.com/golang-jwt/jwt/v5"
)

type badSignMethod struct{}

func (badSignMethod) Alg() string                              { return "HS256" }
func (badSignMethod) Verify(string, []byte, interface{}) error { return nil }
func (badSignMethod) Sign(string, interface{}) (string, error) { return "", errors.New("sign error") }

func TestGenerateJWT_SigningFailure(t *testing.T) {
	origSecret := jwtSecret
	jwtSecret = []byte("secret")
	t.Cleanup(func() { jwtSecret = origSecret })

	origMethod := jwt.SigningMethodHS256
	jwt.SigningMethodHS256 = badSignMethod{}
	t.Cleanup(func() { jwt.SigningMethodHS256 = origMethod })

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/login", nil)
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := &LoginController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	generateJWT(c, 1, "Admin", "Nombre")
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}
