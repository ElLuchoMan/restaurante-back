package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTrabajadorMarshalJSON(t *testing.T) {
	nacimiento := time.Date(1990, time.March, 1, 0, 0, 0, 0, time.UTC)
	ingreso := time.Date(2023, time.February, 10, 9, 15, 30, 0, time.UTC)
	retiro := time.Date(2024, time.April, 5, 18, 0, 0, 0, time.UTC)
	tr := Trabajador{
		FECHA_NACIMIENTO: &nacimiento,
		FECHA_INGRESO:    ingreso,
		FECHA_RETIRO:     &retiro,
	}

	b, err := json.Marshal(tr)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(b, &data); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}

	if data["fechaNacimiento"] != "01-03-1990" {
		t.Errorf("expected fechaNacimiento 01-03-1990, got %v", data["fechaNacimiento"])
	}
	if data["fechaIngreso"] != "10-02-2023 09:15:30" {
		t.Errorf("expected fechaIngreso 10-02-2023 09:15:30, got %v", data["fechaIngreso"])
	}
	if data["fechaRetiro"] != "05-04-2024 18:00:00" {
		t.Errorf("expected fechaRetiro 05-04-2024 18:00:00, got %v", data["fechaRetiro"])
	}
}

func TestTrabajadorTableName(t *testing.T) {
	tr := Trabajador{}
	if tr.TableName() != "trabajador" {
		t.Errorf("expected table name trabajador, got %s", tr.TableName())
	}
}

func TestRolTrabajadorIsValid(t *testing.T) {
	if !RolAdministrador.IsValid() {
		t.Fatalf("RolAdministrador should be valid")
	}
	invalid := RolTrabajador("Chef")
	if invalid.IsValid() {
		t.Fatalf("unexpected valid role for %s", invalid)
	}
}

func TestTrabajadorMarshalJSONNil(t *testing.T) {
	ingreso := time.Date(2023, time.February, 10, 9, 15, 30, 0, time.UTC)
	tr := Trabajador{FECHA_INGRESO: ingreso}

	b, err := json.Marshal(tr)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(b, &data); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}

	if _, ok := data["fechaNacimiento"]; ok {
		t.Errorf("expected fechaNacimiento to be omitted")
	}
	if data["fechaIngreso"] != "10-02-2023 09:15:30" {
		t.Errorf("expected fechaIngreso 10-02-2023 09:15:30, got %v", data["fechaIngreso"])
	}
	if _, ok := data["fechaRetiro"]; ok {
		t.Errorf("expected fechaRetiro to be omitted")
	}
}
