package models

import (
	"encoding/json"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type PushDispositivo struct {
	PkIdPushDispositivo   int64                  `orm:"column(pk_id_push_dispositivo);pk;auto" json:"pushDispositivoId"`
	Plataforma            PlataformaNotificacion `orm:"column(plataforma);type(plataforma_notificacion)" json:"plataforma"`
	Endpoint              *string                `orm:"column(endpoint);type(text);null" json:"endpoint,omitempty"`
	P256dh                *string                `orm:"column(p256dh);type(text);null" json:"p256dh,omitempty"`
	Auth                  *string                `orm:"column(auth);type(text);null" json:"auth,omitempty"`
	FcmToken              *string                `orm:"column(fcm_token);type(text);null" json:"fcmToken,omitempty"`
	Enabled               bool                   `orm:"column(enabled);type(boolean);default(true)" json:"enabled"`
	Locale                *string                `orm:"column(locale);type(text);null" json:"locale,omitempty"`
	TimeZone              *string                `orm:"column(time_zone);type(text);null" json:"timeZone,omitempty"`
	AppVersion            *string                `orm:"column(app_version);type(text);null" json:"appVersion,omitempty"`
	UserAgent             *string                `orm:"column(user_agent);type(text);null" json:"userAgent,omitempty"`
	SubscribedTopics      string                 `orm:"column(subscribed_topics);type(text);null" json:"-"`
	SubscribedTopicsArray []string               `orm:"-" json:"subscribedTopics" swaggertype:"array,string"`
	PkDocumentoCliente    *Cliente               `orm:"column(pk_documento_cliente);rel(fk);null" json:"documentoCliente,omitempty" swaggertype:"integer"`
	PkDocumentoTrabajador *Trabajador            `orm:"column(pk_documento_trabajador);rel(fk);null" json:"documentoTrabajador,omitempty" swaggertype:"integer"`
	CreatedAt             time.Time              `orm:"column(created_at);type(timestamptz);auto_now_add" json:"createdAt"`
	LastSeenAt            *time.Time             `orm:"column(last_seen_at);type(timestamptz);null" json:"lastSeenAt,omitempty"`
}

func (p *PushDispositivo) TableName() string {
	return "push_dispositivo"
}

// BeforeInsert se ejecuta antes de insertar en la base de datos
func (p *PushDispositivo) BeforeInsert() {
	p.serializeSubscribedTopics()
}

// BeforeUpdate se ejecuta antes de actualizar en la base de datos
func (p *PushDispositivo) BeforeUpdate() {
	p.serializeSubscribedTopics()
}

// AfterLoad se ejecuta después de cargar desde la base de datos
func (p *PushDispositivo) AfterLoad() {
	p.deserializeSubscribedTopics()
}

// serializeSubscribedTopics convierte el array a string para la base de datos
func (p *PushDispositivo) serializeSubscribedTopics() {
	if len(p.SubscribedTopicsArray) == 0 {
		p.SubscribedTopics = ""
		return
	}
	jsonBytes, _ := json.Marshal(p.SubscribedTopicsArray)
	p.SubscribedTopics = string(jsonBytes)
}

// deserializeSubscribedTopics convierte el string de la base de datos a array
func (p *PushDispositivo) deserializeSubscribedTopics() {
	if p.SubscribedTopics == "" {
		p.SubscribedTopicsArray = []string{}
		return
	}
	_ = json.Unmarshal([]byte(p.SubscribedTopics), &p.SubscribedTopicsArray)
}

func init() {
	orm.RegisterModel(new(PushDispositivo))
}
