package models

import "github.com/beego/beego/v2/client/orm"

// DetallePedidoPK encapsulates the composite primary key for DetallePedido.
type DetallePedidoPK struct {
	PKIDPedido   int64 `orm:"column(pk_id_pedido)" json:"pedidoId"`
	PKIDProducto int64 `orm:"column(pk_id_producto)" json:"productoId"`
}

// DetallePedido represents a product within a pedido.
// The composite primary key is defined in the embedded DetallePedidoPK struct.
type DetallePedido struct {
	DetallePedidoPK `orm:"pk"`
	Precio          int64 `orm:"column(precio);type(bigint)" json:"precio"`
	Cantidad        int   `orm:"column(cantidad)" json:"cantidad"`
}

func (d *DetallePedido) TableName() string {
	return "detalle_pedido"
}

func init() {
	orm.RegisterModel(new(DetallePedido))
}
