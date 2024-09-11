package models

import "github.com/beego/beego/v2/client/orm"

type Pedido struct {
	PK_ID_PEDIDO      int64  `orm:"pk" json:"PK_ID_PEDIDO"`
	FECHA             string `json:"FECHA"`
	HORA              string `json:"HORA"`
	DELIVERY          bool   `json:"DELIVERY"`
	ESTADO            string `json:"ESTADO"`
	PK_ID_DOMICILIO   *int64 `orm:"null" json:"PK_ID_DOMICILIO"`
	PK_ID_PAGO        *int64 `orm:"null" json:"PK_ID_PAGO"`
	PK_ID_ITEM_PEDIDO *int64 `orm:"null" json:"PK_ID_ITEM_PEDIDO"`
	PK_ID_RESTAURANTE *int64 `orm:"null" json:"PK_ID_RESTAURANTE"`
}

func (c *Pedido) TableName() string {
	return "PEDIDO"
}

func init() {
	orm.RegisterModel(new(Pedido))
}
