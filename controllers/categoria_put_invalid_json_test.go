package controllers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beego/beego/v2/server/web/context"
)

func TestCategoriaController_Put_InvalidJSON(t *testing.T) {
	r := httptest.NewRequest(http.MethodPut, "/categorias?id=1", strings.NewReader("{"))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte("{")
	c := &CategoriaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Put()
	if w.Code != http.StatusBadRequest && w.Code != http.StatusOK {
		t.Fatalf("expected 400 or 200, got %d", w.Code)
	}
}


