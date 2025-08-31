package models

import (
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type ControlNomina struct {
	Fecha  time.Time `orm:"column(fecha);type(date)" json:"fecha"`
	Estado string    `orm:"column(estado);type(text);default(NO GENERADA)" json:"estado"`
}

func (c *ControlNomina) TableName() string {
	return "control_nomina"
}

func init() {
	orm.RegisterModel(new(ControlNomina))
}
