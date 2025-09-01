package models

import "testing"

func TestControlNominaValidEstado(t *testing.T) {
	c := ControlNomina{Estado: EstadoControlNominaGenerada}
	if !c.ValidEstado() {
		t.Errorf("expected estado %s to be valid", c.Estado)
	}
	c.Estado = EstadoControlNomina("OTRO")
	if c.ValidEstado() {
		t.Errorf("expected estado %s to be invalid", c.Estado)
	}
}

func TestControlNominaTableName(t *testing.T) {
	c := ControlNomina{}
	if c.TableName() != "control_nomina" {
		t.Errorf("expected table name control_nomina, got %s", c.TableName())
	}
}
