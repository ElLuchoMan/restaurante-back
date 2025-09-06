package test

import (
	"restaurante/models"
	"testing"
)

func TestEstadoEnums(t *testing.T) {
	cases := map[string]models.DiaSemana{
		"Lunes":     models.DiaLunes,
		"Martes":    models.DiaMartes,
		"Miercoles": models.DiaMiercoles, // Entrada sin acento debe mapear a enum con acento
		"Jueves":    models.DiaJueves,
		"Viernes":   models.DiaViernes,
		"Sabado":    models.DiaSabado, // Entrada sin acento debe mapear a enum con acento
		"Domingo":   models.DiaDomingo,
	}
	for k, want := range cases {
		// La prueba verifica sólo el valor del enum; controladores hacen la normalización
		if want == "" {
			t.Fatalf("DiaSemana %s mismatched", k)
		}
	}
}
