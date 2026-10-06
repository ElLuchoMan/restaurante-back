package pedido

import (
	"net/http"
	"testing"

	"restaurante/internal/notify"
)

// grabador sustituye a notify.Default y guarda los eventos recibidos.
type grabador struct{ eventos []notify.Evento }

func (g *grabador) Notificar(ev notify.Evento) { g.eventos = append(g.eventos, ev) }

func grabar(t *testing.T) *grabador {
	t.Helper()
	prev := notify.Default
	g := &grabador{}
	notify.Default = g
	t.Cleanup(func() { notify.Default = prev })
	return g
}

func unico(t *testing.T, g *grabador) notify.Evento {
	t.Helper()
	if len(g.eventos) != 1 {
		t.Fatalf("esperaba 1 evento, hay %d: %+v", len(g.eventos), g.eventos)
	}
	return g.eventos[0]
}

func TestPostNotificaPedidoCreado(t *testing.T) {
	defer resetFake()
	g := grabar(t)
	p := func(c *PedidoController) { c.Post() }

	serve(count("domicilio", 1), count("restaurante", 1), count("cliente", 1))
	call(t, http.MethodPost, "/pedidos", `{"delivery":true,"pk_id_domicilio":3,"restauranteId":1,"documentoCliente":1001}`, p, http.StatusCreated)
	ev := unico(t, g)
	if ev.Tipo != notify.PedidoCreado || ev.PedidoID != 7 || ev.Cliente != 1001 || ev.DomicilioID != 3 {
		t.Fatalf("evento inesperado: %+v", ev)
	}

	// sin cliente ni domicilio
	g.eventos = nil
	call(t, http.MethodPost, "/pedidos", `{}`, p, http.StatusCreated)
	ev = unico(t, g)
	if ev.Cliente != 0 || ev.DomicilioID != 0 {
		t.Fatalf("no debe haber cliente ni domicilio: %+v", ev)
	}

	// si la creación falla no se notifica
	g.eventos = nil
	call(t, http.MethodPost, "/pedidos", `{"delivery":true}`, p, http.StatusBadRequest)
	if len(g.eventos) != 0 {
		t.Fatalf("no debe notificar ante errores: %+v", g.eventos)
	}
}

func TestAssignDomicilioNotificaATrabajadores(t *testing.T) {
	defer resetFake()
	g := grabar(t)
	serve(pedidoOK(), count("domicilio", 1))
	call(t, http.MethodPost, "/pedidos/asignar-domicilio?pedido_id=10&domicilio_id=3", "", func(c *PedidoController) { c.AssignDomicilio() }, http.StatusOK)
	ev := unico(t, g)
	if ev.Tipo != notify.PedidoDomicilio || ev.PedidoID != 10 || ev.DomicilioID != 3 {
		t.Fatalf("evento inesperado: %+v", ev)
	}
}

func TestUpdateEstadoNotificaSoloSiCambia(t *testing.T) {
	defer resetFake()
	g := grabar(t)
	serve(pedidoOK()) // el pedido está INICIADO
	u := func(c *PedidoController) { c.UpdateEstadoPedido() }

	call(t, http.MethodPut, "/pedidos/actualizar-estado?pedido_id=10&estado=listo", "", u, http.StatusOK)
	ev := unico(t, g)
	if ev.Tipo != notify.PedidoEstado || ev.PedidoID != 10 || ev.Cliente != 1001 || ev.Estado != "LISTO" {
		t.Fatalf("evento inesperado: %+v", ev)
	}

	g.eventos = nil
	call(t, http.MethodPut, "/pedidos/actualizar-estado?pedido_id=10&estado=INICIADO", "", u, http.StatusOK)
	if len(g.eventos) != 0 {
		t.Fatalf("mismo estado: no debe notificar: %+v", g.eventos)
	}
}

func TestAssignPagoNotificaTerminado(t *testing.T) {
	defer resetFake()
	g := grabar(t)
	serve(pedidoOK(), count("pago", 1))
	a := func(c *PedidoController) { c.AssignPago() }
	u := "/pedidos/asignar-pago?pedido_id=10&pago_id=4"

	call(t, http.MethodPost, u, "", a, http.StatusOK)
	ev := unico(t, g)
	if ev.Tipo != notify.PedidoEstado || ev.Estado != "TERMINADO" || ev.Cliente != 1001 {
		t.Fatalf("evento inesperado: %+v", ev)
	}

	g.eventos = nil
	call(t, http.MethodPost, u+"&cambiar_estado=false", "", a, http.StatusOK)
	if len(g.eventos) != 0 {
		t.Fatalf("sin cambio de estado no debe notificar: %+v", g.eventos)
	}
}
