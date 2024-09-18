package models

import "github.com/beego/beego/v2/client/orm"

type Ingrediente struct {
	PK_ID_INGREDIENTE int64  `orm:"column(PK_ID_INGREDIENTE);pk" json:"PK_ID_INGREDIENTE"`
	NOMBRE            string `orm:"column(NOMBRE)" json:"NOMBRE"`
	TIPO              string `orm:"column(TIPO)" json:"TIPO"`
	PESO              int64  `orm:"column(PESO)" json:"PESO"`
	CALORIAS          int64  `orm:"column(CALORIAS)" json:"CALORIAS"`
	PK_ID_INVENTARIO  *int64 `orm:"column(PK_ID_INVENTARIO);null" json:"PK_ID_INVENTARIO"`
}

func (i *Ingrediente) TableName() string {
	return "INGREDIENTE"
}

func init() {
	orm.RegisterModel(new(Ingrediente))
}
