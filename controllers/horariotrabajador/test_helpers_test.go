package horariotrabajador

import (
	"database/sql/driver"
	"errors"
	"strings"
	"testing"
	"time"
)

var (
	horarioCols = []string{"pk_documento_trabajador", "dia", "hora_inicio", "hora_fin"}
	errBoom     = errors.New("boom")
)

// lmt arma una hora como la entrega el driver (año 0 con desfase LMT de
// 9h52m32s): FormatTimeWithLMT la muestra como h:00:00.
func lmt(h int) time.Time {
	return time.Date(0, 1, 1, h, 0, 0, 0, time.UTC).Add(-(9*time.Hour + 52*time.Minute + 32*time.Second))
}

func horarioRow(doc int64, dia string, ini, fin int) []driver.Value {
	return []driver.Value{doc, dia, lmt(ini), lmt(fin)}
}

// serve programa las filas devueltas por cualquier SELECT.
func serve(rows ...[]driver.Value) {
	fakeQuery = func(string, []driver.NamedValue) (driver.Rows, error) {
		return rowsOf(horarioCols, rows...), nil
	}
}

func failQuery() {
	fakeQuery = func(string, []driver.NamedValue) (driver.Rows, error) { return nil, errBoom }
}

func failExec(msg string) {
	e := errors.New(msg)
	fakeExec = func(string, []driver.NamedValue) (driver.Result, error) { return nil, e }
}

func call(t *testing.T, method, target, body string, f func(c *HorarioTrabajadorController), status int) string {
	t.Helper()
	ctx, w := newCtx(method, target, body)
	c := &HorarioTrabajadorController{}
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
