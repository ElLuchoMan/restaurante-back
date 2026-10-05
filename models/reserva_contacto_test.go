package models

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestReservaContactoTableName(t *testing.T) {
	if (&ReservaContacto{}).TableName() != "reserva_contacto" {
		t.Fatal("tabla inesperada")
	}
}

func TestReservaContactoMarshalInvitado(t *testing.T) {
	doc := int64(55)
	tel := "300"
	b, err := json.Marshal(ReservaContacto{PKIDContacto: 2, NombreCompleto: "Ana", Telefono: &tel, DocumentoContacto: &doc})
	if err != nil {
		t.Fatal(err)
	}
	want := `{"contactoId":2,"nombreCompleto":"Ana","telefono":"300","documentoContacto":55}`
	if string(b) != want {
		t.Fatalf("got %s want %s", b, want)
	}
}

// El cliente registrado se responde solo con su documento: nunca con
// contraseña ni demás datos del cliente.
func TestReservaContactoMarshalClienteSinPassword(t *testing.T) {
	cli := &Cliente{PK_DOCUMENTO_CLIENTE: 9, NOMBRE: "Luis", PASSWORD: "secreto", CORREO: "l@x.co"}
	b, err := json.Marshal(&ReservaContacto{PKIDContacto: 1, NombreCompleto: "Luis", PKDocumentoCliente: cli})
	if err != nil {
		t.Fatal(err)
	}
	want := `{"contactoId":1,"nombreCompleto":"Luis","documentoCliente":{"documentoCliente":9}}`
	if string(b) != want {
		t.Fatalf("got %s want %s", b, want)
	}
	if strings.Contains(string(b), "secreto") || strings.Contains(string(b), "password") {
		t.Fatalf("filtración de contraseña: %s", b)
	}
}
