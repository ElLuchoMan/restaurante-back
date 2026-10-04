package models

import (
	"encoding/json"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type Cupon struct {
	PkIdCupon          int64         `orm:"column(pk_id_cupon);pk;auto" json:"cuponId"`
	Codigo             string        `orm:"column(codigo);type(text);unique" json:"codigo"`
	Scope              CuponScope    `orm:"column(scope);type(cupon_scope)" json:"scope"`
	TipoDescuento      TipoDescuento `orm:"column(tipo_descuento);type(tipo_descuento)" json:"tipoDescuento"`
	ValorDescuento     int64         `orm:"column(valor_descuento);type(bigint)" json:"valorDescuento"`
	MaxUsos            *int          `orm:"column(max_usos);type(integer);null" json:"maxUsos,omitempty"`
	LimitePorCliente   *int          `orm:"column(limite_por_cliente);type(integer);null" json:"limitePorCliente,omitempty"`
	MontoMinimo        *int64        `orm:"column(monto_minimo);type(bigint);null" json:"montoMinimo,omitempty"`
	FechaInicio        time.Time     `orm:"column(fecha_inicio);type(date)" json:"fechaInicio"`
	FechaFin           time.Time     `orm:"column(fecha_fin);type(date)" json:"fechaFin"`
	PkIdProducto       *Producto     `orm:"column(pk_id_producto);rel(fk);null" json:"productoId,omitempty" swaggertype:"integer"`
	PkIdCategoria      *Categoria    `orm:"column(pk_id_categoria);rel(fk);null" json:"categoriaId,omitempty" swaggertype:"integer"`
	PkDocumentoCliente *Cliente      `orm:"column(pk_documento_cliente);rel(fk);null" json:"documentoCliente,omitempty" swaggertype:"integer"`
	Activo             bool          `orm:"column(activo);type(boolean);default(true)" json:"activo"`
}

func (c *Cupon) TableName() string {
	return "cupon"
}

func init() {
	orm.RegisterModel(new(Cupon))
}

func (c Cupon) MarshalJSON() ([]byte, error) {

	fiStr := FormatDateUTC(c.FechaInicio)
	ffStr := FormatDateUTC(c.FechaFin)

	return json.Marshal(&struct {
		PkIdCupon          int64         `json:"cuponId"`
		Codigo             string        `json:"codigo"`
		Scope              CuponScope    `json:"scope"`
		TipoDescuento      TipoDescuento `json:"tipoDescuento"`
		ValorDescuento     int64         `json:"valorDescuento"`
		MaxUsos            *int          `json:"maxUsos,omitempty"`
		LimitePorCliente   *int          `json:"limitePorCliente,omitempty"`
		MontoMinimo        *int64        `json:"montoMinimo,omitempty"`
		FechaInicio        string        `json:"fechaInicio"`
		FechaFin           string        `json:"fechaFin"`
		PkIdProducto       *Producto     `json:"productoId,omitempty"`
		PkIdCategoria      *Categoria    `json:"categoriaId,omitempty"`
		PkDocumentoCliente *Cliente      `json:"documentoCliente,omitempty"`
		Activo             bool          `json:"activo"`
	}{
		PkIdCupon:          c.PkIdCupon,
		Codigo:             c.Codigo,
		Scope:              c.Scope,
		TipoDescuento:      c.TipoDescuento,
		ValorDescuento:     c.ValorDescuento,
		MaxUsos:            c.MaxUsos,
		LimitePorCliente:   c.LimitePorCliente,
		MontoMinimo:        c.MontoMinimo,
		FechaInicio:        fiStr,
		FechaFin:           ffStr,
		PkIdProducto:       c.PkIdProducto,
		PkIdCategoria:      c.PkIdCategoria,
		PkDocumentoCliente: c.PkDocumentoCliente,
		Activo:             c.Activo,
	})
}
