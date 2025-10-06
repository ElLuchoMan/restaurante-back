package push

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Tests adicionales para aumentar cobertura de PushController

func TestPushGetAll_QueryError(t *testing.T) {
	controller, recorder, _ := setupPushTest()

	// Configurar mocks
	mockPushOrmer.On("QueryTable", "push_dispositivo").Return(mockPushQS)
	mockPushQS.On("Count").Return(int64(10), nil)
	mockPushQS.On("OrderBy", []string{"-updated_at"}).Return(mockPushQS)
	mockPushQS.On("Limit", 20).Return(mockPushQS)
	mockPushQS.On("Offset", int64(0)).Return(mockPushQS)
	mockPushQS.On("All", mock.AnythingOfType("*[]*models.PushDispositivo"), []string(nil)).Return(int64(0), fmt.Errorf("query error"))

	// Ejecutar
	controller.GetAll()

	// Verificar
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	mockPushOrmer.AssertExpectations(t)
	mockPushQS.AssertExpectations(t)
}

func TestPushGetAll_LimitExceedsMax(t *testing.T) {
	controller, recorder, ctx := setupPushTest()

	// Configurar query parameter con límite > 100
	ctx.Input.SetParam("limit", "200")

	// Configurar mocks
	mockPushOrmer.On("QueryTable", "push_dispositivo").Return(mockPushQS)
	mockPushQS.On("Count").Return(int64(10), nil)
	mockPushQS.On("OrderBy", []string{"-updated_at"}).Return(mockPushQS)
	mockPushQS.On("Limit", 100).Return(mockPushQS) // Debe limitar a 100
	mockPushQS.On("Offset", int64(0)).Return(mockPushQS)
	mockPushQS.On("All", mock.AnythingOfType("*[]*models.PushDispositivo"), []string(nil)).Return(int64(10), nil)

	// Ejecutar
	controller.GetAll()

	// Verificar
	assert.Equal(t, http.StatusOK, recorder.Code)
	mockPushOrmer.AssertExpectations(t)
	mockPushQS.AssertExpectations(t)
}

func TestPushPost_DatabaseError(t *testing.T) {
	controller, recorder, _ := setupPushTest()

	// Preparar request body válido
	dispositivo := map[string]interface{}{
		"token":      "test_token_123",
		"plataforma": "WEB",
		"clienteId":  123,
	}

	body, _ := json.Marshal(dispositivo)
	req := httptest.NewRequest("POST", "/push/dispositivos", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = body
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Configurar mock para error de inserción
	mockPushOrmer.On("Insert", mock.AnythingOfType("*models.PushDispositivo")).Return(int64(0), fmt.Errorf("database error"))

	// Ejecutar
	controller.Post()

	// Verificar
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	mockPushOrmer.AssertExpectations(t)
}

func TestPushGetById_ReadError(t *testing.T) {
	controller, recorder, _ := setupPushTest()

	// Request con ID válido
	req := httptest.NewRequest("GET", "/push/dispositivos?id=1", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Configurar mock para error de lectura
	mockPushOrmer.On("Read", mock.AnythingOfType("*models.PushDispositivo"), []string(nil)).Return(fmt.Errorf("database error"))

	// Ejecutar
	controller.GetById()

	// Verificar
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	mockPushOrmer.AssertExpectations(t)
}

func TestPushPut_InvalidID(t *testing.T) {
	controller, recorder, _ := setupPushTest()

	// Request con ID inválido
	body := []byte(`{"token":"new_token"}`)
	req := httptest.NewRequest("PUT", "/push/dispositivos?id=invalid", bytes.NewBuffer(body))
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

func TestPushPut_NotFound(t *testing.T) {
	controller, recorder, _ := setupPushTest()

	// Request con ID válido
	body := []byte(`{"token":"new_token","plataforma":"WEB"}`)
	req := httptest.NewRequest("PUT", "/push/dispositivos?id=999", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = body
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Configurar mock para no encontrado
	mockPushOrmer.On("Read", mock.AnythingOfType("*models.PushDispositivo"), []string(nil)).Return(orm.ErrNoRows)

	// Ejecutar
	controller.Put()

	// Verificar
	assert.Equal(t, http.StatusNotFound, recorder.Code)
	mockPushOrmer.AssertExpectations(t)
}

func TestPushDelete_InvalidID(t *testing.T) {
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

func TestPushDelete_NotFound(t *testing.T) {
	controller, recorder, _ := setupPushTest()

	// Request con ID válido
	req := httptest.NewRequest("DELETE", "/push/dispositivos?id=999", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Configurar mock para no encontrado
	mockPushOrmer.On("Read", mock.AnythingOfType("*models.PushDispositivo"), []string(nil)).Return(orm.ErrNoRows)

	// Ejecutar
	controller.Delete()

	// Verificar
	assert.Equal(t, http.StatusNotFound, recorder.Code)
	mockPushOrmer.AssertExpectations(t)
}

func TestPushActualizarUltimaVista_NotFound(t *testing.T) {
	controller, recorder, _ := setupPushTest()

	// Request con ID válido
	req := httptest.NewRequest("PUT", "/push/dispositivos/ultima-vista?id=999", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Configurar mock para no encontrado
	mockPushOrmer.On("Read", mock.AnythingOfType("*models.PushDispositivo"), []string(nil)).Return(orm.ErrNoRows)

	// Ejecutar
	controller.ActualizarUltimaVista()

	// Verificar
	assert.Equal(t, http.StatusNotFound, recorder.Code)
	mockPushOrmer.AssertExpectations(t)
}

func TestPushActualizarTopics_InvalidID(t *testing.T) {
	controller, recorder, _ := setupPushTest()

	// Request con ID inválido
	body := []byte(`{"topics":["news","sports"]}`)
	req := httptest.NewRequest("PUT", "/push/dispositivos/topics?id=invalid", bytes.NewBuffer(body))
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

func TestPushListarEnvios_CountError(t *testing.T) {
	controller, recorder, _ := setupPushTest()

	// Configurar mocks
	mockPushOrmer.On("QueryTable", "push_envio").Return(mockPushQS)
	mockPushQS.On("Count").Return(int64(0), fmt.Errorf("database error"))

	// Ejecutar
	controller.ListarEnvios()

	// Verificar
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	mockPushOrmer.AssertExpectations(t)
	mockPushQS.AssertExpectations(t)
}

func TestPushListarEnvios_QueryError(t *testing.T) {
	controller, recorder, _ := setupPushTest()

	// Configurar mocks
	mockPushOrmer.On("QueryTable", "push_envio").Return(mockPushQS)
	mockPushQS.On("Count").Return(int64(10), nil)
	mockPushQS.On("OrderBy", []string{"-created_at"}).Return(mockPushQS)
	mockPushQS.On("Limit", 20).Return(mockPushQS)
	mockPushQS.On("Offset", int64(0)).Return(mockPushQS)
	mockPushQS.On("All", mock.AnythingOfType("*[]*models.PushEnvio"), []string(nil)).Return(int64(0), fmt.Errorf("query error"))

	// Ejecutar
	controller.ListarEnvios()

	// Verificar
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	mockPushOrmer.AssertExpectations(t)
	mockPushQS.AssertExpectations(t)
}
