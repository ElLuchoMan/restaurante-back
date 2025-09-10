package productopedido

import (
	stdctx "context"
	"database/sql/driver"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beego/beego/v2/server/web/context"
)

// Cubre la rama de error al insertar el detalle en Update
func TestProductoPedidoUpdate_InsertDetalleError(t *testing.T) {
	origQ, origE := MockQuery, MockExec
	step := 0
	MockQuery = func(_ stdctx.Context, q string, _ []driver.NamedValue) (driver.Rows, error) {
		lower := strings.ToLower(q)
		switch {
		case strings.Contains(lower, "select pk_id_producto, cantidad from producto"):
			cols := []string{"pk_id_producto", "cantidad"}
			vals := [][]driver.Value{{int64(1), int64(10)}}
			return &mockRows{columns: cols, values: vals}, nil
		case strings.Contains(lower, "insert into") && strings.Contains(lower, "detalle_pedido"):
			return nil, errors.New("insert fail")
		case strings.Contains(lower, "from detalle_pedido"):
			if step == 0 {
				step++
				return &mockRows{columns: []string{"pk_id_pedido"}, values: [][]driver.Value{}}, nil
			}
			cols := []string{"pk_id_detalle", "pk_id_pedido", "pk_id_producto", "cantidad", "precio"}
			return &mockRows{columns: cols, values: [][]driver.Value{}}, nil
		default:
			return &mockRows{columns: []string{"ok"}, values: [][]driver.Value{{int64(1)}}}, nil
		}
	}
	MockExec = func(_ stdctx.Context, _ string, _ []driver.NamedValue) (driver.Result, error) {
		return mockResult{}, nil
	}
	t.Cleanup(func() { MockQuery, MockExec = origQ, origE })

	body := `[{"productoId":1,"cantidad":2}]`
	r := httptest.NewRequest(http.MethodPut, "/producto_pedido?pedido_id=1", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := ProductoPedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Update()
	resp := w.Body.String()
	if !(strings.Contains(resp, "Error al actualizar los productos del pedido") || strings.Contains(resp, "Error al buscar los detalles del pedido")) {
		t.Fatalf("expected insert detalle or pre-query error, body: %s", resp)
	}
}
