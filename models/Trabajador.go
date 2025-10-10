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
	FECHA_NACIMIENTO        *time.Time          `orm:"column(fecha_nacimiento);type(date);null" json:"fechaNacimiento,omitempty"`
	NUEVO                   bool                `orm:"column(nuevo);type(boolean)" json:"nuevo"`
	ROL                     RolTrabajador       `orm:"column(rol);type(text)" json:"rol"`
	FECHA_INGRESO           time.Time           `orm:"column(fecha_ingreso);type(date)" json:"fechaIngreso"`
	FECHA_RETIRO            *time.Time          `orm:"column(fecha_retiro);type(date);null" json:"fechaRetiro,omitempty"`
	PASSWORD                string              `orm:"column(password)" json:"password"`
	HORARIOS                []HorarioTrabajador `orm:"-" json:"horarios,omitempty"`
	PK_ID_RESTAURANTE       *Restaurante        `orm:"column(pk_id_restaurante);rel(fk);null" json:"restauranteId,omitempty" swaggertype:"integer"`
}

func (t *Trabajador) TableName() string {
	return "trabajador"
}

func init() {
	orm.RegisterModel(new(Trabajador))
}

func (d Trabajador) MarshalJSON() ([]byte, error) {
	// Para campos de tipo date: extraer SOLO año/mes/día sin importar timezone
	var fechaNacimientoStr *string
	if d.FECHA_NACIMIENTO != nil {
		str := time.Date(d.FECHA_NACIMIENTO.Year(), d.FECHA_NACIMIENTO.Month(), d.FECHA_NACIMIENTO.Day(), 0, 0, 0, 0, time.UTC).Format("02-01-2006")
		fechaNacimientoStr = &str
	}

	fechaIngresoStr := time.Date(d.FECHA_INGRESO.Year(), d.FECHA_INGRESO.Month(), d.FECHA_INGRESO.Day(), 0, 0, 0, 0, time.UTC).Format("02-01-2006")

	var fechaRetiroStr *string
	if d.FECHA_RETIRO != nil {
		str := time.Date(d.FECHA_RETIRO.Year(), d.FECHA_RETIRO.Month(), d.FECHA_RETIRO.Day(), 0, 0, 0, 0, time.UTC).Format("02-01-2006")
		fechaRetiroStr = &str
	}

	return json.Marshal(&struct {
		PK_DOCUMENTO_TRABAJADOR int64               `json:"documentoTrabajador"`
		NOMBRE                  string              `json:"nombre"`
		APELLIDO                string              `json:"apellido"`
		SUELDO                  int64               `json:"sueldo"`
		TELEFONO                *string             `json:"telefono,omitempty"`
		FECHA_NACIMIENTO        *string             `json:"fechaNacimiento,omitempty"`
		NUEVO                   bool                `json:"nuevo"`
		ROL                     RolTrabajador       `json:"rol"`
		FECHA_INGRESO           string              `json:"fechaIngreso"`
		FECHA_RETIRO            *string             `json:"fechaRetiro,omitempty"`
		PASSWORD                string              `json:"password"`
		HORARIOS                []HorarioTrabajador `json:"horarios,omitempty"`
		PK_ID_RESTAURANTE       *Restaurante        `json:"restauranteId,omitempty"`
	}{
		PK_DOCUMENTO_TRABAJADOR: d.PK_DOCUMENTO_TRABAJADOR,
		NOMBRE:                  d.NOMBRE,
		APELLIDO:                d.APELLIDO,
		SUELDO:                  d.SUELDO,
		TELEFONO:                d.TELEFONO,
		FECHA_NACIMIENTO:        fechaNacimientoStr,
		NUEVO:                   d.NUEVO,
		ROL:                     d.ROL,
		FECHA_INGRESO:           fechaIngresoStr,
		FECHA_RETIRO:            fechaRetiroStr,
		PASSWORD:                d.PASSWORD,
		HORARIOS:                d.HORARIOS,
		PK_ID_RESTAURANTE:       d.PK_ID_RESTAURANTE,
	})
}
