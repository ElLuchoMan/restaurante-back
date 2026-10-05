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

// RestauranteDiaView es la forma que devuelven los endpoints /restaurante_dia:
// una fila de restaurante_dia unida con los datos del restaurante.
type RestauranteDiaView struct {
	RestauranteDiaID  int64     `json:"restauranteDiaId" orm:"column(restaurante_dia_id)" example:"1"`
	RestauranteID     int64     `json:"restauranteId" orm:"column(restaurante_id)" example:"1"`
	NombreRestaurante string    `json:"nombreRestaurante" orm:"column(nombre_restaurante)" example:"El Fogón de María"`
	HoraApertura      string    `json:"horaApertura" orm:"column(hora_apertura)" example:"08:00:00"` // HH:MM:SS
	Dia               DiaSemana `json:"dia" orm:"column(dia)" enums:"Lunes,Martes,Miércoles,Jueves,Viernes,Sábado,Domingo" example:"Lunes"`
}

// DiasSemana lista los valores válidos del enum dia_semana.
var DiasSemana = []DiaSemana{DiaLunes, DiaMartes, DiaMiercoles, DiaJueves, DiaViernes, DiaSabado, DiaDomingo}
