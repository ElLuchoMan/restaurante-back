package reserva

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	loginc "restaurante/controllers/login"
	"restaurante/models"

	"github.com/golang-jwt/jwt/v5"
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
	// consultas registra las lecturas (SELECT) con sus argumentos.
	consultas []execCall
)

// programa instala un fakeQuery que responde según las reglas (la primera
// que coincide) y registra todo el SQL ejecutado (consultas y comandos).
func programa(t *testing.T, reglas ...*res) {
	t.Helper()
	resetFake()
	recorded, execs, consultas = nil, nil, nil
	fakeQuery = func(q string, args []driver.NamedValue) (driver.Rows, error) {
		recorded = append(recorded, q)
		consultas = append(consultas, execCall{q: q, args: args})
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

var (
	dia   = time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)
	hora  = time.Date(2000, 1, 1, 18, 30, 0, 0, time.UTC)
	errDB = errors.New("fallo de base de datos")
)

const (
	qReservaRel   = `FROM "reserva" T0`
	qContacto     = `FROM "reserva_contacto" T0`
	qRestaurante  = `FROM "restaurante"`
	qCliente      = `FROM "cliente"`
	qInsertContac = `INSERT INTO "reserva_contacto"`
	qInsertReserv = `INSERT INTO "reserva"`
	qUpdate       = `UPDATE "reserva"`
)

// filaReserva devuelve una fila de la consulta con relaciones (21 columnas):
// reserva + contacto + restaurante. clienteDoc nil = invitado.
func filaReserva(id int64, estado any, clienteDoc any) []driver.Value {
	return []driver.Value{
		id, dia, hora, int64(4), int64(2), int64(1), estado, "ventana", dia, dia, nil, nil,
		int64(2), "Ana Gómez", "3001234567", int64(55), clienteDoc,
		int64(1), "Sazón Criolla", hora, nil,
	}
}

func reservaRows(filas ...[]driver.Value) *res {
	return &res{match: qReservaRel, cols: 21, rows: filas}
}

// filaContacto: fila de reserva_contacto (5 columnas).
func filaContacto(id int64, docContacto, docCliente any) []driver.Value {
	return []driver.Value{id, "Ana Gómez", "3001234567", docContacto, docCliente}
}

func contactoRows(filas ...[]driver.Value) *res {
	return &res{match: qContacto, cols: 5, rows: filas}
}

func restauranteRows() *res {
	return &res{match: qRestaurante, cols: 4, rows: [][]driver.Value{{int64(1), "Sazón Criolla", hora, nil}}}
}

// clienteRows: cliente con contraseña en la fila para comprobar que no se filtra.
func clienteRows() *res {
	return &res{match: qCliente, cols: 8, rows: [][]driver.Value{{int64(9), "Luis", "Mora", "l@x.co", "Calle 1", "3009998888", nil, "secreto-hash"}}}
}

func vacio(match string, cols int) *res { return &res{match: match, cols: cols} }

func conError(match string) *res { return &res{match: match, err: errDB} }

// run ejecuta un handler del controlador como personal (Mesero) autenticado.
func run(method, target, body string, h func(*ReservaController)) *httptest.ResponseRecorder {
	return runAs(tokenDe(rolMesero, 1), method, target, body, h)
}

// runAs ejecuta un handler con el Authorization dado ("" = invitado sin token).
func runAs(auth, method, target, body string, h func(*ReservaController)) *httptest.ResponseRecorder {
	ctx, w := newCtx(method, target, body)
	if auth != "" {
		ctx.Request.Header.Set("Authorization", auth)
	}
	c := &ReservaController{}
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}
	h(c)
	return w
}

const (
	rolMesero  = "Mesero"
	rolCliente = "Cliente"
)

// tokenDe devuelve un header Authorization Bearer firmado para rol y documento.
func tokenDe(rol string, documento int64) string {
	claims := loginc.Claims{
		Documento: documento, Rol: rol, Nombre: "Prueba",
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))},
	}
	tok, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(loginc.GetJWTSecret())
	if err != nil {
		panic(err)
	}
	return "Bearer " + tok
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

// argQue devuelve el primer argumento de la primera lectura cuyo SQL contiene sub.
func argQue(sub string) driver.Value {
	for _, c := range consultas {
		if strings.Contains(c.q, sub) && len(c.args) > 0 {
			return c.args[0].Value
		}
	}
	return nil
}
