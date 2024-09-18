package models

import "github.com/beego/beego/v2/client/orm"

type Pago struct {
	PK_ID_PAGO        int    `orm:"column(PK_ID_PAGO);pk" json:"PK_ID_PAGO"`
	FECHA             string `orm:"column(FECHA);type(date)" json:"FECHA"`
	HORA              string `orm:"column(HORA);type(time)" json:"HORA"`
	MONTO             int64  `orm:"column(MONTO)" json:"MONTO"`
	ESTADO            string `orm:"column(ESTADO)" json:"ESTADO"`
	PK_ID_METODO_PAGO *int   `orm:"column(PK_ID_METODO_PAGO);null" json:"PK_ID_METODO_PAGO"`
}

func (p *Pago) TableName() string {
	return "PAGO"
}

func init() {
	orm.RegisterModel(new(Pago))
}
