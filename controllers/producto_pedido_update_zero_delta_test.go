package controllers

import (
	stdctx "context"
	"database/sql/driver"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beego/beego/v2/server/web/context"
)

// Cubre rama delta == 0 (mismo total por producto) -> sin validación de stock; reemplazo de detalles
func TestProductoPedidoUpdate_ZeroDelta_ReplacesDetails(t *testing.T) {
    origQ, origE := MockQuery, MockExec
    step := 0
    MockQuery = func(_ stdctx.Context, q string, _ []driver.NamedValue) (driver.Rows, error) {
        lower := strings.ToLower(q)
        if strings.Contains(lower, "detalle_pedido") {
            step++
            // tanto la consulta inicial como la reconsulta final devuelven el mismo registro
            cols := []string{"pk_id_detalle", "pk_id_pedido", "pk_id_producto", "cantidad", "precio"}
            vals := [][]driver.Value{{int64(step), int64(1), int64(1), int64(2), int64(1000)}}
            return &mockRows{columns: cols, values: vals}, nil
        }
        return &mockRows{columns: []string{"ok"}, values: [][]driver.Value{{int64(1)}}}, nil
    }
    MockExec = func(_ stdctx.Context, _ string, _ []driver.NamedValue) (driver.Result, error) { return mockResult{}, nil }
    t.Cleanup(func(){ MockQuery, MockExec = origQ, origE })

    body := `[{"productoId":1,"cantidad":2}]`
    r := httptest.NewRequest(http.MethodPut, "/producto_pedido?pedido_id=1", strings.NewReader(body))
    w := httptest.NewRecorder(); ctx := context.NewContext(); ctx.Reset(w, r); ctx.Input.RequestBody = []byte(body)
    c := ProductoPedidoController{}; c.Ctx = ctx; c.Data = make(map[interface{}]interface{})

    c.Update()
    if w.Code != http.StatusOK { t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String()) }
}


