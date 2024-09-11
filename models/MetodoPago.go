package models

import "github.com/beego/beego/v2/client/orm"

type MetodoPago struct {
	PK_ID_METODO_PAGO int64  `orm:"pk" json:"PK_ID_METODO_PAGO"`
	TIPO              string `json:"TIPO"`
	DETALLE           string `json:"DETALLE"`
	PK_ID_PAGO        *int64 `orm:"null" json:"PK_ID_PAGO"`
}

func (c *MetodoPago) TableName() string {
	return "METODO_PAGO"
}

func init() {
	orm.RegisterModel(new(MetodoPago))
}
