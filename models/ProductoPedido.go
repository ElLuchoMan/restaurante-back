package models

import "github.com/beego/beego/v2/client/orm"

type ProductoPedido struct {
	PK_ID_PRODUCTO_PEDIDO int64                   `orm:"column(PK_ID_PRODUCTO_PEDIDO);pk;auto" json:"productoPedidoId"`
	PK_ID_PEDIDO          int64                   `orm:"column(PK_ID_PEDIDO)" json:"pedidoId"`
	Detalles              []ProductoPedidoDetalle `orm:"-" json:"detalles,omitempty"`
}

func (p *ProductoPedido) TableName() string {
	return "PRODUCTO_PEDIDO"
}

func init() {
	orm.RegisterModel(new(ProductoPedido))
}
