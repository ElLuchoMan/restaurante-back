package models

import (
	"encoding/json"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type CambiosHorario struct {
	PK_ID_CAMBIO_HORARIO int64      `orm:"column(PK_ID_CAMBIO_HORARIO);pk;auto" json:"cambioHorarioId"`
	FECHA                time.Time  `orm:"column(FECHA);type(date)" json:"fechaCambioHorario"`
	HORA_APERTURA        *time.Time `orm:"column(HORA_APERTURA);type(time);null" json:"horaApertura,omitempty"`
	HORA_CIERRE          *time.Time `orm:"column(HORA_CIERRE);type(time);null" json:"horaCierre,omitempty"`
	ABIERTO              bool       `orm:"column(ABIERTO)" json:"abierto"`
}

func (t *CambiosHorario) TableName() string {
	return "CAMBIOS_HORARIO"
}
func init() {
	orm.RegisterModel(new(CambiosHorario))
}

func (t CambiosHorario) MarshalJSON() ([]byte, error) {
	type Alias CambiosHorario
	return json.Marshal(&struct {
		FECHA string `json:"fechaCambioHorario"`
		Alias
	}{
		FECHA: t.FECHA.Format("02-01-2006"),
		Alias: (Alias)(t),
	})
}
