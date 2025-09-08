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
// Valores válidos: INICIADO, EN_PREPARACION, LISTO, TERMINADO, CANCELADO.
type EstadoPedido = string

const (
	EstadoPedidoIniciado      EstadoPedido = "INICIADO"
	EstadoPedidoEnPreparacion EstadoPedido = "EN_PREPARACION"
	EstadoPedidoListo         EstadoPedido = "LISTO"
	EstadoPedidoTerminado     EstadoPedido = "TERMINADO"
	EstadoPedidoCancelado     EstadoPedido = "CANCELADO"
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
	DiaLunes     DiaSemana = "Lunes"
	DiaMartes    DiaSemana = "Martes"
	DiaMiercoles DiaSemana = "Miércoles"
	DiaJueves    DiaSemana = "Jueves"
	DiaViernes   DiaSemana = "Viernes"
	DiaSabado    DiaSemana = "Sábado"
	DiaDomingo   DiaSemana = "Domingo"
)

// RolTrabajador represents valid roles for trabajador. Values are case-sensitive.
type RolTrabajador string

const (
	RolAdministrador RolTrabajador = "Administrador"
	RolMesero        RolTrabajador = "Mesero"
	RolCocinero      RolTrabajador = "Cocinero"
	RolDomiciliario  RolTrabajador = "Domiciliario"
	RolOficiosVarios RolTrabajador = "Oficios_varios"
)

// IsValid reports whether the role is permitted for trabajador
func (r RolTrabajador) IsValid() bool {
	switch r {
	case RolAdministrador, RolMesero, RolCocinero, RolDomiciliario, RolOficiosVarios:
		return true
	}
	return false
}

// EstadoControlNomina represents valid values for control_nomina.estado
type EstadoControlNomina string

const (
	EstadoControlNominaNoGenerada EstadoControlNomina = "NO GENERADA"
	EstadoControlNominaGenerada   EstadoControlNomina = "GENERADA"
	EstadoControlNominaReGenerada EstadoControlNomina = "REGENERADA"
)

// IsValid reports whether the estado is permitted for control_nomina
func (e EstadoControlNomina) IsValid() bool {
	switch e {
	case EstadoControlNominaNoGenerada, EstadoControlNominaGenerada, EstadoControlNominaReGenerada:
		return true
	}
	return false
}
