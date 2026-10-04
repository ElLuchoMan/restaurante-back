package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestIncidenciaMarshalJSON(t *testing.T) {

	loc, err := time.LoadLocation("America/Bogota")
	if err != nil {
		loc = time.FixedZone("UTC-5", -5*60*60)
	}

	fecha := time.Date(2024, time.June, 3, 0, 0, 0, 0, loc)
	i := Incidencia{FECHA: fecha}

	b, err := json.Marshal(i)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(b, &data); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}

	if data["fechaIncidencia"] != "03-06-2024" {
		t.Errorf("expected fechaIncidencia 03-06-2024, got %v", data["fechaIncidencia"])
	}
}

func TestIncidenciaTableName(t *testing.T) {
	i := Incidencia{}
	if i.TableName() != "incidencia" {
		t.Errorf("expected table name incidencia, got %s", i.TableName())
	}
}
