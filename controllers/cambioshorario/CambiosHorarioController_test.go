package cambioshorario

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"restaurante/database"
	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
)

// Driver/mock mínimo para inicializar orm.NewOrm() sin DB real
type mockDriver struct{}

type mockConn struct{}

type mockStmt struct{ query string }

type mockTx struct{}

func (d mockDriver) Open(name string) (driver.Conn, error) { return &mockConn{}, nil }

func (c *mockConn) Prepare(query string) (driver.Stmt, error) { return &mockStmt{query: query}, nil }
func (c *mockConn) Close() error                              { return nil }
func (c *mockConn) Begin() (driver.Tx, error)                 { return &mockTx{}, nil }

func (s *mockStmt) Close() error                                    { return nil }
func (s *mockStmt) NumInput() int                                   { return -1 }
func (s *mockStmt) Exec(args []driver.Value) (driver.Result, error) { return mockResult{}, nil }
func (s *mockStmt) Query(args []driver.Value) (driver.Rows, error)  { return &mockRows{}, nil }

func (mockTx) Commit() error   { return nil }
func (mockTx) Rollback() error { return nil }

type mockResult struct{}

func (mockResult) LastInsertId() (int64, error) { return 1, nil }
func (mockResult) RowsAffected() (int64, error) { return 1, nil }

type mockRows struct{}

func (r *mockRows) Columns() []string              { return []string{} }
func (r *mockRows) Close() error                   { return nil }
func (r *mockRows) Next(dest []driver.Value) error { return io.EOF }

func TestMain(m *testing.M) {
	// Evitar pánico por JWT en inits de otros paquetes
	_ = os.Setenv("JWT_SECRET", "testsecret")
	// Inicializar timezone para pruebas
	database.InitTimezone()
	// Registrar driver y base por defecto para que orm.NewOrm() no falle
	sql.Register("mock", mockDriver{})
	orm.RegisterDriver("mock", orm.DRPostgres)
	_ = orm.RegisterDataBase("default", "mock", "")
	os.Exit(m.Run())
}

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

// Copia íntegra de los tests existentes adaptados al nuevo paquete
// (contenido idéntico al archivo original en controllers/ con package actualizado)

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

func TestCambiosHorario_GetAll_OK(t *testing.T) {
    orig := queryAllCambiosHorario
    queryAllCambiosHorario = func(o orm.Ormer, horarios *[]models.CambiosHorario) (int64, error) {
        ha, _ := time.Parse("15:04:05", "08:00:00")
        hc, _ := time.Parse("15:04:05", "17:00:00")
        *horarios = []models.CambiosHorario{
            {PK_ID_CAMBIO_HORARIO: 1, FECHA: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), ABIERTO: true, HORA_APERTURA: &ha, HORA_CIERRE: hc},
            {PK_ID_CAMBIO_HORARIO: 2, FECHA: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), ABIERTO: false},
        }
        return int64(len(*horarios)), nil
    }
    defer func() { queryAllCambiosHorario = orig }()

    c, w := setupCtx(http.MethodGet, "/cambios_horario", "")
    c.GetAll()
    if w.Code != http.StatusOK {
        t.Fatalf("expected 200, got %d", w.Code)
    }
}

func TestCambiosHorario_GetByCurrentDate_NoRows(t *testing.T) {
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

func TestCambiosHorario_GetByCurrentDate_DBError(t *testing.T) {
    orig := queryCambioHorarioByDate
    queryCambioHorarioByDate = func(o orm.Ormer, date string, ch *models.CambiosHorario) error {
        return errors.New("db fail")
    }
    defer func() { queryCambioHorarioByDate = orig }()

    c, w := setupCtx(http.MethodGet, "/cambios_horario/actual", "")
    c.GetByCurrentDate()

    if w.Code != http.StatusInternalServerError {
        t.Fatalf("expected 500, got %d", w.Code)
    }
}

func TestCambiosHorario_GetByCurrentDate_OK(t *testing.T) {
    orig := queryCambioHorarioByDate
    queryCambioHorarioByDate = func(o orm.Ormer, date string, ch *models.CambiosHorario) error {
        ch.PK_ID_CAMBIO_HORARIO = 10
        ch.FECHA = time.Now().In(database.BogotaZone)
        ch.ABIERTO = true
        ha, _ := time.Parse("15:04:05", "08:30:00")
        hc, _ := time.Parse("15:04:05", "18:00:00")
        ch.HORA_APERTURA = &ha
        ch.HORA_CIERRE = hc
        return nil
    }
    defer func() { queryCambioHorarioByDate = orig }()

    c, w := setupCtx(http.MethodGet, "/cambios_horario/actual", "")
    c.GetByCurrentDate()

    if w.Code != http.StatusOK {
        t.Fatalf("expected 200, got %d", w.Code)
    }
}

func TestCambiosHorario_Post_BadJSON(t *testing.T) {
    c, w := setupCtx(http.MethodPost, "/cambios_horario", "{bad")
    c.Post()
    if w.Code != http.StatusBadRequest {
        t.Fatalf("expected 400, got %d", w.Code)
    }
}

func TestCambiosHorario_Post_MissingFields(t *testing.T) {
    // Falta fechaCambioHorario
    body := `{"abierto": true, "horaApertura":"08:00:00", "horaCierre":"17:00:00"}`
    c, w := setupCtx(http.MethodPost, "/cambios_horario", body)
    c.Post()
    if w.Code != http.StatusBadRequest {
        t.Fatalf("expected 400, got %d", w.Code)
    }

    // Falta horaApertura cuando abierto = true
    body = `{"fechaCambioHorario":"2024-01-01", "abierto": true, "horaCierre":"17:00:00"}`
    c, w = setupCtx(http.MethodPost, "/cambios_horario", body)
    c.Post()
    if w.Code != http.StatusBadRequest {
        t.Fatalf("expected 400, got %d", w.Code)
    }
}

func TestCambiosHorario_Post_InvalidFecha(t *testing.T) {
    c, w := setupCtx(http.MethodPost, "/cambios_horario", `{"fechaCambioHorario":"2024-13-40", "abierto": false}`)
    c.Post()
    if w.Code != http.StatusBadRequest {
        t.Fatalf("expected 400, got %d", w.Code)
    }
}

func TestCambiosHorario_Post_MissingAbierto(t *testing.T) {
    c, w := setupCtx(http.MethodPost, "/cambios_horario", `{"fechaCambioHorario":"2024-01-01"}`)
    c.Post()
    if w.Code != http.StatusBadRequest {
        t.Fatalf("expected 400, got %d", w.Code)
    }
}

func TestCambiosHorario_Post_DBError(t *testing.T) {
    orig := insertCambioHorario
    insertCambioHorario = func(o orm.Ormer, horario *models.CambiosHorario) (int64, error) {
        return 0, errors.New("insert fail")
    }
    defer func() { insertCambioHorario = orig }()

    body := `{"fechaCambioHorario":"2024-01-01", "abierto": false}`
    c, w := setupCtx(http.MethodPost, "/cambios_horario", body)
    c.Post()
    if w.Code != http.StatusInternalServerError {
        t.Fatalf("expected 500, got %d", w.Code)
    }
}

func TestCambiosHorario_Post_OK_Cerrado(t *testing.T) {
    orig := insertCambioHorario
    insertCalled := false
    insertCambioHorario = func(o orm.Ormer, horario *models.CambiosHorario) (int64, error) {
        insertCalled = true
        return 1, nil
    }
    defer func() { insertCambioHorario = orig }()

    body := `{"fechaCambioHorario":"2024-01-02", "abierto": false}`
    c, w := setupCtx(http.MethodPost, "/cambios_horario", body)
    c.Post()
    if w.Code != http.StatusCreated {
        t.Fatalf("expected 201, got %d", w.Code)
    }
    if !insertCalled {
        t.Fatalf("expected insert to be called")
    }
}

func TestCambiosHorario_Post_OK_Abierto(t *testing.T) {
    orig := insertCambioHorario
    insertCambioHorario = func(o orm.Ormer, horario *models.CambiosHorario) (int64, error) { return 1, nil }
    defer func() { insertCambioHorario = orig }()

    body := `{"fechaCambioHorario":"2024-01-03", "abierto": true, "horaApertura":"08:00:00", "horaCierre":"17:00:00"}`
    c, w := setupCtx(http.MethodPost, "/cambios_horario", body)
    c.Post()
    if w.Code != http.StatusCreated {
        t.Fatalf("expected 201, got %d", w.Code)
    }
}

func TestCambiosHorario_Post_InvalidHoraCierre(t *testing.T) {
    c, w := setupCtx(http.MethodPost, "/cambios_horario", `{"fechaCambioHorario":"2024-01-03", "abierto": true, "horaApertura":"08:00:00", "horaCierre":"xx"}`)
    c.Post()
    if w.Code != http.StatusBadRequest {
        t.Fatalf("expected 400, got %d", w.Code)
    }
}

func TestCambiosHorario_Put_BadID(t *testing.T) {
    c, w := setupCtx(http.MethodPut, "/cambios_horario", `{"x":1}`)
    c.Put()
    if w.Code != http.StatusBadRequest {
        t.Fatalf("expected 400, got %d", w.Code)
    }
}

func TestCambiosHorario_Put_BadJSON(t *testing.T) {
    c, w := setupCtx(http.MethodPut, "/cambios_horario?id=1", "{bad")
    c.Put()
    if w.Code != http.StatusBadRequest {
        t.Fatalf("expected 400, got %d", w.Code)
    }
}

func TestCambiosHorario_Put_NotFound(t *testing.T) {
    orig := queryCambioHorarioByID
    queryCambioHorarioByID = func(o orm.Ormer, id int64, horario *models.CambiosHorario) error { return orm.ErrNoRows }
    defer func() { queryCambioHorarioByID = orig }()

    c, w := setupCtx(http.MethodPut, "/cambios_horario?id=1", `{"abierto": false}`)
    c.Put()
    if w.Code != http.StatusOK {
        t.Fatalf("expected 200, got %d", w.Code)
    }
    if !strings.Contains(w.Body.String(), "Cambio de horario no encontrado") {
        t.Errorf("unexpected body: %s", w.Body.String())
    }
}

func TestCambiosHorario_Put_DBErrorOnQuery(t *testing.T) {
    orig := queryCambioHorarioByID
    queryCambioHorarioByID = func(o orm.Ormer, id int64, horario *models.CambiosHorario) error { return errors.New("query fail") }
    defer func() { queryCambioHorarioByID = orig }()

    c, w := setupCtx(http.MethodPut, "/cambios_horario?id=2", `{"abierto": true, "horaApertura":"08:00:00", "horaCierre":"17:00:00"}`)
    c.Put()
    if w.Code != http.StatusInternalServerError {
        t.Fatalf("expected 500, got %d", w.Code)
    }
}

func TestCambiosHorario_Put_UpdateError(t *testing.T) {
    origQ := queryCambioHorarioByID
    origU := updateCambioHorario
    queryCambioHorarioByID = func(o orm.Ormer, id int64, horario *models.CambiosHorario) error { return nil }
    updateCambioHorario = func(o orm.Ormer, horario *models.CambiosHorario) (int64, error) { return 0, errors.New("update fail") }
    defer func() { queryCambioHorarioByID = origQ; updateCambioHorario = origU }()

    c, w := setupCtx(http.MethodPut, "/cambios_horario?id=3", `{"abierto": false}`)
    c.Put()
    if w.Code != http.StatusInternalServerError {
        t.Fatalf("expected 500, got %d", w.Code)
    }
}

func TestCambiosHorario_Put_OK(t *testing.T) {
    origQ := queryCambioHorarioByID
    origU := updateCambioHorario
    queryCambioHorarioByID = func(o orm.Ormer, id int64, horario *models.CambiosHorario) error { return nil }
    updateCambioHorario = func(o orm.Ormer, horario *models.CambiosHorario) (int64, error) { return 1, nil }
    defer func() { queryCambioHorarioByID = origQ; updateCambioHorario = origU }()

    c, w := setupCtx(http.MethodPut, "/cambios_horario?id=4", `{"abierto": true, "horaApertura":"08:00:00", "horaCierre":"17:00:00"}`)
    c.Put()
    if w.Code != http.StatusOK {
        t.Fatalf("expected 200, got %d", w.Code)
    }
}

func TestCambiosHorario_Put_InvalidHoraApertura(t *testing.T) {
    origQ := queryCambioHorarioByID
    queryCambioHorarioByID = func(o orm.Ormer, id int64, horario *models.CambiosHorario) error { return nil }
    defer func() { queryCambioHorarioByID = origQ }()

    c, w := setupCtx(http.MethodPut, "/cambios_horario?id=5", `{"abierto": true, "horaApertura":"xx", "horaCierre":"17:00:00"}`)
    c.Put()
    if w.Code != http.StatusBadRequest {
        t.Fatalf("expected 400, got %d", w.Code)
    }
}

func TestCambiosHorario_Put_InvalidHoraCierre(t *testing.T) {
    origQ := queryCambioHorarioByID
    queryCambioHorarioByID = func(o orm.Ormer, id int64, horario *models.CambiosHorario) error { return nil }
    defer func() { queryCambioHorarioByID = origQ }()

    c, w := setupCtx(http.MethodPut, "/cambios_horario?id=6", `{"abierto": true, "horaApertura":"08:00:00", "horaCierre":"xx"}`)
    c.Put()
    if w.Code != http.StatusBadRequest {
        t.Fatalf("expected 400, got %d", w.Code)
    }
}

func TestCambiosHorario_Delete_BadID(t *testing.T) {
    c, w := setupCtx(http.MethodDelete, "/cambios_horario", "")
    c.Delete()
    if w.Code != http.StatusBadRequest {
        t.Fatalf("expected 400, got %d", w.Code)
    }
}

func TestCambiosHorario_Delete_DBError(t *testing.T) {
    orig := deleteCambioHorarioByID
    deleteCambioHorarioByID = func(o orm.Ormer, id int64) (int64, error) { return 0, errors.New("del fail") }
    defer func() { deleteCambioHorarioByID = orig }()

    c, w := setupCtx(http.MethodDelete, "/cambios_horario?id=1", "")
    c.Delete()
    if w.Code != http.StatusInternalServerError {
        t.Fatalf("expected 500, got %d", w.Code)
    }
}

func TestCambiosHorario_Delete_NotFound(t *testing.T) {
    orig := deleteCambioHorarioByID
    deleteCambioHorarioByID = func(o orm.Ormer, id int64) (int64, error) { return 0, nil }
    defer func() { deleteCambioHorarioByID = orig }()

    c, w := setupCtx(http.MethodDelete, "/cambios_horario?id=2", "")
    c.Delete()
    if w.Code != http.StatusOK {
        t.Fatalf("expected 200, got %d", w.Code)
    }
    if !strings.Contains(w.Body.String(), "no encontrado") {
        t.Errorf("unexpected body: %s", w.Body.String())
    }
}

func TestCambiosHorario_Delete_OK(t *testing.T) {
    orig := deleteCambioHorarioByID
    deleteCambioHorarioByID = func(o orm.Ormer, id int64) (int64, error) { return 1, nil }
    defer func() { deleteCambioHorarioByID = orig }()

    c, w := setupCtx(http.MethodDelete, "/cambios_horario?id=3", "")
    c.Delete()
    if w.Code != http.StatusOK {
        t.Fatalf("expected 200, got %d", w.Code)
    }
}

