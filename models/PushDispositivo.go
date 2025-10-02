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
		p.SubscribedTopics = "" // String vacío para PostgreSQL
		return
	}
	// Crear array de PostgreSQL manualmente
	topics := make([]string, len(p.SubscribedTopicsArray))
	for i, topic := range p.SubscribedTopicsArray {
		// Escapar comillas dobles para PostgreSQL array
		escapedTopic := strings.ReplaceAll(topic, `"`, `""`)
		topics[i] = `"` + escapedTopic + `"`
	}
	p.SubscribedTopics = "{" + strings.Join(topics, ",") + "}"
}

// deserializeSubscribedTopics convierte el string de la base de datos a array
func (p *PushDispositivo) deserializeSubscribedTopics() {
	if p.SubscribedTopics == "" || p.SubscribedTopics == "{}" {
		p.SubscribedTopicsArray = []string{}
		return
	}

	// Parsear array de PostgreSQL manualmente
	if strings.HasPrefix(p.SubscribedTopics, "{") && strings.HasSuffix(p.SubscribedTopics, "}") {
		content := p.SubscribedTopics[1 : len(p.SubscribedTopics)-1] // Remover { }
		if content == "" {
			p.SubscribedTopicsArray = []string{}
			return
		}

		// Dividir por comas y limpiar comillas
		parts := strings.Split(content, ",")
		p.SubscribedTopicsArray = make([]string, len(parts))
		for i, part := range parts {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, `"`) && strings.HasSuffix(part, `"`) {
				part = part[1 : len(part)-1]               // Remover comillas
				part = strings.ReplaceAll(part, `""`, `"`) // Desescapar comillas dobles
			}
			p.SubscribedTopicsArray[i] = part
		}
		return
	}

	// Fallback: intentar como JSON
	_ = json.Unmarshal([]byte(p.SubscribedTopics), &p.SubscribedTopicsArray)
}

func init() {
	orm.RegisterModel(new(PushDispositivo))
}
