package models

import "github.com/beego/beego/v2/client/orm"

// ProductoPedidoDetalle representa cada producto asociado a un pedido.
type ProductoPedidoDetalle struct {
	PKIDProductoPedido int64   `orm:"column(pk_id_producto_pedido);pk" json:"productoPedidoId"`
	PKIDPedido         int64   `orm:"column(pk_id_pedido);pk" json:"pedidoId"`
	PKIDProducto       int64   `orm:"column(pk_id_producto);pk" json:"productoId"`
	CANTIDAD           int     `orm:"column(cantidad)" json:"cantidad"`
	PRECIOUNITARIO     float64 `orm:"column(precio_unitario);null" json:"precioUnitario,omitempty"`
	SUBTOTAL           float64 `orm:"column(subtotal);null" json:"subtotal,omitempty"`
}

// TableName especifica el nombre de la tabla en la base de datos.
func (p *ProductoPedidoDetalle) TableName() string {
	return "producto_pedido_detalle"
}

func init() {
	orm.RegisterModel(new(ProductoPedidoDetalle))
}
