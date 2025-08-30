package models

import "github.com/beego/beego/v2/client/orm"

type ReservaContacto struct {
	PKIDReservaContacto int64   `orm:"column(pk_id_reserva_contacto);pk;auto" json:"reservaContactoId"`
	PKIDReserva         int64   `orm:"column(pk_id_reserva)" json:"reservaId"`
	Nombre              string  `orm:"column(nombre);type(text)" json:"nombre"`
	Telefono            *string `orm:"column(telefono);type(text);null" json:"telefono,omitempty"`
	Email               *string `orm:"column(email);type(text);null" json:"email,omitempty"`
}

func (r *ReservaContacto) TableName() string {
	return "reserva_contacto"
}

func init() {
	orm.RegisterModel(new(ReservaContacto))
}
