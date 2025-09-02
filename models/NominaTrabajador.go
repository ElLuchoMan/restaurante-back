package models

import "github.com/beego/beego/v2/client/orm"

type NominaTrabajador struct {
	PK_ID_NOMINA_TRABAJADOR int64   `orm:"column(pk_id_nomina_trabajador);pk;auto" json:"nominaTrabajadorId"`
	SUELDO_BASE             int64   `orm:"column(sueldo_base)" json:"sueldoBase"`
	MONTO_INCIDENCIAS       *int64  `orm:"column(monto_incidencias);null" json:"montoIncidencias,omitempty"`
	DETALLES                *string `orm:"column(detalles);type(text);null" json:"detalles,omitempty"`
	PK_DOCUMENTO_TRABAJADOR int64   `orm:"column(pk_documento_trabajador);rel(fk)" json:"documentoTrabajador"`
	PK_ID_NOMINA            int64   `orm:"column(pk_id_nomina);rel(fk)" json:"nominaId"`
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
