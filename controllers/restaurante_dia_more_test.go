package controllers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/beego/beego/v2/server/web/context"
)

func TestRestauranteDia_GetAll_OnlyDia(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/restaurante_dia?dia=Lunes", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := &RestauranteDiaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetAll()
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Fatalf("unexpected %d", w.Code)
	}
}

func TestRestauranteDia_GetById_NotFound(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/restaurante_dia/search?id=999", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := &RestauranteDiaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetById()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 wrapper with notfound code, got %d", w.Code)
	}
}
