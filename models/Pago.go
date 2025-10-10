package models

import (
	"encoding/json"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type Pago struct {
	PK_ID_PAGO        int64       `orm:"column(pk_id_pago);pk;auto" json:"pagoId"`
	FECHA             time.Time   `orm:"column(fecha);type(date)" json:"fechaPago"`
	HORA              time.Time   `orm:"column(hora);type(time)" json:"horaPago"`
	MONTO             int64       `orm:"column(monto)" json:"monto"`
	ESTADO_PAGO       EstadoPago  `orm:"column(estado_pago);type(estado_pago)" json:"estadoPago"`
	PK_ID_METODO_PAGO *MetodoPago `orm:"column(pk_id_metodo_pago);rel(fk)" json:"metodoPagoId" swaggertype:"integer"`
	UPDATED_AT        time.Time   `orm:"column(updated_at);type(timestamptz);auto_now" json:"updatedAt"`
	UPDATED_BY        *string     `orm:"column(updated_by);type(text);null" json:"updatedBy,omitempty"`
}

func (p *Pago) TableName() string {
	return "pago"
}

func init() {
	orm.RegisterModel(new(Pago))
}

func (d Pago) MarshalJSON() ([]byte, error) {
	type Alias Pago

	// Cargar zona horaria de Bogotá
	loc, err := time.LoadLocation("America/Bogota")
	if err != nil {
		// Fallback a UTC-5 si no se puede cargar
		loc = time.FixedZone("UTC-5", -5*60*60)
	}

	// Convertir fechas a zona horaria de Bogotá antes de formatear
	fechaBogota := d.FECHA.In(loc)
	horaBogota := d.HORA.In(loc)
	updatedBogota := d.UPDATED_AT.In(loc)

	return json.Marshal(&struct {
		FECHA      string `json:"fechaPago"`
		HORA       string `json:"horaPago"`
		UPDATED_AT string `json:"updatedAt"`
		Alias
	}{
		FECHA:      fechaBogota.Format("02-01-2006"),
		HORA:       horaBogota.Format("15:04:05"),
		UPDATED_AT: updatedBogota.Format("02-01-2006 15:04:05"),
		Alias:      (Alias)(d),
	})
}
