package models

import (
	"encoding/json"
	"strings"
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
	CreatedAt             time.Time              `orm:"column(created_at);type(timestamptz);auto_now_add" json:"createdAt" swaggertype:"string"`
	LastSeenAt            *time.Time             `orm:"column(last_seen_at);type(timestamptz);null" json:"lastSeenAt,omitempty" swaggertype:"string"`
}

func (p *PushDispositivo) TableName() string {
	return "push_dispositivo"
}

func (p *PushDispositivo) BeforeInsert() {
	p.serializeSubscribedTopics()
}

func (p *PushDispositivo) BeforeUpdate() {
	p.serializeSubscribedTopics()
}

func (p *PushDispositivo) AfterLoad() {
	p.deserializeSubscribedTopics()
}

func (p *PushDispositivo) serializeSubscribedTopics() {
	if len(p.SubscribedTopicsArray) == 0 {
		p.SubscribedTopics = ""
		return
	}

	topics := make([]string, len(p.SubscribedTopicsArray))
	for i, topic := range p.SubscribedTopicsArray {

		escapedTopic := strings.ReplaceAll(topic, `"`, `""`)
		topics[i] = `"` + escapedTopic + `"`
	}
	p.SubscribedTopics = "{" + strings.Join(topics, ",") + "}"
}

func (p *PushDispositivo) deserializeSubscribedTopics() {
	if p.SubscribedTopics == "" || p.SubscribedTopics == "{}" {
		p.SubscribedTopicsArray = []string{}
		return
	}

	if strings.HasPrefix(p.SubscribedTopics, "{") && strings.HasSuffix(p.SubscribedTopics, "}") {
		content := p.SubscribedTopics[1 : len(p.SubscribedTopics)-1]
		if content == "" {
			p.SubscribedTopicsArray = []string{}
			return
		}

		parts := strings.Split(content, ",")
		p.SubscribedTopicsArray = make([]string, len(parts))
		for i, part := range parts {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, `"`) && strings.HasSuffix(part, `"`) {
				part = part[1 : len(part)-1]
				part = strings.ReplaceAll(part, `""`, `"`)
			}
			p.SubscribedTopicsArray[i] = part
		}
		return
	}

	_ = json.Unmarshal([]byte(p.SubscribedTopics), &p.SubscribedTopicsArray)
}

func init() {
	orm.RegisterModel(new(PushDispositivo))
}

func (p PushDispositivo) MarshalJSON() ([]byte, error) {

	createdAtStr := FormatTimestampBogota(p.CreatedAt)
	var lastSeenStr *string
	if p.LastSeenAt != nil {
		s := FormatTimestampBogota(*p.LastSeenAt)
		lastSeenStr = &s
	}

	return json.Marshal(&struct {
		PkIdPushDispositivo   int64                  `json:"pushDispositivoId"`
		Plataforma            PlataformaNotificacion `json:"plataforma"`
		Endpoint              *string                `json:"endpoint,omitempty"`
		P256dh                *string                `json:"p256dh,omitempty"`
		Auth                  *string                `json:"auth,omitempty"`
		FcmToken              *string                `json:"fcmToken,omitempty"`
		Enabled               bool                   `json:"enabled"`
		Locale                *string                `json:"locale,omitempty"`
		TimeZone              *string                `json:"timeZone,omitempty"`
		AppVersion            *string                `json:"appVersion,omitempty"`
		UserAgent             *string                `json:"userAgent,omitempty"`
		SubscribedTopicsArray []string               `json:"subscribedTopics" swaggertype:"array,string"`
		PkDocumentoCliente    *Cliente               `json:"documentoCliente,omitempty" swaggertype:"integer"`
		PkDocumentoTrabajador *Trabajador            `json:"documentoTrabajador,omitempty" swaggertype:"integer"`
		CreatedAt             string                 `json:"createdAt"`
		LastSeenAt            *string                `json:"lastSeenAt,omitempty"`
	}{
		PkIdPushDispositivo:   p.PkIdPushDispositivo,
		Plataforma:            p.Plataforma,
		Endpoint:              p.Endpoint,
		P256dh:                p.P256dh,
		Auth:                  p.Auth,
		FcmToken:              p.FcmToken,
		Enabled:               p.Enabled,
		Locale:                p.Locale,
		TimeZone:              p.TimeZone,
		AppVersion:            p.AppVersion,
		UserAgent:             p.UserAgent,
		SubscribedTopicsArray: p.SubscribedTopicsArray,
		PkDocumentoCliente:    p.PkDocumentoCliente,
		PkDocumentoTrabajador: p.PkDocumentoTrabajador,
		CreatedAt:             createdAtStr,
		LastSeenAt:            lastSeenStr,
	})
}
