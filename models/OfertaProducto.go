package models

import (
	"github.com/beego/beego/v2/client/orm"
)

type OfertaProducto struct {
	PkIdOferta   *Oferta   `orm:"column(pk_id_oferta);rel(fk)" json:"ofertaId" swaggertype:"integer"`
	PkIdProducto *Producto `orm:"column(pk_id_producto);rel(fk)" json:"productoId" swaggertype:"integer"`
}

func (o *OfertaProducto) TableName() string {
	return "oferta_producto"
}

func (o *OfertaProducto) TableUnique() [][]string {
	return [][]string{{"PkIdOferta", "PkIdProducto"}}
}

func init() {
	orm.RegisterModel(new(OfertaProducto))
}
