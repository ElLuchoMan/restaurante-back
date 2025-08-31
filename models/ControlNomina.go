package models

import (
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type ControlNomina struct {
	PKIDControlNomina int64     `orm:"column(pk_id_control_nomina);pk;auto" json:"controlNominaId"`
	Fecha             time.Time `orm:"column(fecha);type(date)" json:"fecha"`
	Estado            string    `orm:"column(estado);type(text);default(NO GENERADA)" json:"estado"`
}

func (c *ControlNomina) TableName() string {
	return "control_nomina"
}

func init() {
	orm.RegisterModel(new(ControlNomina))
}
