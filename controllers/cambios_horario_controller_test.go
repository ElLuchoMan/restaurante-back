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

func setupCtx(method, url string, body string) (*CambiosHorarioController, *httptest.ResponseRecorder) {
	r := httptest.NewRequest(method, url, strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	if body != "" {
		ctx.Input.RequestBody = []byte(body)
	}
	c := &CambiosHorarioController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	return c, w
}

func TestCambiosHorario_GetAll_DBError(t *testing.T) {
	orig := queryAllCambiosHorario
	queryAllCambiosHorario = func(o orm.Ormer, horarios *[]models.CambiosHorario) (int64, error) {
		return 0, errors.New("db fail")
	}
	defer func() { queryAllCambiosHorario = orig }()

	c, w := setupCtx(http.MethodGet, "/cambios_horario", "")
	c.GetAll()

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al obtener cambios de horario") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestCambiosHorario_GetAll_Success(t *testing.T) {
	orig := queryAllCambiosHorario
	queryAllCambiosHorario = func(o orm.Ormer, horarios *[]models.CambiosHorario) (int64, error) {
		*horarios = []models.CambiosHorario{
			{PK_ID_CAMBIO_HORARIO: 1, FECHA: time.Date(2024, 10, 10, 0, 0, 0, 0, time.UTC), ABIERTO: true},
			{PK_ID_CAMBIO_HORARIO: 2, FECHA: time.Date(2024, 10, 11, 0, 0, 0, 0, time.UTC), ABIERTO: false},
		}
		return 2, nil
	}
	defer func() { queryAllCambiosHorario = orig }()

	c, w := setupCtx(http.MethodGet, "/cambios_horario", "")
	c.GetAll()

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"cambioHorarioId":1`) {
		t.Errorf("response should include first id, got %s", w.Body.String())
	}
}

func TestCambiosHorario_GetByCurrentDate_DBError(t *testing.T) {
	database.BogotaZone = time.Local
	orig := queryCambioHorarioByDate
	queryCambioHorarioByDate = func(o orm.Ormer, date string, ch *models.CambiosHorario) error {
		return errors.New("db error")
	}
	defer func() { queryCambioHorarioByDate = orig }()

	c, w := setupCtx(http.MethodGet, "/cambios_horario/actual", "")
	c.GetByCurrentDate()

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al consultar cambios de horario") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestCambiosHorario_GetByCurrentDate_NotFound(t *testing.T) {
	database.BogotaZone = time.Local
	orig := queryCambioHorarioByDate
	queryCambioHorarioByDate = func(o orm.Ormer, date string, ch *models.CambiosHorario) error {
		return orm.ErrNoRows
	}
	defer func() { queryCambioHorarioByDate = orig }()

	c, w := setupCtx(http.MethodGet, "/cambios_horario/actual", "")
	c.GetByCurrentDate()

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "No hay cambios de horario para la fecha actual") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestCambiosHorario_Post_InvalidJSON(t *testing.T) {
	c, w := setupCtx(http.MethodPost, "/cambios_horario", "notjson")
	c.Post()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCambiosHorario_Post_MissingFecha(t *testing.T) {
	body := `{"abierto":true,"horaApertura":"08:00:00","horaCierre":"17:00:00"}`
	c, w := setupCtx(http.MethodPost, "/cambios_horario", body)
	c.Post()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "El campo FECHA es obligatorio") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestCambiosHorario_Post_MissingHoraApertura(t *testing.T) {
	body := `{"fechaCambioHorario":"2024-10-10","abierto":true,"horaCierre":"17:00:00"}`
	c, w := setupCtx(http.MethodPost, "/cambios_horario", body)
	c.Post()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "HORA_APERTURA es obligatorio") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestCambiosHorario_Post_AbiertoFalse_SuccessInsert(t *testing.T) {
	orig := insertCambioHorario
	insertCambioHorario = func(o orm.Ormer, horario *models.CambiosHorario) (int64, error) {
		horario.PK_ID_CAMBIO_HORARIO = 99
		return 1, nil
	}
	defer func() { insertCambioHorario = orig }()

	body := `{"fechaCambioHorario":"2024-10-10","abierto":false}`
	c, w := setupCtx(http.MethodPost, "/cambios_horario", body)
	c.Post()

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"cambioHorarioId":99`) {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestCambiosHorario_Post_InsertFail(t *testing.T) {
	orig := insertCambioHorario
	insertCambioHorario = func(o orm.Ormer, horario *models.CambiosHorario) (int64, error) {
		return 0, errors.New("insert fail")
	}
	defer func() { insertCambioHorario = orig }()

	body := `{"fechaCambioHorario":"2024-10-10","abierto":false}`
	c, w := setupCtx(http.MethodPost, "/cambios_horario", body)
	c.Post()

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestCambiosHorario_Put_InvalidID(t *testing.T) {
	c, w := setupCtx(http.MethodPut, "/cambios_horario", "")
	c.Put()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCambiosHorario_Put_InvalidJSON(t *testing.T) {
	c, w := setupCtx(http.MethodPut, "/cambios_horario?id=1", "notjson")
	c.Put()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCambiosHorario_Put_NotFoundPath(t *testing.T) {
	origFind := queryCambioHorarioByID
	queryCambioHorarioByID = func(o orm.Ormer, id int64, horario *models.CambiosHorario) error {
		return orm.ErrNoRows
	}
	defer func() { queryCambioHorarioByID = origFind }()

	c, w := setupCtx(http.MethodPut, "/cambios_horario?id=1", `{"abierto":true}`)
	c.Put()

	// Se respeta tu semántica: 200 con Code=404
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
       if !strings.Contains(w.Body.String(), `"code":404`) || !strings.Contains(w.Body.String(), `"message":"Cambio de horario no encontrado"`) {
               t.Errorf("unexpected body: %s", w.Body.String())
       }
}

func TestCambiosHorario_Put_UpdateSuccess(t *testing.T) {
	origFind := queryCambioHorarioByID
	origUpd := updateCambioHorario
	queryCambioHorarioByID = func(o orm.Ormer, id int64, horario *models.CambiosHorario) error {
		*horario = models.CambiosHorario{
			PK_ID_CAMBIO_HORARIO: 7,
			FECHA:                time.Date(2024, 10, 10, 0, 0, 0, 0, time.UTC),
			ABIERTO:              false,
		}
		return nil
	}
	updateCambioHorario = func(o orm.Ormer, horario *models.CambiosHorario) (int64, error) {
		return 1, nil
	}
	defer func() { queryCambioHorarioByID = origFind; updateCambioHorario = origUpd }()

	body := `{"fechaCambioHorario":"2024-10-11","abierto":true,"horaApertura":"08:00:00","horaCierre":"17:00:00"}`
	c, w := setupCtx(http.MethodPut, "/cambios_horario?id=7", body)
	c.Put()

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"cambioHorarioId":7`) {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestCambiosHorario_Delete_InvalidID(t *testing.T) {
	c, w := setupCtx(http.MethodDelete, "/cambios_horario", "")
	c.Delete()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCambiosHorario_Delete_NotFound(t *testing.T) {
	origDel := deleteCambioHorarioByID
	deleteCambioHorarioByID = func(o orm.Ormer, id int64) (int64, error) {
		return 0, nil
	}
	defer func() { deleteCambioHorarioByID = origDel }()

	c, w := setupCtx(http.MethodDelete, "/cambios_horario?id=123", "")
	c.Delete()

	// Tu semántica: 200 con Code=404
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
       if !strings.Contains(w.Body.String(), `"code":404`) || !strings.Contains(w.Body.String(), `"message":"Cambio de horario no encontrado"`) {
               t.Errorf("unexpected body: %s", w.Body.String())
       }
}

func TestCambiosHorario_Delete_Success(t *testing.T) {
	origDel := deleteCambioHorarioByID
	deleteCambioHorarioByID = func(o orm.Ormer, id int64) (int64, error) {
		return 1, nil
	}
	defer func() { deleteCambioHorarioByID = origDel }()

	c, w := setupCtx(http.MethodDelete, "/cambios_horario?id=1", "")
	c.Delete()

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "eliminado correctamente") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}
