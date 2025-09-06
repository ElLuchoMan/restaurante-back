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
		Fecha:     fecha,
		Estado:    EstadoDomicilioEnCamino,
		CreatedAt: created,
		UpdatedAt: updated,
		Direccion: "X",
		Telefono:  "Y",
		Entregado: false,
	}

	b, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(b, &data); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}

	for _, k := range []string{"fechaDomicilio", "estadoDomicilio", "createdAt", "updatedAt"} {
		if _, ok := data[k]; !ok {
			t.Errorf("expected key %s in JSON output", k)
		}
	}
}

func TestDomicilioTableName(t *testing.T) {
	d := Domicilio{}
	if d.TableName() != "domicilio" {
		t.Errorf("expected table name domicilio, got %s", d.TableName())
	}
}
