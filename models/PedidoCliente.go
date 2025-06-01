package models

import "github.com/beego/beego/v2/client/orm"

type PedidoCliente struct {
	PK_ID_PEDIDO_CLIENTE int64 `orm:"column(PK_ID_PEDIDO_CLIENTE);pk;auto" json:"pedidoClienteId"`
	PK_DOCUMENTO_CLIENTE int64 `orm:"column(PK_DOCUMENTO_CLIENTE);null" json:"documentoCliente"`
	PK_ID_PEDIDO         int   `orm:"column(PK_ID_PEDIDO);null" json:"pedidoId"`
}

func (pc *PedidoCliente) TableName() string {
	return "PEDIDO_CLIENTE"
}

func init() {
	orm.RegisterModel(new(PedidoCliente))
}
