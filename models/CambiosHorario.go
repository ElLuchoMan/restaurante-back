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
	type Alias CambiosHorario

	// Cargar zona horaria de Bogotá
	loc, err := time.LoadLocation("America/Bogota")
	if err != nil {
		// Fallback a UTC-5 si no se puede cargar
		loc = time.FixedZone("UTC-5", -5*60*60)
	}

	return json.Marshal(&struct {
		FECHA string `json:"fechaCambioHorario"`
		Alias
	}{
		FECHA: t.FECHA.In(loc).Format("02-01-2006"),
		Alias: (Alias)(t),
	})
}
