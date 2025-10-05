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
	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock para pushQuerySeter
type mockPushQuerySeter struct {
	mock.Mock
}

func (m *mockPushQuerySeter) All(container interface{}, cols ...string) (int64, error) {
	args := m.Called(container, cols)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockPushQuerySeter) Filter(expr string, args ...interface{}) pushQuerySeter {
	m.Called(expr, args)
	return m
}

func (m *mockPushQuerySeter) OrderBy(exprs ...string) pushQuerySeter {
	m.Called(exprs)
	return m
}

func (m *mockPushQuerySeter) Limit(limit int) pushQuerySeter {
	m.Called(limit)
	return m
}

func (m *mockPushQuerySeter) Offset(offset int64) pushQuerySeter {
	m.Called(offset)
	return m
}

func (m *mockPushQuerySeter) Count() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockPushQuerySeter) One(container interface{}) error {
	args := m.Called(container)
	return args.Error(0)
}

// Mock para pushOrmer
type mockPushOrmer struct {
	mock.Mock
}

func (m *mockPushOrmer) QueryTable(ptrStructOrTableName interface{}) pushQuerySeter {
	args := m.Called(ptrStructOrTableName)
	return args.Get(0).(pushQuerySeter)
}

func (m *mockPushOrmer) Insert(md interface{}) (int64, error) {
	args := m.Called(md)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockPushOrmer) Read(md interface{}, cols ...string) error {
	args := m.Called(md, cols)
	return args.Error(0)
}

func (m *mockPushOrmer) Update(md interface{}, cols ...string) (int64, error) {
	args := m.Called(md, cols)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockPushOrmer) Delete(md interface{}, cols ...string) (int64, error) {
	args := m.Called(md, cols)
	return args.Get(0).(int64), args.Error(1)
}

// Variables globales para mocks
var mockPushOrm *mockPushOrmer
var mockPushQS *mockPushQuerySeter

// Override de las funciones de creación
func init() {
	pushOrmNew = func() pushOrmer {
		return mockPushOrm
	}
	// Mockear el servicio - para tests que no requieren servicio real
	// Los tests que necesiten servicio real deben sobreescribir esto
}

func setupPushTest() (*PushController, *httptest.ResponseRecorder, *context.Context) {
	// Reset mocks
	mockPushOrm = &mockPushOrmer{}
	mockPushQS = &mockPushQuerySeter{}

	// Crear controller
	controller := &PushController{}
	controller.Controller = web.Controller{}

	// Crear response recorder
	recorder := httptest.NewRecorder()

	// Crear request
	req := httptest.NewRequest("GET", "/", nil)

	// Crear contexto usando la forma correcta
	ctx := context.NewContext()
	ctx.Reset(recorder, req)

	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	return controller, recorder, ctx
}

func TestPushGetAll_Success(t *testing.T) {
	controller, recorder, _ := setupPushTest()

	// Configurar mocks
	cliente := &models.Cliente{PK_DOCUMENTO_CLIENTE: 123}
	trabajador := &models.Trabajador{PK_DOCUMENTO_TRABAJADOR: 456}
	dispositivos := []*models.PushDispositivo{
		{PkIdPushDispositivo: 1, Plataforma: "WEB", Enabled: true, PkDocumentoCliente: cliente},
		{PkIdPushDispositivo: 2, Plataforma: "ANDROID", Enabled: true, PkDocumentoTrabajador: trabajador},
	}

	mockPushOrm.On("QueryTable", "push_dispositivo").Return(mockPushQS)
	mockPushQS.On("Count").Return(int64(2), nil)
	mockPushQS.On("OrderBy", []string{"-created_at"}).Return(mockPushQS)
	mockPushQS.On("Limit", 20).Return(mockPushQS)
	mockPushQS.On("Offset", int64(0)).Return(mockPushQS)
	mockPushQS.On("All", mock.AnythingOfType("*[]*models.PushDispositivo"), []string(nil)).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*[]*models.PushDispositivo)
		*arg = dispositivos
	}).Return(int64(2), nil)

	// Ejecutar
	controller.GetAll()

	// Verificar
	assert.Equal(t, http.StatusOK, recorder.Code)
	mockPushOrm.AssertExpectations(t)
	mockPushQS.AssertExpectations(t)
}

func TestPushGetAll_WithFilters(t *testing.T) {
	controller, recorder, ctx := setupPushTest()

	// Configurar query parameters
	ctx.Input.SetParam("cliente_id", "123")
	ctx.Input.SetParam("trabajador_id", "456")
	ctx.Input.SetParam("plataforma", "WEB")
	ctx.Input.SetParam("limit", "10")
	ctx.Input.SetParam("offset", "5")

	// Configurar mocks
	mockPushOrm.On("QueryTable", "push_dispositivo").Return(mockPushQS)
	mockPushQS.On("Filter", "pk_documento_cliente", []interface{}{int64(123)}).Return(mockPushQS)
	mockPushQS.On("Filter", "pk_documento_trabajador", []interface{}{int64(456)}).Return(mockPushQS)
	mockPushQS.On("Filter", "plataforma", []interface{}{"WEB"}).Return(mockPushQS)
	mockPushQS.On("Count").Return(int64(1), nil)
	mockPushQS.On("OrderBy", []string{"-created_at"}).Return(mockPushQS)
	mockPushQS.On("Limit", 10).Return(mockPushQS)
	mockPushQS.On("Offset", int64(5)).Return(mockPushQS)
	mockPushQS.On("All", mock.AnythingOfType("*[]*models.PushDispositivo"), []string(nil)).Return(int64(1), nil)

	// Ejecutar
	controller.GetAll()

	// Verificar
	assert.Equal(t, http.StatusOK, recorder.Code)
	mockPushOrm.AssertExpectations(t)
	mockPushQS.AssertExpectations(t)
}

func TestPushGetAll_CountError(t *testing.T) {
	controller, recorder, _ := setupPushTest()

	// Configurar mocks
	mockPushOrm.On("QueryTable", "push_dispositivo").Return(mockPushQS)
	mockPushQS.On("Count").Return(int64(0), fmt.Errorf("database error"))

	// Ejecutar
	controller.GetAll()

	// Verificar
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	mockPushOrm.AssertExpectations(t)
	mockPushQS.AssertExpectations(t)
}

func TestPushPost_Success(t *testing.T) {
	t.Skip("TODO: Requiere refactorizar controller para inyectar PushService - services.NewPushService(orm.NewOrm()) no es mockeable")
	controller, recorder, ctx := setupPushTest()

	// Preparar request body usando el DTO correcto
	endpoint := "https://test.endpoint.com"
	p256dh := "test_p256dh"
	auth := "test_auth"
	docCliente := int64(123)
	request := models.RegistrarDispositivoRequest{
		Plataforma:         "WEB",
		Endpoint:           &endpoint,
		P256dh:             &p256dh,
		Auth:               &auth,
		PkDocumentoCliente: &docCliente,
		SubscribedTopics:   []string{"promos", "novedades"},
	}

	body, _ := json.Marshal(request)
	ctx.Request = httptest.NewRequest("POST", "/push/dispositivos", bytes.NewBuffer(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Input.RequestBody = body

	// Configurar mocks
	mockPushOrm.On("Insert", mock.AnythingOfType("*models.PushDispositivo")).Return(int64(1), nil)

	// Ejecutar
	controller.Post()

	// Verificar
	assert.Equal(t, http.StatusCreated, recorder.Code)
	mockPushOrm.AssertExpectations(t)
}

func TestPushPost_InvalidJSON(t *testing.T) {
	controller, recorder, ctx := setupPushTest()

	// Request con JSON inválido
	ctx.Request = httptest.NewRequest("POST", "/push/dispositivos", bytes.NewBuffer([]byte("invalid json")))
	ctx.Request.Header.Set("Content-Type", "application/json")

	// Ejecutar
	controller.Post()

	// Verificar
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestPushPost_ValidationError(t *testing.T) {
	controller, recorder, ctx := setupPushTest()

	// Dispositivo con datos inválidos
	dispositivo := models.PushDispositivo{
		Plataforma: "INVALID", // Plataforma inválida
		Enabled:    true,
	}

	body, _ := json.Marshal(dispositivo)
	ctx.Request = httptest.NewRequest("POST", "/push/dispositivos", bytes.NewBuffer(body))
	ctx.Request.Header.Set("Content-Type", "application/json")

	// Ejecutar
	controller.Post()

	// Verificar
	assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
}

func TestPushGetById_Success(t *testing.T) {
	controller, recorder, ctx := setupPushTest()

	// Configurar parámetro ID
	ctx.Input.SetParam(":id", "1")

	// Configurar mock
	dispositivo := models.PushDispositivo{PkIdPushDispositivo: 1, Plataforma: "WEB", Enabled: true}
	mockPushOrm.On("Read", mock.AnythingOfType("*models.PushDispositivo"), []string(nil)).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.PushDispositivo)
		*arg = dispositivo
	}).Return(nil)

	// Ejecutar
	controller.GetById()

	// Verificar
	assert.Equal(t, http.StatusOK, recorder.Code)
	mockPushOrm.AssertExpectations(t)
}

func TestPushGetById_InvalidID(t *testing.T) {
	controller, recorder, ctx := setupPushTest()

	// ID inválido
	ctx.Input.SetParam(":id", "invalid")

	// Ejecutar
	controller.GetById()

	// Verificar
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestPushGetById_NotFound(t *testing.T) {
	controller, recorder, ctx := setupPushTest()

	// Configurar parámetro ID usando query string
	ctx.Request = httptest.NewRequest("GET", "/push/dispositivos/search?id=999", nil)

	// Configurar mock para no encontrado (usar orm.ErrNoRows)
	mockPushOrm.On("Read", mock.AnythingOfType("*models.PushDispositivo"), []string(nil)).Return(orm.ErrNoRows)

	// Ejecutar
	controller.GetById()

	// Verificar
	assert.Equal(t, http.StatusNotFound, recorder.Code)
	mockPushOrm.AssertExpectations(t)
}

func TestPushPut_Success(t *testing.T) {
	t.Skip("TODO: Requiere refactorizar controller para inyectar PushService completo - el servicio usa orm.NewOrm() internamente y accede a la BD")
	// El servicio PushService.ActualizarEstadoDispositivo hace Read + Update en la BD
	// Para testear esto sin BD, necesitaríamos mockear el servicio completo o refactorizar
	// para que el servicio use el Ormer inyectado en lugar de crear uno nuevo
}

func TestPushDelete_Success(t *testing.T) {
	controller, recorder, ctx := setupPushTest()

	// Configurar parámetro ID
	ctx.Input.SetParam(":id", "1")

	// Configurar mocks
	dispositivo := models.PushDispositivo{PkIdPushDispositivo: 1, Plataforma: "WEB"}
	mockPushOrm.On("Read", mock.AnythingOfType("*models.PushDispositivo"), []string(nil)).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.PushDispositivo)
		*arg = dispositivo
	}).Return(nil)
	mockPushOrm.On("Delete", mock.AnythingOfType("*models.PushDispositivo"), []string(nil)).Return(int64(1), nil)

	// Ejecutar
	controller.Delete()

	// Verificar
	assert.Equal(t, http.StatusOK, recorder.Code)
	mockPushOrm.AssertExpectations(t)
}

func TestPushActualizarUltimaVista_Success(t *testing.T) {
	t.Skip("TODO: refactorizar controlador para inyectar servicio en lugar de crear ORM directamente")
	controller, recorder, ctx := setupPushTest()

	// Configurar parámetro ID
	ctx.Input.SetParam(":id", "1")

	// Configurar mocks
	dispositivo := models.PushDispositivo{PkIdPushDispositivo: 1, Plataforma: "WEB"}
	mockPushOrm.On("Read", mock.AnythingOfType("*models.PushDispositivo"), []string(nil)).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.PushDispositivo)
		*arg = dispositivo
	}).Return(nil)
	mockPushOrm.On("Update", mock.AnythingOfType("*models.PushDispositivo"), []string{"last_seen_at"}).Return(int64(1), nil)

	// Ejecutar
	controller.ActualizarUltimaVista()

	// Verificar
	assert.Equal(t, http.StatusOK, recorder.Code)
	mockPushOrm.AssertExpectations(t)
}

func TestPushActualizarUltimaVista_InvalidID(t *testing.T) {
	controller, recorder, ctx := setupPushTest()

	// ID inválido
	ctx.Input.SetParam(":id", "invalid")

	// Ejecutar
	controller.ActualizarUltimaVista()

	// Verificar
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestPushActualizarTopics_Success(t *testing.T) {
	t.Skip("TODO: Requiere refactorizar controller para inyectar PushService completo - el servicio usa orm.NewOrm() internamente y accede a la BD")
	// El servicio PushService.ActualizarTopics hace Read + Update en la BD
}

func TestPushActualizarTopics_InvalidJSON(t *testing.T) {
	controller, recorder, ctx := setupPushTest()

	// Configurar parámetro ID
	ctx.Input.SetParam(":id", "1")

	// Request con JSON inválido
	ctx.Request = httptest.NewRequest("PATCH", "/push/dispositivos/1/topics", bytes.NewBuffer([]byte("invalid json")))
	ctx.Request.Header.Set("Content-Type", "application/json")

	// Ejecutar
	controller.ActualizarTopics()

	// Verificar
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestPushEnviarNotificacion_Success(t *testing.T) {
	t.Skip("TODO: Requiere refactorizar controller para inyectar PushService completo - el servicio usa orm.NewOrm() internamente y accede a la BD")
	// El servicio PushService.EnviarNotificacion hace operaciones complejas con la BD
}

func TestPushEnviarNotificacion_InvalidJSON(t *testing.T) {
	controller, recorder, ctx := setupPushTest()

	// Request con JSON inválido
	ctx.Request = httptest.NewRequest("POST", "/push/enviar", bytes.NewBuffer([]byte("invalid json")))
	ctx.Request.Header.Set("Content-Type", "application/json")

	// Ejecutar
	controller.EnviarNotificacion()

	// Verificar
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestPushListarEnvios_Success(t *testing.T) {
	controller, recorder, _ := setupPushTest()

	// Configurar mocks
	dispositivo1 := &models.PushDispositivo{PkIdPushDispositivo: 1}
	dispositivo2 := &models.PushDispositivo{PkIdPushDispositivo: 2}
	envios := []*models.PushEnvio{
		{PkIdPushEnvio: 1, PkIdPushDispositivo: dispositivo1, Proveedor: "WEB_PUSH", Exito: true},
		{PkIdPushEnvio: 2, PkIdPushDispositivo: dispositivo2, Proveedor: "FCM", Exito: true},
	}

	mockPushOrm.On("QueryTable", "push_envio").Return(mockPushQS)
	mockPushQS.On("Count").Return(int64(2), nil)
	mockPushQS.On("OrderBy", []string{"-sent_at"}).Return(mockPushQS)
	mockPushQS.On("Limit", 20).Return(mockPushQS)
	mockPushQS.On("Offset", int64(0)).Return(mockPushQS)
	mockPushQS.On("All", mock.AnythingOfType("*[]*models.PushEnvio"), []string(nil)).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*[]*models.PushEnvio)
		*arg = envios
	}).Return(int64(2), nil)

	// Ejecutar
	controller.ListarEnvios()

	// Verificar
	assert.Equal(t, http.StatusOK, recorder.Code)
	mockPushOrm.AssertExpectations(t)
	mockPushQS.AssertExpectations(t)
}

func TestPushRegistrarEnvio_Success(t *testing.T) {
	t.Skip("TODO: Requiere refactorizar controller para inyectar PushService completo - el servicio usa orm.NewOrm() internamente y accede a la BD")
}

func TestPushRegistrarEnvio_InvalidJSON(t *testing.T) {
	controller, recorder, ctx := setupPushTest()

	// Request con JSON inválido
	ctx.Request = httptest.NewRequest("POST", "/push/envios", bytes.NewBuffer([]byte("invalid json")))
	ctx.Request.Header.Set("Content-Type", "application/json")

	// Ejecutar
	controller.RegistrarEnvio()

	// Verificar
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

// Test para cobertura de las interfaces adaptadoras
func TestPushAdapterInterfaces(t *testing.T) {
	t.Skip("TODO: Test de adaptadores requiere ORM real")
}
