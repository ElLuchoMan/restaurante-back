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

// Cubre rama de delta negativo (devolver inventario) y reemplazo completo de detalles
func TestProductoPedidoUpdate_NegativeDelta_Success(t *testing.T) {
	// Mock consultas: actuales trae cantidad previa, luego delete ok, luego one reconsulta
	call := 0
	origQ, origE := MockQuery, MockExec
	MockQuery = func(_ stdctx.Context, q string, _ []driver.NamedValue) (driver.Rows, error) {
		lower := strings.ToLower(q)
		if strings.Contains(lower, "detalle_pedido") {
			if call == 0 {
				// consulta inicial: cantidad previa para calcular delta negativo
				cols := []string{"pk_id_detalle", "pk_id_pedido", "pk_id_producto", "cantidad", "precio"}
				vals := [][]driver.Value{{int64(1), int64(1), int64(1), int64(3), int64(1000)}}
				call++
				return &mockRows{columns: cols, values: vals}, nil
			}
			// reconsulta final después del insert
			cols := []string{"pk_id_detalle", "pk_id_pedido", "pk_id_producto", "cantidad", "precio"}
			vals := [][]driver.Value{{int64(2), int64(1), int64(1), int64(1), int64(1000)}}
			return &mockRows{columns: cols, values: vals}, nil
		}
		// otras consultas (BEGIN/COMMIT/UPDATE)
		return &mockRows{columns: []string{"ok"}, values: [][]driver.Value{{int64(1)}}}, nil
	}
	MockExec = func(_ stdctx.Context, q string, _ []driver.NamedValue) (driver.Result, error) {
		// UPDATE producto + DELETE detalles + INSERT detalles
		return mockResult{}, nil
	}
	t.Cleanup(func() { MockQuery, MockExec = origQ, origE })

	body := `[{"productoId":1,"cantidad":1}]` // antes 3 -> delta -2
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
}
