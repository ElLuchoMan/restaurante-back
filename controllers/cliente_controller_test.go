package controllers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beego/beego/v2/server/web/context"
)

func TestNormalizeEmail(t *testing.T) {
	got := normalizeEmail("  Foo@Example.COM  ")
	if got != "foo@example.com" {
		t.Errorf("expected foo@example.com, got %s", got)
	}
}

func TestIsUniqueEmailErr(t *testing.T) {
	uniqueErr := errors.New("pq: duplicate key value violates unique constraint \"uq_cliente_correo\"")
	if !isUniqueEmailErr(uniqueErr) {
		t.Errorf("expected true for unique email error")
	}
	otherErr := errors.New("some other error")
	if isUniqueEmailErr(otherErr) {
		t.Errorf("expected false for non unique email error")
	}
}

func TestClienteGetAllWithoutDB(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/restaurante/v1/clientes", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ClienteController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetAll()

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al obtener clientes") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestClienteGetByIdInvalidID(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/clientes/search", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ClienteController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetById()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestClienteGetByIdDatabaseError(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/clientes/search?id=1", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ClienteController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetById()

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al consultar el cliente") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestClientePostInvalidJSON(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/clientes", strings.NewReader("notjson"))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.CopyBody(1 << 20)
	c := ClienteController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestClientePostDatabaseError(t *testing.T) {
	body := "{\"documentoCliente\":1,\"nombre\":\"Foo\",\"apellido\":\"Bar\",\"direccion\":\"Calle\",\"telefono\":\"123\",\"password\":\"secret\"}"
	r := httptest.NewRequest(http.MethodPost, "/clientes", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.CopyBody(1 << 20)
	c := ClienteController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al crear el cliente") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestClientePutInvalidID(t *testing.T) {
	r := httptest.NewRequest(http.MethodPut, "/clientes", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ClienteController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Put()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestClientePutDatabaseError(t *testing.T) {
	r := httptest.NewRequest(http.MethodPut, "/clientes?id=1", strings.NewReader(`{"nombre":"Foo"}`))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.CopyBody(1 << 20)
	c := ClienteController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Put()

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al buscar el cliente") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestClienteDeleteInvalidID(t *testing.T) {
	r := httptest.NewRequest(http.MethodDelete, "/clientes", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ClienteController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Delete()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestClienteDeleteNotFound(t *testing.T) {
	r := httptest.NewRequest(http.MethodDelete, "/clientes?id=1", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ClienteController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Delete()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Cliente no encontrado") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}
