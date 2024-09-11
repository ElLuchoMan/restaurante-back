package models

import (
	"github.com/beego/beego/v2/client/orm"
)

type Cliente struct {
	PK_DOCUMENTO_CLIENTE int     `orm:"column(PK_DOCUMENTO_CLIENTE);pk"` // Clave primaria
	NOMBRE               string  `orm:"column(NOMBRE)"`                  // Nombre
	APELLIDO             string  `orm:"column(APELLIDO)"`                // Apellido
	DIRECCION            string  `orm:"column(DIRECCION)"`               // Dirección
	TELEFONO             string  `orm:"column(TELEFONO)"`                // Teléfono
	OBSERVACIONES        *string `orm:"column(OBSERVACIONES);null"`      // Observaciones
	PASSWORD             string  `orm:"column(PASSWORD)"`                // Contraseña
}

func (c *Cliente) TableName() string {
	return "CLIENTE"
}

func init() {
	orm.RegisterModel(new(Cliente))
}
