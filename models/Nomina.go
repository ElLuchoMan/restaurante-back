package models

import (
	"encoding/json"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type Nomina struct {
	PK_ID_NOMINA  int64        `orm:"column(pk_id_nomina);pk;auto" json:"nominaId"`
	FECHA         time.Time    `orm:"column(fecha);type(date)" json:"fechaNomina"`
	MONTO         int64        `orm:"column(monto)" json:"monto"`
	ESTADO_NOMINA EstadoNomina `orm:"column(estado_nomina)" json:"estadoNomina"`
	CREATED_AT    time.Time    `orm:"column(created_at);type(timestamp);auto_now_add" json:"createdAt"`
	UPDATED_AT    time.Time    `orm:"column(updated_at);type(timestamp);auto_now" json:"updatedAt"`
	CREATED_BY    *string      `orm:"column(created_by);type(text)" json:"createdBy,omitempty"`
	UPDATED_BY    *string      `orm:"column(updated_by);type(text)" json:"updatedBy,omitempty"`
}

func (n *Nomina) TableName() string {
	return "nomina"
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
