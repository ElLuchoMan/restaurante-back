package models

import "testing"

func TestRestauranteDiaTableUnique(t *testing.T) {
	r := RestauranteDia{}
	u := r.TableUnique()
	if len(u) != 1 || len(u[0]) != 2 || u[0][0] != "PKIDRestaurante" || u[0][1] != "Dia" {
		t.Fatalf("unexpected unique constraints: %#v", u)
	}
}
