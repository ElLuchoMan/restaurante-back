package models

import (
	"encoding/json"
	"testing"
)

func TestNominaTrabajadorTableUnique(t *testing.T) {
	n := NominaTrabajador{}
	expected := [][]string{{"PK_DOCUMENTO_TRABAJADOR", "PK_ID_NOMINA"}}
	if got := n.TableUnique(); len(got) != 1 || got[0][0] != expected[0][0] || got[0][1] != expected[0][1] {
		t.Errorf("expected %v, got %v", expected, got)
	}
}

func TestNominaTrabajadorTableName(t *testing.T) {
	var n NominaTrabajador
	if got := n.TableName(); got != "nomina_trabajador" {
		t.Fatalf("expected table name 'nomina_trabajador', got %q", got)
	}
}

func TestNominaTrabajadorItemJSONShape(t *testing.T) {
	b, err := json.Marshal(NominaTrabajadorItem{PK_ID_NOMINA_TRABAJADOR: 1, SUELDO_BASE: 2, MONTO_INCIDENCIAS: 3, DETALLES: "d", PK_DOCUMENTO_TRABAJADOR: 4, PK_ID_NOMINA: 5})
	if err != nil {
		t.Fatal(err)
	}
	want := `{"nominaTrabajadorId":1,"sueldoBase":2,"montoIncidencias":3,"detalles":"d","documentoTrabajador":4,"nominaId":5}`
	if string(b) != want {
		t.Fatalf("got %s want %s", b, want)
	}
}

func TestNominaTrabajadorDetalleJSONShape(t *testing.T) {
	b, err := json.Marshal(NominaTrabajadorDetalle{PK_ID_NOMINA_TRABAJADOR: 1, NOMBRE: "Ana", APELLIDO: "Gil"})
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"nominaTrabajadorId", "sueldoBase", "montoIncidencias", "detalles", "documentoTrabajador", "nominaId", "nombre", "apellido"} {
		if _, ok := m[k]; !ok {
			t.Errorf("falta %s en %s", k, b)
		}
	}
}
