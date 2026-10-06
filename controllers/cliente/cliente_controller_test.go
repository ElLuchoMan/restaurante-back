package cliente

import (
	"database/sql/driver"
	"errors"
	"net/http"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

var (
	clienteCols = []string{"pk_documento_cliente", "nombre", "apellido", "correo", "direccion", "telefono", "observaciones", "password"}
	errBoom     = errors.New("boom")
)

func clienteRow() []driver.Value {
	return []driver.Value{int64(1001), "Juan", "Pérez", "juan@example.com", "Calle 1", "3001234567", "VIP", "$2a$hash-secreto"}
}

// serveCliente programa el driver: el cliente existe.
func serveCliente() {
	fakeQuery = func(string, []driver.NamedValue) (driver.Rows, error) {
		return rowsOf(clienteCols, clienteRow()), nil
	}
}

func failQuery() {
	fakeQuery = func(string, []driver.NamedValue) (driver.Rows, error) { return nil, errBoom }
}

func call(t *testing.T, method, target, body string, f func(c *ClienteController), status int) string {
	t.Helper()
	ctx, w := newCtx(method, target, body)
	c := &ClienteController{}
	c.Ctx, c.Data = ctx, map[interface{}]interface{}{}
	f(c)
	expect(t, w, status)
	return w.Body.String()
}

func noPassword(t *testing.T, body string) {
	t.Helper()
	if strings.Contains(strings.ToLower(body), "password") || strings.Contains(body, "secreto") {
		t.Fatalf("la respuesta no debe incluir la contraseña: %s", body)
	}
}

func TestGetAll(t *testing.T) {
	defer resetFake()
	g := func(c *ClienteController) { c.GetAll() }

	if b := call(t, http.MethodGet, "/clientes", "", g, http.StatusOK); !strings.Contains(b, `"data":[]`) {
		t.Fatalf("lista vacía debe ser []: %s", b)
	}
	if b := call(t, http.MethodGet, "/clientes?fields=nombre_completo_telefono", "", g, http.StatusOK); !strings.Contains(b, `"data":[]`) {
		t.Fatalf("proyección vacía debe ser []: %s", b)
	}

	var gotQ string
	fakeQuery = func(q string, _ []driver.NamedValue) (driver.Rows, error) {
		gotQ = q
		return rowsOf(clienteCols, clienteRow()), nil
	}
	b := call(t, http.MethodGet, "/clientes", "", g, http.StatusOK)
	noPassword(t, b)
	for _, want := range []string{`"documentoCliente":1001`, `"nombre":"Juan"`, `"observaciones":"VIP"`} {
		if !strings.Contains(b, want) {
			t.Fatalf("falta %s en %s", want, b)
		}
	}
	if !strings.Contains(gotQ, "ORDER BY") || strings.Contains(gotQ, "LIMIT") {
		t.Fatalf("sin limit debe ordenar y no limitar: %s", gotQ)
	}

	call(t, http.MethodGet, "/clientes?limit=5&offset=2", "", g, http.StatusOK)
	if !strings.Contains(gotQ, "LIMIT 5") || !strings.Contains(gotQ, "OFFSET 2") {
		t.Fatalf("paginación no aplicada: %s", gotQ)
	}
	call(t, http.MethodGet, "/clientes?limit=100", "", g, http.StatusOK)

	b = call(t, http.MethodGet, "/clientes?fields=nombre_completo_telefono", "", g, http.StatusOK)
	for _, want := range []string{`"nombre_completo":"Juan Pérez"`, `"telefono":"3001234567"`, `"documentoCliente":1001`} {
		if !strings.Contains(b, want) {
			t.Fatalf("falta %s en %s", want, b)
		}
	}
	noPassword(t, b)

	for _, q := range []string{"fields=otro", "limit=0", "limit=101", "limit=-1", "limit=abc", "offset=-1", "offset=abc"} {
		call(t, http.MethodGet, "/clientes?"+q, "", g, http.StatusBadRequest)
	}

	failQuery()
	call(t, http.MethodGet, "/clientes", "", g, http.StatusInternalServerError)
}

func TestGetById(t *testing.T) {
	defer resetFake()
	g := func(c *ClienteController) { c.GetById() }

	for _, q := range []string{"", "?id=0", "?id=-3", "?id=abc"} {
		call(t, http.MethodGet, "/clientes/search"+q, "", g, http.StatusBadRequest)
	}
	call(t, http.MethodGet, "/clientes/search?id=1001", "", g, http.StatusNotFound)

	serveCliente()
	b := call(t, http.MethodGet, "/clientes/search?id=1001", "", g, http.StatusOK)
	noPassword(t, b)
	if !strings.Contains(b, `"correo":"juan@example.com"`) {
		t.Fatalf("cliente incompleto: %s", b)
	}

	failQuery()
	call(t, http.MethodGet, "/clientes/search?id=1001", "", g, http.StatusInternalServerError)
}

const bodyOK = `{"documentoCliente":1001,"nombre":" Juan ","apellido":"Pérez","correo":" Juan@Example.COM ","password":"Secreta123","telefono":" 3001234567 ","direccion":" Calle 1 ","observaciones":"VIP"}`

func TestPost(t *testing.T) {
	defer resetFake()
	p := func(c *ClienteController) { c.Post() }

	var execArgs []driver.NamedValue
	fakeExec = func(_ string, a []driver.NamedValue) (driver.Result, error) {
		execArgs = a
		return fakeResult{}, nil
	}
	b := call(t, http.MethodPost, "/clientes", bodyOK, p, http.StatusCreated)
	noPassword(t, b)
	for _, want := range []string{`"documentoCliente":1001`, `"nombre":"Juan"`, `"correo":"juan@example.com"`, `"telefono":"3001234567"`, `"direccion":"Calle 1"`} {
		if !strings.Contains(b, want) {
			t.Fatalf("falta %s en %s", want, b)
		}
	}
	hashed := false
	for _, a := range execArgs {
		if s, ok := a.Value.(string); ok {
			if s == "Secreta123" {
				t.Fatalf("la contraseña se guardó en claro")
			}
			if bcrypt.CompareHashAndPassword([]byte(s), []byte("Secreta123")) == nil {
				hashed = true
			}
		}
	}
	if !hashed {
		t.Fatalf("no se guardó el hash bcrypt: %v", execArgs)
	}

	// opcionales omitidos
	call(t, http.MethodPost, "/clientes", `{"documentoCliente":5,"nombre":"A","apellido":"B","correo":"a@b.co","password":"x","telefono":"1"}`, p, http.StatusCreated)

	bads := []string{
		``,
		`{`,
		`[]`,
		`{"documentoCliente":"x"}`,
		`{"nombre":"A","apellido":"B","correo":"a@b.co","password":"x","telefono":"1"}`,
		`{"documentoCliente":0,"nombre":"A","apellido":"B","correo":"a@b.co","password":"x","telefono":"1"}`,
		`{"documentoCliente":1,"nombre":" ","apellido":"B","correo":"a@b.co","password":"x","telefono":"1"}`,
		`{"documentoCliente":1,"nombre":"A","apellido":"","correo":"a@b.co","password":"x","telefono":"1"}`,
		`{"documentoCliente":1,"nombre":"A","apellido":"B","correo":" ","password":"x","telefono":"1"}`,
		`{"documentoCliente":1,"nombre":"A","apellido":"B","correo":"no-es-correo","password":"x","telefono":"1"}`,
		`{"documentoCliente":1,"nombre":"A","apellido":"B","correo":"Ana <a@b.co>","password":"x","telefono":"1"}`,
		`{"documentoCliente":1,"nombre":"A","apellido":"B","correo":"a@b.co","password":"x"}`,
		`{"documentoCliente":1,"nombre":"A","apellido":"B","correo":"a@b.co","password":"x","telefono":" "}`,
		`{"documentoCliente":1,"nombre":"A","apellido":"B","correo":"a@b.co","telefono":"1"}`,
		`{"documentoCliente":1,"nombre":"A","apellido":"B","correo":"a@b.co","password":"` + strings.Repeat("x", 73) + `","telefono":"1"}`,
	}
	for _, body := range bads {
		call(t, http.MethodPost, "/clientes", body, p, http.StatusBadRequest)
	}

	// 72 bytes es el máximo admitido
	call(t, http.MethodPost, "/clientes", `{"documentoCliente":1,"nombre":"A","apellido":"B","correo":"a@b.co","password":"`+strings.Repeat("x", 72)+`","telefono":"1"}`, p, http.StatusCreated)

	// error de hash
	orig := generateFromPassword
	generateFromPassword = func([]byte, int) ([]byte, error) { return nil, errBoom }
	call(t, http.MethodPost, "/clientes", bodyOK, p, http.StatusInternalServerError)
	generateFromPassword = orig

	// conflictos y errores de BD
	cases := []struct {
		err    string
		status int
		msg    string
	}{
		{`pq: duplicate key value violates unique constraint "cliente_correo_key"`, http.StatusConflict, "correo"},
		{`pq: duplicate key value violates unique constraint "cliente_telefono_key"`, http.StatusConflict, "teléfono"},
		{`pq: duplicate key value violates unique constraint "cliente_pkey"`, http.StatusConflict, "documento"},
		{`conexión perdida`, http.StatusInternalServerError, ""},
	}
	for _, tc := range cases {
		e := errors.New(tc.err)
		fakeExec = func(string, []driver.NamedValue) (driver.Result, error) { return nil, e }
		b := call(t, http.MethodPost, "/clientes", bodyOK, p, tc.status)
		if !strings.Contains(b, tc.msg) {
			t.Fatalf("mensaje sin %q: %s", tc.msg, b)
		}
	}
}

func TestPut(t *testing.T) {
	defer resetFake()
	u := func(c *ClienteController) { c.Put() }

	for _, q := range []string{"", "?id=0", "?id=abc"} {
		call(t, http.MethodPut, "/clientes"+q, `{}`, u, http.StatusBadRequest)
	}
	call(t, http.MethodPut, "/clientes?id=1001", `{}`, u, http.StatusNotFound)
	failQuery()
	call(t, http.MethodPut, "/clientes?id=1001", `{}`, u, http.StatusInternalServerError)

	serveCliente()
	var execArgs []driver.NamedValue
	fakeExec = func(_ string, a []driver.NamedValue) (driver.Result, error) {
		execArgs = a
		return fakeResult{}, nil
	}

	// merge: los ausentes se conservan
	b := call(t, http.MethodPut, "/clientes?id=1001", `{"nombre":" Pedro "}`, u, http.StatusOK)
	noPassword(t, b)
	for _, want := range []string{`"nombre":"Pedro"`, `"apellido":"Pérez"`, `"correo":"juan@example.com"`, `"telefono":"3001234567"`, `"direccion":"Calle 1"`, `"observaciones":"VIP"`} {
		if !strings.Contains(b, want) {
			t.Fatalf("falta %s en %s", want, b)
		}
	}
	// la contraseña existente no cambia
	kept := false
	for _, a := range execArgs {
		if a.Value == "$2a$hash-secreto" {
			kept = true
		}
	}
	if !kept {
		t.Fatalf("la contraseña original debía conservarse: %v", execArgs)
	}

	// todos los campos + observaciones null (único anulable)
	all := `{"nombre":"A","apellido":"B","correo":" NUEVO@x.co ","telefono":"3","direccion":"","observaciones":null,"password":"Nueva123"}`
	b = call(t, http.MethodPut, "/clientes?id=1001", all, u, http.StatusOK)
	noPassword(t, b)
	for _, want := range []string{`"correo":"nuevo@x.co"`, `"direccion":""`, `"observaciones":null`} {
		if !strings.Contains(b, want) {
			t.Fatalf("falta %s en %s", want, b)
		}
	}
	ok := false
	for _, a := range execArgs {
		if s, isStr := a.Value.(string); isStr && bcrypt.CompareHashAndPassword([]byte(s), []byte("Nueva123")) == nil {
			ok = true
		}
	}
	if !ok {
		t.Fatalf("no se guardó el hash de la nueva contraseña: %v", execArgs)
	}
	call(t, http.MethodPut, "/clientes?id=1001", `{"observaciones":"Nota"}`, u, http.StatusOK)

	bads := []string{
		``, `{`, `[]`,
		`{"nombre":null}`, `{"apellido":null}`, `{"correo":null}`, `{"telefono":null}`, `{"direccion":null}`, `{"password":null}`,
		`{"nombre":" "}`, `{"apellido":""}`, `{"correo":"x"}`, `{"correo":""}`, `{"telefono":" "}`, `{"password":""}`,
		`{"password":"` + strings.Repeat("x", 73) + `"}`,
		`{"nombre":5}`,
	}
	for _, body := range bads {
		call(t, http.MethodPut, "/clientes?id=1001", body, u, http.StatusBadRequest)
	}

	orig := generateFromPassword
	generateFromPassword = func([]byte, int) ([]byte, error) { return nil, errBoom }
	call(t, http.MethodPut, "/clientes?id=1001", `{"password":"x"}`, u, http.StatusInternalServerError)
	generateFromPassword = orig

	for _, tc := range []struct {
		err    string
		status int
		msg    string
	}{
		{`duplicate key value violates unique constraint "cliente_correo_key"`, http.StatusConflict, "correo"},
		{`duplicate key value violates unique constraint "cliente_telefono_key"`, http.StatusConflict, "teléfono"},
		{`falló`, http.StatusInternalServerError, ""},
	} {
		e := errors.New(tc.err)
		fakeExec = func(string, []driver.NamedValue) (driver.Result, error) { return nil, e }
		b := call(t, http.MethodPut, "/clientes?id=1001", `{"nombre":"Z"}`, u, tc.status)
		if !strings.Contains(b, tc.msg) {
			t.Fatalf("mensaje sin %q: %s", tc.msg, b)
		}
	}
}

func TestDelete(t *testing.T) {
	defer resetFake()
	d := func(c *ClienteController) { c.Delete() }

	for _, q := range []string{"", "?id=0", "?id=abc"} {
		call(t, http.MethodDelete, "/clientes"+q, "", d, http.StatusBadRequest)
	}
	call(t, http.MethodDelete, "/clientes?id=1001", "", d, http.StatusOK)

	fakeAffected = 0
	call(t, http.MethodDelete, "/clientes?id=1001", "", d, http.StatusNotFound)
	fakeAffected = 1

	fakeExec = func(string, []driver.NamedValue) (driver.Result, error) {
		return nil, errors.New(`violates foreign key constraint "pedido_cliente_fkey"`)
	}
	call(t, http.MethodDelete, "/clientes?id=1001", "", d, http.StatusConflict)

	fakeExec = func(string, []driver.NamedValue) (driver.Result, error) { return nil, errBoom }
	call(t, http.MethodDelete, "/clientes?id=1001", "", d, http.StatusInternalServerError)
}
