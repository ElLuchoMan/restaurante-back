package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDomicilioMarshalJSON(t *testing.T) {
	fecha := time.Date(2024, time.August, 10, 0, 0, 0, 0, time.UTC)
	created := time.Date(2024, time.August, 9, 14, 30, 0, 0, time.UTC)
	updated := time.Date(2024, time.August, 11, 15, 45, 0, 0, time.UTC)
	d := Domicilio{
		FECHA:            fecha,
		ESTADO_DOMICILIO: EstadoDomicilioEnCamino,
		CREATED_AT:       created,
		UPDATED_AT:       updated,
	}

	b, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(b, &data); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}

	if data["fechaDomicilio"] != "10-08-2024" {
		t.Errorf("expected fechaDomicilio 10-08-2024, got %v", data["fechaDomicilio"])
	}
	if data["estado"] != EstadoDomicilioEnCamino {
		t.Errorf("expected estado %s, got %v", EstadoDomicilioEnCamino, data["estado"])
	}
	if data["createdAt"] != "09-08-2024 14:30:00" {
		t.Errorf("expected createdAt 09-08-2024 14:30:00, got %v", data["createdAt"])
	}
	if data["updatedAt"] != "11-08-2024 15:45:00" {
		t.Errorf("expected updatedAt 11-08-2024 15:45:00, got %v", data["updatedAt"])
	}
}

func TestDomicilioTableName(t *testing.T) {
	d := Domicilio{}
	if d.TableName() != "domicilio" {
		t.Errorf("expected table name domicilio, got %s", d.TableName())
	}
}
