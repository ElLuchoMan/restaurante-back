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

func TestOfertaService_MotivoNoVigente(t *testing.T) {
	s := NewOfertaService(nil)
	// Martes 2026-10-06 10:30 (hora de Bogotá).
	ahora := time.Date(2026, 10, 6, 10, 30, 0, 0, time.UTC)
	hora := func(h, m int) *time.Time { v := time.Date(0, 1, 1, h, m, 0, 0, time.UTC); return &v }
	base := func() *models.Oferta {
		return &models.Oferta{
			Activo:      true,
			FechaInicio: time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
			FechaFin:    time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC),
		}
	}

	assert.Equal(t, "", s.MotivoNoVigente(base(), ahora))

	o := base()
	o.Activo = false
	assert.Contains(t, s.MotivoNoVigente(o, ahora), "inactiva")

	o = base()
	o.FechaFin = time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC)
	assert.Contains(t, s.MotivoNoVigente(o, ahora), "período")
	o = base()
	o.FechaInicio = time.Date(2026, 10, 7, 0, 0, 0, 0, time.UTC)
	assert.Contains(t, s.MotivoNoVigente(o, ahora), "período")

	o = base()
	o.DiasSemanaArray = []string{"Lunes", " Miércoles "}
	assert.Contains(t, s.MotivoNoVigente(o, ahora), "día")
	o.DiasSemanaArray = []string{"Lunes", " Martes "}
	assert.Equal(t, "", s.MotivoNoVigente(o, ahora))

	o = base()
	o.HoraInicio, o.HoraFin = hora(11, 0), hora(14, 0)
	assert.Contains(t, s.MotivoNoVigente(o, ahora), "horario")
	o.HoraInicio, o.HoraFin = hora(10, 0), hora(10, 30)
	assert.Equal(t, "", s.MotivoNoVigente(o, ahora), "los extremos están incluidos")
	o.HoraInicio, o.HoraFin = hora(8, 0), nil
	assert.Equal(t, "", s.MotivoNoVigente(o, ahora), "sin ambos extremos no hay horario")
}
