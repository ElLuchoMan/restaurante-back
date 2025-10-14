package models

import (
	"encoding/json"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type PushEnvio struct {
	PkIdPushEnvio       int64            `orm:"column(pk_id_push_envio);pk;auto" json:"pushEnvioId"`
	PkIdPushDispositivo *PushDispositivo `orm:"column(pk_id_push_dispositivo);rel(fk)" json:"pushDispositivoId" swaggertype:"integer"`
	Proveedor           ProveedorPush    `orm:"column(proveedor);type(text)" json:"proveedor"`
	Data                string           `orm:"column(data);type(jsonb);null" json:"-"`
	DataObj             json.RawMessage  `orm:"-" json:"data,omitempty" swaggertype:"object"`
	Exito               bool             `orm:"column(exito);type(boolean)" json:"exito"`
	StatusCode          *int             `orm:"column(status_code);type(integer);null" json:"statusCode,omitempty"`
	ErrorCode           *string          `orm:"column(error_code);type(text);null" json:"errorCode,omitempty"`
	SentAt              time.Time        `orm:"column(sent_at);type(timestamptz)" json:"sentAt" swaggertype:"string"`
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

func (p PushEnvio) MarshalJSON() ([]byte, error) {

	sentAtStr := FormatTimestampBogota(p.SentAt)

	return json.Marshal(&struct {
		PkIdPushEnvio       int64            `json:"pushEnvioId"`
		PkIdPushDispositivo *PushDispositivo `json:"pushDispositivoId" swaggertype:"integer"`
		Proveedor           ProveedorPush    `json:"proveedor"`
		Data                json.RawMessage  `json:"data,omitempty" swaggertype:"object"`
		Exito               bool             `json:"exito"`
		StatusCode          *int             `json:"statusCode,omitempty"`
		ErrorCode           *string          `json:"errorCode,omitempty"`
		SentAt              string           `json:"sentAt"`
	}{
		PkIdPushEnvio:       p.PkIdPushEnvio,
		PkIdPushDispositivo: p.PkIdPushDispositivo,
		Proveedor:           p.Proveedor,
		Data:                p.DataObj,
		Exito:               p.Exito,
		StatusCode:          p.StatusCode,
		ErrorCode:           p.ErrorCode,
		SentAt:              sentAtStr,
	})
}
