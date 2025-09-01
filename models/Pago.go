package models

import (
	"encoding/json"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type Pago struct {
	PK_ID_PAGO        int64      `orm:"column(pk_id_pago);pk;auto" json:"pagoId"`
	FECHA             time.Time  `orm:"column(fecha);type(date)" json:"fechaPago"`
	HORA              time.Time  `orm:"column(hora);type(time)" json:"horaPago"`
	MONTO             int64      `orm:"column(monto)" json:"monto"`
	ESTADO_PAGO       EstadoPago `orm:"column(estado_pago);type(text)" json:"estadoPago"`
	PK_ID_METODO_PAGO *int64     `orm:"column(pk_id_metodo_pago);rel(fk);null" json:"metodoPagoId,omitempty"`
	UPDATED_AT        time.Time  `orm:"column(updated_at);type(timestamptz);auto_now" json:"updatedAt"`
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
		HORA       string `json:"horaPago"`
		UPDATED_AT string `json:"updatedAt"`
		Alias
	}{
		FECHA:      d.FECHA.Format("02-01-2006"),
		HORA:       d.HORA.Format("15:04:05"),
		UPDATED_AT: d.UPDATED_AT.Format("02-01-2006 15:04:05"),
		Alias:      (Alias)(d),
	})
}
