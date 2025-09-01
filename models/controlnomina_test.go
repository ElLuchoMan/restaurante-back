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

type dummyOrmer struct{}

func (d dummyOrmer) Insert(interface{}) (int64, error)            { return 1, nil }
func (d dummyOrmer) Update(interface{}, ...string) (int64, error) { return 1, nil }

func TestControlNominaInsertUpdateValidation(t *testing.T) {
	o := dummyOrmer{}
	c := ControlNomina{Estado: EstadoControlNominaGenerada}
	if _, err := c.Insert(o); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c.Estado = EstadoControlNomina("BAD")
	if _, err := c.Insert(o); err == nil {
		t.Fatalf("expected error for invalid estado")
	}
	c.Estado = EstadoControlNominaGenerada
	if _, err := c.Update(o); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c.Estado = EstadoControlNomina("BAD")
	if _, err := c.Update(o); err == nil {
		t.Fatalf("expected error for invalid estado on update")
	}
}
