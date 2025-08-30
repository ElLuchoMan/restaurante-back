package models

import "github.com/beego/beego/v2/client/orm"

type DetallePedido struct {
	PKIDPedido   int64   `orm:"column(pk_id_pedido);pk" json:"pedidoId"`
	PKIDProducto int64   `orm:"column(pk_id_producto);pk" json:"productoId"`
	Precio       float64 `orm:"column(precio)" json:"precio"`
	Cantidad     int     `orm:"column(cantidad)" json:"cantidad"`
}

func (d *DetallePedido) TableName() string {
	return "detalle_pedido"
}

func init() {
	orm.RegisterModel(new(DetallePedido))
}
