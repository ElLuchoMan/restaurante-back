package models

import "github.com/beego/beego/v2/client/orm"

type Plato struct {
	PK_ID_PLATO       int64   `orm:"column(PK_ID_PLATO);pk" json:"PK_ID_PLATO"`
	NOMBRE            string  `orm:"column(NOMBRE)" json:"NOMBRE"`
	CALORIAS          *int64  `orm:"column(CALORIAS);null" json:"CALORIAS"`
	DESCRIPCION       *string `orm:"column(DESCRIPCION);null" json:"DESCRIPCION"`
	PRECIO            int64   `orm:"column(PRECIO)" json:"PRECIO"`
	PERSONALIZADO     bool    `orm:"column(PERSONALIZADO)" json:"PERSONALIZADO"`
	PK_ID_ITEM_PEDIDO *int64  `orm:"column(PK_ID_ITEM_PEDIDO);null" json:"PK_ID_ITEM_PEDIDO"`
}

func (p *Plato) TableName() string {
	return "PLATO"
}

func init() {
	orm.RegisterModel(new(Plato))
}
