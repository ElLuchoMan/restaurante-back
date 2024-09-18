package models

import "github.com/beego/beego/v2/client/orm"

type Nomina struct {
	PK_ID_NOMINA      int64  `orm:"column(PK_ID_NOMINA);pk;auto" json:"PK_ID_NOMINA"`
	FECHA             string `orm:"column(FECHA);type(date)" json:"FECHA"`
	MONTO             int64  `orm:"column(MONTO)" json:"MONTO"`
	ESTADO            string `orm:"column(ESTADO)" json:"ESTADO"`
	PK_ID_RESTAURANTE *int   `orm:"column(PK_ID_RESTAURANTE);null" json:"PK_ID_RESTAURANTE"`
}

func (n *Nomina) TableName() string {
	return "NOMINA"
}

func init() {
	orm.RegisterModel(new(Nomina))
}
