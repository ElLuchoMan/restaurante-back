package push

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"restaurante/models"
	"restaurante/services"

	"github.com/beego/beego/v2/client/orm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func decodeAPI(t *testing.T, body []byte) models.ApiResponse {
	t.Helper()
	var resp models.ApiResponse
	assert.NoError(t, json.Unmarshal(body, &resp))
	return resp
}

func TestPushGetById_ReadError(t *testing.T) {
	controller, recorder, ctx := setupPushTest()
	ctx.Request = httptest.NewRequest("GET", "/push/dispositivos/search?id=7", nil)
	mockPushOrm.On("Read", mock.AnythingOfType("*models.PushDispositivo"), []string(nil)).Return(errors.New("read boom"))

	controller.GetById()

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	resp := decodeAPI(t, recorder.Body.Bytes())
	assert.Equal(t, "Error interno del servidor", resp.Message)
	assert.Equal(t, "read boom", resp.Cause)
	mockPushOrm.AssertExpectations(t)
}

func TestPushDelete_NotFound(t *testing.T) {
	controller, recorder, ctx := setupPushTest()
	ctx.Input.SetParam(":id", "9")
	mockPushOrm.On("Read", mock.AnythingOfType("*models.PushDispositivo"), []string(nil)).Return(orm.ErrNoRows)

	controller.Delete()

	resp := decodeAPI(t, recorder.Body.Bytes())
	assert.Equal(t, http.StatusNotFound, resp.Code)
	assert.Equal(t, "Dispositivo no encontrado", resp.Message)
	mockPushOrm.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
}

func TestPushDelete_ReadError(t *testing.T) {
	controller, recorder, ctx := setupPushTest()
	ctx.Input.SetParam(":id", "9")
	mockPushOrm.On("Read", mock.AnythingOfType("*models.PushDispositivo"), []string(nil)).Return(errors.New("read boom"))

	controller.Delete()

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	resp := decodeAPI(t, recorder.Body.Bytes())
	assert.Equal(t, "Error interno del servidor", resp.Message)
	assert.Equal(t, "read boom", resp.Cause)
}

func TestPushDelete_DeleteError(t *testing.T) {
	controller, recorder, ctx := setupPushTest()
	ctx.Input.SetParam(":id", "9")
	mockPushOrm.On("Read", mock.AnythingOfType("*models.PushDispositivo"), []string(nil)).Return(nil)
	mockPushOrm.On("Delete", mock.AnythingOfType("*models.PushDispositivo"), []string(nil)).Return(int64(0), errors.New("delete boom"))

	controller.Delete()

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	resp := decodeAPI(t, recorder.Body.Bytes())
	assert.Equal(t, "Error al eliminar dispositivo", resp.Message)
	assert.Equal(t, "delete boom", resp.Cause)
}

func TestPushPut_IDFromPathParam(t *testing.T) {
	controller, recorder, ctx := setupPushTest()
	ctx.Input.SetParam(":id", "4")
	ctx.Input.RequestBody = []byte(`{"enabled":false}`)

	mockSvc := new(MockPushServiceInterface)
	mockSvc.On("ActualizarEstadoDispositivo", mock.Anything, int64(4), false).Return(nil)
	origSvc, origOrm := newPushService, newServiceOrm
	newPushService = func(orm.Ormer) services.PushServiceInterface { return mockSvc }
	newServiceOrm = func() orm.Ormer { return nil }
	t.Cleanup(func() { newPushService, newServiceOrm = origSvc, origOrm })

	controller.Put()

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "Estado actualizado correctamente", decodeAPI(t, recorder.Body.Bytes()).Message)
	mockSvc.AssertExpectations(t)
}

func TestPushActualizarUltimaVista_IDFromPathParam(t *testing.T) {
	controller, recorder, ctx := setupPushTest()
	ctx.Input.SetParam(":id", "6")

	mockSvc := new(MockPushServiceInterface)
	mockSvc.On("ActualizarUltimaVista", mock.Anything, int64(6)).Return(nil)
	origSvc := newPushService
	newPushService = func(orm.Ormer) services.PushServiceInterface { return mockSvc }
	t.Cleanup(func() { newPushService = origSvc })

	controller.ActualizarUltimaVista()

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "Última vista actualizada correctamente", decodeAPI(t, recorder.Body.Bytes()).Message)
	mockSvc.AssertExpectations(t)
}

func TestPushListarEnvios_FiltersAndSuccess(t *testing.T) {
	controller, recorder, ctx := setupPushTest()
	ctx.Request = httptest.NewRequest("GET",
		"/push/envios?dispositivo_id=3&fecha_desde=2025-01-01&fecha_hasta=2025-01-31&limit=500&offset=0", nil)

	mockPushOrm.On("QueryTable", "push_envio").Return(mockPushQS)
	mockPushQS.On("Filter", "pk_id_push_dispositivo", mock.Anything).Return(mockPushQS)
	mockPushQS.On("Filter", "sent_at__gte", mock.Anything).Return(mockPushQS)
	mockPushQS.On("Filter", "sent_at__lte", mock.Anything).Return(mockPushQS)
	mockPushQS.On("Count").Return(int64(250), nil)
	mockPushQS.On("OrderBy", []string{"-sent_at"}).Return(mockPushQS)
	mockPushQS.On("Limit", 100).Return(mockPushQS)
	mockPushQS.On("Offset", int64(0)).Return(mockPushQS)
	mockPushQS.On("All", mock.AnythingOfType("*[]*models.PushEnvio"), []string(nil)).Return(int64(0), nil)

	controller.ListarEnvios()

	assert.Equal(t, http.StatusOK, recorder.Code)
	resp := decodeAPI(t, recorder.Body.Bytes())
	data := resp.Data.(map[string]interface{})
	assert.Equal(t, float64(250), data["total"])
	assert.Equal(t, float64(100), data["pageSize"])
	assert.Equal(t, float64(3), data["totalPages"])
	mockPushQS.AssertCalled(t, "Filter", "pk_id_push_dispositivo", mock.Anything)
	mockPushQS.AssertCalled(t, "Filter", "sent_at__gte", mock.Anything)
	mockPushQS.AssertCalled(t, "Filter", "sent_at__lte", mock.Anything)
}

func TestPushListarEnvios_CountErrorCause(t *testing.T) {
	controller, recorder, _ := setupPushTest()
	mockPushOrm.On("QueryTable", "push_envio").Return(mockPushQS)
	mockPushQS.On("Count").Return(int64(0), errors.New("count boom"))

	controller.ListarEnvios()

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	assert.Equal(t, "count boom", decodeAPI(t, recorder.Body.Bytes()).Cause)
}

func TestPushListarEnvios_QueryErrorCause(t *testing.T) {
	controller, recorder, _ := setupPushTest()
	mockPushOrm.On("QueryTable", "push_envio").Return(mockPushQS)
	mockPushQS.On("Count").Return(int64(1), nil)
	mockPushQS.On("OrderBy", []string{"-sent_at"}).Return(mockPushQS)
	mockPushQS.On("Limit", 20).Return(mockPushQS)
	mockPushQS.On("Offset", int64(0)).Return(mockPushQS)
	mockPushQS.On("All", mock.AnythingOfType("*[]*models.PushEnvio"), []string(nil)).Return(int64(0), errors.New("all boom"))

	controller.ListarEnvios()

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	resp := decodeAPI(t, recorder.Body.Bytes())
	assert.Equal(t, "Error al obtener envíos", resp.Message)
	assert.Equal(t, "all boom", resp.Cause)
}
