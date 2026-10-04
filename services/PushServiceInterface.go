package services

import (
	"context"

	"restaurante/models"
)

type PushServiceInterface interface {
	RegistrarDispositivo(ctx context.Context, req *models.RegistrarDispositivoRequest) (*models.PushDispositivo, error)

	ActualizarUltimaVista(ctx context.Context, dispositivoId int64) error

	ActualizarEstadoDispositivo(ctx context.Context, dispositivoId int64, enabled bool) error

	ActualizarTopicsDispositivo(ctx context.Context, dispositivoId int64, topics []string) error

	RegistrarEnvio(ctx context.Context, req *models.RegistrarEnvioRequest) (*models.PushEnvio, error)

	EnviarNotificacion(req *models.EnviarNotificacionRequest) (*models.EnviarNotificacionResponse, error)

	ValidarRegistroDispositivo(req *models.RegistrarDispositivoRequest) error
}

var _ PushServiceInterface = (*PushService)(nil)
