package models

import (
	"encoding/json"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type Reserva struct {
	PK_ID_RESERVA     int64            `orm:"column(pk_id_reserva);pk;auto" json:"reservaId"`
	FECHA             time.Time        `orm:"column(fecha);type(date)" json:"fechaReserva"`
	HORA              time.Time        `orm:"column(hora);type(time);size(8)" json:"horaReserva"`
	PERSONAS          int              `orm:"column(personas)" json:"personas"`
	PK_ID_CONTACTO    *ReservaContacto `orm:"column(pk_id_contacto);rel(fk)" json:"contactoId" swaggertype:"integer"`
	PK_ID_RESTAURANTE *Restaurante     `orm:"column(pk_id_restaurante);rel(fk)" json:"restauranteId" swaggertype:"integer"`
	ESTADO_RESERVA    *EstadoReserva   `orm:"column(estado_reserva);type(estado_reserva);null" json:"estadoReserva,omitempty"`
	INDICACIONES      *string          `orm:"column(indicaciones);null" json:"indicaciones,omitempty"`
	CREATED_AT        time.Time        `orm:"column(created_at);type(timestamptz);auto_now_add" json:"createdAt" swaggertype:"string"`
	UPDATED_AT        time.Time        `orm:"column(updated_at);type(timestamptz);auto_now" json:"updatedAt" swaggertype:"string"`
	CREATED_BY        *string          `orm:"column(created_by);type(text);null" json:"createdBy,omitempty"`
	UPDATED_BY        *string          `orm:"column(updated_by);type(text);null" json:"updatedBy,omitempty"`
}

func (r *Reserva) TableName() string {
	return "reserva"
}

func init() {
	orm.RegisterModel(new(Reserva))
}

func (t Reserva) MarshalJSON() ([]byte, error) {

	fechaStr := FormatDateUTC(t.FECHA)

	horaStr := FormatTimeWithLMT(t.HORA)

	createdAtStr := FormatTimestampBogota(t.CREATED_AT)
	updatedAtStr := FormatTimestampBogota(t.UPDATED_AT)

	return json.Marshal(&struct {
		PK_ID_RESERVA     int64            `json:"reservaId"`
		FECHA             string           `json:"fechaReserva"`
		HORA              string           `json:"horaReserva"`
		PERSONAS          int              `json:"personas"`
		PK_ID_CONTACTO    *ReservaContacto `json:"contactoId"`
		PK_ID_RESTAURANTE *Restaurante     `json:"restauranteId"`
		ESTADO_RESERVA    *EstadoReserva   `json:"estadoReserva,omitempty"`
		INDICACIONES      *string          `json:"indicaciones,omitempty"`
		CREATED_AT        string           `json:"createdAt"`
		UPDATED_AT        string           `json:"updatedAt"`
		CREATED_BY        *string          `json:"createdBy,omitempty"`
		UPDATED_BY        *string          `json:"updatedBy,omitempty"`
	}{
		PK_ID_RESERVA:     t.PK_ID_RESERVA,
		FECHA:             fechaStr,
		HORA:              horaStr,
		PERSONAS:          t.PERSONAS,
		PK_ID_CONTACTO:    t.PK_ID_CONTACTO,
		PK_ID_RESTAURANTE: t.PK_ID_RESTAURANTE,
		ESTADO_RESERVA:    t.ESTADO_RESERVA,
		INDICACIONES:      t.INDICACIONES,
		CREATED_AT:        createdAtStr,
		UPDATED_AT:        updatedAtStr,
		CREATED_BY:        t.CREATED_BY,
		UPDATED_BY:        t.UPDATED_BY,
	})
}
