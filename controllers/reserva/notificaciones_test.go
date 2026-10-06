package reserva

import (
	"net/http"
	"testing"

	"restaurante/internal/notify"
	"restaurante/models"
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

// estadosPut programa la carga inicial (antes) y la relectura (despues) de un
// PUT: la primera consulta de reserva devuelve `antes` y la segunda `despues`.
func estadosPut(antes, despues any, cliente any) []*res {
	primera := reservaRows(filaReserva(3, antes, cliente))
	primera.onlyCall = 1
	segunda := reservaRows(filaReserva(3, despues, cliente))
	segunda.onlyCall = 1
	return []*res{primera, segunda}
}

func TestPostNotificaReservaCreadaInvitado(t *testing.T) {
	g := grabar(t)
	programa(t, postOK(vacio(qContacto, 5))...)
	expect(t, runAs("", "POST", "/reservas", cuerpoNuevoInvitado, (*ReservaController).Post), http.StatusCreated)
	ev := unico(t, g)
	want := notify.Evento{Tipo: notify.ReservaCreada, ReservaID: 7, Estado: "PENDIENTE", Fecha: "2025-01-31", Hora: "18:30:00"}
	if ev != want {
		t.Fatalf("evento %+v != %+v", ev, want)
	}
}

func TestPostNotificaAlClienteDelToken(t *testing.T) {
	g := grabar(t)
	// invitado con el mismo documento que el token de un cliente
	programa(t, postOK(contactoRows(filaContacto(2, int64(55), nil)))...)
	expect(t, runAs(tokCliente(docInvitado), "POST", "/reservas", cuerpoNuevoInvitado, (*ReservaController).Post), http.StatusCreated)
	ev := unico(t, g)
	if ev.Cliente != docInvitado || ev.DesdePersonal {
		t.Fatalf("debe avisar al cliente del token: %+v", ev)
	}
}

func TestPostNotificaAlClienteRegistradoDesdePersonal(t *testing.T) {
	g := grabar(t)
	body := `{"documentoCliente":9,"restauranteId":1,"fechaReserva":"2025-01-31","horaReserva":"18:30:00","personas":2}`
	programa(t, vacio(qContacto, 5), clienteRows(), restauranteRows(), reservaRows(filaReserva(7, "PENDIENTE", int64(9))))
	expect(t, run("POST", "/reservas", body, (*ReservaController).Post), http.StatusCreated)
	ev := unico(t, g)
	if ev.Cliente != docDueno || !ev.DesdePersonal {
		t.Fatalf("debe avisar al cliente del contacto y marcar al personal: %+v", ev)
	}
}

func TestPostSinEstadoCargadoNoFallaNiInventaEstado(t *testing.T) {
	g := grabar(t)
	programa(t, restauranteRows(), contactoRows(filaContacto(2, int64(55), nil)), reservaRows(filaReserva(7, nil, nil)))
	expect(t, runAs("", "POST", "/reservas", cuerpoNuevoInvitado, (*ReservaController).Post), http.StatusCreated)
	if ev := unico(t, g); ev.Estado != "" {
		t.Fatalf("sin estado el evento no debe traerlo: %+v", ev)
	}
}

func TestPostConErrorNoNotifica(t *testing.T) {
	g := grabar(t)
	programa(t, conError(qRestaurante))
	expect(t, runAs("", "POST", "/reservas", cuerpoNuevoInvitado, (*ReservaController).Post), http.StatusInternalServerError)
	if len(g.eventos) != 0 {
		t.Fatalf("no debe notificar ante errores: %+v", g.eventos)
	}
}

func TestPutNotificaCambioDeEstadoDelPersonal(t *testing.T) {
	g := grabar(t)
	programa(t, estadosPut("PENDIENTE", "CONFIRMADA", int64(9))...)
	expect(t, run("PUT", "/reservas?id=3", `{"estadoReserva":"CONFIRMADA"}`, (*ReservaController).Put), http.StatusOK)
	ev := unico(t, g)
	want := notify.Evento{Tipo: notify.ReservaEstado, ReservaID: 3, Cliente: docDueno, Estado: "CONFIRMADA", Fecha: "2025-01-31", Hora: "18:30:00", DesdePersonal: true}
	if ev != want {
		t.Fatalf("evento %+v != %+v", ev, want)
	}
}

func TestPutNotificaCancelacionDelCliente(t *testing.T) {
	g := grabar(t)
	programa(t, estadosPut("PENDIENTE", "CANCELADA", int64(9))...)
	expect(t, runAs(tokCliente(docDueno), "PUT", "/reservas?id=3", `{"estadoReserva":"CANCELADA"}`, (*ReservaController).Put), http.StatusOK)
	ev := unico(t, g)
	if ev.Tipo != notify.ReservaEstado || ev.Estado != "CANCELADA" || ev.DesdePersonal || ev.Cliente != docDueno {
		t.Fatalf("evento inesperado: %+v", ev)
	}
}

func TestPutNoNotificaSiElEstadoNoCambia(t *testing.T) {
	g := grabar(t)
	programa(t, estadosPut("PENDIENTE", "PENDIENTE", int64(9))...)
	expect(t, run("PUT", "/reservas?id=3", `{"personas":6}`, (*ReservaController).Put), http.StatusOK)
	if len(g.eventos) != 0 {
		t.Fatalf("sin cambio de estado no debe notificar: %+v", g.eventos)
	}
}

func TestCambioEstado(t *testing.T) {
	pend, conf := "PENDIENTE", "CONFIRMADA"
	casos := []struct {
		nombre         string
		antes, despues *string
		want           bool
	}{
		{"sin estado final", &pend, nil, false},
		{"estado nuevo desde nulo", nil, &conf, true},
		{"mismo estado", &pend, &pend, false},
		{"estado distinto", &pend, &conf, true},
		{"ambos nulos", nil, nil, false},
	}
	for _, c := range casos {
		if got := cambioEstado(c.antes, c.despues); got != c.want {
			t.Errorf("%s: %v != %v", c.nombre, got, c.want)
		}
	}
}

func TestClienteNotificableSinContacto(t *testing.T) {
	if got := clienteNotificable(nil, &models.Reserva{}); got != 0 {
		t.Fatalf("sin contacto ni token no hay a quién avisar: %d", got)
	}
}
