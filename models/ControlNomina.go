package models

import (
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type ControlNomina struct {
	PKIDControlNomina    int64        `orm:"column(pk_id_control_nomina);pk;auto" json:"controlNominaId"`
	PKIDNominaTrabajador int64        `orm:"column(pk_id_nomina_trabajador)" json:"nominaTrabajadorId"`
	EstadoNomina         EstadoNomina `orm:"column(estado_nomina)" json:"estadoNomina"`
	CreatedAt            time.Time    `orm:"column(created_at);type(timestamp);auto_now_add" json:"createdAt"`
	UpdatedAt            time.Time    `orm:"column(updated_at);type(timestamp);auto_now" json:"updatedAt"`
	CreatedBy            *string      `orm:"column(created_by);type(text)" json:"createdBy,omitempty"`
	UpdatedBy            *string      `orm:"column(updated_by);type(text)" json:"updatedBy,omitempty"`
}

func (c *ControlNomina) TableName() string {
	return "control_nomina"
}

func init() {
	orm.RegisterModel(new(ControlNomina))
}
