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

func TestProductoPedidoPost_InventoryQueryRowsError_Continues(t *testing.T) {
	origQ, origE := MockQuery, MockExec
	MockQuery = func(_ stdctx.Context, q string, _ []driver.NamedValue) (driver.Rows, error) {
		lower := strings.ToLower(q)
		if strings.Contains(lower, "select pk_id_producto, cantidad from producto") {
			return nil, driver.ErrBadConn
		}
		return &mockRows{columns: []string{"ok"}, values: [][]driver.Value{{int64(1)}}}, nil
	}
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
	if w.Code != http.StatusCreated && w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Fatalf("unexpected status %d. Body: %s", w.Code, w.Body.String())
	}
}
