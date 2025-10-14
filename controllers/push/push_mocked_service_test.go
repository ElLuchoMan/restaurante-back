package push

import (
	"bytes"
	stdcontext "context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"restaurante/models"
	"restaurante/services"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockPushServiceInterface struct {
	mock.Mock
}

func (m *MockPushServiceInterface) RegistrarDispositivo(ctx stdcontext.Context, req *models.RegistrarDispositivoRequest) (*models.PushDispositivo, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PushDispositivo), args.Error(1)
}

func (m *MockPushServiceInterface) ActualizarUltimaVista(ctx stdcontext.Context, dispositivoId int64) error {
	args := m.Called(ctx, dispositivoId)
	return args.Error(0)
}

func (m *MockPushServiceInterface) ActualizarEstadoDispositivo(ctx stdcontext.Context, dispositivoId int64, enabled bool) error {
	args := m.Called(ctx, dispositivoId, enabled)
	return args.Error(0)
}

func (m *MockPushServiceInterface) ActualizarTopicsDispositivo(ctx stdcontext.Context, dispositivoId int64, topics []string) error {
	args := m.Called(ctx, dispositivoId, topics)
	return args.Error(0)
}

func (m *MockPushServiceInterface) RegistrarEnvio(ctx stdcontext.Context, req *models.RegistrarEnvioRequest) (*models.PushEnvio, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PushEnvio), args.Error(1)
}

func (m *MockPushServiceInterface) EnviarNotificacion(req *models.EnviarNotificacionRequest) (*models.EnviarNotificacionResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.EnviarNotificacionResponse), args.Error(1)
}

func (m *MockPushServiceInterface) ValidarRegistroDispositivo(req *models.RegistrarDispositivoRequest) error {
	args := m.Called(req)
	return args.Error(0)
}

func TestPushPost_ConServicioMockeado_Exitoso(t *testing.T) {

	originalNewPushService := newPushService
	originalNewServiceOrm := newServiceOrm
	defer func() {
		newPushService = originalNewPushService
		newServiceOrm = originalNewServiceOrm
	}()

	mockService := new(MockPushServiceInterface)
	expectedDispositivo := &models.PushDispositivo{
		PkIdPushDispositivo: 123,
		Plataforma:          models.PlataformaWeb,
		Enabled:             true,
	}
	mockService.On("RegistrarDispositivo", mock.Anything, mock.AnythingOfType("*models.RegistrarDispositivoRequest")).
		Return(expectedDispositivo, nil)

	newPushService = func(o orm.Ormer) services.PushServiceInterface {
		return mockService
	}
	newServiceOrm = func() orm.Ormer {
		return nil
	}

	ctrl := &PushController{}
	ctrl.Data = make(map[interface{}]interface{})

	req := models.RegistrarDispositivoRequest{
		Plataforma: models.PlataformaWeb,
	}
	body, _ := json.Marshal(req)

	r := httptest.NewRequest(http.MethodPost, "/push/dispositivos", bytes.NewReader(body))
	recorder := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(recorder, r)
	ctx.Input.RequestBody = body
	ctrl.Ctx = ctx

	ctrl.Post()

	assert.Equal(t, http.StatusCreated, recorder.Code)

	var response models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Equal(t, http.StatusCreated, response.Code)
	assert.Equal(t, "Dispositivo registrado exitosamente", response.Message)

	mockService.AssertExpectations(t)
}

func TestPushPost_ConServicioMockeado_ErrorDelServicio(t *testing.T) {

	originalNewPushService := newPushService
	originalNewServiceOrm := newServiceOrm
	defer func() {
		newPushService = originalNewPushService
		newServiceOrm = originalNewServiceOrm
	}()

	mockService := new(MockPushServiceInterface)
	mockService.On("RegistrarDispositivo", mock.Anything, mock.AnythingOfType("*models.RegistrarDispositivoRequest")).
		Return(nil, errors.New("error de base de datos"))

	newPushService = func(o orm.Ormer) services.PushServiceInterface {
		return mockService
	}
	newServiceOrm = func() orm.Ormer {
		return nil
	}

	ctrl := &PushController{}
	ctrl.Data = make(map[interface{}]interface{})

	req := models.RegistrarDispositivoRequest{
		Plataforma: models.PlataformaAndroid,
	}
	body, _ := json.Marshal(req)

	r := httptest.NewRequest(http.MethodPost, "/push/dispositivos", bytes.NewReader(body))
	recorder := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(recorder, r)
	ctx.Input.RequestBody = body
	ctrl.Ctx = ctx

	ctrl.Post()

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)

	var response models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.Equal(t, "Error al registrar dispositivo", response.Message)

	mockService.AssertExpectations(t)
}

func TestPushPut_ConServicioMockeado_Exitoso(t *testing.T) {
	originalNewPushService := newPushService
	originalNewServiceOrm := newServiceOrm
	defer func() {
		newPushService = originalNewPushService
		newServiceOrm = originalNewServiceOrm
	}()

	mockService := new(MockPushServiceInterface)
	mockService.On("ActualizarEstadoDispositivo", mock.Anything, int64(1), true).
		Return(nil)

	newPushService = func(o orm.Ormer) services.PushServiceInterface {
		return mockService
	}
	newServiceOrm = func() orm.Ormer {
		return nil
	}

	ctrl := &PushController{}
	ctrl.Data = make(map[interface{}]interface{})

	req := models.ActualizarEstadoDispositivoRequest{
		Enabled: true,
	}
	body, _ := json.Marshal(req)

	r := httptest.NewRequest(http.MethodPut, "/push/dispositivos?id=1", bytes.NewReader(body))
	recorder := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(recorder, r)
	ctx.Input.RequestBody = body
	ctrl.Ctx = ctx

	ctrl.Put()

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Equal(t, http.StatusOK, response.Code)

	mockService.AssertExpectations(t)
}

func TestPushPut_ConServicioMockeado_DispositivoNoEncontrado(t *testing.T) {
	originalNewPushService := newPushService
	originalNewServiceOrm := newServiceOrm
	defer func() {
		newPushService = originalNewPushService
		newServiceOrm = originalNewServiceOrm
	}()

	mockService := new(MockPushServiceInterface)
	mockService.On("ActualizarEstadoDispositivo", mock.Anything, int64(999), false).
		Return(errors.New("dispositivo no encontrado"))

	newPushService = func(o orm.Ormer) services.PushServiceInterface {
		return mockService
	}
	newServiceOrm = func() orm.Ormer {
		return nil
	}

	ctrl := &PushController{}
	ctrl.Data = make(map[interface{}]interface{})

	req := models.ActualizarEstadoDispositivoRequest{
		Enabled: false,
	}
	body, _ := json.Marshal(req)

	r := httptest.NewRequest(http.MethodPut, "/push/dispositivos?id=999", bytes.NewReader(body))
	recorder := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(recorder, r)
	ctx.Input.RequestBody = body
	ctrl.Ctx = ctx

	ctrl.Put()

	assert.Equal(t, http.StatusNotFound, recorder.Code)

	var response models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Equal(t, http.StatusNotFound, response.Code)
	assert.Equal(t, "Dispositivo no encontrado", response.Message)

	mockService.AssertExpectations(t)
}

func TestPushActualizarTopics_ConServicioMockeado_Exitoso(t *testing.T) {
	originalNewPushService := newPushService
	originalNewServiceOrm := newServiceOrm
	defer func() {
		newPushService = originalNewPushService
		newServiceOrm = originalNewServiceOrm
	}()

	mockService := new(MockPushServiceInterface)
	mockService.On("ActualizarTopicsDispositivo", mock.Anything, int64(1), []string{"ofertas", "cupones"}).
		Return(nil)

	newPushService = func(o orm.Ormer) services.PushServiceInterface {
		return mockService
	}
	newServiceOrm = func() orm.Ormer {
		return nil
	}

	ctrl := &PushController{}
	ctrl.Data = make(map[interface{}]interface{})

	req := models.ActualizarTopicsRequest{
		SubscribedTopics: []string{"ofertas", "cupones"},
	}
	body, _ := json.Marshal(req)

	r := httptest.NewRequest(http.MethodPatch, "/push/dispositivos/topics?id=1", bytes.NewReader(body))
	recorder := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(recorder, r)
	ctx.Input.RequestBody = body
	ctrl.Ctx = ctx

	ctrl.ActualizarTopics()

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Equal(t, http.StatusOK, response.Code)

	mockService.AssertExpectations(t)
}

func TestPushActualizarTopics_ConServicioMockeado_ErrorDelServicio(t *testing.T) {
	originalNewPushService := newPushService
	originalNewServiceOrm := newServiceOrm
	defer func() {
		newPushService = originalNewPushService
		newServiceOrm = originalNewServiceOrm
	}()

	mockService := new(MockPushServiceInterface)
	mockService.On("ActualizarTopicsDispositivo", mock.Anything, int64(1), []string{"test"}).
		Return(errors.New("error actualizando topics"))

	newPushService = func(o orm.Ormer) services.PushServiceInterface {
		return mockService
	}
	newServiceOrm = func() orm.Ormer {
		return nil
	}

	ctrl := &PushController{}
	ctrl.Data = make(map[interface{}]interface{})

	req := models.ActualizarTopicsRequest{
		SubscribedTopics: []string{"test"},
	}
	body, _ := json.Marshal(req)

	r := httptest.NewRequest(http.MethodPatch, "/push/dispositivos/topics?id=1", bytes.NewReader(body))
	recorder := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(recorder, r)
	ctx.Input.RequestBody = body
	ctrl.Ctx = ctx

	ctrl.ActualizarTopics()

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Equal(t, http.StatusNotFound, response.Code)

	mockService.AssertExpectations(t)
}

func TestPushActualizarUltimaVista_ConServicioMockeado_Exitoso(t *testing.T) {
	originalNewPushService := newPushService
	originalPushServiceOrmFactory := pushServiceOrmFactory
	defer func() {
		newPushService = originalNewPushService
		pushServiceOrmFactory = originalPushServiceOrmFactory
	}()

	mockService := new(MockPushServiceInterface)
	mockService.On("ActualizarUltimaVista", mock.Anything, int64(1)).
		Return(nil)

	newPushService = func(o orm.Ormer) services.PushServiceInterface {
		return mockService
	}
	pushServiceOrmFactory = func() orm.Ormer {
		return nil
	}

	ctrl := &PushController{}
	ctrl.Data = make(map[interface{}]interface{})

	r := httptest.NewRequest(http.MethodPatch, "/push/dispositivos/visto?id=1", nil)
	recorder := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(recorder, r)
	ctrl.Ctx = ctx

	ctrl.ActualizarUltimaVista()

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, "Última vista actualizada correctamente", response.Message)

	mockService.AssertExpectations(t)
}

func TestPushActualizarUltimaVista_ConServicioMockeado_IDInvalido(t *testing.T) {
	ctrl := &PushController{}
	ctrl.Data = make(map[interface{}]interface{})

	r := httptest.NewRequest(http.MethodPatch, "/push/dispositivos/visto?id=abc", nil)
	recorder := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(recorder, r)
	ctrl.Ctx = ctx

	ctrl.ActualizarUltimaVista()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "ID inválido o ausente", response.Message)
}

func TestPushActualizarUltimaVista_ConServicioMockeado_IDCero(t *testing.T) {
	ctrl := &PushController{}
	ctrl.Data = make(map[interface{}]interface{})

	r := httptest.NewRequest(http.MethodPatch, "/push/dispositivos/visto?id=0", nil)
	recorder := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(recorder, r)
	ctrl.Ctx = ctx

	ctrl.ActualizarUltimaVista()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Equal(t, http.StatusBadRequest, response.Code)
}

func TestPushActualizarUltimaVista_ConServicioMockeado_ErrorDelServicio(t *testing.T) {
	originalNewPushService := newPushService
	originalPushServiceOrmFactory := pushServiceOrmFactory
	defer func() {
		newPushService = originalNewPushService
		pushServiceOrmFactory = originalPushServiceOrmFactory
	}()

	mockService := new(MockPushServiceInterface)
	mockService.On("ActualizarUltimaVista", mock.Anything, int64(999)).
		Return(errors.New("dispositivo no encontrado"))

	newPushService = func(o orm.Ormer) services.PushServiceInterface {
		return mockService
	}
	pushServiceOrmFactory = func() orm.Ormer {
		return nil
	}

	ctrl := &PushController{}
	ctrl.Data = make(map[interface{}]interface{})

	r := httptest.NewRequest(http.MethodPatch, "/push/dispositivos/visto?id=999", nil)
	recorder := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(recorder, r)
	ctrl.Ctx = ctx

	ctrl.ActualizarUltimaVista()

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Equal(t, http.StatusNotFound, response.Code)
	assert.Equal(t, "Dispositivo no encontrado", response.Message)

	mockService.AssertExpectations(t)
}

func TestPushEnviarNotificacion_ConServicioMockeado_Exitoso(t *testing.T) {
	originalNewPushService := newPushService
	originalPushServiceOrmFactory := pushServiceOrmFactory
	defer func() {
		newPushService = originalNewPushService
		pushServiceOrmFactory = originalPushServiceOrmFactory
	}()

	mockService := new(MockPushServiceInterface)
	expectedResponse := &models.EnviarNotificacionResponse{
		EnviosExitosos: 5,
		EnviosFallidos: 0,
	}
	mockService.On("EnviarNotificacion", mock.AnythingOfType("*models.EnviarNotificacionRequest")).
		Return(expectedResponse, nil)

	newPushService = func(o orm.Ormer) services.PushServiceInterface {
		return mockService
	}
	pushServiceOrmFactory = func() orm.Ormer {
		return nil
	}

	ctrl := &PushController{}
	ctrl.Data = make(map[interface{}]interface{})

	req := models.EnviarNotificacionRequest{
		Notificacion: models.ContenidoNotificacion{
			Titulo:  "Título de prueba",
			Mensaje: "Mensaje de prueba",
		},
	}
	body, _ := json.Marshal(req)

	r := httptest.NewRequest(http.MethodPost, "/push/enviar", bytes.NewReader(body))
	recorder := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(recorder, r)
	ctx.Input.RequestBody = body
	ctrl.Ctx = ctx

	ctrl.EnviarNotificacion()

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, "Notificación enviada exitosamente", response.Message)

	mockService.AssertExpectations(t)
}

func TestPushEnviarNotificacion_ConServicioMockeado_TituloVacio(t *testing.T) {
	ctrl := &PushController{}
	ctrl.Data = make(map[interface{}]interface{})

	req := models.EnviarNotificacionRequest{
		Notificacion: models.ContenidoNotificacion{
			Titulo:  "",
			Mensaje: "Mensaje de prueba",
		},
	}
	body, _ := json.Marshal(req)

	r := httptest.NewRequest(http.MethodPost, "/push/enviar", bytes.NewReader(body))
	recorder := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(recorder, r)
	ctx.Input.RequestBody = body
	ctrl.Ctx = ctx

	ctrl.EnviarNotificacion()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "El título es requerido", response.Message)
}

func TestPushEnviarNotificacion_ConServicioMockeado_MensajeVacio(t *testing.T) {
	ctrl := &PushController{}
	ctrl.Data = make(map[interface{}]interface{})

	req := models.EnviarNotificacionRequest{
		Notificacion: models.ContenidoNotificacion{
			Titulo:  "Título de prueba",
			Mensaje: "",
		},
	}
	body, _ := json.Marshal(req)

	r := httptest.NewRequest(http.MethodPost, "/push/enviar", bytes.NewReader(body))
	recorder := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(recorder, r)
	ctx.Input.RequestBody = body
	ctrl.Ctx = ctx

	ctrl.EnviarNotificacion()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "El mensaje es requerido", response.Message)
}

func TestPushEnviarNotificacion_ConServicioMockeado_ErrorDelServicio(t *testing.T) {
	originalNewPushService := newPushService
	originalPushServiceOrmFactory := pushServiceOrmFactory
	defer func() {
		newPushService = originalNewPushService
		pushServiceOrmFactory = originalPushServiceOrmFactory
	}()

	mockService := new(MockPushServiceInterface)
	mockService.On("EnviarNotificacion", mock.AnythingOfType("*models.EnviarNotificacionRequest")).
		Return(nil, errors.New("error al enviar"))

	newPushService = func(o orm.Ormer) services.PushServiceInterface {
		return mockService
	}
	pushServiceOrmFactory = func() orm.Ormer {
		return nil
	}

	ctrl := &PushController{}
	ctrl.Data = make(map[interface{}]interface{})

	req := models.EnviarNotificacionRequest{
		Notificacion: models.ContenidoNotificacion{
			Titulo:  "Título de prueba",
			Mensaje: "Mensaje de prueba",
		},
	}
	body, _ := json.Marshal(req)

	r := httptest.NewRequest(http.MethodPost, "/push/enviar", bytes.NewReader(body))
	recorder := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(recorder, r)
	ctx.Input.RequestBody = body
	ctrl.Ctx = ctx

	ctrl.EnviarNotificacion()

	assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)

	var response models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Equal(t, http.StatusUnprocessableEntity, response.Code)
	assert.Equal(t, "Error al enviar notificación", response.Message)

	mockService.AssertExpectations(t)
}

func TestPushRegistrarEnvio_ConServicioMockeado_Exitoso(t *testing.T) {
	originalNewPushService := newPushService
	originalPushServiceOrmFactory := pushServiceOrmFactory
	defer func() {
		newPushService = originalNewPushService
		pushServiceOrmFactory = originalPushServiceOrmFactory
	}()

	mockService := new(MockPushServiceInterface)
	expectedEnvio := &models.PushEnvio{
		PkIdPushEnvio: 1,
		Proveedor:     models.ProveedorWebPush,
	}
	mockService.On("RegistrarEnvio", mock.Anything, mock.AnythingOfType("*models.RegistrarEnvioRequest")).
		Return(expectedEnvio, nil)

	newPushService = func(o orm.Ormer) services.PushServiceInterface {
		return mockService
	}
	pushServiceOrmFactory = func() orm.Ormer {
		return nil
	}

	ctrl := &PushController{}
	ctrl.Data = make(map[interface{}]interface{})

	req := models.RegistrarEnvioRequest{
		PkIdPushDispositivo: 1,
		Proveedor:           models.ProveedorWebPush,
		Exito:               true,
	}
	body, _ := json.Marshal(req)

	r := httptest.NewRequest(http.MethodPost, "/push/envios", bytes.NewReader(body))
	recorder := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(recorder, r)
	ctx.Input.RequestBody = body
	ctrl.Ctx = ctx

	ctrl.RegistrarEnvio()

	assert.Equal(t, http.StatusCreated, recorder.Code)

	var response models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Equal(t, http.StatusCreated, response.Code)
	assert.Equal(t, "Envío registrado exitosamente", response.Message)

	mockService.AssertExpectations(t)
}

func TestPushRegistrarEnvio_ConServicioMockeado_ProveedorInvalido(t *testing.T) {
	ctrl := &PushController{}
	ctrl.Data = make(map[interface{}]interface{})

	body := []byte(`{"pushDispositivoId":1,"proveedor":"INVALID","exito":true}`)

	r := httptest.NewRequest(http.MethodPost, "/push/envios", bytes.NewReader(body))
	recorder := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(recorder, r)
	ctx.Input.RequestBody = body
	ctrl.Ctx = ctx

	ctrl.RegistrarEnvio()

	assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)

	var response models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Equal(t, http.StatusUnprocessableEntity, response.Code)
	assert.Contains(t, response.Message, "Proveedor no válido")
}

func TestPushRegistrarEnvio_ConServicioMockeado_ErrorDelServicio(t *testing.T) {
	originalNewPushService := newPushService
	originalPushServiceOrmFactory := pushServiceOrmFactory
	defer func() {
		newPushService = originalNewPushService
		pushServiceOrmFactory = originalPushServiceOrmFactory
	}()

	mockService := new(MockPushServiceInterface)
	mockService.On("RegistrarEnvio", mock.Anything, mock.AnythingOfType("*models.RegistrarEnvioRequest")).
		Return(nil, errors.New("error de base de datos"))

	newPushService = func(o orm.Ormer) services.PushServiceInterface {
		return mockService
	}
	pushServiceOrmFactory = func() orm.Ormer {
		return nil
	}

	ctrl := &PushController{}
	ctrl.Data = make(map[interface{}]interface{})

	req := models.RegistrarEnvioRequest{
		PkIdPushDispositivo: 1,
		Proveedor:           models.ProveedorFCM,
		Exito:               true,
	}
	body, _ := json.Marshal(req)

	r := httptest.NewRequest(http.MethodPost, "/push/envios", bytes.NewReader(body))
	recorder := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(recorder, r)
	ctx.Input.RequestBody = body
	ctrl.Ctx = ctx

	ctrl.RegistrarEnvio()

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)

	var response models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.Equal(t, "Error al registrar envío", response.Message)

	mockService.AssertExpectations(t)
}

func TestPushDelete_IDFaltante(t *testing.T) {
	ctrl := &PushController{}
	ctrl.Data = make(map[interface{}]interface{})

	r := httptest.NewRequest(http.MethodDelete, "/push/dispositivos", nil)
	recorder := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(recorder, r)
	ctrl.Ctx = ctx

	ctrl.Delete()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Contains(t, response.Message, "ID")
}

func TestPushDelete_IDInvalido(t *testing.T) {
	ctrl := &PushController{}
	ctrl.Data = make(map[interface{}]interface{})

	r := httptest.NewRequest(http.MethodDelete, "/push/dispositivos?id=abc", nil)
	recorder := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(recorder, r)
	ctrl.Ctx = ctx

	ctrl.Delete()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Contains(t, response.Message, "ID")
}

func TestPushDelete_IDCero(t *testing.T) {
	ctrl := &PushController{}
	ctrl.Data = make(map[interface{}]interface{})

	r := httptest.NewRequest(http.MethodDelete, "/push/dispositivos?id=0", nil)
	recorder := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(recorder, r)
	ctrl.Ctx = ctx

	ctrl.Delete()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Contains(t, response.Message, "ID")
}
