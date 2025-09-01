package models

import (
	"encoding/json"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type Pedido struct {
	PK_ID_PEDIDO         int64        `orm:"column(pk_id_pedido);pk;auto" json:"pedidoId"`
	FECHA                time.Time    `orm:"column(fecha);type(date)" json:"fechaPedido"`
	HORA                 time.Time    `orm:"column(hora);type(time)" json:"horaPedido"`
	DELIVERY             bool         `orm:"column(delivery);type(boolean)" json:"delivery"`
	ESTADO_PEDIDO        EstadoPedido `orm:"column(estado_pedido)" json:"estadoPedido"`
	PK_ID_DOMICILIO      *int64       `orm:"column(pk_id_domicilio);null" json:"domicilioId,omitempty"`
	PK_ID_PAGO           *int64       `orm:"column(pk_id_pago);null" json:"pagoId"`
	PK_ID_RESTAURANTE    int64        `orm:"column(pk_id_restaurante)" json:"restauranteId"`
	PK_DOCUMENTO_CLIENTE int64        `orm:"column(pk_documento_cliente)" json:"documentoCliente"`
	UPDATED_AT           time.Time    `orm:"column(updated_at);type(timestamptz);auto_now" json:"updatedAt"`
	UPDATED_BY           *string      `orm:"column(updated_by);type(text);null" json:"updatedBy,omitempty"`
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
	DocumentoCliente int64  `json:"documentoCliente" orm:"column(pk_documento_cliente)"`
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
		HORA       string `json:"horaPedido"`
		UPDATED_AT string `json:"updatedAt"`
		Alias
	}{
		FECHA:      d.FECHA.Format("02-01-2006"),
		HORA:       d.HORA.Format("15:04:05"),
		UPDATED_AT: d.UPDATED_AT.Format("02-01-2006 15:04:05"),
		Alias:      (Alias)(d),
	})
}
