package models

import "github.com/beego/beego/v2/client/orm"

type Inventario struct {
	PK_ID_INVENTARIO  int64  `orm:"pk" json:"PK_ID_INVENTARIO"`
	FECHA             string `json:"FECHA"`
	CANTIDAD          int64  `json:"CANTIDAD"`
	UNIDAD            int64  `json:"UNIDAD"`
	UNIDAD_MINIMA     int64  `json:"UNIDAD_MINIMA"`
	PK_ID_INGREDIENTE *int64 `orm:"null" json:"PK_ID_INGREDIENTE"`
}

func (c *Inventario) TableName() string {
	return "INVENTARIO"
}

func init() {
	orm.RegisterModel(new(Inventario))
}
