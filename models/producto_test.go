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

func TestProductoUnmarshalJSON_EmptyImageAndNoSubcategoria(t *testing.T) {
	payload := `{"productoId":2,"nombre":"X","precio":10,"estadoProducto":"DISPONIBLE","imagen":"","cantidad":1,"subcategoriaId":0}`
	var p Producto
	if err := json.Unmarshal([]byte(payload), &p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.IMAGEN != "" {
		t.Fatalf("expected empty image, got %q", p.IMAGEN)
	}
	if p.PK_ID_SUBCATEGORIA != nil {
		t.Fatalf("expected nil subcategoria")
	}
}

func TestProductoUnmarshalJSON_InvalidBase64(t *testing.T) {
	payload := `{"productoId":3,"nombre":"Y","precio":10,"estadoProducto":"DISPONIBLE","imagen":"***","cantidad":1}`
	var p Producto
	if err := json.Unmarshal([]byte(payload), &p); err == nil {
		t.Fatalf("expected base64 decode error")
	}
}

func TestProductoUnmarshalJSON_InvalidJSON(t *testing.T) {
	var p Producto
	if err := p.UnmarshalJSON([]byte("{invalid")); err == nil {
		t.Fatalf("expected error for invalid JSON")
	}
}

func TestProductoMarshalJSON_SubcategoriaNil(t *testing.T) {
	p := Producto{PK_ID_PRODUCTO: 10, NOMBRE: "Z", PRECIO: 1, ESTADO_PRODUCTO: EstadoProductoDisponible, CANTIDAD: 1}
	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if strings.Contains(string(data), "\"subcategoriaId\":0") {
		// Es aceptable. El objetivo es ejercer la ruta sin pánico.
	}
}
