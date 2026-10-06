package notify

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"restaurante/models"
	"restaurante/services"
)

// svcFalso implementa services.PushServiceInterface registrando lo enviado.
type svcFalso struct {
	mu       sync.Mutex
	enviados []*models.EnviarNotificacionRequest
	err      error
	panico   bool
	bloqueo  chan struct{}
}

func (s *svcFalso) EnviarNotificacion(r *models.EnviarNotificacionRequest) (*models.EnviarNotificacionResponse, error) {
	if s.panico {
		panic("boom")
	}
	if s.bloqueo != nil {
		<-s.bloqueo
	}
	s.mu.Lock()
	s.enviados = append(s.enviados, r)
	s.mu.Unlock()
	return nil, s.err
}

func (s *svcFalso) RegistrarDispositivo(context.Context, *models.RegistrarDispositivoRequest) (*models.PushDispositivo, bool, error) {
	return nil, false, nil
}
func (s *svcFalso) ActualizarUltimaVista(context.Context, int64) error { return nil }
func (s *svcFalso) ActualizarDispositivo(context.Context, int64, []byte) (*models.PushDispositivo, error) {
	return nil, nil
}
func (s *svcFalso) ActualizarTopicsDispositivo(context.Context, int64, []string) error { return nil }
func (s *svcFalso) RegistrarEnvio(context.Context, *models.RegistrarEnvioRequest) (*models.PushEnvio, error) {
	return nil, nil
}
func (s *svcFalso) ValidarRegistroDispositivo(*models.RegistrarDispositivoRequest) error { return nil }

func (s *svcFalso) total() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.enviados)
}

// preparar sustituye las dependencias del paquete durante el test.
func preparar(t *testing.T, svc *svcFalso, activo bool) {
	t.Helper()
	prevCfg, prevSvc, prevPed := configurado, nuevoServicio, pedidoDeDomicilio
	configurado = func() bool { return activo }
	nuevoServicio = func() services.PushServiceInterface { return svc }
	t.Cleanup(func() { configurado, nuevoServicio, pedidoDeDomicilio = prevCfg, prevSvc, prevPed })
}

func datos(t *testing.T, r *models.EnviarNotificacionRequest) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(r.Notificacion.Datos, &m); err != nil {
		t.Fatalf("datos no es JSON: %v", err)
	}
	return m
}

func TestConstruirTodosLosEventos(t *testing.T) {
	casos := []struct {
		nombre   string
		ev       Evento
		destinos []models.TipoDestinatario
		titulo   string
		contiene string
		tipo     string
		url      string
	}{
		{"pedido creado con cliente y domicilio", Evento{Tipo: PedidoCreado, PedidoID: 5, Cliente: 10, DomicilioID: 3},
			[]models.TipoDestinatario{models.DestinatarioCliente, models.DestinatarioTrabajadores}, "Pedido recibido", "", "PEDIDO", "/cliente/mis-pedidos"},
		{"pedido creado sin cliente ni domicilio", Evento{Tipo: PedidoCreado, PedidoID: 5},
			[]models.TipoDestinatario{models.DestinatarioTrabajadores}, "Nuevo pedido", "(#5)", "PEDIDO_NUEVO", "/admin/pedidos"},
		{"pedido con domicilio asignado", Evento{Tipo: PedidoDomicilio, PedidoID: 5, DomicilioID: 3},
			[]models.TipoDestinatario{models.DestinatarioTrabajadores}, "Nuevo pedido con domicilio", "Domicilio #3", "PEDIDO_DOMICILIO", "/admin/domicilios/consultar"},
		{"pedido en preparación", Evento{Tipo: PedidoEstado, PedidoID: 5, Cliente: 10, Estado: models.EstadoPedidoEnPreparacion},
			[]models.TipoDestinatario{models.DestinatarioCliente}, "Estamos preparando tu pedido", "#5", "PEDIDO", "/cliente/mis-pedidos"},
		{"pedido listo", Evento{Tipo: PedidoEstado, PedidoID: 5, Cliente: 10, Estado: models.EstadoPedidoListo},
			[]models.TipoDestinatario{models.DestinatarioCliente}, "Tu pedido está listo", "", "PEDIDO", ""},
		{"pedido terminado", Evento{Tipo: PedidoEstado, PedidoID: 5, Cliente: 10, Estado: models.EstadoPedidoTerminado},
			[]models.TipoDestinatario{models.DestinatarioCliente}, "Pedido finalizado", "", "PEDIDO", ""},
		{"pedido cancelado", Evento{Tipo: PedidoEstado, PedidoID: 5, Cliente: 10, Estado: models.EstadoPedidoCancelado},
			[]models.TipoDestinatario{models.DestinatarioCliente}, "Pedido cancelado", "", "PEDIDO", ""},
		{"domicilio asignado", Evento{Tipo: DomicilioAsignado, PedidoID: 5, DomicilioID: 3, Cliente: 10},
			[]models.TipoDestinatario{models.DestinatarioCliente}, "Tu domicilio va en camino", "#5", "DOMICILIO", "/cliente/mis-pedidos"},
		{"domicilio entregado", Evento{Tipo: DomicilioEntregado, PedidoID: 5, DomicilioID: 3, Cliente: 10},
			[]models.TipoDestinatario{models.DestinatarioCliente}, "Domicilio entregado", "", "DOMICILIO", ""},
		{"reserva creada por cliente", Evento{Tipo: ReservaCreada, ReservaID: 8, Cliente: 10, Fecha: "2026-10-05", Hora: "19:30:00"},
			[]models.TipoDestinatario{models.DestinatarioCliente, models.DestinatarioTrabajadores}, "Reserva creada", "05/10 19:30", "RESERVA", "/reservas/consultar?reservaId=8"},
		{"reserva creada por invitado", Evento{Tipo: ReservaCreada, ReservaID: 8, Fecha: "2026-10-05", Hora: "19:30:00"},
			[]models.TipoDestinatario{models.DestinatarioTrabajadores}, "Nueva reserva", "#8 para el 05/10 a las 19:30", "RESERVA", "/admin/reservas/hoy"},
		{"reserva creada por personal sin cliente", Evento{Tipo: ReservaCreada, ReservaID: 8, DesdePersonal: true}, nil, "", "", "", ""},
		{"reserva confirmada", Evento{Tipo: ReservaEstado, ReservaID: 8, Cliente: 10, Estado: models.EstadoReservaConfirmada, Fecha: "2026-10-05", Hora: "19:30:00", DesdePersonal: true},
			[]models.TipoDestinatario{models.DestinatarioCliente}, "Reserva confirmada", "quedó confirmada", "RESERVA", ""},
		{"reserva cancelada por el cliente", Evento{Tipo: ReservaEstado, ReservaID: 8, Cliente: 10, Estado: models.EstadoReservaCancelada, Fecha: "2026-10-05", Hora: "19:30:00"},
			[]models.TipoDestinatario{models.DestinatarioCliente, models.DestinatarioTrabajadores}, "Reserva cancelada", "fue cancelada", "RESERVA", ""},
		{"reserva cancelada por personal", Evento{Tipo: ReservaEstado, ReservaID: 8, Cliente: 10, Estado: models.EstadoReservaCancelada, DesdePersonal: true},
			[]models.TipoDestinatario{models.DestinatarioCliente}, "Reserva cancelada", "", "RESERVA", ""},
		{"reserva cumplida", Evento{Tipo: ReservaEstado, ReservaID: 8, Cliente: 10, Estado: models.EstadoReservaCumplida, DesdePersonal: true},
			[]models.TipoDestinatario{models.DestinatarioCliente}, "¡Gracias por visitarnos!", "", "RESERVA", ""},
		{"reserva pendiente", Evento{Tipo: ReservaEstado, ReservaID: 8, Cliente: 10, Estado: models.EstadoReservaPendiente, DesdePersonal: true, Fecha: "x", Hora: "y"},
			[]models.TipoDestinatario{models.DestinatarioCliente}, "Recibimos tu solicitud", "x y", "RESERVA", ""},
		{"reserva sin cliente cancelada por personal", Evento{Tipo: ReservaEstado, ReservaID: 8, Estado: models.EstadoReservaCancelada, DesdePersonal: true}, nil, "", "", "", ""},
		{"pedido estado sin cliente", Evento{Tipo: PedidoEstado, PedidoID: 5, Estado: models.EstadoPedidoListo}, nil, "", "", "", ""},
		{"pedido estado INICIADO no notifica", Evento{Tipo: PedidoEstado, PedidoID: 5, Cliente: 10, Estado: models.EstadoPedidoIniciado}, nil, "", "", "", ""},
		{"domicilio sin cliente", Evento{Tipo: DomicilioAsignado, DomicilioID: 3}, nil, "", "", "", ""},
		{"tipo desconocido", Evento{Tipo: "OTRO"}, nil, "", "", "", ""},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			reqs := construir(c.ev)
			if len(reqs) != len(c.destinos) {
				t.Fatalf("esperaba %d notificaciones, hay %d", len(c.destinos), len(reqs))
			}
			for i, r := range reqs {
				if r.Remitente.Tipo != models.RemitenteSistema {
					t.Fatalf("remitente debe ser SISTEMA: %+v", r.Remitente)
				}
				if r.Destinatarios.Tipo != c.destinos[i] {
					t.Fatalf("destinatario %d: %s != %s", i, r.Destinatarios.Tipo, c.destinos[i])
				}
				if r.Destinatarios.Tipo == models.DestinatarioCliente &&
					(r.Destinatarios.DocumentoCliente == nil || *r.Destinatarios.DocumentoCliente != c.ev.Cliente) {
					t.Fatalf("documento de cliente incorrecto: %+v", r.Destinatarios)
				}
				if n := len([]rune(r.Notificacion.Titulo)); n < 1 || n > 100 {
					t.Fatalf("título fuera de límites: %q", r.Notificacion.Titulo)
				}
				if n := len([]rune(r.Notificacion.Mensaje)); n < 1 || n > 500 {
					t.Fatalf("mensaje fuera de límites: %q", r.Notificacion.Mensaje)
				}
			}
			if len(reqs) == 0 {
				return
			}
			if reqs[0].Notificacion.Titulo != c.titulo {
				t.Fatalf("título %q != %q", reqs[0].Notificacion.Titulo, c.titulo)
			}
			if !strings.Contains(reqs[0].Notificacion.Mensaje, c.contiene) {
				t.Fatalf("mensaje %q no contiene %q", reqs[0].Notificacion.Mensaje, c.contiene)
			}
			d := datos(t, reqs[0])
			if d["tipo"] != c.tipo {
				t.Fatalf("tipo de datos %v != %s", d["tipo"], c.tipo)
			}
			if c.url != "" && d["url"] != c.url {
				t.Fatalf("url %v != %s", d["url"], c.url)
			}
		})
	}
}

func TestFechaHora(t *testing.T) {
	if f, h := fechaHora("2026-10-05", "07:05:00"); f != "05/10" || h != "07:05" {
		t.Fatalf("formato inesperado: %s %s", f, h)
	}
	if f, h := fechaHora("mal", "07:05:00"); f != "mal" || h != "07:05:00" {
		t.Fatalf("fecha inválida debe devolver los originales: %s %s", f, h)
	}
	if f, h := fechaHora("2026-10-05", "mal"); f != "2026-10-05" || h != "mal" {
		t.Fatalf("hora inválida debe devolver los originales: %s %s", f, h)
	}
}

func TestProcesarEnviaTodas(t *testing.T) {
	svc := &svcFalso{}
	preparar(t, svc, true)
	procesar(Evento{Tipo: PedidoCreado, PedidoID: 1, Cliente: 10})
	if svc.total() != 2 {
		t.Fatalf("esperaba 2 envíos, hay %d", svc.total())
	}
}

func TestProcesarContinuaTrasErrorDeEnvio(t *testing.T) {
	svc := &svcFalso{err: errors.New("fcm caído")}
	preparar(t, svc, true)
	procesar(Evento{Tipo: PedidoCreado, PedidoID: 1, Cliente: 10})
	if svc.total() != 2 {
		t.Fatalf("un error no debe cortar los demás envíos: %d", svc.total())
	}
}

func TestProcesarSinNotificaciones(t *testing.T) {
	svc := &svcFalso{}
	creado := false
	preparar(t, svc, true)
	nuevoServicio = func() services.PushServiceInterface { creado = true; return svc }
	procesar(Evento{Tipo: PedidoEstado, PedidoID: 1, Cliente: 10, Estado: models.EstadoPedidoIniciado})
	if creado || svc.total() != 0 {
		t.Fatal("sin notificaciones no debe crearse el servicio ni enviarse nada")
	}
}

func TestProcesarResuelveClienteDeDomicilio(t *testing.T) {
	svc := &svcFalso{}
	preparar(t, svc, true)
	pedidoDeDomicilio = func(id int64) (int64, int64, error) {
		if id != 3 {
			t.Fatalf("domicilio inesperado %d", id)
		}
		return 55, 10, nil
	}
	procesar(Evento{Tipo: DomicilioAsignado, DomicilioID: 3})
	if svc.total() != 1 {
		t.Fatalf("esperaba 1 envío, hay %d", svc.total())
	}
	if m := svc.enviados[0].Notificacion.Mensaje; !strings.Contains(m, "#55") {
		t.Fatalf("debe usar el pedido resuelto: %s", m)
	}
	if *svc.enviados[0].Destinatarios.DocumentoCliente != 10 {
		t.Fatal("debe notificar al cliente del pedido")
	}

	// domicilio sin pedido o sin cliente: no hay a quién avisar
	pedidoDeDomicilio = func(int64) (int64, int64, error) { return 0, 0, nil }
	procesar(Evento{Tipo: DomicilioEntregado, DomicilioID: 3})
	if svc.total() != 1 {
		t.Fatal("sin cliente no debe enviar")
	}

	// error de BD: se registra y no se envía
	pedidoDeDomicilio = func(int64) (int64, int64, error) { return 0, 0, errors.New("db") }
	procesar(Evento{Tipo: DomicilioEntregado, DomicilioID: 3})
	if svc.total() != 1 {
		t.Fatal("con error de BD no debe enviar")
	}
}

func TestResolverNoTocaOtrosEventos(t *testing.T) {
	preparar(t, &svcFalso{}, true)
	pedidoDeDomicilio = func(int64) (int64, int64, error) { t.Fatal("no debe consultarse"); return 0, 0, nil }
	ev, err := resolver(Evento{Tipo: PedidoCreado, PedidoID: 4})
	if err != nil || ev.PedidoID != 4 {
		t.Fatalf("evento alterado: %+v %v", ev, err)
	}
}

func TestAsyncNoHaceNadaSinConfiguracion(t *testing.T) {
	svc := &svcFalso{}
	preparar(t, svc, false)
	a := &asyncNotifier{timeout: time.Second}
	a.Notificar(Evento{Tipo: PedidoCreado, PedidoID: 1, Cliente: 10})
	a.esperar()
	if svc.total() != 0 {
		t.Fatal("sin VAPID/FCM no debe enviarse nada")
	}
}

func TestAsyncEnviaEnSegundoPlano(t *testing.T) {
	svc := &svcFalso{bloqueo: make(chan struct{})}
	preparar(t, svc, true)
	a := &asyncNotifier{timeout: 5 * time.Second}
	inicio := time.Now()
	a.Notificar(Evento{Tipo: PedidoCreado, PedidoID: 1, Cliente: 10})
	if time.Since(inicio) > time.Second {
		t.Fatal("Notificar no debe bloquear")
	}
	close(svc.bloqueo)
	a.esperar()
	if svc.total() != 2 {
		t.Fatalf("esperaba 2 envíos, hay %d", svc.total())
	}
}

func TestAsyncRecuperaPanico(t *testing.T) {
	svc := &svcFalso{panico: true}
	preparar(t, svc, true)
	a := &asyncNotifier{timeout: 5 * time.Second}
	a.Notificar(Evento{Tipo: PedidoCreado, PedidoID: 1, Cliente: 10})
	a.esperar() // si el pánico escapara, el proceso de test caería
}

func TestAsyncAbandonaTrasTiempoLimite(t *testing.T) {
	svc := &svcFalso{bloqueo: make(chan struct{})}
	preparar(t, svc, true)
	a := &asyncNotifier{timeout: 20 * time.Millisecond}
	a.Notificar(Evento{Tipo: PedidoEstado, PedidoID: 1, Cliente: 10, Estado: models.EstadoPedidoListo})
	a.esperar() // regresa por el timeout aunque el envío siga bloqueado
	close(svc.bloqueo)
	// el envío abandonado termina por su cuenta: se espera para no pisar las
	// dependencias que restaura Cleanup
	for i := 0; i < 200 && svc.total() == 0; i++ {
		time.Sleep(5 * time.Millisecond)
	}
	if svc.total() != 1 {
		t.Fatal("el envío abandonado debía completarse")
	}
}

type registro struct{ eventos []Evento }

func (r *registro) Notificar(ev Evento) { r.eventos = append(r.eventos, ev) }

func TestEnviarDelegaEnDefault(t *testing.T) {
	prev := Default
	t.Cleanup(func() { Default = prev })
	r := &registro{}
	Default = r
	Enviar(Evento{Tipo: ReservaCreada, ReservaID: 2})
	if len(r.eventos) != 1 || r.eventos[0].ReservaID != 2 {
		t.Fatalf("evento no entregado: %+v", r.eventos)
	}
}

func TestPedidoDeDomicilioPorDefecto(t *testing.T) {
	defer func() { fakeQuery = nil }()
	fakeQuery = func(q string, args []driver.Value) (driver.Rows, error) {
		if !strings.Contains(q, "FROM pedido WHERE pk_id_domicilio") || len(args) != 1 || args[0] != int64(3) {
			t.Fatalf("consulta inesperada: %s %v", q, args)
		}
		return &fakeRows{columns: []string{"a", "b"}, values: [][]driver.Value{{int64(55), int64(10)}}}, nil
	}
	if p, c, err := pedidoDeDomicilio(3); err != nil || p != 55 || c != 10 {
		t.Fatalf("lectura inesperada: %d %d %v", p, c, err)
	}

	fakeQuery = func(string, []driver.Value) (driver.Rows, error) { return &fakeRows{columns: []string{"a", "b"}}, nil }
	if p, c, err := pedidoDeDomicilio(3); err != nil || p != 0 || c != 0 {
		t.Fatalf("sin filas debe ser 0,0,nil: %d %d %v", p, c, err)
	}

	fakeQuery = func(string, []driver.Value) (driver.Rows, error) { return nil, errors.New("db") }
	if _, _, err := pedidoDeDomicilio(3); err == nil {
		t.Fatal("el error de BD debe propagarse")
	}
}

func TestProveedoresPorDefecto(t *testing.T) {
	if nuevoServicio() == nil {
		t.Fatal("el servicio por defecto no debe ser nil")
	}
	if _, ok := Default.(*asyncNotifier); !ok {
		t.Fatalf("Default debe ser el notificador asíncrono, es %T", Default)
	}
	if configurado == nil {
		t.Fatal("configurado no debe ser nil")
	}
}
