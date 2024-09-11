package models

import "github.com/beego/beego/v2/client/orm"

type Reserva struct {
	PK_ID_RESERVA     int64  `orm:"pk" json:"PK_ID_RESERVA"`
	FECHA             string `json:"FECHA"`
	HORA              string `json:"HORA"`
	PERSONAS          int    `json:"PERSONAS"`
	PK_ID_RESTAURANTE *int64 `orm:"null" json:"PK_ID_RESTAURANTE"`
}

func (c *Reserva) TableName() string {
	return "RESERVA"
}

func init() {
	orm.RegisterModel(new(Reserva))
}
