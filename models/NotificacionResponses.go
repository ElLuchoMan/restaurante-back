package models

import (
	"encoding/json"
	"time"
)

type ValidarCuponResponse struct {
	Aplicable      bool    `json:"aplicable"`
	MontoDescuento int64   `json:"montoDescuento"`
	Motivo         *string `json:"motivo,omitempty"`
}

type OfertaActivaResponse struct {
	OfertaId       int64         `json:"ofertaId"`
	Titulo         string        `json:"titulo"`
	TipoDescuento  TipoDescuento `json:"tipoDescuento"`
	ValorDescuento int64         `json:"valorDescuento"`
	ProductosIds   []int64       `json:"productosIds"`
}

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"pageSize"`
	TotalPages int         `json:"totalPages"`
}

type ValidationErrorResponse struct {
	Message string                 `json:"message"`
	Errors  map[string]interface{} `json:"errors,omitempty"`
}

type ConflictErrorResponse struct {
	Message string      `json:"message"`
	Code    string      `json:"code"`
	Details interface{} `json:"details,omitempty"`
}

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

func (p PushDispositivoResponse) MarshalJSON() ([]byte, error) {
	var lastSeenStr *string
	if p.LastSeenAt != nil {
		s := FormatTimestampBogota(*p.LastSeenAt)
		lastSeenStr = &s
	}
	return json.Marshal(&struct {
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
		CreatedAt           string                 `json:"createdAt"`
		LastSeenAt          *string                `json:"lastSeenAt,omitempty"`
	}{
		PushDispositivoId:   p.PushDispositivoId,
		Plataforma:          p.Plataforma,
		Endpoint:            p.Endpoint,
		FcmToken:            p.FcmToken,
		Enabled:             p.Enabled,
		Locale:              p.Locale,
		TimeZone:            p.TimeZone,
		AppVersion:          p.AppVersion,
		SubscribedTopics:    p.SubscribedTopics,
		DocumentoCliente:    p.DocumentoCliente,
		DocumentoTrabajador: p.DocumentoTrabajador,
		CreatedAt:           FormatTimestampBogota(p.CreatedAt),
		LastSeenAt:          lastSeenStr,
	})
}

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

func (p PushEnvioResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		PushEnvioId       int64           `json:"pushEnvioId"`
		PushDispositivoId int64           `json:"pushDispositivoId"`
		Proveedor         ProveedorPush   `json:"proveedor"`
		Data              json.RawMessage `json:"data,omitempty" swaggertype:"object"`
		Exito             bool            `json:"exito"`
		StatusCode        *int            `json:"statusCode,omitempty"`
		ErrorCode         *string         `json:"errorCode,omitempty"`
		SentAt            string          `json:"sentAt"`
	}{
		PushEnvioId:       p.PushEnvioId,
		PushDispositivoId: p.PushDispositivoId,
		Proveedor:         p.Proveedor,
		Data:              p.Data,
		Exito:             p.Exito,
		StatusCode:        p.StatusCode,
		ErrorCode:         p.ErrorCode,
		SentAt:            FormatTimestampBogota(p.SentAt),
	})
}

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

type EnviarNotificacionResponse struct {
	TotalDispositivos    int                        `json:"totalDispositivos"`
	EnviosExitosos       int                        `json:"enviosExitosos"`
	EnviosFallidos       int                        `json:"enviosFallidos"`
	DetalleEnvios        []DetalleEnvioNotificacion `json:"detalleEnvios"`
	ResumenDestinatarios ResumenDestinatarios       `json:"resumenDestinatarios"`
}

type DetalleEnvioNotificacion struct {
	PushDispositivoId   int64   `json:"pushDispositivoId"`
	Plataforma          string  `json:"plataforma"`
	Exito               bool    `json:"exito"`
	StatusCode          *int    `json:"statusCode,omitempty"`
	ErrorCode           *string `json:"errorCode,omitempty"`
	DocumentoCliente    *int64  `json:"documentoCliente,omitempty"`
	DocumentoTrabajador *int64  `json:"documentoTrabajador,omitempty"`
}

type ResumenDestinatarios struct {
	TipoDestinatario        string   `json:"tipoDestinatario"`
	ClientesNotificados     []int64  `json:"clientesNotificados,omitempty"`
	TrabajadoresNotificados []int64  `json:"trabajadoresNotificados,omitempty"`
	TopicsNotificados       []string `json:"topicsNotificados,omitempty"`
}
