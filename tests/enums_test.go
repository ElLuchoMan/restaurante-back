package test

import (
	"restaurante/models"
	"testing"
)

// TestEstadoEnums verifies that the enumerated constants map to the expected string values.
func TestEstadoEnums(t *testing.T) {
	reservas := map[string]models.EstadoReserva{
		"pendiente":  models.EstadoReservaPendiente,
		"confirmada": models.EstadoReservaConfirmada,
		"cancelada":  models.EstadoReservaCancelada,
		"cumplida":   models.EstadoReservaCumplida,
	}
	for expect, val := range reservas {
		if string(val) != expect {
			t.Errorf("EstadoReserva %s mismatched", expect)
		}
	}

	nominas := map[string]models.EstadoNomina{
		"pago":    models.EstadoNominaPago,
		"no pago": models.EstadoNominaNoPago,
	}
	for expect, val := range nominas {
		if string(val) != expect {
			t.Errorf("EstadoNomina %s mismatched", expect)
		}
	}

	domicilios := map[string]models.EstadoDomicilio{
		"pendiente": models.EstadoDomicilioPendiente,
		"en camino": models.EstadoDomicilioEnCamino,
		"entregado": models.EstadoDomicilioEntregado,
	}
	for expect, val := range domicilios {
		if string(val) != expect {
			t.Errorf("EstadoDomicilio %s mismatched", expect)
		}
	}

	pagos := map[string]models.EstadoPago{
		"pagado":    models.EstadoPagoPagado,
		"pendiente": models.EstadoPagoPendiente,
		"no pago":   models.EstadoPagoNoPago,
	}
	for expect, val := range pagos {
		if string(val) != expect {
			t.Errorf("EstadoPago %s mismatched", expect)
		}
	}

	pedidos := map[string]models.EstadoPedido{
		"iniciado":  models.EstadoPedidoIniciado,
		"terminado": models.EstadoPedidoTerminado,
	}
	for expect, val := range pedidos {
		if string(val) != expect {
			t.Errorf("EstadoPedido %s mismatched", expect)
		}
	}

	productos := map[string]models.EstadoProducto{
		"disponible":    models.EstadoProductoDisponible,
		"no disponible": models.EstadoProductoNoDisponible,
	}
	for expect, val := range productos {
		if string(val) != expect {
			t.Errorf("EstadoProducto %s mismatched", expect)
		}
	}

	dias := map[string]models.DiaSemana{
		"lunes":     models.DiaLunes,
		"martes":    models.DiaMartes,
		"miercoles": models.DiaMiercoles,
		"jueves":    models.DiaJueves,
		"viernes":   models.DiaViernes,
		"sabado":    models.DiaSabado,
		"domingo":   models.DiaDomingo,
	}
	for expect, val := range dias {
		if string(val) != expect {
			t.Errorf("DiaSemana %s mismatched", expect)
		}
	}
}
