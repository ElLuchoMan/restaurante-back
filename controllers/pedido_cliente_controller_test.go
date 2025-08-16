package controllers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beego/beego/v2/server/web/context"
)

func TestPedidoClienteGetAllWithoutDB(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/pedido_clientes", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := PedidoClienteController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetAll()

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al obtener las relaciones de la base de datos") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestPedidoClientePostInvalidJSON(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/pedido_clientes", strings.NewReader("{"))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := PedidoClienteController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error en la solicitud") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestPedidoClientePostMissingFields(t *testing.T) {
	body := `{"documentoCliente":0,"pedidoId":0}`
	r := httptest.NewRequest(http.MethodPost, "/pedido_clientes", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := PedidoClienteController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Faltan campos obligatorios") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestPedidoClientePostClienteNoEncontrado(t *testing.T) {
	body := `{"documentoCliente":1,"pedidoId":1}`
	r := httptest.NewRequest(http.MethodPost, "/pedido_clientes", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := PedidoClienteController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}
