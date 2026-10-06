package horariotrabajador

import (
	"database/sql/driver"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestDiaToDB(t *testing.T) {
	for in, want := range map[string]string{"lunes": "Lunes", " MARTES ": "Martes", "miércoles": "Miércoles", "Sábado": "Sábado", "d": ""} {
		got, ok := diaToDB(in)
		if want == "" {
			if ok {
				t.Fatalf("%q no debía ser válido", in)
			}
			continue
		}
		if !ok || got != want {
			t.Fatalf("diaToDB(%q) = %q, %v", in, got, ok)
		}
	}
	if _, ok := diaToDB(" "); ok {
		t.Fatal("vacío no es un día")
	}
}

func TestGetAll(t *testing.T) {
	defer resetFake()
	g := func(c *HorarioTrabajadorController) { c.GetAll() }

	if b := call(t, http.MethodGet, "/horario_trabajador", "", g, http.StatusOK); !strings.Contains(b, `"data":[]`) {
		t.Fatalf("lista vacía debe ser []: %s", b)
	}

	var gotQ string
	var gotArgs []driver.NamedValue
	fakeQuery = func(q string, a []driver.NamedValue) (driver.Rows, error) {
		gotQ, gotArgs = q, a
		return rowsOf(horarioCols, horarioRow(10, "Lunes", 8, 16)), nil
	}
	b := call(t, http.MethodGet, "/horario_trabajador", "", g, http.StatusOK)
	mustContain(t, b, `"data":[{"documentoTrabajador":10,"dia":"Lunes","horaInicio":"08:00:00","horaFin":"16:00:00"}]`)
	if strings.Contains(gotQ, "WHERE") {
		t.Fatalf("sin filtros no debe haber WHERE: %s", gotQ)
	}

	call(t, http.MethodGet, "/horario_trabajador?documento=10&dia=lunes", "", g, http.StatusOK)
	if !strings.Contains(gotQ, "WHERE pk_documento_trabajador = $1 AND dia = $2") || len(gotArgs) != 2 || gotArgs[1].Value != "Lunes" {
		t.Fatalf("filtros no aplicados: %s %v", gotQ, gotArgs)
	}
	call(t, http.MethodGet, "/horario_trabajador?dia=Sábado", "", g, http.StatusOK)

	for _, q := range []string{"documento=0", "documento=abc", "documento=-1", "dia=Funday"} {
		call(t, http.MethodGet, "/horario_trabajador?"+q, "", g, http.StatusBadRequest)
	}

	failQuery()
	call(t, http.MethodGet, "/horario_trabajador", "", g, http.StatusInternalServerError)
}

func TestPost(t *testing.T) {
	defer resetFake()
	p := func(c *HorarioTrabajadorController) { c.Post() }

	var args []driver.NamedValue
	fakeExec = func(_ string, a []driver.NamedValue) (driver.Result, error) {
		args = a
		return fakeResult{}, nil
	}
	b := call(t, http.MethodPost, "/horario_trabajador", `{"documentoTrabajador":10,"dia":"lunes","horaInicio":"08:00","horaFin":" 16:30:15 "}`, p, http.StatusCreated)
	mustContain(t, b, `"documentoTrabajador":10`, `"dia":"Lunes"`, `"horaInicio":"08:00:00"`, `"horaFin":"16:30:15"`)
	if len(args) != 4 || args[1].Value != "Lunes" {
		t.Fatalf("argumentos del INSERT inesperados: %v", args)
	}

	bads := []string{
		``, `{`, `[]`, `{"documentoTrabajador":"x"}`,
		`{"dia":"Lunes","horaInicio":"08:00","horaFin":"16:00"}`,
		`{"documentoTrabajador":0,"dia":"Lunes","horaInicio":"08:00","horaFin":"16:00"}`,
		`{"documentoTrabajador":10,"horaInicio":"08:00","horaFin":"16:00"}`,
		`{"documentoTrabajador":10,"dia":"Funday","horaInicio":"08:00","horaFin":"16:00"}`,
		`{"documentoTrabajador":10,"dia":"Lunes","horaInicio":"x","horaFin":"16:00"}`,
		`{"documentoTrabajador":10,"dia":"Lunes","horaInicio":"08:00","horaFin":"25:00"}`,
		`{"documentoTrabajador":10,"dia":"Lunes","horaFin":"16:00"}`,
		`{"documentoTrabajador":10,"dia":"Lunes","horaInicio":"16:00","horaFin":"16:00"}`,
		`{"documentoTrabajador":10,"dia":"Lunes","horaInicio":"17:00","horaFin":"16:00"}`,
	}
	for _, body := range bads {
		call(t, http.MethodPost, "/horario_trabajador", body, p, http.StatusBadRequest)
	}

	ok := `{"documentoTrabajador":10,"dia":"Lunes","horaInicio":"08:00","horaFin":"16:00"}`
	failExec(`duplicate key value violates unique constraint "horario_trabajador_pk_documento_trabajador_dia_key"`)
	call(t, http.MethodPost, "/horario_trabajador", ok, p, http.StatusConflict)
	failExec(`violates foreign key constraint "horario_trabajador_fkey"`)
	call(t, http.MethodPost, "/horario_trabajador", ok, p, http.StatusBadRequest)
	failExec(`conexión perdida`)
	call(t, http.MethodPost, "/horario_trabajador", ok, p, http.StatusInternalServerError)
}

func TestPut(t *testing.T) {
	defer resetFake()
	u := func(c *HorarioTrabajadorController) { c.Put() }
	url := "/horario_trabajador?documento=10&dia=Lunes"

	for _, q := range []string{"", "?documento=0&dia=Lunes", "?documento=10", "?documento=10&dia=Funday"} {
		call(t, http.MethodPut, "/horario_trabajador"+q, `{}`, u, http.StatusBadRequest)
	}
	call(t, http.MethodPut, url, `{}`, u, http.StatusNotFound)
	failQuery()
	call(t, http.MethodPut, url, `{}`, u, http.StatusInternalServerError)

	serve(horarioRow(10, "Lunes", 8, 16))
	var args []driver.NamedValue
	fakeExec = func(_ string, a []driver.NamedValue) (driver.Result, error) {
		args = a
		return fakeResult{}, nil
	}

	// merge: la horaInicio ausente se conserva tal como está en la BD (08:00)
	b := call(t, http.MethodPut, url, `{"horaFin":"18:00"}`, u, http.StatusOK)
	mustContain(t, b, `"documentoTrabajador":10`, `"dia":"Lunes"`, `"horaInicio":"08:00:00"`, `"horaFin":"18:00:00"`)
	if got := fmt.Sprint(args[0].Value); !strings.HasPrefix(got, "0001-01-01 08:00:00") {
		t.Fatalf("horaInicio conservada incorrecta: %v", got)
	}
	// sin cambios (body vacío) conserva ambas
	b = call(t, http.MethodPut, url, `{}`, u, http.StatusOK)
	mustContain(t, b, `"horaInicio":"08:00:00"`, `"horaFin":"16:00:00"`)
	// ambos campos; el día se normaliza
	b = call(t, http.MethodPut, "/horario_trabajador?documento=10&dia=lunes", `{"horaInicio":"09:15","horaFin":"17:45:30"}`, u, http.StatusOK)
	mustContain(t, b, `"horaInicio":"09:15:00"`, `"horaFin":"17:45:30"`)

	bads := []string{
		``, `{`, `[]`,
		`{"horaInicio":null}`, `{"horaFin":null}`,
		`{"horaInicio":""}`, `{"horaInicio":"x"}`, `{"horaFin":""}`, `{"horaFin":"99:00"}`,
		`{"horaFin":"07:00"}`,    // anterior a la horaInicio conservada
		`{"horaInicio":"16:00"}`, // igual a la horaFin conservada
		`{"horaInicio":5}`,
	}
	for _, body := range bads {
		call(t, http.MethodPut, url, body, u, http.StatusBadRequest)
	}

	failExec(`falló`)
	call(t, http.MethodPut, url, `{}`, u, http.StatusInternalServerError)
}

func TestDelete(t *testing.T) {
	defer resetFake()
	d := func(c *HorarioTrabajadorController) { c.Delete() }
	url := "/horario_trabajador?documento=10&dia=Lunes"

	for _, q := range []string{"", "?documento=0&dia=Lunes", "?documento=10", "?documento=10&dia=Funday"} {
		call(t, http.MethodDelete, "/horario_trabajador"+q, "", d, http.StatusBadRequest)
	}
	call(t, http.MethodDelete, url, "", d, http.StatusOK)

	fakeAffected = 0
	call(t, http.MethodDelete, url, "", d, http.StatusNotFound)
	fakeAffected = 1

	failExec(`falló`)
	call(t, http.MethodDelete, url, "", d, http.StatusInternalServerError)
}
