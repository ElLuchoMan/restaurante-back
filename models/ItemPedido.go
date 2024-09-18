package models

import "github.com/beego/beego/v2/client/orm"

type ItemPedido struct {
	PK_ID_ITEM_PEDIDO int64  `orm:"column(PK_ID_ITEM_PEDIDO);pk;auto" json:"PK_ID_ITEM_PEDIDO"`
	CANTIDAD          int64  `orm:"column(CANTIDAD)" json:"CANTIDAD"`
	PK_ID_PEDIDO      *int64 `orm:"column(PK_ID_PEDIDO);null" json:"PK_ID_PEDIDO"`
}

func (ip *ItemPedido) TableName() string {
	return "ITEM_PEDIDO"
}

func init() {
	orm.RegisterModel(new(ItemPedido))
}
