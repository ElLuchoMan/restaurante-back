package models

import (
	"encoding/json"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type Pedido struct {
	PK_ID_PEDIDO      int       `orm:"column(pk_id_pedido);pk;auto" json:"pedidoId"`
	FECHA             time.Time `orm:"column(fecha);type(date)" json:"fechaPedido"`
	HORA              string    `orm:"column(hora);type(time)" json:"horaPedido"`
	DELIVERY          bool      `orm:"column(delivery); type(boolean)" json:"delivery"`
	ESTADO_PEDIDO     string    `orm:"column(estado_pedido)" json:"estadoPedido"`
	PK_ID_DOMICILIO   *int      `orm:"column(pk_id_domicilio);null" json:"domicilioId,omitempty"`
	PK_ID_PAGO        *int      `orm:"column(pk_id_pago);null" json:"pagoId"`
	PK_ID_RESTAURANTE *int      `orm:"column(pk_id_restaurante);null" json:"restauranteId"`
	UPDATED_AT        time.Time `orm:"column(updated_at);type(timestamp);auto_now" json:"updatedAt"`
	UPDATED_BY        string    `orm:"column(updated_by)" json:"updatedBy"`
}

type PedidoDetails struct {
	PedidoID         int64  `json:"pedidoId" orm:"column(pk_id_pedido)"`
	Fecha            string `json:"fechaPedido" orm:"column(fecha)"`
	Hora             string `json:"horaPedido" orm:"column(hora)"`
	Delivery         bool   `json:"delivery" orm:"column(delivery)"`
	EstadoPedido     string `json:"estadoPedido" orm:"column(estado_pedido)"`
	MetodoPago       string `json:"metodoPago" orm:"column(metodo_pago)"`
	Productos        string `json:"productos" orm:"column(productos)"`
	PagoID           int64  `json:"pagoId" orm:"column(pago_id)"`
	MetodoPagoID     int64  `json:"metodoPagoId" orm:"column(metodo_pago_id)"`
	DomicilioID      int64  `json:"domicilioId" orm:"column(domicilio_id)"`
	DocumentoCliente int64  `json:"documentoCliente" orm:"column(documento_cliente)"`
}

func (p *Pedido) TableName() string {
	return "pedido"
}

func init() {
	orm.RegisterModel(new(Pedido))
}

func (d Pedido) MarshalJSON() ([]byte, error) {
	type Alias Pedido
	return json.Marshal(&struct {
		FECHA      string `json:"fechaPedido"`
		CREATED_AT string `json:"createdAt"`
		UPDATED_AT string `json:"updatedAt"`
		Alias
	}{
		FECHA:      d.FECHA.Format("02-01-2006"),
		UPDATED_AT: d.UPDATED_AT.Format("02-01-2006 15:04:05"),
		Alias:      (Alias)(d),
	})
}
