package subcategoria

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beego/beego/v2/server/web/context"
	"restaurante/models"
)

// Cubre la rama de JSON inválido en el método Put
func TestSubcategoriaController_Put_BadJSON(t *testing.T) {
	m := newSubMockOrm()
	m.Insert(&models.Subcategoria{NOMBRE: "X"})
	orig := subcatOrmNew
	subcatOrmNew = func() subcatOrmer { return m }
	t.Cleanup(func() { subcatOrmNew = orig })

	r := httptest.NewRequest(http.MethodPut, "/subcategorias?id=1", strings.NewReader("bad"))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte("bad")
	c := &SubcategoriaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Put()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
