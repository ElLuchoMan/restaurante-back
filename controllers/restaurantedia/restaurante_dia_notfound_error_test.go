package restaurantedia

import (
	stdctx "context"
	"database/sql/driver"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beego/beego/v2/server/web/context"
)

func TestRestauranteDia_GetAll_DBError(t *testing.T) {
	origQ := MockQuery
	MockQuery = func(_ stdctx.Context, _ string, _ []driver.NamedValue) (driver.Rows, error) {
		return nil, errors.New("db fail")
	}
	t.Cleanup(func() { MockQuery = origQ })

	r := httptest.NewRequest(http.MethodGet, "/restaurante_dia?restaurante_id=1&dia=Lunes", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := &RestauranteDiaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.GetAll()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestRestauranteDia_GetById_NotFound_Explicit(t *testing.T) {
	origQ := MockQuery
	MockQuery = func(_ stdctx.Context, q string, _ []driver.NamedValue) (driver.Rows, error) {
		if strings.Contains(strings.ToLower(q), "from restaurante_dia") {
			// Simula ErrNoRows devolviendo un Rows sin datos
			cols := []string{"restaurante_id", "nombre_restaurante", "hora_apertura", "dia"}
			return &mockRows{columns: cols, values: [][]driver.Value{}}, nil
		}
		return &mockRows{columns: []string{"ok"}, values: [][]driver.Value{{int64(1)}}}, nil
	}
	t.Cleanup(func() { MockQuery = origQ })

	r := httptest.NewRequest(http.MethodGet, "/restaurante_dia/search?id=999", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := &RestauranteDiaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.GetById()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
