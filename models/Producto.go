package models

import "github.com/beego/beego/v2/client/orm"

type Producto struct {
	PK_ID_PRODUCTO  int64  `orm:"column(PK_ID_PRODUCTO);pk;auto" json:"PK_ID_PRODUCTO"`
	NOMBRE          string `orm:"column(NOMBRE);type(text)" json:"NOMBRE"`
	CALORIAS        *int64 `orm:"column(CALORIAS);type(bigint)" json:"CALORIAS"`
	DESCRIPCION     string `orm:"column(DESCRIPCION);type(text)" json:"DESCRIPCION"`
	PRECIO          int64  `orm:"column(PRECIO);type(bigint)" json:"PRECIO"`
	ESTADO_PRODUCTO string `orm:"column(ESTADO_PRODUCTO);type(text)" json:"ESTADO_PRODUCTO"`
	IMAGEN          string `orm:"column(IMAGEN);null" json:"IMAGEN"`
	CANTIDAD        int    `orm:"column(CANTIDAD);type(integer)" json:"CANTIDAD"`
}

func (p *Producto) TableName() string {
	return "PRODUCTO"
}

func init() {
	orm.RegisterModel(new(Producto))
}
