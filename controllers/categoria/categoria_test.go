package categoria

import (
	"database/sql/driver"
	"errors"
	"net/http"
	"strings"
	"testing"
)

var cols = []string{"pk_id_categoria", "nombre"}

func found() {
	fakeQuery = func(string, []driver.NamedValue) (driver.Rows, error) {
		return rowsOf(cols, []driver.Value{int64(3), "Bebidas"}), nil
	}
}

func boom() {
	fakeQuery = func(string, []driver.NamedValue) (driver.Rows, error) { return nil, errors.New("boom") }
}

func execErr(msg string) {
	fakeExec = func(string, []driver.NamedValue) (driver.Result, error) { return nil, errors.New(msg) }
}

func call(t *testing.T, method, target, body string, f func(c *CategoriaController), status int) string {
	t.Helper()
	ctx, w := newCtx(method, target, body)
	c := &CategoriaController{}
	c.Ctx, c.Data = ctx, map[interface{}]interface{}{}
	f(c)
	expect(t, w, status)
	return w.Body.String()
}

var (
	getAll  = func(c *CategoriaController) { c.GetAll() }
	getByID = func(c *CategoriaController) { c.GetById() }
	post    = func(c *CategoriaController) { c.Post() }
	put     = func(c *CategoriaController) { c.Put() }
	del     = func(c *CategoriaController) { c.Delete() }
)

func TestGetAll(t *testing.T) {
	defer resetFake()
	if b := call(t, http.MethodGet, "/categorias", "", getAll, http.StatusOK); !strings.Contains(b, `"data":[]`) {
		t.Fatalf("lista vacía debe ser []: %s", b)
	}
	found()
	if b := call(t, http.MethodGet, "/categorias", "", getAll, http.StatusOK); !strings.Contains(b, `"nombre":"Bebidas"`) {
		t.Fatalf("falta la categoría: %s", b)
	}
	boom()
	call(t, http.MethodGet, "/categorias", "", getAll, http.StatusInternalServerError)
}

func TestGetById(t *testing.T) {
	defer resetFake()
	for _, q := range []string{"", "?id=abc", "?id=0", "?id=-1"} {
		call(t, http.MethodGet, "/categorias/search"+q, "", getByID, http.StatusBadRequest)
	}
	call(t, http.MethodGet, "/categorias/search?id=3", "", getByID, http.StatusNotFound)
	found()
	if b := call(t, http.MethodGet, "/categorias/search?id=3", "", getByID, http.StatusOK); !strings.Contains(b, `"categoriaId":3`) {
		t.Fatalf("cuerpo inesperado %s", b)
	}
	boom()
	call(t, http.MethodGet, "/categorias/search?id=3", "", getByID, http.StatusInternalServerError)
}

func TestPost(t *testing.T) {
	defer resetFake()
	call(t, http.MethodPost, "/categorias", "", post, http.StatusBadRequest)
	call(t, http.MethodPost, "/categorias", "nojson", post, http.StatusBadRequest)
	call(t, http.MethodPost, "/categorias", `{"nombre":"  "}`, post, http.StatusBadRequest)
	call(t, http.MethodPost, "/categorias", `{}`, post, http.StatusBadRequest)
	if b := call(t, http.MethodPost, "/categorias", `{"nombre":" Postres "}`, post, http.StatusCreated); !strings.Contains(b, `"nombre":"Postres"`) {
		t.Fatalf("cuerpo inesperado %s", b)
	}
	execErr("duplicate key value violates unique constraint (23505)")
	call(t, http.MethodPost, "/categorias", `{"nombre":"Postres"}`, post, http.StatusConflict)
	execErr("boom")
	call(t, http.MethodPost, "/categorias", `{"nombre":"Postres"}`, post, http.StatusInternalServerError)
}

func TestPut(t *testing.T) {
	defer resetFake()
	call(t, http.MethodPut, "/categorias", `{}`, put, http.StatusBadRequest)
	call(t, http.MethodPut, "/categorias?id=3", `{}`, put, http.StatusNotFound)
	boom()
	call(t, http.MethodPut, "/categorias?id=3", `{}`, put, http.StatusInternalServerError)

	found()
	call(t, http.MethodPut, "/categorias?id=3", "", put, http.StatusBadRequest)
	call(t, http.MethodPut, "/categorias?id=3", `{"nombre":null}`, put, http.StatusBadRequest)
	call(t, http.MethodPut, "/categorias?id=3", `{"nombre":" "}`, put, http.StatusBadRequest)
	if b := call(t, http.MethodPut, "/categorias?id=3", `{}`, put, http.StatusOK); !strings.Contains(b, `"nombre":"Bebidas"`) {
		t.Fatalf("sin cambios debe conservar: %s", b)
	}
	if b := call(t, http.MethodPut, "/categorias?id=3", `{"nombre":"Jugos"}`, put, http.StatusOK); !strings.Contains(b, `"nombre":"Jugos"`) {
		t.Fatalf("cuerpo inesperado %s", b)
	}
	execErr("duplicate key value (23505)")
	call(t, http.MethodPut, "/categorias?id=3", `{"nombre":"Jugos"}`, put, http.StatusConflict)
	execErr("boom")
	call(t, http.MethodPut, "/categorias?id=3", `{"nombre":"Jugos"}`, put, http.StatusInternalServerError)
}

func TestDelete(t *testing.T) {
	defer resetFake()
	call(t, http.MethodDelete, "/categorias", "", del, http.StatusBadRequest)
	call(t, http.MethodDelete, "/categorias?id=3", "", del, http.StatusNotFound)
	found()
	call(t, http.MethodDelete, "/categorias?id=3", "", del, http.StatusOK)
	execErr("violates foreign key constraint (23503)")
	call(t, http.MethodDelete, "/categorias?id=3", "", del, http.StatusConflict)
	execErr("boom")
	call(t, http.MethodDelete, "/categorias?id=3", "", del, http.StatusInternalServerError)
}
