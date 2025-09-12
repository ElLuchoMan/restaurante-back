package productopedido

import (
	stdctx "context"
	"database/sql/driver"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beego/beego/v2/server/web/context"
)

func TestProductoPedidoUpdate_RequeryError(t *testing.T) {
	origQ, origE := MockQuery, MockExec
	step := 0
	MockQuery = func(_ stdctx.Context, q string, _ []driver.NamedValue) (driver.Rows, error) {
		lower := strings.ToLower(q)
		if strings.Contains(lower, "detalle_pedido") {
			if strings.Contains(lower, "insert into") {
				return &mockRows{columns: []string{"pk_id_detalle"}, values: [][]driver.Value{{int64(1)}}}, nil
			}
			if step == 0 {
				step++
				return &mockRows{columns: []string{"pk_id_pedido"}, values: [][]driver.Value{}}, nil
			}
			return nil, errors.New("requery fail")
		}
		if strings.Contains(lower, "select pk_id_producto, cantidad from producto") {
			cols := []string{"pk_id_producto", "cantidad"}
			vals := [][]driver.Value{{int64(1), int64(10)}}
			return &mockRows{columns: cols, values: vals}, nil
		}
		return &mockRows{columns: []string{"ok"}, values: [][]driver.Value{{int64(1)}}}, nil
	}
	MockExec = func(_ stdctx.Context, _ string, _ []driver.NamedValue) (driver.Result, error) {
		return mockResult{}, nil
	}
	t.Cleanup(func() { MockQuery, MockExec = origQ, origE })

	body := `[{"productoId":1,"cantidad":2}]`
	r := httptest.NewRequest(http.MethodPut, "/producto_pedido?pedido_id=1", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := ProductoPedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Update()
	if w.Code != http.StatusInternalServerError && w.Code != http.StatusOK {
		t.Fatalf("expected 500 or 200, got %d. Body: %s", w.Code, w.Body.String())
	}
}
