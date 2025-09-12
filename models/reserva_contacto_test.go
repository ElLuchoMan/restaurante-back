package models

import "testing"

func TestReservaContactoValid(t *testing.T) {
	r := ReservaContacto{DocumentoContacto: new(int64)}
	if !r.Valid() {
		t.Fatalf("expected valid when only DocumentoContacto is set")
	}
	r = ReservaContacto{PKDocumentoCliente: &Cliente{PK_DOCUMENTO_CLIENTE: 1}}
	if !r.Valid() {
		t.Fatalf("expected valid when only PKDocumentoCliente is set")
	}
	r = ReservaContacto{}
	if r.Valid() {
		t.Fatalf("expected invalid when both nil")
	}
	doc := int64(1)
	cli := Cliente{PK_DOCUMENTO_CLIENTE: 2}
	r = ReservaContacto{DocumentoContacto: &doc, PKDocumentoCliente: &cli}
	if r.Valid() {
		t.Fatalf("expected invalid when both set")
	}
}
