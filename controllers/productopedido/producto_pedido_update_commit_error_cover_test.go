package productopedido

import (
	stdctx "context"
	"database/sql/driver"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
)

func TestProductoPedidoUpdate_CommitError_Cover(t *testing.T) {
	origQ, origE := MockQuery, MockExec
	origDel := productoPedidoDeleteDetalles
	origReq := productoPedidoRequeryDetalle
	origBegin := productoPedidoBeginTx

	MockQuery = func(_ stdctx.Context, q string, _ []driver.NamedValue) (driver.Rows, error) {
		lower := strings.ToLower(q)
		if strings.Contains(lower, "from detalle_pedido") && !strings.Contains(lower, "insert into") {
			cols := []string{"pk_id_detalle", "pk_id_pedido", "pk_id_producto", "cantidad", "precio"}
			vals := [][]driver.Value{{int64(1), int64(1), int64(1), int64(2), int64(1000)}}
			return &mockRows{columns: cols, values: vals}, nil
		}
		return &mockRows{columns: []string{"ok"}, values: [][]driver.Value{{int64(1)}}}, nil
	}
	MockExec = func(_ stdctx.Context, _ string, _ []driver.NamedValue) (driver.Result, error) {
		return mockResult{}, nil
	}
	productoPedidoDeleteDetalles = func(_ orm.TxOrmer, _ int64) error { return nil }
	productoPedidoRequeryDetalle = func(_ orm.TxOrmer, pedidoID int64, productoID int64, out *models.DetallePedido) error {
		*out = models.DetallePedido{PKIDPedido: &models.Pedido{PK_ID_PEDIDO: pedidoID}, PKIDProducto: &models.Producto{PK_ID_PRODUCTO: productoID}, Cantidad: 2, Precio: 1000}
		return nil
	}
	productoPedidoBeginTx = func(o orm.Ormer) (orm.TxOrmer, error) {
		baseTx, err := o.Begin()
		if err != nil {
			return nil, err
		}
		return commitFailTx{TxOrmer: baseTx}, nil
	}

	t.Cleanup(func() {
		MockQuery, MockExec = origQ, origE
		productoPedidoDeleteDetalles = origDel
		productoPedidoRequeryDetalle = origReq
		productoPedidoBeginTx = origBegin
	})

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
	var resp models.ApiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid response: %v", err)
	}
	if resp.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d. Body: %s", resp.Code, w.Body.String())
	}
}
