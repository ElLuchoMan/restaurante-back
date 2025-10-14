package services

import (
	"context"
	"time"

	"restaurante/models"
)

type OfertaServiceInterface interface {
	ObtenerOfertasActivas(ctx context.Context, restauranteId int64, fecha *time.Time, hora *time.Time, productoId *int64) ([]*models.OfertaActivaResponse, error)

	ValidarReglasNegocioOferta(oferta *models.Oferta) error

	CalcularDescuentoOferta(oferta *models.Oferta, items []models.ValidarCuponItemRequest) (int64, error)
}

var _ OfertaServiceInterface = (*OfertaService)(nil)
