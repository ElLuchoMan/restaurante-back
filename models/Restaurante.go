package models

import (
	"encoding/json"

	"github.com/beego/beego/v2/client/orm"
)

type Restaurante struct {
	PK_ID_RESTAURANTE  int    `orm:"column(PK_ID_RESTAURANTE);pk" json:"pk_id_restaurante"`
	NOMBRE_RESTAURANTE string `orm:"column(NOMBRE_RESTAURANTE)" json:"nombre_restaurante"`
	HORA_APERTURA      string `orm:"column(HORA_APERTURA);type(time)" json:"hora_apertura"`
	DIAS_LABORALES     string `orm:"column(DIAS_LABORALES)" json:"dias_laborales"`
}

func (t *Restaurante) TableName() string {
	return "RESTAURANTE"
}

// Método para establecer los días laborales como una cadena JSON
func (r *Restaurante) SetDiasLaborales(dias []string) error {
	diasJSON, err := json.Marshal(dias)
	if err != nil {
		return err
	}
	r.DIAS_LABORALES = string(diasJSON)
	return nil
}

// Método para obtener los días laborales a partir de la cadena JSON
func (r *Restaurante) GetDiasLaborales() ([]string, error) {
	var dias []string
	err := json.Unmarshal([]byte(r.DIAS_LABORALES), &dias)
	return dias, err
}

func init() {
	orm.RegisterModel(new(Restaurante))
}
