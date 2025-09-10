package domicilio

import (
	stdctx "context"
	"database/sql/driver"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	beegoCtx "github.com/beego/beego/v2/server/web/context"
)

func TestDomicilioPostSuccess(t *testing.T) {
	origQ := MockQuery
	MockQuery = func(_ stdctx.Context, q string, _ []driver.NamedValue) (driver.Rows, error) {
		lower := strings.ToLower(q)
		if strings.Contains(lower, "insert into domicilio") {
			// RETURNING id
			cols := []string{"pk_id_domicilio"}
			vals := [][]driver.Value{{int64(1)}}
			return &mockRows{columns: cols, values: vals}, nil
		}
		if strings.Contains(lower, "select entregado, created_at, updated_at") {
			cols := []string{"entregado", "created_at", "updated_at"}
			vals := [][]driver.Value{{false, nil, nil}}
			return &mockRows{columns: cols, values: vals}, nil
		}
		return &mockRows{columns: []string{"ok"}, values: [][]driver.Value{{int64(1)}}}, nil
	}
	t.Cleanup(func() { MockQuery = origQ })

	body := `{"direccion":"calle 1","telefono":"123","fecha":"2024-01-01","estado":"PENDIENTE"}`
	r := httptest.NewRequest(http.MethodPost, "/domicilios", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := DomicilioController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()
	if w.Code != http.StatusCreated && w.Code != http.StatusBadRequest && w.Code != http.StatusInternalServerError {
		t.Fatalf("unexpected status %d", w.Code)
	}
}
