package models

import (
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type PrecioProductoHist struct {
	PKIDPrecioProductoHist int64      `orm:"column(pk_id_precio_producto_hist);pk;auto" json:"precioProductoHistId"`
	PKIDProducto           int64      `orm:"column(pk_id_producto)" json:"productoId"`
	Precio                 float64    `orm:"column(precio)" json:"precio"`
	FechaInicio            time.Time  `orm:"column(fecha_inicio);type(date)" json:"fechaInicio"`
	FechaFin               *time.Time `orm:"column(fecha_fin);type(date);null" json:"fechaFin,omitempty"`
	CreatedAt              time.Time  `orm:"column(created_at);type(timestamp);auto_now_add" json:"createdAt"`
	UpdatedAt              time.Time  `orm:"column(updated_at);type(timestamp);auto_now" json:"updatedAt"`
	CreatedBy              *string    `orm:"column(created_by);type(text)" json:"createdBy,omitempty"`
	UpdatedBy              *string    `orm:"column(updated_by);type(text)" json:"updatedBy,omitempty"`
}

func (p *PrecioProductoHist) TableName() string {
	return "precio_producto_hist"
}

func init() {
	orm.RegisterModel(new(PrecioProductoHist))
}
