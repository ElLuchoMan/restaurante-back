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
	ESTADO_PEDIDO        EstadoPedido `orm:"column(estado_pedido);type(estado_pedido)" json:"estadoPedido"`
	PK_ID_DOMICILIO      *Domicilio   `orm:"column(pk_id_domicilio);rel(fk);null" json:"domicilioId,omitempty" swaggertype:"integer"`
	PK_ID_PAGO           *Pago        `orm:"column(pk_id_pago);rel(fk);null" json:"pagoId" swaggertype:"integer"`
	PK_ID_RESTAURANTE    *Restaurante `orm:"column(pk_id_restaurante);rel(fk);null" json:"restauranteId" swaggertype:"integer"`
	PK_DOCUMENTO_CLIENTE *Cliente     `orm:"column(pk_documento_cliente);rel(fk);null" json:"documentoCliente" swaggertype:"integer"`
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
	// Cargar zona horaria de Bogotá
	loc, err := time.LoadLocation("America/Bogota")
	if err != nil {
		loc = time.FixedZone("UTC-5", -5*60*60)
	}

	// Para campos de tipo date: extraer SOLO año/mes/día sin importar timezone
	fechaStr := time.Date(d.FECHA.Year(), d.FECHA.Month(), d.FECHA.Day(), 0, 0, 0, 0, time.UTC).Format("02-01-2006")

	// Para campos de tipo time: convertir a zona horaria de Bogotá
	horaEnBogota := d.HORA.In(loc)
	updatedAtEnBogota := d.UPDATED_AT.In(loc)

	return json.Marshal(&struct {
		PK_ID_PEDIDO         int64        `json:"pedidoId"`
		FECHA                string       `json:"fechaPedido"`
		HORA                 string       `json:"horaPedido"`
		DELIVERY             bool         `json:"delivery"`
		ESTADO_PEDIDO        EstadoPedido `json:"estadoPedido"`
		PK_ID_DOMICILIO      *Domicilio   `json:"domicilioId,omitempty"`
		PK_ID_PAGO           *Pago        `json:"pagoId"`
		PK_ID_RESTAURANTE    *Restaurante `json:"restauranteId"`
		PK_DOCUMENTO_CLIENTE *Cliente     `json:"documentoCliente"`
		UPDATED_AT           string       `json:"updatedAt"`
		UPDATED_BY           *string      `json:"updatedBy,omitempty"`
	}{
		PK_ID_PEDIDO:         d.PK_ID_PEDIDO,
		FECHA:                fechaStr,
		HORA:                 horaEnBogota.Format("15:04:05"),
		DELIVERY:             d.DELIVERY,
		ESTADO_PEDIDO:        d.ESTADO_PEDIDO,
		PK_ID_DOMICILIO:      d.PK_ID_DOMICILIO,
		PK_ID_PAGO:           d.PK_ID_PAGO,
		PK_ID_RESTAURANTE:    d.PK_ID_RESTAURANTE,
		PK_DOCUMENTO_CLIENTE: d.PK_DOCUMENTO_CLIENTE,
		UPDATED_AT:           updatedAtEnBogota.Format("02-01-2006 15:04:05"),
		UPDATED_BY:           d.UPDATED_BY,
	})
}
