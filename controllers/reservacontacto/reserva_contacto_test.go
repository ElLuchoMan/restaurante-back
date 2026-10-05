package reservacontacto

import (
	"database/sql/driver"
	"net/http"
	"strings"
	"testing"
)

const qContacto = `FROM "reserva_contacto" T0`

func fila(id int64, docContacto, docCliente any) []driver.Value {
	return []driver.Value{id, "Ana Gómez", "3001234567", docContacto, docCliente}
}

func contactoRows(filas ...[]driver.Value) *res {
	return &res{match: qContacto, cols: 5, rows: filas}
}

func TestGetAllListaContactos(t *testing.T) {
	programa(t, contactoRows(fila(1, int64(55), nil), fila(2, nil, int64(9))))
	w := run("GET", "/reserva_contacto", "", (*ReservaContactoController).GetAll)
	r := expect(t, w, http.StatusOK)
	l := dataList(t, r)
	if len(l) != 2 {
		t.Fatalf("esperaba 2 contactos: %v", l)
	}
	if l[0].(map[string]any)["documentoContacto"] != float64(55) {
		t.Fatalf("contacto invitado: %v", l[0])
	}
	cli := l[1].(map[string]any)["documentoCliente"].(map[string]any)
	if cli["documentoCliente"] != float64(9) || len(cli) != 1 {
		t.Fatalf("el cliente solo debe exponer su documento: %v", cli)
	}
	sinPassword(t, w)
}

func TestGetAllVacioEsListaVacia(t *testing.T) {
	programa(t, contactoRows())
	w := run("GET", "/reserva_contacto", "", (*ReservaContactoController).GetAll)
	expect(t, w, http.StatusOK)
	if !strings.Contains(w.Body.String(), `"data":[]`) {
		t.Fatalf("data debe ser []: %s", w.Body.String())
	}
}

func TestGetAllFiltros(t *testing.T) {
	programa(t, contactoRows(fila(1, int64(55), int64(9))))
	expect(t, run("GET", "/reserva_contacto?documento_contacto=55&documento_cliente=9", "", (*ReservaContactoController).GetAll), http.StatusOK)
	q := sqlQue(qContacto)
	if !strings.Contains(q, `"documento_contacto"`) || !strings.Contains(q, `"pk_documento_cliente"`) {
		t.Fatalf("faltan filtros en el SQL: %s", q)
	}
}

func TestGetAllParametrosInvalidos(t *testing.T) {
	programa(t)
	for _, target := range []string{
		"/reserva_contacto?documento_contacto=x",
		"/reserva_contacto?documento_contacto=0",
		"/reserva_contacto?documento_cliente=-3",
		"/reserva_contacto?documento_cliente=abc",
	} {
		expect(t, run("GET", target, "", (*ReservaContactoController).GetAll), http.StatusBadRequest)
	}
}

func TestGetAllErrorDB(t *testing.T) {
	programa(t, conError(qContacto))
	expect(t, run("GET", "/reserva_contacto", "", (*ReservaContactoController).GetAll), http.StatusInternalServerError)
}

func TestGetByIdOK(t *testing.T) {
	programa(t, contactoRows(fila(3, int64(55), nil)))
	r := expect(t, run("GET", "/reserva_contacto/search?id=3", "", (*ReservaContactoController).GetById), http.StatusOK)
	if dataMap(t, r)["contactoId"] != float64(3) {
		t.Fatalf("contacto inesperado: %v", r.Data)
	}
}

func TestGetByIdInvalido(t *testing.T) {
	programa(t)
	for _, target := range []string{"/reserva_contacto/search", "/reserva_contacto/search?id=0", "/reserva_contacto/search?id=x"} {
		expect(t, run("GET", target, "", (*ReservaContactoController).GetById), http.StatusBadRequest)
	}
}

func TestGetByIdNoEncontradoYError(t *testing.T) {
	programa(t, contactoRows())
	expect(t, run("GET", "/reserva_contacto/search?id=3", "", (*ReservaContactoController).GetById), http.StatusNotFound)

	programa(t, conError(qContacto))
	expect(t, run("GET", "/reserva_contacto/search?id=3", "", (*ReservaContactoController).GetById), http.StatusInternalServerError)
}
