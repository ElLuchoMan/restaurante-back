package restaurantedia

import (
	"database/sql/driver"
	"errors"
	"net/http"
	"strings"
	"testing"
)

var cols = []string{"restaurante_dia_id", "restaurante_id", "nombre_restaurante", "hora_apertura", "dia"}

func call(t *testing.T, target string, f func(c *RestauranteDiaController), status int) string {
	t.Helper()
	ctx, w := newCtx(http.MethodGet, target, "")
	c := &RestauranteDiaController{}
	c.Ctx, c.Data = ctx, map[interface{}]interface{}{}
	f(c)
	expect(t, w, status)
	return w.Body.String()
}

func TestGetAll(t *testing.T) {
	defer resetFake()
	g := func(c *RestauranteDiaController) { c.GetAll() }
	if b := call(t, "/restaurante_dia", g, http.StatusOK); !strings.Contains(b, `"data":[]`) {
		t.Fatalf("lista vacía debe ser []: %s", b)
	}
	var gotQuery string
	var gotArgs []driver.NamedValue
	fakeQuery = func(q string, a []driver.NamedValue) (driver.Rows, error) {
		gotQuery, gotArgs = q, a
		return rowsOf(cols, []driver.Value{int64(5), int64(1), "Mi Resto", "09:00:00", "Miércoles"}), nil
	}
	b := call(t, "/restaurante_dia?restaurante_id=1&dia=miercoles", g, http.StatusOK)
	if !strings.Contains(b, `"restauranteDiaId":5`) || !strings.Contains(b, `"horaApertura":"09:00:00"`) {
		t.Fatalf("falta restauranteDiaId u hora: %s", b)
	}
	if !strings.Contains(gotQuery, "rd.dia") || len(gotArgs) != 2 || gotArgs[1].Value != "Miércoles" {
		t.Fatalf("filtros no aplicados: %s %v", gotQuery, gotArgs)
	}
	call(t, "/restaurante_dia?restaurante_id=abc", g, http.StatusBadRequest)
	call(t, "/restaurante_dia?dia=Funday", g, http.StatusBadRequest)
	fakeQuery = func(string, []driver.NamedValue) (driver.Rows, error) { return nil, errors.New("boom") }
	call(t, "/restaurante_dia", g, http.StatusInternalServerError)
}

func TestGetById(t *testing.T) {
	defer resetFake()
	g := func(c *RestauranteDiaController) { c.GetById() }
	call(t, "/restaurante_dia/search", g, http.StatusBadRequest)
	call(t, "/restaurante_dia/search?id=9", g, http.StatusNotFound)
	fakeQuery = func(string, []driver.NamedValue) (driver.Rows, error) {
		return rowsOf(cols, []driver.Value{int64(9), int64(1), "Mi Resto", "09:00:00", "Lunes"}), nil
	}
	if b := call(t, "/restaurante_dia/search?id=9", g, http.StatusOK); !strings.Contains(b, `"restauranteDiaId":9`) {
		t.Fatalf("cuerpo inesperado: %s", b)
	}
	fakeQuery = func(string, []driver.NamedValue) (driver.Rows, error) { return nil, errors.New("boom") }
	call(t, "/restaurante_dia/search?id=9", g, http.StatusInternalServerError)
}

func TestCanonicalDia(t *testing.T) {
	if canonicalDia(" sábado ") != "Sábado" || canonicalDia("x") != "" {
		t.Fatalf("canonicalDia incorrecto")
	}
}
