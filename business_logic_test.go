package main

import (
	"testing"
	"time"

	"restaurante/models"
)

func TestCuponValidations(t *testing.T) {
	tests := []struct {
		name          string
		tipoDescuento models.TipoDescuento
		valor         int64
		fechaInicio   time.Time
		fechaFin      time.Time
		scope         models.CuponScope
		wantErr       bool
		errContains   string
	}{
		{
			name:          "Porcentaje válido",
			tipoDescuento: models.TipoDescuentoPorcentaje,
			valor:         50,
			fechaInicio:   time.Now(),
			fechaFin:      time.Now().AddDate(0, 0, 7),
			scope:         models.CuponScopeGlobal,
			wantErr:       false,
		},
		{
			name:          "Porcentaje inválido mayor a 100",
			tipoDescuento: models.TipoDescuentoPorcentaje,
			valor:         150,
			fechaInicio:   time.Now(),
			fechaFin:      time.Now().AddDate(0, 0, 7),
			scope:         models.CuponScopeGlobal,
			wantErr:       true,
			errContains:   "porcentaje de descuento debe estar entre 1 y 100",
		},
		{
			name:          "Fechas inválidas",
			tipoDescuento: models.TipoDescuentoPorcentaje,
			valor:         10,
			fechaInicio:   time.Now().AddDate(0, 0, 7),
			fechaFin:      time.Now(),
			scope:         models.CuponScopeGlobal,
			wantErr:       true,
			errContains:   "fecha de fin debe ser posterior",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.tipoDescuento == models.TipoDescuentoPorcentaje {
				if tt.valor < 1 || tt.valor > 100 {
					if !tt.wantErr {
						t.Errorf("Expected error for invalid percentage %d", tt.valor)
					}
					return
				}
			} else if tt.tipoDescuento == models.TipoDescuentoMonto {
				if tt.valor < 0 {
					if !tt.wantErr {
						t.Errorf("Expected error for negative amount %d", tt.valor)
					}
					return
				}
			}

			if tt.fechaFin.Before(tt.fechaInicio) {
				if !tt.wantErr {
					t.Error("Expected error for invalid dates")
				}
				return
			}

			if tt.wantErr {
				t.Error("Expected validation error but none occurred")
			}
		})
	}
}

func TestPushDeviceValidations(t *testing.T) {
	tests := []struct {
		name       string
		plataforma models.PlataformaNotificacion
		endpoint   *string
		p256dh     *string
		auth       *string
		fcmToken   *string
		cliente    *int64
		trabajador *int64
		wantErr    bool
	}{
		{
			name:       "Dispositivo WEB válido",
			plataforma: models.PlataformaWeb,
			endpoint:   stringPtr("https://fcm.googleapis.com/fcm/send/endpoint"),
			p256dh:     stringPtr("p256dh_key"),
			auth:       stringPtr("auth_key"),
			cliente:    int64Ptr(1),
			wantErr:    false,
		},
		{
			name:       "Dispositivo Android válido",
			plataforma: models.PlataformaAndroid,
			fcmToken:   stringPtr("fcm_token"),
			cliente:    int64Ptr(1),
			wantErr:    false,
		},
		{
			name:       "Sin cliente ni trabajador",
			plataforma: models.PlataformaWeb,
			endpoint:   stringPtr("https://fcm.googleapis.com/fcm/send/endpoint"),
			p256dh:     stringPtr("p256dh_key"),
			auth:       stringPtr("auth_key"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if (tt.cliente == nil && tt.trabajador == nil) || (tt.cliente != nil && tt.trabajador != nil) {
				if !tt.wantErr {
					t.Error("Expected error for invalid client/worker specification")
				}
				return
			}

			switch tt.plataforma {
			case models.PlataformaWeb:
				if tt.endpoint == nil || tt.p256dh == nil || tt.auth == nil {
					if !tt.wantErr {
						t.Error("Expected error for missing WEB platform fields")
					}
					return
				}
				if tt.fcmToken != nil {
					if !tt.wantErr {
						t.Error("Expected error for FCM token in WEB platform")
					}
					return
				}
			case models.PlataformaAndroid, models.PlataformaIOS:
				if tt.fcmToken == nil {
					if !tt.wantErr {
						t.Error("Expected error for missing FCM token")
					}
					return
				}
				if tt.endpoint != nil || tt.p256dh != nil || tt.auth != nil {
					if !tt.wantErr {
						t.Error("Expected error for WEB fields in mobile platform")
					}
					return
				}
			}

			if tt.wantErr {
				t.Error("Expected validation error but none occurred")
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}

func int64Ptr(i int64) *int64 {
	return &i
}
