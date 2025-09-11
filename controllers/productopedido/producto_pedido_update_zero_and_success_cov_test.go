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

// Cubre: delta==0 (continue), sort de deltaIDs y newIDs con >=2 elementos,
// reconsulta de detalles, commit exitoso y respuesta final 200.
func TestProductoPedidoUpdate_ZeroDelta_And_SuccessMultipleCov(t *testing.T) {
	origQ, origE := MockQuery, MockExec
	origDel := productoPedidoDeleteDetalles
	origReq := productoPedidoRequeryDetalle
	// evitar problemas del ORM mock en Delete y en reconsulta
	productoPedidoDeleteDetalles = func(_ orm.TxOrmer, _ int64) error { return nil }
	productoPedidoRequeryDetalle = func(_ orm.TxOrmer, pedidoID int64, productoID int64, out *models.DetallePedido) error {
		*out = models.DetallePedido{
			PKIDPedido:   &models.Pedido{PK_ID_PEDIDO: pedidoID},
			PKIDProducto: &models.Producto{PK_ID_PRODUCTO: productoID},
			Cantidad:     1,
			Precio:       1000,
		}
		return nil
	}
	phase := 0
	MockQuery = func(_ stdctx.Context, q string, _ []driver.NamedValue) (driver.Rows, error) {
		lower := strings.ToLower(q)
		// Consulta de detalles actuales (cubre QueryTable.All)
		if strings.Contains(lower, "detalle_pedido") && !strings.Contains(lower, "insert into") {
			if phase == 0 {
				// actuales: p1=1, p2=2 -> con nuevos p1=3 (delta+2), p2=2 (delta 0)
				phase = 1
				cols := []string{"pk_id_detalle", "pk_id_pedido", "pk_id_producto", "cantidad", "precio"}
				vals := [][]driver.Value{{int64(1), int64(1), int64(1), int64(1), int64(1000)}, {int64(2), int64(1), int64(2), int64(2), int64(500)}}
				return &mockRows{columns: cols, values: vals}, nil
			}
		}
		// Validación de inventario: solo p1 está en need
		if strings.Contains(lower, "select pk_id_producto, cantidad from producto") {
			cols := []string{"pk_id_producto", "cantidad"}
			vals := [][]driver.Value{{int64(1), int64(10)}}
			return &mockRows{columns: cols, values: vals}, nil
		}
		// Bloqueo FOR UPDATE y otras consultas auxiliares
		return &mockRows{columns: []string{"ok"}, values: [][]driver.Value{{int64(1)}}}, nil
	}
	MockExec = func(_ stdctx.Context, _ string, _ []driver.NamedValue) (driver.Result, error) {
		// UPDATE/DELETE/INSERT/COMMIT/BEGIN -> ok
		return mockResult{}, nil
	}
	t.Cleanup(func() {
		MockQuery, MockExec = origQ, origE
		productoPedidoDeleteDetalles = origDel
		productoPedidoRequeryDetalle = origReq
	})

	body := `[{"productoId":1,"cantidad":3},{"productoId":2,"cantidad":2}]`
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
	if !strings.Contains(strings.ToLower(w.Body.String()), "actualizados exitosamente") {
		t.Fatalf("expected success message, body: %s", w.Body.String())
	}
}
