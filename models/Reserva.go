package models

import "github.com/beego/beego/v2/client/orm"

type Reserva struct {
	PK_ID_RESERVA     int    `orm:"column(PK_ID_RESERVA);pk;auto" json:"PK_ID_RESERVA"`
	FECHA             string `orm:"column(FECHA);type(date)" json:"FECHA"`
	HORA              string `orm:"column(HORA);type(time)" json:"HORA"`
	PERSONAS          int    `orm:"column(PERSONAS)" json:"PERSONAS"`
	PK_ID_RESTAURANTE *int   `orm:"column(PK_ID_RESTAURANTE);null" json:"PK_ID_RESTAURANTE"`
}

func (r *Reserva) TableName() string {
	return "RESERVA"
}

func init() {
	orm.RegisterModel(new(Reserva))
}
