package services

import (
	"context"

	"restaurante/models"
)

// PushServiceInterface define los métodos del servicio de notificaciones push
// Esta interfaz permite mockear el servicio en tests unitarios
type PushServiceInterface interface {
	// RegistrarDispositivo registra un nuevo dispositivo para notificaciones push
	RegistrarDispositivo(ctx context.Context, req *models.RegistrarDispositivoRequest) (*models.PushDispositivo, error)

	// ActualizarUltimaVista actualiza la fecha de última vista de un dispositivo
	ActualizarUltimaVista(ctx context.Context, dispositivoId int64) error

	// ActualizarEstadoDispositivo actualiza el estado enabled/disabled de un dispositivo
	ActualizarEstadoDispositivo(ctx context.Context, dispositivoId int64, enabled bool) error

	// ActualizarTopicsDispositivo actualiza los topics suscritos de un dispositivo
	ActualizarTopicsDispositivo(ctx context.Context, dispositivoId int64, topics []string) error

	// RegistrarEnvio registra un envío de notificación en la base de datos
	RegistrarEnvio(ctx context.Context, req *models.RegistrarEnvioRequest) (*models.PushEnvio, error)

	// EnviarNotificacion envía una notificación push a los destinatarios especificados
	EnviarNotificacion(req *models.EnviarNotificacionRequest) (*models.EnviarNotificacionResponse, error)

	// ValidarRegistroDispositivo valida que un request de registro sea válido
	ValidarRegistroDispositivo(req *models.RegistrarDispositivoRequest) error
}

// Verificar que PushService implementa la interfaz (compile-time check)
var _ PushServiceInterface = (*PushService)(nil)
