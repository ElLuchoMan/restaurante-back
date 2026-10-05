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
	PASSWORD                string              `orm:"column(password)" json:"-"`
	HORARIOS                []HorarioTrabajador `orm:"-" json:"horarios,omitempty"`
	PK_ID_RESTAURANTE       *Restaurante        `orm:"column(pk_id_restaurante);rel(fk);null" json:"restauranteId,omitempty"`
}

func (t *Trabajador) TableName() string {
	return "trabajador"
}

func init() {
	orm.RegisterModel(new(Trabajador))
}

// MarshalJSON produce exactamente la forma documentada en TrabajadorResponse:
// fechas DD-MM-YYYY, sin password, `restauranteId` como objeto mínimo
// {"restauranteId": n} y `horarios` solo cuando el controlador los cargó
// (lista vacía = []).
// UnmarshalJSON acepta `password` en la entrada (el campo nunca se serializa).
func (d *Trabajador) UnmarshalJSON(b []byte) error {
	type plain Trabajador
	aux := struct {
		*plain
		Password *string `json:"password"`
	}{plain: (*plain)(d)}
	if err := json.Unmarshal(b, &aux); err != nil {
		return err
	}
	if aux.Password != nil {
		d.PASSWORD = *aux.Password
	}
	return nil
}

func (d Trabajador) MarshalJSON() ([]byte, error) {

	var fechaNacimientoStr *string
	if d.FECHA_NACIMIENTO != nil {
		str := FormatDateUTC(*d.FECHA_NACIMIENTO)
		fechaNacimientoStr = &str
	}

	fechaIngresoStr := FormatDateUTC(d.FECHA_INGRESO)

	var fechaRetiroStr *string
	if d.FECHA_RETIRO != nil {
		str := FormatDateUTC(*d.FECHA_RETIRO)
		fechaRetiroStr = &str
	}

	var horarios *[]HorarioTrabajador
	if d.HORARIOS != nil {
		horarios = &d.HORARIOS
	}

	var restaurante *RestauranteRef
	if d.PK_ID_RESTAURANTE != nil {
		restaurante = &RestauranteRef{RestauranteId: d.PK_ID_RESTAURANTE.PK_ID_RESTAURANTE}
	}

	return json.Marshal(&struct {
		PK_DOCUMENTO_TRABAJADOR int64                `json:"documentoTrabajador"`
		NOMBRE                  string               `json:"nombre"`
		APELLIDO                string               `json:"apellido"`
		SUELDO                  int64                `json:"sueldo"`
		TELEFONO                *string              `json:"telefono,omitempty"`
		FECHA_NACIMIENTO        *string              `json:"fechaNacimiento,omitempty"`
		NUEVO                   bool                 `json:"nuevo"`
		ROL                     RolTrabajador        `json:"rol"`
		FECHA_INGRESO           string               `json:"fechaIngreso"`
		FECHA_RETIRO            *string              `json:"fechaRetiro,omitempty"`
		HORARIOS                *[]HorarioTrabajador `json:"horarios,omitempty"`
		PK_ID_RESTAURANTE       *RestauranteRef      `json:"restauranteId,omitempty"`
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
		HORARIOS:                horarios,
		PK_ID_RESTAURANTE:       restaurante,
	})
}
