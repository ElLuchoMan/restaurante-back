package models

import "github.com/beego/beego/v2/client/orm"

type Producto struct {
	PK_ID_PRODUCTO  int64  `orm:"column(pk_id_producto);pk;auto" json:"productoId"`
	NOMBRE          string `orm:"column(nombre);type(text)" json:"nombre"`
	CALORIAS        *int64 `orm:"column(calorias);type(bigint)" json:"calorias"`
	DESCRIPCION     string `orm:"column(descripcion);type(text)" json:"descripcion"`
	PRECIO          int64  `orm:"column(precio);type(bigint)" json:"precio"`
	ESTADO_PRODUCTO string `orm:"column(estado_producto);type(text)" json:"estadoProducto"`
	IMAGEN          string `orm:"column(imagen);null" json:"imagen"`
	CANTIDAD        int    `orm:"column(cantidad);type(integer)" json:"cantidad"`
	CATEGORIA       string `orm:"column(categoria);type(text)" json:"categoria"`
	SUBCATEGORIA    string `orm:"column(subcategoria);type(text)" json:"subcategoria"`
}

func (p *Producto) TableName() string {
	return "producto"
}

func init() {
	orm.RegisterModel(new(Producto))
}
