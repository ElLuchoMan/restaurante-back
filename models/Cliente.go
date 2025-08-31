package models

import (
	"github.com/beego/beego/v2/client/orm"
)

type Cliente struct {
	PK_DOCUMENTO_CLIENTE int64   `orm:"column(pk_documento_cliente);pk" json:"documentoCliente"`
	NOMBRE               string  `orm:"column(nombre);type(text)"        json:"nombre"`
	APELLIDO             string  `orm:"column(apellido);type(text)"      json:"apellido"`
	CORREO               *string `orm:"column(correo);type(text)"        json:"correo"`
	DIRECCION            string  `orm:"column(direccion);type(text)"     json:"direccion"`
	TELEFONO             string  `orm:"column(telefono);type(text)"      json:"telefono"`
	OBSERVACIONES        *string `orm:"column(observaciones);type(text)" json:"observaciones"`
	PASSWORD             string  `orm:"column(password);type(text)"      json:"password"`
}

func (c *Cliente) TableName() string {
	return "cliente"
}

func init() {
	orm.RegisterModel(new(Cliente))
}
