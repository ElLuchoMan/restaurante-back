package models

import (
	"encoding/json"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type Nomina struct {
	PK_ID_NOMINA  int64        `orm:"column(pk_id_nomina);pk;auto" json:"nominaId"`
	FECHA         time.Time    `orm:"column(fecha);type(date);unique" json:"fechaNomina"`
	MONTO         int64        `orm:"column(monto)" json:"monto"`
	ESTADO_NOMINA EstadoNomina `orm:"column(estado_nomina);type(estado_nomina)" json:"estadoNomina"`
}

// NominaResponse es EXACTAMENTE lo que el API devuelve por cada nómina
// (Nomina.MarshalJSON). fechaNomina se responde como DD-MM-YYYY.
type NominaResponse struct {
	NominaID     int64  `json:"nominaId" example:"5"`
	FechaNomina  string `json:"fechaNomina" example:"20-01-2025" description:"DD-MM-YYYY"`
	Monto        int64  `json:"monto" example:"4500000" description:"Lo calcula la base de datos (trigger); el cliente no puede fijarlo"`
	EstadoNomina string `json:"estadoNomina" enums:"PAGO,NO_PAGO" example:"NO_PAGO"`
}

// NominaCreateRequest es el cuerpo de POST /nominas. nominaId y monto NO se
// aceptan (se ignoran): el id lo genera la base y el monto lo calcula el trigger.
type NominaCreateRequest struct {
	FechaNomina  *string `json:"fechaNomina,omitempty" example:"2025-01-20" description:"YYYY-MM-DD; el día debe ser >= 20. Por defecto, hoy"`
	EstadoNomina *string `json:"estadoNomina,omitempty" enums:"PAGO,NO_PAGO" default:"NO_PAGO" example:"NO_PAGO"`
}

// NominaUpdateRequest es el cuerpo opcional de PUT /nominas. El único campo
// editable es estadoNomina (no admite null). Sin cuerpo, o sin ese campo, la
// nómina se marca PAGO. fechaNomina, monto y nominaId se ignoran.
type NominaUpdateRequest struct {
	EstadoNomina *string `json:"estadoNomina,omitempty" enums:"PAGO,NO_PAGO" default:"PAGO" example:"PAGO"`
}

func (n *Nomina) TableName() string {
	return "nomina"
}

func init() {
	orm.RegisterModel(new(Nomina))
}

func (t Nomina) MarshalJSON() ([]byte, error) {
	return json.Marshal(NominaResponse{
		NominaID:     t.PK_ID_NOMINA,
		FechaNomina:  FormatDateUTC(t.FECHA),
		Monto:        t.MONTO,
		EstadoNomina: t.ESTADO_NOMINA,
	})
}
