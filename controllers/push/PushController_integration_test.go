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

		req := httptest.NewRequest("POST", "/restaurante/v1/push/dispositivos", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test_token")

		var parsedRequest models.RegistrarDispositivoRequest
		err = json.Unmarshal(jsonData, &parsedRequest)
		if err != nil {
			t.Errorf("Error parsing request: %v", err)
		}

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

		var parsedRequest models.RegistrarDispositivoRequest
		err = json.Unmarshal(jsonData, &parsedRequest)
		if err != nil {
			t.Errorf("Error parsing request: %v", err)
		}

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

		var parsedResponse map[string]interface{}
		err = json.Unmarshal(jsonData, &parsedResponse)
		if err != nil {
			t.Errorf("Error parsing response: %v", err)
		}

		if int64(parsedResponse["pushDispositivoId"].(float64)) != response.PushDispositivoId {
			t.Errorf("Expected dispositivo ID %d, got %v", response.PushDispositivoId, parsedResponse["pushDispositivoId"])
		}

		if parsedResponse["plataforma"].(string) != string(response.Plataforma) {
			t.Errorf("Expected plataforma %s, got %v", response.Plataforma, parsedResponse["plataforma"])
		}

		topics := parsedResponse["subscribedTopics"].([]interface{})
		if len(topics) != len(response.SubscribedTopics) {
			t.Errorf("Expected %d topics, got %d", len(response.SubscribedTopics), len(topics))
		}

		if _, ok := parsedResponse["createdAt"].(string); !ok {
			t.Errorf("createdAt debe ser string")
		}
		if v, ok := parsedResponse["lastSeenAt"]; ok && v != nil {
			if _, ok2 := v.(string); !ok2 {
				t.Errorf("lastSeenAt debe ser string cuando está presente")
			}
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

		var parsedResponse map[string]interface{}
		err = json.Unmarshal(jsonData, &parsedResponse)
		if err != nil {
			t.Errorf("Error parsing response: %v", err)
		}

		if int64(parsedResponse["pushEnvioId"].(float64)) != response.PushEnvioId {
			t.Errorf("Expected envio ID %d, got %v", response.PushEnvioId, parsedResponse["pushEnvioId"])
		}

		if parsedResponse["proveedor"].(string) != string(response.Proveedor) {
			t.Errorf("Expected proveedor %s, got %v", response.Proveedor, parsedResponse["proveedor"])
		}

		var parsedData map[string]interface{}
		if m, ok := parsedResponse["data"].(map[string]interface{}); ok {
			parsedData = m
		} else if s, ok := parsedResponse["data"].(string); ok {
			if err = json.Unmarshal([]byte(s), &parsedData); err != nil {
				t.Errorf("Error parsing data JSON: %v", err)
			}
		} else {
			t.Errorf("data tiene tipo inesperado: %T", parsedResponse["data"])
		}

		if _, ok := parsedResponse["sentAt"].(string); !ok {
			t.Errorf("sentAt debe ser string")
		}
	})
}

func stringPtr(s string) *string {
	return &s
}

func int64Ptr(i int64) *int64 {
	return &i
}

func intPtr(i int) *int {
	return &i
}
