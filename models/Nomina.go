package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type Nomina struct {
	PK_ID_NOMINA  int64        `orm:"column(pk_id_nomina);pk;auto" json:"nominaId"`
	FECHA         time.Time    `orm:"column(fecha);type(date);unique" json:"fechaNomina"`
	MONTO         int64        `orm:"column(monto)" json:"monto"`
	ESTADO_NOMINA EstadoNomina `orm:"column(estado_nomina);type(estado_nomina)" json:"estadoNomina"`
}

func (n *Nomina) TableName() string {
	return "nomina"
}

func init() {
	orm.RegisterModel(new(Nomina))
}

func (t Nomina) MarshalJSON() ([]byte, error) {
	// FECHA: normalizar a UTC para obtener el día de calendario correcto, sin efectos de zona
	fechaUTC := t.FECHA.UTC()
	fechaStr := fmt.Sprintf("%02d-%02d-%04d", fechaUTC.Day(), int(fechaUTC.Month()), fechaUTC.Year())

	return json.Marshal(&struct {
		PK_ID_NOMINA  int64        `json:"nominaId"`
		FECHA         string       `json:"fechaNomina"`
		MONTO         int64        `json:"monto"`
		ESTADO_NOMINA EstadoNomina `json:"estadoNomina"`
	}{
		PK_ID_NOMINA:  t.PK_ID_NOMINA,
		FECHA:         fechaStr,
		MONTO:         t.MONTO,
		ESTADO_NOMINA: t.ESTADO_NOMINA,
	})
}
func (n *Nomina) UnmarshalJSON(data []byte) error {
	type Alias Nomina
	aux := &struct {
		FECHA string `json:"fechaNomina"`
		*Alias
	}{
		Alias: (*Alias)(n),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if aux.FECHA != "" {
		t, err := time.Parse("2006-01-02", aux.FECHA)
		if err != nil {
			return fmt.Errorf("fechaNomina debe tener formato YYYY-MM-DD: %w", err)
		}
		n.FECHA = t
	}
	return nil
}
