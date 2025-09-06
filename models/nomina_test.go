package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNominaMarshalJSON(t *testing.T) {
	fecha := time.Date(2024, time.July, 5, 0, 0, 0, 0, time.UTC)
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
