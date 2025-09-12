package models

import (
	"encoding/json"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type Incidencia struct {
	PK_ID_INCIDENCIA        int64       `orm:"column(pk_id_incidencia);pk;auto" json:"incidenciaId"`
	FECHA                   time.Time   `orm:"column(fecha);type(date)" json:"fechaIncidencia"`
	MONTO                   int64       `orm:"column(monto)" json:"monto"`
	RESTA                   bool        `orm:"column(resta);type(boolean)" json:"resta"`
	MOTIVO                  string      `orm:"column(motivo);type(text)" json:"motivo"`
	PK_DOCUMENTO_TRABAJADOR *Trabajador `orm:"column(pk_documento_trabajador);rel(fk)" json:"documentoTrabajador" swaggertype:"integer"`
}

func (i *Incidencia) TableName() string {
	return "incidencia"
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
