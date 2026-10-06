package metodopago

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
)

var cols = []string{"pk_id_metodo_pago", "tipo", "detalle"}

func found() {
	fakeQuery = func(string, []driver.NamedValue) (driver.Rows, error) {
		return rowsOf(cols, []driver.Value{int64(3), "NEQUI", "300"}), nil
	}
}

func TestGetAll(t *testing.T) {
	defer resetFake()
	ctx, w := newCtx(http.MethodGet, "/metodos_pago", "")
	c := &MetodoPagoController{}
	c.Ctx, c.Data = ctx, map[interface{}]interface{}{}
	c.GetAll()
	resp := expect(t, w, http.StatusOK)
	if !strings.Contains(w.Body.String(), `"data":[]`) {
		t.Fatalf("lista vacía debe ser []: %s", w.Body.String())
	}
	_ = resp

	found()
	ctx, w = newCtx(http.MethodGet, "/metodos_pago", "")
	c = &MetodoPagoController{}
	c.Ctx, c.Data = ctx, map[interface{}]interface{}{}
	c.GetAll()
	expect(t, w, http.StatusOK)
	if !strings.Contains(w.Body.String(), `"tipo":"NEQUI"`) {
		t.Fatalf("falta el método: %s", w.Body.String())
	}

	fakeQuery = func(string, []driver.NamedValue) (driver.Rows, error) { return nil, errors.New("boom") }
	ctx, w = newCtx(http.MethodGet, "/metodos_pago", "")
	c = &MetodoPagoController{}
	c.Ctx, c.Data = ctx, map[interface{}]interface{}{}
	c.GetAll()
	expect(t, w, http.StatusInternalServerError)
}

func call(t *testing.T, method, target, body string, f func(c *MetodoPagoController), status int) string {
	t.Helper()
	ctx, w := newCtx(method, target, body)
	c := &MetodoPagoController{}
	c.Ctx, c.Data = ctx, map[interface{}]interface{}{}
	f(c)
	expect(t, w, status)
	return w.Body.String()
}

func TestGetById(t *testing.T) {
	defer resetFake()
	g := func(c *MetodoPagoController) { c.GetById() }
	for _, q := range []string{"", "?id=abc", "?id=0"} {
		call(t, http.MethodGet, "/metodos_pago/search"+q, "", g, http.StatusBadRequest)
	}
	call(t, http.MethodGet, "/metodos_pago/search?id=3", "", g, http.StatusNotFound)
	found()
	if b := call(t, http.MethodGet, "/metodos_pago/search?id=3", "", g, http.StatusOK); !strings.Contains(b, `"metodoPagoId":3`) {
		t.Fatalf("cuerpo inesperado %s", b)
	}
	fakeQuery = func(string, []driver.NamedValue) (driver.Rows, error) { return nil, errors.New("boom") }
	call(t, http.MethodGet, "/metodos_pago/search?id=3", "", g, http.StatusInternalServerError)
}

func TestPost(t *testing.T) {
	defer resetFake()
	p := func(c *MetodoPagoController) { c.Post() }
	call(t, http.MethodPost, "/metodos_pago", "nojson", p, http.StatusBadRequest)
	call(t, http.MethodPost, "/metodos_pago", `{"tipo":"  "}`, p, http.StatusBadRequest)
	if b := call(t, http.MethodPost, "/metodos_pago", `{"tipo":"NEQUI","detalle":"300"}`, p, http.StatusCreated); !strings.Contains(b, `"detalle":"300"`) {
		t.Fatalf("falta detalle: %s", b)
	}
	call(t, http.MethodPost, "/metodos_pago", `{"tipo":"EFECTIVO"}`, p, http.StatusCreated)
	fakeExec = func(string, []driver.NamedValue) (driver.Result, error) {
		return nil, errors.New("duplicate key 23505")
	}
	call(t, http.MethodPost, "/metodos_pago", `{"tipo":"X"}`, p, http.StatusConflict)
	fakeExec = func(string, []driver.NamedValue) (driver.Result, error) { return nil, errors.New("boom") }
	call(t, http.MethodPost, "/metodos_pago", `{"tipo":"X"}`, p, http.StatusInternalServerError)
}

func TestPut(t *testing.T) {
	defer resetFake()
	u := func(c *MetodoPagoController) { c.Put() }
	call(t, http.MethodPut, "/metodos_pago", `{}`, u, http.StatusBadRequest)
	call(t, http.MethodPut, "/metodos_pago?id=3", `{}`, u, http.StatusNotFound)
	fakeQuery = func(string, []driver.NamedValue) (driver.Rows, error) { return nil, errors.New("boom") }
	call(t, http.MethodPut, "/metodos_pago?id=3", `{}`, u, http.StatusInternalServerError)

	found()
	call(t, http.MethodPut, "/metodos_pago?id=3", `nojson`, u, http.StatusBadRequest)
	call(t, http.MethodPut, "/metodos_pago?id=3", `{"tipo":null}`, u, http.StatusBadRequest)
	call(t, http.MethodPut, "/metodos_pago?id=3", `{"detalle":null}`, u, http.StatusBadRequest)
	call(t, http.MethodPut, "/metodos_pago?id=3", `{"tipo":" "}`, u, http.StatusBadRequest)

	// merge: solo detalle; el tipo se conserva
	b := call(t, http.MethodPut, "/metodos_pago?id=3", `{"detalle":"nuevo"}`, u, http.StatusOK)
	var resp struct {
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal([]byte(b), &resp); err != nil || resp.Data["tipo"] != "NEQUI" || resp.Data["detalle"] != "nuevo" {
		t.Fatalf("merge incorrecto: %s", b)
	}
	// merge: solo tipo; el detalle se conserva
	b = call(t, http.MethodPut, "/metodos_pago?id=3", `{"tipo":"DAVIPLATA"}`, u, http.StatusOK)
	if !strings.Contains(b, `"detalle":"300"`) || !strings.Contains(b, `"tipo":"DAVIPLATA"`) {
		t.Fatalf("merge incorrecto: %s", b)
	}
	fakeExec = func(string, []driver.NamedValue) (driver.Result, error) { return nil, errors.New("boom") }
	call(t, http.MethodPut, "/metodos_pago?id=3", `{}`, u, http.StatusInternalServerError)
}

func TestDelete(t *testing.T) {
	defer resetFake()
	d := func(c *MetodoPagoController) { c.Delete() }
	call(t, http.MethodDelete, "/metodos_pago", "", d, http.StatusBadRequest)
	call(t, http.MethodDelete, "/metodos_pago?id=3", "", d, http.StatusOK)
	fakeAffected = 0
	call(t, http.MethodDelete, "/metodos_pago?id=3", "", d, http.StatusNotFound)
	fakeExec = func(string, []driver.NamedValue) (driver.Result, error) {
		return nil, errors.New(`violates foreign key constraint`)
	}
	call(t, http.MethodDelete, "/metodos_pago?id=3", "", d, http.StatusConflict)
	fakeExec = func(string, []driver.NamedValue) (driver.Result, error) { return nil, errors.New("boom") }
	call(t, http.MethodDelete, "/metodos_pago?id=3", "", d, http.StatusInternalServerError)
}
