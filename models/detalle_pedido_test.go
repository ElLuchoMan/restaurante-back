package models

import "testing"

func TestDetallePedidoTableUnique(t *testing.T) {
	d := DetallePedido{}
	u := d.TableUnique()
	if len(u) != 1 || len(u[0]) != 2 || u[0][0] != "PKIDPedido" || u[0][1] != "PKIDProducto" {
		t.Fatalf("unexpected unique constraints: %#v", u)
	}
}
