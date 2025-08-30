package models

import "github.com/beego/beego/v2/client/orm"

type PedidoCliente struct {
	PK_ID_PEDIDO_CLIENTE int64 `orm:"column(pk_id_pedido_cliente);pk;auto" json:"pedidoClienteId"`
	PK_DOCUMENTO_CLIENTE int64 `orm:"column(pk_documento_cliente);null" json:"documentoCliente"`
	PK_ID_PEDIDO         int   `orm:"column(pk_id_pedido);null" json:"pedidoId"`
}

func (pc *PedidoCliente) TableName() string {
	return "pedido_cliente"
}

func init() {
	orm.RegisterModel(new(PedidoCliente))
}
