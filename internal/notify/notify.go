// Package notify envía desde el servidor las notificaciones push que dispara
// la operación del negocio (pedidos, domicilios y reservas). Los controladores
// solo informan el evento; el envío es best-effort: ocurre en segundo plano,
// con tiempo límite, nunca hace fallar ni demora la petición y degrada en
// silencio cuando el push no está configurado (sin claves VAPID ni proyecto FCM).
package notify

import (
	"time"

	"restaurante/services"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
)

// Tipo identifica el evento de negocio que originó la notificación.
type Tipo string

const (
	// PedidoCreado: nuevo pedido -> trabajadores y confirmación al cliente.
	PedidoCreado Tipo = "PEDIDO_CREADO"
	// PedidoDomicilio: un pedido existente recibe domicilio -> trabajadores.
	PedidoDomicilio Tipo = "PEDIDO_DOMICILIO"
	// PedidoEstado: el pedido cambió de estado -> cliente.
	PedidoEstado Tipo = "PEDIDO_ESTADO"
	// DomicilioAsignado: un domiciliario tomó el domicilio -> cliente del pedido.
	DomicilioAsignado Tipo = "DOMICILIO_ASIGNADO"
	// DomicilioEntregado: el domicilio se marcó ENTREGADO -> cliente del pedido.
	DomicilioEntregado Tipo = "DOMICILIO_ENTREGADO"
	// ReservaCreada: nueva reserva -> cliente registrado y, si la hizo un
	// cliente/invitado, trabajadores.
	ReservaCreada Tipo = "RESERVA_CREADA"
	// ReservaEstado: la reserva cambió de estado -> cliente registrado y, si el
	// propio cliente/invitado la canceló, trabajadores.
	ReservaEstado Tipo = "RESERVA_ESTADO"
)

// Evento reúne los datos mínimos que necesitan las plantillas. Los campos que
// no aplican al Tipo se dejan en su valor cero.
type Evento struct {
	Tipo Tipo
	// PedidoID y DomicilioID identifican el pedido/domicilio involucrado.
	PedidoID    int64
	DomicilioID int64
	// Cliente es el documento del cliente a notificar (0 = ninguno; en los
	// eventos de domicilio se resuelve a partir del pedido).
	Cliente int64
	// Estado es el nuevo estado (pedido o reserva).
	Estado string
	// ReservaID, Fecha (YYYY-MM-DD) y Hora (HH:MM:SS) describen la reserva.
	ReservaID int64
	Fecha     string
	Hora      string
	// DesdePersonal indica que la acción la hizo un trabajador (no se avisa a
	// los trabajadores de lo que ellos mismos hicieron).
	DesdePersonal bool
}

// Notifier recibe los eventos del negocio. Notificar no bloquea, no devuelve
// error y no entra en pánico.
type Notifier interface {
	Notificar(ev Evento)
}

// Default es el notificador que usan los controladores; los tests lo sustituyen.
var Default Notifier = &asyncNotifier{timeout: 30 * time.Second}

// Variables sustituibles en tests.
var (
	configurado   = services.PushConfigurado
	nuevoServicio = func() services.PushServiceInterface { return services.NewPushService(orm.NewOrm()) }
	// pedidoDeDomicilio devuelve el pedido asociado a un domicilio y su cliente
	// (0 si no hay pedido o no tiene cliente).
	pedidoDeDomicilio = func(domicilioID int64) (pedidoID, cliente int64, err error) {
		err = orm.NewOrm().Raw(
			"SELECT pk_id_pedido, COALESCE(pk_documento_cliente, 0) FROM pedido WHERE pk_id_domicilio = ? LIMIT 1",
			domicilioID,
		).QueryRow(&pedidoID, &cliente)
		if err == orm.ErrNoRows {
			return 0, 0, nil
		}
		return pedidoID, cliente, err
	}
)

// Enviar es el punto de uso para los controladores: delega en Default.
func Enviar(ev Evento) { Default.Notificar(ev) }

func resolver(ev Evento) (Evento, error) {
	if ev.Tipo != DomicilioAsignado && ev.Tipo != DomicilioEntregado {
		return ev, nil
	}
	pedido, cliente, err := pedidoDeDomicilio(ev.DomicilioID)
	if err != nil {
		return ev, err
	}
	ev.PedidoID, ev.Cliente = pedido, cliente
	return ev, nil
}

// procesar resuelve destinatarios, arma las notificaciones y las envía.
func procesar(ev Evento) {
	ev, err := resolver(ev)
	if err != nil {
		logs.Error("[Notify] No se pudo resolver el pedido del domicilio %d (%s): %v", ev.DomicilioID, ev.Tipo, err)
		return
	}
	reqs := construir(ev)
	if len(reqs) == 0 {
		return
	}
	svc := nuevoServicio()
	for _, req := range reqs {
		if _, err := svc.EnviarNotificacion(req); err != nil {
			logs.Error("[Notify] Error al enviar la notificación %q (%s): %v", req.Notificacion.Titulo, ev.Tipo, err)
		}
	}
}
