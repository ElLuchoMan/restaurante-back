package models

import "github.com/beego/beego/v2/client/orm"

type NominaTrabajador struct {
	PK_ID_NOMINA_TRABAJADOR int64   `orm:"column(pk_id_nomina_trabajador);pk;auto" json:"PK_ID_NOMINA_TRABAJADOR"`
	SUELDO_BASE             int64   `orm:"column(sueldo_base)" json:"SUELDO_BASE"`
	MONTO_INCIDENCIAS       *int64  `orm:"column(monto_incidencias);null" json:"MONTO_INCIDENCIAS,omitempty"`
	TOTAL                   *int64  `orm:"column(total);null" json:"TOTAL,omitempty"`
	DETALLES                *string `orm:"column(detalles);type(text);null" json:"DETALLES,omitempty"`
	PK_DOCUMENTO_TRABAJADOR int64   `orm:"column(pk_documento_trabajador);null" json:"PK_DOCUMENTO_TRABAJADOR,omitempty"`
	PK_ID_NOMINA            *int64  `orm:"column(pk_id_nomina);null" json:"PK_ID_NOMINA,omitempty"`
}

type NominaTrabajadorRequest struct {
	PK_DOCUMENTO_TRABAJADOR int64  `json:"PK_DOCUMENTO_TRABAJADOR" example:"1015466494"`
	DETALLES                string `json:"DETALLES,omitempty" example:"Pago correspondiente al mes de enero"`
}

type NominaTrabajadorResponse struct {
	PK_ID_NOMINA_TRABAJADOR int64  `json:"PK_ID_NOMINA_TRABAJADOR" example:"1"`
	SUELDO_BASE             int64  `json:"SUELDO_BASE" example:"2000000"`
	MONTO_INCIDENCIAS       int64  `json:"MONTO_INCIDENCIAS" example:"50000"`
	TOTAL                   int64  `json:"TOTAL" example:"2050000"`
	DETALLES                string `json:"DETALLES,omitempty" example:"Pago correspondiente al mes de enero"`
}

type NominaTrabajadorDetalle struct {
	PK_ID_NOMINA_TRABAJADOR int64  `orm:"column(pk_id_nomina_trabajador)" json:"nominaTrabajadorId"`
	SUELDO_BASE             int64  `orm:"column(sueldo_base)" json:"sueldoBase"`
	MONTO_INCIDENCIAS       int64  `orm:"column(monto_incidencias)" json:"montoIncidencias"`
	TOTAL                   int64  `orm:"column(total)" json:"total"`
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
