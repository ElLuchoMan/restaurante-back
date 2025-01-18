package models

import (
	"encoding/json"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type Reserva struct {
	PK_ID_RESERVA  int       `orm:"column(PK_ID_RESERVA);pk;auto" json:"PK_ID_RESERVA"`
	FECHA          time.Time `orm:"column(FECHA);type(date)" json:"FECHA"`
	HORA           string    `orm:"column(HORA);type(time);size(8)" json:"HORA"`
	PERSONAS       int       `orm:"column(PERSONAS)" json:"PERSONAS"`
	ESTADO_RESERVA *string   `orm:"column(ESTADO_RESERVA);null" json:"ESTADO_RESERVA,omitempty"`
	INDICACIONES   *string   `orm:"column(INDICACIONES);null" json:"INDICACIONES,omitempty"`
	CREATED_AT     time.Time `orm:"column(CREATED_AT);type(timestamp);auto_now_add" json:"CREATED_AT"`
	UPDATED_AT     time.Time `orm:"column(UPDATED_AT);type(timestamp);auto_now" json:"UPDATED_AT"`
	CREATED_BY     *string   `orm:"column(CREATED_BY);type(date)" json:"CREATED_BY,omitempty"`
	UPDATED_BY     *string   `orm:"column(UPDATED_BY);type(date)" json:"UPDATED_BY,omitempty"`
}

func (r *Reserva) TableName() string {
	return "RESERVA"
}

func init() {
	orm.RegisterModel(new(Reserva))
}

func (t Reserva) MarshalJSON() ([]byte, error) {
	type Alias Reserva
	return json.Marshal(&struct {
		FECHA string `json:"FECHA"`
		Alias
	}{
		FECHA: t.FECHA.Format("02-01-2006"),
		Alias: (Alias)(t),
	})
}
