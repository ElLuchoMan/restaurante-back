package models

import "github.com/beego/beego/v2/client/orm"

type ReservaContacto struct {
	PKIDContacto       int64    `orm:"column(pk_id_contacto);pk;auto" json:"contactoId"`
	NombreCompleto     string   `orm:"column(nombre_completo);type(text)" json:"nombreCompleto"`
	Telefono           *string  `orm:"column(telefono);type(text);null" json:"telefono,omitempty"`
	DocumentoContacto  *int64   `orm:"column(documento_contacto);null" json:"documentoContacto,omitempty"`
	PKDocumentoCliente *Cliente `orm:"column(pk_documento_cliente);rel(fk);null" json:"documentoCliente,omitempty" swaggertype:"integer"`
}

func (r *ReservaContacto) TableName() string {
	return "reserva_contacto"
}

// Valid ensures only one of DocumentoContacto or PKDocumentoCliente is set
func (r *ReservaContacto) Valid() bool {
	return (r.DocumentoContacto == nil) != (r.PKDocumentoCliente == nil)
}

func init() {
	orm.RegisterModel(new(ReservaContacto))
}
