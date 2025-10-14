package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNominaMarshalJSON(t *testing.T) {

	loc, err := time.LoadLocation("America/Bogota")
	if err != nil {
		loc = time.FixedZone("UTC-5", -5*60*60)
	}

	fecha := time.Date(2024, time.July, 5, 0, 0, 0, 0, loc)
	n := Nomina{FECHA: fecha}

	b, err := json.Marshal(n)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(b, &data); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}

	if data["fechaNomina"] != "05-07-2024" {
		t.Errorf("expected fechaNomina 05-07-2024, got %v", data["fechaNomina"])
	}
}

func TestNominaTableName(t *testing.T) {
	n := Nomina{}
	if n.TableName() != "nomina" {
		t.Errorf("expected table name nomina, got %s", n.TableName())
	}
}

func TestNominaUnmarshalJSON_OK(t *testing.T) {
	var n Nomina
	payload := []byte(`{"fechaNomina":"2024-08-15"}`)
	if err := json.Unmarshal(payload, &n); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.FECHA.IsZero() {
		t.Fatalf("expected FECHA to be set")
	}
}

func TestNominaUnmarshalJSON_Error(t *testing.T) {
	var n Nomina
	payload := []byte(`{"fechaNomina":"invalid"}`)
	if err := json.Unmarshal(payload, &n); err == nil {
		t.Fatalf("expected error for invalid date format")
	}
}

func TestNominaUnmarshalJSON_EmptyFecha_NoChange(t *testing.T) {
	var n Nomina
	n.MONTO = 123
	payload := []byte(`{}`)
	if err := json.Unmarshal(payload, &n); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.FECHA.IsZero() == false {
		t.Fatalf("expected FECHA to remain zero when field is missing")
	}
	if n.MONTO != 123 {
		t.Fatalf("expected MONTO to remain unchanged, got %d", n.MONTO)
	}
}
