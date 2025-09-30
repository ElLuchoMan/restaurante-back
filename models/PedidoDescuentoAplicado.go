package models

import (
	"encoding/json"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type PedidoDescuentoAplicado struct {
	PkIdPedidoDescuento int64           `orm:"column(pk_id_pedido_descuento);pk;auto" json:"pedidoDescuentoId"`
	PkIdPedido          *Pedido         `orm:"column(pk_id_pedido);rel(fk)" json:"pedidoId" swaggertype:"integer"`
	PkIdCupon           *Cupon          `orm:"column(pk_id_cupon);rel(fk);null" json:"cuponId,omitempty" swaggertype:"integer"`
	PkIdOferta          *Oferta         `orm:"column(pk_id_oferta);rel(fk);null" json:"ofertaId,omitempty" swaggertype:"integer"`
	MontoDescuento      int64           `orm:"column(monto_descuento);type(bigint)" json:"montoDescuento"`
	Detalle             string          `orm:"column(detalle);type(jsonb);null" json:"-"`
	DetalleObj          json.RawMessage `orm:"-" json:"detalle,omitempty" swaggertype:"object"`
	CreatedAt           time.Time       `orm:"column(created_at);type(timestamptz);auto_now_add" json:"createdAt"`
}

func (p *PedidoDescuentoAplicado) TableName() string {
	return "pedido_descuento_aplicado"
}

// BeforeInsert se ejecuta antes de insertar en la base de datos
func (p *PedidoDescuentoAplicado) BeforeInsert() {
	p.serializeDetalle()
}

// BeforeUpdate se ejecuta antes de actualizar en la base de datos
func (p *PedidoDescuentoAplicado) BeforeUpdate() {
	p.serializeDetalle()
}

// AfterLoad se ejecuta después de cargar desde la base de datos
func (p *PedidoDescuentoAplicado) AfterLoad() {
	p.deserializeDetalle()
}

// serializeDetalle convierte el objeto JSON a string para la base de datos
func (p *PedidoDescuentoAplicado) serializeDetalle() {
	if len(p.DetalleObj) == 0 {
		p.Detalle = ""
		return
	}
	p.Detalle = string(p.DetalleObj)
}

// deserializeDetalle convierte el string de la base de datos a objeto JSON
func (p *PedidoDescuentoAplicado) deserializeDetalle() {
	if p.Detalle == "" {
		p.DetalleObj = nil
		return
	}
	p.DetalleObj = json.RawMessage(p.Detalle)
}

func init() {
	orm.RegisterModel(new(PedidoDescuentoAplicado))
}
