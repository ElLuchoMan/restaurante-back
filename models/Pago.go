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
	UPDATED_AT        time.Time   `orm:"column(updated_at);type(timestamptz);auto_now" json:"updatedAt" swaggertype:"string"`
	UPDATED_BY        *string     `orm:"column(updated_by);type(text);null" json:"updatedBy,omitempty"`
}

func (p *Pago) TableName() string {
	return "pago"
}

func init() {
	orm.RegisterModel(new(Pago))
}

func (d Pago) MarshalJSON() ([]byte, error) {

	fechaStr := FormatDateUTC(d.FECHA)

	horaStr := FormatTimeWithLMT(d.HORA)

	updatedAtStr := FormatTimestampBogota(d.UPDATED_AT)

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
		HORA:              horaStr,
		MONTO:             d.MONTO,
		ESTADO_PAGO:       d.ESTADO_PAGO,
		PK_ID_METODO_PAGO: d.PK_ID_METODO_PAGO,
		UPDATED_AT:        updatedAtStr,
		UPDATED_BY:        d.UPDATED_BY,
	})
}
