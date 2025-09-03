package controllers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beego/beego/v2/server/web/context"
)

func TestIncidenciaGetAllWithoutDB(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/incidencias", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := IncidenciaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetAll()

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestIncidenciaGetByDocumentAndDateInvalidDocument(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/incidencias/search", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := IncidenciaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetByDocumentAndDate()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestIncidenciaGetByDocumentAndDateInvalidMonth(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/incidencias/search?documento=1&mes=13&anio=2024", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := IncidenciaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetByDocumentAndDate()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestIncidenciaGetByDocumentAndDateInvalidYear(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/incidencias/search?documento=1&mes=1&anio=1800", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := IncidenciaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetByDocumentAndDate()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestIncidenciaGetByDocumentAndDateNoResults(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/incidencias/search?documento=1&mes=1&anio=2024", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := IncidenciaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetByDocumentAndDate()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "No se encontraron incidencias") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestIncidenciaPostInvalidJSON(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/incidencias", strings.NewReader("notjson"))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := IncidenciaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestIncidenciaPostMissingFecha(t *testing.T) {
	body := `{"monto":100,"resta":false,"motivo":"test"}`
	r := httptest.NewRequest(http.MethodPost, "/incidencias", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := IncidenciaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestIncidenciaPostInvalidFecha(t *testing.T) {
	body := `{"fechaIncidencia":"invalid","monto":100,"resta":false,"motivo":"test"}`
	r := httptest.NewRequest(http.MethodPost, "/incidencias", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := IncidenciaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestIncidenciaPostMissingMonto(t *testing.T) {
	body := `{"fechaIncidencia":"2024-01-01","resta":false,"motivo":"test"}`
	r := httptest.NewRequest(http.MethodPost, "/incidencias", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := IncidenciaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestIncidenciaPostMissingDocumentoTrabajador(t *testing.T) {
	body := `{"fechaIncidencia":"2024-01-01","monto":100,"resta":false,"motivo":"test"}`
	r := httptest.NewRequest(http.MethodPost, "/incidencias", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := IncidenciaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}
func TestIncidenciaPostMissingResta(t *testing.T) {
	body := `{"fechaIncidencia":"2024-01-01","monto":100,"motivo":"test","documentoTrabajador":1}`
	r := httptest.NewRequest(http.MethodPost, "/incidencias", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := IncidenciaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	ctx.Input.CopyBody(1 << 20)
	c.Post()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestIncidenciaPostMissingMotivo(t *testing.T) {
	body := `{"fechaIncidencia":"2024-01-01","monto":100,"resta":false,"documentoTrabajador":1}`
	r := httptest.NewRequest(http.MethodPost, "/incidencias", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := IncidenciaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	ctx.Input.CopyBody(1 << 20)
	c.Post()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestIncidenciaPutInvalidID(t *testing.T) {
	r := httptest.NewRequest(http.MethodPut, "/incidencias?id=abc", strings.NewReader("{}"))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := IncidenciaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Put()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestIncidenciaPutInvalidJSON(t *testing.T) {
	r := httptest.NewRequest(http.MethodPut, "/incidencias?id=1", strings.NewReader("notjson"))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := IncidenciaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Put()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestIncidenciaPutInvalidFecha(t *testing.T) {
	body := `{"fechaIncidencia":"invalid"}`
	r := httptest.NewRequest(http.MethodPut, "/incidencias?id=1", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := IncidenciaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Put()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestIncidenciaPutUpdateError(t *testing.T) {
	body := `{"fechaIncidencia":"2024-01-01","monto":100,"resta":false,"motivo":"test","documentoTrabajador":1}`
	r := httptest.NewRequest(http.MethodPut, "/incidencias?id=1", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := IncidenciaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	ctx.Input.CopyBody(1 << 20)
	c.Put()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestIncidenciaDeleteInvalidID(t *testing.T) {
	r := httptest.NewRequest(http.MethodDelete, "/incidencias?id=abc", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := IncidenciaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Delete()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestIncidenciaDeleteError(t *testing.T) {
	r := httptest.NewRequest(http.MethodDelete, "/incidencias?id=1", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := IncidenciaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Delete()

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}
