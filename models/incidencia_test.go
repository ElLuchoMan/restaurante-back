package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestIncidenciaMarshalJSON(t *testing.T) {
	fecha := time.Date(2024, time.June, 3, 0, 0, 0, 0, time.UTC)
	i := Incidencia{FECHA: fecha}

	b, err := json.Marshal(i)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(b, &data); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}

	if data["fechaincidencia"] != "03-06-2024" {
		t.Errorf("expected fechaincidencia 03-06-2024, got %v", data["fechaincidencia"])
	}
}

func TestIncidenciaTableName(t *testing.T) {
	i := Incidencia{}
	if i.TableName() != "incidencia" {
		t.Errorf("expected table name incidencia, got %s", i.TableName())
	}
}
