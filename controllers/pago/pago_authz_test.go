package pago

import (
	"database/sql/driver"
	"net/http"
	"strings"
	"testing"
)

const bodyPagoPost = `{"fechaPago":"2025-02-01","horaPago":"10:00","monto":100,"estadoPago":"%s","metodoPagoId":2}`

func pagoBody(estado string) string { return strings.Replace(bodyPagoPost, "%s", estado, 1) }

func TestSinTokenEs401(t *testing.T) {
	defer resetFake()
	sinToken(t)
	call(t, http.MethodGet, "/pagos", "", func(c *PagoController) { c.GetAll() }, http.StatusUnauthorized)
	call(t, http.MethodGet, "/pagos/search?id=4", "", func(c *PagoController) { c.GetById() }, http.StatusUnauthorized)
	call(t, http.MethodPost, "/pagos", pagoBody("PENDIENTE"), func(c *PagoController) { c.Post() }, http.StatusUnauthorized)
	call(t, http.MethodPut, "/pagos?id=4", `{}`, func(c *PagoController) { c.Put() }, http.StatusUnauthorized)
	call(t, http.MethodDelete, "/pagos?id=4", "", func(c *PagoController) { c.Delete() }, http.StatusUnauthorized)
}

func TestClienteNoPuedeEditarNiBorrarPagos(t *testing.T) {
	defer resetFake()
	como(t, rolClienteT, 77)
	serve(true, true)
	call(t, http.MethodPut, "/pagos?id=4", `{"monto":1}`, func(c *PagoController) { c.Put() }, http.StatusForbidden)
	call(t, http.MethodDelete, "/pagos?id=4", "", func(c *PagoController) { c.Delete() }, http.StatusForbidden)
}

func TestMeseroPuedeEditarYBorrar(t *testing.T) {
	defer resetFake()
	como(t, rolMesero, 5)
	pagoEstado = "PENDIENTE"
	t.Cleanup(func() { pagoEstado = "PAGADO" })
	serve(true, true)
	call(t, http.MethodPut, "/pagos?id=4", `{"monto":1}`, func(c *PagoController) { c.Put() }, http.StatusOK)
	call(t, http.MethodDelete, "/pagos?id=4", "", func(c *PagoController) { c.Delete() }, http.StatusOK)
}

func TestClienteListaSoloSusPagos(t *testing.T) {
	defer resetFake()
	g := func(c *PagoController) { c.GetAll() }
	var gotQ string
	fakeQuery = func(q string, _ []driver.NamedValue) (driver.Rows, error) {
		gotQ = q
		return rowsOf(pagoCols, pagoRow()), nil
	}
	como(t, rolClienteT, 77)
	call(t, http.MethodGet, "/pagos", "", g, http.StatusOK)
	if !strings.Contains(gotQ, "IN (SELECT pk_id_pago FROM pedido WHERE pk_documento_cliente = 77)") {
		t.Fatalf("falta el filtro por dueño: %s", gotQ)
	}
	como(t, rolMesero, 5)
	call(t, http.MethodGet, "/pagos", "", g, http.StatusOK)
	if strings.Contains(gotQ, "pk_documento_cliente") {
		t.Fatalf("el personal ve todos los pagos: %s", gotQ)
	}
	como(t, rolClienteT, 0)
	call(t, http.MethodGet, "/pagos", "", g, http.StatusForbidden)
}

func TestClienteGetByIdSoloSiEsSuyo(t *testing.T) {
	defer resetFake()
	g := func(c *PagoController) { c.GetById() }
	var args []driver.NamedValue
	cuenta := func(n int64) {
		fakeQuery = func(q string, a []driver.NamedValue) (driver.Rows, error) {
			if strings.Contains(q, `FROM "pedido"`) {
				args = a
				return rowsOf([]string{"count"}, []driver.Value{n}), nil
			}
			return rowsOf(pagoCols, pagoRow()), nil
		}
	}
	como(t, rolClienteT, 77)
	cuenta(1)
	call(t, http.MethodGet, "/pagos/search?id=4", "", g, http.StatusOK)
	if len(args) != 2 || args[1].Value != int64(77) {
		t.Fatalf("debe filtrar por el documento del token: %+v", args)
	}
	cuenta(0)
	b := call(t, http.MethodGet, "/pagos/search?id=4", "", g, http.StatusNotFound)
	if !strings.Contains(b, "Pago no encontrado") {
		t.Fatalf("ajeno debe verse como inexistente: %s", b)
	}
	cuenta(1)
	como(t, rolClienteT, 0)
	call(t, http.MethodGet, "/pagos/search?id=4", "", g, http.StatusNotFound)
	como(t, rolClienteT, 77)
	fakeQuery = func(q string, _ []driver.NamedValue) (driver.Rows, error) {
		if strings.Contains(q, `FROM "pedido"`) {
			return nil, errBoom
		}
		return rowsOf(pagoCols, pagoRow()), nil
	}
	call(t, http.MethodGet, "/pagos/search?id=4", "", g, http.StatusInternalServerError)
	// un pago inexistente sigue siendo 404 para el cliente
	serve(false, false)
	call(t, http.MethodGet, "/pagos/search?id=4", "", g, http.StatusNotFound)
}

func TestClienteSoloCreaPagosPendientes(t *testing.T) {
	defer resetFake()
	p := func(c *PagoController) { c.Post() }
	ped := nuevoPedidoFake()
	ped.dueno = 77
	como(t, rolClienteT, 77)
	conPedido := func(estado string) string {
		return strings.Replace(pagoBody(estado), `"metodoPagoId":2`, `"metodoPagoId":2,"pedidoId":10`, 1)
	}
	call(t, http.MethodPost, "/pagos", conPedido("PENDIENTE"), p, http.StatusCreated)
	call(t, http.MethodPost, "/pagos", conPedido("PAGADO"), p, http.StatusForbidden)
	call(t, http.MethodPost, "/pagos", conPedido("NO_PAGO"), p, http.StatusForbidden)
	// sin pedidoId el cliente no puede crear pagos (el monto sale del pedido)
	call(t, http.MethodPost, "/pagos", pagoBody("PENDIENTE"), p, http.StatusBadRequest)
	como(t, rolMesero, 5)
	call(t, http.MethodPost, "/pagos", pagoBody("PAGADO"), p, http.StatusCreated)
}

func TestPutMonto(t *testing.T) {
	defer resetFake()
	u := func(c *PagoController) { c.Put() }
	const url = "/pagos?id=4"

	// PAGADO: cambiar el monto es 409, repetir el mismo monto no es un cambio
	serve(true, true)
	b := call(t, http.MethodPut, url, `{"monto":70}`, u, http.StatusConflict)
	if !strings.Contains(b, "PAGADO") {
		t.Fatalf("mensaje poco claro: %s", b)
	}
	call(t, http.MethodPut, url, `{"monto":50000}`, u, http.StatusOK)
	// otros campos de un PAGADO sí se pueden editar
	call(t, http.MethodPut, url, `{"updatedBy":"yo"}`, u, http.StatusOK)

	// PENDIENTE sin descuentos: se puede
	pagoEstado = "PENDIENTE"
	t.Cleanup(func() { pagoEstado = "PAGADO" })
	call(t, http.MethodPut, url, `{"monto":70}`, u, http.StatusOK)

	// PENDIENTE con descuento aplicado: 409 y no escribe nada
	descuentos = 1
	t.Cleanup(func() { descuentos = 0 })
	var escribio bool
	fakeExec = func(q string, _ []driver.NamedValue) (driver.Result, error) {
		escribio = true
		return fakeResult{}, nil
	}
	b = call(t, http.MethodPut, url, `{"monto":70}`, u, http.StatusConflict)
	if !strings.Contains(b, "descuento aplicado") || escribio {
		t.Fatalf("debe rechazar sin escribir: %s escribió=%v", b, escribio)
	}
	// con descuento, el mismo monto o sin monto no se toca
	call(t, http.MethodPut, url, `{"monto":50000}`, u, http.StatusOK)
	call(t, http.MethodPut, url, `{}`, u, http.StatusOK)
	descuentos = 0

	// error al consultar los descuentos
	fakeQuery = func(q string, _ []driver.NamedValue) (driver.Rows, error) {
		if strings.Contains(q, "pedido_descuento_aplicado") {
			return nil, errBoom
		}
		return rowsOf(pagoCols, pagoRow()), nil
	}
	call(t, http.MethodPut, url, `{"monto":70}`, u, http.StatusInternalServerError)
}

func TestPutTransaccion(t *testing.T) {
	defer resetFake()
	u := func(c *PagoController) { c.Put() }
	serve(true, true)
	fakeBeginErr = errBoom
	call(t, http.MethodPut, "/pagos?id=4", `{}`, u, http.StatusInternalServerError)
	fakeBeginErr = nil
	fakeCommitErr = errBoom
	call(t, http.MethodPut, "/pagos?id=4", `{}`, u, http.StatusInternalServerError)
	fakeCommitErr = nil
	var q string
	fakeQuery = func(s string, _ []driver.NamedValue) (driver.Rows, error) {
		q = s
		return rowsOf(pagoCols, pagoRow()), nil
	}
	call(t, http.MethodPut, "/pagos?id=4", `{}`, u, http.StatusOK)
	if !strings.Contains(q, "FOR UPDATE") {
		t.Fatalf("el pago debe leerse bloqueado: %s", q)
	}
}
