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
	type Alias Nomina

	// Cargar zona horaria de Bogotá
	loc, err := time.LoadLocation("America/Bogota")
	if err != nil {
		// Fallback a UTC-5 si no se puede cargar
		loc = time.FixedZone("UTC-5", -5*60*60)
	}

	return json.Marshal(&struct {
		FECHA string `json:"fechaNomina"`
		Alias
	}{
		FECHA: t.FECHA.In(loc).Format("02-01-2006"),
		Alias: (Alias)(t),
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
