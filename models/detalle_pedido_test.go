package models

import (
	"reflect"
	"strings"
	"testing"
)

func TestDetallePedidoTableUnique(t *testing.T) {
	d := DetallePedido{}
	u := d.TableUnique()
	if len(u) != 1 || len(u[0]) != 2 || u[0][0] != "PKIDPedido" || u[0][1] != "PKIDProducto" {
		t.Fatalf("unexpected unique constraints: %#v", u)
	}
}

func TestDetallePedidoForeignKeys(t *testing.T) {
	typ := reflect.TypeOf(DetallePedido{})
	for _, name := range []string{"PKIDPedido", "PKIDProducto"} {
		field, ok := typ.FieldByName(name)
		if !ok {
			t.Fatalf("field %s not found", name)
		}
		tag := field.Tag.Get("orm")
		if !strings.Contains(tag, "rel(fk)") {
			t.Errorf("%s missing rel(fk) tag: %q", name, tag)
		}
	}
}
