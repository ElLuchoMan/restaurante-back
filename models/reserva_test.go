package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestReservaMarshalJSON(t *testing.T) {
	fecha := time.Date(2024, time.September, 12, 0, 0, 0, 0, time.UTC)
	estado := EstadoReservaConfirmada
	r := Reserva{FECHA: fecha, ESTADO_RESERVA: &estado}

	b, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(b, &data); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}

	if data["fechaReserva"] != "12-09-2024" {
		t.Errorf("expected fechaReserva 12-09-2024, got %v", data["fechaReserva"])
	}
	if data["estadoReserva"] != string(EstadoReservaConfirmada) {
		t.Errorf("expected estadoReserva %s, got %v", EstadoReservaConfirmada, data["estadoReserva"])
	}
}

func TestReservaTableName(t *testing.T) {
	r := Reserva{}
	if r.TableName() != "reserva" {
		t.Errorf("expected table name reserva, got %s", r.TableName())
	}
}
