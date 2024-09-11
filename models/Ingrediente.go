package models

import "github.com/beego/beego/v2/client/orm"

type Ingrediente struct {
	PK_ID_INGREDIENTE int64  `orm:"pk" json:"PK_ID_INGREDIENTE"`
	NOMBRE            string `json:"NOMBRE"`
	TIPO              string `json:"TIPO"`
	PESO              int64  `json:"PESO"`
	CALORIAS          int64  `json:"CALORIAS"`
	PK_ID_INVENTARIO  *int64 `orm:"null" json:"PK_ID_INVENTARIO"`
}

func (c *Ingrediente) TableName() string {
	return "INGREDIENTE"
}

func init() {
	orm.RegisterModel(new(Ingrediente))
}
