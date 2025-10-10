package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestReservaMarshalJSON(t *testing.T) {
	// Cargar zona horaria de Bogotá para los tests
	loc, err := time.LoadLocation("America/Bogota")
	if err != nil {
		loc = time.FixedZone("UTC-5", -5*60*60)
	}

	fecha := time.Date(2024, time.September, 12, 0, 0, 0, 0, loc)
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

func TestReservaMarshalJSONCreatedUpdatedBy(t *testing.T) {
	r := Reserva{}

	b, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(b, &data); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}

	if _, ok := data["createdBy"]; ok {
		t.Errorf("createdBy should be omitted when nil")
	}
	if _, ok := data["updatedBy"]; ok {
		t.Errorf("updatedBy should be omitted when nil")
	}

	cb, ub := "creator", "updater"
	r.CREATED_BY = &cb
	r.UPDATED_BY = &ub

	b, err = json.Marshal(r)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}

	data = map[string]interface{}{}
	if err := json.Unmarshal(b, &data); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}

	if data["createdBy"] != cb {
		t.Errorf("expected createdBy %s, got %v", cb, data["createdBy"])
	}
	if data["updatedBy"] != ub {
		t.Errorf("expected updatedBy %s, got %v", ub, data["updatedBy"])
	}
}

func TestReservaTableName(t *testing.T) {
	r := Reserva{}
	if r.TableName() != "reserva" {
		t.Errorf("expected table name reserva, got %s", r.TableName())
	}
}
