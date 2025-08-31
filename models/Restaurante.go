package models

import (
	"encoding/json"

	"github.com/beego/beego/v2/client/orm"
)

var jsonMarshal = json.Marshal

type Restaurante struct {
	PK_ID_RESTAURANTE    int    `orm:"column(pk_id_restaurante);pk" json:"restauranteId"`
	NOMBRE_RESTAURANTE   string `orm:"column(nombre_restaurante)" json:"nombreRestaurante"`
	HORA_APERTURA        string `orm:"column(hora_apertura);type(time)" json:"horaApertura"`
	DIAS_LABORALES       string `orm:"column(dias_laborales)" json:"diasLaborales"`
	PK_ID_CAMBIO_HORARIO *int   `orm:"column(pk_id_cambio_horario);null" json:"-"`
	PK_ID_RESERVA        *int   `orm:"column(pk_id_reserva);null" json:"-"`
}

func (t *Restaurante) TableName() string {
	return "restaurante"
}

// Método para establecer los días laborales como una cadena JSON
func (r *Restaurante) SetDiasLaborales(dias []string) error {
	diasJSON, err := jsonMarshal(dias)
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
