package models

type Plato struct {
	PK_ID_PLATO       int64   `orm:"pk" json:"PK_ID_PLATO"`
	NOMBRE            string  `json:"NOMBRE"`
	CALORIAS          *int64  `orm:"null" json:"CALORIAS"`
	DESCRIPCION       *string `orm:"null" json:"DESCRIPCION"`
	PRECIO            int64   `json:"PRECIO"`
	PERSONALIZADO     bool    `json:"PERSONALIZADO"`
	PK_ID_ITEM_PEDIDO *int64  `orm:"null" json:"PK_ID_ITEM_PEDIDO"`
}
