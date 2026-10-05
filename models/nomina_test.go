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

func TestNominaMarshalJSONCamposCompletos(t *testing.T) {
	n := Nomina{PK_ID_NOMINA: 3, FECHA: time.Date(2025, 1, 20, 12, 0, 0, 0, time.UTC), MONTO: 99, ESTADO_NOMINA: EstadoNominaPago}
	b, err := json.Marshal(n)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"nominaId":3,"fechaNomina":"20-01-2025","monto":99,"estadoNomina":"PAGO"}`
	if string(b) != want {
		t.Fatalf("got %s want %s", b, want)
	}
}
