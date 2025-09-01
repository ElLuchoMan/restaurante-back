package models

import (
	"testing"
	"time"
)

func TestHorarioTrabajadorTableUnique(t *testing.T) {
	h := HorarioTrabajador{}
	u := h.TableUnique()
	if len(u) != 1 || len(u[0]) != 2 || u[0][0] != "PK_DOCUMENTO_TRABAJADOR" || u[0][1] != "DIA" {
		t.Fatalf("unexpected unique constraints: %#v", u)
	}
}

func TestHorarioTrabajadorValidHours(t *testing.T) {
	h := HorarioTrabajador{
		HORA_INICIO: time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC),
		HORA_FIN:    time.Date(0, 1, 1, 10, 0, 0, 0, time.UTC),
	}
	if !h.ValidHours() {
		t.Fatalf("expected valid hours")
	}
	h.HORA_FIN = time.Date(0, 1, 1, 8, 0, 0, 0, time.UTC)
	if h.ValidHours() {
		t.Fatalf("expected invalid hours")
	}
}
