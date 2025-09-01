package models

import "testing"

func TestPrecioProductoHistTableUnique(t *testing.T) {
	p := PrecioProductoHist{}
	uniques := p.TableUnique()
	if len(uniques) != 1 || len(uniques[0]) != 2 || uniques[0][0] != "PKIDProducto" || uniques[0][1] != "FechaVigencia" {
		t.Fatalf("unexpected unique constraints: %#v", uniques)
	}
}
