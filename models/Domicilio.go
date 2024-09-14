package models

import (
	"github.com/beego/beego/v2/client/orm"
)

type Domicilio struct {
	PK_ID_DOMICILIO int64  `orm:"column(PK_ID_DOMICILIO);pk;auto"` // Autoincrementable
	DIRECCION       string `orm:"column(DIRECCION)"`               // Dirección del domicilio
	TELEFONO        int64  `orm:"column(TELEFONO)"`                // Teléfono
	ESTADO_PAGO     string `orm:"column(ESTADO_PAGO)"`             // Estado del pago (PAGADO, PENDIENTE, NO PAGO)
	ENTREGADO       bool   `orm:"column(ENTREGADO)"`               // Si el domicilio fue entregado o no
	FECHA           string `orm:"column(FECHA);type(date)"`        // Fecha del domicilio
}

func (d *Domicilio) TableName() string {
	return "DOMICILIO"
}

func init() {
	orm.RegisterModel(new(Domicilio))
}
