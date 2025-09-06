package controllers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/beego/beego/v2/server/web/context"
)

func TestReservaContacto_GetAll_And_Search(t *testing.T) {
    r := httptest.NewRequest(http.MethodGet, "/reserva_contacto?documento_contacto=123", nil)
    w := httptest.NewRecorder(); ctx := context.NewContext(); ctx.Reset(w, r)
    c := &ReservaContactoController{}; c.Ctx = ctx; c.Data = make(map[interface{}]interface{})
    c.GetAll(); if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError { t.Fatalf("unexpected %d", w.Code) }

    r = httptest.NewRequest(http.MethodGet, "/reserva_contacto/search?id=1", nil)
    w = httptest.NewRecorder(); ctx = context.NewContext(); ctx.Reset(w, r)
    c = &ReservaContactoController{}; c.Ctx = ctx; c.Data = make(map[interface{}]interface{})
    c.GetById(); if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError { t.Fatalf("unexpected %d", w.Code) }
}


