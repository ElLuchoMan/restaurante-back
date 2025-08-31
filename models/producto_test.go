package models

import (
	"reflect"
	"strings"
	"testing"

	_ "github.com/lib/pq"
)

// TestProductoImagenFieldType ensures IMAGEN column is declared as bytea.
func TestProductoImagenFieldType(t *testing.T) {
	typ := reflect.TypeOf(Producto{})
	field, ok := typ.FieldByName("IMAGEN")
	if !ok {
		t.Fatal("IMAGEN field not found")
	}
	tag := field.Tag.Get("orm")
	if !strings.Contains(tag, "type(bytea)") {
		t.Fatalf("expected orm tag to contain type(bytea), got %q", tag)
	}
}
