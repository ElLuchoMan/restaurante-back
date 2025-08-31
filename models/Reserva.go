package models

import (
	"encoding/json"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type Reserva struct {
	PK_ID_RESERVA     int            `orm:"column(pk_id_reserva);pk;auto" json:"reservaId"`
	FECHA             time.Time      `orm:"column(fecha);type(date)" json:"fechaReserva"`
	HORA              string         `orm:"column(hora);type(time);size(8)" json:"horaReserva"`
	PERSONAS          int            `orm:"column(personas)" json:"personas"`
	ESTADO_RESERVA    *EstadoReserva `orm:"column(estado_reserva);null" json:"estadoReserva,omitempty"`
	INDICACIONES      *string        `orm:"column(indicaciones);null" json:"indicaciones,omitempty"`
	CREATED_AT        time.Time      `orm:"column(created_at);type(timestamp);auto_now_add" json:"createdAt"`
	UPDATED_AT        time.Time      `orm:"column(updated_at);type(timestamp);auto_now" json:"updatedAt"`
	CREATED_BY        *string        `orm:"column(created_by);type(text);null" json:"createdBy,omitempty"`
	UPDATED_BY        *string        `orm:"column(updated_by);type(text);null" json:"updatedBy,omitempty"`
	NOMBRE_COMPLETO   *string        `orm:"column(nombre_completo);type(text);null" json:"nombreCompleto,omitempty"`
	TELEFONO          *string        `orm:"column(telefono);type(text);null" json:"telefono,omitempty"`
	DOCUMENTO_CLIENTE *int64         `orm:"column(documento_cliente);null" json:"documentoCliente,omitempty"`
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
		Alias
	}{
		FECHA: t.FECHA.Format("02-01-2006"),
		Alias: (Alias)(t),
	})
}
