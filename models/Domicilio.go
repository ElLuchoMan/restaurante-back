package models

import (
	"github.com/beego/beego/v2/client/orm"
)

type Domicilio struct {
	PK_ID_DOMICILIO int    `orm:"column(PK_ID_DOMICILIO);pk;auto"`
	DIRECCION       string `orm:"column(DIRECCION)"`
	TELEFONO        string `orm:"column(TELEFONO)"`
	ESTADO_PAGO     string `orm:"column(ESTADO_PAGO)"`
	ENTREGADO       bool   `orm:"column(ENTREGADO)"`
	FECHA           string `orm:"column(FECHA);type(date)"`
	OBSERVACIONES   string `orm:"column(OBSERVACIONES)"`
}

func (d *Domicilio) TableName() string {
	return "DOMICILIO"
}

func init() {
	orm.RegisterModel(new(Domicilio))
}
