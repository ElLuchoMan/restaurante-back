package models

import (
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type Trabajador struct {
	PK_DOCUMENTO_TRABAJADOR int64      `orm:"column(PK_DOCUMENTO_TRABAJADOR);pk" json:"PK_DOCUMENTO_TRABAJADOR"`
	NOMBRE                  string     `orm:"column(NOMBRE)" json:"NOMBRE"`
	APELLIDO                string     `orm:"column(APELLIDO)" json:"APELLIDO"`
	SUELDO                  int64      `orm:"column(SUELDO)" json:"SUELDO"`
	TELEFONO                string     `orm:"column(TELEFONO)" json:"TELEFONO"`
	FECHA_NACIMIENTO        *time.Time `orm:"column(FECHA_NACIMIENTO);type(date);null" json:"FECHA_NACIMIENTO"`
	NUEVO                   bool       `orm:"column(NUEVO)" json:"NUEVO"`
	ROL                     string     `orm:"column(ROL)" json:"ROL"`
	FECHA_INGRESO           time.Time  `orm:"column(FECHA_INGRESO);type(date)" json:"FECHA_INGRESO"`
	FECHA_RETIRO            *time.Time `orm:"column(FECHA_RETIRO);type(date);null" json:"FECHA_RETIRO"`
	PASSWORD                string     `orm:"column(PASSWORD)" json:"PASSWORD"`
	HORARIO                 string     `orm:"column(HORARIO);type(text)" json:"HORARIO"`
	PK_ID_RESTAURANTE       *int64     `orm:"column(PK_ID_RESTAURANTE);null" json:"PK_ID_RESTAURANTE"`
}

func (t *Trabajador) TableName() string {
	return "TRABAJADOR"
}

func init() {
	orm.RegisterModel(new(Trabajador))
}
