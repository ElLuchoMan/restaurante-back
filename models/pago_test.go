package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestPagoMarshalJSON(t *testing.T) {

	loc, err := time.LoadLocation("America/Bogota")
	if err != nil {
		loc = time.FixedZone("UTC-5", -5*60*60)
	}

	fecha := time.Date(2024, time.August, 16, 0, 0, 0, 0, loc)
	updated := time.Date(2024, time.August, 17, 12, 30, 45, 0, loc)
	p := Pago{
		FECHA:      fecha,
		UPDATED_AT: updated,
	}

	b, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(b, &data); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}

	if data["fechaPago"] != "16-08-2024" {
		t.Errorf("expected fechaPago 16-08-2024, got %v", data["fechaPago"])
	}
	if data["updatedAt"] != "17-08-2024 12:30:45" {
		t.Errorf("expected updatedAt 17-08-2024 12:30:45, got %v", data["updatedAt"])
	}
}

func TestPagoMarshalJSONNilUpdatedBy(t *testing.T) {

	loc, err := time.LoadLocation("America/Bogota")
	if err != nil {
		loc = time.FixedZone("UTC-5", -5*60*60)
	}

	fecha := time.Date(2024, time.August, 16, 0, 0, 0, 0, loc)
	updated := time.Date(2024, time.August, 17, 12, 30, 45, 0, loc)
	p := Pago{
		FECHA:      fecha,
		UPDATED_AT: updated,
		UPDATED_BY: nil,
	}

	b, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(b, &data); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}

	if _, ok := data["updatedBy"]; ok {
		t.Errorf("expected updatedBy to be omitted, got %v", data["updatedBy"])
	}
}
