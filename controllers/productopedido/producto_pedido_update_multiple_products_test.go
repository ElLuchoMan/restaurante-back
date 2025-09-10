package productopedido

import (
	stdctx "context"
	"database/sql/driver"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beego/beego/v2/server/web/context"
)

// Cubre ordenamiento y múltiples productos en Update
func TestProductoPedidoUpdate_MultipleProducts(t *testing.T) {
	origQ, origE := MockQuery, MockExec
	step := 0
	insertStep := 0
	requeryStep := 0
	MockQuery = func(_ stdctx.Context, q string, _ []driver.NamedValue) (driver.Rows, error) {
		lower := strings.ToLower(q)
		switch {
		case strings.Contains(lower, "detalle_pedido"):
			if step == 0 {
				step++
				cols := []string{"pk_id_detalle", "pk_id_pedido", "pk_id_producto", "cantidad", "precio"}
				vals := [][]driver.Value{{int64(1), int64(1), int64(1), int64(2), int64(1000)}, {int64(2), int64(1), int64(2), int64(1), int64(500)}}
				return &mockRows{columns: cols, values: vals}, nil
			}
			requeryStep++
			cols := []string{"pk_id_detalle", "pk_id_pedido", "pk_id_producto", "cantidad", "precio"}
			if requeryStep == 1 {
				vals := [][]driver.Value{{int64(3), int64(1), int64(1), int64(2), int64(1000)}}
				return &mockRows{columns: cols, values: vals}, nil
			}
			vals := [][]driver.Value{{int64(4), int64(1), int64(2), int64(3), int64(500)}}
			return &mockRows{columns: cols, values: vals}, nil
		case strings.Contains(lower, "insert into") && strings.Contains(lower, "detalle_pedido"):
			insertStep++
			return &mockRows{columns: []string{"pk_id_detalle"}, values: [][]driver.Value{{int64(insertStep)}}}, nil
		case strings.Contains(lower, "select pk_id_producto, cantidad from producto"):
			cols := []string{"pk_id_producto", "cantidad"}
			vals := [][]driver.Value{{int64(2), int64(10)}}
			return &mockRows{columns: cols, values: vals}, nil
		default:
			return &mockRows{columns: []string{"ok"}, values: [][]driver.Value{{int64(1)}}}, nil
		}
	}
	MockExec = func(_ stdctx.Context, _ string, _ []driver.NamedValue) (driver.Result, error) {
		return mockResult{}, nil
	}
	t.Cleanup(func() { MockQuery, MockExec = origQ, origE })

	body := `[{"productoId":1,"cantidad":2},{"productoId":2,"cantidad":3}]`
	r := httptest.NewRequest(http.MethodPut, "/producto_pedido?pedido_id=1", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := ProductoPedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Update()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}
}
