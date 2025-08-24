package models

import (
	"encoding/json"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type Domicilio struct {
	PK_ID_DOMICILIO         int       `orm:"column(PK_ID_DOMICILIO);pk;auto" json:"domicilioId"`
	DIRECCION               string    `orm:"column(DIRECCION);type(text)" json:"direccion"`
	TELEFONO                string    `orm:"column(TELEFONO);type(text)" json:"telefono"`
	ESTADO_PAGO             string    `orm:"column(ESTADO_PAGO);type(text)" json:"estadoPago"`
	ENTREGADO               bool      `orm:"column(ENTREGADO);type(boolean)" json:"entregado"`
	FECHA                   time.Time `orm:"column(FECHA);type(date)" json:"fechaDomicilio"`
	OBSERVACIONES           string    `orm:"column(OBSERVACIONES);type(text)" json:"observaciones"`
	CREATED_AT              time.Time `orm:"column(CREATED_AT);type(timestamp);auto_now_add" json:"createdAt"`
	UPDATED_AT              time.Time `orm:"column(UPDATED_AT);type(timestamp);auto_now" json:"updatedAt"`
	CREATED_BY              *string   `orm:"column(CREATED_BY);type(text)" json:"createdBy,omitempty"`
	UPDATED_BY              *string   `orm:"column(UPDATED_BY);type(text)" json:"updatedBy,omitempty"`
	PK_DOCUMENTO_TRABAJADOR *int      `orm:"column(PK_DOCUMENTO_TRABAJADOR);null" json:"trabajadorAsignado,omitempty"`
}

// DomicilioDetail extiende la información de un domicilio con datos del cliente asociado.
type DomicilioDetail struct {
        Domicilio
        NombreCliente    string `json:"nombreCliente" orm:"column(nombre_cliente)"`
        DocumentoCliente int64  `json:"documentoCliente" orm:"column(documento_cliente)"`
}

func (d *Domicilio) TableName() string {
	return "DOMICILIO"
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

// MarshalJSON para DomicilioDetail formatea las fechas igual que el modelo base.
func (d DomicilioDetail) MarshalJSON() ([]byte, error) {
        type Alias DomicilioDetail
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
