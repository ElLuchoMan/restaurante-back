package restaurante

import (
	"database/sql/driver"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"
)

var cols = []string{"pk_id_restaurante", "nombre_restaurante", "hora_apertura", "pk_id_cambio_horario"}

func call(t *testing.T, target string, f func(c *RestauranteController), status int) string {
	t.Helper()
	ctx, w := newCtx(http.MethodGet, target, "")
	c := &RestauranteController{}
	c.Ctx, c.Data = ctx, map[interface{}]interface{}{}
	f(c)
	expect(t, w, status)
	return w.Body.String()
}

func oneRow() {
	fakeQuery = func(string, []driver.NamedValue) (driver.Rows, error) {
		return rowsOf(cols, []driver.Value{int64(1), "Mi Resto", time.Date(1, 1, 1, 9, 0, 0, 0, time.UTC), nil}), nil
	}
}

func TestGetAll(t *testing.T) {
	defer resetFake()
	g := func(c *RestauranteController) { c.GetAll() }
	if b := call(t, "/restaurantes", g, http.StatusOK); !strings.Contains(b, `"data":[]`) {
		t.Fatalf("lista vacía debe ser []: %s", b)
	}
	oneRow()
	if b := call(t, "/restaurantes", g, http.StatusOK); !strings.Contains(b, `"nombreRestaurante":"Mi Resto"`) {
		t.Fatalf("falta el restaurante: %s", b)
	}
	fakeQuery = func(string, []driver.NamedValue) (driver.Rows, error) { return nil, errors.New("boom") }
	call(t, "/restaurantes", g, http.StatusInternalServerError)
}

func TestGetById(t *testing.T) {
	defer resetFake()
	g := func(c *RestauranteController) { c.GetById() }
	call(t, "/restaurantes/search", g, http.StatusBadRequest)
	call(t, "/restaurantes/search?id=0", g, http.StatusBadRequest)
	call(t, "/restaurantes/search?id=1", g, http.StatusNotFound)
	oneRow()
	if b := call(t, "/restaurantes/search?id=1", g, http.StatusOK); !strings.Contains(b, `"restauranteId":1`) {
		t.Fatalf("cuerpo inesperado: %s", b)
	}
	fakeQuery = func(string, []driver.NamedValue) (driver.Rows, error) { return nil, errors.New("boom") }
	call(t, "/restaurantes/search?id=1", g, http.StatusInternalServerError)
}
