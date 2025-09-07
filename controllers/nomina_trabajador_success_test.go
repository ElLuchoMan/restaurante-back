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

func TestNominaTrabajadorPostSuccess(t *testing.T) {
    // Simular consultas necesarias: incidencias (0 filas), sueldo trabajador, última nomina y exist no
    origQ, origE := MockQuery, MockExec
    step := 0
    MockQuery = func(_ stdctx.Context, q string, _ []driver.NamedValue) (driver.Rows, error) {
        step++
        // devolver alguna fila para One de trabajador y de nómina
        if step == 1 { // incidencias -> sin filas
            return &mockRows{columns: []string{"pk_id_incidencia"}, values: [][]driver.Value{}}, nil
        }
        if step == 2 { // trabajador
            cols := []string{"sueldo"}
            vals := [][]driver.Value{{int64(1000)}}
            return &mockRows{columns: cols, values: vals}, nil
        }
        if step == 3 { // ultima nomina
            cols := []string{"pk_id_nomina", "fecha"}
            vals := [][]driver.Value{{int64(1), nil}}
            return &mockRows{columns: cols, values: vals}, nil
        }
        if step == 4 { // exist false -> no devuelve filas que coincidan
            return &mockRows{columns: []string{"pk_id_nomina"}, values: [][]driver.Value{}}, nil
        }
        return &mockRows{columns: []string{"ok"}, values: [][]driver.Value{{int64(1)}}}, nil
    }
    MockExec = func(_ stdctx.Context, _ string, _ []driver.NamedValue) (driver.Result, error) { return mockResult{}, nil }
    t.Cleanup(func(){ MockQuery, MockExec = origQ, origE })

    body := `{"documentoTrabajador":123}`
    r := httptest.NewRequest(http.MethodPost, "/nomina_trabajador", strings.NewReader(body))
    w := httptest.NewRecorder(); ctx := beegoCtx.NewContext(); ctx.Reset(w, r); ctx.Input.RequestBody = []byte(body)
    c := NominaTrabajadorController{}; c.Ctx = ctx; c.Data = make(map[interface{}]interface{})

    c.Post()
    if w.Code != http.StatusCreated && w.Code != http.StatusInternalServerError {
        t.Fatalf("unexpected status %d", w.Code)
    }
}


