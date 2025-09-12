package productopedido

import (
	stdctx "context"
	"database/sql/driver"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
)

func TestProductoPedidoUpdate_Mixed_WithZeroDelta(t *testing.T) {
	origQ, origE := MockQuery, MockExec
	origDel := productoPedidoDeleteDetalles
	origReq := productoPedidoRequeryDetalle

	productoPedidoDeleteDetalles = func(_ orm.TxOrmer, _ int64) error { return nil }
	productoPedidoRequeryDetalle = func(_ orm.TxOrmer, pedidoID int64, productoID int64, out *models.DetallePedido) error {
		*out = models.DetallePedido{PKIDPedido: &models.Pedido{PK_ID_PEDIDO: pedidoID}, PKIDProducto: &models.Producto{PK_ID_PRODUCTO: productoID}, Cantidad: 1, Precio: 1000}
		return nil
	}

	step := 0
	requeryStep := 0
	MockQuery = func(_ stdctx.Context, q string, _ []driver.NamedValue) (driver.Rows, error) {
		lower := strings.ToLower(q)
		switch {
		case strings.Contains(lower, "detalle_pedido") && !strings.Contains(lower, "insert into"):
			if step == 0 {
				step++
				cols := []string{"pk_id_detalle", "pk_id_pedido", "pk_id_producto", "cantidad", "precio"}
				vals := [][]driver.Value{{int64(1), int64(1), int64(1), int64(2), int64(1000)}, {int64(2), int64(1), int64(2), int64(1), int64(800)}, {int64(3), int64(1), int64(3), int64(4), int64(500)}}
				return &mockRows{columns: cols, values: vals}, nil
			}
			requeryStep++
			cols := []string{"pk_id_detalle", "pk_id_pedido", "pk_id_producto", "cantidad", "precio"}
			vals := [][]driver.Value{{int64(10 + requeryStep), int64(1), int64(2), int64(3), int64(900)}}
			return &mockRows{columns: cols, values: vals}, nil
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
	t.Cleanup(func() {
		MockQuery, MockExec = origQ, origE
		productoPedidoDeleteDetalles = origDel
		productoPedidoRequeryDetalle = origReq
	})

	body := `[{"productoId":1,"cantidad":2},{"productoId":2,"cantidad":3},{"productoId":3,"cantidad":3}]`
	r := httptest.NewRequest(http.MethodPut, "/producto_pedido?pedido_id=1", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)

	c := &ProductoPedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Update()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}
}
