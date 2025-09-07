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

func TestIncidenciaPostSuccess(t *testing.T) {
    // Mock insert para que no falle
    origE := MockExec
    origQ := MockQuery
    MockExec = func(_ stdctx.Context, _ string, _ []driver.NamedValue) (driver.Result, error) { return mockResult{}, nil }
    // Algunas bases devuelven el ID con RETURNING -> cubrir Query también
    MockQuery = func(_ stdctx.Context, _ string, _ []driver.NamedValue) (driver.Rows, error) {
        cols := []string{"pk_id_incidencia"}
        vals := [][]driver.Value{{int64(1)}}
        return &mockRows{columns: cols, values: vals}, nil
    }
    t.Cleanup(func(){ MockExec = origE; MockQuery = origQ })

    body := `{"fechaIncidencia":"2024-01-01","monto":100,"resta":false,"motivo":"x","documentoTrabajador":1}`
    r := httptest.NewRequest(http.MethodPost, "/incidencias", strings.NewReader(body))
    w := httptest.NewRecorder()
    ctx := beegoCtx.NewContext(); ctx.Reset(w, r); ctx.Input.RequestBody = []byte(body)
    c := IncidenciaController{}; c.Ctx = ctx; c.Data = make(map[interface{}]interface{})

    c.Post()
    if w.Code != http.StatusCreated { t.Fatalf("expected 201, got %d", w.Code) }
}

func TestIncidenciaPutSuccess(t *testing.T) {
    // Mock update ok (ORM Update -> Exec)
    origE := MockExec
    MockExec = func(_ stdctx.Context, _ string, _ []driver.NamedValue) (driver.Result, error) { return mockResult{}, nil }
    t.Cleanup(func(){ MockExec = origE })

    body := `{"fechaIncidencia":"2024-01-02","monto":110,"resta":true,"motivo":"y","documentoTrabajador":1}`
    r := httptest.NewRequest(http.MethodPut, "/incidencias?id=1", strings.NewReader(body))
    w := httptest.NewRecorder()
    ctx := beegoCtx.NewContext(); ctx.Reset(w, r); ctx.Input.RequestBody = []byte(body)
    c := IncidenciaController{}; c.Ctx = ctx; c.Data = make(map[interface{}]interface{})

    c.Put()
    if w.Code != http.StatusOK { t.Fatalf("expected 200, got %d", w.Code) }
}

func TestIncidenciaGetAllSuccess(t *testing.T) {
    // Basta con que la consulta no falle y devuelva 0 filas
    orig := MockQuery
    MockQuery = func(_ stdctx.Context, _ string, _ []driver.NamedValue) (driver.Rows, error) {
        return &mockRows{columns: []string{"pk_id_incidencia"}, values: [][]driver.Value{}}, nil
    }
    t.Cleanup(func(){ MockQuery = orig })

    r := httptest.NewRequest(http.MethodGet, "/incidencias", nil)
    w := httptest.NewRecorder()
    ctx := beegoCtx.NewContext(); ctx.Reset(w, r)
    c := IncidenciaController{}; c.Ctx = ctx; c.Data = make(map[interface{}]interface{})

    c.GetAll()
    if w.Code != http.StatusOK { t.Fatalf("expected 200, got %d", w.Code) }
}


