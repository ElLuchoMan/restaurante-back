package models

import "github.com/beego/beego/v2/client/orm"

type PedidoCliente struct {
	PK_ID_PEDIDO_CLIENTE int64  `orm:"pk" json:"PK_ID_PEDIDO_CLIENTE"`
	PK_ID_RESTAURANTE    *int64 `orm:"null" json:"PK_ID_RESTAURANTE"`
	PK_DOCUMENTO_CLIENTE *int64 `orm:"null" json:"PK_DOCUMENTO_CLIENTE"`
}

func (c *PedidoCliente) TableName() string {
	return "PEDIDO_CLIENTE"
}

func init() {
	orm.RegisterModel(new(PedidoCliente))
}
