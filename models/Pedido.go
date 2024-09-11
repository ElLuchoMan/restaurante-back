package models

type Pedido struct {
	ID        int     `json:"id"`
	ClienteID int     `json:"cliente_id"`
	Fecha     string  `json:"fecha"`
	Estado    string  `json:"estado"`
	Delivery  bool    `json:"delivery"`
	Platos    []Plato `json:"platos"`
}
