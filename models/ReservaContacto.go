package models

import "github.com/beego/beego/v2/client/orm"

type ReservaContacto struct {
	PKIDContacto       int64   `orm:"column(pk_id_contacto);pk;auto" json:"contactoId"`
	NombreCompleto     string  `orm:"column(nombre_completo);type(text)" json:"nombreCompleto"`
	Telefono           *string `orm:"column(telefono);type(text);null" json:"telefono,omitempty"`
	DocumentoContacto  *int64  `orm:"column(documento_contacto);null" json:"documentoContacto,omitempty"`
	PKDocumentoCliente *int64  `orm:"column(pk_documento_cliente);null" json:"documentoCliente,omitempty"`
}

func (r *ReservaContacto) TableName() string {
	return "reserva_contacto"
}

func init() {
	orm.RegisterModel(new(ReservaContacto))
}

// Validate ensures exactly one of DocumentoContacto or PKDocumentoCliente is set.
func (r ReservaContacto) Validate() bool {
	if (r.DocumentoContacto == nil && r.PKDocumentoCliente == nil) ||
		(r.DocumentoContacto != nil && r.PKDocumentoCliente != nil) {
		return false
	}
	return true
}
