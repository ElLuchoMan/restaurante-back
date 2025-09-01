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

// EstadoControlNomina represents the allowed values for control_nomina.estado.
type EstadoControlNomina string

const (
	EstadoControlNominaNoGenerada EstadoControlNomina = "NO GENERADA"
	EstadoControlNominaGenerada   EstadoControlNomina = "GENERADA"
	EstadoControlNominaRegenerada EstadoControlNomina = "REGENERADA"
)

// IsValid reports whether the EstadoControlNomina has one of the permitted values.
func (e EstadoControlNomina) IsValid() bool {
	switch e {
	case EstadoControlNominaNoGenerada, EstadoControlNominaGenerada, EstadoControlNominaRegenerada:
		return true
	default:
		return false
	}
}

// EstadoPago represents the estado_pago_enum values.
type EstadoPago = string

const (
	EstadoPagoPagado    EstadoPago = "PAGADO"
	EstadoPagoPendiente EstadoPago = "PENDIENTE"
	EstadoPagoNoPago    EstadoPago = "NO_PAGO"
)

// EstadoPedido represents the estado_pedido_enum values.
// Valores válidos: "iniciado", "terminado", "cancelado".
type EstadoPedido = string

const (
	EstadoPedidoIniciado  EstadoPedido = "INICIADO"
	EstadoPedidoTerminado EstadoPedido = "TERMINADO"
	EstadoPedidoCancelado EstadoPedido = "CANCELADO"
)

// RolTrabajador represents the valid roles for a Trabajador.
type RolTrabajador string

const (
	RolAdmin        RolTrabajador = "ADMIN"
	RolMesero       RolTrabajador = "MESERO"
	RolCocinero     RolTrabajador = "COCINERO"
	RolDomiciliario RolTrabajador = "DOMICILIARIO"
)

// IsValid reports whether the role is one of the permitted values.
func (r RolTrabajador) IsValid() bool {
	switch r {
	case RolAdmin, RolMesero, RolCocinero, RolDomiciliario:
		return true
	default:
		return false
	}
}

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
