package push

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPushGetAll_QueryError(t *testing.T) {
	controller, _, _ := setupPushTest()

	mockPushOrm.On("QueryTable", "push_dispositivo").Return(mockPushQS)
	mockPushQS.On("Count").Return(int64(10), nil)
	mockPushQS.On("OrderBy", []string{"-created_at"}).Return(mockPushQS)
	mockPushQS.On("Limit", 20).Return(mockPushQS)
	mockPushQS.On("Offset", int64(0)).Return(mockPushQS)
	mockPushQS.On("All", mock.AnythingOfType("*[]*models.PushDispositivo"), []string(nil)).Return(int64(0), fmt.Errorf("query error"))

	controller.GetAll()

	mockPushOrm.AssertExpectations(t)
	mockPushQS.AssertExpectations(t)
}

func TestPushGetAll_LimitExceedsMax(t *testing.T) {
	controller, recorder, ctx := setupPushTest()

	ctx.Input.SetParam("limit", "200")

	mockPushOrm.On("QueryTable", "push_dispositivo").Return(mockPushQS)
	mockPushQS.On("Count").Return(int64(10), nil)
	mockPushQS.On("OrderBy", []string{"-created_at"}).Return(mockPushQS)
	mockPushQS.On("Limit", 100).Return(mockPushQS)
	mockPushQS.On("Offset", int64(0)).Return(mockPushQS)
	mockPushQS.On("All", mock.AnythingOfType("*[]*models.PushDispositivo"), []string(nil)).Return(int64(10), nil)

	controller.GetAll()

	assert.NotEqual(t, 500, recorder.Code)
	mockPushOrm.AssertExpectations(t)
	mockPushQS.AssertExpectations(t)
}

func TestPushListarEnvios_CountError(t *testing.T) {
	controller, recorder, _ := setupPushTest()

	mockPushOrm.On("QueryTable", "push_envio").Return(mockPushQS)
	mockPushQS.On("Count").Return(int64(0), fmt.Errorf("database error"))

	controller.ListarEnvios()

	assert.Equal(t, 500, recorder.Code)
	mockPushOrm.AssertExpectations(t)
	mockPushQS.AssertExpectations(t)
}

func TestPushListarEnvios_QueryError(t *testing.T) {
	controller, recorder, _ := setupPushTest()

	mockPushOrm.On("QueryTable", "push_envio").Return(mockPushQS)
	mockPushQS.On("Count").Return(int64(10), nil)
	mockPushQS.On("OrderBy", []string{"-sent_at"}).Return(mockPushQS)
	mockPushQS.On("Limit", 20).Return(mockPushQS)
	mockPushQS.On("Offset", int64(0)).Return(mockPushQS)
	mockPushQS.On("All", mock.AnythingOfType("*[]*models.PushEnvio"), []string(nil)).Return(int64(0), fmt.Errorf("query error"))

	controller.ListarEnvios()

	assert.Equal(t, 500, recorder.Code)
	mockPushOrm.AssertExpectations(t)
	mockPushQS.AssertExpectations(t)
}
