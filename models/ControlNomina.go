package models

import (
	"encoding/json"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type ControlNomina struct {
	PK_ID_CONTROL_NOMINA int64               `orm:"column(pk_id_control_nomina);pk;auto" json:"controlNominaId"`
	Fecha                time.Time           `orm:"column(fecha);type(date);unique" json:"fecha"`
	Estado               EstadoControlNomina `orm:"column(estado);type(text);default(NO GENERADA)" json:"estado"`
}

// ControlNominaResponse es EXACTAMENTE lo que el API devuelve por cada
// registro de control de nómina (ControlNomina.MarshalJSON). fecha se responde
// como DD-MM-YYYY.
type ControlNominaResponse struct {
	ControlNominaID int64  `json:"controlNominaId" example:"2"`
	Fecha           string `json:"fecha" example:"20-01-2025" description:"DD-MM-YYYY"`
	Estado          string `json:"estado" enums:"NO GENERADA,GENERADA,REGENERADA" example:"GENERADA"`
}

func (c *ControlNomina) TableName() string {
	return "control_nomina"
}

func init() {
	orm.RegisterModel(new(ControlNomina))
}

func (c ControlNomina) MarshalJSON() ([]byte, error) {
	return json.Marshal(ControlNominaResponse{
		ControlNominaID: c.PK_ID_CONTROL_NOMINA,
		Fecha:           FormatDateUTC(c.Fecha),
		Estado:          string(c.Estado),
	})
}
