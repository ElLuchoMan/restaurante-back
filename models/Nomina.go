package models

import (
	"encoding/json"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type Nomina struct {
	PK_ID_NOMINA  int64     `orm:"column(PK_ID_NOMINA);pk;auto" json:"nominaId"`
	FECHA         time.Time `orm:"column(FECHA);type(date)" json:"fechaNomina"`
	MONTO         int64     `orm:"column(MONTO)" json:"monto"`
	ESTADO_NOMINA string    `orm:"column(ESTADO_NOMINA)" json:"estadoNomina"`
}

func (n *Nomina) TableName() string {
	return "NOMINA"
}

func init() {
	orm.RegisterModel(new(Nomina))
}

func (t Nomina) MarshalJSON() ([]byte, error) {
	type Alias Nomina
	return json.Marshal(&struct {
		FECHA string `json:"fechaNomina"`
		Alias
	}{
		FECHA: t.FECHA.Format("02-01-2006"),
		Alias: (Alias)(t),
	})
}
