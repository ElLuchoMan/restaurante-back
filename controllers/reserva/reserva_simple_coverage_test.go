package reserva

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
)

// Tests simples para mejorar cobertura

func TestReservaPut_UpdateError(t *testing.T) {
	// Setup
	ctx := context.NewContext()
	recorder := httptest.NewRecorder()

	body := map[string]interface{}{
		"personas": float64(4),
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("PUT", "/reservas?id=1", bytes.NewReader(bodyBytes))
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = bodyBytes

	controller := &ReservaController{}
	controller.Init(ctx, "ReservaController", "Put", nil)

	// Mock readReserva
	origRead := readReserva
	defer func() { readReserva = origRead }()
	readReserva = func(o orm.Ormer, r *models.Reserva) error {
		return nil
	}

	// Mock updateReserva para retornar error
	origUpdate := updateReserva
	defer func() { updateReserva = origUpdate }()
	updateReserva = func(o orm.Ormer, r *models.Reserva, cols ...string) (int64, error) {
		return 0, orm.ErrNoRows // Error simulado
	}

	// Execute
	controller.Put()

	// Verify - debería retornar 500
	if recorder.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, recorder.Code)
	}
}

func TestReservaGetAll_Error(t *testing.T) {
	// Setup
	ctx := context.NewContext()
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/reservas", nil)
	ctx.Reset(recorder, req)

	controller := &ReservaController{}
	controller.Init(ctx, "ReservaController", "GetAll", nil)

	// Mock queryAllReservas para retornar error
	origQueryAll := queryAllReservas
	defer func() { queryAllReservas = origQueryAll }()
	queryAllReservas = func(o orm.Ormer, reservas *[]models.Reserva) (int64, error) {
		return 0, orm.ErrNoRows // Error simulado
	}

	// Execute
	controller.GetAll()

	// Verify - debería retornar 500
	if recorder.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, recorder.Code)
	}
}

func TestReservaPost_DocumentoContactoConNombreVacio(t *testing.T) {
	// Setup
	ctx := context.NewContext()
	recorder := httptest.NewRecorder()

	body := map[string]interface{}{
		"documentoContacto": float64(123456),
		"nombreCompleto":    "", // Nombre vacío
		"fecha":             "2025-12-25",
		"hora":              "19:00:00",
		"personas":          float64(4),
		"restauranteId":     float64(1),
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/reservas", bytes.NewReader(bodyBytes))
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = bodyBytes

	controller := &ReservaController{}
	controller.Init(ctx, "ReservaController", "Post", nil)

	// Mock queryReservaContactoByDocumento para no encontrar contacto
	origQuery := queryReservaContactoByDocumento
	defer func() { queryReservaContactoByDocumento = origQuery }()
	queryReservaContactoByDocumento = func(o orm.Ormer, documento int64, rc *models.ReservaContacto) error {
		return orm.ErrNoRows // No encontrado
	}

	// Execute
	controller.Post()

	// Verify - debería retornar 400 porque falta nombreCompleto
	if recorder.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}

func TestReservaPost_DocumentoContactoConTelefono(t *testing.T) {
	// Setup
	ctx := context.NewContext()
	recorder := httptest.NewRecorder()

	telefono := "1234567890"
	body := map[string]interface{}{
		"documentoContacto": float64(123456),
		"nombreCompleto":    "Juan Perez",
		"telefono":          telefono,
		"fechaReserva":      "2025-12-25",
		"horaReserva":       "19:00:00",
		"personas":          float64(4),
		"restauranteId":     float64(1),
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/reservas", bytes.NewReader(bodyBytes))
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = bodyBytes

	controller := &ReservaController{}
	controller.Init(ctx, "ReservaController", "Post", nil)

	// Mock queryReservaContactoByDocumento
	origQuery := queryReservaContactoByDocumento
	defer func() { queryReservaContactoByDocumento = origQuery }()
	queryReservaContactoByDocumento = func(o orm.Ormer, documento int64, rc *models.ReservaContacto) error {
		return orm.ErrNoRows
	}

	// Mock insertReservaContacto
	origInsert := insertReservaContacto
	defer func() { insertReservaContacto = origInsert }()
	contactoInserted := false
	insertReservaContacto = func(o orm.Ormer, rc *models.ReservaContacto) (int64, error) {
		contactoInserted = true
		if rc.Telefono == nil || *rc.Telefono != telefono {
			t.Errorf("Expected telefono to be %s", telefono)
		}
		return 1, nil
	}

	// Mock insertReserva
	origInsertReserva := insertReserva
	defer func() { insertReserva = origInsertReserva }()
	insertReserva = func(o orm.Ormer, r *models.Reserva) (int64, error) {
		return 1, nil
	}

	// Execute
	controller.Post()

	// Verify
	if !contactoInserted {
		t.Error("Expected contacto to be inserted")
	}
	if recorder.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, recorder.Code)
	}
}

func TestReservaPost_DocumentoClienteConTelefono(t *testing.T) {
	// Setup
	ctx := context.NewContext()
	recorder := httptest.NewRecorder()

	body := map[string]interface{}{
		"documentoCliente": float64(789012),
		"fechaReserva":     "2025-12-25",
		"horaReserva":      "19:00:00",
		"personas":         float64(4),
		"restauranteId":    float64(1),
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/reservas", bytes.NewReader(bodyBytes))
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = bodyBytes

	controller := &ReservaController{}
	controller.Init(ctx, "ReservaController", "Post", nil)

	// Mock queryReservaContactoByCliente
	origQueryCliente := queryReservaContactoByCliente
	defer func() { queryReservaContactoByCliente = origQueryCliente }()
	queryReservaContactoByCliente = func(o orm.Ormer, clienteDoc int64, rc *models.ReservaContacto) error {
		return orm.ErrNoRows
	}

	// Mock readCliente
	origReadCliente := readCliente
	defer func() { readCliente = origReadCliente }()
	readCliente = func(o orm.Ormer, c *models.Cliente) error {
		c.NOMBRE = "Maria"
		c.APELLIDO = "Gomez"
		c.TELEFONO = "0987654321"
		return nil
	}

	// Mock insertReservaContacto
	origInsert := insertReservaContacto
	defer func() { insertReservaContacto = origInsert }()
	contactoInserted := false
	insertReservaContacto = func(o orm.Ormer, rc *models.ReservaContacto) (int64, error) {
		contactoInserted = true
		if rc.Telefono == nil || *rc.Telefono != "0987654321" {
			t.Error("Expected telefono from cliente")
		}
		return 1, nil
	}

	// Mock insertReserva
	origInsertReserva := insertReserva
	defer func() { insertReserva = origInsertReserva }()
	insertReserva = func(o orm.Ormer, r *models.Reserva) (int64, error) {
		return 1, nil
	}

	// Execute
	controller.Post()

	// Verify
	if !contactoInserted {
		t.Error("Expected contacto to be inserted")
	}
	if recorder.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, recorder.Code)
	}
}
