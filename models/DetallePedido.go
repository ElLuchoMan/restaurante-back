package models

import "github.com/beego/beego/v2/client/orm"

type DetallePedido struct {
	PK_ID_DETALLE int64     `orm:"column(pk_id_detalle);pk;auto" json:"detalleId"`
	PKIDPedido    *Pedido   `orm:"column(pk_id_pedido);rel(fk)" json:"pedidoId" swaggertype:"integer"`
	PKIDProducto  *Producto `orm:"column(pk_id_producto);rel(fk)" json:"productoId" swaggertype:"integer"`
	Precio        int64     `orm:"column(precio);type(bigint)" json:"precio"`
	Cantidad      int       `orm:"column(cantidad)" json:"cantidad"`
}

func (d *DetallePedido) TableName() string {
	return "detalle_pedido"
}

func (d *DetallePedido) TableUnique() [][]string {
	return [][]string{{"PKIDPedido", "PKIDProducto"}}
}

func init() {
	orm.RegisterModel(new(DetallePedido))
}
