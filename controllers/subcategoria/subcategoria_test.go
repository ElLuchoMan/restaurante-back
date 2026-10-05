package subcategoria

import (
	"database/sql/driver"
	"errors"
	"net/http"
	"strings"
	"testing"
)

var cols = []string{"pk_id_subcategoria", "nombre", "pk_id_categoria", "c_pk", "c_nombre"}

func found() {
	fakeQuery = func(string, []driver.NamedValue) (driver.Rows, error) {
		return rowsOf(cols, []driver.Value{int64(5), "Gaseosas", int64(3), int64(3), "Bebidas"}), nil
	}
}

func boom() {
	fakeQuery = func(string, []driver.NamedValue) (driver.Rows, error) { return nil, errors.New("boom") }
}

func execErr(msg string) {
	fakeExec = func(string, []driver.NamedValue) (driver.Result, error) { return nil, errors.New(msg) }
}

func call(t *testing.T, method, target, body string, f func(c *SubcategoriaController), status int) string {
	t.Helper()
	ctx, w := newCtx(method, target, body)
	c := &SubcategoriaController{}
	c.Ctx, c.Data = ctx, map[interface{}]interface{}{}
	f(c)
	expect(t, w, status)
	return w.Body.String()
}

var (
	getAll  = func(c *SubcategoriaController) { c.GetAll() }
	getByID = func(c *SubcategoriaController) { c.GetById() }
	post    = func(c *SubcategoriaController) { c.Post() }
	put     = func(c *SubcategoriaController) { c.Put() }
	del     = func(c *SubcategoriaController) { c.Delete() }
)

func TestGetAll(t *testing.T) {
	defer resetFake()
	if b := call(t, http.MethodGet, "/subcategorias", "", getAll, http.StatusOK); !strings.Contains(b, `"data":[]`) {
		t.Fatalf("lista vacía debe ser []: %s", b)
	}
	found()
	if b := call(t, http.MethodGet, "/subcategorias?categoria_id=3", "", getAll, http.StatusOK); !strings.Contains(b, `"categoriaId":{"categoriaId":3,"nombre":"Bebidas"}`) {
		t.Fatalf("falta la categoría embebida: %s", b)
	}
	for _, q := range []string{"?categoria_id=abc", "?categoria_id=0"} {
		call(t, http.MethodGet, "/subcategorias"+q, "", getAll, http.StatusBadRequest)
	}
	boom()
	call(t, http.MethodGet, "/subcategorias", "", getAll, http.StatusInternalServerError)
}

func TestGetById(t *testing.T) {
	defer resetFake()
	for _, q := range []string{"", "?id=abc", "?id=0"} {
		call(t, http.MethodGet, "/subcategorias/search"+q, "", getByID, http.StatusBadRequest)
	}
	call(t, http.MethodGet, "/subcategorias/search?id=5", "", getByID, http.StatusNotFound)
	found()
	if b := call(t, http.MethodGet, "/subcategorias/search?id=5", "", getByID, http.StatusOK); !strings.Contains(b, `"subcategoriaId":5`) {
		t.Fatalf("cuerpo inesperado %s", b)
	}
	boom()
	call(t, http.MethodGet, "/subcategorias/search?id=5", "", getByID, http.StatusInternalServerError)
}

func TestPost(t *testing.T) {
	defer resetFake()
	call(t, http.MethodPost, "/subcategorias", "nojson", post, http.StatusBadRequest)
	for _, b := range []string{`{}`, `{"nombre":"x"}`, `{"categoriaId":1}`, `{"nombre":" ","categoriaId":1}`, `{"nombre":"x","categoriaId":0}`} {
		call(t, http.MethodPost, "/subcategorias", b, post, http.StatusBadRequest)
	}
	// Tras insertar, la lectura no encuentra la fila.
	call(t, http.MethodPost, "/subcategorias", `{"nombre":"Gaseosas","categoriaId":3}`, post, http.StatusNotFound)
	found()
	if b := call(t, http.MethodPost, "/subcategorias", `{"nombre":" Gaseosas ","categoriaId":3}`, post, http.StatusCreated); !strings.Contains(b, `"nombre":"Gaseosas"`) {
		t.Fatalf("cuerpo inesperado %s", b)
	}
	execErr("duplicate key value (23505)")
	call(t, http.MethodPost, "/subcategorias", `{"nombre":"Gaseosas","categoriaId":3}`, post, http.StatusConflict)
	execErr("violates foreign key constraint (23503)")
	call(t, http.MethodPost, "/subcategorias", `{"nombre":"Gaseosas","categoriaId":99}`, post, http.StatusBadRequest)
	execErr("boom")
	call(t, http.MethodPost, "/subcategorias", `{"nombre":"Gaseosas","categoriaId":3}`, post, http.StatusInternalServerError)
}

func TestPut(t *testing.T) {
	defer resetFake()
	call(t, http.MethodPut, "/subcategorias", `{}`, put, http.StatusBadRequest)
	call(t, http.MethodPut, "/subcategorias?id=5", `{}`, put, http.StatusNotFound)
	boom()
	call(t, http.MethodPut, "/subcategorias?id=5", `{}`, put, http.StatusInternalServerError)

	found()
	call(t, http.MethodPut, "/subcategorias?id=5", "", put, http.StatusBadRequest)
	call(t, http.MethodPut, "/subcategorias?id=5", `{"nombre":null}`, put, http.StatusBadRequest)
	call(t, http.MethodPut, "/subcategorias?id=5", `{"categoriaId":null}`, put, http.StatusBadRequest)
	call(t, http.MethodPut, "/subcategorias?id=5", `{"nombre":" "}`, put, http.StatusBadRequest)
	call(t, http.MethodPut, "/subcategorias?id=5", `{"categoriaId":0}`, put, http.StatusBadRequest)
	if b := call(t, http.MethodPut, "/subcategorias?id=5", `{}`, put, http.StatusOK); !strings.Contains(b, `"nombre":"Gaseosas"`) {
		t.Fatalf("sin cambios debe conservar: %s", b)
	}
	if b := call(t, http.MethodPut, "/subcategorias?id=5", `{"nombre":"Jugos","categoriaId":3}`, put, http.StatusOK); !strings.Contains(b, `"subcategoriaId":5`) {
		t.Fatalf("cuerpo inesperado %s", b)
	}
	execErr("duplicate key value (23505)")
	call(t, http.MethodPut, "/subcategorias?id=5", `{"nombre":"Jugos"}`, put, http.StatusConflict)
	execErr("violates foreign key constraint (23503)")
	call(t, http.MethodPut, "/subcategorias?id=5", `{"categoriaId":99}`, put, http.StatusBadRequest)
	execErr("boom")
	call(t, http.MethodPut, "/subcategorias?id=5", `{"nombre":"Jugos"}`, put, http.StatusInternalServerError)

	// La relectura posterior a la actualización falla.
	fakeExec = nil
	n := 0
	fakeQuery = func(string, []driver.NamedValue) (driver.Rows, error) {
		n++
		if n == 1 {
			return rowsOf(cols, []driver.Value{int64(5), "Gaseosas", int64(3), int64(3), "Bebidas"}), nil
		}
		return nil, errors.New("boom")
	}
	call(t, http.MethodPut, "/subcategorias?id=5", `{"nombre":"Jugos"}`, put, http.StatusInternalServerError)
}

func TestDelete(t *testing.T) {
	defer resetFake()
	call(t, http.MethodDelete, "/subcategorias", "", del, http.StatusBadRequest)
	call(t, http.MethodDelete, "/subcategorias?id=5", "", del, http.StatusNotFound)
	found()
	call(t, http.MethodDelete, "/subcategorias?id=5", "", del, http.StatusOK)
	execErr("violates foreign key constraint (23503)")
	call(t, http.MethodDelete, "/subcategorias?id=5", "", del, http.StatusConflict)
	execErr("boom")
	call(t, http.MethodDelete, "/subcategorias?id=5", "", del, http.StatusInternalServerError)
}
