package reserva

import (
	"bytes"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	loginc "restaurante/controllers/login"
	"restaurante/internal/ratelimit"
	"restaurante/models"
)

const (
	docDueno    = int64(9)  // cliente registrado dueño de las reservas de prueba
	docInvitado = int64(55) // documento del contacto invitado de filaReserva
	docAjeno    = int64(77)
)

func tokCliente(doc int64) string { return tokenDe(rolCliente, doc) }

func esMinima(t *testing.T, w *bytes.Buffer) {
	t.Helper()
	b := w.String()
	for _, prohibido := range []string{"Ana Gómez", "3001234567", "nombreCompleto", "telefono", "documentoContacto", "documentoCliente", "contactoId", "password"} {
		if strings.Contains(b, prohibido) {
			t.Fatalf("la vista mínima no debe contener %q: %s", prohibido, b)
		}
	}
}

// ---------- GET /reservas/search ----------

func TestGetByIdSinTokenEs401(t *testing.T) {
	programa(t)
	expect(t, runAs("", "GET", "/reservas/search?id=3", "", (*ReservaController).GetById), http.StatusUnauthorized)
	expect(t, runAs("Bearer basura", "GET", "/reservas/search?id=3", "", (*ReservaController).GetById), http.StatusUnauthorized)
}

func TestGetByIdClienteSoloLasSuyas(t *testing.T) {
	// dueño por documento de cliente registrado
	programa(t, reservaRows(filaReserva(3, "PENDIENTE", int64(9))))
	expect(t, runAs(tokCliente(docDueno), "GET", "/reservas/search?id=3", "", (*ReservaController).GetById), http.StatusOK)
	// dueño por documento de contacto (invitado)
	programa(t, reservaRows(filaReserva(3, "PENDIENTE", nil)))
	expect(t, runAs(tokCliente(docInvitado), "GET", "/reservas/search?id=3", "", (*ReservaController).GetById), http.StatusOK)
	// ajeno: 404 idéntico a "no existe" (no revela la reserva ni su existencia)
	programa(t, reservaRows(filaReserva(3, "PENDIENTE", int64(9))))
	ajeno := runAs(tokCliente(docAjeno), "GET", "/reservas/search?id=3", "", (*ReservaController).GetById)
	programa(t)
	inexistente := runAs(tokCliente(docAjeno), "GET", "/reservas/search?id=3", "", (*ReservaController).GetById)
	expect(t, ajeno, http.StatusNotFound)
	expect(t, inexistente, http.StatusNotFound)
	if ajeno.Body.String() != inexistente.Body.String() {
		t.Fatalf("404 distinguibles:\n%s\n%s", ajeno.Body.String(), inexistente.Body.String())
	}
}

func TestPoseeReservaSinContacto(t *testing.T) {
	if poseeReserva(&loginc.Claims{Documento: 1}, &models.Reserva{}) {
		t.Fatal("sin contacto no hay dueño")
	}
}

// ---------- GET /reservas/cliente ----------

func TestClienteSinTokenEs401(t *testing.T) {
	programa(t)
	expect(t, runAs("", "GET", "/reservas/cliente?documentoCliente=9", "", (*ReservaController).GetByDocumentoCliente), http.StatusUnauthorized)
}

func TestClienteUsaDocumentoDelToken(t *testing.T) {
	for _, target := range []string{"/reservas/cliente", "/reservas/cliente?documentoCliente=9"} {
		programa(t, reservaRows(filaReserva(1, "PENDIENTE", int64(9))))
		expect(t, runAs(tokCliente(docDueno), "GET", target, "", (*ReservaController).GetByDocumentoCliente), http.StatusOK)
		if q := sqlQue(qReservaRel); !strings.Contains(q, `"pk_documento_cliente" = $1`) {
			t.Fatalf("debe filtrar por el documento del cliente: %s", q)
		}
		if got := argQue(qReservaRel); got != int64(9) {
			t.Fatalf("documento enviado a la BD: %v", got)
		}
	}
}

func TestClienteDocumentoAjenoEs403(t *testing.T) {
	programa(t)
	expect(t, runAs(tokCliente(docDueno), "GET", "/reservas/cliente?documentoCliente=77", "", (*ReservaController).GetByDocumentoCliente), http.StatusForbidden)
	if sqlEjecutado(qReservaRel) {
		t.Fatal("no debe consultar la BD si el documento no coincide")
	}
}

func TestClienteDocumentoInvalidoEs400YPersonalNecesitaDocumento(t *testing.T) {
	programa(t)
	expect(t, runAs(tokCliente(docDueno), "GET", "/reservas/cliente?documentoCliente=abc", "", (*ReservaController).GetByDocumentoCliente), http.StatusBadRequest)
	expect(t, run("GET", "/reservas/cliente", "", (*ReservaController).GetByDocumentoCliente), http.StatusBadRequest)
	programa(t, reservaRows(filaReserva(1, "PENDIENTE", int64(9))))
	expect(t, run("GET", "/reservas/cliente?documentoCliente=9", "", (*ReservaController).GetByDocumentoCliente), http.StatusOK)
}

// ---------- PUT ----------

func TestPutSinTokenEs401(t *testing.T) {
	programa(t)
	expect(t, runAs("", "PUT", "/reservas?id=3", `{"personas":2}`, (*ReservaController).Put), http.StatusUnauthorized)
	if len(execs) != 0 {
		t.Fatal("no debe escribir")
	}
}

func TestPutClienteAjenoEs404SinEscribir(t *testing.T) {
	programa(t, reservaRows(filaReserva(3, "PENDIENTE", int64(9))))
	expect(t, runAs(tokCliente(docAjeno), "PUT", "/reservas?id=3", `{"personas":2}`, (*ReservaController).Put), http.StatusNotFound)
	if len(execs) != 0 {
		t.Fatalf("no debe escribir: %v", execs)
	}
}

func TestPutClienteDuenoPuedeEditar(t *testing.T) {
	programa(t, putReglas()...)
	programa(t, reservaRows(filaReserva(3, "PENDIENTE", int64(9))))
	expect(t, runAs(tokCliente(docDueno), "PUT", "/reservas?id=3", `{"personas":6,"estadoReserva":"CANCELADA"}`, (*ReservaController).Put), http.StatusOK)
	if execQue(qUpdate) == nil {
		t.Fatal("debía actualizar")
	}
}

func TestPutClienteDuenoMismoContactoPermitido(t *testing.T) {
	programa(t, contactoRows(filaContacto(2, int64(55), nil)), reservaRows(filaReserva(3, "PENDIENTE", nil)))
	expect(t, runAs(tokCliente(docInvitado), "PUT", "/reservas?id=3", `{"documentoContacto":55}`, (*ReservaController).Put), http.StatusOK)
	programa(t, contactoRows(filaContacto(2, nil, int64(9))), reservaRows(filaReserva(3, "PENDIENTE", int64(9))))
	expect(t, runAs(tokCliente(docDueno), "PUT", "/reservas?id=3", `{"documentoCliente":9}`, (*ReservaController).Put), http.StatusOK)
}

func TestPutClienteNoPuedeReasignarNiCambiarEstado(t *testing.T) {
	casos := map[string]string{
		"otro documentoContacto": `{"documentoContacto":77,"nombreCompleto":"X"}`,
		"otro documentoCliente":  `{"documentoCliente":77}`,
		"confirmar":              `{"estadoReserva":"CONFIRMADA"}`,
	}
	for nombre, body := range casos {
		programa(t, reservaRows(filaReserva(3, "PENDIENTE", int64(9))))
		expect(t, runAs(tokCliente(docDueno), "PUT", "/reservas?id=3", body, (*ReservaController).Put), http.StatusForbidden)
		if len(execs) != 0 {
			t.Fatalf("%s: no debe escribir", nombre)
		}
	}
}

func TestPutPersonalPuedeTodo(t *testing.T) {
	programa(t, reservaRows(filaReserva(3, "PENDIENTE", int64(9))))
	expect(t, run("PUT", "/reservas?id=3", `{"estadoReserva":"CONFIRMADA"}`, (*ReservaController).Put), http.StatusOK)
}

// ---------- DELETE ----------

func TestDeleteSinTokenEs401(t *testing.T) {
	programa(t)
	expect(t, runAs("", "DELETE", "/reservas?id=3", "", (*ReservaController).Delete), http.StatusUnauthorized)
}

func TestDeleteClienteSoloLasSuyas(t *testing.T) {
	programa(t, reservaRows(filaReserva(3, "PENDIENTE", int64(9))))
	expect(t, runAs(tokCliente(docAjeno), "DELETE", "/reservas?id=3", "", (*ReservaController).Delete), http.StatusNotFound)
	if len(execs) != 0 {
		t.Fatal("no debe cancelar")
	}
	programa(t, reservaRows(filaReserva(3, "PENDIENTE", int64(9))))
	expect(t, runAs(tokCliente(docDueno), "DELETE", "/reservas?id=3", "", (*ReservaController).Delete), http.StatusOK)
}

// ---------- POST ----------

const cuerpoDocContacto = `{"documentoContacto":55,"nombreCompleto":"Ana Gómez","restauranteId":1,"fechaReserva":"2025-01-31","horaReserva":"18:30:00","personas":2%s}`

func cuerpoCon(extra string) string { return strings.Replace(cuerpoDocContacto, "%s", extra, 1) }

func TestPostInvitadoRecibeVistaMinima(t *testing.T) {
	for _, extra := range []string{``, `,"estadoReserva":""`, `,"estadoReserva":"PENDIENTE"`} {
		programa(t, postOK(contactoRows(filaContacto(2, int64(55), nil)))...)
		w := runAs("", "POST", "/reservas", cuerpoCon(extra), (*ReservaController).Post)
		r := expect(t, w, http.StatusCreated)
		m := dataMap(t, r)
		if m["reservaId"] != float64(7) || m["estadoReserva"] != "PENDIENTE" || m["personas"] != float64(4) {
			t.Fatalf("data inesperada: %v", m)
		}
		if m["restaurante"].(map[string]any)["nombreRestaurante"] != "Sazón Criolla" {
			t.Fatalf("restaurante: %v", m)
		}
		esMinima(t, w.Body)
	}
}

func TestPostInvitadoNoPuedeFijarEstado(t *testing.T) {
	programa(t)
	expect(t, runAs("", "POST", "/reservas", cuerpoCon(`,"estadoReserva":"CONFIRMADA"`), (*ReservaController).Post), http.StatusForbidden)
	expect(t, runAs(tokCliente(docInvitado), "POST", "/reservas", cuerpoCon(`,"estadoReserva":"CUMPLIDA"`), (*ReservaController).Post), http.StatusForbidden)
	if len(execs) != 0 {
		t.Fatal("no debe escribir")
	}
}

func TestPostDocumentoClienteExigeSuToken(t *testing.T) {
	body := `{"documentoCliente":9,"restauranteId":1,"fechaReserva":"2025-01-31","horaReserva":"18:30:00","personas":2}`
	programa(t)
	expect(t, runAs("", "POST", "/reservas", body, (*ReservaController).Post), http.StatusUnauthorized)
	expect(t, runAs(tokCliente(docAjeno), "POST", "/reservas", body, (*ReservaController).Post), http.StatusForbidden)
	if len(execs) != 0 {
		t.Fatal("no debe escribir")
	}
	// con su propio token recibe la reserva completa
	programa(t, contactoRows(filaContacto(2, nil, int64(9))), restauranteRows(), reservaRows(filaReserva(7, "PENDIENTE", int64(9))))
	w := runAs(tokCliente(docDueno), "POST", "/reservas", body, (*ReservaController).Post)
	r := expect(t, w, http.StatusCreated)
	if dataMap(t, r)["contactoId"].(map[string]any)["nombreCompleto"] != "Ana Gómez" {
		t.Fatalf("el dueño debe recibir la reserva completa: %s", w.Body.String())
	}
}

func TestPostClienteConDocumentoContactoPropioRecibeCompleta(t *testing.T) {
	programa(t, postOK(contactoRows(filaContacto(2, int64(55), nil)))...)
	w := runAs(tokCliente(docInvitado), "POST", "/reservas", cuerpoCon(``), (*ReservaController).Post)
	r := expect(t, w, http.StatusCreated)
	if dataMap(t, r)["contactoId"] == nil {
		t.Fatalf("esperaba la reserva completa: %s", w.Body.String())
	}
	// un Cliente que reserva con el documento de otra persona no ve sus datos
	programa(t, postOK(contactoRows(filaContacto(2, int64(55), nil)))...)
	w = runAs(tokCliente(docAjeno), "POST", "/reservas", cuerpoCon(``), (*ReservaController).Post)
	expect(t, w, http.StatusCreated)
	esMinima(t, w.Body)
}

func TestPostDocumentoClienteConDocumentoContactoNoExigeToken(t *testing.T) {
	// prevalece documentoContacto, así que documentoCliente no se usa
	body := `{"documentoContacto":55,"documentoCliente":9,"restauranteId":1,"fechaReserva":"2025-01-31","horaReserva":"18:30:00","personas":2}`
	programa(t, postOK(contactoRows(filaContacto(2, int64(55), nil)))...)
	expect(t, runAs("", "POST", "/reservas", body, (*ReservaController).Post), http.StatusCreated)
}

func TestPostPersonalVeReservaCompleta(t *testing.T) {
	programa(t, postOK(contactoRows(filaContacto(2, int64(55), nil)))...)
	w := run("POST", "/reservas", cuerpoCon(`,"estadoReserva":"CONFIRMADA"`), (*ReservaController).Post)
	r := expect(t, w, http.StatusCreated)
	if dataMap(t, r)["contactoId"].(map[string]any)["telefono"] != "3001234567" {
		t.Fatalf("el personal debe recibir la reserva completa: %s", w.Body.String())
	}
}

// ---------- GET /reservas/consulta ----------

func nuevoLimitador(max int) {
	consultaRL = ratelimit.New(max, time.Minute, 100)
}

func TestConsultaCoincidencias(t *testing.T) {
	casos := map[string]string{
		"telefono":                "3001234567",
		"telefono con espacios":   "300 123-4567",
		"telefono con indicativo": "+57 300 123 4567",
		"documento invitado":      "55",
		"documento cliente":       "9",
	}
	for nombre, contacto := range casos {
		nuevoLimitador(100)
		programa(t, reservaRows(filaReserva(3, "CONFIRMADA", int64(9))))
		w := runAs("", "GET", "/reservas/consulta?reservaId=3&contacto="+url.QueryEscape(contacto), "", (*ReservaController).Consulta)
		r := expect(t, w, http.StatusOK)
		m := dataMap(t, r)
		if m["reservaId"] != float64(3) || m["estadoReserva"] != "CONFIRMADA" || m["fechaReserva"] != "31-01-2025" || m["horaReserva"] != "18:30:00" {
			t.Fatalf("%s: data inesperada: %v", nombre, m)
		}
		esMinima(t, w.Body)
		if len(m) != 6 {
			t.Fatalf("%s: la vista mínima debe tener 6 campos: %v", nombre, m)
		}
	}
}

func TestConsultaNoCoincideEsIgualQueInexistente(t *testing.T) {
	nuevoLimitador(100)
	for _, contacto := range []string{"3009999999", "56", "99999999999999999999", "abc"} {
		programa(t, reservaRows(filaReserva(3, "CONFIRMADA", int64(9))))
		malo := runAs("", "GET", "/reservas/consulta?reservaId=3&contacto="+contacto, "", (*ReservaController).Consulta)
		programa(t)
		inexistente := runAs("", "GET", "/reservas/consulta?reservaId=3&contacto="+contacto, "", (*ReservaController).Consulta)
		if contacto == "abc" {
			expect(t, malo, http.StatusBadRequest)
			continue
		}
		expect(t, malo, http.StatusNotFound)
		expect(t, inexistente, http.StatusNotFound)
		if malo.Body.String() != inexistente.Body.String() {
			t.Fatalf("404 distinguibles para %q:\n%s\n%s", contacto, malo.Body.String(), inexistente.Body.String())
		}
	}
}

func TestConsultaParametrosInvalidos(t *testing.T) {
	nuevoLimitador(100)
	programa(t)
	for _, target := range []string{
		"/reservas/consulta", "/reservas/consulta?contacto=300", "/reservas/consulta?reservaId=0&contacto=300",
		"/reservas/consulta?reservaId=x&contacto=300", "/reservas/consulta?reservaId=3", "/reservas/consulta?reservaId=3&contacto=%20",
	} {
		expect(t, runAs("", "GET", target, "", (*ReservaController).Consulta), http.StatusBadRequest)
	}
	if sqlEjecutado(qReservaRel) {
		t.Fatal("no debe consultar la BD con parámetros inválidos")
	}
}

func TestConsultaErrorBD(t *testing.T) {
	nuevoLimitador(100)
	programa(t, conError(qReservaRel))
	expect(t, runAs("", "GET", "/reservas/consulta?reservaId=3&contacto=300", "", (*ReservaController).Consulta), http.StatusInternalServerError)
}

func TestConsultaRateLimitPorIP(t *testing.T) {
	nuevoLimitador(2)
	programa(t, reservaRows(filaReserva(3, "CONFIRMADA", nil)))
	for i := 0; i < 2; i++ {
		expect(t, runAs("", "GET", "/reservas/consulta?reservaId=3&contacto=55", "", (*ReservaController).Consulta), http.StatusOK)
	}
	w := runAs("", "GET", "/reservas/consulta?reservaId=3&contacto=55", "", (*ReservaController).Consulta)
	expect(t, w, http.StatusTooManyRequests)
}

func TestHelpersConsulta(t *testing.T) {
	if contactoCoincide(nil, "300") {
		t.Fatal("contacto nil no coincide")
	}
	if contactoCoincide(&models.ReservaContacto{}, "") {
		t.Fatal("valor vacío no coincide")
	}
	if contactoCoincide(&models.ReservaContacto{}, "300") {
		t.Fatal("contacto sin teléfono ni documentos no coincide")
	}
	if normalizaTelefono("573001234567") != "3001234567" || normalizaTelefono("1234567890") != "1234567890" {
		t.Fatal("normalización de teléfono")
	}
	v := consultaView(&models.Reserva{PK_ID_RESERVA: 4})
	if v.Restaurante != nil || v.EstadoReserva != nil || v.ReservaID != 4 {
		t.Fatalf("vista sin restaurante ni estado: %+v", v)
	}
}
