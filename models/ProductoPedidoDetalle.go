package models

type ProductoPedidoDetalle struct {
	PK_ID_PRODUCTO_PEDIDO int64 `orm:"column(PK_ID_PRODUCTO_PEDIDO);pk" json:"productoPedidoId"`
	PK_ID_PRODUCTO        int64 `orm:"column(PK_ID_PRODUCTO)" json:"productoId"`
	CANTIDAD              int64 `orm:"column(CANTIDAD)" json:"cantidad"`
	PRECIO                int64 `orm:"column(PRECIO)" json:"precio"`
}

func (p *ProductoPedidoDetalle) TableName() string {
	return "PRODUCTO_PEDIDO_DETALLE"
}
