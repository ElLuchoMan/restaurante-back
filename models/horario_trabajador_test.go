package models

import (
	"encoding/json"
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
		HORA_INICIO: time.Date(1, 1, 1, 9, 0, 0, 0, time.UTC),
		HORA_FIN:    time.Date(1, 1, 1, 10, 0, 0, 0, time.UTC),
	}
	if !h.ValidHours() {
		t.Fatalf("expected valid hours")
	}
	h.HORA_FIN = time.Date(1, 1, 1, 8, 0, 0, 0, time.UTC)
	if h.ValidHours() {
		t.Fatalf("expected invalid hours")
	}
}

func TestHorarioTrabajadorMarshalJSON(t *testing.T) {
	doc := int64(123)
	h := HorarioTrabajador{
		PK_DOCUMENTO_TRABAJADOR: &Trabajador{PK_DOCUMENTO_TRABAJADOR: doc},
		DIA:                     DiaLunes,
		HORA_INICIO:             time.Date(1, 1, 1, 8, 30, 0, 0, time.UTC),
		HORA_FIN:                time.Date(1, 1, 1, 17, 45, 0, 0, time.UTC),
	}
	b, err := json.Marshal(h)
	if err != nil {
		t.Fatalf("MarshalJSON error: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("Unmarshal back error: %v", err)
	}
	if m["documentoTrabajador"].(float64) != float64(doc) {
		t.Fatalf("expected documento %d, got %v", doc, m["documentoTrabajador"])
	}
	if m["dia"].(string) == "" || m["horaInicio"].(string) == "" || m["horaFin"].(string) == "" {
		t.Fatalf("expected formatted fields present: %v", m)
	}

	// Sin trabajador (nil pointer path)
	h.PK_DOCUMENTO_TRABAJADOR = nil
	b, err = json.Marshal(h)
	if err != nil {
		t.Fatalf("MarshalJSON error (nil): %v", err)
	}
}
