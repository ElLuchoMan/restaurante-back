package models

import (
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
