package models

import (
	"github.com/beego/beego/v2/client/orm"
)

type Cliente struct {
	PK_DOCUMENTO_CLIENTE int     `orm:"column(PK_DOCUMENTO_CLIENTE);pk" json:"PK_DOCUMENTO_CLIENTE"`
	NOMBRE               string  `orm:"column(NOMBRE);type(text)" json:"NOMBRE"`
	APELLIDO             string  `orm:"column(APELLIDO);type(text)" json:"APELLIDO"`
	DIRECCION            string  `orm:"column(DIRECCION);type(text)" json:"DIRECCION"`
	TELEFONO             string  `orm:"column(TELEFONO);type(text)" json:"TELEFONO"`
	OBSERVACIONES        *string `orm:"column(OBSERVACIONES);type(text)" json:"OBSERVACIONES"`
	PASSWORD             string  `orm:"column(PASSWORD);type(text)" json:"PASSWORD"`
}

func (c *Cliente) TableName() string {
	return "CLIENTE"
}

func init() {
	orm.RegisterModel(new(Cliente))
}
