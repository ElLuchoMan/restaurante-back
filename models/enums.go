package models

type EstadoDomicilio = string

const (
	EstadoDomicilioPendiente EstadoDomicilio = "PENDIENTE"
	EstadoDomicilioEnCamino  EstadoDomicilio = "EN_CAMINO"
	EstadoDomicilioEntregado EstadoDomicilio = "ENTREGADO"
)

type EstadoNomina = string

const (
	EstadoNominaPago   EstadoNomina = "PAGO"
	EstadoNominaNoPago EstadoNomina = "NO_PAGO"
)

type EstadoPago = string

const (
	EstadoPagoPagado    EstadoPago = "PAGADO"
	EstadoPagoPendiente EstadoPago = "PENDIENTE"
	EstadoPagoNoPago    EstadoPago = "NO_PAGO"
)

type EstadoPedido = string

const (
	EstadoPedidoIniciado      EstadoPedido = "INICIADO"
	EstadoPedidoEnPreparacion EstadoPedido = "EN_PREPARACION"
	EstadoPedidoListo         EstadoPedido = "LISTO"
	EstadoPedidoTerminado     EstadoPedido = "TERMINADO"
	EstadoPedidoCancelado     EstadoPedido = "CANCELADO"
)

type EstadoProducto = string

const (
	EstadoProductoDisponible   EstadoProducto = "DISPONIBLE"
	EstadoProductoNoDisponible EstadoProducto = "NO_DISPONIBLE"
)

type EstadoReserva = string

const (
	EstadoReservaPendiente  EstadoReserva = "PENDIENTE"
	EstadoReservaConfirmada EstadoReserva = "CONFIRMADA"
	EstadoReservaCancelada  EstadoReserva = "CANCELADA"
	EstadoReservaCumplida   EstadoReserva = "CUMPLIDA"
)

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

type RolTrabajador string

const (
	RolAdministrador RolTrabajador = "Administrador"
	RolMesero        RolTrabajador = "Mesero"
	RolCocinero      RolTrabajador = "Cocinero"
	RolDomiciliario  RolTrabajador = "Domiciliario"
	RolOficiosVarios RolTrabajador = "Oficios_varios"
)

func (r RolTrabajador) IsValid() bool {
	switch r {
	case RolAdministrador, RolMesero, RolCocinero, RolDomiciliario, RolOficiosVarios:
		return true
	}
	return false
}

type EstadoControlNomina string

const (
	EstadoControlNominaNoGenerada EstadoControlNomina = "NO GENERADA"
	EstadoControlNominaGenerada   EstadoControlNomina = "GENERADA"
	EstadoControlNominaReGenerada EstadoControlNomina = "REGENERADA"
)

func (e EstadoControlNomina) IsValid() bool {
	switch e {
	case EstadoControlNominaNoGenerada, EstadoControlNominaGenerada, EstadoControlNominaReGenerada:
		return true
	}
	return false
}
