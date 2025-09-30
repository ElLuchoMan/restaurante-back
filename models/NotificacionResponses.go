package models

import (
	"encoding/json"
	"time"
)

// Responses para validación de cupones
type ValidarCuponResponse struct {
	Aplicable      bool    `json:"aplicable"`
	MontoDescuento int64   `json:"montoDescuento"`
	Motivo         *string `json:"motivo,omitempty"`
}

// Responses para ofertas activas
type OfertaActivaResponse struct {
	OfertaId       int64         `json:"ofertaId"`
	Titulo         string        `json:"titulo"`
	TipoDescuento  TipoDescuento `json:"tipoDescuento"`
	ValorDescuento int64         `json:"valorDescuento"`
	ProductosIds   []int64       `json:"productosIds"`
}

// Response genérica para listas paginadas
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"pageSize"`
	TotalPages int         `json:"totalPages"`
}

// Response para errores de validación
type ValidationErrorResponse struct {
	Message string                 `json:"message"`
	Errors  map[string]interface{} `json:"errors,omitempty"`
}

// Response para conflictos
type ConflictErrorResponse struct {
	Message string      `json:"message"`
	Code    string      `json:"code"`
	Details interface{} `json:"details,omitempty"`
}

// Response para dispositivos con información adicional
type PushDispositivoResponse struct {
	PushDispositivoId   int64                  `json:"pushDispositivoId"`
	Plataforma          PlataformaNotificacion `json:"plataforma"`
	Endpoint            *string                `json:"endpoint,omitempty"`
	FcmToken            *string                `json:"fcmToken,omitempty"`
	Enabled             bool                   `json:"enabled"`
	Locale              *string                `json:"locale,omitempty"`
	TimeZone            *string                `json:"timeZone,omitempty"`
	AppVersion          *string                `json:"appVersion,omitempty"`
	SubscribedTopics    []string               `json:"subscribedTopics"`
	DocumentoCliente    *int64                 `json:"documentoCliente,omitempty"`
	DocumentoTrabajador *int64                 `json:"documentoTrabajador,omitempty"`
	CreatedAt           time.Time              `json:"createdAt"`
	LastSeenAt          *time.Time             `json:"lastSeenAt,omitempty"`
}

// Response para envíos con información adicional
type PushEnvioResponse struct {
	PushEnvioId       int64           `json:"pushEnvioId"`
	PushDispositivoId int64           `json:"pushDispositivoId"`
	Proveedor         ProveedorPush   `json:"proveedor"`
	Data              json.RawMessage `json:"data,omitempty" swaggertype:"object"`
	Exito             bool            `json:"exito"`
	StatusCode        *int            `json:"statusCode,omitempty"`
	ErrorCode         *string         `json:"errorCode,omitempty"`
	SentAt            time.Time       `json:"sentAt"`
}

// Response para cupones con información adicional
type CuponResponse struct {
	CuponId          int64         `json:"cuponId"`
	Codigo           string        `json:"codigo"`
	Scope            CuponScope    `json:"scope"`
	TipoDescuento    TipoDescuento `json:"tipoDescuento"`
	ValorDescuento   int64         `json:"valorDescuento"`
	MaxUsos          *int          `json:"maxUsos,omitempty"`
	LimitePorCliente *int          `json:"limitePorCliente,omitempty"`
	MontoMinimo      *int64        `json:"montoMinimo,omitempty"`
	FechaInicio      time.Time     `json:"fechaInicio"`
	FechaFin         time.Time     `json:"fechaFin"`
	ProductoId       *int64        `json:"productoId,omitempty"`
	CategoriaId      *int64        `json:"categoriaId,omitempty"`
	DocumentoCliente *int64        `json:"documentoCliente,omitempty"`
	Activo           bool          `json:"activo"`
	UsosActuales     int           `json:"usosActuales"`
}

// Response para ofertas con información adicional
type OfertaResponse struct {
	OfertaId       int64         `json:"ofertaId"`
	Titulo         string        `json:"titulo"`
	TipoDescuento  TipoDescuento `json:"tipoDescuento"`
	ValorDescuento int64         `json:"valorDescuento"`
	FechaInicio    time.Time     `json:"fechaInicio"`
	FechaFin       time.Time     `json:"fechaFin"`
	DiasSemana     []string      `json:"diasSemana"`
	HoraInicio     *time.Time    `json:"horaInicio,omitempty"`
	HoraFin        *time.Time    `json:"horaFin,omitempty"`
	Activo         bool          `json:"activo"`
	RestauranteId  int64         `json:"restauranteId"`
	ProductosIds   []int64       `json:"productosIds"`
}

// Respuesta del envío de notificación
type EnviarNotificacionResponse struct {
	TotalDispositivos    int                        `json:"totalDispositivos"`
	EnviosExitosos       int                        `json:"enviosExitosos"`
	EnviosFallidos       int                        `json:"enviosFallidos"`
	DetalleEnvios        []DetalleEnvioNotificacion `json:"detalleEnvios"`
	ResumenDestinatarios ResumenDestinatarios       `json:"resumenDestinatarios"`
}

// Detalle de cada envío individual
type DetalleEnvioNotificacion struct {
	PushDispositivoId   int64   `json:"pushDispositivoId"`
	Plataforma          string  `json:"plataforma"`
	Exito               bool    `json:"exito"`
	StatusCode          *int    `json:"statusCode,omitempty"`
	ErrorCode           *string `json:"errorCode,omitempty"`
	DocumentoCliente    *int64  `json:"documentoCliente,omitempty"`
	DocumentoTrabajador *int64  `json:"documentoTrabajador,omitempty"`
}

// Resumen de destinatarios
type ResumenDestinatarios struct {
	TipoDestinatario        string   `json:"tipoDestinatario"`
	ClientesNotificados     []int64  `json:"clientesNotificados,omitempty"`
	TrabajadoresNotificados []int64  `json:"trabajadoresNotificados,omitempty"`
	TopicsNotificados       []string `json:"topicsNotificados,omitempty"`
}
