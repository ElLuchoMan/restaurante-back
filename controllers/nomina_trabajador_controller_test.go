package controllers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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
	body := `{"PK_DOCUMENTO_TRABAJADOR":0}`
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
	if !strings.Contains(w.Body.String(), "PK_DOCUMENTO_TRABAJADOR") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestNominaTrabajadorPostDBError(t *testing.T) {
	body := `{"PK_DOCUMENTO_TRABAJADOR":123}`
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
