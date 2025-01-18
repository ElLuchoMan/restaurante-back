package models

import (
	"encoding/json"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type Trabajador struct {
	PK_DOCUMENTO_TRABAJADOR int64      `orm:"column(PK_DOCUMENTO_TRABAJADOR);pk" json:"PK_DOCUMENTO_TRABAJADOR"`
	NOMBRE                  string     `orm:"column(NOMBRE);type(text)" json:"NOMBRE"`
	APELLIDO                string     `orm:"column(APELLIDO);type(text)" json:"APELLIDO"`
	SUELDO                  int64      `orm:"column(SUELDO)" json:"SUELDO"`
	TELEFONO                *string    `orm:"column(TELEFONO);type(text);null" json:"TELEFONO,omitempty"`
	FECHA_NACIMIENTO        *time.Time `orm:"column(FECHA_NACIMIENTO);type(date)" json:"FECHA_NACIMIENTO,omitempty"`
	NUEVO                   bool       `orm:"column(NUEVO);type(boolean)" json:"NUEVO"`
	ROL                     string     `orm:"column(ROL);type(text)" json:"ROL"`
	FECHA_INGRESO           time.Time  `orm:"column(FECHA_INGRESO);type(date)" json:"FECHA_INGRESO"`
	FECHA_RETIRO            *time.Time `orm:"column(FECHA_RETIRO);type(date);null" json:"FECHA_RETIRO,omitempty"`
	PASSWORD                string     `orm:"column(PASSWORD)" json:"PASSWORD"`
	HORARIO                 *string    `orm:"column(HORARIO);type(text);null" json:"HORARIO,omitempty"`
	PK_ID_RESTAURANTE       *int64     `orm:"column(PK_ID_RESTAURANTE);null" json:"PK_ID_RESTAURANTE,omitempty"`
}

func (t *Trabajador) TableName() string {
	return "TRABAJADOR"
}

func init() {
	orm.RegisterModel(new(Trabajador))
}

func (d Trabajador) MarshalJSON() ([]byte, error) {
	type Alias Trabajador
	return json.Marshal(&struct {
		FECHA_NACIMIENTO string `json:"FECHA"`
		FECHA_INGRESO    string `json:"CREATED_AT"`
		FECHA_RETIRO     string `json:"UPDATED_AT"`
		Alias
	}{
		FECHA_NACIMIENTO: d.FECHA_NACIMIENTO.Format("02-01-2006"),
		FECHA_INGRESO:    d.FECHA_INGRESO.Format("02-01-2006 15:04:05"),
		FECHA_RETIRO:     d.FECHA_RETIRO.Format("02-01-2006 15:04:05"),
		Alias:            (Alias)(d),
	})
}
