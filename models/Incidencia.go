package models

import (
	"encoding/json"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type Incidencia struct {
	PK_ID_INCIDENCIA        int64     `orm:"column(PK_ID_INCIDENCIA);pk;auto" json:"incidenciaId"`
	FECHA                   time.Time `orm:"column(FECHA);type(date)" json:"fechaIncidencia"`
	MONTO                   int64     `orm:"column(MONTO)" json:"monto"`
	RESTA                   bool      `orm:"column(RESTA);type(boolean)" json:"resta"`
	MOTIVO                  string    `orm:"column(MOTIVO);type(text)" json:"motivo"`
	PK_DOCUMENTO_TRABAJADOR *int64    `orm:"column(PK_DOCUMENTO_TRABAJADOR);null" json:"documentoTrabajador,omitempty"`
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
		FECHA string `json:"fechaIncidencia"`
		Alias
	}{
		FECHA: t.FECHA.Format("02-01-2006"),
		Alias: (Alias)(t),
	})
}
