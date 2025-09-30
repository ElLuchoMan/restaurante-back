package models

import (
	"encoding/json"
)

// Requests para Push Dispositivos
type RegistrarDispositivoRequest struct {
	Plataforma            PlataformaNotificacion `json:"plataforma" valid:"required"`
	Endpoint              *string                `json:"endpoint,omitempty"`
	P256dh                *string                `json:"p256dh,omitempty"`
	Auth                  *string                `json:"auth,omitempty"`
	FcmToken              *string                `json:"fcmToken,omitempty"`
	Locale                *string                `json:"locale,omitempty"`
	TimeZone              *string                `json:"timeZone,omitempty"`
	AppVersion            *string                `json:"appVersion,omitempty"`
	UserAgent             *string                `json:"userAgent,omitempty"`
	SubscribedTopics      []string               `json:"subscribedTopics,omitempty" swaggertype:"array,string"`
	PkDocumentoCliente    *int64                 `json:"documentoCliente,omitempty"`
	PkDocumentoTrabajador *int64                 `json:"documentoTrabajador,omitempty"`
}

type ActualizarEstadoDispositivoRequest struct {
	Enabled bool `json:"enabled" valid:"required"`
}

type ActualizarTopicsRequest struct {
	SubscribedTopics []string `json:"subscribedTopics" valid:"required" swaggertype:"array,string"`
}

// Requests para Push Envíos
type RegistrarEnvioRequest struct {
	PkIdPushDispositivo int64           `json:"pushDispositivoId" valid:"required"`
	Proveedor           ProveedorPush   `json:"proveedor" valid:"required"`
	Data                json.RawMessage `json:"data,omitempty" swaggertype:"object"`
	Exito               bool            `json:"exito" valid:"required"`
	StatusCode          *int            `json:"statusCode,omitempty"`
	ErrorCode           *string         `json:"errorCode,omitempty"`
}

// Requests para Cupones
type CrearCuponRequest struct {
	Codigo             string        `json:"codigo" valid:"required,length(3|50)"`
	Scope              CuponScope    `json:"scope" valid:"required"`
	TipoDescuento      TipoDescuento `json:"tipoDescuento" valid:"required"`
	ValorDescuento     int64         `json:"valorDescuento" valid:"required,min(1)"`
	MaxUsos            *int          `json:"maxUsos,omitempty" valid:"min(1)"`
	LimitePorCliente   *int          `json:"limitePorCliente,omitempty" valid:"min(1)"`
	MontoMinimo        *int64        `json:"montoMinimo,omitempty" valid:"min(0)"`
	FechaInicio        string        `json:"fechaInicio" valid:"required"`
	FechaFin           string        `json:"fechaFin" valid:"required"`
	PkIdProducto       *int64        `json:"productoId,omitempty"`
	PkIdCategoria      *int64        `json:"categoriaId,omitempty"`
	PkDocumentoCliente *int64        `json:"documentoCliente,omitempty"`
}

type ValidarCuponRequest struct {
	PedidoId  *int64                    `json:"pedidoId,omitempty"`
	ClienteId int64                     `json:"clienteId" valid:"required"`
	Items     []ValidarCuponItemRequest `json:"items" valid:"required"`
	Codigo    string                    `json:"codigo" valid:"required"`
}

type ValidarCuponItemRequest struct {
	ProductoId int64 `json:"productoId" valid:"required"`
	Cantidad   int   `json:"cantidad" valid:"required,min(1)"`
	Precio     int64 `json:"precio" valid:"required,min(0)"`
}

type RedimirCuponRequest struct {
	ClienteId int64  `json:"clienteId" valid:"required"`
	PedidoId  *int64 `json:"pedidoId,omitempty"`
}

// Requests para Ofertas
type CrearOfertaRequest struct {
	Titulo          string        `json:"titulo" valid:"required,length(3|100)"`
	TipoDescuento   TipoDescuento `json:"tipoDescuento" valid:"required"`
	ValorDescuento  int64         `json:"valorDescuento" valid:"required,min(1)"`
	FechaInicio     string        `json:"fechaInicio" valid:"required"`
	FechaFin        string        `json:"fechaFin" valid:"required"`
	DiasSemana      []string      `json:"diasSemana,omitempty" swaggertype:"array,string"`
	HoraInicio      *string       `json:"horaInicio,omitempty"`
	HoraFin         *string       `json:"horaFin,omitempty"`
	PkIdRestaurante int64         `json:"restauranteId" valid:"required"`
}

type AsociarProductoOfertaRequest struct {
	ProductoId int64 `json:"productoId" valid:"required"`
}

// Requests para Pedido Descuento Aplicado
type AplicarDescuentoRequest struct {
	PkIdCupon      *int64          `json:"cuponId,omitempty"`
	PkIdOferta     *int64          `json:"ofertaId,omitempty"`
	MontoDescuento int64           `json:"montoDescuento" valid:"required,min(0)"`
	Detalle        json.RawMessage `json:"detalle,omitempty" swaggertype:"object"`
}

// Tipos para el remitente de notificaciones
type TipoRemitente string

const (
	RemitenteTrabajador TipoRemitente = "TRABAJADOR"
	RemitenteSistema    TipoRemitente = "SISTEMA"
)

// Tipos para destinatarios de notificaciones
type TipoDestinatario string

const (
	DestinatarioTodos      TipoDestinatario = "TODOS"
	DestinatarioCliente    TipoDestinatario = "CLIENTE"
	DestinatarioTrabajador TipoDestinatario = "TRABAJADOR"
	DestinatarioTopic      TipoDestinatario = "TOPIC"
)

// Estructura del remitente
type RemitenteNotificacion struct {
	Tipo                TipoRemitente `json:"tipo" valid:"required,in(TRABAJADOR|SISTEMA)"`
	DocumentoTrabajador *int64        `json:"documentoTrabajador,omitempty"`
	Nombre              *string       `json:"nombre,omitempty"`
}

// Estructura de destinatarios
type DestinatariosNotificacion struct {
	Tipo                TipoDestinatario `json:"tipo" valid:"required,in(TODOS|CLIENTE|TRABAJADOR|TOPIC)"`
	DocumentoCliente    *int64           `json:"documentoCliente,omitempty"`
	DocumentoTrabajador *int64           `json:"documentoTrabajador,omitempty"`
	Topic               *string          `json:"topic,omitempty"`
}

// Estructura de la notificación
type ContenidoNotificacion struct {
	Titulo  string          `json:"titulo" valid:"required,length(1|100)"`
	Mensaje string          `json:"mensaje" valid:"required,length(1|500)"`
	Datos   json.RawMessage `json:"datos,omitempty" swaggertype:"object"`
}

// Request para enviar notificación
type EnviarNotificacionRequest struct {
	Remitente     RemitenteNotificacion     `json:"remitente" valid:"required"`
	Destinatarios DestinatariosNotificacion `json:"destinatarios" valid:"required"`
	Notificacion  ContenidoNotificacion     `json:"notificacion" valid:"required"`
}
