package models

import "github.com/beego/beego/v2/client/orm"

type Inventario struct {
	PK_ID_INVENTARIO  int64  `orm:"column(PK_ID_INVENTARIO);pk;auto" json:"PK_ID_INVENTARIO"`
	FECHA             string `orm:"column(FECHA);type(date)" json:"FECHA"`
	CANTIDAD          int64  `orm:"column(CANTIDAD)" json:"CANTIDAD"`
	UNIDAD            int64  `orm:"column(UNIDAD)" json:"UNIDAD"`
	UNIDAD_MINIMA     int64  `orm:"column(UNIDAD_MINIMA)" json:"UNIDAD_MINIMA"`
	PK_ID_INGREDIENTE *int64 `orm:"column(PK_ID_INGREDIENTE);null" json:"PK_ID_INGREDIENTE"`
}

func (i *Inventario) TableName() string {
	return "INVENTARIO"
}

func init() {
	orm.RegisterModel(new(Inventario))
}
