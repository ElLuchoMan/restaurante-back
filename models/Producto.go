package models

import (
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type Producto struct {
	PK_ID_PRODUCTO     int64          `orm:"column(pk_id_producto);pk;auto" json:"productoId"`
	NOMBRE             string         `orm:"column(nombre);type(text)" json:"nombre"`
	CALORIAS           *int64         `orm:"column(calorias);type(bigint)" json:"calorias"`
	DESCRIPCION        string         `orm:"column(descripcion);type(text)" json:"descripcion"`
	PRECIO             int64          `orm:"column(precio);type(bigint)" json:"precio"`
	ESTADO_PRODUCTO    EstadoProducto `orm:"column(estado_producto);type(text)" json:"estadoProducto"`
	IMAGEN             []byte         `orm:"column(imagen);type(text);null" json:"imagen"`
	CANTIDAD           int            `orm:"column(cantidad);type(integer)" json:"cantidad"`
	PK_ID_SUBCATEGORIA int64          `orm:"column(pk_id_subcategoria)" json:"subcategoriaId"`
	CREATED_AT         time.Time      `orm:"column(created_at);type(timestamp);auto_now_add" json:"createdAt"`
	UPDATED_AT         time.Time      `orm:"column(updated_at);type(timestamp);auto_now" json:"updatedAt"`
	CREATED_BY         *string        `orm:"column(created_by);type(text)" json:"createdBy,omitempty"`
	UPDATED_BY         *string        `orm:"column(updated_by);type(text)" json:"updatedBy,omitempty"`
}

func (p *Producto) TableName() string {
	return "producto"
}

func init() {
	orm.RegisterModel(new(Producto))
}
