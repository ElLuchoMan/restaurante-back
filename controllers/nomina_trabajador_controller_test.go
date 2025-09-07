package controllers

import (
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

// Mocks para NominaTrabajadorController
type ntMockQS struct{
    all func(interface{}, ...string) (int64, error)
    one func(interface{}, ...string) error
    exist bool
}
func (m ntMockQS) Filter(string, ...interface{}) ntQuerySeter { return m }
func (m ntMockQS) All(res interface{}, cols ...string) (int64, error) { if m.all!=nil { return m.all(res, cols...) }; return 0, nil }
func (m ntMockQS) One(res interface{}, cols ...string) error { if m.one!=nil { return m.one(res, cols...) }; return nil }
func (m ntMockQS) OrderBy(...string) ntQuerySeter { return m }
func (m ntMockQS) Exist() bool { return m.exist }

type ntMockOrm struct{ q map[string]ntMockQS; inserted int }
func (m ntMockOrm) QueryTable(i interface{}) ntQuerySeter { return m.q[fmt.Sprintf("%T", i)] }
func (m *ntMockOrm) Insert(interface{}) (int64, error) { m.inserted++; return 1, nil }

type badInsert struct{ ntMockOrm }
func (b *badInsert) Insert(interface{}) (int64, error) { return 0, fmt.Errorf("insert fail") }

func TestNominaTrabajadorPost_Existente(t *testing.T) {
    orig := nomtraOrmNew
    nomtraOrmNew = func() ntOrmer {
        return &ntMockOrm{ q: map[string]ntMockQS{
            "*models.Incidencia": { all: func(dst interface{}, _ ...string) (int64, error) { return 0, nil } },
            "*models.Trabajador": { one: func(dst interface{}, _ ...string) error { tr := dst.(*models.Trabajador); tr.SUELDO = 1000; return nil } },
            "*models.Nomina": { one: func(dst interface{}, _ ...string) error { n := dst.(*models.Nomina); n.FECHA = time.Now(); n.PK_ID_NOMINA = 1; return nil } },
            "*models.NominaTrabajador": { exist: true, one: func(dst interface{}, _ ...string) error { nt := dst.(*models.NominaTrabajador); nt.SUELDO_BASE = 1000; return nil } },
        } }
    }
    defer func(){ nomtraOrmNew = orig }()

    body := `{"documentoTrabajador":123}`
    r := httptest.NewRequest(http.MethodPost, "/nomina_trabajador", strings.NewReader(body))
    w := httptest.NewRecorder(); ctx := context.NewContext(); ctx.Reset(w, r); ctx.Input.RequestBody = []byte(body)
    c := &NominaTrabajadorController{}; c.Ctx = ctx; c.Data = make(map[interface{}]interface{})
    c.Post()
    if w.Code != http.StatusOK { t.Fatalf("expected 200, got %d", w.Code) }
}

func TestNominaTrabajadorPost_InsertError(t *testing.T) {
    orig := nomtraOrmNew
    nomtraOrmNew = func() ntOrmer {
        return &badInsert{ ntMockOrm{ q: map[string]ntMockQS{
            "*models.Incidencia": { all: func(dst interface{}, _ ...string) (int64, error) { return 0, nil } },
            "*models.Trabajador": { one: func(dst interface{}, _ ...string) error { tr := dst.(*models.Trabajador); tr.SUELDO = 1000; return nil } },
            "*models.Nomina": { one: func(dst interface{}, _ ...string) error { n := dst.(*models.Nomina); n.FECHA = time.Now(); n.PK_ID_NOMINA = 1; return nil } },
            "*models.NominaTrabajador": { exist: false },
        } } }
    }
    defer func(){ nomtraOrmNew = orig }()

    body := `{"documentoTrabajador":123}`
    r := httptest.NewRequest(http.MethodPost, "/nomina_trabajador", strings.NewReader(body))
    w := httptest.NewRecorder(); ctx := context.NewContext(); ctx.Reset(w, r); ctx.Input.RequestBody = []byte(body)
    c := &NominaTrabajadorController{}; c.Ctx = ctx; c.Data = make(map[interface{}]interface{})
    c.Post()
    if w.Code != http.StatusInternalServerError { t.Fatalf("expected 500, got %d", w.Code) }
    var resp models.ApiResponse; _ = json.Unmarshal(w.Body.Bytes(), &resp)
    if resp.Code != http.StatusInternalServerError { t.Fatalf("expected code 500, got %d", resp.Code) }
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
		t.Fatalf("expected status 500, got %d", w.Code)
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

func TestObtenerMesEnEspanol(t *testing.T) {
	if obtenerMesEnEspañol(time.January) != "Enero" {
		t.Errorf("expected Enero")
	}
	if obtenerMesEnEspañol(time.December) != "Diciembre" {
		t.Errorf("expected Diciembre")
	}
}
