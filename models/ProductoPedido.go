package models

import "github.com/beego/beego/v2/client/orm"

type ProductoPedido struct {
	PK_ID_PRODUCTO_PEDIDO int64  `orm:"column(PK_ID_PRODUCTO_PEDIDO);pk;auto" json:"PK_ID_PRODUCTO_PEDIDO"`
	PRECIO_UNITARIO       int64  `orm:"column(PRECIO_UNITARIO)" json:"PRECIO_UNITARIO"`
	CANTIDAD              int    `orm:"column(CANTIDAD)" json:"CANTIDAD"`
	SUBTOTAL              int64  `orm:"column(SUBTOTAL)" json:"SUBTOTAL"`
	PK_ID_PEDIDO          *int64 `orm:"column(PK_ID_PEDIDO);null" json:"PK_ID_PEDIDO,omitempty"`
	PK_ID_PLATO           *int64 `orm:"column(PK_ID_PLATO);null" json:"PK_ID_PLATO,omitempty"`
}

func (p *ProductoPedido) TableName() string {
	return "PRODUCTO_PEDIDO"
}

func init() {
	orm.RegisterModel(new(ProductoPedido))
}
