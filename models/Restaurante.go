package models

import (
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type Restaurante struct {
	PK_ID_RESTAURANTE    int64     `orm:"column(pk_id_restaurante);pk;auto" json:"restauranteId"`
	NOMBRE_RESTAURANTE   string    `orm:"column(nombre_restaurante);type(text)" json:"nombreRestaurante"`
	HORA_APERTURA        time.Time `orm:"column(hora_apertura);type(time)" json:"horaApertura"`
	PK_ID_CAMBIO_HORARIO *int64    `orm:"column(pk_id_cambio_horario);null" json:"cambioHorarioId,omitempty"`
}

func (t *Restaurante) TableName() string {
	return "restaurante"
}

func init() {
	orm.RegisterModel(new(Restaurante))
}
