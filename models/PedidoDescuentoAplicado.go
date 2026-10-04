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
	CreatedAt           time.Time       `orm:"column(created_at);type(timestamptz);auto_now_add" json:"createdAt" swaggertype:"string"`
}

func (p *PedidoDescuentoAplicado) TableName() string {
	return "pedido_descuento_aplicado"
}

func (p *PedidoDescuentoAplicado) BeforeInsert() {
	p.serializeDetalle()
}

func (p *PedidoDescuentoAplicado) BeforeUpdate() {
	p.serializeDetalle()
}

func (p *PedidoDescuentoAplicado) AfterLoad() {
	p.deserializeDetalle()
}

func (p *PedidoDescuentoAplicado) serializeDetalle() {
	if len(p.DetalleObj) == 0 {
		p.Detalle = ""
		return
	}
	p.Detalle = string(p.DetalleObj)
}

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

func (p PedidoDescuentoAplicado) MarshalJSON() ([]byte, error) {

	createdAtStr := FormatTimestampBogota(p.CreatedAt)

	return json.Marshal(&struct {
		PkIdPedidoDescuento int64           `json:"pedidoDescuentoId"`
		PkIdPedido          *Pedido         `json:"pedidoId" swaggertype:"integer"`
		PkIdCupon           *Cupon          `json:"cuponId,omitempty" swaggertype:"integer"`
		PkIdOferta          *Oferta         `json:"ofertaId,omitempty" swaggertype:"integer"`
		MontoDescuento      int64           `json:"montoDescuento"`
		DetalleObj          json.RawMessage `json:"detalle,omitempty" swaggertype:"object"`
		CreatedAt           string          `json:"createdAt"`
	}{
		PkIdPedidoDescuento: p.PkIdPedidoDescuento,
		PkIdPedido:          p.PkIdPedido,
		PkIdCupon:           p.PkIdCupon,
		PkIdOferta:          p.PkIdOferta,
		MontoDescuento:      p.MontoDescuento,
		DetalleObj:          p.DetalleObj,
		CreatedAt:           createdAtStr,
	})
}
