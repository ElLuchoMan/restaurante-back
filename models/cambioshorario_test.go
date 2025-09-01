package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestCambiosHorarioMarshalJSON(t *testing.T) {
	fecha := time.Date(2024, time.January, 2, 0, 0, 0, 0, time.UTC)
	c := CambiosHorario{FECHA: fecha}

	b, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(b, &data); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}

	if data["fechacambiohorario"] != "02-01-2024" {
		t.Errorf("expected fechacambiohorario 02-01-2024, got %v", data["fechacambiohorario"])
	}
}

func TestCambiosHorarioTableName(t *testing.T) {
	c := CambiosHorario{}
	if c.TableName() != "cambios_horario" {
		t.Errorf("expected table name cambios_horario, got %s", c.TableName())
	}
}
