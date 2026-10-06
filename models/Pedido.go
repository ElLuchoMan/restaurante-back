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
	UPDATED_AT           time.Time    `orm:"column(updated_at);type(timestamptz);auto_now" json:"updatedAt" swaggertype:"string"`
	UPDATED_BY           *string      `orm:"column(updated_by);type(text);null" json:"updatedBy,omitempty"`
}

// PedidoDetails es el `data` de GET /pedidos/detalles. A diferencia de Pedido,
// `fechaPedido` va como DD-MM-YYYY, `horaPedido` como HH:MM:SS y las relaciones
// van como número (0 cuando no existen). `productos` es un string que contiene
// un JSON con un arreglo de objetos {pk_id_producto, nombre, cantidad, precio,
// subtotal} (`[]` si el pedido no tiene productos).
type PedidoDetails struct {
	PedidoID         int64  `json:"pedidoId" orm:"column(pk_id_pedido)" example:"10"`
	Fecha            string `json:"fechaPedido" orm:"column(fecha)" example:"31-01-2025"`
	Hora             string `json:"horaPedido" orm:"column(hora)" example:"18:30:00"`
	Delivery         bool   `json:"delivery" orm:"column(delivery)" example:"false"`
	EstadoPedido     string `json:"estadoPedido" orm:"column(estado_pedido)" enums:"INICIADO,EN_PREPARACION,LISTO,TERMINADO,CANCELADO" example:"INICIADO"`
	MetodoPago       string `json:"metodoPago" orm:"column(metodo_pago)" example:"NEQUI"`
	Productos        string `json:"productos" orm:"column(productos)" example:"[{\"pk_id_producto\":1,\"nombre\":\"Bandeja Paisa\",\"cantidad\":2,\"precio\":25000,\"subtotal\":50000}]"`
	PagoID           int64  `json:"pagoId" orm:"column(pago_id)" example:"4"`
	MetodoPagoID     int64  `json:"metodoPagoId" orm:"column(metodo_pago_id)" example:"1"`
	DomicilioID      int64  `json:"domicilioId" orm:"column(domicilio_id)" example:"3"`
	DocumentoCliente int64  `json:"documentoCliente" orm:"column(pk_documento_cliente)" example:"1234567890"`
}

func (p *Pedido) TableName() string {
	return "pedido"
}

func init() {
	orm.RegisterModel(new(Pedido))
}

func (d Pedido) MarshalJSON() ([]byte, error) {

	fechaStr := FormatDateUTC(d.FECHA)

	horaStr := FormatTimeWithLMT(d.HORA)

	updatedAtStr := FormatTimestampBogota(d.UPDATED_AT)

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
		HORA:                 horaStr,
		DELIVERY:             d.DELIVERY,
		ESTADO_PEDIDO:        d.ESTADO_PEDIDO,
		PK_ID_DOMICILIO:      d.PK_ID_DOMICILIO,
		PK_ID_PAGO:           d.PK_ID_PAGO,
		PK_ID_RESTAURANTE:    d.PK_ID_RESTAURANTE,
		PK_DOCUMENTO_CLIENTE: d.PK_DOCUMENTO_CLIENTE,
		UPDATED_AT:           updatedAtStr,
		UPDATED_BY:           d.UPDATED_BY,
	})
}
