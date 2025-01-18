package models

import (
	"encoding/json"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type Incidencia struct {
	PK_ID_INCIDENCIA        int64     `orm:"column(PK_ID_INCIDENCIA);pk;auto" json:"PK_ID_INCIDENCIA"`
	FECHA                   time.Time `orm:"column(FECHA);type(date)" json:"FECHA"`
	MONTO                   int64     `orm:"column(MONTO)" json:"MONTO"`
	RESTA                   bool      `orm:"column(RESTA);type(boolean)" json:"RESTA"`
	MOTIVO                  string    `orm:"column(MOTIVO);type(text)" json:"MOTIVO"`
	PK_DOCUMENTO_TRABAJADOR *int64    `orm:"column(PK_DOCUMENTO_TRABAJADOR);null" json:"PK_DOCUMENTO_TRABAJADOR,omitempty"`
}

func (i *Incidencia) TableName() string {
	return "INCIDENCIA"
}

func init() {
	orm.RegisterModel(new(Incidencia))
}

func (t Incidencia) MarshalJSON() ([]byte, error) {
	type Alias Incidencia
	return json.Marshal(&struct {
		FECHA string `json:"FECHA"`
		Alias
	}{
		FECHA: t.FECHA.Format("02-01-2006"),
		Alias: (Alias)(t),
	})
}
