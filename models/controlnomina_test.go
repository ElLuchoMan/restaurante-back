package models

import "testing"

func TestControlNominaTableName(t *testing.T) {
	c := ControlNomina{}
	if c.TableName() != "control_nomina" {
		t.Errorf("expected table name control_nomina, got %s", c.TableName())
	}
}
