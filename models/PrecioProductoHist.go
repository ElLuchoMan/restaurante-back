package models

import (
	"time"

	"github.com/beego/beego/v2/client/orm"
)

// PrecioProductoHist represents the price of a product at a given time.
// It no longer maintains its own primary key or audit timestamps.
type PrecioProductoHist struct {
	PK_ID_PRECIO_HIST int64     `orm:"column(pk_id_precio_hist);pk;auto" json:"precioHistId"`
	PKIDProducto      int64     `orm:"column(pk_id_producto);rel(fk)" json:"productoId"`
	Precio            int64     `orm:"column(precio);type(bigint)" json:"precio"`
	FechaVigencia     time.Time `orm:"column(fecha_vigencia);type(date)" json:"fechaVigencia"`
}

func (p *PrecioProductoHist) TableName() string {
	return "precio_producto_hist"
}

// TableUnique enforces UNIQUE(pk_id_producto, fecha_vigencia)
func (p *PrecioProductoHist) TableUnique() [][]string {
	return [][]string{{"PKIDProducto", "FechaVigencia"}}
}

func init() {
	orm.RegisterModel(new(PrecioProductoHist))
}
