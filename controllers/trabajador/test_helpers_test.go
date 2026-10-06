package trabajador

import (
	"database/sql/driver"
	"errors"
	"strings"
	"testing"
	"time"
)

var (
	trabajadorCols = []string{"pk_documento_trabajador", "nombre", "apellido", "sueldo", "telefono", "fecha_nacimiento", "nuevo", "rol", "fecha_ingreso", "fecha_retiro", "password", "pk_id_restaurante"}
	horarioCols    = []string{"pk_documento_trabajador", "dia", "hora_inicio", "hora_fin"}
	errBoom        = errors.New("boom")
)

func fecha(y int, m time.Month, d int) time.Time { return time.Date(y, m, d, 12, 0, 0, 0, time.UTC) }

// trabajadorRow es un trabajador activo (sin fecha de retiro).
func trabajadorRow() []driver.Value {
	return []driver.Value{int64(10), "María", "Gómez", int64(2000000), "3012223344", fecha(1990, 5, 20), false, "Mesero", fecha(2025, 1, 31), nil, "$2a$hash-secreto", int64(1)}
}

func retiradoRow() []driver.Value {
	r := trabajadorRow()
	r[9] = fecha(2025, 6, 30)
	return r
}

func horarioRow() []driver.Value {
	return []driver.Value{int64(10), "Lunes", lmt(8), lmt(16)}
}

// lmt arma una hora como la entrega el driver (año 0 con desfase LMT de
// 9h52m32s): FormatTimeWithLMT la muestra como h:00:00.
func lmt(h int) time.Time {
	return time.Date(0, 1, 1, h, 0, 0, 0, time.UTC).Add(-(9*time.Hour + 52*time.Minute + 32*time.Second))
}

// serve programa el driver con las filas de trabajador y horario dadas.
func serve(trab [][]driver.Value, horarios [][]driver.Value) {
	fakeQuery = func(q string, _ []driver.NamedValue) (driver.Rows, error) {
		if strings.Contains(q, "horario_trabajador") {
			return rowsOf(horarioCols, horarios...), nil
		}
		return rowsOf(trabajadorCols, trab...), nil
	}
}

// serveErr hace fallar la consulta de trabajador o la de horarios.
func serveErr(trab, horarios bool, ok [][]driver.Value) {
	fakeQuery = func(q string, _ []driver.NamedValue) (driver.Rows, error) {
		if strings.Contains(q, "horario_trabajador") {
			if horarios {
				return nil, errBoom
			}
			return rowsOf(horarioCols), nil
		}
		if trab {
			return nil, errBoom
		}
		return rowsOf(trabajadorCols, ok...), nil
	}
}

func call(t *testing.T, method, target, body string, f func(c *TrabajadorController), status int) string {
	t.Helper()
	ctx, w := newCtx(method, target, body)
	c := &TrabajadorController{}
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

func mustContain(t *testing.T, body string, wants ...string) {
	t.Helper()
	for _, w := range wants {
		if !strings.Contains(body, w) {
			t.Fatalf("falta %s en %s", w, body)
		}
	}
}
