package models

// EstadoDomicilio represents the estado_domicilio_enum values.
type EstadoDomicilio = string

const (
	EstadoDomicilioPendiente EstadoDomicilio = "PENDIENTE"
	EstadoDomicilioEnCamino  EstadoDomicilio = "EN_CAMINO"
	EstadoDomicilioEntregado EstadoDomicilio = "ENTREGADO"
)

// EstadoNomina represents the estado_nomina_enum values.
type EstadoNomina = string

const (
	EstadoNominaPago   EstadoNomina = "PAGO"
	EstadoNominaNoPago EstadoNomina = "NO_PAGO"
)

// EstadoPago represents the estado_pago_enum values.
type EstadoPago = string

const (
	EstadoPagoPagado    EstadoPago = "PAGADO"
	EstadoPagoPendiente EstadoPago = "PENDIENTE"
	EstadoPagoNoPago    EstadoPago = "NO_PAGO"
)

// EstadoPedido represents the estado_pedido_enum values.
type EstadoPedido = string

const (
	EstadoPedidoIniciado  EstadoPedido = "INICIADO"
	EstadoPedidoTerminado EstadoPedido = "TERMINADO"
)

// EstadoProducto represents the estado_producto_enum values.
type EstadoProducto = string

const (
	EstadoProductoDisponible   EstadoProducto = "DISPONIBLE"
	EstadoProductoNoDisponible EstadoProducto = "NO_DISPONIBLE"
)

// EstadoReserva represents the estado_reserva_enum values.
type EstadoReserva = string

const (
	EstadoReservaPendiente  EstadoReserva = "PENDIENTE"
	EstadoReservaConfirmada EstadoReserva = "CONFIRMADA"
	EstadoReservaCancelada  EstadoReserva = "CANCELADA"
	EstadoReservaCumplida   EstadoReserva = "CUMPLIDA"
)

// DiaSemana represents the dia_semana_enum values.
type DiaSemana = string

const (
	DiaLunes     DiaSemana = "LUNES"
	DiaMartes    DiaSemana = "MARTES"
	DiaMiercoles DiaSemana = "MIERCOLES"
	DiaJueves    DiaSemana = "JUEVES"
	DiaViernes   DiaSemana = "VIERNES"
	DiaSabado    DiaSemana = "SABADO"
	DiaDomingo   DiaSemana = "DOMINGO"
)
