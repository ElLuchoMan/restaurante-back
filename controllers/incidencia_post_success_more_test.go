package controllers

import (
	stdctx "context"
	"database/sql/driver"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	beegoCtx "github.com/beego/beego/v2/server/web/context"
)

func TestIncidenciaPostSuccess_RestaTrue_MontoZero(t *testing.T) {
    origE, origQ := MockExec, MockQuery
    MockExec = func(_ stdctx.Context, _ string, _ []driver.NamedValue) (driver.Result, error) { return mockResult{}, nil }
    MockQuery = func(_ stdctx.Context, _ string, _ []driver.NamedValue) (driver.Rows, error) {
        cols := []string{"pk_id_incidencia"}
        vals := [][]driver.Value{{int64(2)}}
        return &mockRows{columns: cols, values: vals}, nil
    }
    t.Cleanup(func(){ MockExec, MockQuery = origE, origQ })

    body := `{"fechaIncidencia":"2024-03-01","monto":0,"resta":true,"motivo":"desc","documentoTrabajador":1}`
    r := httptest.NewRequest(http.MethodPost, "/incidencias", strings.NewReader(body))
    w := httptest.NewRecorder()
    ctx := beegoCtx.NewContext(); ctx.Reset(w, r); ctx.Input.RequestBody = []byte(body)
    c := IncidenciaController{}; c.Ctx = ctx; c.Data = make(map[interface{}]interface{})

    c.Post()
    if w.Code != http.StatusCreated && w.Code != http.StatusOK { t.Fatalf("unexpected status %d", w.Code) }
}


