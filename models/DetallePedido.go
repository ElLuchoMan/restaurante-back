package models

import "github.com/beego/beego/v2/client/orm"

// DetallePedidoPK encapsulates the composite primary key for DetallePedido.
// DetallePedido represents a product within a pedido.
type DetallePedido struct {
	PK_ID_DETALLE int64 `orm:"column(pk_id_detalle);pk;auto" json:"detalleId"`
	PKIDPedido    int64 `orm:"column(pk_id_pedido)" json:"pedidoId"`
	PKIDProducto  int64 `orm:"column(pk_id_producto)" json:"productoId"`
	Precio        int64 `orm:"column(precio);type(bigint)" json:"precio"`
	Cantidad      int   `orm:"column(cantidad)" json:"cantidad"`
}

func (d *DetallePedido) TableName() string {
	return "detalle_pedido"
}

func init() {
	orm.RegisterModel(new(DetallePedido))
}
