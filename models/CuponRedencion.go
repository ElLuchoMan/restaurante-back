package models

import (
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type CuponRedencion struct {
	PkIdCuponRedencion int64     `orm:"column(pk_id_cupon_redencion);pk;auto" json:"cuponRedencionId"`
	PkIdCupon          *Cupon    `orm:"column(pk_id_cupon);rel(fk)" json:"cuponId" swaggertype:"integer"`
	PkDocumentoCliente *Cliente  `orm:"column(pk_documento_cliente);rel(fk)" json:"documentoCliente" swaggertype:"integer"`
	PkIdPedido         *Pedido   `orm:"column(pk_id_pedido);rel(fk);null" json:"pedidoId,omitempty" swaggertype:"integer"`
	MontoDescuento     int64     `orm:"column(monto_descuento);type(bigint)" json:"montoDescuento"`
	CreatedAt          time.Time `orm:"column(created_at);type(timestamptz);auto_now_add" json:"createdAt"`
}

func (c *CuponRedencion) TableName() string {
	return "cupon_redencion"
}

func init() {
	orm.RegisterModel(new(CuponRedencion))
}
