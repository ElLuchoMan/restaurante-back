package cambioshorario

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

var (
	cambioCols = []string{"pk_id_cambio_horario", "fecha", "hora_apertura", "hora_cierre", "abierto"}
	errBoom    = errors.New("boom")
)

// lmt arma una hora como la entrega el driver (año 0 con desfase LMT de
// 9h52m32s): FormatTimeWithLMT la muestra como h:00:00.
func lmt(h int) time.Time {
	return time.Date(0, 1, 1, h, 0, 0, 0, time.UTC).Add(-(9*time.Hour + 52*time.Minute + 32*time.Second))
}

func fecha(y int, m time.Month, d int) time.Time { return time.Date(y, m, d, 12, 0, 0, 0, time.UTC) }

func rowAbierto() []driver.Value {
	return []driver.Value{int64(3), fecha(2025, 12, 24), lmt(9), lmt(20), true}
}

func rowCerrado() []driver.Value {
	return []driver.Value{int64(4), fecha(2025, 12, 25), lmt(0), lmt(23), false}
}

func serve(rows ...[]driver.Value) {
	fakeQuery = func(string, []driver.NamedValue) (driver.Rows, error) {
		return rowsOf(cambioCols, rows...), nil
	}
}

func failQuery() {
	fakeQuery = func(string, []driver.NamedValue) (driver.Rows, error) { return nil, errBoom }
}

func failExec(msg string) {
	e := errors.New(msg)
	fakeExec = func(string, []driver.NamedValue) (driver.Result, error) { return nil, e }
}

func call(t *testing.T, method, target, body string, f func(c *CambiosHorarioController), status int) string {
	t.Helper()
	ctx, w := newCtx(method, target, body)
	c := &CambiosHorarioController{}
	c.Ctx, c.Data = ctx, map[interface{}]interface{}{}
	f(c)
	expect(t, w, status)
	return w.Body.String()
}

func mustContain(t *testing.T, body string, wants ...string) {
	t.Helper()
	for _, w := range wants {
		if !strings.Contains(body, w) {
			t.Fatalf("falta %s en %s", w, body)
		}
	}
}

func TestGetAll(t *testing.T) {
	defer resetFake()
	g := func(c *CambiosHorarioController) { c.GetAll() }

	mustContain(t, call(t, http.MethodGet, "/cambios_horario", "", g, http.StatusOK), `"data":[]`)

	serve(rowAbierto(), rowCerrado())
	b := call(t, http.MethodGet, "/cambios_horario", "", g, http.StatusOK)
	mustContain(t, b,
		`{"cambioHorarioId":3,"fechaCambioHorario":"24-12-2025","horaApertura":"09:00:00","horaCierre":"20:00:00","abierto":true}`,
		`"cambioHorarioId":4`, `"abierto":false`)

	failQuery()
	call(t, http.MethodGet, "/cambios_horario", "", g, http.StatusInternalServerError)
}

func TestGetByCurrentDate(t *testing.T) {
	defer resetFake()
	defer func() { now = time.Now }()
	g := func(c *CambiosHorarioController) { c.GetByCurrentDate() }

	// 2025-12-25 02:00 UTC es todavía 24-dic a las 21:00 en Colombia.
	now = func() time.Time { return time.Date(2025, 12, 25, 2, 0, 0, 0, time.UTC) }
	var gotArgs []driver.NamedValue
	fakeQuery = func(_ string, a []driver.NamedValue) (driver.Rows, error) {
		gotArgs = a
		return rowsOf(cambioCols, rowAbierto()), nil
	}
	b := call(t, http.MethodGet, "/cambios_horario/actual", "", g, http.StatusOK)
	mustContain(t, b, `"fechaCambioHorario":"24-12-2025"`, `"horaApertura":"09:00:00"`)
	if got := fmt.Sprint(gotArgs[0].Value); !strings.HasPrefix(got, "2025-12-24") {
		t.Fatalf("debía consultar la fecha local de Colombia: %v", got)
	}

	serve()
	call(t, http.MethodGet, "/cambios_horario/actual", "", g, http.StatusNotFound)
	failQuery()
	call(t, http.MethodGet, "/cambios_horario/actual", "", g, http.StatusInternalServerError)
}

func TestPost(t *testing.T) {
	defer resetFake()
	p := func(c *CambiosHorarioController) { c.Post() }

	b := call(t, http.MethodPost, "/cambios_horario", `{"fechaCambioHorario":"2025-12-24","abierto":true,"horaApertura":"09:00","horaCierre":" 20:30:15 "}`, p, http.StatusCreated)
	mustContain(t, b, `"cambioHorarioId":7`, `"fechaCambioHorario":"24-12-2025"`, `"horaApertura":"09:00:00"`, `"horaCierre":"20:30:15"`, `"abierto":true`)

	// cerrado: se ignoran las horas enviadas y se fuerzan 00:00:00 - 23:59:59
	b = call(t, http.MethodPost, "/cambios_horario", `{"fechaCambioHorario":"2025-12-25","abierto":false,"horaApertura":"10:00","horaCierre":"11:00"}`, p, http.StatusCreated)
	mustContain(t, b, `"horaApertura":"00:00:00"`, `"horaCierre":"23:59:59"`, `"abierto":false`)

	bads := []string{
		``, `{`, `[]`, `{"fechaCambioHorario":5}`,
		`{"abierto":false}`,
		`{"fechaCambioHorario":" ","abierto":false}`,
		`{"fechaCambioHorario":"24-12-2025","abierto":false}`,
		`{"fechaCambioHorario":"2025-12-24"}`,
		`{"fechaCambioHorario":"2025-12-24","abierto":true,"horaCierre":"20:00"}`,
		`{"fechaCambioHorario":"2025-12-24","abierto":true,"horaApertura":" ","horaCierre":"20:00"}`,
		`{"fechaCambioHorario":"2025-12-24","abierto":true,"horaApertura":"09:00"}`,
		`{"fechaCambioHorario":"2025-12-24","abierto":true,"horaApertura":"09:00","horaCierre":""}`,
		`{"fechaCambioHorario":"2025-12-24","abierto":true,"horaApertura":"x","horaCierre":"20:00"}`,
		`{"fechaCambioHorario":"2025-12-24","abierto":true,"horaApertura":"09:00","horaCierre":"99:00"}`,
	}
	for _, body := range bads {
		call(t, http.MethodPost, "/cambios_horario", body, p, http.StatusBadRequest)
	}

	ok := `{"fechaCambioHorario":"2025-12-24","abierto":false}`
	failExec(`duplicate key value violates unique constraint "cambios_horario_fecha_key"`)
	call(t, http.MethodPost, "/cambios_horario", ok, p, http.StatusConflict)
	failExec(`conexión perdida`)
	call(t, http.MethodPost, "/cambios_horario", ok, p, http.StatusInternalServerError)
}

func TestPut(t *testing.T) {
	defer resetFake()
	u := func(c *CambiosHorarioController) { c.Put() }

	for _, q := range []string{"", "?id=0", "?id=abc"} {
		call(t, http.MethodPut, "/cambios_horario"+q, `{}`, u, http.StatusBadRequest)
	}
	call(t, http.MethodPut, "/cambios_horario?id=3", `{}`, u, http.StatusNotFound)
	failQuery()
	call(t, http.MethodPut, "/cambios_horario?id=3", `{}`, u, http.StatusInternalServerError)

	serve(rowAbierto())
	var args []driver.NamedValue
	fakeExec = func(_ string, a []driver.NamedValue) (driver.Result, error) {
		args = a
		return fakeResult{}, nil
	}

	// merge: lo ausente se conserva (horas de pared, sin desfase)
	b := call(t, http.MethodPut, "/cambios_horario?id=3", `{"horaCierre":"22:00"}`, u, http.StatusOK)
	mustContain(t, b, `"cambioHorarioId":3`, `"fechaCambioHorario":"24-12-2025"`, `"horaApertura":"09:00:00"`, `"horaCierre":"22:00:00"`, `"abierto":true`)
	joined := fmt.Sprint(args)
	if !strings.Contains(joined, "0001-01-01 09:00:00") || !strings.Contains(joined, "0001-01-01 22:00:00") {
		t.Fatalf("las horas guardadas deben ser de pared: %v", args)
	}
	mustContain(t, call(t, http.MethodPut, "/cambios_horario?id=3", `{}`, u, http.StatusOK), `"horaApertura":"09:00:00"`, `"horaCierre":"20:00:00"`)
	b = call(t, http.MethodPut, "/cambios_horario?id=3", `{"fechaCambioHorario":"2025-12-31","horaApertura":"10:15:30"}`, u, http.StatusOK)
	mustContain(t, b, `"fechaCambioHorario":"31-12-2025"`, `"horaApertura":"10:15:30"`, `"horaCierre":"20:00:00"`)

	// pasar a cerrado fuerza las horas
	b = call(t, http.MethodPut, "/cambios_horario?id=3", `{"abierto":false,"horaCierre":"22:00"}`, u, http.StatusOK)
	mustContain(t, b, `"horaApertura":"00:00:00"`, `"horaCierre":"23:59:59"`, `"abierto":false`)

	// un cambio cerrado sigue cerrado aunque lleguen horas
	serve(rowCerrado())
	b = call(t, http.MethodPut, "/cambios_horario?id=4", `{"horaApertura":"10:00"}`, u, http.StatusOK)
	mustContain(t, b, `"horaApertura":"00:00:00"`, `"abierto":false`)
	// reabrir exige ambas horas
	call(t, http.MethodPut, "/cambios_horario?id=4", `{"abierto":true}`, u, http.StatusBadRequest)
	call(t, http.MethodPut, "/cambios_horario?id=4", `{"abierto":true,"horaApertura":"09:00"}`, u, http.StatusBadRequest)
	call(t, http.MethodPut, "/cambios_horario?id=4", `{"abierto":true,"horaCierre":"09:00"}`, u, http.StatusBadRequest)
	b = call(t, http.MethodPut, "/cambios_horario?id=4", `{"abierto":true,"horaApertura":"09:00","horaCierre":"18:00"}`, u, http.StatusOK)
	mustContain(t, b, `"horaApertura":"09:00:00"`, `"horaCierre":"18:00:00"`, `"abierto":true`)

	// fila sin horaApertura (columna nula)
	serve([]driver.Value{int64(5), fecha(2025, 12, 26), nil, lmt(20), true})
	b = call(t, http.MethodPut, "/cambios_horario?id=5", `{"horaCierre":"21:00"}`, u, http.StatusOK)
	if strings.Contains(b, "horaApertura") {
		t.Fatalf("horaApertura nula debe omitirse: %s", b)
	}

	serve(rowAbierto())
	bads := []string{
		``, `{`, `[]`,
		`{"fechaCambioHorario":null}`, `{"abierto":null}`, `{"horaApertura":null}`, `{"horaCierre":null}`,
		`{"fechaCambioHorario":"x"}`, `{"fechaCambioHorario":""}`,
		`{"horaApertura":"x"}`, `{"horaApertura":""}`, `{"horaCierre":"99:99"}`, `{"horaCierre":""}`,
		`{"abierto":"si"}`,
	}
	for _, body := range bads {
		call(t, http.MethodPut, "/cambios_horario?id=3", body, u, http.StatusBadRequest)
	}

	failExec(`duplicate key value violates unique constraint "cambios_horario_fecha_key"`)
	call(t, http.MethodPut, "/cambios_horario?id=3", `{}`, u, http.StatusConflict)
	failExec(`falló`)
	call(t, http.MethodPut, "/cambios_horario?id=3", `{}`, u, http.StatusInternalServerError)
}

func TestDelete(t *testing.T) {
	defer resetFake()
	d := func(c *CambiosHorarioController) { c.Delete() }

	for _, q := range []string{"", "?id=0", "?id=abc"} {
		call(t, http.MethodDelete, "/cambios_horario"+q, "", d, http.StatusBadRequest)
	}
	call(t, http.MethodDelete, "/cambios_horario?id=3", "", d, http.StatusOK)
	fakeAffected = 0
	call(t, http.MethodDelete, "/cambios_horario?id=3", "", d, http.StatusNotFound)
	fakeAffected = 1

	failExec(`violates foreign key constraint "restaurante_cambio_horario_fkey"`)
	call(t, http.MethodDelete, "/cambios_horario?id=3", "", d, http.StatusConflict)
	failExec(`falló`)
	call(t, http.MethodDelete, "/cambios_horario?id=3", "", d, http.StatusInternalServerError)
}
