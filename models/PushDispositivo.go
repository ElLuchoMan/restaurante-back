package models

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type PushDispositivo struct {
	PkIdPushDispositivo   int64                  `orm:"column(pk_id_push_dispositivo);pk;auto" json:"pushDispositivoId"`
	Plataforma            PlataformaNotificacion `orm:"column(plataforma);type(plataforma_notificacion)" json:"plataforma" enums:"WEB,ANDROID,IOS"`
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
	CreatedAt             time.Time              `orm:"column(created_at);type(timestamptz);auto_now_add" json:"createdAt" swaggertype:"string" example:"05-10-2026 14:30:00"`
	LastSeenAt            *time.Time             `orm:"column(last_seen_at);type(timestamptz);null" json:"lastSeenAt,omitempty" swaggertype:"string" example:"05-10-2026 14:30:00"`
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

// MarshalJSON serializa el dispositivo tal como lo expone el API: las FK
// documentoCliente / documentoTrabajador salen como número de documento (nunca
// como el struct Cliente/Trabajador completo, que incluiría la contraseña),
// subscribedTopics siempre es una lista (nunca null) y las fechas son
// "DD-MM-YYYY HH:MM:SS" en hora de Bogotá.
func (p PushDispositivo) MarshalJSON() ([]byte, error) {

	createdAtStr := FormatTimestampBogota(p.CreatedAt)
	var lastSeenStr *string
	if p.LastSeenAt != nil {
		s := FormatTimestampBogota(*p.LastSeenAt)
		lastSeenStr = &s
	}
	var documentoCliente, documentoTrabajador *int64
	if p.PkDocumentoCliente != nil {
		documentoCliente = &p.PkDocumentoCliente.PK_DOCUMENTO_CLIENTE
	}
	if p.PkDocumentoTrabajador != nil {
		documentoTrabajador = &p.PkDocumentoTrabajador.PK_DOCUMENTO_TRABAJADOR
	}

	return json.Marshal(&struct {
		PkIdPushDispositivo int64                  `json:"pushDispositivoId"`
		Plataforma          PlataformaNotificacion `json:"plataforma"`
		Endpoint            *string                `json:"endpoint,omitempty"`
		P256dh              *string                `json:"p256dh,omitempty"`
		Auth                *string                `json:"auth,omitempty"`
		FcmToken            *string                `json:"fcmToken,omitempty"`
		Enabled             bool                   `json:"enabled"`
		Locale              *string                `json:"locale,omitempty"`
		TimeZone            *string                `json:"timeZone,omitempty"`
		AppVersion          *string                `json:"appVersion,omitempty"`
		UserAgent           *string                `json:"userAgent,omitempty"`
		SubscribedTopics    []string               `json:"subscribedTopics"`
		DocumentoCliente    *int64                 `json:"documentoCliente,omitempty"`
		DocumentoTrabajador *int64                 `json:"documentoTrabajador,omitempty"`
		CreatedAt           string                 `json:"createdAt"`
		LastSeenAt          *string                `json:"lastSeenAt,omitempty"`
	}{
		PkIdPushDispositivo: p.PkIdPushDispositivo,
		Plataforma:          p.Plataforma,
		Endpoint:            p.Endpoint,
		P256dh:              p.P256dh,
		Auth:                p.Auth,
		FcmToken:            p.FcmToken,
		Enabled:             p.Enabled,
		Locale:              p.Locale,
		TimeZone:            p.TimeZone,
		AppVersion:          p.AppVersion,
		UserAgent:           p.UserAgent,
		SubscribedTopics:    emptySliceIfNil(p.SubscribedTopicsArray),
		DocumentoCliente:    documentoCliente,
		DocumentoTrabajador: documentoTrabajador,
		CreatedAt:           createdAtStr,
		LastSeenAt:          lastSeenStr,
	})
}
