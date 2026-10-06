package incidencia

import (
	"database/sql/driver"
	"errors"
	"strings"
	"testing"
	"time"
)

var (
	incidenciaCols = []string{"pk_id_incidencia", "fecha", "monto", "resta", "motivo", "pk_documento_trabajador"}
	errBoom        = errors.New("boom")
)

func fecha(y int, m time.Month, d int) time.Time { return time.Date(y, m, d, 12, 0, 0, 0, time.UTC) }

func incidenciaRow() []driver.Value {
	return []driver.Value{int64(5), fecha(2025, 1, 31), int64(50000), true, "Descuento por retraso", int64(10)}
}

func serve(rows ...[]driver.Value) {
	fakeQuery = func(string, []driver.NamedValue) (driver.Rows, error) {
		return rowsOf(incidenciaCols, rows...), nil
	}
}

func failQuery() {
	fakeQuery = func(string, []driver.NamedValue) (driver.Rows, error) { return nil, errBoom }
}

func failExec(msg string) {
	e := errors.New(msg)
	fakeExec = func(string, []driver.NamedValue) (driver.Result, error) { return nil, e }
}

func call(t *testing.T, method, target, body string, f func(c *IncidenciaController), status int) string {
	t.Helper()
	ctx, w := newCtx(method, target, body)
	c := &IncidenciaController{}
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
