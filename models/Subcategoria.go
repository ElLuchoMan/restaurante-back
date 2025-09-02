package models

import (
	"github.com/beego/beego/v2/client/orm"
)

type Subcategoria struct {
	PK_ID_SUBCATEGORIA int64      `orm:"column(pk_id_subcategoria);pk;auto" json:"subcategoriaId"`
	PK_ID_CATEGORIA    *Categoria `orm:"column(pk_id_categoria);rel(fk)" json:"categoriaId"`
	NOMBRE             string     `orm:"column(nombre);type(text)" json:"nombre"`
}

func (s *Subcategoria) TableName() string {
	return "subcategoria"
}

func init() {
	orm.RegisterModel(new(Subcategoria))
}
