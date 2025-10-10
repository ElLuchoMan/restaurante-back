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
	// Cargar zona horaria de Bogotá
	loc, err := time.LoadLocation("America/Bogota")
	if err != nil {
		loc = time.FixedZone("UTC-5", -5*60*60)
	}

	// Para campos de tipo date: extraer SOLO año/mes/día sin importar timezone
	fechaStr := time.Date(d.FECHA.Year(), d.FECHA.Month(), d.FECHA.Day(), 0, 0, 0, 0, time.UTC).Format("02-01-2006")

	// Para campos de tipo time: convertir a zona horaria de Bogotá
	horaEnBogota := d.HORA.In(loc)
	updatedAtEnBogota := d.UPDATED_AT.In(loc)

	return json.Marshal(&struct {
		PK_ID_PAGO        int64       `json:"pagoId"`
		FECHA             string      `json:"fechaPago"`
		HORA              string      `json:"horaPago"`
		MONTO             int64       `json:"monto"`
		ESTADO_PAGO       EstadoPago  `json:"estadoPago"`
		PK_ID_METODO_PAGO *MetodoPago `json:"metodoPagoId"`
		UPDATED_AT        string      `json:"updatedAt"`
		UPDATED_BY        *string     `json:"updatedBy,omitempty"`
	}{
		PK_ID_PAGO:        d.PK_ID_PAGO,
		FECHA:             fechaStr,
		HORA:              horaEnBogota.Format("15:04:05"),
		MONTO:             d.MONTO,
		ESTADO_PAGO:       d.ESTADO_PAGO,
		PK_ID_METODO_PAGO: d.PK_ID_METODO_PAGO,
		UPDATED_AT:        updatedAtEnBogota.Format("02-01-2006 15:04:05"),
		UPDATED_BY:        d.UPDATED_BY,
	})
}
