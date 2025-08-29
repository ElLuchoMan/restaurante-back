package models

import "github.com/beego/beego/v2/client/orm"

// ProductoPedidoDetalle representa un producto específico dentro de un pedido.
type ProductoPedidoDetalle struct {
	PK_ID_PRODUCTO_PEDIDO_DETALLE int64   `orm:"column(PK_ID_PRODUCTO_PEDIDO_DETALLE);pk;auto" json:"productoPedidoDetalleId"`
	PK_ID_PRODUCTO_PEDIDO         int64   `orm:"column(PK_ID_PRODUCTO_PEDIDO)" json:"productoPedidoId"`
	PK_ID_PRODUCTO                int64   `orm:"column(PK_ID_PRODUCTO)" json:"productoId"`
	CANTIDAD                      int     `orm:"column(CANTIDAD)" json:"cantidad"`
	PRECIO_UNITARIO               float64 `orm:"column(PRECIO_UNITARIO)" json:"precioUnitario"`
	SUBTOTAL                      float64 `orm:"column(SUBTOTAL)" json:"subtotal"`
}

func (p *ProductoPedidoDetalle) TableName() string {
	return "PRODUCTO_PEDIDO_DETALLE"
}

func init() {
	orm.RegisterModel(new(ProductoPedidoDetalle))
}
