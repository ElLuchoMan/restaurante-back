package models

import (
	"encoding/json"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type PushEnvio struct {
	PkIdPushEnvio       int64            `orm:"column(pk_id_push_envio);pk;auto" json:"pushEnvioId"`
	PkIdPushDispositivo *PushDispositivo `orm:"column(pk_id_push_dispositivo);rel(fk)" json:"pushDispositivoId" swaggertype:"integer"`
	Proveedor           ProveedorPush    `orm:"column(proveedor);type(text)" json:"proveedor" enums:"WEB_PUSH,FCM"`
	Data                string           `orm:"column(data);type(jsonb);null" json:"-"`
	DataObj             json.RawMessage  `orm:"-" json:"data,omitempty" swaggertype:"object"`
	Exito               bool             `orm:"column(exito);type(boolean)" json:"exito"`
	StatusCode          *int             `orm:"column(status_code);type(integer);null" json:"statusCode,omitempty"`
	ErrorCode           *string          `orm:"column(error_code);type(text);null" json:"errorCode,omitempty"`
	SentAt              time.Time        `orm:"column(sent_at);type(timestamptz)" json:"sentAt" swaggertype:"string" example:"05-10-2026 14:30:00"`
}

func (p *PushEnvio) TableName() string {
	return "push_envio"
}

func (p *PushEnvio) BeforeInsert() {
	p.serializeData()
}

func (p *PushEnvio) BeforeUpdate() {
	p.serializeData()
}

func (p *PushEnvio) AfterLoad() {
	p.deserializeData()
}

func (p *PushEnvio) serializeData() {
	if len(p.DataObj) == 0 {
		p.Data = ""
		return
	}
	p.Data = string(p.DataObj)
}

func (p *PushEnvio) deserializeData() {
	if p.Data == "" {
		p.DataObj = nil
		return
	}
	p.DataObj = json.RawMessage(p.Data)
}

func init() {
	orm.RegisterModel(new(PushEnvio))
}

// MarshalJSON serializa el envío tal como lo expone el API: pushDispositivoId
// es el id numérico del dispositivo (no el struct PushDispositivo) y sentAt
// es "DD-MM-YYYY HH:MM:SS" en hora de Bogotá.
func (p PushEnvio) MarshalJSON() ([]byte, error) {

	sentAtStr := FormatTimestampBogota(p.SentAt)
	var dispositivoID int64
	if p.PkIdPushDispositivo != nil {
		dispositivoID = p.PkIdPushDispositivo.PkIdPushDispositivo
	}

	return json.Marshal(&struct {
		PkIdPushEnvio       int64           `json:"pushEnvioId"`
		PkIdPushDispositivo int64           `json:"pushDispositivoId"`
		Proveedor           ProveedorPush   `json:"proveedor"`
		Data                json.RawMessage `json:"data,omitempty"`
		Exito               bool            `json:"exito"`
		StatusCode          *int            `json:"statusCode,omitempty"`
		ErrorCode           *string         `json:"errorCode,omitempty"`
		SentAt              string          `json:"sentAt"`
	}{
		PkIdPushEnvio:       p.PkIdPushEnvio,
		PkIdPushDispositivo: dispositivoID,
		Proveedor:           p.Proveedor,
		Data:                p.DataObj,
		Exito:               p.Exito,
		StatusCode:          p.StatusCode,
		ErrorCode:           p.ErrorCode,
		SentAt:              sentAtStr,
	})
}
