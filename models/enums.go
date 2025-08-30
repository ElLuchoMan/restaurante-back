package models

// EstadoDomicilio represents the estado_domicilio_enum values.
type EstadoDomicilio = string

const (
	EstadoDomicilioPendiente EstadoDomicilio = "pendiente"
	EstadoDomicilioEnCamino  EstadoDomicilio = "en camino"
	EstadoDomicilioEntregado EstadoDomicilio = "entregado"
)

// EstadoNomina represents the estado_nomina_enum values.
type EstadoNomina = string

const (
	EstadoNominaPago   EstadoNomina = "pago"
	EstadoNominaNoPago EstadoNomina = "no pago"
)

// EstadoPago represents the estado_pago_enum values.
type EstadoPago = string

const (
	EstadoPagoPagado    EstadoPago = "pagado"
	EstadoPagoPendiente EstadoPago = "pendiente"
	EstadoPagoNoPago    EstadoPago = "no pago"
)

// EstadoPedido represents the estado_pedido_enum values.
type EstadoPedido = string

const (
	EstadoPedidoIniciado  EstadoPedido = "iniciado"
	EstadoPedidoTerminado EstadoPedido = "terminado"
)

// EstadoProducto represents the estado_producto_enum values.
type EstadoProducto = string

const (
	EstadoProductoDisponible   EstadoProducto = "disponible"
	EstadoProductoNoDisponible EstadoProducto = "no disponible"
)

// EstadoReserva represents the estado_reserva_enum values.
type EstadoReserva = string

const (
	EstadoReservaPendiente  EstadoReserva = "pendiente"
	EstadoReservaConfirmada EstadoReserva = "confirmada"
	EstadoReservaCancelada  EstadoReserva = "cancelada"
	EstadoReservaCumplida   EstadoReserva = "cumplida"
)

// DiaSemana represents the dia_semana_enum values.
type DiaSemana = string

const (
	DiaLunes     DiaSemana = "lunes"
	DiaMartes    DiaSemana = "martes"
	DiaMiercoles DiaSemana = "miercoles"
	DiaJueves    DiaSemana = "jueves"
	DiaViernes   DiaSemana = "viernes"
	DiaSabado    DiaSemana = "sabado"
	DiaDomingo   DiaSemana = "domingo"
)
