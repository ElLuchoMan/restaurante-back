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

func TestNominaTrabajadorUnmarshalJSON_NumericIDs(t *testing.T) {
	var n NominaTrabajador
	payload := []byte(`{"documentoTrabajador":101,"nominaId":7,"sueldoBase":1000}`)
	if err := json.Unmarshal(payload, &n); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.PK_DOCUMENTO_TRABAJADOR == nil || n.PK_DOCUMENTO_TRABAJADOR.PK_DOCUMENTO_TRABAJADOR != 101 {
		t.Fatalf("unexpected trabajador: %+v", n.PK_DOCUMENTO_TRABAJADOR)
	}
	if n.PK_ID_NOMINA == nil || n.PK_ID_NOMINA.PK_ID_NOMINA != 7 {
		t.Fatalf("unexpected nomina: %+v", n.PK_ID_NOMINA)
	}
}

func TestNominaTrabajadorUnmarshalJSON_ObjectIDs(t *testing.T) {
	var n NominaTrabajador
	payload := []byte(`{"pk_documento_trabajador":{"documentoTrabajador":202},"pk_id_nomina":{"nominaId":9}}`)
	if err := json.Unmarshal(payload, &n); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.PK_DOCUMENTO_TRABAJADOR == nil || n.PK_DOCUMENTO_TRABAJADOR.PK_DOCUMENTO_TRABAJADOR != 202 {
		t.Fatalf("unexpected trabajador: %+v", n.PK_DOCUMENTO_TRABAJADOR)
	}
	if n.PK_ID_NOMINA == nil || n.PK_ID_NOMINA.PK_ID_NOMINA != 9 {
		t.Fatalf("unexpected nomina: %+v", n.PK_ID_NOMINA)
	}
}
