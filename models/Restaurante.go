package models

import (
	"encoding/json"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type Restaurante struct {
	PK_ID_RESTAURANTE    int64           `orm:"column(pk_id_restaurante);pk;auto" json:"restauranteId"`
	NOMBRE_RESTAURANTE   string          `orm:"column(nombre_restaurante);type(text)" json:"nombreRestaurante"`
	HORA_APERTURA        time.Time       `orm:"column(hora_apertura);type(time)" json:"horaApertura"`
	PK_ID_CAMBIO_HORARIO *CambiosHorario `orm:"column(pk_id_cambio_horario);rel(fk);null" json:"cambioHorarioId,omitempty" swaggertype:"integer"`
}

func (t *Restaurante) TableName() string {
	return "restaurante"
}

func init() {
	orm.RegisterModel(new(Restaurante))
}

func (r Restaurante) MarshalJSON() ([]byte, error) {

	h := r.HORA_APERTURA
	horaStr := FormatTimeWithLMT(h)

	return json.Marshal(&struct {
		PK_ID_RESTAURANTE    int64           `json:"restauranteId"`
		NOMBRE_RESTAURANTE   string          `json:"nombreRestaurante"`
		HORA_APERTURA        string          `json:"horaApertura"`
		PK_ID_CAMBIO_HORARIO *CambiosHorario `json:"cambioHorarioId,omitempty" swaggertype:"integer"`
	}{
		PK_ID_RESTAURANTE:    r.PK_ID_RESTAURANTE,
		NOMBRE_RESTAURANTE:   r.NOMBRE_RESTAURANTE,
		HORA_APERTURA:        horaStr,
		PK_ID_CAMBIO_HORARIO: r.PK_ID_CAMBIO_HORARIO,
	})
}
