package models

import "github.com/beego/beego/v2/client/orm"

type ProductoPedido struct {
	PK_ID_PRODUCTO_PEDIDO int64  `orm:"column(PK_ID_PRODUCTO_PEDIDO);pk;auto" json:"PK_ID_PRODUCTO_PEDIDO"`
	DETALLES_PRODUCTOS    string `orm:"column(DETALLES_PRODUCTOS);type(jsonb)" json:"DETALLES_PRODUCTOS"` // JSONB para consolidar productos
	PK_ID_PEDIDO          int64  `orm:"column(PK_ID_PEDIDO)" json:"PK_ID_PEDIDO"`
}

func (p *ProductoPedido) TableName() string {
	return "PRODUCTO_PEDIDO"
}

func init() {
	orm.RegisterModel(new(ProductoPedido))
}
