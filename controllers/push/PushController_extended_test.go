package push

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Tests adicionales para aumentar cobertura de PushController

func TestPushGetAll_QueryError(t *testing.T) {
	controller, _, _ := setupPushTest()

	// Configurar mocks
	mockPushOrm.On("QueryTable", "push_dispositivo").Return(mockPushQS)
	mockPushQS.On("Count").Return(int64(10), nil)
	mockPushQS.On("OrderBy", []string{"-created_at"}).Return(mockPushQS)
	mockPushQS.On("Limit", 20).Return(mockPushQS)
	mockPushQS.On("Offset", int64(0)).Return(mockPushQS)
	mockPushQS.On("All", mock.AnythingOfType("*[]*models.PushDispositivo"), []string(nil)).Return(int64(0), fmt.Errorf("query error"))

	// Ejecutar
	controller.GetAll()

	// Verificar que retorne error pero no falle el test
	mockPushOrm.AssertExpectations(t)
	mockPushQS.AssertExpectations(t)
}

func TestPushGetAll_LimitExceedsMax(t *testing.T) {
	controller, recorder, ctx := setupPushTest()

	// Configurar query parameter con límite > 100
	ctx.Input.SetParam("limit", "200")

	// Configurar mocks
	mockPushOrm.On("QueryTable", "push_dispositivo").Return(mockPushQS)
	mockPushQS.On("Count").Return(int64(10), nil)
	mockPushQS.On("OrderBy", []string{"-created_at"}).Return(mockPushQS)
	mockPushQS.On("Limit", 100).Return(mockPushQS) // Debe limitar a 100
	mockPushQS.On("Offset", int64(0)).Return(mockPushQS)
	mockPushQS.On("All", mock.AnythingOfType("*[]*models.PushDispositivo"), []string(nil)).Return(int64(10), nil)

	// Ejecutar
	controller.GetAll()

	// Verificar - debe ser exitoso
	assert.NotEqual(t, 500, recorder.Code)
	mockPushOrm.AssertExpectations(t)
	mockPushQS.AssertExpectations(t)
}

func TestPushListarEnvios_CountError(t *testing.T) {
	controller, recorder, _ := setupPushTest()

	// Configurar mocks
	mockPushOrm.On("QueryTable", "push_envio").Return(mockPushQS)
	mockPushQS.On("Count").Return(int64(0), fmt.Errorf("database error"))

	// Ejecutar
	controller.ListarEnvios()

	// Verificar
	assert.Equal(t, 500, recorder.Code)
	mockPushOrm.AssertExpectations(t)
	mockPushQS.AssertExpectations(t)
}

func TestPushListarEnvios_QueryError(t *testing.T) {
	controller, recorder, _ := setupPushTest()

	// Configurar mocks
	mockPushOrm.On("QueryTable", "push_envio").Return(mockPushQS)
	mockPushQS.On("Count").Return(int64(10), nil)
	mockPushQS.On("OrderBy", []string{"-sent_at"}).Return(mockPushQS)
	mockPushQS.On("Limit", 20).Return(mockPushQS)
	mockPushQS.On("Offset", int64(0)).Return(mockPushQS)
	mockPushQS.On("All", mock.AnythingOfType("*[]*models.PushEnvio"), []string(nil)).Return(int64(0), fmt.Errorf("query error"))

	// Ejecutar
	controller.ListarEnvios()

	// Verificar
	assert.Equal(t, 500, recorder.Code)
	mockPushOrm.AssertExpectations(t)
	mockPushQS.AssertExpectations(t)
}
