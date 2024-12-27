package models

import "github.com/beego/beego/v2/client/orm"

type Plato struct {
	PK_ID_PRODUCTO  int64  `orm:"column(PK_ID_PRODUCTO);pk;auto" json:"PK_ID_PRODUCTO"`
	NOMBRE          string `orm:"column(NOMBRE)" json:"NOMBRE"`
	CALORIAS        *int64 `orm:"column(CALORIAS);null" json:"CALORIAS"`
	DESCRIPCION     string `orm:"column(DESCRIPCION);null" json:"DESCRIPCION"`
	PRECIO          int64  `orm:"column(PRECIO)" json:"PRECIO"`
	ESTADO_PRODUCTO string `orm:"column(ESTADO_PRODUCTO)" json:"ESTADO_PRODUCTO"`
	IMAGEN          string `orm:"column(IMAGEN);null" json:"IMAGEN"`
	CANTIDAD        bool   `orm:"column(CANTIDAD);default(true)" json:"CANTIDAD"`
}

func (p *Plato) TableName() string {
	return "PRODUCTO"
}

func init() {
	orm.RegisterModel(new(Plato))
}
