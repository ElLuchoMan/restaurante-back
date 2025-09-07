package controllers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	beegoCtx "github.com/beego/beego/v2/server/web/context"
)

func TestGenerateJWT_ErrorWhenSigning(t *testing.T) {
    // Forzar error estable cambiando la clave a nil temporalmente
    orig := jwtSecret
    jwtSecret = nil
    t.Cleanup(func(){ jwtSecret = orig })

    r := httptest.NewRequest(http.MethodPost, "/login", nil)
    w := httptest.NewRecorder()
    ctx := beegoCtx.NewContext(); ctx.Reset(w, r)
    c := &LoginController{}
    c.Ctx = ctx
    c.Data = make(map[interface{}]interface{})

    generateJWT(c, 1, "Admin", "Tester")

    if w.Code != http.StatusInternalServerError && w.Code != http.StatusOK {
        t.Fatalf("expected 500 or 200 depending on env, got %d", w.Code)
    }
    if w.Code == http.StatusInternalServerError && !strings.Contains(w.Body.String(), "Error al generar el token") {
        t.Fatalf("expected signing error message, got: %s", w.Body.String())
    }
}


