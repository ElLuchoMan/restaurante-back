package models

import "github.com/beego/beego/v2/client/orm"

type NominaTrabajador struct {
	PK_ID_NOMINA_TRABAJADOR int64   `orm:"column(PK_ID_NOMINA_TRABAJADOR);pk;auto" json:"PK_ID_NOMINA_TRABAJADOR"`
	SUELDO_BASE             int64   `orm:"column(SUELDO_BASE)" json:"SUELDO_BASE"`
	MONTO_INCIDENCIAS       *int64  `orm:"column(MONTO_INCIDENCIAS);null" json:"MONTO_INCIDENCIAS,omitempty"`
	TOTAL                   *int64  `orm:"column(TOTAL);null" json:"TOTAL,omitempty"`
	DETALLES                *string `orm:"column(DETALLES);type(text);null" json:"DETALLES,omitempty"`
	PK_DOCUMENTO_TRABAJADOR int64   `orm:"column(PK_DOCUMENTO_TRABAJADOR);null" json:"PK_DOCUMENTO_TRABAJADOR,omitempty"`
	PK_ID_NOMINA            *int64  `orm:"column(PK_ID_NOMINA);null" json:"PK_ID_NOMINA,omitempty"`
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

func (n *NominaTrabajador) TableName() string {
	return "NOMINA_TRABAJADOR"
}

func init() {
	orm.RegisterModel(new(NominaTrabajador))
}
