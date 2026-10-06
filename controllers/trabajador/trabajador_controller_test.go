package trabajador

import (
	"database/sql/driver"
	"errors"
	"net/http"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestGetAll(t *testing.T) {
	defer resetFake()
	g := func(c *TrabajadorController) { c.GetAll() }

	b := call(t, http.MethodGet, "/trabajadores", "", g, http.StatusOK)
	mustContain(t, b, `"data":[]`)

	var gotQ string
	fakeQuery = func(q string, _ []driver.NamedValue) (driver.Rows, error) {
		if strings.Contains(q, "horario_trabajador") {
			return rowsOf(horarioCols, horarioRow()), nil
		}
		gotQ = q
		return rowsOf(trabajadorCols, trabajadorRow(), func() []driver.Value { r := trabajadorRow(); r[0] = int64(11); return r }()), nil
	}
	b = call(t, http.MethodGet, "/trabajadores", "", g, http.StatusOK)
	noPassword(t, b)
	mustContain(t, b, `"documentoTrabajador":10`, `"fechaIngreso":"31-01-2025"`, `"fechaNacimiento":"20-05-1990"`, `"restauranteId":{"restauranteId":1}`,
		`"horarios":[{"documentoTrabajador":10,"dia":"Lunes","horaInicio":"08:00:00","horaFin":"16:00:00"}]`,
		`"documentoTrabajador":11`, `"horarios":[]`)
	if !strings.Contains(gotQ, "IS NULL") {
		t.Fatalf("por defecto debe excluir retirados: %s", gotQ)
	}

	call(t, http.MethodGet, "/trabajadores?incluir_retirados=true&rol=Mesero&fecha_ingreso=2025-01-31", "", g, http.StatusOK)
	if strings.Contains(gotQ, "IS NULL") || !strings.Contains(gotQ, "rol") || !strings.Contains(gotQ, "fecha_ingreso") {
		t.Fatalf("filtros no aplicados: %s", gotQ)
	}
	call(t, http.MethodGet, "/trabajadores?solo_retirados=true&fechaIngreso=2025-01-31", "", g, http.StatusOK)
	if !strings.Contains(gotQ, "IS NOT NULL") {
		t.Fatalf("solo_retirados debe filtrar por retirados: %s", gotQ)
	}

	for _, q := range []string{"incluir_retirados=quiza", "solo_retirados=x", "rol=Gerente", "fecha_ingreso=31-01-2025"} {
		call(t, http.MethodGet, "/trabajadores?"+q, "", g, http.StatusBadRequest)
	}

	serveErr(true, false, nil)
	call(t, http.MethodGet, "/trabajadores", "", g, http.StatusInternalServerError)
	serveErr(false, true, [][]driver.Value{trabajadorRow()})
	call(t, http.MethodGet, "/trabajadores", "", g, http.StatusInternalServerError)
}

func TestGetById(t *testing.T) {
	defer resetFake()
	g := func(c *TrabajadorController) { c.GetById() }

	for _, q := range []string{"", "?id=0", "?id=abc"} {
		call(t, http.MethodGet, "/trabajadores/search"+q, "", g, http.StatusBadRequest)
	}
	call(t, http.MethodGet, "/trabajadores/search?id=10", "", g, http.StatusNotFound)

	serve([][]driver.Value{retiradoRow()}, nil)
	b := call(t, http.MethodGet, "/trabajadores/search?id=10", "", g, http.StatusOK)
	noPassword(t, b)
	mustContain(t, b, `"fechaRetiro":"30-06-2025"`, `"horarios":[]`)

	serve([][]driver.Value{trabajadorRow()}, [][]driver.Value{horarioRow()})
	b = call(t, http.MethodGet, "/trabajadores/search?id=10", "", g, http.StatusOK)
	mustContain(t, b, `"dia":"Lunes"`)

	serveErr(true, false, nil)
	call(t, http.MethodGet, "/trabajadores/search?id=10", "", g, http.StatusInternalServerError)
	serveErr(false, true, [][]driver.Value{trabajadorRow()})
	call(t, http.MethodGet, "/trabajadores/search?id=10", "", g, http.StatusInternalServerError)
}

const postOK = `{"documentoTrabajador":10,"nombre":" María ","apellido":"Gómez","rol":"Mesero","fechaIngreso":"2025-01-31","sueldo":2000000,"password":"Secreta123"}`

func TestPost(t *testing.T) {
	defer resetFake()
	p := func(c *TrabajadorController) { c.Post() }

	var execArgs []driver.NamedValue
	fakeExec = func(_ string, a []driver.NamedValue) (driver.Result, error) {
		execArgs = a
		return fakeResult{}, nil
	}
	b := call(t, http.MethodPost, "/trabajadores", postOK, p, http.StatusCreated)
	noPassword(t, b)
	mustContain(t, b, `"nombre":"María"`, `"rol":"Mesero"`, `"fechaIngreso":"31-01-2025"`, `"nuevo":false`, `"horarios":[]`)
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
		t.Fatalf("no se guardó el hash bcrypt")
	}

	full := `{"documentoTrabajador":11,"nombre":"A","apellido":"B","rol":"Administrador","fechaIngreso":"2025-01-31","sueldo":0,"password":"x","nuevo":true,"telefono":" 300 ","restauranteId":2,"fechaNacimiento":"1990-05-20"}`
	b = call(t, http.MethodPost, "/trabajadores", full, p, http.StatusCreated)
	mustContain(t, b, `"nuevo":true`, `"telefono":"300"`, `"restauranteId":{"restauranteId":2}`, `"fechaNacimiento":"20-05-1990"`, `"rol":"Administrador"`)
	// teléfono y fecha de nacimiento vacíos equivalen a no informados
	b = call(t, http.MethodPost, "/trabajadores", `{"documentoTrabajador":11,"nombre":"A","apellido":"B","rol":"Cocinero","fechaIngreso":"2025-01-31","sueldo":1,"password":"x","telefono":" ","fechaNacimiento":" "}`, p, http.StatusCreated)
	if strings.Contains(b, "telefono") || strings.Contains(b, "fechaNacimiento") {
		t.Fatalf("campos vacíos deben omitirse: %s", b)
	}

	base := func(extra string) string {
		return `{"documentoTrabajador":10,"nombre":"A","apellido":"B","rol":"Mesero","fechaIngreso":"2025-01-31","sueldo":1,"password":"x"` + extra + `}`
	}
	bads := []string{
		``, `{`, `[]`, `{"documentoTrabajador":"x"}`,
		`{"nombre":"A","apellido":"B","rol":"Mesero","fechaIngreso":"2025-01-31","sueldo":1,"password":"x"}`,
		`{"documentoTrabajador":0,"nombre":"A","apellido":"B","rol":"Mesero","fechaIngreso":"2025-01-31","sueldo":1,"password":"x"}`,
		`{"documentoTrabajador":10,"nombre":" ","apellido":"B","rol":"Mesero","fechaIngreso":"2025-01-31","sueldo":1,"password":"x"}`,
		`{"documentoTrabajador":10,"nombre":"A","apellido":"","rol":"Mesero","fechaIngreso":"2025-01-31","sueldo":1,"password":"x"}`,
		`{"documentoTrabajador":10,"nombre":"A","apellido":"B","fechaIngreso":"2025-01-31","sueldo":1,"password":"x"}`,
		`{"documentoTrabajador":10,"nombre":"A","apellido":"B","rol":"Gerente","fechaIngreso":"2025-01-31","sueldo":1,"password":"x"}`,
		`{"documentoTrabajador":10,"nombre":"A","apellido":"B","rol":"Mesero","sueldo":1,"password":"x"}`,
		`{"documentoTrabajador":10,"nombre":"A","apellido":"B","rol":"Mesero","fechaIngreso":"31-01-2025","sueldo":1,"password":"x"}`,
		`{"documentoTrabajador":10,"nombre":"A","apellido":"B","rol":"Mesero","fechaIngreso":"2025-01-31","password":"x"}`,
		`{"documentoTrabajador":10,"nombre":"A","apellido":"B","rol":"Mesero","fechaIngreso":"2025-01-31","sueldo":-1,"password":"x"}`,
		`{"documentoTrabajador":10,"nombre":"A","apellido":"B","rol":"Mesero","fechaIngreso":"2025-01-31","sueldo":1}`,
		`{"documentoTrabajador":10,"nombre":"A","apellido":"B","rol":"Mesero","fechaIngreso":"2025-01-31","sueldo":1,"password":"` + strings.Repeat("x", 73) + `"}`,
		base(`,"restauranteId":0`),
		base(`,"fechaNacimiento":"20/05/1990"`),
	}
	for _, body := range bads {
		call(t, http.MethodPost, "/trabajadores", body, p, http.StatusBadRequest)
	}

	orig := generateFromPassword
	generateFromPassword = func([]byte, int) ([]byte, error) { return nil, errBoom }
	call(t, http.MethodPost, "/trabajadores", postOK, p, http.StatusInternalServerError)
	generateFromPassword = orig

	for _, tc := range []struct {
		err    string
		status int
		msg    string
	}{
		{`duplicate key value violates unique constraint "trabajador_pkey"`, http.StatusConflict, "documento"},
		{`duplicate key value violates unique constraint "trabajador_telefono_key"`, http.StatusConflict, "teléfono"},
		{`violates foreign key constraint "trabajador_pk_id_restaurante_fkey"`, http.StatusBadRequest, "restauranteId"},
		{`conexión perdida`, http.StatusInternalServerError, "Error al crear"},
	} {
		e := errors.New(tc.err)
		fakeExec = func(string, []driver.NamedValue) (driver.Result, error) { return nil, e }
		b := call(t, http.MethodPost, "/trabajadores", postOK, p, tc.status)
		mustContain(t, b, tc.msg)
	}
}

func TestPut(t *testing.T) {
	defer resetFake()
	u := func(c *TrabajadorController) { c.Put() }

	for _, q := range []string{"", "?id=0", "?id=abc"} {
		call(t, http.MethodPut, "/trabajadores"+q, `{}`, u, http.StatusBadRequest)
	}
	call(t, http.MethodPut, "/trabajadores?id=10", `{}`, u, http.StatusNotFound)
	serveErr(true, false, nil)
	call(t, http.MethodPut, "/trabajadores?id=10", `{}`, u, http.StatusInternalServerError)
	serveErr(false, true, [][]driver.Value{trabajadorRow()})
	call(t, http.MethodPut, "/trabajadores?id=10", `{}`, u, http.StatusInternalServerError)

	serve([][]driver.Value{trabajadorRow()}, [][]driver.Value{horarioRow()})
	var execArgs []driver.NamedValue
	fakeExec = func(_ string, a []driver.NamedValue) (driver.Result, error) {
		execArgs = a
		return fakeResult{}, nil
	}

	// merge: los ausentes se conservan
	b := call(t, http.MethodPut, "/trabajadores?id=10", `{"nombre":" Ana "}`, u, http.StatusOK)
	noPassword(t, b)
	mustContain(t, b, `"nombre":"Ana"`, `"apellido":"Gómez"`, `"sueldo":2000000`, `"telefono":"3012223344"`, `"rol":"Mesero"`,
		`"fechaNacimiento":"20-05-1990"`, `"restauranteId":{"restauranteId":1}`, `"dia":"Lunes"`)
	kept := false
	for _, a := range execArgs {
		if a.Value == "$2a$hash-secreto" {
			kept = true
		}
	}
	if !kept {
		t.Fatalf("la contraseña original debía conservarse: %v", execArgs)
	}

	all := `{"apellido":"Z","rol":"Cocinero","sueldo":5,"nuevo":true,"telefono":" 311 ","fechaIngreso":"2024-01-01","fechaRetiro":"2025-02-02","fechaNacimiento":"1991-01-01","restauranteId":3,"password":"Nueva123"}`
	b = call(t, http.MethodPut, "/trabajadores?id=10", all, u, http.StatusOK)
	noPassword(t, b)
	mustContain(t, b, `"apellido":"Z"`, `"rol":"Cocinero"`, `"nuevo":true`, `"telefono":"311"`, `"fechaIngreso":"01-01-2024"`,
		`"fechaRetiro":"02-02-2025"`, `"fechaNacimiento":"01-01-1991"`, `"restauranteId":{"restauranteId":3}`)
	ok := false
	for _, a := range execArgs {
		if s, isStr := a.Value.(string); isStr && bcrypt.CompareHashAndPassword([]byte(s), []byte("Nueva123")) == nil {
			ok = true
		}
	}
	if !ok {
		t.Fatalf("no se guardó el hash de la nueva contraseña")
	}

	// null limpia los campos anulables
	b = call(t, http.MethodPut, "/trabajadores?id=10", `{"telefono":null,"fechaNacimiento":null,"fechaRetiro":null,"restauranteId":null}`, u, http.StatusOK)
	for _, no := range []string{"telefono", "fechaNacimiento", "fechaRetiro", "restauranteId"} {
		if strings.Contains(b, `"`+no+`"`) {
			t.Fatalf("%s debía limpiarse: %s", no, b)
		}
	}
	// teléfono vacío equivale a null
	b = call(t, http.MethodPut, "/trabajadores?id=10", `{"telefono":" "}`, u, http.StatusOK)
	if strings.Contains(b, `"telefono"`) {
		t.Fatalf("teléfono vacío debía limpiarse: %s", b)
	}

	bads := []string{
		``, `{`, `[]`,
		`{"nombre":null}`, `{"apellido":null}`, `{"rol":null}`, `{"sueldo":null}`, `{"nuevo":null}`, `{"fechaIngreso":null}`, `{"password":null}`,
		`{"nombre":" "}`, `{"apellido":""}`, `{"rol":"Gerente"}`, `{"rol":""}`, `{"sueldo":-1}`,
		`{"fechaIngreso":"x"}`, `{"fechaRetiro":"x"}`, `{"fechaNacimiento":"x"}`,
		`{"restauranteId":0}`, `{"password":""}`, `{"password":"` + strings.Repeat("x", 73) + `"}`,
		`{"fechaRetiro":"2024-12-31"}`, // anterior al ingreso (2025-01-31)
		`{"nombre":5}`,
	}
	for _, body := range bads {
		call(t, http.MethodPut, "/trabajadores?id=10", body, u, http.StatusBadRequest)
	}

	orig := generateFromPassword
	generateFromPassword = func([]byte, int) ([]byte, error) { return nil, errBoom }
	call(t, http.MethodPut, "/trabajadores?id=10", `{"password":"x"}`, u, http.StatusInternalServerError)
	generateFromPassword = orig

	for _, tc := range []struct {
		err    string
		status int
	}{
		{`duplicate key value violates unique constraint "trabajador_telefono_key"`, http.StatusConflict},
		{`violates foreign key constraint "fk"`, http.StatusBadRequest},
		{`falló`, http.StatusInternalServerError},
	} {
		e := errors.New(tc.err)
		fakeExec = func(string, []driver.NamedValue) (driver.Result, error) { return nil, e }
		call(t, http.MethodPut, "/trabajadores?id=10", `{"nombre":"Z"}`, u, tc.status)
	}
}

func TestDelete(t *testing.T) {
	defer resetFake()
	d := func(c *TrabajadorController) { c.Delete() }

	for _, q := range []string{"", "?id=0", "?id=abc"} {
		call(t, http.MethodDelete, "/trabajadores"+q, "", d, http.StatusBadRequest)
	}
	call(t, http.MethodDelete, "/trabajadores?id=10", "", d, http.StatusNotFound)
	serveErr(true, false, nil)
	call(t, http.MethodDelete, "/trabajadores?id=10", "", d, http.StatusInternalServerError)

	serve([][]driver.Value{retiradoRow()}, nil)
	call(t, http.MethodDelete, "/trabajadores?id=10", "", d, http.StatusConflict)

	serveErr(false, true, [][]driver.Value{trabajadorRow()})
	call(t, http.MethodDelete, "/trabajadores?id=10", "", d, http.StatusInternalServerError)

	serve([][]driver.Value{trabajadorRow()}, [][]driver.Value{horarioRow()})
	b := call(t, http.MethodDelete, "/trabajadores?id=10", "", d, http.StatusOK)
	noPassword(t, b)
	mustContain(t, b, `"fechaRetiro":"`, `"dia":"Lunes"`)

	fakeExec = func(string, []driver.NamedValue) (driver.Result, error) { return nil, errBoom }
	call(t, http.MethodDelete, "/trabajadores?id=10", "", d, http.StatusInternalServerError)
}
