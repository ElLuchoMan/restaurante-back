package controllers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beego/beego/v2/server/web/context"
)

func TestNominaGetAllWithoutDB(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/nominas", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
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
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := NominaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestNominaPostMissingDatesForAutoGeneration(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/nominas?generar_nomina_automatica=true", strings.NewReader("{}"))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := NominaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "fecha_inicio") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestNominaPostDBError(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/nominas", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte("{\"estadoNomina\":\"OTRO\"}")
	c := NominaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al crear la nómina") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestNominaPutInvalidID(t *testing.T) {
	r := httptest.NewRequest(http.MethodPut, "/nominas", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
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
	ctx := context.NewContext()
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

func TestNominaDeleteInvalidID(t *testing.T) {
	r := httptest.NewRequest(http.MethodDelete, "/nominas", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
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
	ctx := context.NewContext()
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
