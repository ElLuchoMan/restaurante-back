package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beego/beego/v2/server/web/context"
)

func TestReservaGetAllWithoutDB(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/reservas", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ReservaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetAll()

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al obtener reservas") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestReservaGetByIdInvalidID(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/reservas/search", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ReservaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetById()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestReservaPostInvalidJSON(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/reservas", strings.NewReader("notjson"))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ReservaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestReservaPostInvalidDate(t *testing.T) {
	payload := map[string]interface{}{
		"fechaReserva":     "2024-13-01",
		"horaReserva":      "12:00:00",
		"personas":         2,
		"documentoCliente": 123,
	}
	body, _ := json.Marshal(payload)
	r := httptest.NewRequest(http.MethodPost, "/reservas", bytes.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = body
	c := ReservaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Formato de fecha inválido") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestReservaPostMissingHora(t *testing.T) {
	payload := map[string]interface{}{
		"fechaReserva":     "2024-01-01",
		"personas":         2,
		"documentoCliente": 123,
	}
	body, _ := json.Marshal(payload)
	r := httptest.NewRequest(http.MethodPost, "/reservas", bytes.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = body
	c := ReservaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "El campo HORA no puede estar vacío") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestReservaPutInvalidID(t *testing.T) {
	r := httptest.NewRequest(http.MethodPut, "/reservas", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ReservaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Put()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestReservaGetByParameterInvalidFecha(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/reservas/parameter?fecha=2024-13-01", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ReservaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetByParameter()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "formato YYYY-MM-DD") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestReservaGetByParameterDBError(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/reservas/parameter?documentoCliente=123&fecha=2024-10-10", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ReservaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetByParameter()

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al obtener reservas") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestReservaDeleteInvalidID(t *testing.T) {
	r := httptest.NewRequest(http.MethodDelete, "/reservas", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ReservaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Delete()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}
