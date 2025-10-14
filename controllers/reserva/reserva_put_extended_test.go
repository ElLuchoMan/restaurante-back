package reserva

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
)

func TestReservaPut_BadJSON(t *testing.T) {

	ctx := context.NewContext()
	recorder := httptest.NewRecorder()

	invalidJSON := []byte("{invalid json}")
	req := httptest.NewRequest("PUT", "/reservas?id=1", bytes.NewReader(invalidJSON))
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = invalidJSON

	controller := &ReservaController{}
	controller.Init(ctx, "ReservaController", "Put", nil)

	origRead := readReserva
	defer func() { readReserva = origRead }()
	readReserva = func(o orm.Ormer, r *models.Reserva) error {
		return nil
	}

	controller.Put()

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}

func TestReservaPut_InvalidFechaFormat(t *testing.T) {

	ctx := context.NewContext()
	recorder := httptest.NewRecorder()

	body := map[string]interface{}{
		"fechaReserva": "invalid-date",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("PUT", "/reservas?id=1", bytes.NewReader(bodyBytes))
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = bodyBytes

	controller := &ReservaController{}
	controller.Init(ctx, "ReservaController", "Put", nil)

	origRead := readReserva
	defer func() { readReserva = origRead }()
	readReserva = func(o orm.Ormer, r *models.Reserva) error {
		return nil
	}

	controller.Put()

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}

func TestReservaPut_InvalidHoraFormat(t *testing.T) {

	ctx := context.NewContext()
	recorder := httptest.NewRecorder()

	body := map[string]interface{}{
		"horaReserva": "invalid-time",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("PUT", "/reservas?id=1", bytes.NewReader(bodyBytes))
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = bodyBytes

	controller := &ReservaController{}
	controller.Init(ctx, "ReservaController", "Put", nil)

	origRead := readReserva
	defer func() { readReserva = origRead }()
	readReserva = func(o orm.Ormer, r *models.Reserva) error {
		return nil
	}

	controller.Put()

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}

func TestReservaPut_InvalidEstado(t *testing.T) {

	ctx := context.NewContext()
	recorder := httptest.NewRecorder()

	body := map[string]interface{}{
		"estadoReserva": "ESTADO_INVALIDO",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("PUT", "/reservas?id=1", bytes.NewReader(bodyBytes))
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = bodyBytes

	controller := &ReservaController{}
	controller.Init(ctx, "ReservaController", "Put", nil)

	origRead := readReserva
	defer func() { readReserva = origRead }()
	readReserva = func(o orm.Ormer, r *models.Reserva) error {
		return nil
	}

	controller.Put()

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}

func TestReservaPut_UpdatePersonas(t *testing.T) {

	ctx := context.NewContext()
	recorder := httptest.NewRecorder()

	body := map[string]interface{}{
		"personas": float64(6),
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("PUT", "/reservas?id=1", bytes.NewReader(bodyBytes))
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = bodyBytes

	controller := &ReservaController{}
	controller.Init(ctx, "ReservaController", "Put", nil)

	origRead := readReserva
	defer func() { readReserva = origRead }()
	readReserva = func(o orm.Ormer, r *models.Reserva) error {
		return nil
	}

	origUpdate := updateReserva
	defer func() { updateReserva = origUpdate }()
	updateReserva = func(o orm.Ormer, r *models.Reserva, cols ...string) (int64, error) {
		if r.PERSONAS != 6 {
			t.Errorf("Expected personas to be 6, got %d", r.PERSONAS)
		}
		return 1, nil
	}

	controller.Put()

	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, recorder.Code)
	}
}

func TestReservaPut_UpdateIndicaciones(t *testing.T) {

	ctx := context.NewContext()
	recorder := httptest.NewRecorder()

	indicaciones := "Sin cebolla por favor"
	body := map[string]interface{}{
		"indicaciones": indicaciones,
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("PUT", "/reservas?id=1", bytes.NewReader(bodyBytes))
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = bodyBytes

	controller := &ReservaController{}
	controller.Init(ctx, "ReservaController", "Put", nil)

	origRead := readReserva
	defer func() { readReserva = origRead }()
	readReserva = func(o orm.Ormer, r *models.Reserva) error {
		return nil
	}

	origUpdate := updateReserva
	defer func() { updateReserva = origUpdate }()
	updateReserva = func(o orm.Ormer, r *models.Reserva, cols ...string) (int64, error) {
		if r.INDICACIONES == nil || *r.INDICACIONES != indicaciones {
			t.Errorf("Expected indicaciones to be '%s'", indicaciones)
		}
		return 1, nil
	}

	controller.Put()

	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, recorder.Code)
	}
}

func TestReservaPut_UpdatedBy(t *testing.T) {

	ctx := context.NewContext()
	recorder := httptest.NewRecorder()

	updatedBy := "admin@test.com"
	body := map[string]interface{}{
		"updatedBy": updatedBy,
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("PUT", "/reservas?id=1", bytes.NewReader(bodyBytes))
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = bodyBytes

	controller := &ReservaController{}
	controller.Init(ctx, "ReservaController", "Put", nil)

	origRead := readReserva
	defer func() { readReserva = origRead }()
	readReserva = func(o orm.Ormer, r *models.Reserva) error {
		return nil
	}

	origUpdate := updateReserva
	defer func() { updateReserva = origUpdate }()
	updateReserva = func(o orm.Ormer, r *models.Reserva, cols ...string) (int64, error) {
		if r.UPDATED_BY == nil || *r.UPDATED_BY != updatedBy {
			t.Errorf("Expected updatedBy to be '%s'", updatedBy)
		}
		return 1, nil
	}

	controller.Put()

	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, recorder.Code)
	}
}

func TestReservaPut_UpdateRestaurante(t *testing.T) {

	ctx := context.NewContext()
	recorder := httptest.NewRecorder()

	body := map[string]interface{}{
		"restauranteId": float64(2),
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("PUT", "/reservas?id=1", bytes.NewReader(bodyBytes))
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = bodyBytes

	controller := &ReservaController{}
	controller.Init(ctx, "ReservaController", "Put", nil)

	origRead := readReserva
	defer func() { readReserva = origRead }()
	readReserva = func(o orm.Ormer, r *models.Reserva) error {
		return nil
	}

	origUpdate := updateReserva
	defer func() { updateReserva = origUpdate }()
	updateReserva = func(o orm.Ormer, r *models.Reserva, cols ...string) (int64, error) {
		if r.PK_ID_RESTAURANTE == nil || r.PK_ID_RESTAURANTE.PK_ID_RESTAURANTE != 2 {
			t.Error("Expected restauranteId to be 2")
		}
		return 1, nil
	}

	controller.Put()

	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, recorder.Code)
	}
}

func TestReservaPut_ValidFechaAndHora(t *testing.T) {

	ctx := context.NewContext()
	recorder := httptest.NewRecorder()

	body := map[string]interface{}{
		"fechaReserva": "2025-12-25",
		"horaReserva":  "19:30:00",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("PUT", "/reservas?id=1", bytes.NewReader(bodyBytes))
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = bodyBytes

	controller := &ReservaController{}
	controller.Init(ctx, "ReservaController", "Put", nil)

	origRead := readReserva
	defer func() { readReserva = origRead }()
	readReserva = func(o orm.Ormer, r *models.Reserva) error {
		return nil
	}

	origUpdate := updateReserva
	defer func() { updateReserva = origUpdate }()
	updateReserva = func(o orm.Ormer, r *models.Reserva, cols ...string) (int64, error) {

		expectedDate, _ := time.Parse("2006-01-02", "2025-12-25")
		expectedNormalized := time.Date(expectedDate.Year(), expectedDate.Month(), expectedDate.Day(), 12, 0, 0, 0, time.UTC)
		if !r.FECHA.Equal(expectedNormalized) {
			t.Errorf("Expected fecha to be normalized to noon UTC (2025-12-25 12:00:00 UTC), got %v", r.FECHA)
		}
		return 1, nil
	}

	controller.Put()

	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, recorder.Code)
	}
}

func TestReservaPut_ValidEstado(t *testing.T) {

	ctx := context.NewContext()
	recorder := httptest.NewRecorder()

	body := map[string]interface{}{
		"estadoReserva": "CONFIRMADA",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("PUT", "/reservas?id=1", bytes.NewReader(bodyBytes))
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = bodyBytes

	controller := &ReservaController{}
	controller.Init(ctx, "ReservaController", "Put", nil)

	origRead := readReserva
	defer func() { readReserva = origRead }()
	readReserva = func(o orm.Ormer, r *models.Reserva) error {
		return nil
	}

	origUpdate := updateReserva
	defer func() { updateReserva = origUpdate }()
	updateReserva = func(o orm.Ormer, r *models.Reserva, cols ...string) (int64, error) {
		if r.ESTADO_RESERVA == nil || *r.ESTADO_RESERVA != models.EstadoReservaConfirmada {
			t.Error("Expected estado to be CONFIRMADA")
		}
		return 1, nil
	}

	controller.Put()

	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, recorder.Code)
	}
}
