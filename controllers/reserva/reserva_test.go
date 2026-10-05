package reserva

import (
	"database/sql/driver"
	"net/http"
	"strings"
	"testing"
)

const cuerpoNuevoInvitado = `{"documentoContacto":55,"nombreCompleto":"Ana Gómez","telefono":"3001234567","restauranteId":1,"fechaReserva":"2025-01-31","horaReserva":"18:30:00","personas":4,"indicaciones":"ventana","createdBy":"admin@x.co"}`

// ---------- GET /reservas ----------

func TestGetAllConRelacionesSinPassword(t *testing.T) {
	programa(t, reservaRows(filaReserva(1, "PENDIENTE", int64(9))))
	w := run("GET", "/reservas", "", (*ReservaController).GetAll)
	r := expect(t, w, http.StatusOK)
	l := dataList(t, r)
	if len(l) != 1 {
		t.Fatalf("esperaba 1 reserva: %v", l)
	}
	rv := l[0].(map[string]any)
	contacto := rv["contactoId"].(map[string]any)
	if contacto["nombreCompleto"] != "Ana Gómez" || contacto["telefono"] != "3001234567" {
		t.Fatalf("contacto sin cargar: %v", contacto)
	}
	if contacto["documentoCliente"].(map[string]any)["documentoCliente"] != float64(9) {
		t.Fatalf("documentoCliente inesperado: %v", contacto)
	}
	if rv["restauranteId"].(map[string]any)["nombreRestaurante"] != "Sazón Criolla" {
		t.Fatalf("restaurante sin cargar: %v", rv)
	}
	if rv["fechaReserva"] != "31-01-2025" {
		t.Fatalf("fecha de respuesta debe ser DD-MM-YYYY: %v", rv["fechaReserva"])
	}
	if !sqlEjecutado(`INNER JOIN "reserva_contacto"`) || !sqlEjecutado(`INNER JOIN "restaurante"`) {
		t.Fatalf("la consulta debe cargar contacto y restaurante: %v", recorded)
	}
	sinPassword(t, w)
}

func TestGetAllVacioEsLista(t *testing.T) {
	programa(t)
	w := run("GET", "/reservas", "", (*ReservaController).GetAll)
	expect(t, w, http.StatusOK)
	if !strings.Contains(w.Body.String(), `"data":[]`) {
		t.Fatalf("data debe ser []: %s", w.Body.String())
	}
}

func TestGetAllError(t *testing.T) {
	programa(t, conError(qReservaRel))
	expect(t, run("GET", "/reservas", "", (*ReservaController).GetAll), http.StatusInternalServerError)
}

// ---------- GET /reservas/search ----------

func TestGetByIdOK(t *testing.T) {
	programa(t, reservaRows(filaReserva(3, "CONFIRMADA", nil)))
	w := run("GET", "/reservas/search?id=3", "", (*ReservaController).GetById)
	r := expect(t, w, http.StatusOK)
	m := dataMap(t, r)
	if m["reservaId"] != float64(3) || m["estadoReserva"] != "CONFIRMADA" {
		t.Fatalf("reserva inesperada: %v", m)
	}
	if _, ok := m["contactoId"].(map[string]any)["documentoCliente"]; ok {
		t.Fatalf("invitado no debe traer documentoCliente: %v", m)
	}
	if !sqlEjecutado(`INNER JOIN "reserva_contacto"`) {
		t.Fatalf("search debe cargar relaciones: %v", recorded)
	}
}

func TestGetByIdParametrosInvalidos(t *testing.T) {
	programa(t)
	for _, target := range []string{"/reservas/search", "/reservas/search?id=0", "/reservas/search?id=abc", "/reservas/search?id=-2"} {
		expect(t, run("GET", target, "", (*ReservaController).GetById), http.StatusBadRequest)
	}
}

func TestGetByIdNoEncontradaEs404(t *testing.T) {
	programa(t)
	w := run("GET", "/reservas/search?id=99", "", (*ReservaController).GetById)
	r := expect(t, w, http.StatusNotFound)
	if r.Message != "Reserva no encontrada" {
		t.Fatalf("mensaje: %s", r.Message)
	}
}

func TestGetByIdError(t *testing.T) {
	programa(t, conError(qReservaRel))
	expect(t, run("GET", "/reservas/search?id=1", "", (*ReservaController).GetById), http.StatusInternalServerError)
}

// ---------- GET /reservas/parameter ----------

func TestGetByParameterConRelacionesYFiltros(t *testing.T) {
	programa(t, reservaRows(filaReserva(1, "PENDIENTE", nil)))
	w := run("GET", "/reservas/parameter?contactoId=2&fecha=2025-01-31", "", (*ReservaController).GetByParameter)
	r := expect(t, w, http.StatusOK)
	rv := dataList(t, r)[0].(map[string]any)
	if rv["contactoId"].(map[string]any)["nombreCompleto"] != "Ana Gómez" {
		t.Fatalf("contacto sin cargar: %v", rv)
	}
	q := sqlQue(qReservaRel)
	if !strings.Contains(q, `INNER JOIN "reserva_contacto"`) || !strings.Contains(q, `INNER JOIN "restaurante"`) {
		t.Fatalf("parameter debe cargar relaciones: %s", q)
	}
	if !strings.Contains(q, `T0."pk_id_contacto" = $1`) || !strings.Contains(q, `T0."fecha" = $2`) {
		t.Fatalf("faltan filtros: %s", q)
	}
	sinPassword(t, w)
}

func TestGetByParameterSinFiltrosVacioEsLista(t *testing.T) {
	programa(t)
	w := run("GET", "/reservas/parameter", "", (*ReservaController).GetByParameter)
	r := expect(t, w, http.StatusOK)
	if r.Message != "No se encontraron reservas" || !strings.Contains(w.Body.String(), `"data":[]`) {
		t.Fatalf("respuesta inesperada: %s", w.Body.String())
	}
}

func TestGetByParameterInvalidos(t *testing.T) {
	programa(t)
	for _, target := range []string{
		"/reservas/parameter?contactoId=abc",
		"/reservas/parameter?contactoId=0",
		"/reservas/parameter?fecha=31-01-2025",
	} {
		expect(t, run("GET", target, "", (*ReservaController).GetByParameter), http.StatusBadRequest)
	}
}

func TestGetByParameterError(t *testing.T) {
	programa(t, conError(qReservaRel))
	expect(t, run("GET", "/reservas/parameter", "", (*ReservaController).GetByParameter), http.StatusInternalServerError)
}

// ---------- GET /reservas/documento ----------

func TestGetByDocumentoComoCliente(t *testing.T) {
	programa(t, reservaRows(filaReserva(1, "PENDIENTE", int64(9))))
	r := expect(t, run("GET", "/reservas/documento?documento=9&fecha=2025-01-31", "", (*ReservaController).GetByDocumento), http.StatusOK)
	if len(dataList(t, r)) != 1 || r.Message != "Reservas obtenidas exitosamente" {
		t.Fatalf("respuesta inesperada: %+v", r)
	}
}

func TestGetByDocumentoCaeAContacto(t *testing.T) {
	programa(t,
		&res{match: `T3."pk_documento_cliente" = $1`, cols: 21},
		reservaRows(filaReserva(1, "PENDIENTE", nil)),
	)
	r := expect(t, run("GET", "/reservas/documento?documento=55", "", (*ReservaController).GetByDocumento), http.StatusOK)
	if len(dataList(t, r)) != 1 {
		t.Fatalf("esperaba 1 reserva por contacto: %+v", r)
	}
}

func TestGetByDocumentoVacio(t *testing.T) {
	programa(t)
	w := run("GET", "/reservas/documento?documento=55", "", (*ReservaController).GetByDocumento)
	r := expect(t, w, http.StatusOK)
	if r.Message != "No se encontraron reservas para este documento" || !strings.Contains(w.Body.String(), `"data":[]`) {
		t.Fatalf("respuesta inesperada: %s", w.Body.String())
	}
}

func TestGetByDocumentoInvalidos(t *testing.T) {
	programa(t)
	for _, target := range []string{"/reservas/documento", "/reservas/documento?documento=0", "/reservas/documento?documento=1&fecha=hoy"} {
		expect(t, run("GET", target, "", (*ReservaController).GetByDocumento), http.StatusBadRequest)
	}
}

func TestGetByDocumentoErrores(t *testing.T) {
	programa(t, conError(`T3."pk_documento_cliente"`))
	expect(t, run("GET", "/reservas/documento?documento=1", "", (*ReservaController).GetByDocumento), http.StatusInternalServerError)

	programa(t, &res{match: `T3."pk_documento_cliente"`, cols: 21}, conError(`T1."documento_contacto" = $1`))
	expect(t, run("GET", "/reservas/documento?documento=1", "", (*ReservaController).GetByDocumento), http.StatusInternalServerError)
}

// ---------- GET /reservas/cliente ----------

func TestGetByDocumentoClienteOK(t *testing.T) {
	programa(t, reservaRows(filaReserva(1, "PENDIENTE", int64(9))))
	w := run("GET", "/reservas/cliente?documentoCliente=9&fecha=2025-01-31", "", (*ReservaController).GetByDocumentoCliente)
	r := expect(t, w, http.StatusOK)
	if r.Message != "Reservas del cliente obtenidas exitosamente" || len(dataList(t, r)) != 1 {
		t.Fatalf("respuesta inesperada: %+v", r)
	}
	sinPassword(t, w)
}

func TestGetByDocumentoClienteVacio(t *testing.T) {
	programa(t)
	w := run("GET", "/reservas/cliente?documentoCliente=9", "", (*ReservaController).GetByDocumentoCliente)
	r := expect(t, w, http.StatusOK)
	if r.Message != "No se encontraron reservas para este cliente" || !strings.Contains(w.Body.String(), `"data":[]`) {
		t.Fatalf("respuesta inesperada: %s", w.Body.String())
	}
}

func TestGetByDocumentoClienteInvalidosYError(t *testing.T) {
	programa(t)
	for _, target := range []string{"/reservas/cliente", "/reservas/cliente?documentoCliente=x", "/reservas/cliente?documentoCliente=1&fecha=bad"} {
		expect(t, run("GET", target, "", (*ReservaController).GetByDocumentoCliente), http.StatusBadRequest)
	}
	programa(t, conError(qReservaRel))
	expect(t, run("GET", "/reservas/cliente?documentoCliente=1", "", (*ReservaController).GetByDocumentoCliente), http.StatusInternalServerError)
}

// ---------- POST /reservas ----------

// postOK agrega las reglas de un POST exitoso: restaurante y relectura final.
func postOK(extra ...*res) []*res {
	return append(extra, restauranteRows(), reservaRows(filaReserva(7, "PENDIENTE", nil)))
}

func TestPostInvitadoNuevo(t *testing.T) {
	programa(t, postOK(vacio(qContacto, 5))...)
	w := run("POST", "/reservas", cuerpoNuevoInvitado, (*ReservaController).Post)
	r := expect(t, w, http.StatusCreated)
	if dataMap(t, r)["reservaId"] != float64(7) {
		t.Fatalf("data inesperada: %v", r.Data)
	}
	if !sqlEjecutado(qInsertContac) || !sqlEjecutado(qInsertReserv) {
		t.Fatalf("debe insertar contacto y reserva: %v", recorded)
	}
	sinPassword(t, w)
}

func TestPostInvitadoExistenteNoCreaContacto(t *testing.T) {
	programa(t, postOK(contactoRows(filaContacto(2, int64(55), nil)))...)
	expect(t, run("POST", "/reservas", cuerpoNuevoInvitado, (*ReservaController).Post), http.StatusCreated)
	if sqlEjecutado(qInsertContac) {
		t.Fatalf("no debe crear contacto existente: %v", recorded)
	}
}

func TestPostClienteNuevoNoFiltraPassword(t *testing.T) {
	body := `{"documentoCliente":9,"restauranteId":1,"fechaReserva":"2025-01-31","horaReserva":"18:30:00","personas":2,"estadoReserva":"CONFIRMADA"}`
	programa(t, vacio(qContacto, 5), clienteRows(), restauranteRows(), reservaRows(filaReserva(7, "CONFIRMADA", int64(9))))
	w := run("POST", "/reservas", body, (*ReservaController).Post)
	r := expect(t, w, http.StatusCreated)
	c := dataMap(t, r)["contactoId"].(map[string]any)
	ref := c["documentoCliente"].(map[string]any)
	if ref["documentoCliente"] != float64(9) || len(ref) != 1 {
		t.Fatalf("documentoCliente solo debe traer el documento: %v", c)
	}
	if !sqlEjecutado(qInsertContac) {
		t.Fatalf("debe crear el contacto del cliente: %v", recorded)
	}
	sinPassword(t, w)
}

func TestPostClienteSinTelefonoYContactoExistente(t *testing.T) {
	body := `{"documentoCliente":9,"restauranteId":1,"fechaReserva":"2025-01-31","horaReserva":"18:30:00","personas":2}`
	sinTel := &res{match: qCliente, cols: 8, rows: [][]driver.Value{{int64(9), "Luis", "Mora", "l@x.co", "Calle 1", "", nil, "secreto-hash"}}}
	programa(t, vacio(qContacto, 5), sinTel, restauranteRows(), reservaRows(filaReserva(7, "PENDIENTE", int64(9))))
	expect(t, run("POST", "/reservas", body, (*ReservaController).Post), http.StatusCreated)

	programa(t, contactoRows(filaContacto(2, nil, int64(9))), restauranteRows(), reservaRows(filaReserva(7, "PENDIENTE", int64(9))))
	expect(t, run("POST", "/reservas", body, (*ReservaController).Post), http.StatusCreated)
	if sqlEjecutado(qInsertContac) || sqlEjecutado(qCliente) {
		t.Fatalf("con contacto existente no debe leer cliente ni insertar: %v", recorded)
	}
}

func TestPostPrevaleceDocumentoContacto(t *testing.T) {
	programa(t, postOK(contactoRows(filaContacto(2, int64(55), nil)))...)
	body := `{"documentoContacto":55,"documentoCliente":9,"restauranteId":1,"fechaReserva":"2025-01-31","horaReserva":"18:30:00","personas":2}`
	expect(t, run("POST", "/reservas", body, (*ReservaController).Post), http.StatusCreated)
	if sqlEjecutado(qCliente) {
		t.Fatalf("no debe consultar cliente: %v", recorded)
	}
}

func TestPostValidaciones400(t *testing.T) {
	casos := map[string]string{
		"json roto":        `{`,
		"vacío":            ``,
		"tipo erróneo":     `{"personas":"muchas"}`,
		"sin fecha":        `{"documentoContacto":55,"horaReserva":"18:30:00","personas":2,"restauranteId":1}`,
		"fecha mala":       `{"documentoContacto":55,"fechaReserva":"31-01-2025","horaReserva":"18:30:00","personas":2,"restauranteId":1}`,
		"fecha vacía":      `{"fechaReserva":"","documentoContacto":55,"horaReserva":"18:30:00","personas":2,"restauranteId":1}`,
		"sin hora":         `{"documentoContacto":55,"fechaReserva":"2025-01-31","personas":2,"restauranteId":1}`,
		"hora mala":        `{"documentoContacto":55,"fechaReserva":"2025-01-31","horaReserva":"6pm","personas":2,"restauranteId":1}`,
		"hora vacía":       `{"fechaReserva":"2025-01-31","horaReserva":"","documentoContacto":55,"personas":2,"restauranteId":1}`,
		"sin personas":     `{"documentoContacto":55,"fechaReserva":"2025-01-31","horaReserva":"18:30:00","restauranteId":1}`,
		"personas cero":    `{"documentoContacto":55,"fechaReserva":"2025-01-31","horaReserva":"18:30:00","personas":0,"restauranteId":1}`,
		"estado inválido":  `{"documentoContacto":55,"nombreCompleto":"Ana","restauranteId":1,"fechaReserva":"2025-01-31","horaReserva":"18:30:00","personas":4,"estadoReserva":"OTRO"}`,
		"sin restaurante":  `{"documentoContacto":55,"fechaReserva":"2025-01-31","horaReserva":"18:30:00","personas":2}`,
		"restaurante cero": `{"documentoContacto":55,"restauranteId":0,"fechaReserva":"2025-01-31","horaReserva":"18:30:00","personas":2}`,
		"sin contacto":     `{"restauranteId":1,"fechaReserva":"2025-01-31","horaReserva":"18:30:00","personas":2}`,
		"documento cero":   `{"documentoContacto":0,"nombreCompleto":"A","restauranteId":1,"fechaReserva":"2025-01-31","horaReserva":"18:30:00","personas":2}`,
		"cliente negativo": `{"documentoCliente":-3,"restauranteId":1,"fechaReserva":"2025-01-31","horaReserva":"18:30:00","personas":2}`,
		"sin nombre":       `{"documentoContacto":55,"restauranteId":1,"fechaReserva":"2025-01-31","horaReserva":"18:30:00","personas":2}`,
		"nombre en blanco": `{"documentoContacto":55,"nombreCompleto":"  ","restauranteId":1,"fechaReserva":"2025-01-31","horaReserva":"18:30:00","personas":2}`,
	}
	for nombre, body := range casos {
		programa(t, restauranteRows(), vacio(qContacto, 5))
		w := run("POST", "/reservas", body, (*ReservaController).Post)
		if w.Code != http.StatusBadRequest {
			t.Errorf("%s: esperaba 400, obtuve %d: %s", nombre, w.Code, w.Body.String())
			continue
		}
		decode(t, w)
		if sqlEjecutado(qInsertReserv) {
			t.Errorf("%s: no debe insertar la reserva", nombre)
		}
	}
}

func TestPostValidaAntesDeTocarLaBase(t *testing.T) {
	programa(t, restauranteRows(), vacio(qContacto, 5))
	body := `{"documentoContacto":55,"nombreCompleto":"Ana","restauranteId":1,"fechaReserva":"mala","horaReserva":"18:30:00","personas":2}`
	expect(t, run("POST", "/reservas", body, (*ReservaController).Post), http.StatusBadRequest)
	if len(recorded) != 0 {
		t.Fatalf("una validación fallida no debe tocar la base: %v", recorded)
	}
}

func TestPostNoEncontrados404(t *testing.T) {
	programa(t, vacio(qRestaurante, 4))
	expect(t, run("POST", "/reservas", cuerpoNuevoInvitado, (*ReservaController).Post), http.StatusNotFound)

	programa(t, restauranteRows(), vacio(qContacto, 5), vacio(qCliente, 8))
	body := `{"documentoCliente":9,"restauranteId":1,"fechaReserva":"2025-01-31","horaReserva":"18:30:00","personas":2}`
	r := expect(t, run("POST", "/reservas", body, (*ReservaController).Post), http.StatusNotFound)
	if r.Message != "Cliente no encontrado" {
		t.Fatalf("mensaje: %s", r.Message)
	}
}

func TestPostErrores500(t *testing.T) {
	cliente := `{"documentoCliente":9,"restauranteId":1,"fechaReserva":"2025-01-31","horaReserva":"18:30:00","personas":2}`
	casos := []struct {
		nombre string
		body   string
		reglas []*res
	}{
		{"restaurante", cuerpoNuevoInvitado, []*res{conError(qRestaurante)}},
		{"consulta contacto invitado", cuerpoNuevoInvitado, []*res{restauranteRows(), conError(qContacto)}},
		{"insert contacto", cuerpoNuevoInvitado, []*res{restauranteRows(), vacio(qContacto, 5), conError(qInsertContac)}},
		{"insert reserva", cuerpoNuevoInvitado, []*res{restauranteRows(), contactoRows(filaContacto(2, int64(55), nil)), conError(qInsertReserv)}},
		{"relectura", cuerpoNuevoInvitado, []*res{restauranteRows(), contactoRows(filaContacto(2, int64(55), nil)), conError(qReservaRel)}},
		{"consulta contacto cliente", cliente, []*res{restauranteRows(), conError(qContacto)}},
		{"lectura cliente", cliente, []*res{restauranteRows(), vacio(qContacto, 5), conError(qCliente)}},
		{"insert contacto cliente", cliente, []*res{restauranteRows(), vacio(qContacto, 5), clienteRows(), conError(qInsertContac)}},
	}
	for _, c := range casos {
		programa(t, c.reglas...)
		w := run("POST", "/reservas", c.body, (*ReservaController).Post)
		if w.Code != http.StatusInternalServerError {
			t.Errorf("%s: esperaba 500, obtuve %d: %s", c.nombre, w.Code, w.Body.String())
			continue
		}
		decode(t, w)
	}
}

// ---------- PUT /reservas ----------

func putReglas(extra ...*res) []*res {
	return append(extra, reservaRows(filaReserva(3, "PENDIENTE", nil)))
}

func TestPutMergeSoloActualizaLoEnviado(t *testing.T) {
	programa(t, putReglas()...)
	w := run("PUT", "/reservas?id=3", `{"personas":6}`, (*ReservaController).Put)
	r := expect(t, w, http.StatusOK)
	if dataMap(t, r)["reservaId"] != float64(3) {
		t.Fatalf("data inesperada: %v", r.Data)
	}
	up := execQue(qUpdate)
	if up == nil {
		t.Fatalf("no hubo UPDATE: %v", recorded)
	}
	if !strings.Contains(up.q, `"personas" = $`) || !strings.Contains(up.q, `"updated_at"`) {
		t.Fatalf("UPDATE inesperado: %s", up.q)
	}
	for _, no := range []string{`"fecha"`, `"hora"`, `"estado_reserva"`, `"indicaciones"`, `"pk_id_contacto"`, `"pk_id_restaurante"`, `"created_at"`} {
		if strings.Contains(up.q, no) {
			t.Fatalf("el campo ausente %s no debe actualizarse: %s", no, up.q)
		}
	}
}

func TestPutCuerpoVacioSoloUpdatedAt(t *testing.T) {
	programa(t, putReglas()...)
	expect(t, run("PUT", "/reservas?id=3", `{}`, (*ReservaController).Put), http.StatusOK)
	up := execQue(qUpdate)
	if up == nil || len(up.args) != 2 { // updated_at + clave primaria
		t.Fatalf("UPDATE inesperado: %v", up)
	}
}

func TestPutTodosLosCampos(t *testing.T) {
	programa(t, putReglas(restauranteRows(), contactoRows(filaContacto(2, int64(55), nil)))...)
	body := `{"documentoContacto":55,"restauranteId":1,"fechaReserva":"2025-02-01","horaReserva":"19:00:00","personas":5,"estadoReserva":"CONFIRMADA","indicaciones":"fondo","updatedBy":"op@x.co"}`
	expect(t, run("PUT", "/reservas?id=3", body, (*ReservaController).Put), http.StatusOK)
	up := execQue(qUpdate)
	for _, col := range []string{`"fecha"`, `"hora"`, `"personas"`, `"estado_reserva"`, `"indicaciones"`, `"updated_by"`, `"pk_id_restaurante"`, `"pk_id_contacto"`, `"updated_at"`} {
		if up == nil || !strings.Contains(up.q, col) {
			t.Fatalf("falta %s en UPDATE: %v", col, up)
		}
	}
}

func TestPutNullLimpiaSoloAnulables(t *testing.T) {
	programa(t, putReglas()...)
	expect(t, run("PUT", "/reservas?id=3", `{"indicaciones":null,"updatedBy":null}`, (*ReservaController).Put), http.StatusOK)
	up := execQue(qUpdate)
	if !strings.Contains(up.q, `"indicaciones"`) || !strings.Contains(up.q, `"updated_by"`) {
		t.Fatalf("null debe limpiar indicaciones y updatedBy: %s", up.q)
	}

	for _, campo := range []string{"fechaReserva", "horaReserva", "personas", "estadoReserva", "restauranteId", "documentoContacto", "documentoCliente"} {
		programa(t, putReglas()...)
		w := run("PUT", "/reservas?id=3", `{"`+campo+`":null}`, (*ReservaController).Put)
		if w.Code != http.StatusBadRequest {
			t.Errorf("%s:null debe ser 400, obtuve %d", campo, w.Code)
		}
		if sqlEjecutado(qUpdate) {
			t.Errorf("%s:null no debe actualizar", campo)
		}
	}
}

func TestPutValidaciones400(t *testing.T) {
	casos := map[string]string{
		"json roto":            `{`,
		"vacío":                ``,
		"tipo erróneo":         `{"personas":"x"}`,
		"fecha":                `{"fechaReserva":"01/02/2025"}`,
		"fecha vacía":          `{"fechaReserva":""}`,
		"hora":                 `{"horaReserva":"7pm"}`,
		"personas cero":        `{"personas":0}`,
		"estado":               `{"estadoReserva":"X"}`,
		"estado vacío":         `{"estadoReserva":""}`,
		"restaurante cero":     `{"restauranteId":0}`,
		"contacto doc cero":    `{"documentoContacto":0}`,
		"cliente doc negativo": `{"documentoCliente":-1}`,
		"contacto sin nombre":  `{"documentoContacto":77}`,
	}
	for nombre, body := range casos {
		programa(t, putReglas(vacio(qContacto, 5))...)
		w := run("PUT", "/reservas?id=3", body, (*ReservaController).Put)
		if w.Code != http.StatusBadRequest {
			t.Errorf("%s: esperaba 400, obtuve %d: %s", nombre, w.Code, w.Body.String())
			continue
		}
		decode(t, w)
		if sqlEjecutado(qUpdate) {
			t.Errorf("%s: no debe actualizar", nombre)
		}
	}
}

func TestPutIdInvalido(t *testing.T) {
	programa(t)
	for _, target := range []string{"/reservas", "/reservas?id=0", "/reservas?id=zz"} {
		expect(t, run("PUT", target, `{}`, (*ReservaController).Put), http.StatusBadRequest)
	}
}

func TestPutNoEncontrados404(t *testing.T) {
	programa(t)
	expect(t, run("PUT", "/reservas?id=3", `{}`, (*ReservaController).Put), http.StatusNotFound)

	programa(t, putReglas(vacio(qRestaurante, 4))...)
	expect(t, run("PUT", "/reservas?id=3", `{"restauranteId":9}`, (*ReservaController).Put), http.StatusNotFound)

	programa(t, putReglas(vacio(qContacto, 5), vacio(qCliente, 8))...)
	expect(t, run("PUT", "/reservas?id=3", `{"documentoCliente":9}`, (*ReservaController).Put), http.StatusNotFound)
}

func TestPutCambiaContactoCliente(t *testing.T) {
	programa(t, putReglas(contactoRows(filaContacto(2, nil, int64(9))))...)
	expect(t, run("PUT", "/reservas?id=3", `{"documentoCliente":9}`, (*ReservaController).Put), http.StatusOK)
	if !strings.Contains(execQue(qUpdate).q, `"pk_id_contacto"`) {
		t.Fatal("debe actualizar el contacto")
	}
}

func TestPutErrores500(t *testing.T) {
	casos := []struct {
		nombre string
		body   string
		reglas []*res
	}{
		{"carga", `{}`, []*res{conError(qReservaRel)}},
		{"restaurante", `{"restauranteId":1}`, putReglas(conError(qRestaurante))},
		{"contacto", `{"documentoContacto":55}`, putReglas(conError(qContacto))},
		{"update", `{}`, putReglas(conError(qUpdate))},
		{"relectura", `{}`, []*res{
			{match: qReservaRel, cols: 21, rows: [][]driver.Value{filaReserva(3, "PENDIENTE", nil)}, onlyCall: 1},
			{match: qReservaRel, err: errDB, onlyCall: 1},
		}},
	}
	for _, c := range casos {
		programa(t, c.reglas...)
		w := run("PUT", "/reservas?id=3", c.body, (*ReservaController).Put)
		if w.Code != http.StatusInternalServerError {
			t.Errorf("%s: esperaba 500, obtuve %d: %s", c.nombre, w.Code, w.Body.String())
			continue
		}
		decode(t, w)
	}
}

// ---------- DELETE /reservas (cancelar) ----------

func TestDeleteCancelaUsandoColumnasValidas(t *testing.T) {
	programa(t, reservaRows(filaReserva(3, "PENDIENTE", nil)))
	w := run("DELETE", "/reservas?id=3", "", (*ReservaController).Delete)
	r := expect(t, w, http.StatusOK)
	if dataMap(t, r)["estadoReserva"] != "CANCELADA" {
		t.Fatalf("estado inesperado: %v", r.Data)
	}
	up := execQue(qUpdate)
	if up == nil || !strings.Contains(up.q, `"estado_reserva" = $1`) || !strings.Contains(up.q, `"updated_at" = $2`) {
		t.Fatalf("UPDATE inesperado: %v", up)
	}
	if up.args[0].Value != "CANCELADA" {
		t.Fatalf("estado enviado: %v", up.args)
	}
	sinPassword(t, w)
}

func TestDeleteSinEstadoPrevioTambienCancela(t *testing.T) {
	programa(t, reservaRows(filaReserva(3, nil, nil)))
	expect(t, run("DELETE", "/reservas?id=3", "", (*ReservaController).Delete), http.StatusOK)
}

func TestDeleteYaCanceladaEs409(t *testing.T) {
	programa(t, reservaRows(filaReserva(3, "CANCELADA", nil)))
	expect(t, run("DELETE", "/reservas?id=3", "", (*ReservaController).Delete), http.StatusConflict)
	if sqlEjecutado(qUpdate) {
		t.Fatal("no debe actualizar")
	}
}

func TestDeleteInvalidoNoEncontradoYErrores(t *testing.T) {
	programa(t)
	for _, target := range []string{"/reservas", "/reservas?id=0", "/reservas?id=x"} {
		expect(t, run("DELETE", target, "", (*ReservaController).Delete), http.StatusBadRequest)
	}
	expect(t, run("DELETE", "/reservas?id=3", "", (*ReservaController).Delete), http.StatusNotFound)

	programa(t, conError(qReservaRel))
	expect(t, run("DELETE", "/reservas?id=3", "", (*ReservaController).Delete), http.StatusInternalServerError)

	programa(t, reservaRows(filaReserva(3, "PENDIENTE", nil)), conError(qUpdate))
	expect(t, run("DELETE", "/reservas?id=3", "", (*ReservaController).Delete), http.StatusInternalServerError)
}
