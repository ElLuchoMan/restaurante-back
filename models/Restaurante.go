package models

import (
	"encoding/json"

	"github.com/beego/beego/v2/client/orm"
)

type Restaurante struct {
	PK_ID_RESTAURANTE  int64  `orm:"pk" json:"PK_ID_RESTAURANTE"`
	NOMBRE_RESTAURANTE string `json:"NOMBRE_RESTAURANTE"`
	HORA_APERTURA      string `json:"HORA_APERTURA"`
	DIAS_LABORALES     string `orm:"type(text)" json:"DIAS_LABORALES"`
}

func (c *Restaurante) TableName() string {
	return "RESTAURANTE"
}

func (r *Restaurante) SetDiasLaborales(dias []string) error {
	diasJSON, err := json.Marshal(dias)
	if err != nil {
		return err
	}
	r.DIAS_LABORALES = string(diasJSON)
	return nil
}

func (r *Restaurante) GetDiasLaborales() ([]string, error) {
	var dias []string
	err := json.Unmarshal([]byte(r.DIAS_LABORALES), &dias)
	return dias, err
}

func init() {
	orm.RegisterModel(new(Restaurante))
}
