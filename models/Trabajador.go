package models

import "github.com/beego/beego/v2/client/orm"

type Trabajador struct {
	PK_DOCUMENTO_TRABAJADOR int64   `orm:"pk" json:"PK_DOCUMENTO_TRABAJADOR"`
	NOMBRE                  string  `json:"NOMBRE"`
	APELLIDO                string  `json:"APELLIDO"`
	SUELDO                  int64   `json:"SUELDO"`
	TELEFONO                int64   `json:"TELEFONO"`
	FECHA_NACIMIENTO        *string `orm:"null" json:"FECHA_NACIMIENTO"`
	NUEVO                   bool    `json:"NUEVO"`
	ROL                     string  `json:"ROL"`
	FECHA_INGRESO           string  `json:"FECHA_INGRESO"`
	FECHA_RETIRO            *string `orm:"null" json:"FECHA_RETIRO"`
	PASSWORD                string  `json:"PASSWORD"`
	PK_ID_RESTAURANTE       *int64  `orm:"null" json:"PK_ID_RESTAURANTE"`
}

func (c *Trabajador) TableName() string {
	return "TRABAJADOR"
}

func init() {
	orm.RegisterModel(new(Trabajador))
}
