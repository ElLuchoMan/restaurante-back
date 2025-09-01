package models

import (
	"encoding/json"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type Reserva struct {
	PK_ID_RESERVA     int64         `orm:"column(pk_id_reserva);pk;auto" json:"reservaId"`
	FECHA             time.Time     `orm:"column(fecha);type(date)" json:"fechaReserva"`
	HORA              time.Time     `orm:"column(hora);type(time);size(8)" json:"horaReserva"`
	PERSONAS          int           `orm:"column(personas)" json:"personas"`
	PK_ID_CONTACTO    int64         `orm:"column(pk_id_contacto)" json:"contactoId"`
	PK_ID_RESTAURANTE int64         `orm:"column(pk_id_restaurante)" json:"restauranteId"`
	ESTADO_RESERVA    EstadoReserva `orm:"column(estado_reserva);type(text);default(PENDIENTE)" json:"estadoReserva"`
	INDICACIONES      *string       `orm:"column(indicaciones);null" json:"indicaciones,omitempty"`
	CREATED_AT        time.Time     `orm:"column(created_at);type(timestamp);auto_now_add" json:"createdAt"`
	UPDATED_AT        time.Time     `orm:"column(updated_at);type(timestamp);auto_now" json:"updatedAt"`
	CREATED_BY        *string       `orm:"column(created_by);type(text);null" json:"createdBy,omitempty"`
	UPDATED_BY        *string       `orm:"column(updated_by);type(text);null" json:"updatedBy,omitempty"`
}

func (r *Reserva) TableName() string {
	return "reserva"
}

func init() {
	orm.RegisterModel(new(Reserva))
}

func (t Reserva) MarshalJSON() ([]byte, error) {
	type Alias Reserva
	return json.Marshal(&struct {
		FECHA string `json:"fechaReserva"`
		HORA  string `json:"horaReserva"`
		Alias
	}{
		FECHA: t.FECHA.Format("02-01-2006"),
		HORA:  t.HORA.Format("15:04:05"),
		Alias: (Alias)(t),
	})
}
