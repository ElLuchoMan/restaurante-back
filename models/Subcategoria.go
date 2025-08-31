package models

import (
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type Subcategoria struct {
	PK_ID_SUBCATEGORIA int64     `orm:"column(pk_id_subcategoria);pk;auto" json:"subcategoriaId"`
	PK_ID_CATEGORIA    int64     `orm:"column(pk_id_categoria)" json:"categoriaId"`
	NOMBRE             string    `orm:"column(nombre);type(text)" json:"nombre"`
	DESCRIPCION        string    `orm:"column(descripcion);type(text);null" json:"descripcion"`
	CREATED_AT         time.Time `orm:"column(created_at);type(timestamp);auto_now_add" json:"createdAt"`
	UPDATED_AT         time.Time `orm:"column(updated_at);type(timestamp);auto_now" json:"updatedAt"`
}

func (s *Subcategoria) TableName() string {
	return "subcategoria"
}

func init() {
	orm.RegisterModel(new(Subcategoria))
}
