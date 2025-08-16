package controllers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"restaurante/database"
	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
)

func TestCambiosHorarioGetAllWithoutDB(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/cambios_horario", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := CambiosHorarioController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetAll()

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al obtener cambios de horario") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestCambiosHorarioPostInvalidJSON(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/cambios_horario", strings.NewReader("notjson"))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := CambiosHorarioController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestCambiosHorarioGetByCurrentDateWithoutDB(t *testing.T) {
	database.BogotaZone = time.Local
	r := httptest.NewRequest(http.MethodGet, "/cambios_horario/actual", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := CambiosHorarioController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetByCurrentDate()

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al consultar cambios de horario") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestCambiosHorarioGetByCurrentDateNotFound(t *testing.T) {
	database.BogotaZone = time.Local
	original := queryCambioHorarioByDate
	queryCambioHorarioByDate = func(o orm.Ormer, date string, ch *models.CambiosHorario) error {
		return orm.ErrNoRows
	}
	defer func() { queryCambioHorarioByDate = original }()

	r := httptest.NewRequest(http.MethodGet, "/cambios_horario/actual", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := CambiosHorarioController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetByCurrentDate()

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "No hay cambios de horario para la fecha actual") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestCambiosHorarioPostMissingFecha(t *testing.T) {
	body := `{"abierto":true,"horaApertura":"08:00:00","horaCierre":"17:00:00"}`
	r := httptest.NewRequest(http.MethodPost, "/cambios_horario", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := CambiosHorarioController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "FECHA es obligatorio") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestCambiosHorarioPostMissingHoraApertura(t *testing.T) {
	body := `{"fechaCambioHorario":"2024-10-10","abierto":true,"horaCierre":"17:00:00"}`
	r := httptest.NewRequest(http.MethodPost, "/cambios_horario", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := CambiosHorarioController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "HORA_APERTURA es obligatorio") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestCambiosHorarioPostAbiertoFalse(t *testing.T) {
	body := `{"fechaCambioHorario":"2024-10-10","abierto":false}`
	r := httptest.NewRequest(http.MethodPost, "/cambios_horario", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := CambiosHorarioController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al crear el cambio de horario") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestCambiosHorarioPutInvalidID(t *testing.T) {
	r := httptest.NewRequest(http.MethodPut, "/cambios_horario", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := CambiosHorarioController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Put()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestCambiosHorarioPutInvalidJSON(t *testing.T) {
	r := httptest.NewRequest(http.MethodPut, "/cambios_horario?id=1", strings.NewReader("notjson"))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte("notjson")
	c := CambiosHorarioController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Put()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestCambiosHorarioDeleteInvalidID(t *testing.T) {
	r := httptest.NewRequest(http.MethodDelete, "/cambios_horario", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := CambiosHorarioController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Delete()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestCambiosHorarioGetAllSuccess(t *testing.T) {
	original := queryAllCambiosHorario
	queryAllCambiosHorario = func(o orm.Ormer, horarios *[]models.CambiosHorario) (int64, error) {
		apertura, _ := time.Parse("15:04:05", "08:00:00")
		cierre, _ := time.Parse("15:04:05", "17:00:00")
		*horarios = []models.CambiosHorario{{
			PK_ID_CAMBIO_HORARIO: 1,
			FECHA:                time.Date(2024, 10, 10, 0, 0, 0, 0, time.UTC),
			HORA_APERTURA:        &apertura,
			HORA_CIERRE:          &cierre,
			ABIERTO:              true,
		}}
		return 1, nil
	}
	defer func() { queryAllCambiosHorario = original }()

	r := httptest.NewRequest(http.MethodGet, "/cambios_horario", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := CambiosHorarioController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetAll()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Cambios de horario obtenidos correctamente") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestCambiosHorarioGetByCurrentDateSuccess(t *testing.T) {
	database.BogotaZone = time.UTC
	original := queryCambioHorarioByDate
	queryCambioHorarioByDate = func(o orm.Ormer, date string, ch *models.CambiosHorario) error {
		apertura, _ := time.Parse("15:04:05", "09:00:00")
		cierre, _ := time.Parse("15:04:05", "18:00:00")
		ch.PK_ID_CAMBIO_HORARIO = 2
		ch.FECHA = time.Date(2024, 10, 10, 0, 0, 0, 0, time.UTC)
		ch.HORA_APERTURA = &apertura
		ch.HORA_CIERRE = &cierre
		ch.ABIERTO = true
		return nil
	}
	defer func() { queryCambioHorarioByDate = original }()

	r := httptest.NewRequest(http.MethodGet, "/cambios_horario/actual", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := CambiosHorarioController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetByCurrentDate()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Cambio de horario encontrado para la fecha actual") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestCambiosHorarioPostInvalidFecha(t *testing.T) {
	body := `{"fechaCambioHorario":"2024/10/10","abierto":true,"horaApertura":"08:00:00","horaCierre":"17:00:00"}`
	r := httptest.NewRequest(http.MethodPost, "/cambios_horario", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := CambiosHorarioController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestCambiosHorarioPostMissingAbierto(t *testing.T) {
	body := `{"fechaCambioHorario":"2024-10-10"}`
	r := httptest.NewRequest(http.MethodPost, "/cambios_horario", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := CambiosHorarioController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestCambiosHorarioPostInvalidHoraApertura(t *testing.T) {
	body := `{"fechaCambioHorario":"2024-10-10","abierto":true,"horaApertura":"invalid","horaCierre":"17:00:00"}`
	r := httptest.NewRequest(http.MethodPost, "/cambios_horario", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := CambiosHorarioController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestCambiosHorarioPostMissingHoraCierre(t *testing.T) {
	body := `{"fechaCambioHorario":"2024-10-10","abierto":true,"horaApertura":"08:00:00"}`
	r := httptest.NewRequest(http.MethodPost, "/cambios_horario", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := CambiosHorarioController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestCambiosHorarioPostInvalidHoraCierre(t *testing.T) {
	body := `{"fechaCambioHorario":"2024-10-10","abierto":true,"horaApertura":"08:00:00","horaCierre":"invalid"}`
	r := httptest.NewRequest(http.MethodPost, "/cambios_horario", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := CambiosHorarioController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestCambiosHorarioPostSuccess(t *testing.T) {
	original := insertCambioHorario
	insertCambioHorario = func(o orm.Ormer, h *models.CambiosHorario) (int64, error) {
		h.PK_ID_CAMBIO_HORARIO = 1
		return 1, nil
	}
	defer func() { insertCambioHorario = original }()

	body := `{"fechaCambioHorario":"2024-10-10","abierto":true,"horaApertura":"08:00:00","horaCierre":"17:00:00"}`
	r := httptest.NewRequest(http.MethodPost, "/cambios_horario", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := CambiosHorarioController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Cambio de horario creado correctamente") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestCambiosHorarioPostAbiertoFalseSuccess(t *testing.T) {
	original := insertCambioHorario
	insertCambioHorario = func(o orm.Ormer, h *models.CambiosHorario) (int64, error) {
		h.PK_ID_CAMBIO_HORARIO = 2
		return 1, nil
	}
	defer func() { insertCambioHorario = original }()

	body := `{"fechaCambioHorario":"2024-10-10","abierto":false}`
	r := httptest.NewRequest(http.MethodPost, "/cambios_horario", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := CambiosHorarioController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", w.Code)
	}
}

func TestCambiosHorarioPutNotFound(t *testing.T) {
	original := queryCambioHorarioByID
	queryCambioHorarioByID = func(o orm.Ormer, id int64, ch *models.CambiosHorario) error {
		return orm.ErrNoRows
	}
	defer func() { queryCambioHorarioByID = original }()

	r := httptest.NewRequest(http.MethodPut, "/cambios_horario?id=1", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(`{}`)
	c := CambiosHorarioController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Put()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Cambio de horario no encontrado") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestCambiosHorarioPutDBError(t *testing.T) {
	original := queryCambioHorarioByID
	queryCambioHorarioByID = func(o orm.Ormer, id int64, ch *models.CambiosHorario) error {
		return errors.New("db error")
	}
	defer func() { queryCambioHorarioByID = original }()

	r := httptest.NewRequest(http.MethodPut, "/cambios_horario?id=1", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(`{}`)
	c := CambiosHorarioController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Put()

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestCambiosHorarioPutInvalidHoraApertura(t *testing.T) {
	original := queryCambioHorarioByID
	queryCambioHorarioByID = func(o orm.Ormer, id int64, ch *models.CambiosHorario) error { return nil }
	defer func() { queryCambioHorarioByID = original }()

	body := `{"abierto":true,"horaApertura":"bad"}`
	r := httptest.NewRequest(http.MethodPut, "/cambios_horario?id=1", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := CambiosHorarioController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Put()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestCambiosHorarioPutInvalidHoraCierre(t *testing.T) {
	original := queryCambioHorarioByID
	queryCambioHorarioByID = func(o orm.Ormer, id int64, ch *models.CambiosHorario) error { return nil }
	defer func() { queryCambioHorarioByID = original }()

	body := `{"abierto":true,"horaApertura":"08:00:00","horaCierre":"bad"}`
	r := httptest.NewRequest(http.MethodPut, "/cambios_horario?id=1", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := CambiosHorarioController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Put()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestCambiosHorarioPutUpdateError(t *testing.T) {
	originalQuery := queryCambioHorarioByID
	queryCambioHorarioByID = func(o orm.Ormer, id int64, ch *models.CambiosHorario) error { return nil }
	originalUpdate := updateCambioHorario
	updateCambioHorario = func(o orm.Ormer, ch *models.CambiosHorario) (int64, error) {
		return 0, errors.New("update error")
	}
	defer func() {
		queryCambioHorarioByID = originalQuery
		updateCambioHorario = originalUpdate
	}()

	body := `{"abierto":true,"horaApertura":"08:00:00","horaCierre":"17:00:00"}`
	r := httptest.NewRequest(http.MethodPut, "/cambios_horario?id=1", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := CambiosHorarioController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Put()

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestCambiosHorarioPutSuccess(t *testing.T) {
	originalQuery := queryCambioHorarioByID
	queryCambioHorarioByID = func(o orm.Ormer, id int64, ch *models.CambiosHorario) error { return nil }
	originalUpdate := updateCambioHorario
	updateCambioHorario = func(o orm.Ormer, ch *models.CambiosHorario) (int64, error) { return 1, nil }
	defer func() {
		queryCambioHorarioByID = originalQuery
		updateCambioHorario = originalUpdate
	}()

	body := `{"abierto":true,"horaApertura":"08:00:00","horaCierre":"17:00:00"}`
	r := httptest.NewRequest(http.MethodPut, "/cambios_horario?id=1", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := CambiosHorarioController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Put()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestCambiosHorarioDeleteNotFound(t *testing.T) {
	original := deleteCambioHorarioByID
	deleteCambioHorarioByID = func(o orm.Ormer, id int64) (int64, error) { return 0, nil }
	defer func() { deleteCambioHorarioByID = original }()

	r := httptest.NewRequest(http.MethodDelete, "/cambios_horario?id=1", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := CambiosHorarioController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Delete()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Cambio de horario no encontrado") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestCambiosHorarioDeleteDBError(t *testing.T) {
	original := deleteCambioHorarioByID
	deleteCambioHorarioByID = func(o orm.Ormer, id int64) (int64, error) { return 0, errors.New("db") }
	defer func() { deleteCambioHorarioByID = original }()

	r := httptest.NewRequest(http.MethodDelete, "/cambios_horario?id=1", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := CambiosHorarioController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Delete()

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestCambiosHorarioDeleteSuccess(t *testing.T) {
	original := deleteCambioHorarioByID
	deleteCambioHorarioByID = func(o orm.Ormer, id int64) (int64, error) { return 1, nil }
	defer func() { deleteCambioHorarioByID = original }()

	r := httptest.NewRequest(http.MethodDelete, "/cambios_horario?id=1", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := CambiosHorarioController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Delete()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}
