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

func TestProductoPedidoUpdate_Happy_Commit_Success_Cover(t *testing.T) {
	origQ, origE := MockQuery, MockExec
	origDel := productoPedidoDeleteDetalles
	origReq := productoPedidoRequeryDetalle
	origCommit := productoPedidoCommit
	origQueryAct := productoPedidoQueryActualesFn

	productoPedidoQueryActualesFn = func(_ orm.Ormer, _ int64) ([]models.DetallePedido, error) {
		return []models.DetallePedido{{
			PKIDPedido:   &models.Pedido{PK_ID_PEDIDO: 1},
			PKIDProducto: &models.Producto{PK_ID_PRODUCTO: 1},
			Cantidad:     2,
			Precio:       1000,
		}}, nil
	}

	MockQuery = func(_ stdctx.Context, _ string, _ []driver.NamedValue) (driver.Rows, error) {
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
	productoPedidoCommit = func(tx orm.TxOrmer) error { return nil }

	t.Cleanup(func() {
		MockQuery, MockExec = origQ, origE
		productoPedidoDeleteDetalles = origDel
		productoPedidoRequeryDetalle = origReq
		productoPedidoCommit = origCommit
		productoPedidoQueryActualesFn = origQueryAct
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
	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", resp.Code, w.Body.String())
	}
}
