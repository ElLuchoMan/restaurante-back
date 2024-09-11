package models

import "github.com/beego/beego/v2/client/orm"

type Pago struct {
	PK_ID_PAGO        int64  `orm:"pk" json:"PK_ID_PAGO"`
	FECHA             string `json:"FECHA"`
	HORA              string `json:"HORA"`
	MONTO             int64  `json:"MONTO"`
	ESTADO            string `json:"ESTADO"`
	PK_ID_METODO_PAGO *int64 `orm:"null" json:"PK_ID_METODO_PAGO"`
}

func (c *Pago) TableName() string {
	return "PAGO"
}

func init() {
	orm.RegisterModel(new(Pago))
}
