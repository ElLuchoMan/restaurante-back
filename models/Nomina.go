package models

import "github.com/beego/beego/v2/client/orm"

type Nomina struct {
	PK_ID_NOMINA      int64   `orm:"pk" json:"PK_ID_NOMINA"`
	FECHA             string  `json:"FECHA"`
	MONTO             int64   `json:"MONTO"`
	ESTADO            *string `orm:"null" json:"ESTADO"`
	PK_ID_RESTAURANTE *int64  `orm:"null" json:"PK_ID_RESTAURANTE"`
}

func (c *Nomina) TableName() string {
	return "NOMINA"
}

func init() {
	orm.RegisterModel(new(Nomina))
}
