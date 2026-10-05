package services

import (
	"context"

	"restaurante/models"
)

type PushServiceInterface interface {
	RegistrarDispositivo(ctx context.Context, req *models.RegistrarDispositivoRequest) (dispositivo *models.PushDispositivo, created bool, err error)

	ActualizarUltimaVista(ctx context.Context, dispositivoId int64) error

	ActualizarDispositivo(ctx context.Context, dispositivoId int64, body []byte) (*models.PushDispositivo, error)

	ActualizarTopicsDispositivo(ctx context.Context, dispositivoId int64, topics []string) error

	RegistrarEnvio(ctx context.Context, req *models.RegistrarEnvioRequest) (*models.PushEnvio, error)

	EnviarNotificacion(req *models.EnviarNotificacionRequest) (*models.EnviarNotificacionResponse, error)

	ValidarRegistroDispositivo(req *models.RegistrarDispositivoRequest) error
}

var _ PushServiceInterface = (*PushService)(nil)
