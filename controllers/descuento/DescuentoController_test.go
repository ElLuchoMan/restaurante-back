package descuento

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"restaurante/models"
	"restaurante/services"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock para descuentoQuerySeter
type mockDescuentoQuerySeter struct {
	mock.Mock
}

func (m *mockDescuentoQuerySeter) All(container interface{}, cols ...string) (int64, error) {
	args := m.Called(container, cols)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockDescuentoQuerySeter) Filter(expr string, args ...interface{}) descuentoQuerySeter {
	m.Called(expr, args)
	return m
}

func (m *mockDescuentoQuerySeter) OrderBy(exprs ...string) descuentoQuerySeter {
	m.Called(exprs)
	return m
}

func (m *mockDescuentoQuerySeter) Limit(limit int) descuentoQuerySeter {
	m.Called(limit)
	return m
}

func (m *mockDescuentoQuerySeter) Offset(offset int64) descuentoQuerySeter {
	m.Called(offset)
	return m
}

func (m *mockDescuentoQuerySeter) Count() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockDescuentoQuerySeter) One(container interface{}) error {
	args := m.Called(container)
	return args.Error(0)
}

// Mock para descuentoOrmer
type mockDescuentoOrmer struct {
	mock.Mock
}

func (m *mockDescuentoOrmer) QueryTable(ptrStructOrTableName interface{}) descuentoQuerySeter {
	args := m.Called(ptrStructOrTableName)
	return args.Get(0).(descuentoQuerySeter)
}

func (m *mockDescuentoOrmer) Insert(md interface{}) (int64, error) {
	args := m.Called(md)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockDescuentoOrmer) Read(md interface{}, cols ...string) error {
	args := m.Called(md, cols)
	return args.Error(0)
}

func (m *mockDescuentoOrmer) Update(md interface{}, cols ...string) (int64, error) {
	args := m.Called(md, cols)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockDescuentoOrmer) Delete(md interface{}, cols ...string) (int64, error) {
	args := m.Called(md, cols)
	return args.Get(0).(int64), args.Error(1)
}

// Variables globales para mocks
var mockDescOrmer *mockDescuentoOrmer
var mockDescQS *mockDescuentoQuerySeter

// Override de las funciones de creación
func init() {
	descOrmNew = func() descuentoOrmer {
		return mockDescOrmer
	}
	// Mockear el servicio - los métodos del servicio no usan el ORM directamente en este caso
	newDescuentoService = func(o orm.Ormer) *services.DescuentoService {
		return services.NewDescuentoService(nil)
	}
}

func setupDescuentoTest() (*DescuentoController, *httptest.ResponseRecorder, *context.Context) {
	// Reset mocks
	mockDescOrmer = &mockDescuentoOrmer{}
	mockDescQS = &mockDescuentoQuerySeter{}

	// Crear controller
	controller := &DescuentoController{}
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

func TestDescuentoGetAll_Success(t *testing.T) {
	t.Skip("TODO: Requiere refactorizar controller para inyectar DescuentoService completo - el servicio llama a orm.NewOrm() directamente")
	controller, recorder, ctx := setupDescuentoTest()

	// Configurar query parameter
	ctx.Input.SetParam("pedido_id", "1")

	// Configurar mock para leer pedido
	pedido := models.Pedido{PK_ID_PEDIDO: 1}
	mockDescOrmer.On("Read", mock.AnythingOfType("*models.Pedido"), []string(nil)).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.Pedido)
		*arg = pedido
	}).Return(nil)

	// Ejecutar
	controller.GetAll()

	// Verificar
	assert.Equal(t, http.StatusOK, recorder.Code)
	mockDescOrmer.AssertExpectations(t)
}

func TestDescuentoGetAll_InvalidPedidoId(t *testing.T) {
	controller, recorder, ctx := setupDescuentoTest()

	// ID inválido usando query string
	ctx.Request = httptest.NewRequest("GET", "/descuentos/pedidos?pedido_id=invalid", nil)

	// Ejecutar
	controller.GetAll()

	// Verificar
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestDescuentoGetAll_MissingPedidoId(t *testing.T) {
	controller, recorder, _ := setupDescuentoTest()

	// Sin pedido_id
	// Ejecutar
	controller.GetAll()

	// Verificar
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestDescuentoGetAll_PedidoNotFound(t *testing.T) {
	controller, recorder, ctx := setupDescuentoTest()

	// Configurar query parameter usando query string
	ctx.Request = httptest.NewRequest("GET", "/descuentos/pedidos?pedido_id=999", nil)

	// Configurar mock para pedido no encontrado
	mockDescOrmer.On("Read", mock.AnythingOfType("*models.Pedido"), []string(nil)).Return(orm.ErrNoRows)

	// Ejecutar
	controller.GetAll()

	// Verificar
	assert.Equal(t, http.StatusNotFound, recorder.Code) // Debe devolver 404 cuando el pedido no existe
	mockDescOrmer.AssertExpectations(t)
}

func TestDescuentoGetAll_DatabaseError(t *testing.T) {
	controller, recorder, ctx := setupDescuentoTest()

	// Configurar query parameter usando query string
	ctx.Request = httptest.NewRequest("GET", "/descuentos/pedidos?pedido_id=1", nil)

	// Configurar mock para error de base de datos
	mockDescOrmer.On("Read", mock.AnythingOfType("*models.Pedido"), []string(nil)).Return(fmt.Errorf("database error"))

	// Ejecutar
	controller.GetAll()

	// Verificar
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	mockDescOrmer.AssertExpectations(t)
}

func TestDescuentoPost_Success(t *testing.T) {
	t.Skip("TODO: Requiere refactorizar controller para inyectar DescuentoService completo - el servicio llama a orm.NewOrm() directamente")
	controller, recorder, ctx := setupDescuentoTest()

	// Preparar request body
	cuponId := int64(1)
	payload := map[string]interface{}{
		"cuponId":        cuponId,
		"montoDescuento": 5000,
		"detalle":        map[string]string{"tipo": "cupon", "codigo": "TEST"},
	}

	body, _ := json.Marshal(payload)
	ctx.Request = httptest.NewRequest("POST", "/descuentos/pedidos?pedido_id=1", bytes.NewBuffer(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Input.RequestBody = body // IMPORTANTE: establecer RequestBody

	// Ejecutar
	controller.Post()

	// Verificar
	assert.Equal(t, http.StatusCreated, recorder.Code)
}

func TestDescuentoPost_InvalidJSON(t *testing.T) {
	controller, recorder, ctx := setupDescuentoTest()

	// Request con JSON inválido
	body := []byte("invalid json")
	ctx.Request = httptest.NewRequest("POST", "/descuentos/pedidos?pedido_id=1", bytes.NewBuffer(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Input.RequestBody = body

	// Ejecutar
	controller.Post()

	// Verificar
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestDescuentoPost_ValidationError(t *testing.T) {
	controller, recorder, ctx := setupDescuentoTest()

	// Descuento con datos inválidos - sin cuponId ni ofertaId
	payload := map[string]interface{}{
		"montoDescuento": 5000,
	}

	body, _ := json.Marshal(payload)
	ctx.Request = httptest.NewRequest("POST", "/descuentos/pedidos?pedido_id=1", bytes.NewBuffer(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Input.RequestBody = body

	// Ejecutar
	controller.Post()

	// Verificar
	assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
}

func TestDescuentoPost_DatabaseError(t *testing.T) {
	t.Skip("TODO: Requiere refactorizar controller para inyectar DescuentoService completo - el servicio llama a orm.NewOrm() directamente")
	controller, recorder, ctx := setupDescuentoTest()

	// Preparar request body válido
	cuponId := int64(1)
	payload := map[string]interface{}{
		"cuponId":        cuponId,
		"montoDescuento": 5000,
		"detalle":        map[string]string{"tipo": "cupon", "codigo": "TEST"},
	}

	body, _ := json.Marshal(payload)
	ctx.Request = httptest.NewRequest("POST", "/descuentos/pedidos?pedido_id=1", bytes.NewBuffer(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Input.RequestBody = body

	// Ejecutar
	controller.Post()

	// Verificar
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}

// Test para cobertura de las interfaces adaptadoras
func TestDescuentoAdapterInterfaces(t *testing.T) {
	t.Skip("TODO: Test de adaptadores requiere refactorización - los adaptadores necesitan un ORM real para funcionar")
}
