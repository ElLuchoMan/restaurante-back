package models

import "github.com/beego/beego/v2/client/orm"

// ProductoPedidoDetalle representa cada producto asociado a un pedido.
type ProductoPedidoDetalle struct {
	PKIDProductoPedidoDetalle int64   `orm:"column(PK_ID_PRODUCTO_PEDIDO_DETALLE);pk;auto" json:"detalleId"`
	PKIDPedido                int64   `orm:"column(PK_ID_PEDIDO)" json:"pedidoId"`
	PKIDProducto              int64   `orm:"column(PK_ID_PRODUCTO)" json:"productoId"`
	CANTIDAD                  int     `orm:"column(CANTIDAD)" json:"cantidad"`
	PRECIOUNITARIO            float64 `orm:"column(PRECIO_UNITARIO);null" json:"precioUnitario,omitempty"`
	SUBTOTAL                  float64 `orm:"column(SUBTOTAL);null" json:"subtotal,omitempty"`
}

// TableName especifica el nombre de la tabla en la base de datos.
func (p *ProductoPedidoDetalle) TableName() string {
	return "PRODUCTO_PEDIDO_DETALLE"
}

func init() {
	orm.RegisterModel(new(ProductoPedidoDetalle))
}
