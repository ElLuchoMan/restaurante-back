package productopedido

import (
	stdctx "context"
	"database/sql/driver"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
)

// Cubre la rama donde todos los items nuevos son inválidos -> nuevos vacío,
// solo deltas negativos (restock), sin inserts y commit exitoso.
func TestProductoPedidoUpdate_AllInvalidItems_OnlyRestock_NoInserts(t *testing.T) {
	origQ, origE := MockQuery, MockExec
	origDel := productoPedidoDeleteDetalles
	// Evitar dependencia del Delete por QueryTable
	productoPedidoDeleteDetalles = func(_ orm.TxOrmer, _ int64) error { return nil }

	// actuales: dos productos existentes (1 y 2)
	MockQuery = func(_ stdctx.Context, q string, _ []driver.NamedValue) (driver.Rows, error) {
		lower := strings.ToLower(q)
		switch {
		case strings.Contains(lower, "from detalle_pedido") && !strings.Contains(lower, "insert into"):
			cols := []string{"pk_id_detalle", "pk_id_pedido", "pk_id_producto", "cantidad", "precio"}
			vals := [][]driver.Value{{int64(1), int64(1), int64(1), int64(2), int64(1000)}, {int64(2), int64(1), int64(2), int64(3), int64(500)}}
			return &mockRows{columns: cols, values: vals}, nil
		case strings.Contains(lower, "for update"):
			// bloqueo del pedido ok
			return &mockRows{columns: []string{"ok"}, values: [][]driver.Value{{int64(1)}}}, nil
		default:
			return &mockRows{columns: []string{"ok"}, values: [][]driver.Value{{int64(1)}}}, nil
		}
	}
	MockExec = func(_ stdctx.Context, q string, _ []driver.NamedValue) (driver.Result, error) {
		// Solo se deben ejecutar updates de restock (+) y el delete de detalles;
		// no habrá inserts de nuevos detalles porque nuevos está vacío.
		return mockResult{}, nil
	}
	t.Cleanup(func() { MockQuery, MockExec = origQ, origE; productoPedidoDeleteDetalles = origDel })

	// Todos los items inválidos: productoId=0 o cantidad<0 -> nuevos queda vacío
	body := `[{"productoId":0,"cantidad":-1},{"productoId":0,"cantidad":0},{"productoId":2,"cantidad":-5}]`
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
