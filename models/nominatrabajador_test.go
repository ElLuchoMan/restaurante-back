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

func TestNominaTrabajadorUnmarshalJSON_InvalidValues_Ignored(t *testing.T) {
    var n NominaTrabajador
    // Valores inválidos en IDs (strings), y seteo de campos opcionales
    payload := []byte(`{"documentoTrabajador":"oops","pk_id_nomina":"bad","sueldoBase":2000,"montoIncidencias":50,"detalles":"texto"}`)
    if err := json.Unmarshal(payload, &n); err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if n.PK_DOCUMENTO_TRABAJADOR != nil {
        t.Fatalf("expected trabajador to remain nil on invalid input, got %+v", n.PK_DOCUMENTO_TRABAJADOR)
    }
    if n.PK_ID_NOMINA != nil {
        t.Fatalf("expected nomina to remain nil on invalid input, got %+v", n.PK_ID_NOMINA)
    }
    if n.SUELDO_BASE != 2000 {
        t.Fatalf("expected sueldoBase 2000, got %d", n.SUELDO_BASE)
    }
    if n.MONTO_INCIDENCIAS == nil || *n.MONTO_INCIDENCIAS != 50 {
        t.Fatalf("expected montoIncidencias 50, got %+v", n.MONTO_INCIDENCIAS)
    }
    if n.DETALLES == nil || *n.DETALLES != "texto" {
        t.Fatalf("expected detalles 'texto', got %+v", n.DETALLES)
    }
}

func TestNominaTrabajadorUnmarshalJSON_MainObjectIDs(t *testing.T) {
    var n NominaTrabajador
    payload := []byte(`{"documentoTrabajador":{"documentoTrabajador":404},"nominaId":{"nominaId":22}}`)
    if err := json.Unmarshal(payload, &n); err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if n.PK_DOCUMENTO_TRABAJADOR == nil || n.PK_DOCUMENTO_TRABAJADOR.PK_DOCUMENTO_TRABAJADOR != 404 {
        t.Fatalf("unexpected trabajador: %+v", n.PK_DOCUMENTO_TRABAJADOR)
    }
    if n.PK_ID_NOMINA == nil || n.PK_ID_NOMINA.PK_ID_NOMINA != 22 {
        t.Fatalf("unexpected nomina: %+v", n.PK_ID_NOMINA)
    }
}

func TestNominaTrabajadorUnmarshalJSON_AltNumericIDs(t *testing.T) {
    var n NominaTrabajador
    payload := []byte(`{"pk_documento_trabajador":303,"pk_id_nomina":11}`)
    if err := json.Unmarshal(payload, &n); err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if n.PK_DOCUMENTO_TRABAJADOR == nil || n.PK_DOCUMENTO_TRABAJADOR.PK_DOCUMENTO_TRABAJADOR != 303 {
        t.Fatalf("unexpected trabajador: %+v", n.PK_DOCUMENTO_TRABAJADOR)
    }
    if n.PK_ID_NOMINA == nil || n.PK_ID_NOMINA.PK_ID_NOMINA != 11 {
        t.Fatalf("unexpected nomina: %+v", n.PK_ID_NOMINA)
    }
}

func TestNominaTrabajadorUnmarshalJSON_InvalidJSON(t *testing.T) {
    var n NominaTrabajador
    payload := []byte(`{`)
    if err := json.Unmarshal(payload, &n); err == nil {
        t.Fatalf("expected error for invalid JSON input")
    }
}

func TestNominaTrabajadorUnmarshalJSON_MainOverridesAlt(t *testing.T) {
    var n NominaTrabajador
    payload := []byte(`{"documentoTrabajador":111,"pk_documento_trabajador":999,"nominaId":12,"pk_id_nomina":34}`)
    if err := json.Unmarshal(payload, &n); err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if n.PK_DOCUMENTO_TRABAJADOR == nil || n.PK_DOCUMENTO_TRABAJADOR.PK_DOCUMENTO_TRABAJADOR != 111 {
        t.Fatalf("expected main field to win for trabajador, got %+v", n.PK_DOCUMENTO_TRABAJADOR)
    }
    if n.PK_ID_NOMINA == nil || n.PK_ID_NOMINA.PK_ID_NOMINA != 12 {
        t.Fatalf("expected main field to win for nomina, got %+v", n.PK_ID_NOMINA)
    }
}

func TestNominaTrabajadorUnmarshalJSON_MainNullPointersZeroValue(t *testing.T) {
    var n NominaTrabajador
    payload := []byte(`{"documentoTrabajador":null,"pk_documento_trabajador":202,"nominaId":null,"pk_id_nomina":303}`)
    if err := json.Unmarshal(payload, &n); err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    // La implementación actual intenta unmarshaling 'null' en struct y deja punteros a cero valor
    if n.PK_DOCUMENTO_TRABAJADOR == nil || n.PK_ID_NOMINA == nil {
        t.Fatalf("expected zero-valued pointers when main fields are null; got trabajador=%+v nomina=%+v", n.PK_DOCUMENTO_TRABAJADOR, n.PK_ID_NOMINA)
    }
    if n.PK_DOCUMENTO_TRABAJADOR.PK_DOCUMENTO_TRABAJADOR != 0 || n.PK_ID_NOMINA.PK_ID_NOMINA != 0 {
        t.Fatalf("expected zero IDs when main fields are null; got trabajador=%+v nomina=%+v", n.PK_DOCUMENTO_TRABAJADOR, n.PK_ID_NOMINA)
    }
}

func TestNominaTrabajadorUnmarshalJSON_NoIDsProvided(t *testing.T) {
    var n NominaTrabajador
    payload := []byte(`{"sueldoBase":500}`)
    if err := json.Unmarshal(payload, &n); err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if n.PK_DOCUMENTO_TRABAJADOR != nil || n.PK_ID_NOMINA != nil {
        t.Fatalf("expected both IDs to be nil, got trabajador=%+v nomina=%+v", n.PK_DOCUMENTO_TRABAJADOR, n.PK_ID_NOMINA)
    }
}

func TestNominaTrabajadorUnmarshalJSON_NonNumberNonObject(t *testing.T) {
    var n NominaTrabajador
    // Forzar que los helpers devuelvan nil (no número y no objeto)
    payload := []byte(`{"documentoTrabajador":"notNumberOrObject","nominaId":"alsoBad"}`)
    if err := json.Unmarshal(payload, &n); err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if n.PK_DOCUMENTO_TRABAJADOR != nil || n.PK_ID_NOMINA != nil {
        t.Fatalf("expected nil pointers when values are strings, got trabajador=%+v nomina=%+v", n.PK_DOCUMENTO_TRABAJADOR, n.PK_ID_NOMINA)
    }
}