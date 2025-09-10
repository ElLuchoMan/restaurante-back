package categoria

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beego/beego/v2/server/web/context"
)

func TestCategoria_Put_BadID(t *testing.T) {
	orig := catOrmNew
	catOrmNew = func() categoriaOrmer { return fakeCatOrm{} }
	t.Cleanup(func() { catOrmNew = orig })

	r := httptest.NewRequest(http.MethodPut, "/categorias", strings.NewReader(`{"nombre":"X"}`))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(`{"nombre":"X"}`)
	c := &CategoriaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Put()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
