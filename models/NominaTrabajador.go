package models

import (
	"github.com/beego/beego/v2/client/orm"
)

type NominaTrabajador struct {
	PK_ID_NOMINA_TRABAJADOR int64       `orm:"column(pk_id_nomina_trabajador);pk;auto" json:"nominaTrabajadorId"`
	SUELDO_BASE             int64       `orm:"column(sueldo_base)" json:"sueldoBase"`
	MONTO_INCIDENCIAS       *int64      `orm:"column(monto_incidencias);null" json:"montoIncidencias,omitempty"`
	DETALLES                *string     `orm:"column(detalles);type(text);null" json:"detalles,omitempty"`
	PK_DOCUMENTO_TRABAJADOR *Trabajador `orm:"column(pk_documento_trabajador);rel(fk)" json:"documentoTrabajador"`
	PK_ID_NOMINA            *Nomina     `orm:"column(pk_id_nomina);rel(fk)" json:"nominaId"`
}

// NominaTrabajadorRequest es el cuerpo de POST /nomina_trabajador. Cualquier
// otro campo (p. ej. detalles) se ignora: sueldo, incidencias y detalle los
// calcula el backend a partir de la última nómina.
type NominaTrabajadorRequest struct {
	PK_DOCUMENTO_TRABAJADOR int64 `json:"documentoTrabajador" binding:"required" example:"1015466494"`
}

// NominaTrabajadorItem es la forma única de una relación nómina-trabajador en
// GET /nomina_trabajador, GET /nomina_trabajador/search y POST /nomina_trabajador.
// Las FK se responden como ids numéricos; montoIncidencias y detalles nulos en
// base de datos se responden como 0 y "".
type NominaTrabajadorItem struct {
	PK_ID_NOMINA_TRABAJADOR int64  `orm:"column(pk_id_nomina_trabajador)" json:"nominaTrabajadorId" example:"15"`
	SUELDO_BASE             int64  `orm:"column(sueldo_base)" json:"sueldoBase" example:"2000000"`
	MONTO_INCIDENCIAS       int64  `orm:"column(monto_incidencias)" json:"montoIncidencias" example:"50000"`
	DETALLES                string `orm:"column(detalles)" json:"detalles" example:"Nómina del mes de Enero de 2025 más incidencias si aplica"`
	PK_DOCUMENTO_TRABAJADOR int64  `orm:"column(pk_documento_trabajador)" json:"documentoTrabajador" example:"1015466494"`
	PK_ID_NOMINA            int64  `orm:"column(pk_id_nomina)" json:"nominaId" example:"5"`
}

// NominaTrabajadorDetalle es la fila de GET /nomina_trabajador/mes: lo mismo
// que NominaTrabajadorItem más el nombre y apellido del trabajador.
type NominaTrabajadorDetalle struct {
	PK_ID_NOMINA_TRABAJADOR int64  `orm:"column(pk_id_nomina_trabajador)" json:"nominaTrabajadorId" example:"15"`
	SUELDO_BASE             int64  `orm:"column(sueldo_base)" json:"sueldoBase" example:"2000000"`
	MONTO_INCIDENCIAS       int64  `orm:"column(monto_incidencias)" json:"montoIncidencias" example:"50000"`
	DETALLES                string `orm:"column(detalles)" json:"detalles" example:"Nómina del mes de Enero de 2025 más incidencias si aplica"`
	PK_DOCUMENTO_TRABAJADOR int64  `orm:"column(pk_documento_trabajador)" json:"documentoTrabajador" example:"1015466494"`
	PK_ID_NOMINA            int64  `orm:"column(pk_id_nomina)" json:"nominaId" example:"5"`
	NOMBRE                  string `orm:"column(nombre)" json:"nombre" example:"Juan"`
	APELLIDO                string `orm:"column(apellido)" json:"apellido" example:"Pérez"`
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
