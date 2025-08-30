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
}
