package controllers

import (
	stdctx "context"
	"database/sql/driver"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/beego/beego/v2/server/web/context"
)

func TestControlNominaGetAllSuccessOne(t *testing.T) {
	orig := MockQuery
	MockQuery = func(_ stdctx.Context, q string, _ []driver.NamedValue) (driver.Rows, error) {
		cols := []string{"pk_id_control_nomina", "fecha", "estado"}
		vals := [][]driver.Value{{int64(1), "2024-01-01", "NO GENERADA"}}
		return &mockRows{columns: cols, values: vals}, nil
	}
	t.Cleanup(func() { MockQuery = orig })

	r := httptest.NewRequest(http.MethodGet, "/control_nomina", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := &ControlNominaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.GetAll()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestControlNominaGetByIdNotFound(t *testing.T) {
	orig := MockQuery
	MockQuery = func(_ stdctx.Context, q string, _ []driver.NamedValue) (driver.Rows, error) {
		cols := []string{"pk_id_control_nomina"}
		vals := [][]driver.Value{}
		return &mockRows{columns: cols, values: vals}, nil
	}
	t.Cleanup(func() { MockQuery = orig })

	r := httptest.NewRequest(http.MethodGet, "/control_nomina/search?id=1", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := &ControlNominaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.GetById()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestControlNominaGetByIdSuccessOne(t *testing.T) {
	orig := MockQuery
	MockQuery = func(_ stdctx.Context, q string, _ []driver.NamedValue) (driver.Rows, error) {
		cols := []string{"pk_id_control_nomina", "fecha", "estado"}
		vals := [][]driver.Value{{int64(2), "2024-02-01", "GENERADA"}}
		return &mockRows{columns: cols, values: vals}, nil
	}
	t.Cleanup(func() { MockQuery = orig })

	r := httptest.NewRequest(http.MethodGet, "/control_nomina/search?id=2", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := &ControlNominaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.GetById()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestControlNominaGetAllDBError(t *testing.T) {
	orig := MockQuery
	MockQuery = func(_ stdctx.Context, q string, _ []driver.NamedValue) (driver.Rows, error) {
		return nil, errors.New("db error")
	}
	t.Cleanup(func() { MockQuery = orig })

	r := httptest.NewRequest(http.MethodGet, "/control_nomina", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := &ControlNominaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.GetAll()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}
