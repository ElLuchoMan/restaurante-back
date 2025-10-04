package reserva

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"

	"restaurante/database"
	"restaurante/models"
)

var (
	originalQueryAllReservas     = queryAllReservas
	originalReadReserva          = readReserva
	originalInsertReserva        = insertReserva
	originalUpdateReserva        = updateReserva
	originalQueryReservasByParam = queryReservasByParam
)

func resetReservaMocks() {
	ormNew = orm.NewOrm
	queryAllReservas = func(o orm.Ormer, reservas *[]models.Reserva) (int64, error) {
		return o.QueryTable(new(models.Reserva)).All(reservas)
	}
	readReserva = func(o orm.Ormer, r *models.Reserva) error { return o.Read(r) }
	insertReserva = func(o orm.Ormer, r *models.Reserva) (int64, error) { return o.Insert(r) }
	updateReserva = func(o orm.Ormer, r *models.Reserva, cols ...string) (int64, error) { return o.Update(r, cols...) }
	queryReservasByParam = func(o orm.Ormer, contacto int64, fecha time.Time, useContacto, useFecha bool, reservas *[]models.Reserva) (int64, error) {
		qs := o.QueryTable(new(models.Reserva))
		if useContacto {
			qs = qs.Filter("pk_id_contacto", contacto)
		}
		if useFecha {
			qs = qs.Filter("fecha", fecha)
		}
		return qs.All(reservas)
	}
}

func TestReservaGetAllWithoutDB(t *testing.T) {
	t.Cleanup(resetReservaMocks)
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

func TestReservaDefaultFuncsCoverage(t *testing.T) {
	o := orm.NewOrm()
	var r models.Reserva
	var rs []models.Reserva
	originalQueryAllReservas(o, &rs)
	_ = originalReadReserva(o, &r)
	originalInsertReserva(o, &r)
	originalUpdateReserva(o, &r)
	originalQueryReservasByParam(o, 1, time.Now(), true, true, &rs)
}

func TestReservaGetByIdInvalidID(t *testing.T) {
	t.Cleanup(resetReservaMocks)
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
	t.Cleanup(resetReservaMocks)
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
	t.Cleanup(resetReservaMocks)
	payload := map[string]interface{}{
		"fechaReserva":      "2024-13-01",
		"horaReserva":       "12:00:00",
		"personas":          2,
		"contactoId":        123,
		"restauranteId":     1,
		"documentoContacto": 123456789,
		"nombreCompleto":    "Test User",
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
	t.Cleanup(resetReservaMocks)
	payload := map[string]interface{}{
		"fechaReserva":      "2024-01-01",
		"personas":          2,
		"contactoId":        123,
		"restauranteId":     1,
		"documentoContacto": 123456789,
		"nombreCompleto":    "Test User",
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
	t.Cleanup(resetReservaMocks)
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
	t.Cleanup(resetReservaMocks)
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
	t.Cleanup(resetReservaMocks)
	r := httptest.NewRequest(http.MethodGet, "/reservas/parameter?contactoId=123&fecha=2024-10-10", nil)
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
	t.Cleanup(resetReservaMocks)
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

func TestReservaGetAllSuccess(t *testing.T) {
	originalZone := database.BogotaZone
	database.BogotaZone = time.UTC
	t.Cleanup(func() { database.BogotaZone = originalZone })
	hora, _ := time.Parse(time.RFC3339, "2024-01-01T12:00:00Z")
	db := []models.Reserva{{
		PK_ID_RESERVA: int64(1),
		CREATED_AT:    time.Now(),
		UPDATED_AT:    time.Now(),
		FECHA:         time.Now(),
		HORA:          hora,
	}}
	ormNew = func() orm.Ormer { return nil }
	queryAllReservas = func(o orm.Ormer, reservas *[]models.Reserva) (int64, error) {
		*reservas = append(*reservas, db...)
		return int64(len(db)), nil
	}
	t.Cleanup(resetReservaMocks)
	r := httptest.NewRequest(http.MethodGet, "/reservas", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ReservaController{}
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}
	c.GetAll()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestReservaGetByIdScenarios(t *testing.T) {
	originalZone := database.BogotaZone
	database.BogotaZone = time.UTC
	t.Cleanup(func() { database.BogotaZone = originalZone })
	hora, _ := time.Parse(time.RFC3339, "2024-01-01T12:00:00Z")
	db := map[int64]models.Reserva{1: {
		PK_ID_RESERVA: int64(1),
		FECHA:         time.Now(),
		CREATED_AT:    time.Now(),
		UPDATED_AT:    time.Now(),
		HORA:          hora,
	}}
	ormNew = func() orm.Ormer { return nil }
	readReserva = func(o orm.Ormer, r *models.Reserva) error {
		res, ok := db[r.PK_ID_RESERVA]
		if !ok {
			return orm.ErrNoRows
		}
		*r = res
		return nil
	}
	t.Cleanup(resetReservaMocks)
	r := httptest.NewRequest(http.MethodGet, "/reservas/search?id=2", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ReservaController{}
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}
	c.GetById()
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "Reserva no encontrada") {
		t.Fatalf("not found failed")
	}
	r = httptest.NewRequest(http.MethodGet, "/reservas/search?id=1", nil)
	w = httptest.NewRecorder()
	ctx = context.NewContext()
	ctx.Reset(w, r)
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}
	c.GetById()
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "\"reservaId\":1") {
		t.Fatalf("success failed")
	}
}

func TestReservaPostMissingFecha(t *testing.T) {
	t.Cleanup(resetReservaMocks)
	payload := map[string]interface{}{
		"horaReserva":   "12:00:00",
		"personas":      2,
		"contactoId":    123,
		"restauranteId": 1,
	}
	body, _ := json.Marshal(payload)
	r := httptest.NewRequest(http.MethodPost, "/reservas", bytes.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = body
	c := ReservaController{}
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}
	c.Post()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestReservaPostInvalidHora(t *testing.T) {
	t.Cleanup(resetReservaMocks)
	payload := map[string]interface{}{
		"fechaReserva":  "2024-01-01",
		"horaReserva":   "99:99:99",
		"personas":      2,
		"contactoId":    123,
		"restauranteId": 1,
	}
	body, _ := json.Marshal(payload)
	r := httptest.NewRequest(http.MethodPost, "/reservas", bytes.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = body
	c := ReservaController{}
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}

	c.Post()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestReservaPostMissingPersonas(t *testing.T) {
	t.Cleanup(resetReservaMocks)
	payload := map[string]interface{}{
		"fechaReserva":  "2024-01-01",
		"horaReserva":   "12:00:00",
		"contactoId":    123,
		"restauranteId": 1,
	}
	body, _ := json.Marshal(payload)
	r := httptest.NewRequest(http.MethodPost, "/reservas", bytes.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = body
	c := ReservaController{}
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}

	c.Post()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestReservaPostInvalidEstado(t *testing.T) {
	t.Cleanup(resetReservaMocks)
	payload := map[string]interface{}{
		"fechaReserva":  "2024-01-01",
		"horaReserva":   "12:00:00",
		"personas":      2,
		"contactoId":    123,
		"restauranteId": 1,
		"estadoReserva": "DESCONOCIDO",
	}
	body, _ := json.Marshal(payload)
	r := httptest.NewRequest(http.MethodPost, "/reservas", bytes.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = body
	c := ReservaController{}
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}

	c.Post()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestReservaPostInvalidContacto(t *testing.T) {
	t.Cleanup(resetReservaMocks)
	payload := map[string]interface{}{
		"fechaReserva":  "2024-01-01",
		"horaReserva":   "12:00:00",
		"personas":      2,
		"contactoId":    "abc",
		"restauranteId": 1,
	}
	body, _ := json.Marshal(payload)
	r := httptest.NewRequest(http.MethodPost, "/reservas", bytes.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = body
	c := ReservaController{}
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}

	c.Post()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestReservaPostInvalidRestaurante(t *testing.T) {
	t.Cleanup(resetReservaMocks)
	payload := map[string]interface{}{
		"fechaReserva":  "2024-01-01",
		"horaReserva":   "12:00:00",
		"personas":      2,
		"contactoId":    123,
		"restauranteId": "abc",
	}
	body, _ := json.Marshal(payload)
	r := httptest.NewRequest(http.MethodPost, "/reservas", bytes.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = body
	c := ReservaController{}
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}

	c.Post()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestReservaPostInvalidReservaContacto(t *testing.T) {
	t.Cleanup(resetReservaMocks)
	payload := map[string]interface{}{
		"fechaReserva":      "2024-01-01",
		"horaReserva":       "12:00:00",
		"personas":          2,
		"contactoId":        123,
		"restauranteId":     1,
		"documentoContacto": 1,
		"documentoCliente":  2,
	}
	body, _ := json.Marshal(payload)
	r := httptest.NewRequest(http.MethodPost, "/reservas", bytes.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = body
	c := ReservaController{}
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}

	c.Post()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestReservaPostInsertError(t *testing.T) {
	ormNew = func() orm.Ormer { return nil }
	insertReserva = func(o orm.Ormer, r *models.Reserva) (int64, error) { return 0, errors.New("db") }
	// Mockear funciones de contacto
	queryReservaContactoByDocumento = func(o orm.Ormer, documento int64, rc *models.ReservaContacto) error {
		return orm.ErrNoRows // No encontrado, para que cree uno nuevo
	}
	insertReservaContacto = func(o orm.Ormer, rc *models.ReservaContacto) (int64, error) {
		return 1, nil
	}
	t.Cleanup(resetReservaMocks)
	payload := map[string]interface{}{
		"fechaReserva":      "2024-01-01",
		"horaReserva":       "12:00:00",
		"personas":          2,
		"contactoId":        123,
		"restauranteId":     1,
		"documentoContacto": 123456789,
		"nombreCompleto":    "Test User",
	}
	body, _ := json.Marshal(payload)
	r := httptest.NewRequest(http.MethodPost, "/reservas", bytes.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = body
	c := ReservaController{}
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}

	c.Post()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestReservaPostSuccess(t *testing.T) {
	db := make([]models.Reserva, 0)
	ormNew = func() orm.Ormer { return nil }
	insertReserva = func(o orm.Ormer, r *models.Reserva) (int64, error) {
		db = append(db, *r)
		return 1, nil
	}
	// Mockear funciones de contacto
	queryReservaContactoByDocumento = func(o orm.Ormer, documento int64, rc *models.ReservaContacto) error {
		return orm.ErrNoRows // No encontrado, para que cree uno nuevo
	}
	insertReservaContacto = func(o orm.Ormer, rc *models.ReservaContacto) (int64, error) {
		return 1, nil
	}
	t.Cleanup(resetReservaMocks)
	payload := map[string]interface{}{
		"fechaReserva":      "2024-01-01",
		"horaReserva":       "12:00:00",
		"personas":          2,
		"contactoId":        123,
		"restauranteId":     1,
		"estadoReserva":     models.EstadoReservaConfirmada,
		"indicaciones":      "Ninguna",
		"createdBy":         "admin",
		"nombreCompleto":    "John Doe",
		"telefono":          "123",
		"documentoContacto": 999,
	}
	body, _ := json.Marshal(payload)
	r := httptest.NewRequest(http.MethodPost, "/reservas", bytes.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = body
	c := ReservaController{}
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}

	c.Post()
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	var resp struct {
		Data map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if createdBy, ok := resp.Data["createdBy"].(string); !ok || createdBy != "admin" {
		t.Fatalf("expected createdBy admin, got %v", resp.Data["createdBy"])
	}
}

func TestReservaPutScenarios(t *testing.T) {
	hora, _ := time.Parse("15:04:05", "12:00:00")
	db := map[int64]models.Reserva{1: {
		PK_ID_RESERVA: int64(1),
		FECHA:         time.Now(),
		CREATED_AT:    time.Now(),
		UPDATED_AT:    time.Now(),
		HORA:          hora,
	}}
	ormNew = func() orm.Ormer { return nil }
	readReserva = func(o orm.Ormer, r *models.Reserva) error {
		res, ok := db[r.PK_ID_RESERVA]
		if !ok {
			return orm.ErrNoRows
		}
		*r = res
		return nil
	}
	updateReserva = func(o orm.Ormer, r *models.Reserva, cols ...string) (int64, error) {
		db[r.PK_ID_RESERVA] = *r
		return 1, nil
	}
	t.Cleanup(resetReservaMocks)

	r := httptest.NewRequest(http.MethodPut, "/reservas?id=2", strings.NewReader("{}"))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ReservaController{}
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}

	c.Put()
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "Reserva no encontrada") {
		t.Fatalf("not found failed")
	}

	r = httptest.NewRequest(http.MethodPut, "/reservas?id=1", strings.NewReader("{invalid"))
	w = httptest.NewRecorder()
	ctx = context.NewContext()
	ctx.Reset(w, r)
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}
	c.Put()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}

	payload := map[string]interface{}{"fechaReserva": "2024-13-01", "horaReserva": "12:00:00", "personas": 2, "contactoId": 123, "restauranteId": 1}
	body, _ := json.Marshal(payload)
	r = httptest.NewRequest(http.MethodPut, "/reservas?id=1", bytes.NewReader(body))
	w = httptest.NewRecorder()
	ctx = context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = body
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}
	c.Put()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}

	payload["fechaReserva"] = "2024-01-01"
	payload["horaReserva"] = "99:99:99"
	body, _ = json.Marshal(payload)
	r = httptest.NewRequest(http.MethodPut, "/reservas?id=1", bytes.NewReader(body))
	w = httptest.NewRecorder()
	ctx = context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = body
	c.Ctx = ctx
	c.Put()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}

	payload["horaReserva"] = "12:00:00"
	payload["contactoId"] = "abc"
	body, _ = json.Marshal(payload)
	r = httptest.NewRequest(http.MethodPut, "/reservas?id=1", bytes.NewReader(body))
	w = httptest.NewRecorder()
	ctx = context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = body
	c.Ctx = ctx
	c.Put()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}

	payload["contactoId"] = 123
	payload["restauranteId"] = "abc"
	body, _ = json.Marshal(payload)
	r = httptest.NewRequest(http.MethodPut, "/reservas?id=1", bytes.NewReader(body))
	w = httptest.NewRecorder()
	ctx = context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = body
	c.Ctx = ctx
	c.Put()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}

	payload["restauranteId"] = 1
	body, _ = json.Marshal(payload)
	updateReserva = func(o orm.Ormer, r *models.Reserva, cols ...string) (int64, error) { return 0, errors.New("db") }
	r = httptest.NewRequest(http.MethodPut, "/reservas?id=1", bytes.NewReader(body))
	w = httptest.NewRecorder()
	ctx = context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = body
	c.Ctx = ctx
	c.Put()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}

	updateReserva = func(o orm.Ormer, r *models.Reserva, cols ...string) (int64, error) {
		db[r.PK_ID_RESERVA] = *r
		return 1, nil
	}
	payload = map[string]interface{}{
		"fechaReserva":      "2024-01-02",
		"horaReserva":       "13:00:00",
		"personas":          3,
		"estadoReserva":     models.EstadoReservaConfirmada,
		"indicaciones":      "OK",
		"updatedBy":         "admin",
		"nombreCompleto":    "John",
		"telefono":          "123",
		"contactoId":        123,
		"restauranteId":     1,
		"documentoContacto": 777,
	}
	body, _ = json.Marshal(payload)
	r = httptest.NewRequest(http.MethodPut, "/reservas?id=1", bytes.NewReader(body))
	w = httptest.NewRecorder()
	ctx = context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = body
	c.Ctx = ctx
	c.Put()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp struct {
		Data map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if updatedBy, ok := resp.Data["updatedBy"].(string); !ok || updatedBy != "admin" {
		t.Fatalf("expected updatedBy admin, got %v", resp.Data["updatedBy"])
	}
}

func TestReservaGetByParameterSuccess(t *testing.T) {
	originalZone := database.BogotaZone
	database.BogotaZone = time.UTC
	t.Cleanup(func() { database.BogotaZone = originalZone })
	hora, _ := time.Parse(time.RFC3339, "2024-01-01T12:00:00Z")
	db := []models.Reserva{{
		PK_ID_RESERVA:  int64(1),
		PK_ID_CONTACTO: &models.ReservaContacto{PKIDContacto: 123},
		FECHA:          time.Now(),
		CREATED_AT:     time.Now(),
		UPDATED_AT:     time.Now(),
		HORA:           hora,
	}}
	ormNew = func() orm.Ormer { return nil }
	queryReservasByParam = func(o orm.Ormer, contacto int64, fecha time.Time, useContacto, useFecha bool, reservas *[]models.Reserva) (int64, error) {
		*reservas = append(*reservas, db...)
		return int64(len(db)), nil
	}
	t.Cleanup(resetReservaMocks)
	r := httptest.NewRequest(http.MethodGet, "/reservas/parameter?contactoId=123&fecha=2024-01-01", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ReservaController{}
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}

	c.GetByParameter()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestReservaGetByParameterNoResults(t *testing.T) {
	originalZone := database.BogotaZone
	database.BogotaZone = time.UTC
	t.Cleanup(func() { database.BogotaZone = originalZone })
	ormNew = func() orm.Ormer { return nil }
	queryReservasByParam = func(o orm.Ormer, contacto int64, fecha time.Time, useContacto, useFecha bool, reservas *[]models.Reserva) (int64, error) {
		return 0, nil
	}
	t.Cleanup(resetReservaMocks)
	r := httptest.NewRequest(http.MethodGet, "/reservas/parameter?contactoId=123&fecha=2024-01-01", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ReservaController{}
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}

	c.GetByParameter()
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "No se encontraron reservas") {
		t.Fatalf("expected no results, got status %d body %s", w.Code, w.Body.String())
	}
}

func TestReservaDeleteScenarios(t *testing.T) {
	db := map[int64]models.Reserva{1: {PK_ID_RESERVA: int64(1)}}
	ormNew = func() orm.Ormer { return nil }
	readReserva = func(o orm.Ormer, r *models.Reserva) error {
		res, ok := db[r.PK_ID_RESERVA]
		if !ok {
			return orm.ErrNoRows
		}
		*r = res
		return nil
	}
	updateReserva = func(o orm.Ormer, r *models.Reserva, cols ...string) (int64, error) {
		db[r.PK_ID_RESERVA] = *r
		return 1, nil
	}
	t.Cleanup(resetReservaMocks)

	r := httptest.NewRequest(http.MethodDelete, "/reservas?id=2", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ReservaController{}
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}

	c.Delete()
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "Reserva no encontrada") {
		t.Fatalf("not found failed")
	}

	updateReserva = func(o orm.Ormer, r *models.Reserva, cols ...string) (int64, error) { return 0, errors.New("db") }
	r = httptest.NewRequest(http.MethodDelete, "/reservas?id=1", nil)
	w = httptest.NewRecorder()
	ctx = context.NewContext()
	ctx.Reset(w, r)
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}
	c.Delete()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}

	updateReserva = func(o orm.Ormer, r *models.Reserva, cols ...string) (int64, error) {
		db[r.PK_ID_RESERVA] = *r
		return 1, nil
	}
	r = httptest.NewRequest(http.MethodDelete, "/reservas?id=1", nil)
	w = httptest.NewRecorder()
	ctx = context.NewContext()
	ctx.Reset(w, r)
	c.Ctx = ctx
	c.Delete()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestReservaPutInvalidEstadoIgnored(t *testing.T) {
	pend := models.EstadoReservaPendiente
	db := map[int64]models.Reserva{1: {PK_ID_RESERVA: int64(1), ESTADO_RESERVA: &pend}}
	ormNew = func() orm.Ormer { return nil }
	readReserva = func(o orm.Ormer, r *models.Reserva) error {
		if res, ok := db[r.PK_ID_RESERVA]; ok {
			*r = res
			return nil
		}
		return orm.ErrNoRows
	}
	var updated models.Reserva
	updateReserva = func(o orm.Ormer, r *models.Reserva, cols ...string) (int64, error) {
		updated = *r
		return 1, nil
	}
	t.Cleanup(resetReservaMocks)

	payload := map[string]interface{}{
		"estadoReserva": "invalido",
		"fechaReserva":  "2024-01-01",
		"horaReserva":   "12:00:00",
		"personas":      2,
		"contactoId":    123,
		"restauranteId": 1,
	}
	body, _ := json.Marshal(payload)
	r := httptest.NewRequest(http.MethodPut, "/reservas?id=1", bytes.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = body
	c := ReservaController{}
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}

	c.Put()

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if updated.ESTADO_RESERVA == nil || *updated.ESTADO_RESERVA != pend {
		t.Fatalf("estado should remain %s, got %v", pend, updated.ESTADO_RESERVA)
	}
}

func TestReservaPutInvalidReservaContacto(t *testing.T) {
	ormNew = func() orm.Ormer { return nil }
	readReserva = func(o orm.Ormer, r *models.Reserva) error { return nil }
	t.Cleanup(resetReservaMocks)
	payload := map[string]interface{}{
		"fechaReserva":      "2024-01-01",
		"horaReserva":       "12:00:00",
		"personas":          2,
		"contactoId":        123,
		"restauranteId":     1,
		"documentoContacto": 1,
		"documentoCliente":  2,
	}
	body, _ := json.Marshal(payload)
	r := httptest.NewRequest(http.MethodPut, "/reservas?id=1", bytes.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = body
	c := ReservaController{}
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}

	c.Put()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
