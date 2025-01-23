package models

import (
	"encoding/json"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type Pago struct {
	PK_ID_PAGO        int       `orm:"column(PK_ID_PAGO);pk;auto" json:"pagoId"`
	FECHA             time.Time `orm:"column(FECHA);type(date)" json:"fechaPago"`
	HORA              string    `orm:"column(HORA);type(time)" json:"horaPago"`
	MONTO             int64     `orm:"column(MONTO)" json:"monto"`
	ESTADO_PAGO       string    `orm:"column(ESTADO_PAGO);type(text)" json:"estadoPago"`
	PK_ID_METODO_PAGO int       `orm:"column(PK_ID_METODO_PAGO);null" json:"metodoPagoId"`
	UPDATED_AT        time.Time `orm:"column(UPDATED_AT);type(timestamp);auto_now" json:"updatedAt"`
	UPDATED_BY        string    `orm:"column(UPDATED_BY)" json:"updatedBy"`
}

func (p *Pago) TableName() string {
	return "PAGO"
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
