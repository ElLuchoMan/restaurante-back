package nomina

import (
	"database/sql/driver"
	"net/http"
	"strings"
	"testing"
	"time"
)

const (
	qNomina    = `FROM "nomina"`
	qMes       = `"fecha" >= `
	qPorID     = `"pk_id_nomina" = `
	qInsert    = `INSERT INTO "nomina"`
	qUpdate    = `UPDATE "nomina"`
	qControlIn = `INSERT INTO control_nomina`
)

var dia = time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC)

func fila(id int64, fecha time.Time, estado string) []driver.Value {
	return []driver.Value{id, fecha, int64(4500000), estado}
}

func nominaRows(match string, filas ...[]driver.Value) *res {
	return &res{match: match, cols: 4, rows: filas}
}

// ---------- GET /nominas ----------

func TestGetAllFiltrosYFormato(t *testing.T) {
	filas := [][]driver.Value{
		fila(1, dia, "NO_PAGO"),
		fila(2, time.Date(2025, 2, 25, 0, 0, 0, 0, time.UTC), "PAGO"),
		fila(3, time.Date(2024, 2, 21, 0, 0, 0, 0, time.UTC), "PAGO"),
	}
	casos := map[string]int{
		"/nominas":                        3,
		"/nominas?fecha=2025-01-20":       1,
		"/nominas?mes=2":                  2,
		"/nominas?anio=2025":              2,
		"/nominas?mes=2&anio=2025":        1,
		"/nominas?fecha=2025-01-21":       0,
		"/nominas?fecha=2025-01-20&mes=2": 0,
	}
	for target, esperadas := range casos {
		programa(t, nominaRows(qNomina, filas...))
		w := run("GET", target, "", (*NominaController).GetAll)
		r := expect(t, w, http.StatusOK)
		l := dataList(t, r)
		if len(l) != esperadas {
			t.Errorf("%s: esperaba %d nóminas, obtuve %d", target, esperadas, len(l))
		}
		if esperadas == 0 && !strings.Contains(w.Body.String(), `"data":[]`) {
			t.Errorf("%s: data debe ser []: %s", target, w.Body.String())
		}
	}
	programa(t, nominaRows(qNomina, filas[0]))
	r := expect(t, run("GET", "/nominas", "", (*NominaController).GetAll), http.StatusOK)
	m := dataList(t, r)[0].(map[string]any)
	if m["fechaNomina"] != "20-01-2025" || m["estadoNomina"] != "NO_PAGO" || m["monto"] != float64(4500000) || m["nominaId"] != float64(1) {
		t.Fatalf("forma inesperada: %v", m)
	}
}

func TestGetAllParametrosInvalidos(t *testing.T) {
	programa(t)
	for _, target := range []string{
		"/nominas?fecha=20-01-2025", "/nominas?mes=13", "/nominas?mes=0", "/nominas?mes=x",
		"/nominas?anio=0", "/nominas?anio=abc", "/nominas?anio=10000",
	} {
		expect(t, run("GET", target, "", (*NominaController).GetAll), http.StatusBadRequest)
	}
	if sqlEjecutado(qNomina) {
		t.Fatal("no debe consultar con parámetros inválidos")
	}
}

func TestGetAllErrorDB(t *testing.T) {
	programa(t, conError(qNomina))
	expect(t, run("GET", "/nominas", "", (*NominaController).GetAll), http.StatusInternalServerError)
}

// ---------- POST /nominas ----------

func TestPostCreaIgnorandoNominaIDYMonto(t *testing.T) {
	programa(t,
		nominaRows(qMes),
		nominaRows(qPorID, fila(7, dia, "NO_PAGO")),
	)
	body := `{"nominaId":99,"monto":123,"fechaNomina":"2025-01-20","estadoNomina":"NO_PAGO"}`
	w := run("POST", "/nominas", body, (*NominaController).Post)
	r := expect(t, w, http.StatusCreated)
	if dataMap(t, r)["nominaId"] != float64(7) {
		t.Fatalf("debe devolver la nómina leída tras el trigger: %v", r.Data)
	}
	ins := execQue(qInsert)
	if ins == nil {
		t.Fatal("debe insertar")
	}
	for _, a := range ins.args {
		if a.Value == int64(99) || a.Value == int64(123) {
			t.Fatalf("nominaId/monto del cuerpo no deben llegar al INSERT: %v", ins.args)
		}
	}
	if strings.Contains(ins.q, `"pk_id_nomina"`) && !strings.Contains(ins.q, "RETURNING") {
		t.Fatalf("el id lo genera la base: %s", ins.q)
	}
	if ins.args[1].Value != int64(0) {
		t.Fatalf("monto debe insertarse en 0 (lo calcula el trigger): %v", ins.args)
	}
}

func TestPostSinCuerpoUsaHoyYValidaDia(t *testing.T) {
	programa(t, nominaRows(qMes), nominaRows(qPorID, fila(7, dia, "NO_PAGO")))
	w := run("POST", "/nominas", "  ", (*NominaController).Post)
	hoy := time.Now().Day()
	if hoy < 20 {
		expect(t, w, http.StatusBadRequest)
		return
	}
	expect(t, w, http.StatusCreated)
}

func TestPostEstadoPorDefectoNoPago(t *testing.T) {
	programa(t, nominaRows(qMes), nominaRows(qPorID, fila(7, dia, "NO_PAGO")))
	expect(t, run("POST", "/nominas", `{"fechaNomina":"2025-01-20"}`, (*NominaController).Post), http.StatusCreated)
	ins := execQue(qInsert)
	if ins == nil || ins.args[2].Value != "NO_PAGO" {
		t.Fatalf("estado por defecto NO_PAGO: %v", ins)
	}
}

func TestPostValidaciones400(t *testing.T) {
	casos := map[string]string{
		"json inválido":    `{`,
		"fecha formato":    `{"fechaNomina":"20/01/2025"}`,
		"día antes del 20": `{"fechaNomina":"2025-01-19"}`,
		"estado inválido":  `{"fechaNomina":"2025-01-20","estadoNomina":"X"}`,
		"estado vacío":     `{"fechaNomina":"2025-01-20","estadoNomina":""}`,
	}
	for nombre, body := range casos {
		programa(t)
		expect(t, run("POST", "/nominas", body, (*NominaController).Post), http.StatusBadRequest)
		if sqlEjecutado(qNomina) {
			t.Errorf("%s: no debe consultar ni insertar", nombre)
		}
	}
}

func TestPostMesExistenteMarcaRegenerada(t *testing.T) {
	programa(t, nominaRows(qMes, fila(5, dia, "PAGO")))
	r := expect(t, run("POST", "/nominas", `{"fechaNomina":"2025-01-25"}`, (*NominaController).Post), http.StatusOK)
	if dataMap(t, r)["nominaId"] != float64(5) {
		t.Fatalf("debe devolver la existente: %v", r.Data)
	}
	if execQue(qControlIn) == nil {
		t.Fatal("debe marcar control_nomina como REGENERADA")
	}
	if execQue(qInsert) != nil {
		t.Fatal("no debe insertar otra nómina")
	}
}

func TestPostErrores(t *testing.T) {
	body := `{"fechaNomina":"2025-01-20"}`
	// error validando el mes
	programa(t, conError(qMes))
	expect(t, run("POST", "/nominas", body, (*NominaController).Post), http.StatusInternalServerError)

	// error marcando control_nomina
	programa(t, nominaRows(qMes, fila(5, dia, "PAGO")), conError(qControlIn))
	expect(t, run("POST", "/nominas", body, (*NominaController).Post), http.StatusInternalServerError)

	// error genérico al insertar
	programa(t, nominaRows(qMes), conError(qInsert))
	expect(t, run("POST", "/nominas", body, (*NominaController).Post), http.StatusInternalServerError)

	// carrera: la fecha ya existe (unicidad) -> 409
	programa(t, nominaRows(qMes), &res{match: qInsert, err: errUnique})
	expect(t, run("POST", "/nominas", body, (*NominaController).Post), http.StatusConflict)

	// error al releer la nómina creada
	programa(t, nominaRows(qMes), conError(qPorID))
	expect(t, run("POST", "/nominas", body, (*NominaController).Post), http.StatusInternalServerError)
}

// ---------- PUT /nominas ----------

func TestPutSinCuerpoMarcaPago(t *testing.T) {
	programa(t, nominaRows(qNomina, fila(5, dia, "NO_PAGO")))
	w := run("PUT", "/nominas?id=5", "", (*NominaController).Put)
	r := expect(t, w, http.StatusOK)
	if dataMap(t, r)["estadoNomina"] != "PAGO" {
		t.Fatalf("estado: %v", r.Data)
	}
	up := execQue(qUpdate)
	if up == nil || !strings.Contains(up.q, `"estado_nomina"`) || strings.Contains(up.q, `"monto"`) || strings.Contains(up.q, `"fecha"`) {
		t.Fatalf("solo debe actualizar estado_nomina: %v", up)
	}
}

func TestPutConEstadoExplicitoYCamposIgnorados(t *testing.T) {
	programa(t, nominaRows(qNomina, fila(5, dia, "PAGO")))
	r := expect(t, run("PUT", "/nominas?id=5", `{"estadoNomina":"NO_PAGO","monto":1,"fechaNomina":"2030-01-01","nominaId":9}`, (*NominaController).Put), http.StatusOK)
	m := dataMap(t, r)
	if m["estadoNomina"] != "NO_PAGO" || m["monto"] != float64(4500000) || m["fechaNomina"] != "20-01-2025" || m["nominaId"] != float64(5) {
		t.Fatalf("solo cambia el estado: %v", m)
	}
}

func TestPutObjetoVacioMarcaPago(t *testing.T) {
	programa(t, nominaRows(qNomina, fila(5, dia, "NO_PAGO")))
	expect(t, run("PUT", "/nominas?id=5", `{}`, (*NominaController).Put), http.StatusOK)
}

func TestPutYaTeniaEseEstadoEs409(t *testing.T) {
	programa(t, nominaRows(qNomina, fila(5, dia, "PAGO")))
	expect(t, run("PUT", "/nominas?id=5", "", (*NominaController).Put), http.StatusConflict)
	if sqlEjecutado(qUpdate) {
		t.Fatal("no debe actualizar")
	}
}

func TestPutValidaciones400(t *testing.T) {
	programa(t, nominaRows(qNomina, fila(5, dia, "NO_PAGO")))
	for _, target := range []string{"/nominas", "/nominas?id=0", "/nominas?id=-1", "/nominas?id=x"} {
		expect(t, run("PUT", target, "", (*NominaController).Put), http.StatusBadRequest)
	}
	for nombre, body := range map[string]string{
		"json inválido": `{`, "null": `{"estadoNomina":null}`, "estado inválido": `{"estadoNomina":"X"}`, "tipo": `{"estadoNomina":3}`,
	} {
		programa(t, nominaRows(qNomina, fila(5, dia, "NO_PAGO")))
		w := run("PUT", "/nominas?id=5", body, (*NominaController).Put)
		if w.Code != http.StatusBadRequest {
			t.Errorf("%s: esperaba 400, obtuve %d", nombre, w.Code)
		}
		decode(t, w)
		if sqlEjecutado(qUpdate) {
			t.Errorf("%s: no debe actualizar", nombre)
		}
	}
}

func TestPutNoEncontradoYErrores(t *testing.T) {
	programa(t, nominaRows(qNomina))
	expect(t, run("PUT", "/nominas?id=5", "", (*NominaController).Put), http.StatusNotFound)

	programa(t, conError(qNomina))
	expect(t, run("PUT", "/nominas?id=5", "", (*NominaController).Put), http.StatusInternalServerError)

	programa(t, nominaRows(qNomina, fila(5, dia, "NO_PAGO")), conError(qUpdate))
	expect(t, run("PUT", "/nominas?id=5", "", (*NominaController).Put), http.StatusInternalServerError)
}

// ---------- DELETE /nominas ----------

func TestDeleteMarcaNoPago(t *testing.T) {
	programa(t, nominaRows(qNomina, fila(5, dia, "PAGO")))
	r := expect(t, run("DELETE", "/nominas?id=5", "", (*NominaController).Delete), http.StatusOK)
	if dataMap(t, r)["estadoNomina"] != "NO_PAGO" {
		t.Fatalf("estado: %v", r.Data)
	}
	if execQue(qUpdate) == nil {
		t.Fatal("debe actualizar")
	}
}

func TestDeleteYaNoPagoEs409(t *testing.T) {
	programa(t, nominaRows(qNomina, fila(5, dia, "NO_PAGO")))
	expect(t, run("DELETE", "/nominas?id=5", "", (*NominaController).Delete), http.StatusConflict)
}

func TestDeleteInvalidoNoEncontradoYErrores(t *testing.T) {
	programa(t)
	for _, target := range []string{"/nominas", "/nominas?id=0", "/nominas?id=x"} {
		expect(t, run("DELETE", target, "", (*NominaController).Delete), http.StatusBadRequest)
	}
	expect(t, run("DELETE", "/nominas?id=5", "", (*NominaController).Delete), http.StatusNotFound)

	programa(t, conError(qNomina))
	expect(t, run("DELETE", "/nominas?id=5", "", (*NominaController).Delete), http.StatusInternalServerError)

	programa(t, nominaRows(qNomina, fila(5, dia, "PAGO")), conError(qUpdate))
	expect(t, run("DELETE", "/nominas?id=5", "", (*NominaController).Delete), http.StatusInternalServerError)
}
