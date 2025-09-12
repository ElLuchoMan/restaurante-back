package models

import "github.com/beego/beego/v2/client/orm"

type MetodoPago struct {
	PK_ID_METODO_PAGO int64  `orm:"column(pk_id_metodo_pago);pk;auto" json:"metodoPagoId"`
	TIPO              string `orm:"column(tipo);type(text)" json:"tipo"`
	DETALLE           string `orm:"column(detalle);type(text)" json:"detalle"`
}

func (m *MetodoPago) TableName() string {
	return "metodo_pago"
}

func init() {
	orm.RegisterModel(new(MetodoPago))
}
