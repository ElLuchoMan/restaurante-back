package models

import (
	"encoding/json"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type Domicilio struct {
	PK_ID_DOMICILIO int       `orm:"column(PK_ID_DOMICILIO);pk;auto" json:"PK_ID_DOMICILIO"`
	DIRECCION       string    `orm:"column(DIRECCION);type(text)" json:"DIRECCION"`
	TELEFONO        string    `orm:"column(TELEFONO);type(text)" json:"TELEFONO"`
	ESTADO_PAGO     string    `orm:"column(ESTADO_PAGO);type(text)" json:"ESTADO_PAGO"`
	ENTREGADO       bool      `orm:"column(ENTREGADO);type(boolean)" json:"ENTREGADO"`
	FECHA           time.Time `orm:"column(FECHA);type(date)" json:"FECHA"`
	OBSERVACIONES   string    `orm:"column(OBSERVACIONES);type(text)" json:"OBSERVACIONES"`
	CREATED_AT      time.Time `orm:"column(CREATED_AT);type(timestamp);auto_now_add" json:"CREATED_AT"`
	UPDATED_AT      time.Time `orm:"column(UPDATED_AT);type(timestamp);auto_now" json:"UPDATED_AT"`
	CREATED_BY      *string   `orm:"column(CREATED_BY);type(date)" json:"CREATED_BY,omitempty"`
	UPDATED_BY      *string   `orm:"column(UPDATED_BY);type(date)" json:"UPDATED_BY,omitempty"`
}

func (d *Domicilio) TableName() string {
	return "DOMICILIO"
}

func init() {
	orm.RegisterModel(new(Domicilio))
}

func (d Domicilio) MarshalJSON() ([]byte, error) {
	type Alias Domicilio
	return json.Marshal(&struct {
		FECHA      string `json:"FECHA"`
		CREATED_AT string `json:"CREATED_AT"`
		UPDATED_AT string `json:"UPDATED_AT"`
		Alias
	}{
		FECHA:      d.FECHA.Format("02-01-2006"),
		CREATED_AT: d.CREATED_AT.Format("02-01-2006 15:04:05"),
		UPDATED_AT: d.UPDATED_AT.Format("02-01-2006 15:04:05"),
		Alias:      (Alias)(d),
	})
}
