package models

type Plato struct {
	ID            int    `json:"id"`
	Nombre        string `json:"nombre"`
	Calorias      int    `json:"calorias"`
	Descripcion   string `json:"descripcion"`
	Precio        int    `json:"precio"`
	Personalizado bool   `json:"personalizado"`
}
