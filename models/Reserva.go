package models

import (
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type Reserva struct {
	PK_ID_RESERVA  int       `orm:"column(PK_ID_RESERVA);pk;auto" json:"PK_ID_RESERVA"`
	FECHA          time.Time `orm:"column(FECHA);type(date)" json:"FECHA"`
	HORA           string    `orm:"column(HORA);type(time)" json:"HORA"`
	PERSONAS       int       `orm:"column(PERSONAS)" json:"PERSONAS"`
	ESTADO_RESERVA string    `orm:"column(ESTADO_RESERVA);type(text)" json:"ESTADO_RESERVA"`
	INDICACIONES   string    `orm:"column(INDICACIONES);type(text)" json:"INDICACIONES"`
	CREATED_AT     string    `orm:"column(CREATED_AT);type(date)" json:"CREATED_AT"`
	UPDATED_AT     string    `orm:"column(UPDATED_AT);type(date)" json:"UPDATED_AT"`
	UPDATED_BY     string    `orm:"column(UPDATED_BY);type(text)" json:"UPDATED_BY"`
}

func (r *Reserva) TableName() string {
	return "RESERVA"
}

func init() {
	orm.RegisterModel(new(Reserva))
}
