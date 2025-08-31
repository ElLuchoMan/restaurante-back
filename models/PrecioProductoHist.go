package models

import (
	"time"

	"github.com/beego/beego/v2/client/orm"
)

// PrecioProductoHist represents the price of a product at a given time.
// It no longer maintains its own primary key or audit timestamps.
type PrecioProductoHist struct {
	PKIDProducto  int64     `orm:"column(pk_id_producto);pk" json:"productoId"`
	Precio        int64     `orm:"column(precio);type(bigint)" json:"precio"`
	FechaVigencia time.Time `orm:"column(fecha_vigencia);type(date)" json:"fechaVigencia"` // part of composite PK
}

func (p *PrecioProductoHist) TableName() string {
	return "precio_producto_hist"
}

func init() {
	orm.RegisterModel(new(PrecioProductoHist))
}
