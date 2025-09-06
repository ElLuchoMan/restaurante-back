package models

import (
	"encoding/json"

	"github.com/beego/beego/v2/client/orm"
)

type NominaTrabajador struct {
	PK_ID_NOMINA_TRABAJADOR int64       `orm:"column(pk_id_nomina_trabajador);pk;auto" json:"nominaTrabajadorId"`
	SUELDO_BASE             int64       `orm:"column(sueldo_base)" json:"sueldoBase"`
	MONTO_INCIDENCIAS       *int64      `orm:"column(monto_incidencias);null" json:"montoIncidencias,omitempty"`
	DETALLES                *string     `orm:"column(detalles);type(text);null" json:"detalles,omitempty"`
	PK_DOCUMENTO_TRABAJADOR *Trabajador `orm:"column(pk_documento_trabajador);rel(fk)" json:"documentoTrabajador" swaggertype:"integer"`
	PK_ID_NOMINA            *Nomina     `orm:"column(pk_id_nomina);rel(fk)" json:"nominaId" swaggertype:"integer"`
}

type NominaTrabajadorRequest struct {
	PK_DOCUMENTO_TRABAJADOR int64  `json:"documentoTrabajador" example:"1015466494"`
	DETALLES                string `json:"detalles,omitempty" example:"Pago correspondiente al mes de enero"`
}

type NominaTrabajadorResponse struct {
	SUELDO_BASE             int64  `json:"sueldoBase" example:"2000000"`
	MONTO_INCIDENCIAS       int64  `json:"montoIncidencias" example:"50000"`
	DETALLES                string `json:"detalles,omitempty" example:"Pago correspondiente al mes de enero"`
	PK_DOCUMENTO_TRABAJADOR int64  `json:"documentoTrabajador" example:"1015466494"`
}

type NominaTrabajadorDetalle struct {
	SUELDO_BASE             int64  `orm:"column(sueldo_base)" json:"sueldoBase"`
	MONTO_INCIDENCIAS       int64  `orm:"column(monto_incidencias)" json:"montoIncidencias"`
	DETALLES                string `orm:"column(detalles)" json:"detalles"`
	PK_DOCUMENTO_TRABAJADOR int64  `orm:"column(pk_documento_trabajador)" json:"documentoTrabajador"`
	PK_ID_NOMINA            int64  `orm:"column(pk_id_nomina)" json:"nominaId"`
	NOMBRE                  string `orm:"column(nombre)" json:"nombre"`
	APELLIDO                string `orm:"column(apellido)" json:"apellido"`
}

func (n *NominaTrabajador) TableName() string {
	return "nomina_trabajador"
}

func init() {
	orm.RegisterModel(new(NominaTrabajador))
}

func (n *NominaTrabajador) TableUnique() [][]string {
	return [][]string{
		{"PK_DOCUMENTO_TRABAJADOR", "PK_ID_NOMINA"},
	}
}

// UnmarshalJSON permite aceptar tanto claves explícitas como alternativas y
// números para referencias de FK. Soporta JSON con:
// - "nominaId" o "pk_id_nomina" (número o objeto Nomina)
// - "documentoTrabajador" o "pk_documento_trabajador" (número o objeto Trabajador)
func (n *NominaTrabajador) UnmarshalJSON(data []byte) error {
	type alias struct {
		SUELDO_BASE             *int64          `json:"sueldoBase,omitempty"`
		MONTO_INCIDENCIAS       *int64          `json:"montoIncidencias,omitempty"`
		DETALLES                *string         `json:"detalles,omitempty"`
		// posibles nombres entrantes
		DocumentoTrabajador     json.RawMessage `json:"documentoTrabajador,omitempty"`
		DocumentoTrabajadorAlt  json.RawMessage `json:"pk_documento_trabajador,omitempty"`
		NominaID                json.RawMessage `json:"nominaId,omitempty"`
		NominaIDAlt             json.RawMessage `json:"pk_id_nomina,omitempty"`
	}

	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}

	if a.SUELDO_BASE != nil {
		n.SUELDO_BASE = *a.SUELDO_BASE
	}
	n.MONTO_INCIDENCIAS = a.MONTO_INCIDENCIAS
	n.DETALLES = a.DETALLES

	// helper para procesar un RawMessage que puede ser número o objeto
	parseIDToTrabajador := func(raw json.RawMessage) (*Trabajador, error) {
		if len(raw) == 0 {
			return nil, nil
		}
		// intentar número
		var idNum int64
		if err := json.Unmarshal(raw, &idNum); err == nil {
			return &Trabajador{PK_DOCUMENTO_TRABAJADOR: idNum}, nil
		}
		// intentar objeto
		var t Trabajador
		if err := json.Unmarshal(raw, &t); err == nil {
			return &t, nil
		}
		return nil, nil
	}

	// Parse documento trabajador (prefiere el campo principal)
	if len(a.DocumentoTrabajador) != 0 {
		if tr, err := parseIDToTrabajador(a.DocumentoTrabajador); err == nil {
			n.PK_DOCUMENTO_TRABAJADOR = tr
		}
	} else if len(a.DocumentoTrabajadorAlt) != 0 {
		if tr, err := parseIDToTrabajador(a.DocumentoTrabajadorAlt); err == nil {
			n.PK_DOCUMENTO_TRABAJADOR = tr
		}
	}

	// Parse nomina id -> Nomina pointer
	parseIDToNomina := func(raw json.RawMessage) (*Nomina, error) {
		if len(raw) == 0 {
			return nil, nil
		}
		var idNum int64
		if err := json.Unmarshal(raw, &idNum); err == nil {
			return &Nomina{PK_ID_NOMINA: idNum}, nil
		}
		var m Nomina
		if err := json.Unmarshal(raw, &m); err == nil {
			return &m, nil
		}
		return nil, nil
	}

	if len(a.NominaID) != 0 {
		if nm, err := parseIDToNomina(a.NominaID); err == nil {
			n.PK_ID_NOMINA = nm
		}
	} else if len(a.NominaIDAlt) != 0 {
		if nm, err := parseIDToNomina(a.NominaIDAlt); err == nil {
			n.PK_ID_NOMINA = nm
		}
	}

	return nil
}
