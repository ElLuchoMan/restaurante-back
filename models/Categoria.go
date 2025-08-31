package models

import (
	"github.com/beego/beego/v2/client/orm"
)

type Categoria struct {
	PK_ID_CATEGORIA int64  `orm:"column(pk_id_categoria);pk;auto" json:"categoriaId"`
	NOMBRE          string `orm:"column(nombre);type(text)" json:"nombre"`
}

func (c *Categoria) TableName() string {
	return "categoria"
}

func init() {
	orm.RegisterModel(new(Categoria))
}
