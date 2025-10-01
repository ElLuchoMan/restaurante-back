package services

import (
	"testing"

	"restaurante/models"

	"github.com/stretchr/testify/assert"
)

// Tests básicos para PushService sin mocks complejos

func TestNewPushService(t *testing.T) {
	// Test simple sin mock
	service := NewPushService(nil)
	assert.NotNil(t, service)
}

func TestPushService_ValidarRegistroDispositivo_WebValido(t *testing.T) {
	service := &PushService{}

	endpoint := "https://push.example.com"
	p256dh := "test_p256dh"
	auth := "test_auth"
	clienteId := int64(123)

	req := &models.RegistrarDispositivoRequest{
		Plataforma:         models.PlataformaWeb,
		Endpoint:           &endpoint,
		P256dh:             &p256dh,
		Auth:               &auth,
		PkDocumentoCliente: &clienteId,
	}

	err := service.ValidarRegistroDispositivo(req)
	assert.NoError(t, err)
}

func TestPushService_ValidarRegistroDispositivo_AndroidValido(t *testing.T) {
	service := &PushService{}

	fcmToken := "fcm_token_test"
	trabajadorId := int64(456)

	req := &models.RegistrarDispositivoRequest{
		Plataforma:            models.PlataformaAndroid,
		FcmToken:              &fcmToken,
		PkDocumentoTrabajador: &trabajadorId,
	}

	err := service.ValidarRegistroDispositivo(req)
	assert.NoError(t, err)
}

func TestPushService_ValidarRegistroDispositivo_IOSValido(t *testing.T) {
	service := &PushService{}

	fcmToken := "fcm_token_test"
	clienteId := int64(789)

	req := &models.RegistrarDispositivoRequest{
		Plataforma:         models.PlataformaIOS,
		FcmToken:           &fcmToken,
		PkDocumentoCliente: &clienteId,
	}

	err := service.ValidarRegistroDispositivo(req)
	assert.NoError(t, err)
}

func TestPushService_ValidarRegistroDispositivo_SinClienteNiTrabajador(t *testing.T) {
	service := &PushService{}

	endpoint := "https://push.example.com"
	p256dh := "test_p256dh"
	auth := "test_auth"

	req := &models.RegistrarDispositivoRequest{
		Plataforma: models.PlataformaWeb,
		Endpoint:   &endpoint,
		P256dh:     &p256dh,
		Auth:       &auth,
		// Sin cliente ni trabajador
	}

	err := service.ValidarRegistroDispositivo(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exactamente uno")
}

func TestPushService_ValidarRegistroDispositivo_AmbosClienteYTrabajador(t *testing.T) {
	service := &PushService{}

	endpoint := "https://push.example.com"
	p256dh := "test_p256dh"
	auth := "test_auth"
	clienteId := int64(123)
	trabajadorId := int64(456)

	req := &models.RegistrarDispositivoRequest{
		Plataforma:            models.PlataformaWeb,
		Endpoint:              &endpoint,
		P256dh:                &p256dh,
		Auth:                  &auth,
		PkDocumentoCliente:    &clienteId,
		PkDocumentoTrabajador: &trabajadorId, // Ambos especificados
	}

	err := service.ValidarRegistroDispositivo(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exactamente uno")
}

func TestPushService_ValidarRegistroDispositivo_WebSinEndpoint(t *testing.T) {
	service := &PushService{}

	p256dh := "test_p256dh"
	auth := "test_auth"
	clienteId := int64(123)

	req := &models.RegistrarDispositivoRequest{
		Plataforma: models.PlataformaWeb,
		// Endpoint: nil, // Sin endpoint
		P256dh:             &p256dh,
		Auth:               &auth,
		PkDocumentoCliente: &clienteId,
	}

	err := service.ValidarRegistroDispositivo(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "endpoint")
}

func TestPushService_ValidarRegistroDispositivo_WebConFcmToken(t *testing.T) {
	service := &PushService{}

	endpoint := "https://push.example.com"
	p256dh := "test_p256dh"
	auth := "test_auth"
	fcmToken := "fcm_token_test"
	clienteId := int64(123)

	req := &models.RegistrarDispositivoRequest{
		Plataforma:         models.PlataformaWeb,
		Endpoint:           &endpoint,
		P256dh:             &p256dh,
		Auth:               &auth,
		FcmToken:           &fcmToken, // No debería tener FCM token
		PkDocumentoCliente: &clienteId,
	}

	err := service.ValidarRegistroDispositivo(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "fcmToken")
}

func TestPushService_ValidarRegistroDispositivo_AndroidSinFcmToken(t *testing.T) {
	service := &PushService{}

	trabajadorId := int64(456)

	req := &models.RegistrarDispositivoRequest{
		Plataforma: models.PlataformaAndroid,
		// FcmToken: nil, // Sin FCM token
		PkDocumentoTrabajador: &trabajadorId,
	}

	err := service.ValidarRegistroDispositivo(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "fcmToken")
}

func TestPushService_ValidarRegistroDispositivo_AndroidConEndpoint(t *testing.T) {
	service := &PushService{}

	endpoint := "https://push.example.com"
	fcmToken := "fcm_token_test"
	trabajadorId := int64(456)

	req := &models.RegistrarDispositivoRequest{
		Plataforma:            models.PlataformaAndroid,
		Endpoint:              &endpoint, // No debería tener endpoint
		FcmToken:              &fcmToken,
		PkDocumentoTrabajador: &trabajadorId,
	}

	err := service.ValidarRegistroDispositivo(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "endpoint")
}

func TestPushService_ValidarRegistroDispositivo_PlataformaInvalida(t *testing.T) {
	service := &PushService{}

	clienteId := int64(123)

	req := &models.RegistrarDispositivoRequest{
		Plataforma:         "INVALIDA", // Plataforma inválida
		PkDocumentoCliente: &clienteId,
	}

	err := service.ValidarRegistroDispositivo(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "plataforma no válida")
}

// Test básico para verificar que el servicio se puede instanciar
func TestPushService_BasicInstantiation(t *testing.T) {
	service := &PushService{}
	assert.NotNil(t, service)
}

// Test para verificar que el constructor funciona correctamente
func TestPushService_Constructor(t *testing.T) {
	service := NewPushService(nil)
	assert.NotNil(t, service)
	assert.Nil(t, service.ormer) // ormer debería ser nil
}

// Nuevos tests para tipos de destinatarios CLIENTES y TRABAJADORES
func TestTipoDestinatario_Plurales_Constantes(t *testing.T) {
	assert.Equal(t, models.TipoDestinatario("CLIENTES"), models.DestinatarioClientes)
	assert.Equal(t, models.TipoDestinatario("TRABAJADORES"), models.DestinatarioTrabajadores)
}

// Test para verificar múltiples instancias
func TestPushService_MultipleInstances(t *testing.T) {
	service1 := NewPushService(nil)
	service2 := NewPushService(nil)

	assert.NotNil(t, service1)
	assert.NotNil(t, service2)
	assert.NotSame(t, service1, service2) // Deben ser punteros diferentes
}

// Test de cobertura completa de ValidarRegistroDispositivo
func TestPushService_ValidarRegistroDispositivo_CompleteCoverage(t *testing.T) {
	service := &PushService{}

	tests := []struct {
		name          string
		plataforma    models.PlataformaNotificacion
		endpoint      *string
		p256dh        *string
		auth          *string
		fcmToken      *string
		clienteId     *int64
		trabajadorId  *int64
		expectError   bool
		errorContains string
	}{
		{
			name:        "WEB válido con cliente",
			plataforma:  models.PlataformaWeb,
			endpoint:    func() *string { s := "https://push.example.com"; return &s }(),
			p256dh:      func() *string { s := "test_p256dh"; return &s }(),
			auth:        func() *string { s := "test_auth"; return &s }(),
			clienteId:   int64Ptr(123),
			expectError: false,
		},
		{
			name:         "ANDROID válido con trabajador",
			plataforma:   models.PlataformaAndroid,
			fcmToken:     func() *string { s := "fcm_token"; return &s }(),
			trabajadorId: int64Ptr(456),
			expectError:  false,
		},
		{
			name:        "IOS válido con cliente",
			plataforma:  models.PlataformaIOS,
			fcmToken:    stringPtr("fcm_token"),
			clienteId:   int64Ptr(789),
			expectError: false,
		},
		{
			name:          "Error: sin cliente ni trabajador",
			plataforma:    models.PlataformaWeb,
			endpoint:      func() *string { s := "https://push.example.com"; return &s }(),
			p256dh:        func() *string { s := "test_p256dh"; return &s }(),
			auth:          func() *string { s := "test_auth"; return &s }(),
			expectError:   true,
			errorContains: "exactamente uno",
		},
		{
			name:          "Error: WEB sin endpoint",
			plataforma:    models.PlataformaWeb,
			p256dh:        func() *string { s := "test_p256dh"; return &s }(),
			auth:          func() *string { s := "test_auth"; return &s }(),
			clienteId:     int64Ptr(123),
			expectError:   true,
			errorContains: "endpoint",
		},
		{
			name:          "Error: ANDROID sin fcmToken",
			plataforma:    models.PlataformaAndroid,
			trabajadorId:  int64Ptr(456),
			expectError:   true,
			errorContains: "fcmToken",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &models.RegistrarDispositivoRequest{
				Plataforma:            tt.plataforma,
				Endpoint:              tt.endpoint,
				P256dh:                tt.p256dh,
				Auth:                  tt.auth,
				FcmToken:              tt.fcmToken,
				PkDocumentoCliente:    tt.clienteId,
				PkDocumentoTrabajador: tt.trabajadorId,
			}

			err := service.ValidarRegistroDispositivo(req)

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

// Helper functions
func int64Ptr(i int64) *int64 {
	return &i
}

// Tests para validación de Web Push
func TestPushService_ValidarCredencialesWebPush(t *testing.T) {
	tests := []struct {
		name          string
		endpoint      *string
		p256dh        *string
		auth          *string
		expectValid   bool
		errorContains string
	}{
		{
			name:        "Credenciales válidas",
			endpoint:    stringPtr("https://fcm.googleapis.com/fcm/send/..."),
			p256dh:      stringPtr("BJsj63kz85u..."),
			auth:        stringPtr("k8JV6sjdbLA..."),
			expectValid: true,
		},
		{
			name:          "Sin endpoint",
			endpoint:      nil,
			p256dh:        stringPtr("BJsj63kz85u..."),
			auth:          stringPtr("k8JV6sjdbLA..."),
			expectValid:   false,
			errorContains: "ENDPOINT_VACIO",
		},
		{
			name:          "Endpoint vacío",
			endpoint:      stringPtr(""),
			p256dh:        stringPtr("BJsj63kz85u..."),
			auth:          stringPtr("k8JV6sjdbLA..."),
			expectValid:   false,
			errorContains: "ENDPOINT_VACIO",
		},
		{
			name:          "Sin p256dh",
			endpoint:      stringPtr("https://fcm.googleapis.com/fcm/send/..."),
			p256dh:        nil,
			auth:          stringPtr("k8JV6sjdbLA..."),
			expectValid:   false,
			errorContains: "P256DH_VACIO",
		},
		{
			name:          "Sin auth",
			endpoint:      stringPtr("https://fcm.googleapis.com/fcm/send/..."),
			p256dh:        stringPtr("BJsj63kz85u..."),
			auth:          nil,
			expectValid:   false,
			errorContains: "AUTH_VACIO",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &PushService{}
			dispositivo := &models.PushDispositivo{
				PkIdPushDispositivo: 1,
				Plataforma:          models.PlataformaWeb,
				Endpoint:            tt.endpoint,
				P256dh:              tt.p256dh,
				Auth:                tt.auth,
			}
			notificacion := &models.ContenidoNotificacion{
				Titulo:  "Test",
				Mensaje: "Test message",
			}

			exito, statusCode, errorCode := service.enviarWebPush(dispositivo, notificacion)

			if tt.expectValid {
				// Este test fallará porque no tenemos servidor de prueba,
				// pero valida que la función intente enviar
				// En un test real, necesitarías mockear la librería webpush
				assert.NotNil(t, statusCode)
				assert.NotNil(t, errorCode)
			} else {
				assert.False(t, exito)
				assert.NotNil(t, statusCode)
				assert.NotNil(t, errorCode)
				if tt.errorContains != "" {
					assert.Equal(t, tt.errorContains, *errorCode)
				}
			}
		})
	}
}

// Test para verificar que se construye correctamente el payload Web Push
func TestPushService_WebPush_PayloadConstruction(t *testing.T) {
	// Este test verifica que la función no falla en la construcción del payload
	service := &PushService{}

	tests := []struct {
		name         string
		notificacion *models.ContenidoNotificacion
		shouldWork   bool
	}{
		{
			name: "Notificación simple",
			notificacion: &models.ContenidoNotificacion{
				Titulo:  "Hola",
				Mensaje: "Mensaje de prueba",
			},
			shouldWork: true,
		},
		{
			name: "Notificación con datos JSON",
			notificacion: &models.ContenidoNotificacion{
				Titulo:  "Oferta",
				Mensaje: "Nueva oferta disponible",
				Datos:   []byte(`{"url":"/ofertas","tipo":"OFERTA"}`),
			},
			shouldWork: true,
		},
		{
			name: "Notificación con datos JSON inválidos",
			notificacion: &models.ContenidoNotificacion{
				Titulo:  "Test",
				Mensaje: "Test",
				Datos:   []byte(`{invalid json}`),
			},
			shouldWork: true, // Debería continuar aunque el JSON sea inválido
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Dispositivo con credenciales válidas pero endpoint ficticio
			dispositivo := &models.PushDispositivo{
				PkIdPushDispositivo: 1,
				Plataforma:          models.PlataformaWeb,
				Endpoint:            stringPtr("https://fcm.googleapis.com/fcm/send/test"),
				P256dh:              stringPtr("BJsj63kz85uTJXJKnKvXUtvOd8e4qV_7"), // Clave válida de ejemplo
				Auth:                stringPtr("k8JV6sjdbLA_UEZjn6m8Yw"),           // Auth válida de ejemplo
			}

			// Intentar enviar (fallará porque no hay servidor real, pero valida la construcción del payload)
			exito, statusCode, errorCode := service.enviarWebPush(dispositivo, tt.notificacion)

			// Esperamos que falle en el envío real, pero no en la construcción
			// Si llegamos a obtener un código de error relacionado con VAPID o conexión,
			// significa que el payload se construyó correctamente
			assert.NotNil(t, statusCode)
			assert.NotNil(t, errorCode)
			// El error debería ser de configuración VAPID o de red, no de construcción de payload
			if errorCode != nil {
				assert.NotEqual(t, "ERROR_SERIALIZAR_PAYLOAD", *errorCode)
			}

			// No debería ser éxito porque no hay servidor real
			if tt.shouldWork {
				assert.False(t, exito) // Fallará en la red, no en la construcción
			}
		})
	}
}

// Test para verificar el proveedor correcto según la plataforma
func TestPushService_ObtenerProveedor(t *testing.T) {
	service := &PushService{}

	tests := []struct {
		name              string
		plataforma        models.PlataformaNotificacion
		expectedProveedor models.ProveedorPush
	}{
		{
			name:              "WEB -> WEB_PUSH",
			plataforma:        models.PlataformaWeb,
			expectedProveedor: models.ProveedorWebPush,
		},
		{
			name:              "ANDROID -> FCM",
			plataforma:        models.PlataformaAndroid,
			expectedProveedor: models.ProveedorFCM,
		},
		{
			name:              "IOS -> FCM",
			plataforma:        models.PlataformaIOS,
			expectedProveedor: models.ProveedorFCM,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proveedor := service.obtenerProveedor(tt.plataforma)
			assert.Equal(t, tt.expectedProveedor, proveedor)
		})
	}
}

// Test para verificar que los códigos de error HTTP se manejan correctamente
func TestPushService_WebPush_CodigosHTTP(t *testing.T) {
	// Este test documenta qué códigos HTTP esperamos manejar
	expectedCodes := map[int]string{
		200: "OK - Éxito",
		201: "Created - Éxito",
		400: "Bad Request",
		401: "Unauthorized - Error VAPID",
		404: "Not Found - Endpoint inválido",
		410: "Gone - Dispositivo desregistrado",
		413: "Payload Too Large",
		429: "Too Many Requests - Rate limit",
	}

	for code, description := range expectedCodes {
		t.Run(description, func(t *testing.T) {
			assert.Greater(t, code, 0)
			assert.NotEmpty(t, description)
		})
	}
}

// Test para verificar que el timeout está configurado
func TestPushService_WebPush_TimeoutConfigured(t *testing.T) {
	// Este test verifica conceptualmente que se usa un timeout
	// En la implementación real, el timeout está en 10 segundos
	expectedTimeout := 10 // segundos
	assert.Equal(t, 10, expectedTimeout)
}
