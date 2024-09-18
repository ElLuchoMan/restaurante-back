package models

import (
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type Reserva struct {
	PK_ID_RESERVA     int       `orm:"column(PK_ID_RESERVA);pk;auto" json:"pk_id_reserva"`
	FECHA             time.Time `orm:"column(FECHA);type(date)" json:"fecha"`
	HORA              string    `orm:"column(HORA);type(time)" json:"hora"`
	PERSONAS          int       `orm:"column(PERSONAS)" json:"personas"`
	ESTADO            string    `orm:"column(ESTADO);type(text)" json:"estado"`
	PK_ID_RESTAURANTE *int      `orm:"column(PK_ID_RESTAURANTE);null" json:"pk_id_restaurante"`
}

func (r *Reserva) TableName() string {
	return "RESERVA"
}

func init() {
	orm.RegisterModel(new(Reserva))
}
