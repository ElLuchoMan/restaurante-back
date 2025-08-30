package models

import (
	"github.com/beego/beego/v2/client/orm"
)

type Cliente struct {
	PK_DOCUMENTO_CLIENTE int     `orm:"column(PK_DOCUMENTO_CLIENTE);pk" json:"documentoCliente"`
	NOMBRE               string  `orm:"column(NOMBRE);type(text)"        json:"nombre"`
	APELLIDO             string  `orm:"column(APELLIDO);type(text)"      json:"apellido"`
	CORREO               *string `orm:"column(CORREO);type(text)"        json:"correo"`
	DIRECCION            string  `orm:"column(DIRECCION);type(text)"     json:"direccion"`
	TELEFONO             string  `orm:"column(TELEFONO);type(text)"      json:"telefono"`
	OBSERVACIONES        *string `orm:"column(OBSERVACIONES);type(text)" json:"observaciones"`
	PASSWORD             string  `orm:"column(PASSWORD);type(text)"      json:"password"`
}

func (c *Cliente) TableName() string {
	return "CLIENTE"
}

func init() {
	orm.RegisterModel(new(Cliente))
}
