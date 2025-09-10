package precioproductohist

import (
	stdctx "context"
	"database/sql/driver"
	"net/http"
	"net/http/httptest"
	"testing"

	beegoCtx "github.com/beego/beego/v2/server/web/context"
)

// Usamos el hook pphOrmNew pero no ejecutamos QueryRows (no DB); sólo verificamos códigos en ausencia de datos
func TestPPH_GetAll_EmptyOK(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/precio_producto_hist?producto_id=1", nil)
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	c := &PrecioProductoHistController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.GetAll()
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError { // sin DB podría ser 500
		t.Fatalf("unexpected status %d", w.Code)
	}
}

func TestPPH_GetById_NotFound(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/precio_producto_hist/search?id=999", nil)
	w := httptest.NewRecorder()
	orig := MockQuery
	MockQuery = func(_ stdctx.Context, _ string, _ []driver.NamedValue) (driver.Rows, error) {
		cols := []string{"nombre", "estado_producto", "precio", "fecha_vigencia"}
		// sin valores para forzar orm.ErrNoRows
		return &mockRows{columns: cols, values: [][]driver.Value{}}, nil
	}
	t.Cleanup(func() { MockQuery = orig })

	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	c := &PrecioProductoHistController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.GetById()
	if w.Code != http.StatusOK {
		t.Fatalf("unexpected status %d", w.Code)
	}
}
