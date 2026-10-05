package nominatrabajador

import (
	"database/sql/driver"
	"net/http"
	"strings"
	"testing"
	"time"

	"restaurante/models"
)

const (
	qNT         = `FROM "nomina_trabajador" T0`
	qTrabajador = `FROM "trabajador" T0`
	qNomina     = `FROM "nomina" T0`
	qIncidencia = `FROM "incidencia" T0`
	qInsertNT   = `INSERT INTO "nomina_trabajador"`
	qRawNT      = `FROM "nomina_trabajador" nt`
	qRawMes     = `JOIN "trabajador" t`
)

var dia = time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC)

func filaNT(id int64) []driver.Value {
	return []driver.Value{id, int64(2000000), int64(50000), "detalle", int64(1015), int64(5)}
}

func filaTrabajador() []driver.Value {
	return []driver.Value{int64(1015), "Juan", "Pérez", int64(2000000), nil, nil, false, "Mesero", dia, nil, "secreto-hash", nil}
}

func trabajadorRows() *res {
	return &res{match: qTrabajador, cols: 12, rows: [][]driver.Value{filaTrabajador()}}
}

func nominaRows() *res {
	return &res{match: qNomina, cols: 4, rows: [][]driver.Value{{int64(5), dia, int64(0), "NO_PAGO"}}}
}

func ntRows(filas ...[]driver.Value) *res { return &res{match: qNT, cols: 6, rows: filas} }

func incidenciaRows(montos ...int64) *res {
	var filas [][]driver.Value
	for i, m := range montos {
		filas = append(filas, []driver.Value{int64(i + 1), dia, m, i%2 == 0, "motivo", int64(1015)})
	}
	return &res{match: qIncidencia, cols: 6, rows: filas}
}

var colsItem = []string{"pk_id_nomina_trabajador", "sueldo_base", "monto_incidencias", "detalles", "pk_documento_trabajador", "pk_id_nomina"}

func rawItems(filas ...[]driver.Value) *res {
	return &res{match: qRawNT, names: colsItem, rows: filas}
}

// ---------- GET /nomina_trabajador ----------

func TestGetAllItemsConIdsYNulosComoCero(t *testing.T) {
	programa(t, ntRows(filaNT(15), []driver.Value{int64(16), int64(1), nil, nil, int64(2), int64(5)}))
	w := run("GET", "/nomina_trabajador", "", (*NominaTrabajadorController).GetAll)
	l := dataList(t, expect(t, w, http.StatusOK))
	if len(l) != 2 {
		t.Fatalf("esperaba 2: %v", l)
	}
	m := l[0].(map[string]any)
	if m["nominaTrabajadorId"] != float64(15) || m["documentoTrabajador"] != float64(1015) || m["nominaId"] != float64(5) || m["detalles"] != "detalle" {
		t.Fatalf("forma inesperada: %v", m)
	}
	n := l[1].(map[string]any)
	if n["montoIncidencias"] != float64(0) || n["detalles"] != "" {
		t.Fatalf("nulos como 0 y \"\": %v", n)
	}
	sinPassword(t, w)
}

func TestItemDeSinFK(t *testing.T) {
	item := itemDe(modeloVacio())
	if item.PK_DOCUMENTO_TRABAJADOR != 0 || item.PK_ID_NOMINA != 0 {
		t.Fatalf("sin FK debe dar 0: %+v", item)
	}
}

func TestGetAllVacioYError(t *testing.T) {
	programa(t, ntRows())
	w := run("GET", "/nomina_trabajador", "", (*NominaTrabajadorController).GetAll)
	expect(t, w, http.StatusOK)
	if !strings.Contains(w.Body.String(), `"data":[]`) {
		t.Fatalf("data debe ser []: %s", w.Body.String())
	}
	programa(t, conError(qNT))
	expect(t, run("GET", "/nomina_trabajador", "", (*NominaTrabajadorController).GetAll), http.StatusInternalServerError)
}

// ---------- POST /nomina_trabajador ----------

func reglasPost(extra ...*res) []*res {
	return append([]*res{trabajadorRows(), nominaRows()}, extra...)
}

func TestPostCreaConFormaUnica(t *testing.T) {
	programa(t, reglasPost(ntRows(), incidenciaRows(100000, 30000, 5000))...)
	w := run("POST", "/nomina_trabajador", `{"documentoTrabajador":1015,"detalles":"hack","sueldoBase":1}`, (*NominaTrabajadorController).Post)
	r := expect(t, w, http.StatusCreated)
	m := dataMap(t, r)
	// incidencias: +100000 (resta=true -> resta), 30000 (suma), 5000 (resta)
	if m["nominaTrabajadorId"] != float64(7) || m["documentoTrabajador"] != float64(1015) || m["nominaId"] != float64(5) ||
		m["sueldoBase"] != float64(2000000) || m["montoIncidencias"] != float64(-75000) ||
		m["detalles"] != "Nómina del mes de Enero de 2025 más incidencias si aplica" {
		t.Fatalf("forma inesperada: %v", m)
	}
	ins := execQue(qInsertNT)
	if ins == nil {
		t.Fatal("debe insertar")
	}
	for _, a := range ins.args {
		if a.Value == "hack" || a.Value == int64(1) {
			t.Fatalf("el cuerpo no debe fijar sueldo ni detalle: %v", ins.args)
		}
	}
	sinPassword(t, w)
}

func TestPostExistenteMismaForma(t *testing.T) {
	programa(t, reglasPost(ntRows(filaNT(15)))...)
	r := expect(t, run("POST", "/nomina_trabajador", `{"documentoTrabajador":1015}`, (*NominaTrabajadorController).Post), http.StatusOK)
	m := dataMap(t, r)
	if m["nominaTrabajadorId"] != float64(15) || m["documentoTrabajador"] != float64(1015) {
		t.Fatalf("debe devolver la existente con la misma forma: %v", m)
	}
	if execQue(qInsertNT) != nil {
		t.Fatal("no debe insertar")
	}
}

func TestPostValidaciones400(t *testing.T) {
	for nombre, body := range map[string]string{
		"vacío": ``, "json inválido": `{`, "sin documento": `{}`, "documento cero": `{"documentoTrabajador":0}`,
		"documento negativo": `{"documentoTrabajador":-5}`, "tipo": `{"documentoTrabajador":"x"}`,
	} {
		programa(t)
		w := run("POST", "/nomina_trabajador", body, (*NominaTrabajadorController).Post)
		if w.Code != http.StatusBadRequest {
			t.Errorf("%s: esperaba 400, obtuve %d", nombre, w.Code)
		}
		decode(t, w)
	}
}

func TestPostTrabajadorNoEncontrado404(t *testing.T) {
	programa(t)
	expect(t, run("POST", "/nomina_trabajador", `{"documentoTrabajador":1015}`, (*NominaTrabajadorController).Post), http.StatusNotFound)
}

func TestPostSinNominaEs422(t *testing.T) {
	programa(t, trabajadorRows())
	expect(t, run("POST", "/nomina_trabajador", `{"documentoTrabajador":1015}`, (*NominaTrabajadorController).Post), http.StatusUnprocessableEntity)
}

func TestPostErrores(t *testing.T) {
	body := `{"documentoTrabajador":1015}`
	casos := []struct {
		nombre string
		status int
		reglas []*res
	}{
		{"trabajador", 500, []*res{conError(qTrabajador)}},
		{"nómina", 500, []*res{trabajadorRows(), conError(qNomina)}},
		{"existente", 500, reglasPost(conError(qNT))},
		{"incidencias", 500, reglasPost(ntRows(), conError(qIncidencia))},
		{"insert", 500, reglasPost(ntRows(), incidenciaRows(), conError(qInsertNT))},
		{"duplicado", 409, reglasPost(ntRows(), incidenciaRows(), &res{match: qInsertNT, err: errUnique})},
	}
	for _, c := range casos {
		programa(t, c.reglas...)
		w := run("POST", "/nomina_trabajador", body, (*NominaTrabajadorController).Post)
		if w.Code != c.status {
			t.Errorf("%s: esperaba %d, obtuve %d: %s", c.nombre, c.status, w.Code, w.Body.String())
			continue
		}
		decode(t, w)
	}
}

func TestMesEnEspañol(t *testing.T) {
	for m := time.January; m <= time.December; m++ {
		if obtenerMesEnEspañol(m) == "" {
			t.Fatalf("mes %d sin nombre", m)
		}
	}
	if obtenerMesEnEspañol(time.December) != "Diciembre" {
		t.Fatal("diciembre")
	}
}

// ---------- GET /nomina_trabajador/search ----------

func TestSearchFiltrosArmanSQL(t *testing.T) {
	programa(t, rawItems(filaNT(15)))
	w := run("GET", "/nomina_trabajador/search?documento=1015&actual=true&pagas=true&mes=1&anio=2025", "", (*NominaTrabajadorController).GetByTrabajador)
	l := dataList(t, expect(t, w, http.StatusOK))
	if len(l) != 1 || l[0].(map[string]any)["nominaTrabajadorId"] != float64(15) {
		t.Fatalf("resultado: %v", l)
	}
	q := sqlQue(qRawNT)
	for _, frag := range []string{`MAX("fecha")`, `'PAGO'`, `EXTRACT(MONTH`, `EXTRACT(YEAR`} {
		if !strings.Contains(q, frag) {
			t.Errorf("falta %q en %s", frag, q)
		}
	}
	if strings.Contains(q, `'NO_PAGO'`) {
		t.Error("pagas no debe incluir NO_PAGO")
	}
}

func TestSearchNoPagasSoloMesOAnio(t *testing.T) {
	programa(t, rawItems())
	w := run("GET", "/nomina_trabajador/search?documento=1015&no_pagas=1&anio=2025", "", (*NominaTrabajadorController).GetByTrabajador)
	expect(t, w, http.StatusOK)
	if !strings.Contains(w.Body.String(), `"data":[]`) {
		t.Fatalf("data debe ser []: %s", w.Body.String())
	}
	q := sqlQue(qRawNT)
	if !strings.Contains(q, `'NO_PAGO'`) || !strings.Contains(q, `EXTRACT(YEAR`) || strings.Contains(q, `EXTRACT(MONTH`) {
		t.Fatalf("SQL inesperado: %s", q)
	}
}

func TestSearchParametrosInvalidos(t *testing.T) {
	programa(t)
	for _, target := range []string{
		"/s", "/s?documento=0", "/s?documento=x", "/s?documento=-1",
		"/s?documento=1&actual=quizas", "/s?documento=1&pagas=x", "/s?documento=1&no_pagas=x",
		"/s?documento=1&pagas=true&no_pagas=true",
		"/s?documento=1&mes=13", "/s?documento=1&mes=0", "/s?documento=1&mes=x",
		"/s?documento=1&anio=0", "/s?documento=1&anio=x",
	} {
		expect(t, run("GET", target, "", (*NominaTrabajadorController).GetByTrabajador), http.StatusBadRequest)
	}
	if sqlEjecutado(qRawNT) {
		t.Fatal("no debe consultar con parámetros inválidos")
	}
}

func TestSearchErrorDB(t *testing.T) {
	programa(t, conError(qRawNT))
	expect(t, run("GET", "/s?documento=1015", "", (*NominaTrabajadorController).GetByTrabajador), http.StatusInternalServerError)
}

// ---------- GET /nomina_trabajador/mes ----------

var colsDetalle = append(append([]string{}, colsItem...), "nombre", "apellido")

func TestMesIncluyeNominaTrabajadorID(t *testing.T) {
	fila := append(filaNT(15), "Juan", "Pérez")
	programa(t, &res{match: qRawMes, names: colsDetalle, rows: [][]driver.Value{fila}})
	w := run("GET", "/nomina_trabajador/mes?mes=1&anio=2025", "", (*NominaTrabajadorController).GetNominasByMes)
	m := dataList(t, expect(t, w, http.StatusOK))[0].(map[string]any)
	if m["nominaTrabajadorId"] != float64(15) || m["nombre"] != "Juan" || m["apellido"] != "Pérez" || m["nominaId"] != float64(5) {
		t.Fatalf("forma inesperada: %v", m)
	}
	if !strings.Contains(sqlQue(qRawMes), `nt."pk_id_nomina_trabajador"`) {
		t.Fatal("el SELECT debe incluir pk_id_nomina_trabajador")
	}
}

func TestMesPorDefectoEsElActual(t *testing.T) {
	programa(t, &res{match: qRawMes, names: colsDetalle})
	w := run("GET", "/nomina_trabajador/mes", "", (*NominaTrabajadorController).GetNominasByMes)
	expect(t, w, http.StatusOK)
	if !strings.Contains(w.Body.String(), `"data":[]`) {
		t.Fatalf("data debe ser []: %s", w.Body.String())
	}
	if len(execArgsQuery) != 2 || execArgsQuery[0] != int64(time.Now().Month()) || execArgsQuery[1] != int64(time.Now().Year()) {
		t.Fatalf("debe usar mes y año actuales: %v", execArgsQuery)
	}
}

func TestMesInvalidosYError(t *testing.T) {
	programa(t)
	for _, target := range []string{"/m?mes=13", "/m?mes=x", "/m?anio=0", "/m?anio=x"} {
		expect(t, run("GET", target, "", (*NominaTrabajadorController).GetNominasByMes), http.StatusBadRequest)
	}
	programa(t, conError(qRawMes))
	expect(t, run("GET", "/m?mes=1&anio=2025", "", (*NominaTrabajadorController).GetNominasByMes), http.StatusInternalServerError)
}

func modeloVacio() models.NominaTrabajador { return models.NominaTrabajador{} }
