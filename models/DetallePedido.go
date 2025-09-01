package models

import "github.com/beego/beego/v2/client/orm"

// DetallePedido represents a product within un pedido.
type DetallePedido struct {
	PK_ID_DETALLE int64 `orm:"column(pk_id_detalle);pk;auto" json:"detalleId"`
	PKIDPedido    int64 `orm:"column(pk_id_pedido);unique(pedido_producto)" json:"pedidoId"`
	PKIDProducto  int64 `orm:"column(pk_id_producto);unique(pedido_producto)" json:"productoId"`
	Precio        int64 `orm:"column(precio);type(bigint)" json:"precio"`
	Cantidad      int   `orm:"column(cantidad)" json:"cantidad"`
}

func (d *DetallePedido) TableName() string {
	return "detalle_pedido"
}

func init() {
	orm.RegisterModel(new(DetallePedido))
}
