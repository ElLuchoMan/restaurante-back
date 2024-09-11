package models

import "github.com/beego/beego/v2/client/orm"

type PlatoIngrediente struct {
	PK_ID_PLATO_INGREDIENTE int64  `orm:"pk" json:"PK_ID_PLATO_INGREDIENTE"`
	CANTIDAD                int64  `json:"CANTIDAD"`
	PK_ID_INGREDIENTE       *int64 `orm:"null" json:"PK_ID_INGREDIENTE"`
	PK_ID_PLATO             *int64 `orm:"null" json:"PK_ID_PLATO"`
}

func (c *PlatoIngrediente) TableName() string {
	return "PLATO_INGREDIENTE"
}

func init() {
	orm.RegisterModel(new(PlatoIngrediente))
}
