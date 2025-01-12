package models

import (
	"encoding/json"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type CambiosHorario struct {
	PK_ID_CAMBIO_HORARIO int64      `orm:"column(PK_ID_CAMBIO_HORARIO);pk;auto" json:"PK_ID_CAMBIO_HORARIO"`
	FECHA                time.Time  `orm:"column(FECHA);type(date)" json:"FECHA"`
	HORA_APERTURA        *time.Time `orm:"column(HORA_APERTURA);type(time);null" json:"HORA_APERTURA,omitempty"`
	HORA_CIERRE          *time.Time `orm:"column(HORA_CIERRE);type(time);null" json:"HORA_CIERRE,omitempty"`
	ABIERTO              bool       `orm:"column(ABIERTO)" json:"ABIERTO"`
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
		FECHA string `json:"FECHA"`
		Alias
	}{
		FECHA: t.FECHA.Format("2006-01-02"),
		Alias: (Alias)(t),
	})
}
