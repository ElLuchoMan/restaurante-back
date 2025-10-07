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

// TestPut_InvalidID cubre el caso de ID inválido en Put
func TestPut_InvalidID(t *testing.T) {
	controller, recorder, _ := setupPushTest()

	// Request con ID inválido
	payload := models.ActualizarEstadoDispositivoRequest{
		Enabled: true,
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("PUT", "/push/dispositivos?id=invalid", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = body
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Ejecutar
	controller.Put()

	// Verificar
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

// TestPut_MissingID cubre el caso de ID ausente en Put
func TestPut_MissingID(t *testing.T) {
	controller, recorder, _ := setupPushTest()

	// Request sin ID
	payload := models.ActualizarEstadoDispositivoRequest{
		Enabled: false,
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("PUT", "/push/dispositivos", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = body
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Ejecutar
	controller.Put()

	// Verificar
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

// TestPut_InvalidJSON cubre el caso de JSON inválido en Put
func TestPut_InvalidJSON(t *testing.T) {
	controller, recorder, _ := setupPushTest()

	// Request con JSON inválido
	req := httptest.NewRequest("PUT", "/push/dispositivos?id=1", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = []byte("invalid json")
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Ejecutar
	controller.Put()

	// Verificar
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

// TestDelete_InvalidID cubre el caso de ID inválido en Delete
func TestDelete_InvalidID(t *testing.T) {
	controller, recorder, _ := setupPushTest()

	// Request con ID inválido
	req := httptest.NewRequest("DELETE", "/push/dispositivos?id=invalid", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Ejecutar
	controller.Delete()

	// Verificar
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

// TestDelete_MissingID cubre el caso de ID ausente en Delete
func TestDelete_MissingID(t *testing.T) {
	controller, recorder, _ := setupPushTest()

	// Request sin ID
	req := httptest.NewRequest("DELETE", "/push/dispositivos", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Ejecutar
	controller.Delete()

	// Verificar
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

// TestDelete_ZeroID cubre el caso de ID = 0 en Delete
func TestDelete_ZeroID(t *testing.T) {
	controller, recorder, _ := setupPushTest()

	// Request con ID = 0
	req := httptest.NewRequest("DELETE", "/push/dispositivos?id=0", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Ejecutar
	controller.Delete()

	// Verificar
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

// TestPut_ZeroID cubre el caso de ID = 0 en Put
func TestPut_ZeroID(t *testing.T) {
	controller, recorder, _ := setupPushTest()

	// Request con ID = 0
	payload := models.ActualizarEstadoDispositivoRequest{
		Enabled: true,
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("PUT", "/push/dispositivos?id=0", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = body
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Ejecutar
	controller.Put()

	// Verificar
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}
