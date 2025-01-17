package models

import (
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type Pedido struct {
	PK_ID_PEDIDO      int       `orm:"column(PK_ID_PEDIDO);pk;auto" json:"PK_ID_PAGO"`
	FECHA             time.Time `orm:"column(FECHA);type(date)" json:"FECHA"`
	HORA              string    `orm:"column(HORA);type(time)" json:"HORA"`
	DELIVERY          bool      `orm:"column(DELIVERY); type(boolean)" json:"DELIVERY"`
	ESTADO_PEDIDO     string    `orm:"column(ESTADO_PEDIDO)"`
	PK_ID_DOMICILIO   *int      `orm:"column(PK_ID_DOMICILIO);null"`
	PK_ID_PAGO        *int      `orm:"column(PK_ID_PAGO);null"`
	PK_ID_RESTAURANTE *int      `orm:"column(PK_ID_RESTAURANTE);null"`
	UPDATED_AT        time.Time `orm:"column(UPDATED_AT);type(timestamp);auto_now" json:"UPDATED_AT"`
	UPDATED_BY        string    `orm:"column(UPDATED_BY)" json:"UPDATED_BY"`
}

type PedidoDetails struct {
	PKIDPedido   int64  `json:"PK_ID_PEDIDO" orm:"column(PK_ID_PEDIDO)"`
	Fecha        string `json:"FECHA" orm:"column(FECHA)"`
	Hora         string `json:"HORA" orm:"column(HORA)"`
	Delivery     bool   `json:"DELIVERY" orm:"column(DELIVERY)"`
	EstadoPedido string `json:"ESTADO_PEDIDO" orm:"column(ESTADO_PEDIDO)"`
	MetodoPago   string `json:"METODO_PAGO"`
	Productos    string `json:"PRODUCTOS"`
}

func (p *Pedido) TableName() string {
	return "PEDIDO"
}

func init() {
	orm.RegisterModel(new(Pedido))
}
