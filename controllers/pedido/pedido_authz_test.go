package pedido

import (
	"database/sql/driver"
	"net/http"
	"strings"
	"testing"
	"time"
)

const cliente1001 = int64(1001) // dueño del pedido que devuelve pedidoRow

func ajeno(t *testing.T) { como(t, rolClienteT, 55) }
func dueno(t *testing.T) { como(t, rolClienteT, cliente1001) }

// pedidoSinPagoRow es el pedido 10 del cliente 1001 sin pago asignado.
func pedidoSinPagoRow() []driver.Value {
	return []driver.Value{int64(10), fechaPedido, time.Date(2000, 1, 1, 18, 30, 0, 0, time.UTC), false, "INICIADO",
		nil, nil, int64(1), cliente1001, updatedPedido, "cajero"}
}

func pedidoSinPago() rt {
	return rt{`FROM "pedido"`, func() driver.Rows { return rowsOf(pedidoCols, pedidoSinPagoRow()) }}
}

func TestSinTokenEs401(t *testing.T) {
	defer resetFake()
	sinToken(t)
	serve(pedidoOK())
	call(t, http.MethodGet, "/pedidos", "", func(c *PedidoController) { c.GetAll() }, http.StatusUnauthorized)
	call(t, http.MethodPost, "/pedidos", `{}`, func(c *PedidoController) { c.Post() }, http.StatusUnauthorized)
	call(t, http.MethodPost, "/pedidos/asignar-domicilio?pedido_id=10&domicilio_id=3", "", func(c *PedidoController) { c.AssignDomicilio() }, http.StatusUnauthorized)
	call(t, http.MethodPost, "/pedidos/asignar-pago?pedido_id=10&pago_id=4", "", func(c *PedidoController) { c.AssignPago() }, http.StatusUnauthorized)
	call(t, http.MethodPut, "/pedidos/actualizar-estado?pedido_id=10&estado=LISTO", "", func(c *PedidoController) { c.UpdateEstadoPedido() }, http.StatusUnauthorized)
	call(t, http.MethodGet, "/pedidos/detalles?pedido_id=10", "", func(c *PedidoController) { c.GetPedidoDetails() }, http.StatusUnauthorized)
}

func TestGetAllCliente(t *testing.T) {
	defer resetFake()
	g := func(c *PedidoController) { c.GetAll() }
	var gotQ string
	var gotArgs []driver.NamedValue
	fakeQuery = func(q string, a []driver.NamedValue) (driver.Rows, error) {
		gotQ, gotArgs = q, a
		return rowsOf(pedidoCols, pedidoRow()), nil
	}
	como(t, rolClienteT, 77)
	// sin parámetro: se fuerza el documento del token
	call(t, http.MethodGet, "/pedidos", "", g, http.StatusOK)
	if !strings.Contains(gotQ, "p.pk_documento_cliente = $1") || len(gotArgs) != 1 || gotArgs[0].Value != int64(77) {
		t.Fatalf("debe filtrar por el cliente del token: %s %+v", gotQ, gotArgs)
	}
	// su propio documento explícito
	call(t, http.MethodGet, "/pedidos?cliente=77", "", g, http.StatusOK)
	// el de otro
	call(t, http.MethodGet, "/pedidos?cliente=78", "", g, http.StatusForbidden)
	// token de cliente sin documento
	como(t, rolClienteT, 0)
	call(t, http.MethodGet, "/pedidos", "", g, http.StatusForbidden)
	// el personal no se ve limitado
	como(t, rolMesero, 5)
	call(t, http.MethodGet, "/pedidos", "", g, http.StatusOK)
	if strings.Contains(gotQ, "pk_documento_cliente") {
		t.Fatalf("el personal ve todos los pedidos: %s", gotQ)
	}
}

func TestPostCliente(t *testing.T) {
	defer resetFake()
	g := grabar(t)
	p := func(c *PedidoController) { c.Post() }
	serve(count("cliente", 1), count("restaurante", 1), count("domicilio", 1))
	var insertArgs []driver.NamedValue
	fakeExec = func(q string, a []driver.NamedValue) (driver.Result, error) {
		if strings.HasPrefix(q, `INSERT INTO "pedido"`) {
			insertArgs = a
		}
		return fakeResult{}, nil
	}
	como(t, rolClienteT, cliente1001)

	// sin documento en el body: sale del token
	call(t, http.MethodPost, "/pedidos", `{"restauranteId":1}`, p, http.StatusCreated)
	if ev := unico(t, g); ev.Cliente != cliente1001 {
		t.Fatalf("el pedido debe ser del cliente del token: %+v", ev)
	}
	var tieneDoc bool
	for _, a := range insertArgs {
		if a.Value == cliente1001 {
			tieneDoc = true
		}
	}
	if !tieneDoc {
		t.Fatalf("el INSERT debe llevar el documento del token: %+v", insertArgs)
	}
	// el suyo explícito
	call(t, http.MethodPost, "/pedidos", `{"documentoCliente":1001}`, p, http.StatusCreated)
	// el de otro es 403 y no inserta ni notifica
	g.eventos, insertArgs = nil, nil
	body := call(t, http.MethodPost, "/pedidos", `{"documentoCliente":1002}`, p, http.StatusForbidden)
	if !strings.Contains(body, "otro cliente") || insertArgs != nil || len(g.eventos) != 0 {
		t.Fatalf("no debe crear a nombre de otro: %s %v %v", body, insertArgs, g.eventos)
	}
	// token de cliente sin documento
	como(t, rolClienteT, 0)
	call(t, http.MethodPost, "/pedidos", `{}`, p, http.StatusForbidden)
	// el cliente del token ya no existe
	como(t, rolClienteT, cliente1001)
	serve(count("cliente", 0))
	call(t, http.MethodPost, "/pedidos", `{}`, p, http.StatusNotFound)
}

func TestPostPersonal(t *testing.T) {
	defer resetFake()
	g := grabar(t)
	p := func(c *PedidoController) { c.Post() }
	como(t, rolMesero, 5)
	// a nombre de un cliente existente: se notifica al cliente
	serve(count("cliente", 1))
	call(t, http.MethodPost, "/pedidos", `{"documentoCliente":1001}`, p, http.StatusCreated)
	if ev := unico(t, g); ev.Cliente != cliente1001 {
		t.Fatalf("el cliente indicado debe recibir el aviso: %+v", ev)
	}
	// cliente inexistente
	serve(count("cliente", 0))
	call(t, http.MethodPost, "/pedidos", `{"documentoCliente":1001}`, p, http.StatusNotFound)
	// pedido de mostrador sin cliente
	g.eventos = nil
	call(t, http.MethodPost, "/pedidos", `{}`, p, http.StatusCreated)
	if ev := unico(t, g); ev.Cliente != 0 {
		t.Fatalf("sin cliente no hay aviso al cliente: %+v", ev)
	}
}

func TestDetallesCliente(t *testing.T) {
	defer resetFake()
	d := func(c *PedidoController) { c.GetPedidoDetails() }
	fakeQuery = func(string, []driver.NamedValue) (driver.Rows, error) {
		return rowsOf(detallesCols, []driver.Value{int64(10), "31-01-2025", "18:30:00", false, "INICIADO", "NEQUI", "[]", int64(4), int64(1), int64(0), cliente1001}), nil
	}
	const url = "/pedidos/detalles?pedido_id=10"
	dueno(t)
	call(t, http.MethodGet, url, "", d, http.StatusOK)
	ajeno(t)
	if b := call(t, http.MethodGet, url, "", d, http.StatusNotFound); !strings.Contains(b, "Pedido no encontrado") {
		t.Fatalf("ajeno debe verse como inexistente: %s", b)
	}
	como(t, rolClienteT, 0)
	call(t, http.MethodGet, url, "", d, http.StatusNotFound)
	como(t, rolMesero, 5)
	call(t, http.MethodGet, url, "", d, http.StatusOK)
	// un pedido sin cliente solo lo ve el personal
	fakeQuery = func(string, []driver.NamedValue) (driver.Rows, error) {
		return rowsOf(detallesCols, []driver.Value{int64(10), "31-01-2025", "18:30:00", false, "INICIADO", "", "[]", int64(0), int64(0), int64(0), int64(0)}), nil
	}
	call(t, http.MethodGet, url, "", d, http.StatusOK)
	como(t, rolClienteT, 0)
	call(t, http.MethodGet, url, "", d, http.StatusNotFound)
}

func TestActualizarEstadoSoloPersonal(t *testing.T) {
	defer resetFake()
	serve(pedidoOK())
	u := func(c *PedidoController) { c.UpdateEstadoPedido() }
	const url = "/pedidos/actualizar-estado?pedido_id=10&estado=CANCELADO"
	dueno(t)
	call(t, http.MethodPut, url, "", u, http.StatusForbidden)
	como(t, rolMesero, 5)
	call(t, http.MethodPut, url, "", u, http.StatusOK)
}

func TestAssignDomicilioCliente(t *testing.T) {
	defer resetFake()
	g := grabar(t)
	a := func(c *PedidoController) { c.AssignDomicilio() }
	const url = "/pedidos/asignar-domicilio?pedido_id=10&domicilio_id=3"

	// pedido propio y domicilio libre
	serve(count("pedido", 0), count("domicilio", 1), pedidoOK())
	dueno(t)
	call(t, http.MethodPost, url, "", a, http.StatusOK)
	unico(t, g)
	// pedido ajeno: igual que inexistente
	ajeno(t)
	if b := call(t, http.MethodPost, url, "", a, http.StatusNotFound); !strings.Contains(b, "Pedido no encontrado") {
		t.Fatalf("pedido ajeno: %s", b)
	}
	// domicilio que ya es de otro pedido
	dueno(t)
	serve(count("pedido", 1), count("domicilio", 1), pedidoOK())
	if b := call(t, http.MethodPost, url, "", a, http.StatusNotFound); !strings.Contains(b, "Domicilio no encontrado") {
		t.Fatalf("domicilio de otro pedido: %s", b)
	}
	// el personal no tiene esa restricción (ni consulta)
	como(t, rolMesero, 5)
	call(t, http.MethodPost, url, "", a, http.StatusOK)
	// fallo de la consulta de pertenencia
	dueno(t)
	fakeQuery = func(q string, _ []driver.NamedValue) (driver.Rows, error) {
		switch {
		case strings.Contains(q, `COUNT(*) FROM "pedido"`):
			return nil, errBoom
		case strings.Contains(q, `COUNT(*) FROM "domicilio"`):
			return rowsOf(countCols, []driver.Value{int64(1)}), nil
		}
		return rowsOf(pedidoCols, pedidoRow()), nil
	}
	call(t, http.MethodPost, url, "", a, http.StatusInternalServerError)
}

func TestAssignPagoCliente(t *testing.T) {
	defer resetFake()
	g := grabar(t)
	a := func(c *PedidoController) { c.AssignPago() }
	const base = "/pedidos/asignar-pago?pedido_id=10&pago_id=4"

	var seen []string
	fakeExec = func(q string, _ []driver.NamedValue) (driver.Result, error) {
		seen = append(seen, q)
		return fakeResult{}, nil
	}
	dueno(t)
	// el cliente nunca termina el pedido ni marca el pago: ni explícito ni por defecto (403 antes de tocar la BD)
	serve(count("pedido", 0), count("pago", 1), pedidoSinPago())
	call(t, http.MethodPost, base, "", a, http.StatusForbidden)
	call(t, http.MethodPost, base+"&cambiar_estado=true", "", a, http.StatusForbidden)
	if len(seen) != 0 || len(g.eventos) != 0 {
		t.Fatalf("no debe escribir ni notificar: %v %v", seen, g.eventos)
	}
	// vincula su pago a su pedido sin cambiar estados
	b := call(t, http.MethodPost, base+"&cambiar_estado=false", "", a, http.StatusOK)
	contains(t, b, `"estadoPedido":"INICIADO"`)
	for _, q := range seen {
		if strings.HasPrefix(q, `UPDATE "pago"`) {
			t.Fatalf("no debe tocar el pago: %v", seen)
		}
	}
	// pedido ajeno
	ajeno(t)
	call(t, http.MethodPost, base+"&cambiar_estado=false", "", a, http.StatusNotFound)
	dueno(t)
	// pago que ya es de otro pedido
	serve(count("pedido", 1), count("pago", 1), pedidoSinPago())
	if b := call(t, http.MethodPost, base+"&cambiar_estado=false", "", a, http.StatusNotFound); !strings.Contains(b, "Pago no encontrado") {
		t.Fatalf("pago de otro pedido: %s", b)
	}
	// el pedido ya tiene pago: no se reemplaza
	serve(count("pedido", 0), count("pago", 1), pedidoOK())
	if b := call(t, http.MethodPost, base+"&cambiar_estado=false", "", a, http.StatusConflict); !strings.Contains(b, "ya tiene un pago") {
		t.Fatalf("reemplazo de pago: %s", b)
	}
	// error al comprobar la pertenencia del pago
	fakeQuery = func(q string, _ []driver.NamedValue) (driver.Rows, error) {
		switch {
		case strings.Contains(q, `COUNT(*) FROM "pedido"`):
			return nil, errBoom
		case strings.Contains(q, `COUNT(*) FROM "pago"`):
			return rowsOf(countCols, []driver.Value{int64(1)}), nil
		}
		return rowsOf(pedidoCols, pedidoSinPagoRow()), nil
	}
	call(t, http.MethodPost, base+"&cambiar_estado=false", "", a, http.StatusInternalServerError)

	// el personal conserva el comportamiento por defecto (TERMINADO + PAGADO) y puede reemplazar el pago
	como(t, rolDomi, 9)
	serve(count("pago", 1), pedidoOK())
	g.eventos = nil
	call(t, http.MethodPost, base, "", a, http.StatusOK)
	if ev := unico(t, g); ev.Estado != "TERMINADO" {
		t.Fatalf("evento inesperado: %+v", ev)
	}
}
