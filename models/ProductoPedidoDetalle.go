package models

import "github.com/beego/beego/v2/client/orm"

// ProductoPedidoDetalle representa cada producto asociado a un pedido.
type ProductoPedidoDetalle struct {
	PKIDProductoPedido int64   `orm:"column(PK_ID_PRODUCTO_PEDIDO);pk" json:"productoPedidoId"`
	PKIDProducto       int64   `orm:"column(PK_ID_PRODUCTO);pk" json:"productoId"`
	CANTIDAD           int     `orm:"column(CANTIDAD)" json:"cantidad"`
	PRECIO             float64 `orm:"column(PRECIO);null" json:"precio,omitempty"`
}

// TableName especifica el nombre de la tabla en la base de datos.
func (p *ProductoPedidoDetalle) TableName() string {
	return "PRODUCTO_PEDIDO_DETALLE"
}

func init() {
	orm.RegisterModel(new(ProductoPedidoDetalle))
}
