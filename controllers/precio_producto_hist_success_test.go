package controllers

import (
	stdctx "context"
	"database/sql/driver"
	"net/http"
	"net/http/httptest"
	"testing"

	beegoCtx "github.com/beego/beego/v2/server/web/context"
)

func TestPPH_GetAll_Success(t *testing.T) {
	orig := MockQuery
	MockQuery = func(_ stdctx.Context, _ string, _ []driver.NamedValue) (driver.Rows, error) {
		cols := []string{"nombre", "estado_producto", "precio", "fecha_vigencia"}
		vals := [][]driver.Value{{"Cafe", "DISPONIBLE", int64(1000), nil}}
		// fecha_vigencia como time.Time puede ir nil; beego ignora si no mapea
		return &mockRows{columns: cols, values: vals}, nil
	}
	t.Cleanup(func() { MockQuery = orig })

	r := httptest.NewRequest(http.MethodGet, "/precio_producto_hist?producto_id=1", nil)
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	c := &PrecioProductoHistController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.GetAll()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestPPH_GetById_Success(t *testing.T) {
	orig := MockQuery
	MockQuery = func(_ stdctx.Context, _ string, _ []driver.NamedValue) (driver.Rows, error) {
		cols := []string{"nombre", "estado_producto", "precio", "fecha_vigencia"}
		vals := [][]driver.Value{{"Te", "DISPONIBLE", int64(2000), nil}}
		return &mockRows{columns: cols, values: vals}, nil
	}
	t.Cleanup(func() { MockQuery = orig })

	r := httptest.NewRequest(http.MethodGet, "/precio_producto_hist/search?id=1", nil)
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	c := &PrecioProductoHistController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.GetById()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestPPH_GetAll_Error(t *testing.T) {
	orig := MockQuery
	// Forzar error genérico
	MockQuery = func(_ stdctx.Context, _ string, _ []driver.NamedValue) (driver.Rows, error) {
		return nil, driver.ErrBadConn
	}
	t.Cleanup(func() { MockQuery = orig })

	r := httptest.NewRequest(http.MethodGet, "/precio_producto_hist", nil)
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	c := &PrecioProductoHistController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.GetAll()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestPPH_GetAll_SuccessWithFecha(t *testing.T) {
	orig := MockQuery
	MockQuery = func(_ stdctx.Context, q string, _ []driver.NamedValue) (driver.Rows, error) {
		cols := []string{"nombre", "estado_producto", "precio", "fecha_vigencia"}
		vals := [][]driver.Value{{"Cafe", "DISPONIBLE", int64(1000), nil}}
		return &mockRows{columns: cols, values: vals}, nil
	}
	t.Cleanup(func() { MockQuery = orig })

	r := httptest.NewRequest(http.MethodGet, "/precio_producto_hist?fecha=2024-01-01", nil)
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	c := &PrecioProductoHistController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.GetAll()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
