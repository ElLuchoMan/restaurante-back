package models

import (
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type Nomina struct {
	PK_ID_NOMINA  int64     `orm:"column(PK_ID_NOMINA);pk;auto" json:"PK_ID_NOMINA"`
	FECHA         time.Time `orm:"column(FECHA);type(date)" json:"FECHA"`
	MONTO         int64     `orm:"column(MONTO)" json:"MONTO"`
	ESTADO_NOMINA string    `orm:"column(ESTADO_NOMINA)" json:"ESTADO_NOMINA"`
}

func (n *Nomina) TableName() string {
	return "NOMINA"
}

func init() {
	orm.RegisterModel(new(Nomina))
}
