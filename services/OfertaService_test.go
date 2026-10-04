package services

import (
	"testing"
	"time"

	"restaurante/models"

	"github.com/stretchr/testify/assert"
)

func TestNewOfertaService(t *testing.T) {

	service := NewOfertaService(nil)
	assert.NotNil(t, service)
}

func TestOfertaService_ValidarReglasNegocioOferta_TipoDescuentoPorcentaje(t *testing.T) {
	service := &OfertaService{}

	fechaInicio, _ := time.Parse("2006-01-02", "2025-01-01")
	fechaFin, _ := time.Parse("2006-01-02", "2025-12-31")
	restaurante := &models.Restaurante{PK_ID_RESTAURANTE: 1}

	oferta := &models.Oferta{
		TipoDescuento:   models.TipoDescuentoPorcentaje,
		ValorDescuento:  10,
		FechaInicio:     fechaInicio,
		FechaFin:        fechaFin,
		PkIdRestaurante: restaurante,
	}

	err := service.ValidarReglasNegocioOferta(oferta)
	assert.NoError(t, err)

	oferta.ValorDescuento = 150
	err = service.ValidarReglasNegocioOferta(oferta)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "porcentaje")
}

func TestOfertaService_ValidarReglasNegocioOferta_TipoDescuentoMonto(t *testing.T) {
	service := &OfertaService{}

	fechaInicio, _ := time.Parse("2006-01-02", "2025-01-01")
	fechaFin, _ := time.Parse("2006-01-02", "2025-12-31")
	restaurante := &models.Restaurante{PK_ID_RESTAURANTE: 1}

	oferta := &models.Oferta{
		TipoDescuento:   models.TipoDescuentoMonto,
		ValorDescuento:  5000,
		FechaInicio:     fechaInicio,
		FechaFin:        fechaFin,
		PkIdRestaurante: restaurante,
	}

	err := service.ValidarReglasNegocioOferta(oferta)
	assert.NoError(t, err)

	oferta.ValorDescuento = -1000
	err = service.ValidarReglasNegocioOferta(oferta)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "monto")
}

func TestOfertaService_ValidarReglasNegocioOferta_FechasInvalidas(t *testing.T) {
	service := &OfertaService{}

	fechaInicio, _ := time.Parse("2006-01-02", "2025-12-31")
	fechaFin, _ := time.Parse("2006-01-02", "2025-01-01")
	restaurante := &models.Restaurante{PK_ID_RESTAURANTE: 1}

	oferta := &models.Oferta{
		TipoDescuento:   models.TipoDescuentoPorcentaje,
		ValorDescuento:  10,
		FechaInicio:     fechaInicio,
		FechaFin:        fechaFin,
		PkIdRestaurante: restaurante,
	}

	err := service.ValidarReglasNegocioOferta(oferta)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "fecha")
}

func TestOfertaService_ObtenerDiaSemanaEspanol(t *testing.T) {
	service := &OfertaService{}

	tests := []struct {
		weekday  time.Weekday
		expected string
	}{
		{time.Monday, "Lunes"},
		{time.Tuesday, "Martes"},
		{time.Wednesday, "Miércoles"},
		{time.Thursday, "Jueves"},
		{time.Friday, "Viernes"},
		{time.Saturday, "Sábado"},
		{time.Sunday, "Domingo"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := service.obtenerDiaSemanaEspanol(tt.weekday)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestOfertaService_BasicInstantiation(t *testing.T) {
	service := &OfertaService{}
	assert.NotNil(t, service)
}

func TestOfertaService_Constructor(t *testing.T) {
	service := NewOfertaService(nil)
	assert.NotNil(t, service)
	assert.Nil(t, service.ormer)
}

func TestOfertaService_MultipleInstances(t *testing.T) {
	service1 := NewOfertaService(nil)
	service2 := NewOfertaService(nil)

	assert.NotNil(t, service1)
	assert.NotNil(t, service2)
	assert.NotSame(t, service1, service2)
}

func TestOfertaService_ValidarReglasNegocioOferta_CompleteCoverage(t *testing.T) {
	service := &OfertaService{}

	fechaInicio, _ := time.Parse("2006-01-02", "2025-01-01")
	fechaFin, _ := time.Parse("2006-01-02", "2025-12-31")
	restaurante := &models.Restaurante{PK_ID_RESTAURANTE: 1}

	tests := []struct {
		name           string
		tipoDescuento  models.TipoDescuento
		valorDescuento int64
		fechaInicio    time.Time
		fechaFin       time.Time
		restaurante    *models.Restaurante
		expectError    bool
		errorContains  string
	}{
		{
			name:           "Válido porcentaje",
			tipoDescuento:  models.TipoDescuentoPorcentaje,
			valorDescuento: 50,
			fechaInicio:    fechaInicio,
			fechaFin:       fechaFin,
			restaurante:    restaurante,
			expectError:    false,
		},
		{
			name:           "Válido monto",
			tipoDescuento:  models.TipoDescuentoMonto,
			valorDescuento: 1000,
			fechaInicio:    fechaInicio,
			fechaFin:       fechaFin,
			restaurante:    restaurante,
			expectError:    false,
		},
		{
			name:           "Error: porcentaje mayor a 100",
			tipoDescuento:  models.TipoDescuentoPorcentaje,
			valorDescuento: 150,
			fechaInicio:    fechaInicio,
			fechaFin:       fechaFin,
			restaurante:    restaurante,
			expectError:    true,
			errorContains:  "porcentaje",
		},
		{
			name:           "Error: porcentaje menor a 1",
			tipoDescuento:  models.TipoDescuentoPorcentaje,
			valorDescuento: 0,
			fechaInicio:    fechaInicio,
			fechaFin:       fechaFin,
			restaurante:    restaurante,
			expectError:    true,
			errorContains:  "porcentaje",
		},
		{
			name:           "Error: monto negativo",
			tipoDescuento:  models.TipoDescuentoMonto,
			valorDescuento: -500,
			fechaInicio:    fechaInicio,
			fechaFin:       fechaFin,
			restaurante:    restaurante,
			expectError:    true,
			errorContains:  "monto",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oferta := &models.Oferta{
				TipoDescuento:   tt.tipoDescuento,
				ValorDescuento:  tt.valorDescuento,
				FechaInicio:     tt.fechaInicio,
				FechaFin:        tt.fechaFin,
				PkIdRestaurante: tt.restaurante,
			}

			err := service.ValidarReglasNegocioOferta(oferta)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
