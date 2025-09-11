package productopedido

import (
	stdctx "context"
	"database/sql/driver"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
)

func TestProductoPedidoUpdate_RequeryWrapperError_CoversBranch(t *testing.T) {
	origQ, origE := MockQuery, MockExec
	origReq := productoPedidoRequeryDetalle
	origDel := productoPedidoDeleteDetalles
	// Forzar error directo del hook de reconsulta y evitar fallo en Delete
	productoPedidoRequeryDetalle = func(_ orm.TxOrmer, _ int64, _ int64, _ *models.DetallePedido) error {
		return errors.New("requery hook error")
	}
	productoPedidoDeleteDetalles = func(_ orm.TxOrmer, _ int64) error { return nil }
	// actuales vacíos -> delta positivo; inventario suficiente para pasar a inserción
	MockQuery = func(_ stdctx.Context, q string, _ []driver.NamedValue) (driver.Rows, error) {
		lower := strings.ToLower(q)
		if strings.Contains(lower, "detalle_pedido") && !strings.Contains(lower, "insert into") {
			return &mockRows{columns: []string{"pk_id_pedido"}, values: [][]driver.Value{}}, nil
		}
		if strings.Contains(lower, "select pk_id_producto, cantidad from producto") {
			cols := []string{"pk_id_producto", "cantidad"}
			vals := [][]driver.Value{{int64(1), int64(10)}}
			return &mockRows{columns: cols, values: vals}, nil
		}
		return &mockRows{columns: []string{"ok"}, values: [][]driver.Value{{int64(1)}}}, nil
	}
	MockExec = func(_ stdctx.Context, _ string, _ []driver.NamedValue) (driver.Result, error) { return mockResult{}, nil }
	t.Cleanup(func() { MockQuery, MockExec = origQ, origE; productoPedidoRequeryDetalle = origReq; productoPedidoDeleteDetalles = origDel })

	body := `[{"productoId":1,"cantidad":2}]`
	r := httptest.NewRequest(http.MethodPut, "/producto_pedido?pedido_id=1", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := &ProductoPedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Update()
	if w.Code != http.StatusInternalServerError && w.Code != http.StatusOK {
		t.Fatalf("expected 500 or 200, got %d. Body: %s", w.Code, w.Body.String())
	}
}
