package models

import (
	"github.com/beego/beego/v2/client/orm"
)

type Pedido struct {
	PK_ID_PEDIDO      int    `orm:"column(PK_ID_PEDIDO);pk"`
	FECHA             string `orm:"column(FECHA)"`
	HORA              string `orm:"column(HORA)"`
	DELIVERY          bool   `orm:"column(DELIVERY)"`
	ESTADO            string `orm:"column(ESTADO)"`
	PK_ID_DOMICILIO   *int   `orm:"column(PK_ID_DOMICILIO);null"`
	PK_ID_PAGO        *int   `orm:"column(PK_ID_PAGO);null"`
	PK_ID_ITEM_PEDIDO *int   `orm:"column(PK_ID_ITEM_PEDIDO);null"`
	PK_ID_RESTAURANTE *int   `orm:"column(PK_ID_RESTAURANTE);null"`
}

func (p *Pedido) TableName() string {
	return "PEDIDO"
}

func init() {
	orm.RegisterModel(new(Pedido))
}
