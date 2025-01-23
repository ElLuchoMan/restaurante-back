package models

import "github.com/beego/beego/v2/client/orm"

type MetodoPago struct {
	PK_ID_METODO_PAGO int    `orm:"column(PK_ID_METODO_PAGO);pk;auto" json:"metodoPagoId"`
	TIPO              string `orm:"column(TIPO);size(50)" json:"tipo"`
	DETALLE           string `orm:"column(DETALLE);type(text);null" json:"detalle"`
	PK_ID_PAGO        *int   `orm:"column(PK_ID_PAGO);null" json:"-"`
}

func (m *MetodoPago) TableName() string {
	return "METODO_PAGO"
}

func init() {
	orm.RegisterModel(new(MetodoPago))
}
