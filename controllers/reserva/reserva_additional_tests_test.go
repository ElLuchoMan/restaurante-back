package reserva

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
)

func TestReservaGetAll_Success(t *testing.T) {

	ctx := context.NewContext()
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/reservas", nil)
	ctx.Reset(recorder, req)

	controller := &ReservaController{}
	controller.Init(ctx, "ReservaController", "GetAll", nil)

	origQueryAll := queryAllReservas
	defer func() { queryAllReservas = origQueryAll }()

	estadoPendiente := models.EstadoReservaPendiente
	mockReservas := []models.Reserva{
		{
			PK_ID_RESERVA:  1,
			FECHA:          time.Now(),
			HORA:           time.Now(),
			PERSONAS:       4,
			ESTADO_RESERVA: &estadoPendiente,
		},
	}
	queryAllReservas = func(o orm.Ormer, reservas *[]models.Reserva) (int64, error) {
		*reservas = mockReservas
		return int64(len(mockReservas)), nil
	}

	controller.GetAll()

	if recorder.Code != 200 {
		t.Errorf("Expected status 200, got %d", recorder.Code)
	}
}

func TestReservaPost_MissingDocumentos(t *testing.T) {

	ctx := context.NewContext()
	recorder := httptest.NewRecorder()

	body := map[string]interface{}{
		"fecha":          "2025-01-20",
		"hora":           "19:00:00",
		"personas":       4,
		"restauranteId":  1,
		"nombreCompleto": "Juan Perez",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/reservas", bytes.NewReader(bodyBytes))
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = bodyBytes

	controller := &ReservaController{}
	controller.Init(ctx, "ReservaController", "Post", nil)

	controller.Post()

	if recorder.Code == 201 {
		t.Errorf("Expected error status, got 201")
	}
}

func TestReservaGetByDocumento_WithFecha(t *testing.T) {

	ctx := context.NewContext()
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/reservas/documento?documento=123456&fecha=2025-01-20", nil)
	ctx.Reset(recorder, req)
	ctx.Input.SetParam("documento", "123456")
	ctx.Input.SetParam("fecha", "2025-01-20")

	controller := &ReservaController{}
	controller.Init(ctx, "ReservaController", "GetByDocumento", nil)

	origQueryByCliente := queryReservasByDocumentoCliente
	defer func() { queryReservasByDocumentoCliente = origQueryByCliente }()

	queryReservasByDocumentoCliente = func(o orm.Ormer, documentoCliente int64, fecha time.Time, useFecha bool, reservas *[]models.Reserva) (int64, error) {
		*reservas = []models.Reserva{{PK_ID_RESERVA: 1}}
		return 1, nil
	}

	controller.GetByDocumento()

	if recorder.Code != 200 {
		t.Errorf("Expected status 200, got %d", recorder.Code)
	}
}

func TestReservaDelete_Success(t *testing.T) {

	ctx := context.NewContext()
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/reservas?id=1", nil)
	ctx.Reset(recorder, req)
	ctx.Input.SetParam("id", "1")

	controller := &ReservaController{}
	controller.Init(ctx, "ReservaController", "Delete", nil)

	origRead := readReserva
	origUpdate := updateReserva
	defer func() {
		readReserva = origRead
		updateReserva = origUpdate
	}()

	estadoPendiente := models.EstadoReservaPendiente
	readReserva = func(o orm.Ormer, r *models.Reserva) error {
		r.PK_ID_RESERVA = 1
		r.ESTADO_RESERVA = &estadoPendiente
		return nil
	}

	updateReserva = func(o orm.Ormer, r *models.Reserva, cols ...string) (int64, error) {
		if r.ESTADO_RESERVA == nil || *r.ESTADO_RESERVA != models.EstadoReservaCancelada {
			return 0, fmt.Errorf("expected estado cancelada")
		}
		return 1, nil
	}

	controller.Delete()

	if recorder.Code != 200 {
		t.Errorf("Expected status 200, got %d", recorder.Code)
	}
}

func TestCreateOrFindReservaContacto_NewContactoNoLoggeado(t *testing.T) {

	origQueryByDoc := queryReservaContactoByDocumento
	origInsert := insertReservaContacto
	defer func() {
		queryReservaContactoByDocumento = origQueryByDoc
		insertReservaContacto = origInsert
	}()

	queryReservaContactoByDocumento = func(o orm.Ormer, documento int64, rc *models.ReservaContacto) error {
		return orm.ErrNoRows
	}

	insertReservaContacto = func(o orm.Ormer, rc *models.ReservaContacto) (int64, error) {
		return 123, nil
	}

	input := map[string]interface{}{
		"documentoContacto": float64(987654321),
		"nombreCompleto":    "Test User",
		"telefono":          "3001234567",
	}

	contacto, err := createOrFindReservaContacto(nil, input)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if contacto == nil {
		t.Fatal("Expected contacto, got nil")
	}
	if contacto.PKIDContacto != 123 {
		t.Errorf("Expected PKIDContacto 123, got %d", contacto.PKIDContacto)
	}
}

func TestCreateOrFindReservaContacto_MissingNombreCompleto(t *testing.T) {

	origQueryByDoc := queryReservaContactoByDocumento
	defer func() { queryReservaContactoByDocumento = origQueryByDoc }()

	queryReservaContactoByDocumento = func(o orm.Ormer, documento int64, rc *models.ReservaContacto) error {
		return orm.ErrNoRows
	}

	input := map[string]interface{}{
		"documentoContacto": float64(987654321),
	}

	contacto, err := createOrFindReservaContacto(nil, input)

	if err == nil {
		t.Error("Expected error for missing nombreCompleto")
	}
	if contacto != nil {
		t.Error("Expected nil contacto on error")
	}
}
