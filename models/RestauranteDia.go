package models

import "github.com/beego/beego/v2/client/orm"

type RestauranteDia struct {
	PK_ID_RESTAURANTE_DIA int64     `orm:"column(pk_id_restaurante_dia);pk;auto" json:"restauranteDiaId"`
	PKIDRestaurante       int64     `orm:"column(pk_id_restaurante);unique(restaurante_dia)" json:"restauranteId"`
	Dia                   DiaSemana `orm:"column(dia);type(text);unique(restaurante_dia)" json:"dia"`
}

func (r *RestauranteDia) TableName() string {
	return "restaurante_dia"
}

func init() {
	orm.RegisterModel(new(RestauranteDia))
}
