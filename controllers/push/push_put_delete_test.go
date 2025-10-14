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

func TestPut_InvalidID(t *testing.T) {
	controller, recorder, _ := setupPushTest()

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

	controller.Put()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestPut_MissingID(t *testing.T) {
	controller, recorder, _ := setupPushTest()

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

	controller.Put()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestPut_InvalidJSON(t *testing.T) {
	controller, recorder, _ := setupPushTest()

	req := httptest.NewRequest("PUT", "/push/dispositivos?id=1", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = []byte("invalid json")
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	controller.Put()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestDelete_InvalidID(t *testing.T) {
	controller, recorder, _ := setupPushTest()

	req := httptest.NewRequest("DELETE", "/push/dispositivos?id=invalid", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	controller.Delete()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestDelete_MissingID(t *testing.T) {
	controller, recorder, _ := setupPushTest()

	req := httptest.NewRequest("DELETE", "/push/dispositivos", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	controller.Delete()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestDelete_ZeroID(t *testing.T) {
	controller, recorder, _ := setupPushTest()

	req := httptest.NewRequest("DELETE", "/push/dispositivos?id=0", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	controller.Delete()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestPut_ZeroID(t *testing.T) {
	controller, recorder, _ := setupPushTest()

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

	controller.Put()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}
