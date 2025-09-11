package productopedido

import (
    stdctx "context"
    "database/sql/driver"
    "encoding/json"
    "errors"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"

    "restaurante/models"

    "github.com/beego/beego/v2/client/orm"
    "github.com/beego/beego/v2/server/web/context"
)

// Cubre la rama de error al restaurar inventario cuando delta < 0
func TestProductoPedidoUpdate_StockRestoreError(t *testing.T) {
    origQ, origE := MockQuery, MockExec
    origDel := productoPedidoDeleteDetalles
    origReq := productoPedidoRequeryDetalle

    MockQuery = func(_ stdctx.Context, q string, _ []driver.NamedValue) (driver.Rows, error) {
        lower := strings.ToLower(q)
        if strings.Contains(lower, "detalle_pedido") {
            // existentes con cantidad 3 para delta negativo
            cols := []string{"pk_id_detalle", "pk_id_pedido", "pk_id_producto", "cantidad", "precio"}
            vals := [][]driver.Value{{int64(1), int64(1), int64(1), int64(3), int64(1000)}}
            return &mockRows{columns: cols, values: vals}, nil
        }
        return &mockRows{columns: []string{"ok"}, values: [][]driver.Value{{int64(1)}}}, nil
    }
    MockExec = func(_ stdctx.Context, q string, _ []driver.NamedValue) (driver.Result, error) {
        if strings.Contains(strings.ToLower(q), "update producto set cantidad = cantidad +") {
            return nil, errors.New("restore fail")
        }
        return mockResult{}, nil
    }
    productoPedidoDeleteDetalles = func(_ orm.TxOrmer, _ int64) error { return nil }
    productoPedidoRequeryDetalle = func(_ orm.TxOrmer, _ int64, _ int64, _ *models.DetallePedido) error { return nil }

    t.Cleanup(func() {
        MockQuery, MockExec = origQ, origE
        productoPedidoDeleteDetalles = origDel
        productoPedidoRequeryDetalle = origReq
    })

    body := `[{"productoId":1,"cantidad":1}]` // antes 3 -> delta -2
    r := httptest.NewRequest(http.MethodPut, "/producto_pedido?pedido_id=1", strings.NewReader(body))
    w := httptest.NewRecorder()
    ctx := context.NewContext()
    ctx.Reset(w, r)
    ctx.Input.RequestBody = []byte(body)

    c := &ProductoPedidoController{}
    c.Ctx = ctx
    c.Data = make(map[interface{}]interface{})

    c.Update()
    if w.Code != http.StatusInternalServerError {
        t.Fatalf("expected 500, got %d. Body: %s", w.Code, w.Body.String())
    }
    var resp models.ApiResponse
    if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
        t.Fatalf("invalid response: %v", err)
    }
    if resp.Message != "Error al ajustar inventario" {
        t.Fatalf("unexpected message: %s", resp.Message)
    }
}

