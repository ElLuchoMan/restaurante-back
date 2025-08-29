package models

import (
	"encoding/json"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type HorarioTrabajador struct {
	PK_DOCUMENTO_TRABAJADOR int64     `orm:"column(PK_DOCUMENTO_TRABAJADOR);pk" json:"documentoTrabajador"`
	DIA                     string    `orm:"column(DIA);type(text)" json:"dia"`
	HORA_INICIO             time.Time `orm:"column(HORA_INICIO);type(time)" json:"horaInicio"`
	HORA_FIN                time.Time `orm:"column(HORA_FIN);type(time)" json:"horaFin"`
}

func (h *HorarioTrabajador) TableName() string {
	return "HORARIO_TRABAJADOR"
}

func init() {
	orm.RegisterModel(new(HorarioTrabajador))
}

func (h HorarioTrabajador) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		PK_DOCUMENTO_TRABAJADOR int64  `json:"documentoTrabajador"`
		DIA                     string `json:"dia"`
		HORA_INICIO             string `json:"horaInicio"`
		HORA_FIN                string `json:"horaFin"`
	}{
		PK_DOCUMENTO_TRABAJADOR: h.PK_DOCUMENTO_TRABAJADOR,
		DIA:                     h.DIA,
		HORA_INICIO:             h.HORA_INICIO.Format("15:04:05"),
		HORA_FIN:                h.HORA_FIN.Format("15:04:05"),
	})
}