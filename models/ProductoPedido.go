package models

import "github.com/beego/beego/v2/client/orm"

type ProductoPedido struct {
	PK_ID_PRODUCTO_PEDIDO int64  `orm:"column(pk_id_producto_pedido);pk;auto" json:"productoPedidoId"`
	DETALLES_PRODUCTOS    string `orm:"column(detalles_productos);type(jsonb)" json:"detallesProductos"`
	PK_ID_PEDIDO          int64  `orm:"column(pk_id_pedido)" json:"pedidoId"`
}

func (p *ProductoPedido) TableName() string {
	return "producto_pedido"
}

func init() {
	orm.RegisterModel(new(ProductoPedido))
}
