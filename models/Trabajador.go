package models

import (
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type Trabajador struct {
	PK_DOCUMENTO_TRABAJADOR int64      `orm:"column(PK_DOCUMENTO_TRABAJADOR);pk" json:"pk_documento_trabajador"`
	NOMBRE                  string     `orm:"column(NOMBRE)" json:"nombre"`
	APELLIDO                string     `orm:"column(APELLIDO)" json:"apellido"`
	SUELDO                  int64      `orm:"column(SUELDO)" json:"sueldo"`
	TELEFONO                string     `orm:"column(TELEFONO)" json:"telefono"`
	FECHA_NACIMIENTO        *time.Time `orm:"column(FECHA_NACIMIENTO);type(date);null" json:"fecha_nacimiento"`
	NUEVO                   bool       `orm:"column(NUEVO)" json:"nuevo"`
	ROL                     string     `orm:"column(ROL)" json:"rol"`
	FECHA_INGRESO           time.Time  `orm:"column(FECHA_INGRESO);type(date)" json:"fecha_ingreso"`
	FECHA_RETIRO            *time.Time `orm:"column(FECHA_RETIRO);type(date);null" json:"fecha_retiro"`
	PASSWORD                string     `orm:"column(PASSWORD)" json:"-"`
	PK_ID_RESTAURANTE       *int64     `orm:"column(PK_ID_RESTAURANTE);null" json:"pk_id_restaurante"`
}

func (t *Trabajador) TableName() string {
	return "TRABAJADOR"
}

func init() {
	orm.RegisterModel(new(Trabajador))
}
