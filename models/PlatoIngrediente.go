package models

import "github.com/beego/beego/v2/client/orm"

type PlatoIngrediente struct {
	PK_ID_PLATO_INGREDIENTE int64  `orm:"column(PK_ID_PLATO_INGREDIENTE);pk;auto" json:"PK_ID_PLATO_INGREDIENTE"`
	CANTIDAD                int64  `orm:"column(CANTIDAD)" json:"CANTIDAD"`
	PK_ID_INGREDIENTE       *int64 `orm:"column(PK_ID_INGREDIENTE);null" json:"PK_ID_INGREDIENTE"`
	PK_ID_PLATO             *int64 `orm:"column(PK_ID_PLATO);null" json:"PK_ID_PLATO"`
}

func (pi *PlatoIngrediente) TableName() string {
	return "PLATO_INGREDIENTE"
}

func init() {
	orm.RegisterModel(new(PlatoIngrediente))
}
