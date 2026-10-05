package controlnomina

import (
	"database/sql/driver"
	"net/http"
	"strings"
	"testing"
	"time"
)

const qControl = `FROM "control_nomina" T0`

var dia = time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC)

func controlRows(filas ...[]driver.Value) *res {
	return &res{match: qControl, cols: 3, rows: filas}
}

func fila(id int64, estado string) []driver.Value { return []driver.Value{id, dia, estado} }

func TestGetAllLista(t *testing.T) {
	programa(t, controlRows(fila(1, "GENERADA"), fila(2, "NO GENERADA")))
	r := expect(t, run("GET", "/control_nomina", "", (*ControlNominaController).GetAll), http.StatusOK)
	l := dataList(t, r)
	if len(l) != 2 {
		t.Fatalf("esperaba 2 registros: %v", l)
	}
	m := l[0].(map[string]any)
	if m["fecha"] != "20-01-2025" || m["estado"] != "GENERADA" || m["controlNominaId"] != float64(1) {
		t.Fatalf("registro inesperado: %v", m)
	}
}

func TestGetAllVacioEsListaVacia(t *testing.T) {
	programa(t, controlRows())
	w := run("GET", "/control_nomina?fecha=2025-01-20", "", (*ControlNominaController).GetAll)
	expect(t, w, http.StatusOK)
	if !strings.Contains(w.Body.String(), `"data":[]`) {
		t.Fatalf("data debe ser []: %s", w.Body.String())
	}
	if !strings.Contains(sqlQue(qControl), `"fecha"`) {
		t.Fatal("debe filtrar por fecha")
	}
}

func TestGetAllFechaInvalida(t *testing.T) {
	programa(t)
	expect(t, run("GET", "/control_nomina?fecha=20-01-2025", "", (*ControlNominaController).GetAll), http.StatusBadRequest)
	if sqlEjecutado(qControl) {
		t.Fatal("no debe consultar con fecha inválida")
	}
}

func TestGetAllErrorDB(t *testing.T) {
	programa(t, conError(qControl))
	expect(t, run("GET", "/control_nomina", "", (*ControlNominaController).GetAll), http.StatusInternalServerError)
}

func TestGetByIdOK(t *testing.T) {
	programa(t, controlRows(fila(4, "REGENERADA")))
	r := expect(t, run("GET", "/control_nomina/search?id=4", "", (*ControlNominaController).GetById), http.StatusOK)
	if dataMap(t, r)["estado"] != "REGENERADA" {
		t.Fatalf("registro inesperado: %v", r.Data)
	}
}

func TestGetByIdInvalido(t *testing.T) {
	programa(t)
	for _, target := range []string{"/control_nomina/search", "/control_nomina/search?id=0", "/control_nomina/search?id=x"} {
		expect(t, run("GET", target, "", (*ControlNominaController).GetById), http.StatusBadRequest)
	}
}

func TestGetByIdNoEncontradoYError(t *testing.T) {
	programa(t, controlRows())
	expect(t, run("GET", "/control_nomina/search?id=4", "", (*ControlNominaController).GetById), http.StatusNotFound)

	programa(t, conError(qControl))
	expect(t, run("GET", "/control_nomina/search?id=4", "", (*ControlNominaController).GetById), http.StatusInternalServerError)
}
