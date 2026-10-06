package notify

import (
	"encoding/json"
	"fmt"
	"time"

	"restaurante/models"
)

// Rutas del front a las que lleva cada notificación.
const (
	urlMisPedidos   = "/cliente/mis-pedidos"
	urlDomicilios   = "/admin/domicilios/consultar"
	urlPedidosAdmin = "/admin/pedidos"
	urlReservasHoy  = "/admin/reservas/hoy"
)

func paraCliente(doc int64, titulo, mensaje string, datos map[string]any) *models.EnviarNotificacionRequest {
	return armar(models.DestinatariosNotificacion{Tipo: models.DestinatarioCliente, DocumentoCliente: &doc}, titulo, mensaje, datos)
}

func paraTrabajadores(titulo, mensaje string, datos map[string]any) *models.EnviarNotificacionRequest {
	return armar(models.DestinatariosNotificacion{Tipo: models.DestinatarioTrabajadores}, titulo, mensaje, datos)
}

func armar(dest models.DestinatariosNotificacion, titulo, mensaje string, datos map[string]any) *models.EnviarNotificacionRequest {
	// Un map[string]any de escalares siempre se serializa.
	raw, _ := json.Marshal(datos)
	return &models.EnviarNotificacionRequest{
		Remitente:     models.RemitenteNotificacion{Tipo: models.RemitenteSistema},
		Destinatarios: dest,
		Notificacion:  models.ContenidoNotificacion{Titulo: titulo, Mensaje: mensaje, Datos: raw},
	}
}

// construir devuelve las notificaciones a enviar para el evento (puede ser
// ninguna).
func construir(ev Evento) []*models.EnviarNotificacionRequest {
	switch ev.Tipo {
	case PedidoCreado:
		return pedidoCreado(ev)
	case PedidoDomicilio:
		return []*models.EnviarNotificacionRequest{avisoDomicilioPedido(ev)}
	case PedidoEstado:
		return pedidoEstado(ev)
	case DomicilioAsignado:
		return domicilioCliente(ev, "Tu domicilio va en camino",
			fmt.Sprintf("Un domiciliario tomó tu pedido #%d y va en camino.", ev.PedidoID), "EN_CAMINO")
	case DomicilioEntregado:
		return domicilioCliente(ev, "Domicilio entregado",
			fmt.Sprintf("Tu pedido #%d fue entregado. ¡Buen provecho!", ev.PedidoID), "ENTREGADO")
	case ReservaCreada:
		return reservaCreada(ev)
	case ReservaEstado:
		return reservaEstado(ev)
	}
	return nil
}

func avisoDomicilioPedido(ev Evento) *models.EnviarNotificacionRequest {
	return paraTrabajadores(
		"Nuevo pedido con domicilio",
		fmt.Sprintf("Se ha creado un nuevo pedido (#%d) que requiere asignación de domiciliario. Domicilio #%d.", ev.PedidoID, ev.DomicilioID),
		map[string]any{"tipo": "PEDIDO_DOMICILIO", "pedidoId": ev.PedidoID, "domicilioId": ev.DomicilioID, "url": urlDomicilios},
	)
}

func pedidoCreado(ev Evento) []*models.EnviarNotificacionRequest {
	var out []*models.EnviarNotificacionRequest
	if ev.Cliente > 0 {
		out = append(out, paraCliente(ev.Cliente, "Pedido recibido",
			"Recibimos tu pedido exitosamente. Pronto recibirás actualizaciones sobre su estado. ¡Gracias por tu compra!",
			map[string]any{"tipo": "PEDIDO", "pedidoId": ev.PedidoID, "url": urlMisPedidos}))
	}
	if ev.DomicilioID > 0 {
		return append(out, avisoDomicilioPedido(ev))
	}
	return append(out, paraTrabajadores("Nuevo pedido",
		fmt.Sprintf("Se ha creado un nuevo pedido (#%d).", ev.PedidoID),
		map[string]any{"tipo": "PEDIDO_NUEVO", "pedidoId": ev.PedidoID, "url": urlPedidosAdmin}))
}

func pedidoEstado(ev Evento) []*models.EnviarNotificacionRequest {
	if ev.Cliente <= 0 {
		return nil
	}
	var titulo, mensaje string
	switch ev.Estado {
	case models.EstadoPedidoEnPreparacion:
		titulo, mensaje = "Estamos preparando tu pedido", fmt.Sprintf("Tu pedido #%d está en preparación.", ev.PedidoID)
	case models.EstadoPedidoListo:
		titulo, mensaje = "Tu pedido está listo", fmt.Sprintf("Tu pedido #%d está listo.", ev.PedidoID)
	case models.EstadoPedidoTerminado:
		titulo, mensaje = "Pedido finalizado", fmt.Sprintf("Tu pedido #%d fue finalizado. ¡Gracias por tu compra!", ev.PedidoID)
	case models.EstadoPedidoCancelado:
		titulo, mensaje = "Pedido cancelado", fmt.Sprintf("Tu pedido #%d fue cancelado. Si necesitas ayuda, contáctanos.", ev.PedidoID)
	default:
		return nil
	}
	return []*models.EnviarNotificacionRequest{paraCliente(ev.Cliente, titulo, mensaje,
		map[string]any{"tipo": "PEDIDO", "pedidoId": ev.PedidoID, "estado": ev.Estado, "url": urlMisPedidos})}
}

func domicilioCliente(ev Evento, titulo, mensaje, estado string) []*models.EnviarNotificacionRequest {
	if ev.Cliente <= 0 {
		return nil
	}
	return []*models.EnviarNotificacionRequest{paraCliente(ev.Cliente, titulo, mensaje,
		map[string]any{"tipo": "DOMICILIO", "pedidoId": ev.PedidoID, "domicilioId": ev.DomicilioID, "estado": estado, "url": urlMisPedidos})}
}

// fechaHora devuelve la fecha como dd/mm y la hora como HH:MM (24 h, como en
// es-CO); si no se pueden interpretar, devuelve los textos originales.
func fechaHora(fecha, hora string) (string, string) {
	f, err := time.Parse("2006-01-02", fecha)
	if err != nil {
		return fecha, hora
	}
	h, err := time.Parse("15:04:05", hora)
	if err != nil {
		return fecha, hora
	}
	return f.Format("02/01"), h.Format("15:04")
}

func urlReserva(id int64) string {
	return fmt.Sprintf("/reservas/consultar?reservaId=%d", id)
}

func datosReserva(ev Evento, estado, url string) map[string]any {
	return map[string]any{"tipo": "RESERVA", "reservaId": ev.ReservaID, "estado": estado, "url": url}
}

func reservaCreada(ev Evento) []*models.EnviarNotificacionRequest {
	fecha, hora := fechaHora(ev.Fecha, ev.Hora)
	var out []*models.EnviarNotificacionRequest
	if ev.Cliente > 0 {
		out = append(out, paraCliente(ev.Cliente, "Reserva creada",
			fmt.Sprintf("Recibimos tu solicitud para el %s %s. Te llamaremos días antes para confirmar.", fecha, hora),
			datosReserva(ev, models.EstadoReservaPendiente, urlReserva(ev.ReservaID))))
	}
	if !ev.DesdePersonal {
		out = append(out, paraTrabajadores("Nueva reserva",
			fmt.Sprintf("Nueva reserva #%d para el %s a las %s.", ev.ReservaID, fecha, hora),
			datosReserva(ev, models.EstadoReservaPendiente, urlReservasHoy)))
	}
	return out
}

func reservaEstado(ev Evento) []*models.EnviarNotificacionRequest {
	fecha, hora := fechaHora(ev.Fecha, ev.Hora)
	var out []*models.EnviarNotificacionRequest
	if ev.Cliente > 0 {
		var titulo, mensaje string
		switch ev.Estado {
		case models.EstadoReservaConfirmada:
			titulo, mensaje = "Reserva confirmada", fmt.Sprintf("Tu reserva para el %s a las %s quedó confirmada. ¡Te esperamos!", fecha, hora)
		case models.EstadoReservaCancelada:
			titulo, mensaje = "Reserva cancelada", fmt.Sprintf("Tu reserva del %s a las %s fue cancelada. Si necesitas ayuda, contáctanos.", fecha, hora)
		case models.EstadoReservaCumplida:
			titulo, mensaje = "¡Gracias por visitarnos!", "¿Nos dejas tu opinión? Te tomará 30 segundos."
		default:
			titulo, mensaje = "Recibimos tu solicitud", fmt.Sprintf("Estamos validando tu reserva del %s %s. Te contactaremos para confirmar.", fecha, hora)
		}
		out = append(out, paraCliente(ev.Cliente, titulo, mensaje, datosReserva(ev, ev.Estado, urlReserva(ev.ReservaID))))
	}
	if !ev.DesdePersonal && ev.Estado == models.EstadoReservaCancelada {
		out = append(out, paraTrabajadores("Reserva cancelada por el cliente",
			fmt.Sprintf("Se canceló la reserva #%d del %s a las %s.", ev.ReservaID, fecha, hora),
			datosReserva(ev, ev.Estado, urlReservasHoy)))
	}
	return out
}
