package models

import (
	"github.com/beego/beego/v2/client/orm"
)

type Domicilio struct {
	PK_ID_DOMICILIO int    `orm:"column(PK_ID_DOMICILIO);pk;auto" json:"PK_ID_DOMICILIO"`
	DIRECCION       string `orm:"column(DIRECCION);type(text)" json:"DIRECCION"`
	TELEFONO        string `orm:"column(TELEFONO);type(text)" json:"TELEFONO"`
	ESTADO_PAGO     string `orm:"column(ESTADO_PAGO);type(text)" json:"ESTADO_PAGO"`
	ENTREGADO       bool   `orm:"column(ENTREGADO);type(boolean)" json:"ENTREGADO"`
	FECHA           string `orm:"column(FECHA);type(date);type(text)" json:"FECHA"`
	OBSERVACIONES   string `orm:"column(OBSERVACIONES);type(text)" json:"OBSERVACIONES"`
}

func (d *Domicilio) TableName() string {
	return "DOMICILIO"
}

func init() {
	orm.RegisterModel(new(Domicilio))
}
