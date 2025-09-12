package test

import (
	"restaurante/models"
	"testing"
)

func TestEstadoEnums(t *testing.T) {
	cases := map[string]models.DiaSemana{
		"Lunes":     models.DiaLunes,
		"Martes":    models.DiaMartes,
		"Miercoles": models.DiaMiercoles,
		"Jueves":    models.DiaJueves,
		"Viernes":   models.DiaViernes,
		"Sabado":    models.DiaSabado,
		"Domingo":   models.DiaDomingo,
	}
	for k, want := range cases {
		if want == "" {
			t.Fatalf("DiaSemana %s mismatched", k)
		}
	}
}
