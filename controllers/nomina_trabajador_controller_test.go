package controllers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
	c := NominaTrabajadorController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}
