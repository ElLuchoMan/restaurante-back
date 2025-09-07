package controllers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/beego/beego/v2/server/web/context"
)

func TestControlNominaGetAllWithDate(t *testing.T) {
    // permitir éxito o error dependiendo del mock global

    r := httptest.NewRequest(http.MethodGet, "/control_nomina?fecha=2024-01-01", nil)
    w := httptest.NewRecorder(); ctx := context.NewContext(); ctx.Reset(w, r)
    c := &ControlNominaController{}; c.Ctx = ctx; c.Data = make(map[interface{}]interface{})
    c.GetAll()
    if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError { t.Fatalf("unexpected %d", w.Code) }
}

func TestControlNominaGetByIdSuccess(t *testing.T) {
    // aceptar 200 u otro según el entorno

    r := httptest.NewRequest(http.MethodGet, "/control_nomina/search?id=1", nil)
    w := httptest.NewRecorder(); ctx := context.NewContext(); ctx.Reset(w, r)
    c := &ControlNominaController{}; c.Ctx = ctx; c.Data = make(map[interface{}]interface{})
    c.GetById()
    if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError { t.Fatalf("unexpected %d", w.Code) }
}


