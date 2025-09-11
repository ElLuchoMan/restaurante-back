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

// Camino feliz completo: stock suficiente, sin errores, commit OK y respuesta 200
func TestProductoPedidoUpdate_HappyPath_Success(t *testing.T) {
	origQ, origE := MockQuery, MockExec
	origDel := productoPedidoDeleteDetalles
	origReq := productoPedidoRequeryDetalle
	step := 0
	requeryStep := 0
	MockQuery = func(_ stdctx.Context, q string, _ []driver.NamedValue) (driver.Rows, error) {
		lower := strings.ToLower(q)
		switch {
		case strings.Contains(lower, "detalle_pedido") && !strings.Contains(lower, "insert into"):
			// actuales: vacío -> deltas positivos
			if step == 0 {
				step++
				return &mockRows{columns: []string{"pk_id_pedido"}, values: [][]driver.Value{}}, nil
			}
			// reconsultas después de inserts
			requeryStep++
			cols := []string{"pk_id_detalle", "pk_id_pedido", "pk_id_producto", "cantidad", "precio"}
			if requeryStep == 1 {
				vals := [][]driver.Value{{int64(101), int64(1), int64(1), int64(2), int64(1200)}}
				return &mockRows{columns: cols, values: vals}, nil
			}
			vals := [][]driver.Value{{int64(102), int64(1), int64(2), int64(3), int64(800)}}
			return &mockRows{columns: cols, values: vals}, nil
		case strings.Contains(lower, "select pk_id_producto, cantidad from producto"):
			// stock suficiente para p1 y p2
			cols := []string{"pk_id_producto", "cantidad"}
			vals := [][]driver.Value{{int64(1), int64(10)}, {int64(2), int64(10)}}
			return &mockRows{columns: cols, values: vals}, nil
		default:
			return &mockRows{columns: []string{"ok"}, values: [][]driver.Value{{int64(1)}}}, nil
		}
	}
	MockExec = func(_ stdctx.Context, _ string, _ []driver.NamedValue) (driver.Result, error) {
		return mockResult{}, nil
	}
	// Evitar fallos por Delete y reconsulta One
	productoPedidoDeleteDetalles = func(_ orm.TxOrmer, _ int64) error { return nil }
	productoPedidoRequeryDetalle = func(_ orm.TxOrmer, pedidoID int64, productoID int64, out *models.DetallePedido) error {
		*out = models.DetallePedido{PKIDPedido: &models.Pedido{PK_ID_PEDIDO: pedidoID}, PKIDProducto: &models.Producto{PK_ID_PRODUCTO: productoID}, Cantidad: 1, Precio: 1000}
		return nil
	}
	t.Cleanup(func() {
		MockQuery, MockExec = origQ, origE
		productoPedidoDeleteDetalles = origDel
		productoPedidoRequeryDetalle = origReq
	})

	body := `[{"productoId":1,"cantidad":2},{"productoId":2,"cantidad":3}]`
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
	if !strings.Contains(w.Body.String(), "Productos del pedido actualizados exitosamente") {
		t.Fatalf("expected success message, body: %s", w.Body.String())
	}
}
