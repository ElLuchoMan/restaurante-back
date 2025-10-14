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

func TestActualizarUltimaVista_InvalidID(t *testing.T) {
	controller, recorder, _ := setupPushTest()

	req := httptest.NewRequest("PATCH", "/push/dispositivos/visto?id=invalid", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	controller.ActualizarUltimaVista()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestActualizarUltimaVista_MissingID(t *testing.T) {
	controller, recorder, _ := setupPushTest()

	req := httptest.NewRequest("PATCH", "/push/dispositivos/visto", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	controller.ActualizarUltimaVista()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestActualizarTopics_InvalidID(t *testing.T) {
	controller, recorder, _ := setupPushTest()

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

	controller.ActualizarTopics()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestActualizarTopics_MissingID(t *testing.T) {
	controller, recorder, _ := setupPushTest()

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

	controller.ActualizarTopics()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestActualizarTopics_InvalidJSON(t *testing.T) {
	controller, recorder, _ := setupPushTest()

	req := httptest.NewRequest("PATCH", "/push/dispositivos/topics?id=1", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = []byte("invalid json")
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	controller.ActualizarTopics()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestEnviarNotificacion_InvalidJSON(t *testing.T) {
	controller, recorder, _ := setupPushTest()

	req := httptest.NewRequest("POST", "/push/enviar", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = []byte("invalid json")
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	controller.EnviarNotificacion()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestRegistrarEnvio_InvalidJSON(t *testing.T) {
	controller, recorder, _ := setupPushTest()

	req := httptest.NewRequest("POST", "/push/envios", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = []byte("invalid json")
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	controller.RegistrarEnvio()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}
