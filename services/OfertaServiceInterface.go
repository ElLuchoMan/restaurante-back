package services

import (
	"context"
	"time"

	"restaurante/models"
)

// OfertaServiceInterface define los métodos del servicio de ofertas
// Esta interfaz permite mockear el servicio en tests unitarios
type OfertaServiceInterface interface {
	// ObtenerOfertasActivas obtiene ofertas activas según los criterios especificados
	ObtenerOfertasActivas(ctx context.Context, restauranteId int64, fecha *time.Time, hora *time.Time, productoId *int64) ([]*models.OfertaActivaResponse, error)

	// ValidarReglasNegocioOferta valida las reglas de negocio de una oferta
	ValidarReglasNegocioOferta(oferta *models.Oferta) error

	// CalcularDescuentoOferta calcula el descuento de una oferta para items específicos
	CalcularDescuentoOferta(oferta *models.Oferta, items []models.ValidarCuponItemRequest) (int64, error)
}

// Verificar que OfertaService implementa la interfaz (compile-time check)
var _ OfertaServiceInterface = (*OfertaService)(nil)
