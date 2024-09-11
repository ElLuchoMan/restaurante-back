package models

import "github.com/beego/beego/v2/client/orm"

type ItemPedido struct {
	PK_ID_ITEM_PEDIDO int64  `orm:"pk" json:"PK_ID_ITEM_PEDIDO"`
	CANTIDAD          int64  `json:"CANTIDAD"`
	PK_ID_PEDIDO      *int64 `orm:"null" json:"PK_ID_PEDIDO"`
}

func (c *ItemPedido) TableName() string {
	return "ITEM_PEDIDO"
}

func init() {
	orm.RegisterModel(new(ItemPedido))
}
