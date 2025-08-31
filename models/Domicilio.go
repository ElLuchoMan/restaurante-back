package models

import (
	"encoding/json"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type Domicilio struct {
	PK_ID_DOMICILIO         int64           `orm:"column(pk_id_domicilio);pk;auto" json:"domicilioId"`
	DIRECCION               string          `orm:"column(direccion);type(text)" json:"direccion"`
	TELEFONO                string          `orm:"column(telefono);type(text)" json:"telefono"`
	ESTADO_DOMICILIO        EstadoDomicilio `orm:"column(estado_domicilio);type(text)" json:"estado"`
	ENTREGADO               bool            `orm:"column(entregado);type(boolean)" json:"entregado"`
	FECHA                   time.Time       `orm:"column(fecha);type(date)" json:"fechaDomicilio"`
	OBSERVACIONES           string          `orm:"column(observaciones);type(text)" json:"observaciones"`
	CREATED_AT              time.Time       `orm:"column(created_at);type(timestamp);auto_now_add" json:"createdAt"`
	UPDATED_AT              time.Time       `orm:"column(updated_at);type(timestamp);auto_now" json:"updatedAt"`
	CREATED_BY              *string         `orm:"column(created_by);type(text)" json:"createdBy,omitempty"`
	UPDATED_BY              *string         `orm:"column(updated_by);type(text)" json:"updatedBy,omitempty"`
	PK_DOCUMENTO_TRABAJADOR *int            `orm:"column(pk_documento_trabajador);null" json:"trabajadorAsignado,omitempty"`
}

func (d *Domicilio) TableName() string {
	return "domicilio"
}

func init() {
	orm.RegisterModel(new(Domicilio))
}

// Personalizar el formato de fecha para la API
func (d Domicilio) MarshalJSON() ([]byte, error) {
	type Alias Domicilio
	return json.Marshal(&struct {
		FECHA      string `json:"fechaDomicilio"`
		CREATED_AT string `json:"createdAt"`
		UPDATED_AT string `json:"updatedAt"`
		Alias
	}{
		FECHA:      d.FECHA.Format("02-01-2006"),
		CREATED_AT: d.CREATED_AT.Format("02-01-2006 15:04:05"),
		UPDATED_AT: d.UPDATED_AT.Format("02-01-2006 15:04:05"),
		Alias:      (Alias)(d),
	})
}
