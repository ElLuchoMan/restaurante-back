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
	// Cargar zona horaria de Bogotá
	loc, err := time.LoadLocation("America/Bogota")
	if err != nil {
		loc = time.FixedZone("UTC-5", -5*60*60)
	}

	// Para campos de tipo date: extraer SOLO año/mes/día sin importar timezone
	fechaStr := time.Date(t.FECHA.Year(), t.FECHA.Month(), t.FECHA.Day(), 0, 0, 0, 0, time.UTC).Format("02-01-2006")

	var horaAperturaStr *string
	if t.HORA_APERTURA != nil {
		horaApertura := t.HORA_APERTURA.In(loc)
		str := horaApertura.Format("15:04:05")
		horaAperturaStr = &str
	}

	horaCierreEnBogota := t.HORA_CIERRE.In(loc)

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
		HORA_CIERRE:          horaCierreEnBogota.Format("15:04:05"),
		ABIERTO:              t.ABIERTO,
	})
}
