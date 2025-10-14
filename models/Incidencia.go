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

	fechaStr := FormatDateUTC(t.FECHA)

	var docTrabajador int64
	if t.PK_DOCUMENTO_TRABAJADOR != nil {
		docTrabajador = t.PK_DOCUMENTO_TRABAJADOR.PK_DOCUMENTO_TRABAJADOR
	}

	return json.Marshal(&struct {
		PK_ID_INCIDENCIA        int64  `json:"incidenciaId"`
		FECHA                   string `json:"fechaIncidencia"`
		MONTO                   int64  `json:"monto"`
		RESTA                   bool   `json:"resta"`
		MOTIVO                  string `json:"motivo"`
		PK_DOCUMENTO_TRABAJADOR int64  `json:"documentoTrabajador"`
	}{
		PK_ID_INCIDENCIA:        t.PK_ID_INCIDENCIA,
		FECHA:                   fechaStr,
		MONTO:                   t.MONTO,
		RESTA:                   t.RESTA,
		MOTIVO:                  t.MOTIVO,
		PK_DOCUMENTO_TRABAJADOR: docTrabajador,
	})
}
