package incidencia

import (
	"database/sql/driver"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

const wantJSON = `{"incidenciaId":5,"fechaIncidencia":"31-01-2025","monto":50000,"resta":true,"motivo":"Descuento por retraso","documentoTrabajador":10}`

func TestGetAll(t *testing.T) {
	defer resetFake()
	g := func(c *IncidenciaController) { c.GetAll() }

	mustContain(t, call(t, http.MethodGet, "/incidencias", "", g, http.StatusOK), `"data":[]`)

	serve(incidenciaRow())
	mustContain(t, call(t, http.MethodGet, "/incidencias", "", g, http.StatusOK), `"data":[`+wantJSON+`]`)

	failQuery()
	call(t, http.MethodGet, "/incidencias", "", g, http.StatusInternalServerError)
}

func TestGetByDocumentAndDate(t *testing.T) {
	defer resetFake()
	g := func(c *IncidenciaController) { c.GetByDocumentAndDate() }

	// sin resultados: 200 con lista vacía (no 200 con code 404)
	mustContain(t, call(t, http.MethodGet, "/incidencias/search?documento=10&mes=1&anio=2025", "", g, http.StatusOK), `"data":[]`)

	var args []driver.NamedValue
	fakeQuery = func(_ string, a []driver.NamedValue) (driver.Rows, error) {
		args = a
		return rowsOf(incidenciaCols, incidenciaRow()), nil
	}
	mustContain(t, call(t, http.MethodGet, "/incidencias/search?documento=10&mes=2&anio=2025", "", g, http.StatusOK), `"data":[`+wantJSON+`]`)
	if len(args) != 3 || !strings.HasPrefix(fmt.Sprint(args[1].Value), "2025-02-01") || !strings.HasPrefix(fmt.Sprint(args[2].Value), "2025-02-28") {
		t.Fatalf("rango del mes incorrecto: %v", args)
	}
	// diciembre cierra el 31
	call(t, http.MethodGet, "/incidencias/search?documento=10&mes=12&anio=2025", "", g, http.StatusOK)
	if !strings.HasPrefix(fmt.Sprint(args[2].Value), "2025-12-31") {
		t.Fatalf("rango de diciembre incorrecto: %v", args)
	}

	year := time.Now().Year()
	for _, q := range []string{
		"", "?mes=1&anio=2025", "?documento=0&mes=1&anio=2025", "?documento=x&mes=1&anio=2025",
		"?documento=10&anio=2025", "?documento=10&mes=0&anio=2025", "?documento=10&mes=13&anio=2025", "?documento=10&mes=x&anio=2025",
		"?documento=10&mes=1", "?documento=10&mes=1&anio=1899", "?documento=10&mes=1&anio=x",
		fmt.Sprintf("?documento=10&mes=1&anio=%d", year+1),
	} {
		call(t, http.MethodGet, "/incidencias/search"+q, "", g, http.StatusBadRequest)
	}
	call(t, http.MethodGet, fmt.Sprintf("/incidencias/search?documento=10&mes=1&anio=%d", year), "", g, http.StatusOK)

	failQuery()
	call(t, http.MethodGet, "/incidencias/search?documento=10&mes=1&anio=2025", "", g, http.StatusInternalServerError)
}

const postOK = `{"documentoTrabajador":10,"fechaIncidencia":"2025-01-31","monto":50000,"resta":true,"motivo":" Descuento por retraso "}`

func TestPost(t *testing.T) {
	defer resetFake()
	p := func(c *IncidenciaController) { c.Post() }

	b := call(t, http.MethodPost, "/incidencias", postOK, p, http.StatusCreated)
	mustContain(t, b, `"incidenciaId":7`, `"fechaIncidencia":"31-01-2025"`, `"monto":50000`, `"resta":true`, `"motivo":"Descuento por retraso"`, `"documentoTrabajador":10`)
	// monto 0 y resta=false son válidos si están presentes
	call(t, http.MethodPost, "/incidencias", `{"documentoTrabajador":10,"fechaIncidencia":"2025-01-31","monto":0,"resta":false,"motivo":"Bono"}`, p, http.StatusCreated)

	bads := []string{
		``, `{`, `[]`, `{"monto":"x"}`,
		`{"documentoTrabajador":10,"monto":1,"resta":true,"motivo":"m"}`,
		`{"documentoTrabajador":10,"fechaIncidencia":" ","monto":1,"resta":true,"motivo":"m"}`,
		`{"documentoTrabajador":10,"fechaIncidencia":"31-01-2025","monto":1,"resta":true,"motivo":"m"}`,
		`{"documentoTrabajador":10,"fechaIncidencia":"2025-01-31","resta":true,"motivo":"m"}`,
		`{"documentoTrabajador":10,"fechaIncidencia":"2025-01-31","monto":-5,"resta":true,"motivo":"m"}`,
		`{"documentoTrabajador":10,"fechaIncidencia":"2025-01-31","monto":1,"motivo":"m"}`,
		`{"documentoTrabajador":10,"fechaIncidencia":"2025-01-31","monto":1,"resta":true}`,
		`{"documentoTrabajador":10,"fechaIncidencia":"2025-01-31","monto":1,"resta":true,"motivo":"  "}`,
		`{"fechaIncidencia":"2025-01-31","monto":1,"resta":true,"motivo":"m"}`,
		`{"documentoTrabajador":0,"fechaIncidencia":"2025-01-31","monto":1,"resta":true,"motivo":"m"}`,
	}
	for _, body := range bads {
		call(t, http.MethodPost, "/incidencias", body, p, http.StatusBadRequest)
	}

	failExec(`violates foreign key constraint "incidencia_trabajador_fkey"`)
	call(t, http.MethodPost, "/incidencias", postOK, p, http.StatusBadRequest)
	failExec(`conexión perdida`)
	call(t, http.MethodPost, "/incidencias", postOK, p, http.StatusInternalServerError)
}

func TestPut(t *testing.T) {
	defer resetFake()
	u := func(c *IncidenciaController) { c.Put() }

	for _, q := range []string{"", "?id=0", "?id=abc"} {
		call(t, http.MethodPut, "/incidencias"+q, `{}`, u, http.StatusBadRequest)
	}
	call(t, http.MethodPut, "/incidencias?id=5", `{}`, u, http.StatusNotFound)
	failQuery()
	call(t, http.MethodPut, "/incidencias?id=5", `{}`, u, http.StatusInternalServerError)

	serve(incidenciaRow())
	// merge: sin cambios se conserva todo
	b := call(t, http.MethodPut, "/incidencias?id=5", `{}`, u, http.StatusOK)
	mustContain(t, b, wantJSON)
	b = call(t, http.MethodPut, "/incidencias?id=5", `{"monto":1000}`, u, http.StatusOK)
	mustContain(t, b, `"monto":1000`, `"motivo":"Descuento por retraso"`, `"resta":true`, `"documentoTrabajador":10`, `"fechaIncidencia":"31-01-2025"`)
	b = call(t, http.MethodPut, "/incidencias?id=5", `{"documentoTrabajador":11,"fechaIncidencia":"2025-02-01","monto":0,"resta":false,"motivo":" Bono "}`, u, http.StatusOK)
	mustContain(t, b, `"documentoTrabajador":11`, `"fechaIncidencia":"01-02-2025"`, `"monto":0`, `"resta":false`, `"motivo":"Bono"`)

	bads := []string{
		``, `{`, `[]`,
		`{"documentoTrabajador":null}`, `{"fechaIncidencia":null}`, `{"monto":null}`, `{"resta":null}`, `{"motivo":null}`,
		`{"fechaIncidencia":"x"}`, `{"fechaIncidencia":""}`, `{"monto":-1}`, `{"motivo":" "}`, `{"documentoTrabajador":0}`,
		`{"monto":"x"}`,
	}
	for _, body := range bads {
		call(t, http.MethodPut, "/incidencias?id=5", body, u, http.StatusBadRequest)
	}

	failExec(`violates foreign key constraint "incidencia_trabajador_fkey"`)
	call(t, http.MethodPut, "/incidencias?id=5", `{"documentoTrabajador":99}`, u, http.StatusBadRequest)
	failExec(`falló`)
	call(t, http.MethodPut, "/incidencias?id=5", `{}`, u, http.StatusInternalServerError)
}

func TestDelete(t *testing.T) {
	defer resetFake()
	d := func(c *IncidenciaController) { c.Delete() }

	for _, q := range []string{"", "?id=0", "?id=abc"} {
		call(t, http.MethodDelete, "/incidencias"+q, "", d, http.StatusBadRequest)
	}
	call(t, http.MethodDelete, "/incidencias?id=5", "", d, http.StatusOK)

	// id inexistente: 404 real
	fakeAffected = 0
	call(t, http.MethodDelete, "/incidencias?id=999", "", d, http.StatusNotFound)
	fakeAffected = 1

	failExec(`falló`)
	call(t, http.MethodDelete, "/incidencias?id=5", "", d, http.StatusInternalServerError)
}
