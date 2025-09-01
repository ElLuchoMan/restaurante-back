package models

import (
	"encoding/json"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type HorarioTrabajador struct {
	PK_ID_HORARIO_TRABAJADOR int64     `orm:"column(pk_id_horario_trabajador);pk;auto" json:"horarioTrabajadorId"`
	PK_DOCUMENTO_TRABAJADOR  int64     `orm:"column(pk_documento_trabajador);unique(trabajador_dia)" json:"documentoTrabajador"`
	DIA                      DiaSemana `orm:"column(dia);type(text);unique(trabajador_dia)" json:"dia"`
	HORA_INICIO              time.Time `orm:"column(hora_inicio);type(time)" json:"horaInicio"`
	HORA_FIN                 time.Time `orm:"column(hora_fin);type(time)" json:"horaFin"`
}

func (h *HorarioTrabajador) TableName() string {
	return "horario_trabajador"
}

func init() {
	orm.RegisterModel(new(HorarioTrabajador))
}

// ValidateHoras ensures hora fin is after hora inicio.
func (h HorarioTrabajador) ValidateHoras() bool {
	return h.HORA_FIN.After(h.HORA_INICIO)
}

func (h HorarioTrabajador) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		PK_DOCUMENTO_TRABAJADOR int64  `json:"documentoTrabajador"`
		DIA                     string `json:"dia"`
		HORA_INICIO             string `json:"horaInicio"`
		HORA_FIN                string `json:"horaFin"`
	}{
		PK_DOCUMENTO_TRABAJADOR: h.PK_DOCUMENTO_TRABAJADOR,
		DIA:                     string(h.DIA),
		HORA_INICIO:             h.HORA_INICIO.Format("15:04:05"),
		HORA_FIN:                h.HORA_FIN.Format("15:04:05"),
	})
}
