package controllers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
