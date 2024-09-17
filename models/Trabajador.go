package models

import "github.com/beego/beego/v2/client/orm"

type Trabajador struct {
	PK_DOCUMENTO_TRABAJADOR int64   `orm:"column(PK_DOCUMENTO_TRABAJADOR);pk" json:"PK_DOCUMENTO_TRABAJADOR"`
	NOMBRE                  string  `orm:"column(NOMBRE)" json:"NOMBRE"`
	APELLIDO                string  `orm:"column(APELLIDO)" json:"APELLIDO"`
	SUELDO                  int64   `orm:"column(SUELDO)" json:"SUELDO"`
	TELEFONO                int64   `orm:"column(TELEFONO)" json:"TELEFONO"`
	FECHA_NACIMIENTO        *string `orm:"column(FECHA_NACIMIENTO);type(date);null" json:"FECHA_NACIMIENTO"`
	NUEVO                   bool    `orm:"column(NUEVO)" json:"NUEVO"`
	ROL                     string  `orm:"column(ROL)" json:"ROL"`
	FECHA_INGRESO           string  `orm:"column(FECHA_INGRESO);type(date)" json:"FECHA_INGRESO"`
	FECHA_RETIRO            *string `orm:"column(FECHA_RETIRO);type(date);null" json:"FECHA_RETIRO"`
	PASSWORD                string  `orm:"column(PASSWORD)" json:"PASSWORD"`
	PK_ID_RESTAURANTE       *int64  `orm:"column(PK_ID_RESTAURANTE);null" json:"PK_ID_RESTAURANTE"`
}

func (c *Trabajador) TableName() string {
	return "TRABAJADOR"
}

func init() {
	orm.RegisterModel(new(Trabajador))
}
