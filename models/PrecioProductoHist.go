package models

import (
	"time"

	"github.com/beego/beego/v2/client/orm"
)

// PrecioProductoHist represents the price of a product at a given time.
// It no longer maintains its own primary key or audit timestamps.
type PrecioProductoHist struct {
	PKIDProducto  int64     `orm:"column(pk_id_producto)" json:"productoId"`
	Precio        float64   `orm:"column(precio)" json:"precio"`
	FechaVigencia time.Time `orm:"column(fecha_vigencia);type(date)" json:"fechaVigencia"`
}

func (p *PrecioProductoHist) TableName() string {
	return "precio_producto_hist"
}

func init() {
	orm.RegisterModel(new(PrecioProductoHist))
}
