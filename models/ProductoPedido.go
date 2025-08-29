package models

import "github.com/beego/beego/v2/client/orm"

// ProductoPedido representa el vínculo entre un pedido y sus productos.
// Los detalles específicos de cada producto se almacenan en la tabla
// PRODUCTO_PEDIDO_DETALLE.
type ProductoPedido struct {
	PK_ID_PRODUCTO_PEDIDO int64 `orm:"column(PK_ID_PRODUCTO_PEDIDO);pk;auto" json:"productoPedidoId"`
	PK_ID_PEDIDO          int64 `orm:"column(PK_ID_PEDIDO)" json:"pedidoId"`
}

func (p *ProductoPedido) TableName() string {
	return "PRODUCTO_PEDIDO"
}

func init() {
	orm.RegisterModel(new(ProductoPedido))
}
