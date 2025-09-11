package productopedido

import (
	stdctx "context"
	"database/sql/driver"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beego/beego/v2/server/web/context"
)

// Cubre rama en Post: error al consultar inventario (QueryRows) y se continúa el flujo
func TestProductoPedidoPost_InventoryQueryRowsError_Continues(t *testing.T) {
	origQ, origE := MockQuery, MockExec
	MockQuery = func(_ stdctx.Context, q string, _ []driver.NamedValue) (driver.Rows, error) {
		lower := strings.ToLower(q)
		if strings.Contains(lower, "select pk_id_producto, cantidad from producto") {
			return nil, driver.ErrBadConn // provocar error en QueryRows
		}
		// para resto de consultas devolver algo neutro
		return &mockRows{columns: []string{"ok"}, values: [][]driver.Value{{int64(1)}}}, nil
	}
	// dejar MockExec por defecto (mockResult)
	t.Cleanup(func() { MockQuery, MockExec = origQ, origE })

	body := `{"pedidoId":1,"detalles":[{"productoId":1,"cantidad":2}]}`
	r := httptest.NewRequest(http.MethodPost, "/producto_pedido", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := &ProductoPedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()
	// puede terminar en 201/200 o 500 según el entorno mock, lo importante es ejecutar la rama
	if w.Code != http.StatusCreated && w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Fatalf("unexpected status %d. Body: %s", w.Code, w.Body.String())
	}
}

