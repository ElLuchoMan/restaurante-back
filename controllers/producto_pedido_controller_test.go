package controllers

import (
	"bytes"
	stdctx "context"
	"database/sql/driver"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	beegoCtx "github.com/beego/beego/v2/server/web/context"
)

func TestProductoPedidoGetAllMissingParam(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/producto_pedido", nil)
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	c := ProductoPedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetAll()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "pedido_id") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestProductoPedidoGetAllDBError(t *testing.T) {
	MockQuery = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
		return nil, errors.New("db error")
	}
	defer func() { MockQuery = nil }()

	r := httptest.NewRequest(http.MethodGet, "/producto_pedido?pedido_id=1", nil)
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	c := ProductoPedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetAll()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al obtener los productos del pedido") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestProductoPedidoGetAllSuccess(t *testing.T) {
	MockQuery = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
		cols := []string{"producto_id", "nombre", "cantidad", "precio_unitario", "subtotal"}
		vals := [][]driver.Value{{int64(1), "Cafe", int64(1), float64(10), float64(10)}}
		return &mockRows{columns: cols, values: vals}, nil
	}
	defer func() { MockQuery = nil }()

	r := httptest.NewRequest(http.MethodGet, "/producto_pedido?pedido_id=1", nil)
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	c := ProductoPedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetAll()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "detallesProductos") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestProductoPedidoPostInvalidJSON(t *testing.T) {
	body := "{"
	r := httptest.NewRequest(http.MethodPost, "/producto_pedido", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := ProductoPedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Datos inválidos") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestProductoPedidoPostMissingFields(t *testing.T) {
	body := `{"pedidoId":0,"detallesProductos":[]}`
	r := httptest.NewRequest(http.MethodPost, "/producto_pedido", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := ProductoPedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "obligatorios") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestProductoPedidoPostDBError(t *testing.T) {
	MockExec = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Result, error) {
		return nil, errors.New("db error")
	}
	defer func() { MockExec = nil }()

	body := `{"pedidoId":1,"detallesProductos":[{"productoId":1,"cantidad":1}]}`
	r := httptest.NewRequest(http.MethodPost, "/producto_pedido", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := ProductoPedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al crear el pedido con productos") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestProductoPedidoPostSuccess(t *testing.T) {
	MockExec = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Result, error) {
		return mockResult{}, nil
	}
	defer func() { MockExec = nil }()

	body := `{"pedidoId":1,"detallesProductos":[{"productoId":1,"cantidad":1}]}`
	r := httptest.NewRequest(http.MethodPost, "/producto_pedido", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := ProductoPedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "pedidoId") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestProductoPedidoUpdateMissingParam(t *testing.T) {
	body := "[]"
	r := httptest.NewRequest(http.MethodPut, "/producto_pedido", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := ProductoPedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Update()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "pedido_id") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestProductoPedidoUpdateInvalidJSON(t *testing.T) {
	body := "{"
	r := httptest.NewRequest(http.MethodPut, "/producto_pedido?pedido_id=1", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := ProductoPedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Update()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Datos inválidos") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestProductoPedidoUpdateEmptyList(t *testing.T) {
	body := "[]"
	r := httptest.NewRequest(http.MethodPut, "/producto_pedido?pedido_id=1", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := ProductoPedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Update()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "no puede estar vacía") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestProductoPedidoUpdateDBError(t *testing.T) {
	MockQuery = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
		return nil, errors.New("db error")
	}
	defer func() { MockQuery = nil }()

	body := `[{"productoId":1,"cantidad":1}]`
	r := httptest.NewRequest(http.MethodPut, "/producto_pedido?pedido_id=1", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := ProductoPedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Update()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al buscar el pedido") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}
