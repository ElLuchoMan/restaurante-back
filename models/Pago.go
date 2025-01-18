package models

import (
	"encoding/json"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type Pago struct {
	PK_ID_PAGO        int       `orm:"column(PK_ID_PAGO);pk;auto" json:"PK_ID_PAGO"`
	FECHA             time.Time `orm:"column(FECHA);type(date)" json:"FECHA"`
	HORA              string    `orm:"column(HORA);type(time)" json:"HORA"`
	MONTO             int64     `orm:"column(MONTO)" json:"MONTO"`
	ESTADO_PAGO       string    `orm:"column(ESTADO_PAGO);type(text)" json:"ESTADO_PAGO"`
	PK_ID_METODO_PAGO int       `orm:"column(PK_ID_METODO_PAGO);null" json:"PK_ID_METODO_PAGO"`
	UPDATED_AT        time.Time `orm:"column(UPDATED_AT);type(timestamp);auto_now" json:"UPDATED_AT"`
	UPDATED_BY        string    `orm:"column(UPDATED_BY)" json:"UPDATED_BY"`
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
		FECHA      string `json:"FECHA"`
		CREATED_AT string `json:"CREATED_AT"`
		UPDATED_AT string `json:"UPDATED_AT"`
		Alias
	}{
		FECHA:      d.FECHA.Format("02-01-2006"),
		UPDATED_AT: d.UPDATED_AT.Format("02-01-2006 15:04:05"),
		Alias:      (Alias)(d),
	})
}
