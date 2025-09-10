package domicilio

import (
	stdctx "context"
	"database/sql/driver"
	"net/http"
	"net/http/httptest"
	"testing"

	beegoCtx "github.com/beego/beego/v2/server/web/context"
)

func TestDomicilioAsignarDomiciliarioSuccess(t *testing.T) {
	origQ, origE := MockQuery, MockExec
	// UPDATE devuelve filas afectadas
	MockExec = func(_ stdctx.Context, _ string, _ []driver.NamedValue) (driver.Result, error) {
		return mockResult{}, nil
	}
	// SELECT COUNT(1) no se usa cuando affected==1, pero dejamos un fallback
	MockQuery = func(_ stdctx.Context, _ string, _ []driver.NamedValue) (driver.Rows, error) {
		cols := []string{"count"}
		vals := [][]driver.Value{{int64(1)}}
		return &mockRows{columns: cols, values: vals}, nil
	}
	t.Cleanup(func() { MockQuery, MockExec = origQ, origE })

	r := httptest.NewRequest(http.MethodPost, "/domicilios/asignar?domicilio_id=1&trabajador_id=1", nil)
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	c := DomicilioController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.AsignarDomiciliario()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
