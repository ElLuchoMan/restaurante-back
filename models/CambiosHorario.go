package models

import (
	"encoding/json"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type CambiosHorario struct {
	PK_ID_CAMBIO_HORARIO int64      `orm:"column(pk_id_cambio_horario);pk;auto" json:"cambioHorarioId"`
	FECHA                time.Time  `orm:"column(fecha);type(date)" json:"fechaCambioHorario"`
	HORA_APERTURA        *time.Time `orm:"column(hora_apertura);type(time);null" json:"horaApertura,omitempty"`
	HORA_CIERRE          time.Time  `orm:"column(hora_cierre);type(time)" json:"horaCierre"`
	ABIERTO              bool       `orm:"column(abierto)" json:"abierto"`
}

func (t *CambiosHorario) TableName() string {
	return "cambios_horario"
}
func init() {
	orm.RegisterModel(new(CambiosHorario))
}

func (t CambiosHorario) MarshalJSON() ([]byte, error) {

	fechaStr := FormatDateUTC(t.FECHA)

	var horaAperturaStr *string
	if t.HORA_APERTURA != nil {
		str := FormatTimeWithLMT(*t.HORA_APERTURA)
		horaAperturaStr = &str
	}

	horaCierreStr := FormatTimeWithLMT(t.HORA_CIERRE)

	return json.Marshal(&struct {
		PK_ID_CAMBIO_HORARIO int64   `json:"cambioHorarioId"`
		FECHA                string  `json:"fechaCambioHorario"`
		HORA_APERTURA        *string `json:"horaApertura,omitempty"`
		HORA_CIERRE          string  `json:"horaCierre"`
		ABIERTO              bool    `json:"abierto"`
	}{
		PK_ID_CAMBIO_HORARIO: t.PK_ID_CAMBIO_HORARIO,
		FECHA:                fechaStr,
		HORA_APERTURA:        horaAperturaStr,
		HORA_CIERRE:          horaCierreStr,
		ABIERTO:              t.ABIERTO,
	})
}
