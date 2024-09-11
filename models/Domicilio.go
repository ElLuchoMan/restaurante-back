package models

import "github.com/beego/beego/v2/client/orm"

type Domicilio struct {
	PK_ID_DOMICILIO int64  `orm:"pk" json:"PK_ID_DOMICILIO"`
	DIRECCION       string `json:"DIRECCION"`
	TELEFONO        int64  `json:"TELEFONO"`
	ESTADO_PAGO     string `json:"ESTADO_PAGO"`
	ENTREGADO       bool   `json:"ENTREGADO"`
	FECHA           string `json:"FECHA"`
}

func (c *Domicilio) TableName() string {
	return "DOMICILIO"
}

func init() {
	orm.RegisterModel(new(Domicilio))
}
