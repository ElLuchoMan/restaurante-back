package models

import (
	"encoding/json"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type Trabajador struct {
	PK_DOCUMENTO_TRABAJADOR int64               `orm:"column(PK_DOCUMENTO_TRABAJADOR);pk" json:"documentoTrabajador"`
	NOMBRE                  string              `orm:"column(NOMBRE);type(text)" json:"nombre"`
	APELLIDO                string              `orm:"column(APELLIDO);type(text)" json:"apellido"`
	SUELDO                  int64               `orm:"column(SUELDO)" json:"sueldo"`
	TELEFONO                *string             `orm:"column(TELEFONO);type(text);null" json:"telefono,omitempty"`
	FECHA_NACIMIENTO        *time.Time          `orm:"column(FECHA_NACIMIENTO);type(date)" json:"fechaNacimiento,omitempty"`
	NUEVO                   bool                `orm:"column(NUEVO);type(boolean)" json:"nuevo"`
	ROL                     string              `orm:"column(ROL);type(text)" json:"rol"`
	FECHA_INGRESO           time.Time           `orm:"column(FECHA_INGRESO);type(date)" json:"fechaIngreso"`
	FECHA_RETIRO            *time.Time          `orm:"column(FECHA_RETIRO);type(date);null" json:"fechaRetiro,omitempty"`
	PASSWORD                string              `orm:"column(PASSWORD)" json:"password"`
	HORARIOS                []HorarioTrabajador `orm:"-" json:"horarios,omitempty"`
	PK_ID_RESTAURANTE       *int64              `orm:"column(PK_ID_RESTAURANTE);null" json:"restauranteId,omitempty"`
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
		FECHA_NACIMIENTO *string `json:"fechaNacimiento,omitempty"`
		FECHA_INGRESO    string  `json:"fechaIngreso"`
		FECHA_RETIRO     *string `json:"fechaRetiro,omitempty"`
		Alias
	}{
		FECHA_NACIMIENTO: func() *string {
			if d.FECHA_NACIMIENTO != nil {
				str := d.FECHA_NACIMIENTO.Format("02-01-2006")
				return &str
			}
			return nil
		}(),
		FECHA_INGRESO: d.FECHA_INGRESO.Format("02-01-2006 15:04:05"),
		FECHA_RETIRO: func() *string {
			if d.FECHA_RETIRO != nil {
				str := d.FECHA_RETIRO.Format("02-01-2006 15:04:05")
				return &str
			}
			return nil
		}(),
		Alias: (Alias)(d),
	})
}
