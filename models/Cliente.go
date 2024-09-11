package models

import (
	"github.com/beego/beego/v2/client/orm"
)

type Cliente struct {
	PK_DOCUMENTO_CLIENTE int     `orm:"column(PK_DOCUMENTO_CLIENTE);pk"`
	NOMBRE               string  `orm:"column(NOMBRE)"`
	APELLIDO             string  `orm:"column(APELLIDO)"`
	DIRECCION            string  `orm:"column(DIRECCION)"`
	TELEFONO             string  `orm:"column(TELEFONO)"`
	OBSERVACIONES        *string `orm:"column(OBSERVACIONES);null"`
	PASSWORD             string  `orm:"column(PASSWORD)"`
}

func (c *Cliente) TableName() string {
	return "CLIENTE"
}

func init() {
	orm.RegisterModel(new(Cliente))
}
