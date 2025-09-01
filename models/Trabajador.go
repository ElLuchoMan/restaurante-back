package models

import (
	"encoding/json"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type Trabajador struct {
	PK_DOCUMENTO_TRABAJADOR int64               `orm:"column(pk_documento_trabajador);pk" json:"documentoTrabajador"`
	NOMBRE                  string              `orm:"column(nombre);type(text)" json:"nombre"`
	APELLIDO                string              `orm:"column(apellido);type(text)" json:"apellido"`
	SUELDO                  int64               `orm:"column(sueldo)" json:"sueldo"`
	TELEFONO                *string             `orm:"column(telefono);type(text);unique;null" json:"telefono,omitempty"`
	FECHA_NACIMIENTO        *time.Time          `orm:"column(fecha_nacimiento);type(date)" json:"fechaNacimiento,omitempty"`
	NUEVO                   bool                `orm:"column(nuevo);type(boolean)" json:"nuevo"`
	ROL                     string              `orm:"column(rol);type(text)" json:"rol"`
	FECHA_INGRESO           time.Time           `orm:"column(fecha_ingreso);type(date)" json:"fechaIngreso"`
	FECHA_RETIRO            *time.Time          `orm:"column(fecha_retiro);type(date);null" json:"fechaRetiro,omitempty"`
	PASSWORD                string              `orm:"column(password)" json:"password"`
	HORARIOS                []HorarioTrabajador `orm:"-" json:"horarios,omitempty"`
	PK_ID_RESTAURANTE       *int64              `orm:"column(pk_id_restaurante);null" json:"restauranteId,omitempty"`
}

func (t *Trabajador) TableName() string {
	return "trabajador"
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
