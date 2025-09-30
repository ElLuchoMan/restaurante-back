package push

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"restaurante/models"

	"github.com/beego/beego/v2/server/web"
)

func TestPushController_Integration(t *testing.T) {
	// Configurar Beego para tests
	web.BConfig.RunMode = "test"

	t.Run("RegistrarDispositivo - Web válido", func(t *testing.T) {
		request := &models.RegistrarDispositivoRequest{
			Plataforma:            models.PlataformaWeb,
			Endpoint:              stringPtr("https://fcm.googleapis.com/fcm/send/test"),
			P256dh:                stringPtr("test_p256dh_key"),
			Auth:                  stringPtr("test_auth_key"),
			Locale:                stringPtr("es-CO"),
			TimeZone:              stringPtr("America/Bogota"),
			AppVersion:            stringPtr("1.0.0"),
			UserAgent:             stringPtr("Mozilla/5.0 (Windows NT 10.0; Win64; x64)"),
			SubscribedTopics:      []string{"general", "ofertas"},
			PkDocumentoCliente:    int64Ptr(12345678),
			PkDocumentoTrabajador: nil,
		}

		jsonData, err := json.Marshal(request)
		if err != nil {
			t.Fatalf("Error marshaling request: %v", err)
		}

		// Crear request HTTP
		req := httptest.NewRequest("POST", "/restaurante/v1/push/dispositivos", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test_token")

		// Verificar que el request se puede parsear
		var parsedRequest models.RegistrarDispositivoRequest
		err = json.Unmarshal(jsonData, &parsedRequest)
		if err != nil {
			t.Errorf("Error parsing request: %v", err)
		}

		// Verificar campos específicos de web
		if parsedRequest.Plataforma != models.PlataformaWeb {
			t.Errorf("Expected plataforma %s, got %s", models.PlataformaWeb, parsedRequest.Plataforma)
		}

		if parsedRequest.Endpoint == nil || *parsedRequest.Endpoint != *request.Endpoint {
			t.Errorf("Expected endpoint %v, got %v", request.Endpoint, parsedRequest.Endpoint)
		}

		if parsedRequest.FcmToken != nil {
			t.Error("Web device should not have FCM token")
		}

		if len(parsedRequest.SubscribedTopics) != 2 {
			t.Errorf("Expected 2 subscribed topics, got %d", len(parsedRequest.SubscribedTopics))
		}
	})

	t.Run("RegistrarDispositivo - Android válido", func(t *testing.T) {
		request := &models.RegistrarDispositivoRequest{
			Plataforma:            models.PlataformaAndroid,
			FcmToken:              stringPtr("android_fcm_token_123"),
			Locale:                stringPtr("es-CO"),
			TimeZone:              stringPtr("America/Bogota"),
			AppVersion:            stringPtr("2.1.0"),
			SubscribedTopics:      []string{"pedidos", "promociones"},
			PkDocumentoTrabajador: int64Ptr(87654321),
			PkDocumentoCliente:    nil,
		}

		jsonData, err := json.Marshal(request)
		if err != nil {
			t.Fatalf("Error marshaling request: %v", err)
		}

		// Verificar que el request se puede parsear
		var parsedRequest models.RegistrarDispositivoRequest
		err = json.Unmarshal(jsonData, &parsedRequest)
		if err != nil {
			t.Errorf("Error parsing request: %v", err)
		}

		// Verificar campos específicos de Android
		if parsedRequest.Plataforma != models.PlataformaAndroid {
			t.Errorf("Expected plataforma %s, got %s", models.PlataformaAndroid, parsedRequest.Plataforma)
		}

		if parsedRequest.FcmToken == nil || *parsedRequest.FcmToken != *request.FcmToken {
			t.Errorf("Expected FCM token %v, got %v", request.FcmToken, parsedRequest.FcmToken)
		}

		if parsedRequest.Endpoint != nil {
			t.Error("Android device should not have endpoint")
		}

		if parsedRequest.PkDocumentoTrabajador == nil || *parsedRequest.PkDocumentoTrabajador != *request.PkDocumentoTrabajador {
			t.Errorf("Expected trabajador ID %v, got %v", request.PkDocumentoTrabajador, parsedRequest.PkDocumentoTrabajador)
		}
	})

	t.Run("ActualizarEstadoDispositivo - Request válido", func(t *testing.T) {
		request := &models.ActualizarEstadoDispositivoRequest{
			Enabled: false,
		}

		jsonData, err := json.Marshal(request)
		if err != nil {
			t.Fatalf("Error marshaling request: %v", err)
		}

		// Verificar que el request se puede parsear
		var parsedRequest models.ActualizarEstadoDispositivoRequest
		err = json.Unmarshal(jsonData, &parsedRequest)
		if err != nil {
			t.Errorf("Error parsing request: %v", err)
		}

		if parsedRequest.Enabled != request.Enabled {
			t.Errorf("Expected enabled %v, got %v", request.Enabled, parsedRequest.Enabled)
		}
	})

	t.Run("ActualizarTopics - Request válido", func(t *testing.T) {
		request := &models.ActualizarTopicsRequest{
			SubscribedTopics: []string{"general", "ofertas", "pedidos"},
		}

		jsonData, err := json.Marshal(request)
		if err != nil {
			t.Fatalf("Error marshaling request: %v", err)
		}

		// Verificar que el request se puede parsear
		var parsedRequest models.ActualizarTopicsRequest
		err = json.Unmarshal(jsonData, &parsedRequest)
		if err != nil {
			t.Errorf("Error parsing request: %v", err)
		}

		if len(parsedRequest.SubscribedTopics) != 3 {
			t.Errorf("Expected 3 topics, got %d", len(parsedRequest.SubscribedTopics))
		}

		expectedTopics := []string{"general", "ofertas", "pedidos"}
		for i, topic := range parsedRequest.SubscribedTopics {
			if topic != expectedTopics[i] {
				t.Errorf("Expected topic %s at index %d, got %s", expectedTopics[i], i, topic)
			}
		}
	})

	t.Run("RegistrarEnvio - Request válido", func(t *testing.T) {
		data := map[string]interface{}{
			"title": "Nuevo pedido",
			"body":  "Tienes un nuevo pedido #123",
			"icon":  "/icon-192x192.png",
		}
		dataBytes, _ := json.Marshal(data)

		request := &models.RegistrarEnvioRequest{
			PkIdPushDispositivo: 1,
			Proveedor:           models.ProveedorWebPush,
			Data:                dataBytes,
			Exito:               true,
			StatusCode:          intPtr(200),
		}

		jsonData, err := json.Marshal(request)
		if err != nil {
			t.Fatalf("Error marshaling request: %v", err)
		}

		// Verificar que el request se puede parsear
		var parsedRequest models.RegistrarEnvioRequest
		err = json.Unmarshal(jsonData, &parsedRequest)
		if err != nil {
			t.Errorf("Error parsing request: %v", err)
		}

		if parsedRequest.PkIdPushDispositivo != request.PkIdPushDispositivo {
			t.Errorf("Expected dispositivo ID %d, got %d", request.PkIdPushDispositivo, parsedRequest.PkIdPushDispositivo)
		}

		if parsedRequest.Proveedor != request.Proveedor {
			t.Errorf("Expected proveedor %s, got %s", request.Proveedor, parsedRequest.Proveedor)
		}

		if parsedRequest.Exito != request.Exito {
			t.Errorf("Expected exito %v, got %v", request.Exito, parsedRequest.Exito)
		}

		// Verificar que los datos JSON se preservan
		var parsedData map[string]interface{}
		err = json.Unmarshal(parsedRequest.Data, &parsedData)
		if err != nil {
			t.Errorf("Error parsing data JSON: %v", err)
		}

		if parsedData["title"] != "Nuevo pedido" {
			t.Errorf("Expected title 'Nuevo pedido', got %v", parsedData["title"])
		}
	})
}

func TestPushController_ResponseStructures(t *testing.T) {
	t.Run("PushDispositivoResponse - Serialización", func(t *testing.T) {
		now := time.Now()
		response := &models.PushDispositivoResponse{
			PushDispositivoId:   1,
			Plataforma:          models.PlataformaWeb,
			Endpoint:            stringPtr("https://fcm.googleapis.com/fcm/send/test"),
			Enabled:             true,
			Locale:              stringPtr("es-CO"),
			TimeZone:            stringPtr("America/Bogota"),
			AppVersion:          stringPtr("1.0.0"),
			SubscribedTopics:    []string{"general", "ofertas"},
			DocumentoCliente:    int64Ptr(12345678),
			DocumentoTrabajador: nil,
			CreatedAt:           now,
			LastSeenAt:          &now,
		}

		jsonData, err := json.Marshal(response)
		if err != nil {
			t.Fatalf("Error marshaling response: %v", err)
		}

		// Verificar que se puede deserializar
		var parsedResponse models.PushDispositivoResponse
		err = json.Unmarshal(jsonData, &parsedResponse)
		if err != nil {
			t.Errorf("Error parsing response: %v", err)
		}

		// Verificar campos clave
		if parsedResponse.PushDispositivoId != response.PushDispositivoId {
			t.Errorf("Expected dispositivo ID %d, got %d", response.PushDispositivoId, parsedResponse.PushDispositivoId)
		}

		if parsedResponse.Plataforma != response.Plataforma {
			t.Errorf("Expected plataforma %s, got %s", response.Plataforma, parsedResponse.Plataforma)
		}

		if len(parsedResponse.SubscribedTopics) != len(response.SubscribedTopics) {
			t.Errorf("Expected %d topics, got %d", len(response.SubscribedTopics), len(parsedResponse.SubscribedTopics))
		}
	})

	t.Run("PushEnvioResponse - Serialización", func(t *testing.T) {
		data := map[string]interface{}{
			"title": "Test notification",
			"body":  "Test message",
		}
		dataBytes, _ := json.Marshal(data)

		response := &models.PushEnvioResponse{
			PushEnvioId:       1,
			PushDispositivoId: 1,
			Proveedor:         models.ProveedorFCM,
			Data:              dataBytes,
			Exito:             true,
			StatusCode:        intPtr(200),
			SentAt:            time.Now(),
		}

		jsonData, err := json.Marshal(response)
		if err != nil {
			t.Fatalf("Error marshaling response: %v", err)
		}

		// Verificar que se puede deserializar
		var parsedResponse models.PushEnvioResponse
		err = json.Unmarshal(jsonData, &parsedResponse)
		if err != nil {
			t.Errorf("Error parsing response: %v", err)
		}

		// Verificar campos clave
		if parsedResponse.PushEnvioId != response.PushEnvioId {
			t.Errorf("Expected envio ID %d, got %d", response.PushEnvioId, parsedResponse.PushEnvioId)
		}

		if parsedResponse.Proveedor != response.Proveedor {
			t.Errorf("Expected proveedor %s, got %s", response.Proveedor, parsedResponse.Proveedor)
		}

		// Verificar que los datos JSON se preservan
		var parsedData map[string]interface{}
		err = json.Unmarshal(parsedResponse.Data, &parsedData)
		if err != nil {
			t.Errorf("Error parsing data JSON: %v", err)
		}
	})
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func int64Ptr(i int64) *int64 {
	return &i
}

func intPtr(i int) *int {
	return &i
}
