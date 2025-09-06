package test

import (
	"restaurante/models"
	"testing"
)

// TestEstadoEnums verifies that the enumerated constants map to the expected string values.
func TestEstadoEnums(t *testing.T) {
	reservas := map[string]models.EstadoReserva{
		"PENDIENTE":  models.EstadoReservaPendiente,
		"CONFIRMADA": models.EstadoReservaConfirmada,
		"CANCELADA":  models.EstadoReservaCancelada,
		"CUMPLIDA":   models.EstadoReservaCumplida,
	}
	for expect, val := range reservas {
		if string(val) != expect {
			t.Errorf("EstadoReserva %s mismatched", expect)
		}
	}

	nominas := map[string]models.EstadoNomina{
		"PAGO":    models.EstadoNominaPago,
		"NO_PAGO": models.EstadoNominaNoPago,
	}
	for expect, val := range nominas {
		if string(val) != expect {
			t.Errorf("EstadoNomina %s mismatched", expect)
		}
	}

	domicilios := map[string]models.EstadoDomicilio{
		"PENDIENTE": models.EstadoDomicilioPendiente,
		"EN_CAMINO": models.EstadoDomicilioEnCamino,
		"ENTREGADO": models.EstadoDomicilioEntregado,
	}
	for expect, val := range domicilios {
		if string(val) != expect {
			t.Errorf("EstadoDomicilio %s mismatched", expect)
		}
	}

	pagos := map[string]models.EstadoPago{
		"PAGADO":    models.EstadoPagoPagado,
		"PENDIENTE": models.EstadoPagoPendiente,
		"NO_PAGO":   models.EstadoPagoNoPago,
	}
	for expect, val := range pagos {
		if string(val) != expect {
			t.Errorf("EstadoPago %s mismatched", expect)
		}
	}

	pedidos := map[string]models.EstadoPedido{
		"INICIADO":  models.EstadoPedidoIniciado,
		"TERMINADO": models.EstadoPedidoTerminado,
		"CANCELADO": models.EstadoPedidoCancelado,
	}
	for expect, val := range pedidos {
		if string(val) != expect {
			t.Errorf("EstadoPedido %s mismatched", expect)
		}
	}

	productos := map[string]models.EstadoProducto{
		"DISPONIBLE":    models.EstadoProductoDisponible,
		"NO_DISPONIBLE": models.EstadoProductoNoDisponible,
	}
	for expect, val := range productos {
		if string(val) != expect {
			t.Errorf("EstadoProducto %s mismatched", expect)
		}
	}

	dias := map[string]models.DiaSemana{
		"Lunes":     models.DiaLunes,
		"Martes":    models.DiaMartes,
		"Miercoles": models.DiaMiercoles,
		"Jueves":    models.DiaJueves,
		"Viernes":   models.DiaViernes,
		"Sabado":    models.DiaSabado,
		"Domingo":   models.DiaDomingo,
	}
	for expect, val := range dias {
		if string(val) != expect {
			t.Errorf("DiaSemana %s mismatched", expect)
		}
	}
}
