package controllers

import (
        "net/http"
        "net/http/httptest"
        "strings"
        "testing"
        "time"

        "restaurante/database"

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
