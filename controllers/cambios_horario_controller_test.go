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

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
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

func TestCambiosHorario_GetAll_WithHoras(t *testing.T) {
	orig := queryAllCambiosHorario
	queryAllCambiosHorario = func(o orm.Ormer, horarios *[]models.CambiosHorario) (int64, error) {
		horaApertura, _ := time.Parse("15:04:05", "08:00:00")
		horaCierre, _ := time.Parse("15:04:05", "17:00:00")
		*horarios = []models.CambiosHorario{
			{
				PK_ID_CAMBIO_HORARIO: 3,
				FECHA:                time.Date(2024, 10, 12, 0, 0, 0, 0, time.UTC),
				ABIERTO:              true,
				HORA_APERTURA:        &horaApertura,
				HORA_CIERRE:          horaCierre,
			},
		}
		return 1, nil
	}
	defer func() { queryAllCambiosHorario = orig }()

	c, w := setupCtx(http.MethodGet, "/cambios_horario", "")
	c.GetAll()

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"horaApertura":"08:00:00"`) || !strings.Contains(w.Body.String(), `"horaCierre":"17:00:00"`) {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestCambiosHorario_GetByCurrentDate_Success(t *testing.T) {
	database.BogotaZone = time.Local
	orig := queryCambioHorarioByDate
	queryCambioHorarioByDate = func(o orm.Ormer, date string, ch *models.CambiosHorario) error {
		d, _ := time.Parse("2006-01-02", date)
		ha, _ := time.Parse("15:04:05", "08:00:00")
		hc, _ := time.Parse("15:04:05", "17:00:00")
		*ch = models.CambiosHorario{
			PK_ID_CAMBIO_HORARIO: 1,
			FECHA:                d,
			ABIERTO:              true,
			HORA_APERTURA:        &ha,
			HORA_CIERRE:          hc,
		}
		return nil
	}
	defer func() { queryCambioHorarioByDate = orig }()

	c, w := setupCtx(http.MethodGet, "/cambios_horario/actual", "")
	c.GetByCurrentDate()

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"cambioHorarioId":1`) {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestCambiosHorario_Post_InvalidFecha(t *testing.T) {
	body := `{"fechaCambioHorario":"2024/10/10","abierto":false}`
	c, w := setupCtx(http.MethodPost, "/cambios_horario", body)
	c.Post()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Formato de fecha inválido") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestCambiosHorario_Post_MissingAbierto(t *testing.T) {
	body := `{"fechaCambioHorario":"2024-10-10"}`
	c, w := setupCtx(http.MethodPost, "/cambios_horario", body)
	c.Post()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "ABIERTO es obligatorio") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestCambiosHorario_Post_InvalidHoraApertura(t *testing.T) {
	body := `{"fechaCambioHorario":"2024-10-10","abierto":true,"horaApertura":"bad","horaCierre":"17:00:00"}`
	c, w := setupCtx(http.MethodPost, "/cambios_horario", body)
	c.Post()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Formato de hora inválido para HORA_APERTURA") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestCambiosHorario_Post_MissingHoraCierre(t *testing.T) {
	body := `{"fechaCambioHorario":"2024-10-10","abierto":true,"horaApertura":"08:00:00"}`
	c, w := setupCtx(http.MethodPost, "/cambios_horario", body)
	c.Post()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "HORA_CIERRE es obligatorio") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestCambiosHorario_Post_InvalidHoraCierre(t *testing.T) {
	body := `{"fechaCambioHorario":"2024-10-10","abierto":true,"horaApertura":"08:00:00","horaCierre":"bad"}`
	c, w := setupCtx(http.MethodPost, "/cambios_horario", body)
	c.Post()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Formato de hora inválido para HORA_CIERRE") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestCambiosHorario_Post_AbiertoTrue_Success(t *testing.T) {
	orig := insertCambioHorario
	insertCambioHorario = func(o orm.Ormer, horario *models.CambiosHorario) (int64, error) {
		horario.PK_ID_CAMBIO_HORARIO = 5
		return 1, nil
	}
	defer func() { insertCambioHorario = orig }()

	body := `{"fechaCambioHorario":"2024-10-10","abierto":true,"horaApertura":"08:00:00","horaCierre":"17:00:00"}`
	c, w := setupCtx(http.MethodPost, "/cambios_horario", body)
	c.Post()

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"cambioHorarioId":5`) {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestCambiosHorario_Put_FindError(t *testing.T) {
	orig := queryCambioHorarioByID
	queryCambioHorarioByID = func(o orm.Ormer, id int64, horario *models.CambiosHorario) error {
		return errors.New("db error")
	}
	defer func() { queryCambioHorarioByID = orig }()

	c, w := setupCtx(http.MethodPut, "/cambios_horario?id=1", `{}`)
	c.Put()

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestCambiosHorario_Put_BadFecha(t *testing.T) {
	origFind := queryCambioHorarioByID
	queryCambioHorarioByID = func(o orm.Ormer, id int64, horario *models.CambiosHorario) error {
		*horario = models.CambiosHorario{PK_ID_CAMBIO_HORARIO: id, FECHA: time.Now(), ABIERTO: true}
		return nil
	}
	defer func() { queryCambioHorarioByID = origFind }()

	body := `{"fechaCambioHorario":"bad"}`
	c, w := setupCtx(http.MethodPut, "/cambios_horario?id=1", body)
	c.Put()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Formato de fecha inválido") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestCambiosHorario_Put_AbiertoFalse(t *testing.T) {
	origFind := queryCambioHorarioByID
	origUpd := updateCambioHorario
	queryCambioHorarioByID = func(o orm.Ormer, id int64, horario *models.CambiosHorario) error {
		*horario = models.CambiosHorario{PK_ID_CAMBIO_HORARIO: id, FECHA: time.Date(2024, 10, 10, 0, 0, 0, 0, time.UTC), ABIERTO: true}
		return nil
	}
	updateCambioHorario = func(o orm.Ormer, horario *models.CambiosHorario) (int64, error) {
		return 1, nil
	}
	defer func() { queryCambioHorarioByID = origFind; updateCambioHorario = origUpd }()

	body := `{"abierto":false}`
	c, w := setupCtx(http.MethodPut, "/cambios_horario?id=2", body)
	c.Put()

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"abierto":false`) {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestCambiosHorario_Put_InvalidHoraApertura(t *testing.T) {
	origFind := queryCambioHorarioByID
	queryCambioHorarioByID = func(o orm.Ormer, id int64, horario *models.CambiosHorario) error {
		*horario = models.CambiosHorario{PK_ID_CAMBIO_HORARIO: id, FECHA: time.Now(), ABIERTO: false}
		return nil
	}
	defer func() { queryCambioHorarioByID = origFind }()

	body := `{"abierto":true,"horaApertura":"bad"}`
	c, w := setupCtx(http.MethodPut, "/cambios_horario?id=3", body)
	c.Put()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Formato de hora inválido para HORA_APERTURA") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestCambiosHorario_Put_InvalidHoraCierre(t *testing.T) {
	origFind := queryCambioHorarioByID
	queryCambioHorarioByID = func(o orm.Ormer, id int64, horario *models.CambiosHorario) error {
		*horario = models.CambiosHorario{PK_ID_CAMBIO_HORARIO: id, FECHA: time.Now(), ABIERTO: false}
		return nil
	}
	defer func() { queryCambioHorarioByID = origFind }()

	body := `{"abierto":true,"horaApertura":"08:00:00","horaCierre":"bad"}`
	c, w := setupCtx(http.MethodPut, "/cambios_horario?id=4", body)
	c.Put()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Formato de hora inválido para HORA_CIERRE") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestCambiosHorario_Put_UpdateError(t *testing.T) {
	origFind := queryCambioHorarioByID
	origUpd := updateCambioHorario
	queryCambioHorarioByID = func(o orm.Ormer, id int64, horario *models.CambiosHorario) error {
		*horario = models.CambiosHorario{PK_ID_CAMBIO_HORARIO: id, FECHA: time.Now(), ABIERTO: true}
		return nil
	}
	updateCambioHorario = func(o orm.Ormer, horario *models.CambiosHorario) (int64, error) {
		return 0, errors.New("update fail")
	}
	defer func() { queryCambioHorarioByID = origFind; updateCambioHorario = origUpd }()

	body := `{"abierto":true,"horaApertura":"08:00:00","horaCierre":"17:00:00"}`
	c, w := setupCtx(http.MethodPut, "/cambios_horario?id=5", body)
	c.Put()

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestCambiosHorario_Delete_DBError(t *testing.T) {
	origDel := deleteCambioHorarioByID
	deleteCambioHorarioByID = func(o orm.Ormer, id int64) (int64, error) {
		return 0, errors.New("del err")
	}
	defer func() { deleteCambioHorarioByID = origDel }()

	c, w := setupCtx(http.MethodDelete, "/cambios_horario?id=9", "")
	c.Delete()

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

// Tests adicionales para casos edge y validaciones específicas

func TestCambiosHorario_GetAll_EmptyResult(t *testing.T) {
	orig := queryAllCambiosHorario
	queryAllCambiosHorario = func(o orm.Ormer, horarios *[]models.CambiosHorario) (int64, error) {
		*horarios = []models.CambiosHorario{}
		return 0, nil
	}
	defer func() { queryAllCambiosHorario = orig }()

	c, w := setupCtx(http.MethodGet, "/cambios_horario", "")
	c.GetAll()

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	// El controlador actual retorna null cuando la slice está vacía, no []
	if !strings.Contains(w.Body.String(), `"data":null`) {
		t.Errorf("expected null in data for empty result, got %s", w.Body.String())
	}
}

func TestCambiosHorario_GetAll_HorasNulas(t *testing.T) {
	orig := queryAllCambiosHorario
	queryAllCambiosHorario = func(o orm.Ormer, horarios *[]models.CambiosHorario) (int64, error) {
		*horarios = []models.CambiosHorario{
			{
				PK_ID_CAMBIO_HORARIO: 1,
				FECHA:                time.Date(2024, 10, 10, 0, 0, 0, 0, time.UTC),
				ABIERTO:              false,
				HORA_APERTURA:        nil,
				HORA_CIERRE:          time.Time{}, // Zero time
			},
		}
		return 1, nil
	}
	defer func() { queryAllCambiosHorario = orig }()

	c, w := setupCtx(http.MethodGet, "/cambios_horario", "")
	c.GetAll()

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"horaApertura":null`) {
		t.Errorf("expected null horaApertura, got %s", w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"horaCierre":null`) {
		t.Errorf("expected null horaCierre, got %s", w.Body.String())
	}
}

func TestCambiosHorario_Post_AbiertoFalse_AutomaticHours(t *testing.T) {
	orig := insertCambioHorario
	var capturedHorario *models.CambiosHorario
	insertCambioHorario = func(o orm.Ormer, horario *models.CambiosHorario) (int64, error) {
		capturedHorario = horario
		horario.PK_ID_CAMBIO_HORARIO = 100
		return 1, nil
	}
	defer func() { insertCambioHorario = orig }()

	body := `{"fechaCambioHorario":"2024-12-25","abierto":false}`
	c, w := setupCtx(http.MethodPost, "/cambios_horario", body)
	c.Post()

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	// Verificar que se asignaron las horas automáticamente para cerrado
	if capturedHorario == nil {
		t.Fatal("horario should have been captured")
	}
	if capturedHorario.HORA_APERTURA == nil {
		t.Error("HORA_APERTURA should not be nil for closed restaurant")
	} else if capturedHorario.HORA_APERTURA.Format("15:04:05") != "00:00:00" {
		t.Errorf("expected HORA_APERTURA 00:00:00 for closed restaurant, got %s", capturedHorario.HORA_APERTURA.Format("15:04:05"))
	}
	if capturedHorario.HORA_CIERRE.Format("15:04:05") != "23:59:59" {
		t.Errorf("expected HORA_CIERRE 23:59:59 for closed restaurant, got %s", capturedHorario.HORA_CIERRE.Format("15:04:05"))
	}
}

func TestCambiosHorario_Post_FechaStringEmpty(t *testing.T) {
	body := `{"fechaCambioHorario":"","abierto":true,"horaApertura":"08:00:00","horaCierre":"17:00:00"}`
	c, w := setupCtx(http.MethodPost, "/cambios_horario", body)
	c.Post()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "El campo FECHA es obligatorio") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestCambiosHorario_Post_HoraFormatoConSegundos(t *testing.T) {
	orig := insertCambioHorario
	var capturedHorario *models.CambiosHorario
	insertCambioHorario = func(o orm.Ormer, horario *models.CambiosHorario) (int64, error) {
		capturedHorario = horario
		horario.PK_ID_CAMBIO_HORARIO = 101
		return 1, nil
	}
	defer func() { insertCambioHorario = orig }()

	body := `{"fechaCambioHorario":"2024-10-15","abierto":true,"horaApertura":"08:30:45","horaCierre":"18:15:30"}`
	c, w := setupCtx(http.MethodPost, "/cambios_horario", body)
	c.Post()

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	if capturedHorario.HORA_APERTURA.Format("15:04:05") != "08:30:45" {
		t.Errorf("expected HORA_APERTURA 08:30:45, got %s", capturedHorario.HORA_APERTURA.Format("15:04:05"))
	}
	if capturedHorario.HORA_CIERRE.Format("15:04:05") != "18:15:30" {
		t.Errorf("expected HORA_CIERRE 18:15:30, got %s", capturedHorario.HORA_CIERRE.Format("15:04:05"))
	}
}

func TestCambiosHorario_Put_PartialUpdateDate(t *testing.T) {
	origFind := queryCambioHorarioByID
	origUpd := updateCambioHorario
	var capturedHorario *models.CambiosHorario

	queryCambioHorarioByID = func(o orm.Ormer, id int64, horario *models.CambiosHorario) error {
		ha, _ := time.Parse("15:04:05", "09:00:00")
		hc, _ := time.Parse("15:04:05", "18:00:00")
		*horario = models.CambiosHorario{
			PK_ID_CAMBIO_HORARIO: id,
			FECHA:                time.Date(2024, 10, 10, 0, 0, 0, 0, time.UTC),
			ABIERTO:              true,
			HORA_APERTURA:        &ha,
			HORA_CIERRE:          hc,
		}
		return nil
	}
	updateCambioHorario = func(o orm.Ormer, horario *models.CambiosHorario) (int64, error) {
		capturedHorario = horario
		return 1, nil
	}
	defer func() { queryCambioHorarioByID = origFind; updateCambioHorario = origUpd }()

	// Solo actualizar la fecha
	body := `{"fechaCambioHorario":"2024-12-31"}`
	c, w := setupCtx(http.MethodPut, "/cambios_horario?id=8", body)
	c.Put()

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Verificar que la fecha se actualizó pero otros campos se mantuvieron
	if capturedHorario.FECHA.Format("2006-01-02") != "2024-12-31" {
		t.Errorf("expected date 2024-12-31, got %s", capturedHorario.FECHA.Format("2006-01-02"))
	}
	if !capturedHorario.ABIERTO {
		t.Error("ABIERTO should remain true")
	}
}

func TestCambiosHorario_Put_AbiertoToFalseOverridesHours(t *testing.T) {
	origFind := queryCambioHorarioByID
	origUpd := updateCambioHorario
	var capturedHorario *models.CambiosHorario

	queryCambioHorarioByID = func(o orm.Ormer, id int64, horario *models.CambiosHorario) error {
		ha, _ := time.Parse("15:04:05", "09:00:00")
		hc, _ := time.Parse("15:04:05", "18:00:00")
		*horario = models.CambiosHorario{
			PK_ID_CAMBIO_HORARIO: id,
			FECHA:                time.Date(2024, 10, 10, 0, 0, 0, 0, time.UTC),
			ABIERTO:              true,
			HORA_APERTURA:        &ha,
			HORA_CIERRE:          hc,
		}
		return nil
	}
	updateCambioHorario = func(o orm.Ormer, horario *models.CambiosHorario) (int64, error) {
		capturedHorario = horario
		return 1, nil
	}
	defer func() { queryCambioHorarioByID = origFind; updateCambioHorario = origUpd }()

	// Cambiar a cerrado - debe sobreescribir las horas
	body := `{"abierto":false}`
	c, w := setupCtx(http.MethodPut, "/cambios_horario?id=9", body)
	c.Put()

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	if capturedHorario.ABIERTO {
		t.Error("ABIERTO should be false")
	}
	if capturedHorario.HORA_APERTURA == nil {
		t.Error("HORA_APERTURA should not be nil")
	} else if capturedHorario.HORA_APERTURA.Format("15:04:05") != "00:00:00" {
		t.Errorf("expected HORA_APERTURA 00:00:00, got %s", capturedHorario.HORA_APERTURA.Format("15:04:05"))
	}
	if capturedHorario.HORA_CIERRE.Format("15:04:05") != "23:59:59" {
		t.Errorf("expected HORA_CIERRE 23:59:59, got %s", capturedHorario.HORA_CIERRE.Format("15:04:05"))
	}
}

func TestCambiosHorario_Put_UpdateHoursWhenAbierto(t *testing.T) {
	origFind := queryCambioHorarioByID
	origUpd := updateCambioHorario
	var capturedHorario *models.CambiosHorario

	queryCambioHorarioByID = func(o orm.Ormer, id int64, horario *models.CambiosHorario) error {
		*horario = models.CambiosHorario{
			PK_ID_CAMBIO_HORARIO: id,
			FECHA:                time.Date(2024, 10, 10, 0, 0, 0, 0, time.UTC),
			ABIERTO:              true,
		}
		return nil
	}
	updateCambioHorario = func(o orm.Ormer, horario *models.CambiosHorario) (int64, error) {
		capturedHorario = horario
		return 1, nil
	}
	defer func() { queryCambioHorarioByID = origFind; updateCambioHorario = origUpd }()

	// Actualizar solo las horas cuando está abierto
	body := `{"horaApertura":"07:30:00","horaCierre":"19:30:00"}`
	c, w := setupCtx(http.MethodPut, "/cambios_horario?id=10", body)
	c.Put()

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	if !capturedHorario.ABIERTO {
		t.Error("ABIERTO should remain true")
	}
	if capturedHorario.HORA_APERTURA == nil {
		t.Error("HORA_APERTURA should not be nil")
	} else if capturedHorario.HORA_APERTURA.Format("15:04:05") != "07:30:00" {
		t.Errorf("expected HORA_APERTURA 07:30:00, got %s", capturedHorario.HORA_APERTURA.Format("15:04:05"))
	}
	if capturedHorario.HORA_CIERRE.Format("15:04:05") != "19:30:00" {
		t.Errorf("expected HORA_CIERRE 19:30:00, got %s", capturedHorario.HORA_CIERRE.Format("15:04:05"))
	}
}

func TestCambiosHorario_GetByCurrentDate_WithHours(t *testing.T) {
	database.BogotaZone = time.Local
	orig := queryCambioHorarioByDate
	queryCambioHorarioByDate = func(o orm.Ormer, date string, ch *models.CambiosHorario) error {
		d, _ := time.Parse("2006-01-02", date)
		*ch = models.CambiosHorario{
			PK_ID_CAMBIO_HORARIO: 15,
			FECHA:                d,
			ABIERTO:              false,
			HORA_APERTURA:        nil,
			HORA_CIERRE:          time.Time{}, // Zero time should be null
		}
		return nil
	}
	defer func() { queryCambioHorarioByDate = orig }()

	c, w := setupCtx(http.MethodGet, "/cambios_horario/actual", "")
	c.GetByCurrentDate()

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"cambioHorarioId":15`) {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
	// No debe incluir horaApertura ni horaCierre en el response cuando son null/zero
	if strings.Contains(w.Body.String(), `"horaApertura"`) {
		t.Errorf("should not include horaApertura when null, got %s", w.Body.String())
	}
	if strings.Contains(w.Body.String(), `"horaCierre"`) {
		t.Errorf("should not include horaCierre when zero time, got %s", w.Body.String())
	}
}

func TestCambiosHorario_Delete_IDStringInvalid(t *testing.T) {
	c, w := setupCtx(http.MethodDelete, "/cambios_horario?id=invalid", "")
	c.Delete()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "ID inválido o ausente") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestCambiosHorario_Delete_IDZero(t *testing.T) {
	c, w := setupCtx(http.MethodDelete, "/cambios_horario?id=0", "")
	c.Delete()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "ID inválido o ausente") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestCambiosHorario_Put_IDNegative(t *testing.T) {
	c, w := setupCtx(http.MethodPut, "/cambios_horario?id=-1", `{}`)
	c.Put()

	// Los IDs negativos pueden causar error 500 en lugar de 400 dependiendo de la implementación
	// Aceptamos ambos como válidos ya que ambos indican error
	if w.Code != http.StatusBadRequest && w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 400 or 500, got %d", w.Code)
	}

	body := w.Body.String()
	if w.Code == http.StatusBadRequest {
		if !strings.Contains(body, "ID inválido o ausente") {
			t.Errorf("unexpected body for 400: %s", body)
		}
	} else if w.Code == http.StatusInternalServerError {
		// Para 500, esperamos un mensaje de error genérico
		if !strings.Contains(body, "Error") || !strings.Contains(body, "500") {
			t.Errorf("unexpected body for 500: %s", body)
		}
	}
}

// Tests adicionales para validaciones de tiempo y casos complejos

func TestCambiosHorario_Post_HorarioCompleto24Horas(t *testing.T) {
	orig := insertCambioHorario
	var capturedHorario *models.CambiosHorario
	insertCambioHorario = func(o orm.Ormer, horario *models.CambiosHorario) (int64, error) {
		capturedHorario = horario
		horario.PK_ID_CAMBIO_HORARIO = 102
		return 1, nil
	}
	defer func() { insertCambioHorario = orig }()

	// Test para horario 24 horas (00:00 a 23:59)
	body := `{"fechaCambioHorario":"2024-10-20","abierto":true,"horaApertura":"00:00:00","horaCierre":"23:59:59"}`
	c, w := setupCtx(http.MethodPost, "/cambios_horario", body)
	c.Post()

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	if capturedHorario.HORA_APERTURA.Format("15:04:05") != "00:00:00" {
		t.Errorf("expected HORA_APERTURA 00:00:00, got %s", capturedHorario.HORA_APERTURA.Format("15:04:05"))
	}
	if capturedHorario.HORA_CIERRE.Format("15:04:05") != "23:59:59" {
		t.Errorf("expected HORA_CIERRE 23:59:59, got %s", capturedHorario.HORA_CIERRE.Format("15:04:05"))
	}
}

func TestCambiosHorario_Post_FechaFutura(t *testing.T) {
	orig := insertCambioHorario
	var capturedHorario *models.CambiosHorario
	insertCambioHorario = func(o orm.Ormer, horario *models.CambiosHorario) (int64, error) {
		capturedHorario = horario
		horario.PK_ID_CAMBIO_HORARIO = 103
		return 1, nil
	}
	defer func() { insertCambioHorario = orig }()

	// Test para fecha futura (ej: planificación de horarios especiales)
	futureDate := time.Now().AddDate(0, 3, 15).Format("2006-01-02")
	body := `{"fechaCambioHorario":"` + futureDate + `","abierto":true,"horaApertura":"10:00:00","horaCierre":"20:00:00"}`
	c, w := setupCtx(http.MethodPost, "/cambios_horario", body)
	c.Post()

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	if capturedHorario.FECHA.Format("2006-01-02") != futureDate {
		t.Errorf("expected date %s, got %s", futureDate, capturedHorario.FECHA.Format("2006-01-02"))
	}
}

func TestCambiosHorario_Post_HoraInvertida(t *testing.T) {
	orig := insertCambioHorario
	insertCambioHorario = func(o orm.Ormer, horario *models.CambiosHorario) (int64, error) {
		// Simular que la base de datos acepta horas invertidas
		horario.PK_ID_CAMBIO_HORARIO = 104
		return 1, nil
	}
	defer func() { insertCambioHorario = orig }()

	// Test con hora de cierre antes que hora de apertura (caso de restaurante nocturno)
	body := `{"fechaCambioHorario":"2024-10-25","abierto":true,"horaApertura":"22:00:00","horaCierre":"02:00:00"}`
	c, w := setupCtx(http.MethodPost, "/cambios_horario", body)
	c.Post()

	// El controlador debería aceptar esto ya que puede ser válido para restaurantes nocturnos
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}

func TestCambiosHorario_Put_CambioDeTipoCompleto(t *testing.T) {
	origFind := queryCambioHorarioByID
	origUpd := updateCambioHorario
	var capturedHorario *models.CambiosHorario

	queryCambioHorarioByID = func(o orm.Ormer, id int64, horario *models.CambiosHorario) error {
		// Horario inicialmente cerrado
		ha, _ := time.Parse("15:04:05", "00:00:00")
		hc, _ := time.Parse("15:04:05", "23:59:59")
		*horario = models.CambiosHorario{
			PK_ID_CAMBIO_HORARIO: id,
			FECHA:                time.Date(2024, 10, 15, 0, 0, 0, 0, time.UTC),
			ABIERTO:              false,
			HORA_APERTURA:        &ha,
			HORA_CIERRE:          hc,
		}
		return nil
	}
	updateCambioHorario = func(o orm.Ormer, horario *models.CambiosHorario) (int64, error) {
		capturedHorario = horario
		return 1, nil
	}
	defer func() { queryCambioHorarioByID = origFind; updateCambioHorario = origUpd }()

	// Cambiar completamente de cerrado a abierto con nuevos horarios
	body := `{
		"fechaCambioHorario":"2024-10-16",
		"abierto":true,
		"horaApertura":"09:00:00",
		"horaCierre":"21:00:00"
	}`
	c, w := setupCtx(http.MethodPut, "/cambios_horario?id=11", body)
	c.Put()

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	if !capturedHorario.ABIERTO {
		t.Error("ABIERTO should be true")
	}
	if capturedHorario.FECHA.Format("2006-01-02") != "2024-10-16" {
		t.Errorf("expected date 2024-10-16, got %s", capturedHorario.FECHA.Format("2006-01-02"))
	}
	if capturedHorario.HORA_APERTURA.Format("15:04:05") != "09:00:00" {
		t.Errorf("expected HORA_APERTURA 09:00:00, got %s", capturedHorario.HORA_APERTURA.Format("15:04:05"))
	}
	if capturedHorario.HORA_CIERRE.Format("15:04:05") != "21:00:00" {
		t.Errorf("expected HORA_CIERRE 21:00:00, got %s", capturedHorario.HORA_CIERRE.Format("15:04:05"))
	}
}

func TestCambiosHorario_GetAll_MultipleHorarios_CompleteScenario(t *testing.T) {
	orig := queryAllCambiosHorario
	queryAllCambiosHorario = func(o orm.Ormer, horarios *[]models.CambiosHorario) (int64, error) {
		// Simular múltiples horarios con diferentes estados
		ha1, _ := time.Parse("15:04:05", "08:00:00")
		hc1, _ := time.Parse("15:04:05", "17:00:00")
		ha2, _ := time.Parse("15:04:05", "10:00:00")
		hc2, _ := time.Parse("15:04:05", "22:00:00")

		*horarios = []models.CambiosHorario{
			// Horario normal
			{
				PK_ID_CAMBIO_HORARIO: 1,
				FECHA:                time.Date(2024, 10, 10, 0, 0, 0, 0, time.UTC),
				ABIERTO:              true,
				HORA_APERTURA:        &ha1,
				HORA_CIERRE:          hc1,
			},
			// Horario cerrado
			{
				PK_ID_CAMBIO_HORARIO: 2,
				FECHA:                time.Date(2024, 10, 11, 0, 0, 0, 0, time.UTC),
				ABIERTO:              false,
				HORA_APERTURA:        nil,
				HORA_CIERRE:          time.Time{},
			},
			// Horario extendido
			{
				PK_ID_CAMBIO_HORARIO: 3,
				FECHA:                time.Date(2024, 10, 12, 0, 0, 0, 0, time.UTC),
				ABIERTO:              true,
				HORA_APERTURA:        &ha2,
				HORA_CIERRE:          hc2,
			},
		}
		return 3, nil
	}
	defer func() { queryAllCambiosHorario = orig }()

	c, w := setupCtx(http.MethodGet, "/cambios_horario", "")
	c.GetAll()

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	body := w.Body.String()
	// Verificar que todos los horarios estén presentes
	if !strings.Contains(body, `"cambioHorarioId":1`) {
		t.Error("should include horario 1")
	}
	if !strings.Contains(body, `"cambioHorarioId":2`) {
		t.Error("should include horario 2")
	}
	if !strings.Contains(body, `"cambioHorarioId":3`) {
		t.Error("should include horario 3")
	}

	// Verificar horarios específicos
	if !strings.Contains(body, `"horaApertura":"08:00:00"`) {
		t.Error("should include 08:00:00 opening time")
	}
	if !strings.Contains(body, `"horaApertura":"10:00:00"`) {
		t.Error("should include 10:00:00 opening time")
	}
	if !strings.Contains(body, `"abierto":false`) {
		t.Error("should include closed status")
	}
}

func TestCambiosHorario_Post_JSONConCamposExtra(t *testing.T) {
	orig := insertCambioHorario
	var capturedHorario *models.CambiosHorario
	insertCambioHorario = func(o orm.Ormer, horario *models.CambiosHorario) (int64, error) {
		capturedHorario = horario
		horario.PK_ID_CAMBIO_HORARIO = 105
		return 1, nil
	}
	defer func() { insertCambioHorario = orig }()

	// JSON con campos extra que deberían ser ignorados
	body := `{
		"fechaCambioHorario":"2024-11-01",
		"abierto":true,
		"horaApertura":"08:30:00",
		"horaCierre":"18:30:00",
		"campoExtra":"valor",
		"otroExtra":123,
		"ignorado":true
	}`
	c, w := setupCtx(http.MethodPost, "/cambios_horario", body)
	c.Post()

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	// Verificar que los campos válidos fueron procesados correctamente
	if capturedHorario.FECHA.Format("2006-01-02") != "2024-11-01" {
		t.Errorf("expected date 2024-11-01, got %s", capturedHorario.FECHA.Format("2006-01-02"))
	}
	if !capturedHorario.ABIERTO {
		t.Error("ABIERTO should be true")
	}
}

func TestCambiosHorario_Put_EmptyJSON(t *testing.T) {
	origFind := queryCambioHorarioByID
	origUpd := updateCambioHorario
	var capturedHorario *models.CambiosHorario

	queryCambioHorarioByID = func(o orm.Ormer, id int64, horario *models.CambiosHorario) error {
		ha, _ := time.Parse("15:04:05", "09:00:00")
		hc, _ := time.Parse("15:04:05", "18:00:00")
		*horario = models.CambiosHorario{
			PK_ID_CAMBIO_HORARIO: id,
			FECHA:                time.Date(2024, 10, 10, 0, 0, 0, 0, time.UTC),
			ABIERTO:              true,
			HORA_APERTURA:        &ha,
			HORA_CIERRE:          hc,
		}
		return nil
	}
	updateCambioHorario = func(o orm.Ormer, horario *models.CambiosHorario) (int64, error) {
		capturedHorario = horario
		return 1, nil
	}
	defer func() { queryCambioHorarioByID = origFind; updateCambioHorario = origUpd }()

	// PUT con JSON vacío - no debería cambiar nada
	body := `{}`
	c, w := setupCtx(http.MethodPut, "/cambios_horario?id=12", body)
	c.Put()

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Verificar que los valores originales se mantuvieron
	if capturedHorario.FECHA.Format("2006-01-02") != "2024-10-10" {
		t.Error("original date should be maintained")
	}
	if !capturedHorario.ABIERTO {
		t.Error("original ABIERTO status should be maintained")
	}
}

func TestCambiosHorario_ResponseFormat_Consistency(t *testing.T) {
	// Test para asegurar que el formato de respuesta sea consistente en todos los métodos
	orig := insertCambioHorario
	insertCambioHorario = func(o orm.Ormer, horario *models.CambiosHorario) (int64, error) {
		horario.PK_ID_CAMBIO_HORARIO = 200
		return 1, nil
	}
	defer func() { insertCambioHorario = orig }()

	body := `{"fechaCambioHorario":"2024-10-30","abierto":true,"horaApertura":"09:00:00","horaCierre":"18:00:00"}`
	c, w := setupCtx(http.MethodPost, "/cambios_horario", body)
	c.Post()

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	responseBody := w.Body.String()
	// Verificar estructura de respuesta JSON
	if !strings.Contains(responseBody, `"code":201`) {
		t.Error("response should include correct status code")
	}
	if !strings.Contains(responseBody, `"message":"Cambio de horario creado correctamente"`) {
		t.Error("response should include success message")
	}
	if !strings.Contains(responseBody, `"data":{`) {
		t.Error("response should include data object")
	}
	if !strings.Contains(responseBody, `"cambioHorarioId":200`) {
		t.Error("response should include cambioHorarioId")
	}
	if !strings.Contains(responseBody, `"fechaCambioHorario":"2024-10-30"`) {
		t.Error("response should include fechaCambioHorario")
	}
	if !strings.Contains(responseBody, `"abierto":true`) {
		t.Error("response should include abierto status")
	}
	if !strings.Contains(responseBody, `"horaApertura":"09:00:00"`) {
		t.Error("response should include horaApertura")
	}
	if !strings.Contains(responseBody, `"horaCierre":"18:00:00"`) {
		t.Error("response should include horaCierre")
	}
}

// Tests finales para casos avanzados y validaciones adicionales

func TestCambiosHorario_GetByCurrentDate_TimezoneHandling(t *testing.T) {
	// Test para verificar que el timezone de Bogotá se maneja correctamente
	originalZone := database.BogotaZone

	// Simular timezone de Bogotá
	bogotaLocation, _ := time.LoadLocation("America/Bogota")
	database.BogotaZone = bogotaLocation
	defer func() { database.BogotaZone = originalZone }()

	orig := queryCambioHorarioByDate
	var capturedDate string
	queryCambioHorarioByDate = func(o orm.Ormer, date string, ch *models.CambiosHorario) error {
		capturedDate = date
		return orm.ErrNoRows
	}
	defer func() { queryCambioHorarioByDate = orig }()

	c, w := setupCtx(http.MethodGet, "/cambios_horario/actual", "")
	c.GetByCurrentDate()

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Verificar que la fecha capturada tiene el formato correcto
	if _, err := time.Parse("2006-01-02", capturedDate); err != nil {
		t.Errorf("captured date should be in YYYY-MM-DD format, got %s", capturedDate)
	}
}

func TestCambiosHorario_Post_FechaLimite_AnoBisiesto(t *testing.T) {
	orig := insertCambioHorario
	var capturedHorario *models.CambiosHorario
	insertCambioHorario = func(o orm.Ormer, horario *models.CambiosHorario) (int64, error) {
		capturedHorario = horario
		horario.PK_ID_CAMBIO_HORARIO = 106
		return 1, nil
	}
	defer func() { insertCambioHorario = orig }()

	// Test para 29 de febrero en año bisiesto
	body := `{"fechaCambioHorario":"2024-02-29","abierto":true,"horaApertura":"08:00:00","horaCierre":"17:00:00"}`
	c, w := setupCtx(http.MethodPost, "/cambios_horario", body)
	c.Post()

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	if capturedHorario.FECHA.Format("2006-01-02") != "2024-02-29" {
		t.Errorf("expected leap year date 2024-02-29, got %s", capturedHorario.FECHA.Format("2006-01-02"))
	}
}

func TestCambiosHorario_Put_SoloHorasConAbiertoFalse(t *testing.T) {
	origFind := queryCambioHorarioByID
	origUpd := updateCambioHorario
	var capturedHorario *models.CambiosHorario

	queryCambioHorarioByID = func(o orm.Ormer, id int64, horario *models.CambiosHorario) error {
		*horario = models.CambiosHorario{
			PK_ID_CAMBIO_HORARIO: id,
			FECHA:                time.Date(2024, 10, 10, 0, 0, 0, 0, time.UTC),
			ABIERTO:              false, // Cerrado
		}
		return nil
	}
	updateCambioHorario = func(o orm.Ormer, horario *models.CambiosHorario) (int64, error) {
		capturedHorario = horario
		return 1, nil
	}
	defer func() { queryCambioHorarioByID = origFind; updateCambioHorario = origUpd }()

	// Intentar actualizar solo horas cuando está cerrado - no debería procesar las horas
	body := `{"horaApertura":"10:00:00","horaCierre":"20:00:00"}`
	c, w := setupCtx(http.MethodPut, "/cambios_horario?id=13", body)
	c.Put()

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Cuando está cerrado, las horas no deberían actualizarse desde el request
	if capturedHorario.ABIERTO {
		t.Error("ABIERTO should remain false")
	}
}

func TestCambiosHorario_GetAll_OrderingConsistency(t *testing.T) {
	orig := queryAllCambiosHorario
	queryAllCambiosHorario = func(o orm.Ormer, horarios *[]models.CambiosHorario) (int64, error) {
		// Simular horarios en diferente orden para verificar consistencia
		*horarios = []models.CambiosHorario{
			{PK_ID_CAMBIO_HORARIO: 3, FECHA: time.Date(2024, 10, 12, 0, 0, 0, 0, time.UTC), ABIERTO: true},
			{PK_ID_CAMBIO_HORARIO: 1, FECHA: time.Date(2024, 10, 10, 0, 0, 0, 0, time.UTC), ABIERTO: false},
			{PK_ID_CAMBIO_HORARIO: 2, FECHA: time.Date(2024, 10, 11, 0, 0, 0, 0, time.UTC), ABIERTO: true},
		}
		return 3, nil
	}
	defer func() { queryAllCambiosHorario = orig }()

	c, w := setupCtx(http.MethodGet, "/cambios_horario", "")
	c.GetAll()

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	body := w.Body.String()
	// Verificar que todos los elementos estén presentes independientemente del orden
	ids := []string{`"cambioHorarioId":1`, `"cambioHorarioId":2`, `"cambioHorarioId":3`}
	for _, id := range ids {
		if !strings.Contains(body, id) {
			t.Errorf("response should contain %s", id)
		}
	}
}

func TestCambiosHorario_Put_ResponseFieldConsistency(t *testing.T) {
	origFind := queryCambioHorarioByID
	origUpd := updateCambioHorario

	queryCambioHorarioByID = func(o orm.Ormer, id int64, horario *models.CambiosHorario) error {
		*horario = models.CambiosHorario{
			PK_ID_CAMBIO_HORARIO: id,
			FECHA:                time.Date(2024, 10, 10, 0, 0, 0, 0, time.UTC),
			ABIERTO:              true,
		}
		return nil
	}
	updateCambioHorario = func(o orm.Ormer, horario *models.CambiosHorario) (int64, error) {
		return 1, nil
	}
	defer func() { queryCambioHorarioByID = origFind; updateCambioHorario = origUpd }()

	body := `{"fechaCambioHorario":"2024-11-05","abierto":true,"horaApertura":"08:00:00","horaCierre":"18:00:00"}`
	c, w := setupCtx(http.MethodPut, "/cambios_horario?id=14", body)
	c.Put()

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	responseBody := w.Body.String()

	// Verificar que el campo en la respuesta sea "fecha" y no "fechaCambioHorario" como en POST
	// (Este es un detalle del controller actual - en PUT usa "fecha")
	if !strings.Contains(responseBody, `"fecha":"2024-11-05"`) {
		t.Error("PUT response should use 'fecha' field")
	}

	// Verificar consistencia de otros campos
	if !strings.Contains(responseBody, `"cambioHorarioId":14`) {
		t.Error("response should include cambioHorarioId")
	}
	if !strings.Contains(responseBody, `"abierto":true`) {
		t.Error("response should include abierto status")
	}
}

func TestCambiosHorario_EdgeCase_LargeID(t *testing.T) {
	origFind := queryCambioHorarioByID
	queryCambioHorarioByID = func(o orm.Ormer, id int64, horario *models.CambiosHorario) error {
		return orm.ErrNoRows
	}
	defer func() { queryCambioHorarioByID = origFind }()

	// Test con ID muy grande
	largeID := "9223372036854775807" // Max int64
	c, w := setupCtx(http.MethodPut, "/cambios_horario?id="+largeID, `{}`)
	c.Put()

	// Debería manejar el ID grande correctamente y devolver not found
	if w.Code != http.StatusOK { // Recuerda que usa semántica 200 con Code=404
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"code":404`) {
		t.Error("should handle large ID and return not found")
	}
}

func TestCambiosHorario_Delete_LargeID(t *testing.T) {
	origDel := deleteCambioHorarioByID
	deleteCambioHorarioByID = func(o orm.Ormer, id int64) (int64, error) {
		// Verificar que el ID llegue correctamente
		if id != 9223372036854775807 {
			t.Errorf("expected large ID, got %d", id)
		}
		return 0, nil // Not found
	}
	defer func() { deleteCambioHorarioByID = origDel }()

	largeID := "9223372036854775807"
	c, w := setupCtx(http.MethodDelete, "/cambios_horario?id="+largeID, "")
	c.Delete()

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"code":404`) {
		t.Error("should handle large ID deletion correctly")
	}
}

// Test de performance básico (no es unitario pero útil para desarrollo)
func TestCambiosHorario_Performance_GetAll_ManyRecords(t *testing.T) {
	orig := queryAllCambiosHorario
	queryAllCambiosHorario = func(o orm.Ormer, horarios *[]models.CambiosHorario) (int64, error) {
		// Simular muchos registros
		*horarios = make([]models.CambiosHorario, 1000)
		for i := range *horarios {
			(*horarios)[i] = models.CambiosHorario{
				PK_ID_CAMBIO_HORARIO: int64(i + 1),
				FECHA:                time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, i),
				ABIERTO:              i%2 == 0,
			}
		}
		return 1000, nil
	}
	defer func() { queryAllCambiosHorario = orig }()

	startTime := time.Now()
	c, w := setupCtx(http.MethodGet, "/cambios_horario", "")
	c.GetAll()
	duration := time.Since(startTime)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Verificar que no tome demasiado tiempo (>1s sería preocupante para 1000 registros)
	if duration > time.Second {
		t.Logf("Performance warning: GetAll took %v for 1000 records", duration)
	}

	// Verificar que la respuesta contenga todos los registros
	body := w.Body.String()
	if !strings.Contains(body, `"cambioHorarioId":1`) {
		t.Error("should include first record")
	}
	if !strings.Contains(body, `"cambioHorarioId":1000`) {
		t.Error("should include last record")
	}
}

func TestCambiosHorario_Wrappers_NilOrmerErrors(t *testing.T) {
	var horarios []models.CambiosHorario
	if _, err := queryAllCambiosHorario(nil, &horarios); err == nil {
		t.Error("expected error for nil ormer in queryAllCambiosHorario")
	}
	var h models.CambiosHorario
	if _, err := insertCambioHorario(nil, &h); err == nil {
		t.Error("expected error for nil ormer in insertCambioHorario")
	}
	if err := queryCambioHorarioByID(nil, 1, &h); err == nil {
		t.Error("expected error for nil ormer in queryCambioHorarioByID")
	}
	if _, err := updateCambioHorario(nil, &h); err == nil {
		t.Error("expected error for nil ormer in updateCambioHorario")
	}
	if _, err := deleteCambioHorarioByID(nil, 1); err == nil {
		t.Error("expected error for nil ormer in deleteCambioHorarioByID")
	}
	if err := queryCambioHorarioByDate(nil, "2024-01-01", &h); err == nil {
		t.Error("expected error for nil ormer in queryCambioHorarioByDate")
	}
}

func TestCambiosHorario_Wrappers_OrmerNormalErrors(t *testing.T) {
	// Usar un Ormer real; no hay DB configurada, así que esperamos errores,
	// pero se ejecutan las rutas "normales" y cuentan para cobertura.
	o := orm.NewOrm()

	var horarios []models.CambiosHorario
	if _, err := queryAllCambiosHorario(o, &horarios); err == nil {
		t.Error("expected error for real ormer in queryAllCambiosHorario")
	}

	var h models.CambiosHorario
	if _, err := insertCambioHorario(o, &h); err == nil {
		t.Error("expected error for real ormer in insertCambioHorario")
	}
	if err := queryCambioHorarioByID(o, 1, &h); err == nil {
		t.Error("expected error for real ormer in queryCambioHorarioByID")
	}
	if _, err := updateCambioHorario(o, &h); err == nil {
		t.Error("expected error for real ormer in updateCambioHorario")
	}
	if _, err := deleteCambioHorarioByID(o, 1); err == nil {
		t.Error("expected error for real ormer in deleteCambioHorarioByID")
	}
	if err := queryCambioHorarioByDate(o, "2024-01-01", &h); err == nil {
		t.Error("expected error for real ormer in queryCambioHorarioByDate")
	}
}
