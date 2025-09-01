package controllers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beego/beego/v2/server/web/context"
	"restaurante/models"
)

func int64Ptr(i int64) *int64 { return &i }

func TestValidateProducto(t *testing.T) {
	valid := &models.Producto{NOMBRE: "A", PRECIO: 10, ESTADO_PRODUCTO: "disponible"}
	if err := validateProducto(valid); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	tests := []models.Producto{
		{PRECIO: 10, ESTADO_PRODUCTO: "disponible"},                                      // missing name
		{NOMBRE: "B", PRECIO: 0, ESTADO_PRODUCTO: "disponible"},                          // zero price
		{NOMBRE: "B", PRECIO: -5, ESTADO_PRODUCTO: "disponible"},                         // negative price
		{NOMBRE: "B", PRECIO: 10, CALORIAS: int64Ptr(-1), ESTADO_PRODUCTO: "disponible"}, // negative calories
		{NOMBRE: "B", PRECIO: 10, ESTADO_PRODUCTO: "OTRO"},                               // invalid estado
	}

	for _, p := range tests {
		if err := validateProducto(&p); err == nil {
			t.Errorf("expected error for producto %+v", p)
		}
	}
}

func TestProductoGetAllWithoutDB(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/restaurante/v1/productos", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ProductoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetAll()

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al obtener productos") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestProductoGetByIdInvalidID(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/productos/search", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ProductoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetById()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestProductoGetByIdNotFound(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/productos/search?id=1", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ProductoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetById()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "error al buscar el producto") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestProductoPostMissingNombre(t *testing.T) {
	body := bytes.NewBufferString(`{"estadoProducto":"disponible","precio":10}`)
	r := httptest.NewRequest(http.MethodPost, "/productos", body)
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ProductoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "nombre") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestProductoPutInvalidID(t *testing.T) {
	r := httptest.NewRequest(http.MethodPut, "/productos", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ProductoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Put()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestProductoDeleteInvalidID(t *testing.T) {
	r := httptest.NewRequest(http.MethodDelete, "/productos", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ProductoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Delete()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}
