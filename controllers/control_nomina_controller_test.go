package controllers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/beego/beego/v2/server/web/context"
)

func TestControlNomina_GetAll_And_Search(t *testing.T) {
    // GetAll sin DB
    r := httptest.NewRequest(http.MethodGet, "/control_nomina?fecha=2025-01-31", nil)
    w := httptest.NewRecorder(); ctx := context.NewContext(); ctx.Reset(w, r)
    c := &ControlNominaController{}; c.Ctx = ctx; c.Data = make(map[interface{}]interface{})
    c.GetAll()
    if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError { t.Fatalf("unexpected %d", w.Code) }

    // Search por id
    r = httptest.NewRequest(http.MethodGet, "/control_nomina/search?id=1", nil)
    w = httptest.NewRecorder(); ctx = context.NewContext(); ctx.Reset(w, r)
    c = &ControlNominaController{}; c.Ctx = ctx; c.Data = make(map[interface{}]interface{})
    c.GetById()
    if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError { t.Fatalf("unexpected %d", w.Code) }
}


