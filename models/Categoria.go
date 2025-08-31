package models

import (
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type Categoria struct {
	PK_ID_CATEGORIA int64     `orm:"column(pk_id_categoria);pk;auto" json:"categoriaId"`
	NOMBRE          string    `orm:"column(nombre);type(text)" json:"nombre"`
	DESCRIPCION     string    `orm:"column(descripcion);type(text);null" json:"descripcion"`
	CREATED_AT      time.Time `orm:"column(created_at);type(timestamp);auto_now_add" json:"createdAt"`
	UPDATED_AT      time.Time `orm:"column(updated_at);type(timestamp);auto_now" json:"updatedAt"`
}

func (c *Categoria) TableName() string {
	return "categoria"
}

func init() {
	orm.RegisterModel(new(Categoria))
}
