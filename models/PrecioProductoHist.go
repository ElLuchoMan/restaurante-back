package models

import (
	"encoding/json"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type PrecioProductoHist struct {
	PK_ID_PRECIO_HIST int64     `orm:"column(pk_id_precio_hist);pk;auto" json:"precioHistId"`
	PKIDProducto      *Producto `orm:"column(pk_id_producto);rel(fk)" json:"productoId"`
	Precio            int64     `orm:"column(precio);type(bigint)" json:"precio"`
	FechaVigencia     time.Time `orm:"column(fecha_vigencia);type(date)" json:"fechaVigencia"`
}

func (p *PrecioProductoHist) TableName() string {
	return "precio_producto_hist"
}

func (p *PrecioProductoHist) TableUnique() [][]string {
	return [][]string{{"PKIDProducto", "FechaVigencia"}}
}

func init() {
	orm.RegisterModel(new(PrecioProductoHist))
}

func (p PrecioProductoHist) MarshalJSON() ([]byte, error) {

	fechaStr := FormatDateUTC(p.FechaVigencia)

	var productoID *int64
	if p.PKIDProducto != nil {
		id := p.PKIDProducto.PK_ID_PRODUCTO
		productoID = &id
	}

	return json.Marshal(&struct {
		PK_ID_PRECIO_HIST int64  `json:"precioHistId"`
		PKIDProducto      *int64 `json:"productoId"`
		Precio            int64  `json:"precio"`
		FechaVigencia     string `json:"fechaVigencia"`
	}{
		PK_ID_PRECIO_HIST: p.PK_ID_PRECIO_HIST,
		PKIDProducto:      productoID,
		Precio:            p.Precio,
		FechaVigencia:     fechaStr,
	})
}
