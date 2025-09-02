package models

import (
	"encoding/json"
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

func TestProductoMarshalUnmarshalJSON(t *testing.T) {
	calorias := int64(100)
	img := "hello"
	p := Producto{
		PK_ID_PRODUCTO:     1,
		NOMBRE:             "Test",
		CALORIAS:           &calorias,
		DESCRIPCION:        func() *string { s := "desc"; return &s }(),
		PRECIO:             5,
		ESTADO_PRODUCTO:    EstadoProductoDisponible,
		IMAGEN:             img,
		CANTIDAD:           2,
		PK_ID_SUBCATEGORIA: &Subcategoria{PK_ID_SUBCATEGORIA: 3},
	}
	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if !strings.Contains(string(data), "aGVsbG8=") {
		t.Fatalf("expected base64 image in json: %s", string(data))
	}
	var out Producto
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !reflect.DeepEqual(out.IMAGEN, p.IMAGEN) {
		t.Fatalf("expected image %v, got %v", p.IMAGEN, out.IMAGEN)
	}
}
