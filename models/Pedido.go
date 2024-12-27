package models

import (
	"github.com/beego/beego/v2/client/orm"
)

type Pedido struct {
	PK_ID_PEDIDO      int    `orm:"column(PK_ID_PEDIDO);pk"`
	FECHA             string `orm:"column(FECHA);type(date)"`
	HORA              string `orm:"column(HORA);type(time)"`
	DELIVERY          bool   `orm:"column(DELIVERY)"`
	ESTADO_PEDIDO     string `orm:"column(ESTADO_PEDIDO)"`
	PK_ID_DOMICILIO   *int   `orm:"column(PK_ID_DOMICILIO);null"`
	PK_ID_PAGO        *int   `orm:"column(PK_ID_PAGO);null"`
	PK_ID_RESTAURANTE *int   `orm:"column(PK_ID_RESTAURANTE);null"`
	UPDATED_AT        string `orm:"column(UPDATED_AT);type(date)" json:"UPDATED_AT"`
	UPDATED_BY        string `orm:"column(UPDATED_BY)" json:"UPDATED_BY"`
}

func (p *Pedido) TableName() string {
	return "PEDIDO"
}

func init() {
	orm.RegisterModel(new(Pedido))
}
