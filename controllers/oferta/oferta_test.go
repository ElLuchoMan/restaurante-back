package oferta

import (
	"database/sql/driver"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"
)

func day(y int, m time.Month, d int) time.Time { return time.Date(y, m, d, 0, 0, 0, 0, time.UTC) }

func ofertaVals() map[string]driver.Value {
	return map[string]driver.Value{
		"pk_id_oferta":          int64(1),
		"titulo":                "Martes",
		"tipo_descuento":        "PORCENTAJE",
		"valor_descuento":       int64(30),
		"fecha_inicio":          day(2025, 1, 1),
		"fecha_fin":             day(2025, 12, 31),
		"dias_semana":           "{Martes,\"Miércoles\"}",
		"hora_inicio":           time.Date(2000, 1, 1, 8, 0, 0, 0, time.UTC),
		"hora_fin":              time.Date(2000, 1, 1, 18, 0, 0, 0, time.UTC),
		"activo":                true,
		"T0.pk_id_restaurante":  int64(1),
		"T1.pk_id_restaurante":  int64(1),
		"T1.nombre_restaurante": "Sede",
		"T1.hora_apertura":      time.Date(2000, 1, 1, 8, 0, 0, 0, time.UTC),
	}
}

type env struct {
	vals    map[string]driver.Value
	execs   []string
	args    map[string][]driver.NamedValue
	queries []string
}

// setup programa una oferta existente (activa) y registra las escrituras.
func setup(mod func(v map[string]driver.Value)) *env {
	e := &env{vals: ofertaVals(), args: map[string][]driver.NamedValue{}}
	if mod != nil {
		mod(e.vals)
	}
	one := func(q string) (driver.Rows, error) { return rowFor(q, e.vals), nil }
	fakeQuery = queryFn(map[string]func(string) (driver.Rows, error){
		`FROM "oferta"`: one,
		`FROM "producto"`: func(q string) (driver.Rows, error) {
			return rowFor(q, map[string]driver.Value{"pk_id_producto": int64(2)}), nil
		},
		`FROM "oferta_producto"`: func(q string) (driver.Rows, error) {
			return rowFor(q, map[string]driver.Value{"pk_id_oferta": int64(1), "pk_id_producto": int64(2)}), nil
		},
	})
	fakeExec = func(q string, a []driver.NamedValue) (driver.Result, error) {
		e.execs = append(e.execs, q)
		e.args[q] = a
		return fakeResult{}, nil
	}
	return e
}

func (e *env) has(sub string) bool {
	for _, s := range e.execs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

func (e *env) argsOf(sub string) []driver.NamedValue {
	for q, a := range e.args {
		if strings.Contains(q, sub) {
			return a
		}
	}
	return nil
}

func execErrOn(sub, msg string) {
	prev := fakeExec
	fakeExec = func(q string, a []driver.NamedValue) (driver.Result, error) {
		if strings.Contains(q, sub) {
			return nil, errors.New(msg)
		}
		return prev(q, a)
	}
}

func queryErrOn(sub string) {
	prev := fakeQuery
	fakeQuery = func(q string, a []driver.NamedValue) (driver.Rows, error) {
		if strings.Contains(q, sub) {
			return nil, errors.New("boom")
		}
		return prev(q, a)
	}
}

func call(t *testing.T, method, target, body string, f func(c *OfertaController), status int) string {
	t.Helper()
	ctx, w := newCtx(method, target, body)
	c := &OfertaController{}
	c.Ctx, c.Data = ctx, map[interface{}]interface{}{}
	f(c)
	expect(t, w, status)
	return w.Body.String()
}

var (
	getAll   = func(c *OfertaController) { c.GetAll() }
	getByID  = func(c *OfertaController) { c.GetById() }
	post     = func(c *OfertaController) { c.Post() }
	put      = func(c *OfertaController) { c.Put() }
	del      = func(c *OfertaController) { c.Delete() }
	activas  = func(c *OfertaController) { c.ObtenerOfertasActivas() }
	asociar  = func(c *OfertaController) { c.AsociarProducto() }
	desasoc  = func(c *OfertaController) { c.DesasociarProducto() }
	contains = func(t *testing.T, b string, parts ...string) {
		t.Helper()
		for _, p := range parts {
			if !strings.Contains(b, p) {
				t.Fatalf("falta %s en %s", p, b)
			}
		}
	}
)

func TestGetAll(t *testing.T) {
	defer resetFake()
	fakeQuery = queryFn(map[string]func(string) (driver.Rows, error){
		"COUNT(": func(q string) (driver.Rows, error) { return rowFor(q, map[string]driver.Value{"count": int64(0)}), nil },
	})
	if b := call(t, http.MethodGet, "/ofertas", "", getAll, http.StatusOK); !strings.Contains(b, `"data":[]`) || !strings.Contains(b, `"totalPages":0`) {
		t.Fatalf("lista vacía debe ser []: %s", b)
	}
	setup(func(v map[string]driver.Value) { v["count"] = int64(1) })
	b := call(t, http.MethodGet, "/ofertas?activo=true&restaurante_id=1&titulo=mar&limit=500&offset=0", "", getAll, http.StatusOK)
	contains(t, b, `"diasSemana":["Martes","Miércoles"]`, `"nombreRestaurante":"Sede"`, `"pageSize":100`, `"page":1`)
	for _, q := range []string{"?activo=quizas", "?restaurante_id=abc", "?restaurante_id=0", "?limit=0", "?limit=abc", "?offset=-1", "?offset=abc"} {
		call(t, http.MethodGet, "/ofertas"+q, "", getAll, http.StatusBadRequest)
	}
	queryErrOn("COUNT(")
	call(t, http.MethodGet, "/ofertas", "", getAll, http.StatusInternalServerError)
	setup(nil)
	fakeQuery = queryFn(map[string]func(string) (driver.Rows, error){
		"COUNT(": func(q string) (driver.Rows, error) { return rowFor(q, map[string]driver.Value{"count": int64(0)}), nil },
		"SELECT": func(string) (driver.Rows, error) { return nil, errors.New("boom") },
	})
	call(t, http.MethodGet, "/ofertas", "", getAll, http.StatusInternalServerError)
}

func TestGetById(t *testing.T) {
	defer resetFake()
	for _, q := range []string{"", "?id=abc", "?id=0"} {
		call(t, http.MethodGet, "/ofertas/search"+q, "", getByID, http.StatusBadRequest)
	}
	call(t, http.MethodGet, "/ofertas/search?id=1", "", getByID, http.StatusNotFound)
	setup(nil)
	contains(t, call(t, http.MethodGet, "/ofertas/search?id=1", "", getByID, http.StatusOK), `"ofertaId":1`, `"restauranteId":{"restauranteId":1,"nombreRestaurante":"Sede"`)
	queryErrOn(`FROM "oferta"`)
	call(t, http.MethodGet, "/ofertas/search?id=1", "", getByID, http.StatusInternalServerError)
}

const okPost = `{"titulo":" Martes de gaseosas ","tipoDescuento":"PORCENTAJE","valorDescuento":30,"fechaInicio":"2025-01-01","fechaFin":"2025-12-31","diasSemana":["Martes"],"horaInicio":"08:00","horaFin":"18:00:00","restauranteId":1}`

func TestPost(t *testing.T) {
	defer resetFake()
	call(t, http.MethodPost, "/ofertas", "", post, http.StatusBadRequest)
	call(t, http.MethodPost, "/ofertas", "nojson", post, http.StatusBadRequest)
	mut := func(old, new string) string { return strings.Replace(okPost, old, new, 1) }
	for _, b := range []string{
		mut(`"PORCENTAJE"`, `"OTRO"`),
		mut(`"2025-01-01"`, `"01/01/2025"`),
		mut(`"2025-12-31"`, `"nada"`),
		mut(`"08:00"`, `"8am"`),
		mut(`"18:00:00"`, `"6pm"`),
		mut(`" Martes de gaseosas "`, `"  "`),
		mut(`"restauranteId":1`, `"restauranteId":0`),
		mut(`"valorDescuento":30`, `"valorDescuento":150`),
		mut(`"fechaFin":"2025-12-31"`, `"fechaFin":"2024-12-31"`),
		mut(`"diasSemana":["Martes"]`, `"diasSemana":["Funday"]`),
		mut(`"horaFin":"18:00:00",`, ``),
	} {
		call(t, http.MethodPost, "/ofertas", b, post, http.StatusUnprocessableEntity)
	}

	e := setup(nil)
	b := call(t, http.MethodPost, "/ofertas", okPost, post, http.StatusCreated)
	contains(t, b, `"ofertaId":1`, `"nombreRestaurante":"Sede"`)
	a := e.argsOf(`INSERT INTO "oferta"`)
	found := false
	for _, v := range a {
		if v.Value == `{"Martes"}` {
			found = true
		}
	}
	if !found {
		t.Fatalf("los días deben guardarse como arreglo PG: %v", a)
	}
	// Sin días ni horario.
	call(t, http.MethodPost, "/ofertas", `{"titulo":"Todo","tipoDescuento":"MONTO","valorDescuento":0,"fechaInicio":"2025-01-01","fechaFin":"2025-01-01","restauranteId":1}`, post, http.StatusCreated)

	execErrOn("INSERT INTO", "duplicate key value (23505)")
	call(t, http.MethodPost, "/ofertas", okPost, post, http.StatusConflict)
	resetFake()
	setup(nil)
	execErrOn("INSERT INTO", "violates foreign key constraint (23503)")
	call(t, http.MethodPost, "/ofertas", okPost, post, http.StatusBadRequest)
	resetFake()
	setup(nil)
	execErrOn("INSERT INTO", "boom")
	call(t, http.MethodPost, "/ofertas", okPost, post, http.StatusInternalServerError)
	resetFake()
	fakeQuery = queryFn(map[string]func(string) (driver.Rows, error){`FROM "oferta"`: func(string) (driver.Rows, error) { return nil, errors.New("boom") }})
	call(t, http.MethodPost, "/ofertas", okPost, post, http.StatusInternalServerError)
}

func TestPut(t *testing.T) {
	defer resetFake()
	call(t, http.MethodPut, "/ofertas", `{}`, put, http.StatusBadRequest)
	call(t, http.MethodPut, "/ofertas?id=1", `{}`, put, http.StatusNotFound)
	fakeQuery = queryFn(map[string]func(string) (driver.Rows, error){`FROM "oferta"`: func(string) (driver.Rows, error) { return nil, errors.New("boom") }})
	call(t, http.MethodPut, "/ofertas?id=1", `{}`, put, http.StatusInternalServerError)

	e := setup(nil)
	for _, b := range []string{"", "nojson", `{"titulo":null}`, `{"tipoDescuento":null}`, `{"diasSemana":null}`, `{"restauranteId":null}`, `{"activo":null}`, `{"valorDescuento":"x"}`} {
		call(t, http.MethodPut, "/ofertas?id=1", b, put, http.StatusBadRequest)
	}
	for _, b := range []string{
		`{"tipoDescuento":"OTRO"}`, `{"fechaInicio":"x"}`, `{"fechaFin":"x"}`, `{"horaInicio":"x"}`, `{"horaFin":"x"}`,
		`{"titulo":" "}`, `{"restauranteId":0}`, `{"valorDescuento":500}`, `{"diasSemana":["Funday"]}`,
		`{"horaInicio":null}`, `{"horaFin":null}`,
	} {
		call(t, http.MethodPut, "/ofertas?id=1", b, put, http.StatusUnprocessableEntity)
	}

	// {} conserva todo y responde 200.
	b := call(t, http.MethodPut, "/ofertas?id=1", `{}`, put, http.StatusOK)
	contains(t, b, `"titulo":"Martes"`, `"diasSemana":["Martes","Miércoles"]`, `"horaInicio":"08:00:00"`)
	if !e.has(`UPDATE "oferta"`) {
		t.Fatalf("debe actualizar: %v", e.execs)
	}

	// Merge de todos los campos; null limpia el horario.
	b = call(t, http.MethodPut, "/ofertas?id=1",
		`{"titulo":"Nuevo","tipoDescuento":"MONTO","valorDescuento":500,"fechaInicio":"2025-02-01","fechaFin":"2025-03-01","diasSemana":[],"horaInicio":null,"horaFin":null,"restauranteId":2,"activo":false}`, put, http.StatusOK)
	contains(t, b, `"ofertaId":1`)
	// Horas nuevas.
	call(t, http.MethodPut, "/ofertas?id=1", `{"horaInicio":"09:00","horaFin":"10:00"}`, put, http.StatusOK)

	execErrOn(`UPDATE "oferta"`, "duplicate key value (23505)")
	call(t, http.MethodPut, "/ofertas?id=1", `{"titulo":"Otro"}`, put, http.StatusConflict)
	execErrOn(`UPDATE "oferta"`, "violates foreign key constraint (23503)")
	call(t, http.MethodPut, "/ofertas?id=1", `{"restauranteId":99}`, put, http.StatusBadRequest)
	resetFake()
	setup(nil)
	execErrOn(`UPDATE "oferta"`, "boom")
	call(t, http.MethodPut, "/ofertas?id=1", `{"titulo":"Otro"}`, put, http.StatusInternalServerError)

	// La relectura posterior falla.
	setup(nil)
	n := 0
	prev := fakeQuery
	fakeQuery = func(q string, a []driver.NamedValue) (driver.Rows, error) {
		n++
		if n > 1 {
			return nil, errors.New("boom")
		}
		return prev(q, a)
	}
	call(t, http.MethodPut, "/ofertas?id=1", `{"titulo":"Otro"}`, put, http.StatusInternalServerError)
}

func TestDelete(t *testing.T) {
	defer resetFake()
	for _, q := range []string{"", "?id=abc", "?id=0"} {
		call(t, http.MethodDelete, "/ofertas"+q, "", del, http.StatusBadRequest)
	}
	call(t, http.MethodDelete, "/ofertas?id=1", "", del, http.StatusNotFound)
	e := setup(nil)
	call(t, http.MethodDelete, "/ofertas?id=1", "", del, http.StatusOK)
	if !e.has(`UPDATE "oferta"`) {
		t.Fatalf("debe desactivar: %v", e.execs)
	}
	execErrOn(`UPDATE "oferta"`, "boom")
	call(t, http.MethodDelete, "/ofertas?id=1", "", del, http.StatusInternalServerError)
	e = setup(func(v map[string]driver.Value) { v["activo"] = false })
	call(t, http.MethodDelete, "/ofertas?id=1", "", del, http.StatusBadRequest)
	if len(e.execs) != 0 {
		t.Fatalf("no debe escribir: %v", e.execs)
	}
}

func TestActivas(t *testing.T) {
	defer resetFake()
	for _, q := range []string{"", "?restaurante_id=abc", "?restaurante_id=0",
		"?restaurante_id=1&fecha=x", "?restaurante_id=1&hora=x", "?restaurante_id=1&producto_id=x", "?restaurante_id=1&producto_id=0"} {
		call(t, http.MethodGet, "/ofertas/activas"+q, "", activas, http.StatusBadRequest)
	}
	if b := call(t, http.MethodGet, "/ofertas/activas?restaurante_id=1", "", activas, http.StatusOK); !strings.Contains(b, `"data":[]`) {
		t.Fatalf("lista vacía debe ser []: %s", b)
	}
	setup(nil)
	// 2025-01-07 es martes.
	b := call(t, http.MethodGet, "/ofertas/activas?restaurante_id=1&fecha=2025-01-07&hora=10:00&producto_id=2", "", activas, http.StatusOK)
	contains(t, b, `"ofertaId":1`, `"productosIds":[2]`)
	// Producto que no está en la oferta: lista vacía.
	if b := call(t, http.MethodGet, "/ofertas/activas?restaurante_id=1&fecha=2025-01-07&hora=10:00&producto_id=9", "", activas, http.StatusOK); !strings.Contains(b, `"data":[]`) {
		t.Fatalf("debe ser []: %s", b)
	}
	queryErrOn(`FROM "oferta"`)
	call(t, http.MethodGet, "/ofertas/activas?restaurante_id=1", "", activas, http.StatusInternalServerError)
}

func TestAsociar(t *testing.T) {
	defer resetFake()
	call(t, http.MethodPost, "/ofertas/productos", `{"productoId":2}`, asociar, http.StatusBadRequest)
	call(t, http.MethodPost, "/ofertas/productos?id=1", "nojson", asociar, http.StatusBadRequest)
	call(t, http.MethodPost, "/ofertas/productos?id=1", `{}`, asociar, http.StatusBadRequest)
	call(t, http.MethodPost, "/ofertas/productos?id=1", `{"productoId":2}`, asociar, http.StatusNotFound) // oferta

	e := setup(nil)
	b := call(t, http.MethodPost, "/ofertas/productos?id=1", `{"productoId":2}`, asociar, http.StatusCreated)
	contains(t, b, `"ofertaId":1`, `"productoId":2`)
	if !e.has("INSERT INTO oferta_producto") {
		t.Fatalf("debe insertar: %v", e.execs)
	}

	// producto inexistente
	setup(nil)
	fakeQuery = queryFn(map[string]func(string) (driver.Rows, error){`FROM "oferta"`: func(q string) (driver.Rows, error) { return rowFor(q, ofertaVals()), nil }})
	call(t, http.MethodPost, "/ofertas/productos?id=1", `{"productoId":2}`, asociar, http.StatusNotFound)
	// errores de lectura
	queryErrOn(`FROM "oferta"`)
	call(t, http.MethodPost, "/ofertas/productos?id=1", `{"productoId":2}`, asociar, http.StatusInternalServerError)
	setup(nil)
	queryErrOn(`FROM "producto"`)
	call(t, http.MethodPost, "/ofertas/productos?id=1", `{"productoId":2}`, asociar, http.StatusInternalServerError)

	setup(nil)
	execErrOn("INSERT INTO oferta_producto", "duplicate key value (23505)")
	call(t, http.MethodPost, "/ofertas/productos?id=1", `{"productoId":2}`, asociar, http.StatusConflict)
	resetFake()
	setup(nil)
	execErrOn("INSERT INTO oferta_producto", "boom")
	call(t, http.MethodPost, "/ofertas/productos?id=1", `{"productoId":2}`, asociar, http.StatusInternalServerError)
}

func TestDesasociar(t *testing.T) {
	defer resetFake()
	for _, q := range []string{"", "?id=abc", "?id=1", "?id=1&producto_id=0"} {
		call(t, http.MethodDelete, "/ofertas/productos"+q, "", desasoc, http.StatusBadRequest)
	}
	fakeAffected = 0
	call(t, http.MethodDelete, "/ofertas/productos?id=1&producto_id=2", "", desasoc, http.StatusNotFound)
	fakeAffected = 1
	e := setup(nil)
	call(t, http.MethodDelete, "/ofertas/productos?id=1&producto_id=2", "", desasoc, http.StatusOK)
	if !e.has("DELETE FROM oferta_producto") {
		t.Fatalf("debe borrar: %v", e.execs)
	}
	execErrOn("DELETE FROM oferta_producto", "boom")
	call(t, http.MethodDelete, "/ofertas/productos?id=1&producto_id=2", "", desasoc, http.StatusInternalServerError)
}
