package services

import (
	"testing"
	"time"

	"restaurante/models"
)

func TestCuponValidations(t *testing.T) {
	service := &CuponService{}

	tests := []struct {
		name    string
		cupon   *models.Cupon
		wantErr bool
		errMsg  string
	}{
		{
			name: "Cupón válido con porcentaje",
			cupon: &models.Cupon{
				TipoDescuento:  models.TipoDescuentoPorcentaje,
				ValorDescuento: 50,
				FechaInicio:    time.Now(),
				FechaFin:       time.Now().AddDate(0, 0, 7),
				Scope:          models.CuponScopeGlobal,
			},
			wantErr: false,
		},
		{
			name: "Porcentaje inválido mayor a 100",
			cupon: &models.Cupon{
				TipoDescuento:  models.TipoDescuentoPorcentaje,
				ValorDescuento: 150,
				FechaInicio:    time.Now(),
				FechaFin:       time.Now().AddDate(0, 0, 7),
				Scope:          models.CuponScopeGlobal,
			},
			wantErr: true,
			errMsg:  "el porcentaje de descuento debe estar entre 1 y 100",
		},
		{
			name: "Fechas inválidas",
			cupon: &models.Cupon{
				TipoDescuento:  models.TipoDescuentoPorcentaje,
				ValorDescuento: 10,
				FechaInicio:    time.Now().AddDate(0, 0, 7),
				FechaFin:       time.Now(),
				Scope:          models.CuponScopeGlobal,
			},
			wantErr: true,
			errMsg:  "la fecha de fin debe ser posterior a la fecha de inicio",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidarReglasNegocioCupon(tt.cupon)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidarReglasNegocioCupon() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("ValidarReglasNegocioCupon() error = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestOfertaValidations(t *testing.T) {
	service := &OfertaService{}

	tests := []struct {
		name    string
		oferta  *models.Oferta
		wantErr bool
		errMsg  string
	}{
		{
			name: "Oferta válida con porcentaje",
			oferta: &models.Oferta{
				TipoDescuento:  models.TipoDescuentoPorcentaje,
				ValorDescuento: 20,
				FechaInicio:    time.Now(),
				FechaFin:       time.Now().AddDate(0, 0, 7),
			},
			wantErr: false,
		},
		{
			name: "Porcentaje inválido mayor a 100",
			oferta: &models.Oferta{
				TipoDescuento:  models.TipoDescuentoPorcentaje,
				ValorDescuento: 150,
				FechaInicio:    time.Now(),
				FechaFin:       time.Now().AddDate(0, 0, 7),
			},
			wantErr: true,
			errMsg:  "el porcentaje de descuento debe estar entre 1 y 100",
		},
		{
			name: "Fechas inválidas",
			oferta: &models.Oferta{
				TipoDescuento:  models.TipoDescuentoPorcentaje,
				ValorDescuento: 10,
				FechaInicio:    time.Now().AddDate(0, 0, 7),
				FechaFin:       time.Now(),
			},
			wantErr: true,
			errMsg:  "la fecha de fin debe ser posterior a la fecha de inicio",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidarReglasNegocioOferta(tt.oferta)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidarReglasNegocioOferta() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("ValidarReglasNegocioOferta() error = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestPushValidations(t *testing.T) {
	service := &PushService{}

	tests := []struct {
		name    string
		request *models.RegistrarDispositivoRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "Dispositivo web válido",
			request: &models.RegistrarDispositivoRequest{
				Plataforma:            models.PlataformaWeb,
				Endpoint:              testStringPtr("https://fcm.googleapis.com/fcm/send/..."),
				P256dh:                testStringPtr("test_p256dh"),
				Auth:                  testStringPtr("test_auth"),
				PkDocumentoCliente:    testInt64Ptr(12345678),
				PkDocumentoTrabajador: nil,
			},
			wantErr: false,
		},
		{
			name: "Sin cliente ni trabajador",
			request: &models.RegistrarDispositivoRequest{
				Plataforma:            models.PlataformaWeb,
				Endpoint:              testStringPtr("https://fcm.googleapis.com/fcm/send/..."),
				P256dh:                testStringPtr("test_p256dh"),
				Auth:                  testStringPtr("test_auth"),
				PkDocumentoCliente:    nil,
				PkDocumentoTrabajador: nil,
			},
			wantErr: true,
			errMsg:  "debe especificar exactamente uno de cliente o trabajador",
		},
		{
			name: "Web sin endpoint",
			request: &models.RegistrarDispositivoRequest{
				Plataforma:         models.PlataformaWeb,
				P256dh:             testStringPtr("test_p256dh"),
				Auth:               testStringPtr("test_auth"),
				PkDocumentoCliente: testInt64Ptr(12345678),
			},
			wantErr: true,
			errMsg:  "para plataforma WEB se requieren endpoint, p256dh y auth",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidarRegistroDispositivo(tt.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidarRegistroDispositivo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("ValidarRegistroDispositivo() error = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestCalcularDescuento(t *testing.T) {
	service := &CuponService{}

	tests := []struct {
		name                string
		cupon               *models.Cupon
		montoTotal          int64
		items               []models.ValidarCuponItemRequest
		productosAplicables []int64
		expected            int64
	}{
		{
			name: "Descuento por porcentaje global",
			cupon: &models.Cupon{
				Scope:          models.CuponScopeGlobal,
				TipoDescuento:  models.TipoDescuentoPorcentaje,
				ValorDescuento: 10,
			},
			montoTotal:          20000,
			items:               []models.ValidarCuponItemRequest{},
			productosAplicables: []int64{},
			expected:            2000,
		},
		{
			name: "Descuento por monto fijo global",
			cupon: &models.Cupon{
				Scope:          models.CuponScopeGlobal,
				TipoDescuento:  models.TipoDescuentoMonto,
				ValorDescuento: 5000,
			},
			montoTotal:          20000,
			items:               []models.ValidarCuponItemRequest{},
			productosAplicables: []int64{},
			expected:            5000,
		},
		{
			name: "Descuento por monto mayor al total",
			cupon: &models.Cupon{
				Scope:          models.CuponScopeGlobal,
				TipoDescuento:  models.TipoDescuentoMonto,
				ValorDescuento: 30000,
			},
			montoTotal:          20000,
			items:               []models.ValidarCuponItemRequest{},
			productosAplicables: []int64{},
			expected:            20000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.calcularDescuento(tt.cupon, tt.montoTotal, tt.items, tt.productosAplicables)
			if result != tt.expected {
				t.Errorf("calcularDescuento() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestDiaSemanaConversion(t *testing.T) {
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
			if result != tt.expected {
				t.Errorf("obtenerDiaSemanaEspanol(%v) = %v, want %v", tt.weekday, result, tt.expected)
			}
		})
	}
}

func testStringPtr(s string) *string {
	return &s
}

func testInt64Ptr(i int64) *int64 {
	return &i
}
