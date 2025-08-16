package controllers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beego/beego/v2/server/web/context"
)

func TestPedidoGetAllWithoutDB(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/pedidos", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := PedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetAll()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al obtener los pedidos") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestPedidoPostDatabaseError(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/pedidos", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := PedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al crear el pedido") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestPedidoAssignDomicilioNotFound(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/pedidos/asignar-domicilio?pedido_id=1&domicilio_id=1", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := PedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.AssignDomicilio()

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Pedido no encontrado") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestPedidoAssignPagoNotFound(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/pedidos/asignar-pago?pedido_id=1&pago_id=1", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := PedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.AssignPago()

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Pedido no encontrado") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestPedidoUpdateEstadoPedidoNotFound(t *testing.T) {
	r := httptest.NewRequest(http.MethodPut, "/pedidos/actualizar-estado?pedido_id=1&estado=TERMINADO", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := PedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.UpdateEstadoPedido()

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Pedido no encontrado") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestPedidoGetPedidoDetailsMissingID(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/pedidos/detalles", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := PedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetPedidoDetails()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "El parámetro 'pedido_id' es obligatorio") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestPedidoGetPedidoDetailsDBError(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/pedidos/detalles?pedido_id=1", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := PedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetPedidoDetails()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al obtener los detalles del pedido") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}
