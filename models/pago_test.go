package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestPagoMarshalJSON(t *testing.T) {
	fecha := time.Date(2024, time.August, 16, 0, 0, 0, 0, time.UTC)
	updated := time.Date(2024, time.August, 17, 12, 30, 45, 0, time.UTC)
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
