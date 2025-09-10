package categoria

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"restaurante/models"

	"github.com/beego/beego/v2/server/web/context"
)

func TestCategoriaController_Put_InvalidJSON(t *testing.T) {
	// Mock ormer to avoid hitting real database and ensure Read succeeds
	m := newCatMockOrm()
	// Insert a dummy category with ID=1 so that Read returns nil error
	_, _ = m.Insert(&models.Categoria{NOMBRE: "test"})
	orig := catOrmNew
	catOrmNew = func() categoriaOrmer { return m }
	defer func() { catOrmNew = orig }()

	r := httptest.NewRequest(http.MethodPut, "/categorias?id=1", strings.NewReader("{"))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte("{")
	c := &CategoriaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Put()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
