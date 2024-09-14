package models

import (
	"encoding/json"

	"github.com/beego/beego/v2/client/orm"
)

type Restaurante struct {
	PK_ID_RESTAURANTE  int    `orm:"column(PK_ID_RESTAURANTE);pk"`
	NOMBRE_RESTAURANTE string `orm:"column(NOMBRE_RESTAURANTE)"`
	HORA_APERTURA      string `orm:"column(HORA_APERTURA);type(time)"`
	DIAS_LABORALES     string `orm:"column(DIAS_LABORALES);type(text)"`
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
