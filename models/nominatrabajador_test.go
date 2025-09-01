package models

import (
	"reflect"
	"testing"
)

func TestNominaTrabajadorTableUnique(t *testing.T) {
	n := NominaTrabajador{}
	expected := [][]string{{"PK_DOCUMENTO_TRABAJADOR", "PK_ID_NOMINA"}}
	if got := n.TableUnique(); !reflect.DeepEqual(got, expected) {
		t.Errorf("expected %v, got %v", expected, got)
	}
}
