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

func TestProductoPedidoPost_Success_Transaction(t *testing.T) {
    origQ, origE := MockQuery, MockExec
    step := 0
    MockQuery = func(_ stdctx.Context, q string, _ []driver.NamedValue) (driver.Rows, error) {
        lower := strings.ToLower(q)
        if strings.Contains(lower, "select pk_id_producto, cantidad from producto") {
            cols := []string{"pk_id_producto", "cantidad"}
            vals := [][]driver.Value{{int64(1), int64(10)}}
            return &mockRows{columns: cols, values: vals}, nil
        }
        if strings.Contains(lower, "insert into") && strings.Contains(lower, "detalle_pedido") {
            // retorno del INSERT ... RETURNING pk_id_detalle
            return &mockRows{columns: []string{"pk_id_detalle"}, values: [][]driver.Value{{int64(1)}}}, nil
        }
        if strings.Contains(lower, "detalle_pedido") {
            // reconsulta después del insert para obtener precio definitivo
            cols := []string{"pk_id_detalle", "pk_id_pedido", "pk_id_producto", "cantidad", "precio"}
            vals := [][]driver.Value{{int64(1), int64(1), int64(1), int64(2), int64(1000)}}
            return &mockRows{columns: cols, values: vals}, nil
        }
        // otras consultas (BEGIN/COMMIT/UPDATE) se satisfacen vía MockExec
        step++
        return &mockRows{columns: []string{"ok"}, values: [][]driver.Value{{int64(step)}}}, nil
    }
    MockExec = func(_ stdctx.Context, _ string, _ []driver.NamedValue) (driver.Result, error) { return mockResult{}, nil }
    t.Cleanup(func(){ MockQuery, MockExec = origQ, origE })

    body := `{"pedidoId":1,"detalles":[{"productoId":1,"cantidad":2}]}`
    r := httptest.NewRequest(http.MethodPost, "/producto_pedido", strings.NewReader(body))
    w := httptest.NewRecorder(); ctx := context.NewContext(); ctx.Reset(w, r); ctx.Input.RequestBody = []byte(body)
    c := ProductoPedidoController{}; c.Ctx = ctx; c.Data = make(map[interface{}]interface{})

    c.Post()
    if w.Code != http.StatusCreated && w.Code != http.StatusOK { t.Fatalf("unexpected status %d. Body: %s", w.Code, w.Body.String()) }
}


