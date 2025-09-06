package controllers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/beego/beego/v2/server/web/context"
)

func TestRestauranteDia_GetAll_And_Search(t *testing.T) {
    r := httptest.NewRequest(http.MethodGet, "/restaurante_dia?restaurante_id=1&dia=Lunes", nil)
    w := httptest.NewRecorder(); ctx := context.NewContext(); ctx.Reset(w, r)
    c := &RestauranteDiaController{}; c.Ctx = ctx; c.Data = make(map[interface{}]interface{})
    c.GetAll()
    if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError { t.Fatalf("unexpected %d", w.Code) }

    r = httptest.NewRequest(http.MethodGet, "/restaurante_dia/search?id=1", nil)
    w = httptest.NewRecorder(); ctx = context.NewContext(); ctx.Reset(w, r)
    c = &RestauranteDiaController{}; c.Ctx = ctx; c.Data = make(map[interface{}]interface{})
    c.GetById()
    if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError { t.Fatalf("unexpected %d", w.Code) }
}


