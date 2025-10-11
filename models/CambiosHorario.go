package models

import (
	"encoding/json"
	"fmt"
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
	// FECHA: normalizar a UTC para obtener el día de calendario correcto, sin efectos de zona
	fechaUTC := t.FECHA.UTC()
	fechaStr := fmt.Sprintf("%02d-%02d-%04d", fechaUTC.Day(), int(fechaUTC.Month()), fechaUTC.Year())

	// HORA: algunos timezones históricos (LMT) en America/Bogota afectan horas con año 0000
	// Detectamos año antiguo y ajustamos con doble desfase LMT (~09:52:32) para recuperar hora de pared
	var horaAperturaStr *string
	if t.HORA_APERTURA != nil {
		horaAdj := *t.HORA_APERTURA
		if horaAdj.Year() < 1900 {
			horaAdj = horaAdj.Add(9*time.Hour + 52*time.Minute + 32*time.Second)
		}
		str := fmt.Sprintf("%02d:%02d:%02d", horaAdj.Hour(), horaAdj.Minute(), horaAdj.Second())
		horaAperturaStr = &str
	}

	horaCierreAdj := t.HORA_CIERRE
	if horaCierreAdj.Year() < 1900 {
		horaCierreAdj = horaCierreAdj.Add(9*time.Hour + 52*time.Minute + 32*time.Second)
	}
	horaCierreStr := fmt.Sprintf("%02d:%02d:%02d", horaCierreAdj.Hour(), horaCierreAdj.Minute(), horaCierreAdj.Second())

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
