package models

import (
	"time"

	"fmt"
	"github.com/beego/beego/v2/client/orm"
)

type ControlNomina struct {
	PK_ID_CONTROL_NOMINA int64               `orm:"column(pk_id_control_nomina);pk;auto" json:"controlNominaId"`
	Fecha                time.Time           `orm:"column(fecha);type(date);unique" json:"fecha"`
	Estado               EstadoControlNomina `orm:"column(estado);type(text);default(NO GENERADA)" json:"estado"`
}

func (c *ControlNomina) TableName() string {
	return "control_nomina"
}

func init() {
	orm.RegisterModel(new(ControlNomina))
}

// ValidEstado indica si el estado actual está permitido
func (c *ControlNomina) ValidEstado() bool {
	return c.Estado.IsValid()
}

// Insert validates Estado before inserting
type simpleOrmer interface {
	Insert(interface{}) (int64, error)
	Update(interface{}, ...string) (int64, error)
}

func (c *ControlNomina) Insert(o simpleOrmer) (int64, error) {
	if !c.ValidEstado() {
		return 0, fmt.Errorf("estado inválido: %s", c.Estado)
	}
	return o.Insert(c)
}

// Update validates Estado before updating
func (c *ControlNomina) Update(o simpleOrmer, cols ...string) (int64, error) {
	if !c.ValidEstado() {
		return 0, fmt.Errorf("estado inválido: %s", c.Estado)
	}
	return o.Update(c, cols...)
}
