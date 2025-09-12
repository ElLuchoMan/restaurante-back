package models

import "github.com/beego/beego/v2/client/orm"

type RestauranteDia struct {
	PK_ID_RESTAURANTE_DIA int64        `orm:"column(pk_id_restaurante_dia);pk;auto" json:"restauranteDiaId"`
	PKIDRestaurante       *Restaurante `orm:"column(pk_id_restaurante);rel(fk)" json:"restauranteId" swaggertype:"integer"`
	Dia                   DiaSemana    `orm:"column(dia);type(dia_semana)" json:"dia"`
}

func (r *RestauranteDia) TableName() string {
	return "restaurante_dia"
}

func (r *RestauranteDia) TableUnique() [][]string {
	return [][]string{{"PKIDRestaurante", "Dia"}}
}

func init() {
	orm.RegisterModel(new(RestauranteDia))
}
