package models

import "github.com/beego/beego/v2/client/orm"

type Producto struct {
	PK_ID_PRODUCTO  int64  `orm:"column(PK_ID_PRODUCTO);pk;auto" json:"productoId"`
	NOMBRE          string `orm:"column(NOMBRE);type(text)" json:"nombre"`
	CALORIAS        *int64 `orm:"column(CALORIAS);type(bigint)" json:"calorias"`
	DESCRIPCION     string `orm:"column(DESCRIPCION);type(text)" json:"descripcion"`
	PRECIO          int64  `orm:"column(PRECIO);type(bigint)" json:"precio"`
	ESTADO_PRODUCTO string `orm:"column(ESTADO_PRODUCTO);type(text)" json:"estadoProducto"`
	IMAGEN          string `orm:"column(IMAGEN);null" json:"imagen"`
	CANTIDAD        int    `orm:"column(CANTIDAD);type(integer)" json:"cantidad"`
	CATEGORIA       string `orm:"column(CATEGORIA);type(text)" json:"categoria"`
	SUBCATEGORIA    string `orm:"column(SUBCATEGORIA);type(text)" json:"subcategoria"`
}

func (p *Producto) TableName() string {
	return "PRODUCTO"
}

func init() {
	orm.RegisterModel(new(Producto))
}
