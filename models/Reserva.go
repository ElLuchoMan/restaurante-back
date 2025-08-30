package models

import (
	"encoding/json"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type Reserva struct {
	PK_ID_RESERVA     int       `orm:"column(PK_ID_RESERVA);pk;auto" json:"reservaId"`
	FECHA             time.Time `orm:"column(FECHA);type(date)" json:"fechaReserva"`
	HORA              string    `orm:"column(HORA);type(time);size(8)" json:"horaReserva"`
	PERSONAS          int       `orm:"column(PERSONAS)" json:"personas"`
	ESTADO_RESERVA    *string   `orm:"column(ESTADO_RESERVA);null" json:"estadoReserva,omitempty"`
	INDICACIONES      *string   `orm:"column(INDICACIONES);null" json:"indicaciones,omitempty"`
	CREATED_AT        time.Time `orm:"column(CREATED_AT);type(timestamp);auto_now_add" json:"createdAt"`
	UPDATED_AT        time.Time `orm:"column(UPDATED_AT);type(timestamp);auto_now" json:"updatedAt"`
	CREATED_BY        *string   `orm:"column(CREATED_BY);type(date)" json:"createdBy,omitempty"`
	UPDATED_BY        *string   `orm:"column(UPDATED_BY);type(date)" json:"updatedBy,omitempty"`
	NOMBRE_COMPLETO   *string   `orm:"column(NOMBRE_COMPLETO);type(text);null" json:"nombreCompleto,omitempty"`
	TELEFONO          *string   `orm:"column(TELEFONO);type(text);null" json:"telefono,omitempty"`
	DOCUMENTO_CLIENTE *int64    `orm:"column(DOCUMENTO_CLIENTE);null" json:"documentoCliente,omitempty"`
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
		FECHA string `json:"fechaReserva"`
		Alias
	}{
		FECHA: t.FECHA.Format("02-01-2006"),
		Alias: (Alias)(t),
	})
}
