package models

import "github.com/beego/beego/v2/client/orm"

type RestauranteDia struct {
	PKIDRestaurante int64     `orm:"column(pk_id_restaurante);pk" json:"restauranteId"`
	Dia             DiaSemana `orm:"column(dia);type(text)" json:"dia"`
}

func (r *RestauranteDia) TableName() string {
	return "restaurante_dia"
}

func init() {
	orm.RegisterModel(new(RestauranteDia))
}
