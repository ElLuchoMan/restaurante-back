package models

import (
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type ControlNomina struct {
	PK_ID_CONTROL_NOMINA int64               `orm:"column(pk_id_control_nomina);pk;auto" json:"controlNominaId"`
	Fecha                time.Time           `orm:"column(fecha);type(date);unique" json:"fecha"`
	Estado               EstadoControlNomina `orm:"column(estado);type(text);default(NO GENERADA)" json:"estado"`
}

// Validate ensures Estado has a permitted value.
func (c ControlNomina) Validate() bool {
	return c.Estado.IsValid()
}

func (c *ControlNomina) TableName() string {
	return "control_nomina"
}

func init() {
	orm.RegisterModel(new(ControlNomina))
}
