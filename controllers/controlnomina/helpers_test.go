package controlnomina

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"restaurante/models"
)

// res es el resultado programado para las consultas cuyo SQL contiene match.
type res struct {
	match string
	cols  int
	rows  [][]driver.Value
	err   error
	// onlyCall (1-based) limita la regla a la N-ésima consulta que coincide.
	onlyCall int
	seen     int
}

type execCall struct {
	q    string
	args []driver.NamedValue
}

var (
	recorded []string
	// execs registra los comandos (INSERT/UPDATE) con sus argumentos.
	execs []execCall
)

// programa instala un fakeQuery que responde según las reglas (la primera
// que coincide) y registra todo el SQL ejecutado (consultas y comandos).
func programa(t *testing.T, reglas ...*res) {
	t.Helper()
	resetFake()
	recorded, execs = nil, nil
	fakeQuery = func(q string, args []driver.NamedValue) (driver.Rows, error) {
		recorded = append(recorded, q)
		for _, r := range reglas {
			if !strings.Contains(q, r.match) {
				continue
			}
			r.seen++
			if r.onlyCall != 0 && r.onlyCall != r.seen {
				continue
			}
			if r.err != nil {
				return nil, r.err
			}
			return rowsOf(colNames(r.cols), r.rows...), nil
		}
		return rowsOf(colNames(1)), nil
	}
	fakeExec = func(q string, args []driver.NamedValue) (driver.Result, error) {
		recorded = append(recorded, q)
		execs = append(execs, execCall{q: q, args: args})
		for _, r := range reglas {
			if r.err != nil && strings.Contains(q, r.match) {
				return nil, r.err
			}
		}
		return fakeResult{}, nil
	}
	t.Cleanup(resetFake)
}

func colNames(n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = fmt.Sprintf("c%d", i)
	}
	return out
}

func sqlEjecutado(sub string) bool { return sqlQue(sub) != "" }

func sqlQue(sub string) string {
	for _, q := range recorded {
		if strings.Contains(q, sub) {
			return q
		}
	}
	return ""
}

func execQue(sub string) *execCall {
	for i := range execs {
		if strings.Contains(execs[i].q, sub) {
			return &execs[i]
		}
	}
	return nil
}

var errDB = errors.New("fallo de base de datos")

func vacio(match string, cols int) *res { return &res{match: match, cols: cols} }

func conError(match string) *res { return &res{match: match, err: errDB} }

// run ejecuta un handler del controlador con una petición simulada.
func run(method, target, body string, h func(*ControlNominaController)) *httptest.ResponseRecorder {
	ctx, w := newCtx(method, target, body)
	c := &ControlNominaController{}
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}
	h(c)
	return w
}

func dataMap(t *testing.T, r models.ApiResponse) map[string]any {
	t.Helper()
	m, ok := r.Data.(map[string]any)
	if !ok {
		t.Fatalf("data no es un objeto: %#v", r.Data)
	}
	return m
}

func dataList(t *testing.T, r models.ApiResponse) []any {
	t.Helper()
	l, ok := r.Data.([]any)
	if !ok {
		t.Fatalf("data no es una lista: %#v", r.Data)
	}
	return l
}

// sinPassword falla si el cuerpo menciona contraseñas.
func sinPassword(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()
	b := strings.ToLower(w.Body.String())
	if strings.Contains(b, "password") || strings.Contains(b, "secreto") {
		t.Fatalf("la respuesta filtra la contraseña: %s", w.Body.String())
	}
}
