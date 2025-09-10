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

// Ensure that Post handles requests where all items are filtered out
// (e.g., invalid product IDs or non-positive quantities) without failing
// and skips the inventory validation block.
func TestProductoPedidoPost_AllInvalidItems(t *testing.T) {
	origQ := MockQuery
	// Only the row lock is expected; return a dummy row for any query.
	MockQuery = func(_ stdctx.Context, _ string, _ []driver.NamedValue) (driver.Rows, error) {
		return &mockRows{columns: []string{"ok"}, values: [][]driver.Value{{int64(1)}}}, nil
	}
	t.Cleanup(func() { MockQuery = origQ })

	body := `{"pedidoId":1,"detalles":[{"productoId":0,"cantidad":2},{"productoId":5,"cantidad":0}]}`
	r := httptest.NewRequest(http.MethodPost, "/producto_pedido", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)

	c := ProductoPedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()
	if w.Code != http.StatusCreated && w.Code != http.StatusOK {
		t.Fatalf("unexpected status %d. Body: %s", w.Code, w.Body.String())
	}
}
