package controllers

import (
	"crypto"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	beegoCtx "github.com/beego/beego/v2/server/web/context"
	"github.com/dgrijalva/jwt-go"
)

func TestGenerateJWT_ErrorWhenSigning(t *testing.T) {
	origSecret := jwtSecret
	jwtSecret = []byte("testsecret")
	t.Cleanup(func() { jwtSecret = origSecret })

	origHash := jwt.SigningMethodHS256.Hash
	jwt.SigningMethodHS256.Hash = crypto.Hash(0)
	t.Cleanup(func() { jwt.SigningMethodHS256.Hash = origHash })

	r := httptest.NewRequest(http.MethodPost, "/login", nil)
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	c := &LoginController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	generateJWT(c, 1, "Admin", "Tester")

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al generar el token") {
		t.Fatalf("expected signing error message, got: %s", w.Body.String())
	}
}
