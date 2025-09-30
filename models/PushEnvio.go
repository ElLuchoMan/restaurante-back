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
	SentAt              time.Time        `orm:"column(sent_at);type(timestamptz)" json:"sentAt"`
}

func (p *PushEnvio) TableName() string {
	return "push_envio"
}

// BeforeInsert se ejecuta antes de insertar en la base de datos
func (p *PushEnvio) BeforeInsert() {
	p.serializeData()
}

// BeforeUpdate se ejecuta antes de actualizar en la base de datos
func (p *PushEnvio) BeforeUpdate() {
	p.serializeData()
}

// AfterLoad se ejecuta después de cargar desde la base de datos
func (p *PushEnvio) AfterLoad() {
	p.deserializeData()
}

// serializeData convierte el objeto JSON a string para la base de datos
func (p *PushEnvio) serializeData() {
	if len(p.DataObj) == 0 {
		p.Data = ""
		return
	}
	p.Data = string(p.DataObj)
}

// deserializeData convierte el string de la base de datos a objeto JSON
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
