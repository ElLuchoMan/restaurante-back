package models

import (
	"encoding/json"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type Pago struct {
	PK_ID_PAGO        int        `orm:"column(pk_id_pago);pk;auto" json:"pagoId"`
	FECHA             time.Time  `orm:"column(fecha);type(date)" json:"fechaPago"`
	HORA              string     `orm:"column(hora);type(time)" json:"horaPago"`
	MONTO             int64      `orm:"column(monto)" json:"monto"`
	ESTADO_PAGO       EstadoPago `orm:"column(estado_pago);type(text)" json:"estadoPago"`
	PK_ID_METODO_PAGO int        `orm:"column(pk_id_metodo_pago);null" json:"metodoPagoId"`
	CREATED_AT        time.Time  `orm:"column(created_at);type(timestamp);auto_now_add" json:"createdAt"`
	UPDATED_AT        time.Time  `orm:"column(updated_at);type(timestamp);auto_now" json:"updatedAt"`
	CREATED_BY        *string    `orm:"column(created_by);type(text)" json:"createdBy,omitempty"`
	UPDATED_BY        *string    `orm:"column(updated_by);type(text);null" json:"updatedBy,omitempty"`
}

func (p *Pago) TableName() string {
	return "pago"
}

func init() {
	orm.RegisterModel(new(Pago))
}

func (d Pago) MarshalJSON() ([]byte, error) {
	type Alias Pago
	return json.Marshal(&struct {
		FECHA      string `json:"fechaPago"`
		UPDATED_AT string `json:"updatedAt"`
		Alias
	}{
		FECHA:      d.FECHA.Format("02-01-2006"),
		UPDATED_AT: d.UPDATED_AT.Format("02-01-2006 15:04:05"),
		Alias:      (Alias)(d),
	})
}
