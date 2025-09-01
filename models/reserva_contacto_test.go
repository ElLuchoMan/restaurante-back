package models

import "testing"

func TestReservaContactoValidate(t *testing.T) {
	rc := ReservaContacto{}
	if rc.Validate() {
		t.Fatalf("expected false when both identifiers are nil")
	}
	doc := int64(1)
	cli := int64(2)
	rc.DocumentoContacto = &doc
	rc.PKDocumentoCliente = &cli
	if rc.Validate() {
		t.Fatalf("expected false when both identifiers are set")
	}
	rc = ReservaContacto{DocumentoContacto: &doc}
	if !rc.Validate() {
		t.Fatalf("expected true when only DocumentoContacto is set")
	}
	rc = ReservaContacto{PKDocumentoCliente: &cli}
	if !rc.Validate() {
		t.Fatalf("expected true when only PKDocumentoCliente is set")
	}
}
