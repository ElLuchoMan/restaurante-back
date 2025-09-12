package nomina

import (
	"context"
	"database/sql/driver"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	webCtx "github.com/beego/beego/v2/server/web/context"
)

func TestNominaGetAllWithoutDB(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/nominas", nil)
	w := httptest.NewRecorder()
	ctx := webCtx.NewContext()
	ctx.Reset(w, r)
	c := NominaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetAll()

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al obtener nóminas") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestNominaPostInvalidJSON(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/nominas", strings.NewReader("notjson"))
	w := httptest.NewRecorder()
	ctx := webCtx.NewContext()
	ctx.Reset(w, r)
	c := NominaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestNominaPostDBError(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/nominas", nil)
	w := httptest.NewRecorder()
	ctx := webCtx.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte("{\"estadoNomina\":\"OTRO\",\"fechaNomina\":\"2024-01-20\"}")
	c := NominaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	orig := findExistingNominaFn
	findExistingNominaFn = func(o orm.Ormer, fecha time.Time) (*models.Nomina, error) { return nil, orm.ErrNoRows }
	t.Cleanup(func() { findExistingNominaFn = orig })

	c.Post()

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al crear la nómina") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestNominaPostBeforeDay20(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/nominas", nil)
	w := httptest.NewRecorder()
	ctx := webCtx.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte("{\"fechaNomina\":\"2024-01-10\"}")
	c := NominaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	orig := findExistingNominaFn
	findExistingNominaFn = func(o orm.Ormer, fecha time.Time) (*models.Nomina, error) { return nil, orm.ErrNoRows }
	t.Cleanup(func() { findExistingNominaFn = orig })

	c.Post()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "No se puede generar una nómina") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestNominaPostExistingNomina(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/nominas", nil)
	w := httptest.NewRecorder()
	ctx := webCtx.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte("{\"fechaNomina\":\"2024-01-20\"}")
	c := NominaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	origFind := findExistingNominaFn
	origExec := MockExec
	findExistingNominaFn = func(o orm.Ormer, fecha time.Time) (*models.Nomina, error) {
		return &models.Nomina{PK_ID_NOMINA: 1, FECHA: fecha}, nil
	}
	MockExec = func(ctx context.Context, q string, args []driver.NamedValue) (driver.Result, error) {
		return mockResult{}, nil
	}
	t.Cleanup(func() { findExistingNominaFn = origFind; MockExec = origExec })

	c.Post()
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "REGENERADA") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestNominaPostExistingNominaExecError(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/nominas", nil)
	w := httptest.NewRecorder()
	ctx := webCtx.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte("{\"fechaNomina\":\"2024-01-20\"}")
	c := NominaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	origFind := findExistingNominaFn
	origExec := MockExec
	findExistingNominaFn = func(o orm.Ormer, fecha time.Time) (*models.Nomina, error) {
		return &models.Nomina{PK_ID_NOMINA: 1, FECHA: fecha}, nil
	}
	MockExec = func(ctx context.Context, q string, args []driver.NamedValue) (driver.Result, error) {
		return nil, errors.New("exec error")
	}
	t.Cleanup(func() { findExistingNominaFn = origFind; MockExec = origExec })

	c.Post()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al marcar nómina como REGENERADA") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestNominaPostDefaultDate(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/nominas", nil)
	w := httptest.NewRecorder()
	ctx := webCtx.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte("{\"estadoNomina\":\"PAGO\"}")
	c := NominaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	origFind := findExistingNominaFn
	origExec := MockExec
	findExistingNominaFn = func(o orm.Ormer, fecha time.Time) (*models.Nomina, error) {
		return &models.Nomina{PK_ID_NOMINA: 1, FECHA: fecha}, nil
	}
	MockExec = func(ctx context.Context, q string, args []driver.NamedValue) (driver.Result, error) {
		return mockResult{}, nil
	}
	t.Cleanup(func() { findExistingNominaFn = origFind; MockExec = origExec })

	c.Post()
	expected := http.StatusOK
	if time.Now().Day() < 20 {
		expected = http.StatusBadRequest
	}
	if w.Code != expected {
		t.Fatalf("unexpected status %d", w.Code)
	}
}

func TestNominaPostFindExistingError(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/nominas", nil)
	w := httptest.NewRecorder()
	ctx := webCtx.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte("{\"fechaNomina\":\"2024-01-20\"}")
	c := NominaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	orig := findExistingNominaFn
	findExistingNominaFn = func(o orm.Ormer, fecha time.Time) (*models.Nomina, error) {
		return nil, errors.New("boom")
	}
	t.Cleanup(func() { findExistingNominaFn = orig })

	c.Post()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al validar nóminas del mes") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestNominaPostQueryError(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/nominas", nil)
	w := httptest.NewRecorder()
	ctx := webCtx.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte("{\"estadoNomina\":\"PAGO\",\"fechaNomina\":\"2024-01-20\"}")
	c := NominaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	origFind := findExistingNominaFn
	origExec := MockExec
	origQuery := MockQuery
	findExistingNominaFn = func(o orm.Ormer, fecha time.Time) (*models.Nomina, error) { return nil, orm.ErrNoRows }
	MockExec = func(ctx context.Context, q string, args []driver.NamedValue) (driver.Result, error) {
		return mockResult{}, nil
	}
	MockQuery = func(ctx context.Context, q string, args []driver.NamedValue) (driver.Rows, error) {
		if strings.Contains(strings.ToUpper(q), "INSERT") {
			return &mockRows{columns: []string{"pk_id_nomina"}, values: [][]driver.Value{{int64(1)}}}, nil
		}
		return nil, errors.New("mock query error")
	}
	t.Cleanup(func() { findExistingNominaFn = origFind; MockExec = origExec; MockQuery = origQuery })

	c.Post()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al verificar la nómina generada") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestNominaPostSuccess(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/nominas", nil)
	w := httptest.NewRecorder()
	ctx := webCtx.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte("{\"estadoNomina\":\"PAGO\",\"fechaNomina\":\"2024-01-20\"}")
	c := NominaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	origFind := findExistingNominaFn
	origExec := MockExec
	origQuery := MockQuery
	findExistingNominaFn = func(o orm.Ormer, fecha time.Time) (*models.Nomina, error) { return nil, orm.ErrNoRows }
	MockExec = func(ctx context.Context, q string, args []driver.NamedValue) (driver.Result, error) {
		return mockResult{}, nil
	}
	MockQuery = func(ctx context.Context, q string, args []driver.NamedValue) (driver.Rows, error) {
		qUpper := strings.ToUpper(q)
		if strings.Contains(qUpper, "INSERT") {
			return &mockRows{columns: []string{"pk_id_nomina"}, values: [][]driver.Value{{int64(1)}}}, nil
		}
		return &mockRows{
			columns: []string{"pk_id_nomina", "fecha", "monto", "estado_nomina"},
			values:  [][]driver.Value{{int64(1), time.Now(), int64(0), string(models.EstadoNominaPago)}},
		}, nil
	}
	t.Cleanup(func() { findExistingNominaFn = origFind; MockExec = origExec; MockQuery = origQuery })

	c.Post()
	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Nómina creada correctamente") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestNominaPutInvalidID(t *testing.T) {
	r := httptest.NewRequest(http.MethodPut, "/nominas", nil)
	w := httptest.NewRecorder()
	ctx := webCtx.NewContext()
	ctx.Reset(w, r)
	c := NominaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Put()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestNominaPutUpdateError(t *testing.T) {
	r := httptest.NewRequest(http.MethodPut, "/nominas?id=1", nil)
	w := httptest.NewRecorder()
	ctx := webCtx.NewContext()
	ctx.Reset(w, r)
	c := NominaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Put()

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al actualizar el estado de la nómina") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestNominaPutNotFound(t *testing.T) {
	r := httptest.NewRequest(http.MethodPut, "/nominas?id=1", nil)
	w := httptest.NewRecorder()
	ctx := webCtx.NewContext()
	ctx.Reset(w, r)
	c := NominaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	orig := readNominaFn
	readNominaFn = func(o orm.Ormer, n *models.Nomina) error { return orm.ErrNoRows }
	t.Cleanup(func() { readNominaFn = orig })

	c.Put()
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Nómina no encontrada") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestNominaDeleteInvalidID(t *testing.T) {
	r := httptest.NewRequest(http.MethodDelete, "/nominas", nil)
	w := httptest.NewRecorder()
	ctx := webCtx.NewContext()
	ctx.Reset(w, r)
	c := NominaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Delete()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestNominaDeleteNotFound(t *testing.T) {
	r := httptest.NewRequest(http.MethodDelete, "/nominas?id=1", nil)
	w := httptest.NewRecorder()
	ctx := webCtx.NewContext()
	ctx.Reset(w, r)
	c := NominaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Delete()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Nómina no encontrada") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestEstadosNominaPermitidos(t *testing.T) {
	if !estadosNominaPermitidos[models.EstadoNominaPago] || !estadosNominaPermitidos[models.EstadoNominaNoPago] {
		t.Fatalf("expected valid states to be allowed")
	}
	if estadosNominaPermitidos[models.EstadoNomina("otro")] {
		t.Fatalf("unexpected state should not be allowed")
	}
}

func TestNominaGetAllWithFilters(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/nominas?fecha=2024-01-01&mes=1&anio=2024", nil)
	w := httptest.NewRecorder()
	ctx := webCtx.NewContext()
	ctx.Reset(w, r)
	c := NominaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	saved := queryAllNominas
	queryAllNominas = func(o orm.Ormer, out *[]models.Nomina) (int64, error) {
		*out = []models.Nomina{{FECHA: parseDate("2024-01-01").FECHA}}
		return 1, nil
	}
	t.Cleanup(func() { queryAllNominas = saved })

	c.GetAll()
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestNominaGetAllNoResults(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/nominas?fecha=2024-01-02", nil)
	w := httptest.NewRecorder()
	ctx := webCtx.NewContext()
	ctx.Reset(w, r)
	c := NominaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	saved := queryAllNominas
	queryAllNominas = func(o orm.Ormer, out *[]models.Nomina) (int64, error) {
		*out = []models.Nomina{{FECHA: parseDate("2024-01-01").FECHA}}
		return 1, nil
	}
	t.Cleanup(func() { queryAllNominas = saved })

	c.GetAll()
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "No se encontraron") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestNominaGetAllMesMismatch(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/nominas?mes=2", nil)
	w := httptest.NewRecorder()
	ctx := webCtx.NewContext()
	ctx.Reset(w, r)
	c := NominaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	saved := queryAllNominas
	queryAllNominas = func(o orm.Ormer, out *[]models.Nomina) (int64, error) {
		*out = []models.Nomina{{FECHA: parseDate("2024-01-01").FECHA}}
		return 1, nil
	}
	t.Cleanup(func() { queryAllNominas = saved })

	c.GetAll()
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "No se encontraron") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestNominaGetAllAnioMismatch(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/nominas?anio=2023", nil)
	w := httptest.NewRecorder()
	ctx := webCtx.NewContext()
	ctx.Reset(w, r)
	c := NominaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	saved := queryAllNominas
	queryAllNominas = func(o orm.Ormer, out *[]models.Nomina) (int64, error) {
		*out = []models.Nomina{{FECHA: parseDate("2024-01-01").FECHA}}
		return 1, nil
	}
	t.Cleanup(func() { queryAllNominas = saved })

	c.GetAll()
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "No se encontraron") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func parseDate(s string) (n models.Nomina) { n.FECHA, _ = time.Parse("2006-01-02", s); return }

func TestNominaPutAlreadyPaid(t *testing.T) {
	r := httptest.NewRequest(http.MethodPut, "/nominas?id=1", nil)
	w := httptest.NewRecorder()
	ctx := webCtx.NewContext()
	ctx.Reset(w, r)
	c := NominaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	origRead := readNominaFn
	readNominaFn = func(o orm.Ormer, n *models.Nomina) error { n.ESTADO_NOMINA = models.EstadoNominaPago; return nil }
	t.Cleanup(func() { readNominaFn = origRead })

	c.Put()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestNominaPutSuccess(t *testing.T) {
	r := httptest.NewRequest(http.MethodPut, "/nominas?id=1", nil)
	w := httptest.NewRecorder()
	ctx := webCtx.NewContext()
	ctx.Reset(w, r)
	c := NominaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	origRead := readNominaFn
	origUpdate := updateNominaFn
	readNominaFn = func(o orm.Ormer, n *models.Nomina) error { n.ESTADO_NOMINA = models.EstadoNominaNoPago; return nil }
	updateNominaFn = func(o orm.Ormer, n *models.Nomina, cols ...string) (int64, error) { return 1, nil }
	t.Cleanup(func() { readNominaFn = origRead; updateNominaFn = origUpdate })

	c.Put()
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestNominaDeleteSuccess(t *testing.T) {
	r := httptest.NewRequest(http.MethodDelete, "/nominas?id=1", nil)
	w := httptest.NewRecorder()
	ctx := webCtx.NewContext()
	ctx.Reset(w, r)
	c := NominaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	origRead := readNominaFn
	origUpdate := updateNominaFn
	readNominaFn = func(o orm.Ormer, n *models.Nomina) error { n.ESTADO_NOMINA = models.EstadoNominaPago; return nil }
	updateNominaFn = func(o orm.Ormer, n *models.Nomina, cols ...string) (int64, error) { return 1, nil }
	t.Cleanup(func() { readNominaFn = origRead; updateNominaFn = origUpdate })

	c.Delete()
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestNominaDeleteUpdateError(t *testing.T) {
	r := httptest.NewRequest(http.MethodDelete, "/nominas?id=1", nil)
	w := httptest.NewRecorder()
	ctx := webCtx.NewContext()
	ctx.Reset(w, r)
	c := NominaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	origRead := readNominaFn
	origUpdate := updateNominaFn
	readNominaFn = func(o orm.Ormer, n *models.Nomina) error { return nil }
	updateNominaFn = func(o orm.Ormer, n *models.Nomina, cols ...string) (int64, error) { return 0, errors.New("boom") }
	t.Cleanup(func() { readNominaFn = origRead; updateNominaFn = origUpdate })

	c.Delete()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al eliminar lógicamente la nómina") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}
func TestFindExistingNominaFnSuccess(t *testing.T) {
	o := ormNewNomina()
	fecha := time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC)

	origQuery := MockQuery
	MockQuery = func(ctx context.Context, q string, args []driver.NamedValue) (driver.Rows, error) {
		return &mockRows{
			columns: []string{"pk_id_nomina", "fecha", "monto", "estado_nomina"},
			values:  [][]driver.Value{{int64(1), fecha, int64(100), string(models.EstadoNominaPago)}},
		}, nil
	}
	t.Cleanup(func() { MockQuery = origQuery })

	existing, err := findExistingNominaFn(o, fecha)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if existing == nil || existing.PK_ID_NOMINA != 1 {
		t.Fatalf("unexpected result: %#v", existing)
	}
}

func TestFindExistingNominaFnError(t *testing.T) {
	o := ormNewNomina()
	fecha := time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC)

	origQuery := MockQuery
	MockQuery = func(ctx context.Context, q string, args []driver.NamedValue) (driver.Rows, error) {
		return nil, errors.New("query error")
	}
	t.Cleanup(func() { MockQuery = origQuery })

	existing, err := findExistingNominaFn(o, fecha)
	if err == nil || existing != nil {
		t.Fatalf("expected error finding nomina")
	}
}
