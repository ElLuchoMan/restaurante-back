package models

import (
	"encoding/json"

	"github.com/beego/beego/v2/client/orm"
)

type Cliente struct {
	PK_DOCUMENTO_CLIENTE int64   `orm:"column(pk_documento_cliente);pk" json:"documentoCliente"`
	NOMBRE               string  `orm:"column(nombre);type(text)"        json:"nombre"`
	APELLIDO             string  `orm:"column(apellido);type(text)"      json:"apellido"`
	CORREO               string  `orm:"column(correo);type(text);unique"        json:"correo" valid:"email"`
	DIRECCION            string  `orm:"column(direccion);type(text)"     json:"direccion"`
	TELEFONO             string  `orm:"column(telefono);type(text);unique"      json:"telefono"`
	OBSERVACIONES        *string `orm:"column(observaciones);type(text);null" json:"observaciones"`
	PASSWORD             string  `orm:"column(password);type(text)"      json:"-"`
}

// UnmarshalJSON acepta `password` en la entrada (el campo se serializa como
// "-": nunca sale en una respuesta, pero sí se puede recibir).
func (c *Cliente) UnmarshalJSON(b []byte) error {
	type plain Cliente
	aux := struct {
		*plain
		Password *string `json:"password"`
	}{plain: (*plain)(c)}
	if err := json.Unmarshal(b, &aux); err != nil {
		return err
	}
	if aux.Password != nil {
		c.PASSWORD = *aux.Password
	}
	return nil
}

func (c *Cliente) TableName() string {
	return "cliente"
}

func init() {
	orm.RegisterModel(new(Cliente))
}
