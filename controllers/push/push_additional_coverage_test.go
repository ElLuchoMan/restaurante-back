package push

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"restaurante/models"

	"github.com/beego/beego/v2/server/web/context"
	"github.com/stretchr/testify/assert"
)

// TestActualizarUltimaVista_InvalidID cubre el caso de ID inválido
func TestActualizarUltimaVista_InvalidID(t *testing.T) {
	controller, recorder, _ := setupPushTest()

	// Request con ID inválido
	req := httptest.NewRequest("PATCH", "/push/dispositivos/visto?id=invalid", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Ejecutar
	controller.ActualizarUltimaVista()

	// Verificar
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

// TestActualizarUltimaVista_MissingID cubre el caso de ID ausente
func TestActualizarUltimaVista_MissingID(t *testing.T) {
	controller, recorder, _ := setupPushTest()

	// Request sin ID
	req := httptest.NewRequest("PATCH", "/push/dispositivos/visto", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Ejecutar
	controller.ActualizarUltimaVista()

	// Verificar
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

// TestActualizarTopics_InvalidID cubre el caso de ID inválido
func TestActualizarTopics_InvalidID(t *testing.T) {
	controller, recorder, _ := setupPushTest()

	// Request con ID inválido
	payload := models.ActualizarTopicsRequest{
		SubscribedTopics: []string{"test-topic"},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("PATCH", "/push/dispositivos/topics?id=invalid", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = body
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Ejecutar
	controller.ActualizarTopics()

	// Verificar
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

// TestActualizarTopics_MissingID cubre el caso de ID ausente
func TestActualizarTopics_MissingID(t *testing.T) {
	controller, recorder, _ := setupPushTest()

	// Request sin ID
	payload := models.ActualizarTopicsRequest{
		SubscribedTopics: []string{"test-topic"},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("PATCH", "/push/dispositivos/topics", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = body
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Ejecutar
	controller.ActualizarTopics()

	// Verificar
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

// TestActualizarTopics_InvalidJSON cubre el caso de JSON inválido
func TestActualizarTopics_InvalidJSON(t *testing.T) {
	controller, recorder, _ := setupPushTest()

	// Request con JSON inválido
	req := httptest.NewRequest("PATCH", "/push/dispositivos/topics?id=1", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = []byte("invalid json")
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Ejecutar
	controller.ActualizarTopics()

	// Verificar
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

// TestEnviarNotificacion_InvalidJSON cubre el caso de JSON inválido
func TestEnviarNotificacion_InvalidJSON(t *testing.T) {
	controller, recorder, _ := setupPushTest()

	// Request con JSON inválido
	req := httptest.NewRequest("POST", "/push/enviar", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = []byte("invalid json")
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Ejecutar
	controller.EnviarNotificacion()

	// Verificar
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

// TestRegistrarEnvio_InvalidJSON cubre el caso de JSON inválido
func TestRegistrarEnvio_InvalidJSON(t *testing.T) {
	controller, recorder, _ := setupPushTest()

	// Request con JSON inválido
	req := httptest.NewRequest("POST", "/push/envios", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = []byte("invalid json")
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Ejecutar
	controller.RegistrarEnvio()

	// Verificar
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}
