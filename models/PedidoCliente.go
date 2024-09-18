package models

import "github.com/beego/beego/v2/client/orm"

type PedidoCliente struct {
	PK_ID_PEDIDO_CLIENTE int64  `orm:"column(PK_ID_PEDIDO_CLIENTE);pk;auto" json:"PK_ID_PEDIDO_CLIENTE"`
	PK_DOCUMENTO_CLIENTE *int64 `orm:"column(PK_DOCUMENTO_CLIENTE);null" json:"PK_DOCUMENTO_CLIENTE"`
	PK_ID_RESTAURANTE    *int   `orm:"column(PK_ID_RESTAURANTE);null" json:"PK_ID_RESTAURANTE"`
}

func (pc *PedidoCliente) TableName() string {
	return "PEDIDO_CLIENTE"
}

func init() {
	orm.RegisterModel(new(PedidoCliente))
}
