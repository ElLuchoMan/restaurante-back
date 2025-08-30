package models

import (
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type RestauranteDia struct {
	PKIDRestaurante int64     `orm:"column(pk_id_restaurante);pk" json:"restauranteId"`
	Dia             DiaSemana `orm:"column(dia);type(text);pk" json:"dia"`
	HoraApertura    time.Time `orm:"column(hora_apertura);type(time)" json:"horaApertura"`
	HoraCierre      time.Time `orm:"column(hora_cierre);type(time)" json:"horaCierre"`
}

func (r *RestauranteDia) TableName() string {
	return "restaurante_dia"
}

func init() {
	orm.RegisterModel(new(RestauranteDia))
}
