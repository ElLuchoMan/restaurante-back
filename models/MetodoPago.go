package models

import "github.com/beego/beego/v2/client/orm"

type MetodoPago struct {
	PK_ID_METODO_PAGO int    `orm:"column(PK_ID_METODO_PAGO);pk" json:"PK_ID_METODO_PAGO"`
	TIPO              string `orm:"column(TIPO)" json:"TIPO"`
	DETALLE           string `orm:"column(DETALLE)" json:"DETALLE"`
	PK_ID_PAGO        *int   `orm:"column(PK_ID_PAGO);null" json:"PK_ID_PAGO"`
}

func (m *MetodoPago) TableName() string {
	return "METODO_PAGO"
}

func init() {
	orm.RegisterModel(new(MetodoPago))
}
