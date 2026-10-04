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

var mockPushOrm *mockPushOrmer
var mockPushQS *mockPushQuerySeter

func init() {
	pushOrmNew = func() pushOrmer {
		return mockPushOrm
	}

}

func setupPushTest() (*PushController, *httptest.ResponseRecorder, *context.Context) {

	mockPushOrm = &mockPushOrmer{}
	mockPushQS = &mockPushQuerySeter{}

	controller := &PushController{}
	controller.Controller = web.Controller{}

	recorder := httptest.NewRecorder()

	req := httptest.NewRequest("GET", "/", nil)

	ctx := context.NewContext()
	ctx.Reset(recorder, req)

	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	return controller, recorder, ctx
}

func TestPushGetAll_Success(t *testing.T) {
	controller, recorder, _ := setupPushTest()

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

	controller.GetAll()

	assert.Equal(t, http.StatusOK, recorder.Code)
	mockPushOrm.AssertExpectations(t)
	mockPushQS.AssertExpectations(t)
}

func TestPushGetAll_WithFilters(t *testing.T) {
	controller, recorder, ctx := setupPushTest()

	ctx.Input.SetParam("cliente_id", "123")
	ctx.Input.SetParam("trabajador_id", "456")
	ctx.Input.SetParam("plataforma", "WEB")
	ctx.Input.SetParam("limit", "10")
	ctx.Input.SetParam("offset", "5")

	mockPushOrm.On("QueryTable", "push_dispositivo").Return(mockPushQS)
	mockPushQS.On("Filter", "pk_documento_cliente", []interface{}{int64(123)}).Return(mockPushQS)
	mockPushQS.On("Filter", "pk_documento_trabajador", []interface{}{int64(456)}).Return(mockPushQS)
	mockPushQS.On("Filter", "plataforma", []interface{}{"WEB"}).Return(mockPushQS)
	mockPushQS.On("Count").Return(int64(1), nil)
	mockPushQS.On("OrderBy", []string{"-created_at"}).Return(mockPushQS)
	mockPushQS.On("Limit", 10).Return(mockPushQS)
	mockPushQS.On("Offset", int64(5)).Return(mockPushQS)
	mockPushQS.On("All", mock.AnythingOfType("*[]*models.PushDispositivo"), []string(nil)).Return(int64(1), nil)

	controller.GetAll()

	assert.Equal(t, http.StatusOK, recorder.Code)
	mockPushOrm.AssertExpectations(t)
	mockPushQS.AssertExpectations(t)
}

func TestPushGetAll_CountError(t *testing.T) {
	controller, recorder, _ := setupPushTest()

	mockPushOrm.On("QueryTable", "push_dispositivo").Return(mockPushQS)
	mockPushQS.On("Count").Return(int64(0), fmt.Errorf("database error"))

	controller.GetAll()

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	mockPushOrm.AssertExpectations(t)
	mockPushQS.AssertExpectations(t)
}

func TestPushPost_InvalidJSON(t *testing.T) {
	controller, recorder, ctx := setupPushTest()

	ctx.Request = httptest.NewRequest("POST", "/push/dispositivos", bytes.NewBuffer([]byte("invalid json")))
	ctx.Request.Header.Set("Content-Type", "application/json")

	controller.Post()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestPushPost_ValidationError(t *testing.T) {
	controller, recorder, ctx := setupPushTest()

	dispositivo := models.PushDispositivo{
		Plataforma: "INVALID",
		Enabled:    true,
	}

	body, _ := json.Marshal(dispositivo)
	ctx.Request = httptest.NewRequest("POST", "/push/dispositivos", bytes.NewBuffer(body))
	ctx.Request.Header.Set("Content-Type", "application/json")

	controller.Post()

	assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
}

func TestPushGetById_Success(t *testing.T) {
	controller, recorder, ctx := setupPushTest()

	ctx.Input.SetParam(":id", "1")

	dispositivo := models.PushDispositivo{PkIdPushDispositivo: 1, Plataforma: "WEB", Enabled: true}
	mockPushOrm.On("Read", mock.AnythingOfType("*models.PushDispositivo"), []string(nil)).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.PushDispositivo)
		*arg = dispositivo
	}).Return(nil)

	controller.GetById()

	assert.Equal(t, http.StatusOK, recorder.Code)
	mockPushOrm.AssertExpectations(t)
}

func TestPushGetById_InvalidID(t *testing.T) {
	controller, recorder, ctx := setupPushTest()

	ctx.Input.SetParam(":id", "invalid")

	controller.GetById()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestPushGetById_NotFound(t *testing.T) {
	controller, recorder, ctx := setupPushTest()

	ctx.Request = httptest.NewRequest("GET", "/push/dispositivos/search?id=999", nil)

	mockPushOrm.On("Read", mock.AnythingOfType("*models.PushDispositivo"), []string(nil)).Return(orm.ErrNoRows)

	controller.GetById()

	assert.Equal(t, http.StatusNotFound, recorder.Code)
	mockPushOrm.AssertExpectations(t)
}

func TestPushDelete_Success(t *testing.T) {
	controller, recorder, ctx := setupPushTest()

	ctx.Input.SetParam(":id", "1")

	dispositivo := models.PushDispositivo{PkIdPushDispositivo: 1, Plataforma: "WEB"}
	mockPushOrm.On("Read", mock.AnythingOfType("*models.PushDispositivo"), []string(nil)).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.PushDispositivo)
		*arg = dispositivo
	}).Return(nil)
	mockPushOrm.On("Delete", mock.AnythingOfType("*models.PushDispositivo"), []string(nil)).Return(int64(1), nil)

	controller.Delete()

	assert.Equal(t, http.StatusOK, recorder.Code)
	mockPushOrm.AssertExpectations(t)
}

func TestPushActualizarUltimaVista_InvalidID(t *testing.T) {
	controller, recorder, ctx := setupPushTest()

	ctx.Input.SetParam(":id", "invalid")

	controller.ActualizarUltimaVista()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestPushActualizarTopics_InvalidJSON(t *testing.T) {
	controller, recorder, ctx := setupPushTest()

	ctx.Input.SetParam(":id", "1")

	ctx.Request = httptest.NewRequest("PATCH", "/push/dispositivos/1/topics", bytes.NewBuffer([]byte("invalid json")))
	ctx.Request.Header.Set("Content-Type", "application/json")

	controller.ActualizarTopics()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestPushEnviarNotificacion_InvalidJSON(t *testing.T) {
	controller, recorder, ctx := setupPushTest()

	ctx.Request = httptest.NewRequest("POST", "/push/enviar", bytes.NewBuffer([]byte("invalid json")))
	ctx.Request.Header.Set("Content-Type", "application/json")

	controller.EnviarNotificacion()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestPushListarEnvios_Success(t *testing.T) {
	controller, recorder, _ := setupPushTest()

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

	controller.ListarEnvios()

	assert.Equal(t, http.StatusOK, recorder.Code)
	mockPushOrm.AssertExpectations(t)
	mockPushQS.AssertExpectations(t)
}

func TestPushRegistrarEnvio_InvalidJSON(t *testing.T) {
	controller, recorder, ctx := setupPushTest()

	ctx.Request = httptest.NewRequest("POST", "/push/envios", bytes.NewBuffer([]byte("invalid json")))
	ctx.Request.Header.Set("Content-Type", "application/json")

	controller.RegistrarEnvio()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}
