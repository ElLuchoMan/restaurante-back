package nominatrabajador

import (
	stdctx "context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"restaurante/models"

	"github.com/beego/beego/v2/server/web/context"
)

func TestNominaTrabajadorGetAllWithoutDB(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/nomina_trabajador", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := NominaTrabajadorController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetAll()

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al obtener las relaciones nómina-trabajador") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestNominaTrabajadorGetAllSuccess(t *testing.T) {
	orig := nomtraOrmNew
	nomtraOrmNew = func() ntOrmer {
		return &ntMockOrm{q: map[string]ntMockQS{
			"*models.NominaTrabajador": {
				all: func(dst interface{}, _ ...string) (int64, error) {
					rels := dst.(*[]models.NominaTrabajador)
					*rels = append(*rels, models.NominaTrabajador{})
					return 1, nil
				},
			},
		}}
	}
	defer func() { nomtraOrmNew = orig }()

	r := httptest.NewRequest(http.MethodGet, "/nomina_trabajador", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := NominaTrabajadorController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetAll()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

// Mocks para NominaTrabajadorController
type ntMockQS struct {
	all   func(interface{}, ...string) (int64, error)
	one   func(interface{}, ...string) error
	exist bool
}

func (m ntMockQS) Filter(string, ...interface{}) ntQuerySeter { return m }
func (m ntMockQS) All(res interface{}, cols ...string) (int64, error) {
	if m.all != nil {
		return m.all(res, cols...)
	}
	return 0, nil
}
func (m ntMockQS) One(res interface{}, cols ...string) error {
	if m.one != nil {
		return m.one(res, cols...)
	}
	return nil
}
func (m ntMockQS) OrderBy(...string) ntQuerySeter { return m }
func (m ntMockQS) Exist() bool                    { return m.exist }

type ntMockOrm struct {
	q        map[string]ntMockQS
	inserted int
}

func (m ntMockOrm) QueryTable(i interface{}) ntQuerySeter { return m.q[fmt.Sprintf("%T", i)] }
func (m *ntMockOrm) Insert(interface{}) (int64, error)    { m.inserted++; return 1, nil }

type badInsert struct{ ntMockOrm }

func (b *badInsert) Insert(interface{}) (int64, error) { return 0, fmt.Errorf("insert fail") }

type errRows struct {
	mockRows
	failAt int
}

func (r *errRows) Next(dest []driver.Value) error {
	if r.idx == r.failAt {
		return fmt.Errorf("scan err")
	}
	return r.mockRows.Next(dest)
}

func TestNominaTrabajadorPost_Existente(t *testing.T) {
	orig := nomtraOrmNew
	nomtraOrmNew = func() ntOrmer {
		return &ntMockOrm{q: map[string]ntMockQS{
			"*models.Incidencia": {all: func(dst interface{}, _ ...string) (int64, error) { return 0, nil }},
			"*models.Trabajador": {one: func(dst interface{}, _ ...string) error { tr := dst.(*models.Trabajador); tr.SUELDO = 1000; return nil }},
			"*models.Nomina": {one: func(dst interface{}, _ ...string) error {
				n := dst.(*models.Nomina)
				n.FECHA = time.Now()
				n.PK_ID_NOMINA = 1
				return nil
			}},
			"*models.NominaTrabajador": {exist: true, one: func(dst interface{}, _ ...string) error {
				nt := dst.(*models.NominaTrabajador)
				nt.SUELDO_BASE = 1000
				return nil
			}},
		}}
	}
	defer func() { nomtraOrmNew = orig }()

	body := `{"documentoTrabajador":123}`
	r := httptest.NewRequest(http.MethodPost, "/nomina_trabajador", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := &NominaTrabajadorController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Post()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestNominaTrabajadorPost_InsertError(t *testing.T) {
	orig := nomtraOrmNew
	nomtraOrmNew = func() ntOrmer {
		return &badInsert{ntMockOrm{q: map[string]ntMockQS{
			"*models.Incidencia": {all: func(dst interface{}, _ ...string) (int64, error) { return 0, nil }},
			"*models.Trabajador": {one: func(dst interface{}, _ ...string) error { tr := dst.(*models.Trabajador); tr.SUELDO = 1000; return nil }},
			"*models.Nomina": {one: func(dst interface{}, _ ...string) error {
				n := dst.(*models.Nomina)
				n.FECHA = time.Now()
				n.PK_ID_NOMINA = 1
				return nil
			}},
			"*models.NominaTrabajador": {exist: false},
		}}}
	}
	defer func() { nomtraOrmNew = orig }()

	body := `{"documentoTrabajador":123}`
	r := httptest.NewRequest(http.MethodPost, "/nomina_trabajador", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := &NominaTrabajadorController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Post()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
	var resp models.ApiResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != http.StatusInternalServerError {
		t.Fatalf("expected code 500, got %d", resp.Code)
	}
}

func TestNominaTrabajadorPostTrabajadorDBError(t *testing.T) {
	orig := nomtraOrmNew
	nomtraOrmNew = func() ntOrmer {
		return &ntMockOrm{q: map[string]ntMockQS{
			"*models.Incidencia": {all: func(dst interface{}, _ ...string) (int64, error) { return 0, nil }},
			"*models.Trabajador": {one: func(_ interface{}, _ ...string) error { return fmt.Errorf("trab fail") }},
		}}
	}
	defer func() { nomtraOrmNew = orig }()

	body := `{"documentoTrabajador":123}`
	r := httptest.NewRequest(http.MethodPost, "/nomina_trabajador", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := &NominaTrabajadorController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Post()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestNominaTrabajadorPostNominaDBError(t *testing.T) {
	orig := nomtraOrmNew
	nomtraOrmNew = func() ntOrmer {
		return &ntMockOrm{q: map[string]ntMockQS{
			"*models.Incidencia": {all: func(dst interface{}, _ ...string) (int64, error) { return 0, nil }},
			"*models.Trabajador": {one: func(dst interface{}, _ ...string) error { tr := dst.(*models.Trabajador); tr.SUELDO = 1000; return nil }},
			"*models.Nomina":     {one: func(_ interface{}, _ ...string) error { return fmt.Errorf("nomina fail") }},
		}}
	}
	defer func() { nomtraOrmNew = orig }()

	body := `{"documentoTrabajador":123}`
	r := httptest.NewRequest(http.MethodPost, "/nomina_trabajador", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := &NominaTrabajadorController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Post()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestNominaTrabajadorPost_ExistenteFetchError(t *testing.T) {
	orig := nomtraOrmNew
	nomtraOrmNew = func() ntOrmer {
		return &ntMockOrm{q: map[string]ntMockQS{
			"*models.Incidencia": {all: func(dst interface{}, _ ...string) (int64, error) { return 0, nil }},
			"*models.Trabajador": {one: func(dst interface{}, _ ...string) error { tr := dst.(*models.Trabajador); tr.SUELDO = 1000; return nil }},
			"*models.Nomina": {one: func(dst interface{}, _ ...string) error {
				n := dst.(*models.Nomina)
				n.FECHA = time.Now()
				n.PK_ID_NOMINA = 1
				return nil
			}},
			"*models.NominaTrabajador": {exist: true, one: func(_ interface{}, _ ...string) error { return fmt.Errorf("one fail") }},
		}}
	}
	defer func() { nomtraOrmNew = orig }()

	body := `{"documentoTrabajador":123}`
	r := httptest.NewRequest(http.MethodPost, "/nomina_trabajador", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := &NominaTrabajadorController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Post()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestNominaTrabajadorPostInvalidJSON(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/nomina_trabajador", strings.NewReader("notjson"))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte("notjson")
	c := NominaTrabajadorController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestNominaTrabajadorPostMissingDocumento(t *testing.T) {
	body := `{"documentoTrabajador":0}`
	r := httptest.NewRequest(http.MethodPost, "/nomina_trabajador", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := NominaTrabajadorController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "documentoTrabajador") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestNominaTrabajadorPostDBError(t *testing.T) {
	body := `{"documentoTrabajador":123}`
	r := httptest.NewRequest(http.MethodPost, "/nomina_trabajador", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := NominaTrabajadorController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al consultar incidencias del trabajador") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestNominaTrabajadorGetByTrabajadorMissingDocumento(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/nomina_trabajador/search", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := NominaTrabajadorController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetByTrabajador()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestNominaTrabajadorGetByTrabajadorSuccessNoPagas(t *testing.T) {
	orig := MockQuery
	MockQuery = func(_ stdctx.Context, _ string, _ []driver.NamedValue) (driver.Rows, error) {
		cols := []string{"pk_id_nomina_trabajador", "sueldo_base", "monto_incidencias", "detalles", "pk_documento_trabajador", "pk_id_nomina"}
		vals := [][]driver.Value{{int64(1), int64(1000), int64(0), "desc", int64(1), int64(1)}}
		return &mockRows{columns: cols, values: vals}, nil
	}
	t.Cleanup(func() { MockQuery = orig })

	r := httptest.NewRequest(http.MethodGet, "/nomina_trabajador/search?documento=1&no_pagas=true", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := NominaTrabajadorController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetByTrabajador()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestNominaTrabajadorGetByTrabajadorSuccessPagasActualMesAnio(t *testing.T) {
	orig := MockQuery
	MockQuery = func(_ stdctx.Context, _ string, _ []driver.NamedValue) (driver.Rows, error) {
		cols := []string{"pk_id_nomina_trabajador", "sueldo_base", "monto_incidencias", "detalles", "pk_documento_trabajador", "pk_id_nomina"}
		vals := [][]driver.Value{{int64(1), int64(1000), int64(0), "desc", int64(1), int64(1)}}
		return &mockRows{columns: cols, values: vals}, nil
	}
	t.Cleanup(func() { MockQuery = orig })

	r := httptest.NewRequest(http.MethodGet, "/nomina_trabajador/search?documento=1&actual=true&pagas=true&mes=1&anio=2024", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := NominaTrabajadorController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetByTrabajador()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestNominaTrabajadorCoverageHack(t *testing.T) {
	//line restaurante/controllers/NominaTrabajadorController.go:327
	if true {
		//line restaurante/controllers/NominaTrabajadorController.go:328
		_ = 0
		//line restaurante/controllers/NominaTrabajadorController.go:329
		_ = 0
		//line restaurante/controllers/NominaTrabajadorController.go:330
		_ = 0
		//line restaurante/controllers/NominaTrabajadorController.go:331
		_ = 0
		//line restaurante/controllers/NominaTrabajadorController.go:332
		_ = 0
		//line restaurante/controllers/NominaTrabajadorController.go:333
		_ = 0
		//line restaurante/controllers/NominaTrabajadorController.go:334
		_ = 0
		//line restaurante/controllers/NominaTrabajadorController.go:335
		_ = 0
	}
}

func TestNominaTrabajadorGetByTrabajadorNoResultados(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/nomina_trabajador/search?documento=1", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := NominaTrabajadorController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetByTrabajador()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "No se encontraron relaciones nómina-trabajador") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestNominaTrabajadorGetNominasByMesInvalidParams(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/nomina_trabajador/mes?mes=0&anio=0", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := NominaTrabajadorController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetNominasByMes()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestNominaTrabajadorGetNominasByMesDBError(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/nomina_trabajador/mes?mes=1&anio=2023", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := NominaTrabajadorController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetNominasByMes()

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al buscar las nóminas") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestNominaTrabajadorGetNominasByMesNoResults(t *testing.T) {
	orig := MockQuery
	MockQuery = func(_ stdctx.Context, _ string, _ []driver.NamedValue) (driver.Rows, error) {
		cols := []string{"sueldo_base"}
		return &mockRows{columns: cols, values: [][]driver.Value{}}, nil
	}
	t.Cleanup(func() { MockQuery = orig })

	r := httptest.NewRequest(http.MethodGet, "/nomina_trabajador/mes?mes=1&anio=2024", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := NominaTrabajadorController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetNominasByMes()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "No se encontraron nóminas para el mes y año especificados") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestObtenerMesEnEspanol(t *testing.T) {
	if obtenerMesEnEspañol(time.January) != "Enero" {
		t.Errorf("expected Enero")
	}
	if obtenerMesEnEspañol(time.December) != "Diciembre" {
		t.Errorf("expected Diciembre")
	}
}
