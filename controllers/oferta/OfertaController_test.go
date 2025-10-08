package oferta

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

// Mock para ofertaQuerySeter
type mockOfertaQuerySeter struct {
	mock.Mock
}

func (m *mockOfertaQuerySeter) All(container interface{}, cols ...string) (int64, error) {
	args := m.Called(container, cols)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockOfertaQuerySeter) Filter(expr string, args ...interface{}) ofertaQuerySeter {
	m.Called(expr, args)
	return m
}

func (m *mockOfertaQuerySeter) OrderBy(exprs ...string) ofertaQuerySeter {
	m.Called(exprs)
	return m
}

func (m *mockOfertaQuerySeter) Limit(limit int) ofertaQuerySeter {
	m.Called(limit)
	return m
}

func (m *mockOfertaQuerySeter) Offset(offset int64) ofertaQuerySeter {
	m.Called(offset)
	return m
}

func (m *mockOfertaQuerySeter) Count() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockOfertaQuerySeter) One(container interface{}) error {
	args := m.Called(container)
	return args.Error(0)
}

// Mock para ofertaOrmer
type mockOfertaOrmer struct {
	mock.Mock
}

func (m *mockOfertaOrmer) QueryTable(ptrStructOrTableName interface{}) ofertaQuerySeter {
	args := m.Called(ptrStructOrTableName)
	return args.Get(0).(ofertaQuerySeter)
}

func (m *mockOfertaOrmer) Insert(md interface{}) (int64, error) {
	args := m.Called(md)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockOfertaOrmer) Read(md interface{}, cols ...string) error {
	args := m.Called(md, cols)
	return args.Error(0)
}

func (m *mockOfertaOrmer) Update(md interface{}, cols ...string) (int64, error) {
	args := m.Called(md, cols)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockOfertaOrmer) Delete(md interface{}, cols ...string) (int64, error) {
	args := m.Called(md, cols)
	return args.Get(0).(int64), args.Error(1)
}

// Variables globales para mocks
var mockOfertOrmer *mockOfertaOrmer
var mockOfertQS *mockOfertaQuerySeter

// Override de las funciones de creación
func init() {
	ofertOrmNew = func() ofertaOrmer {
		return mockOfertOrmer
	}
	// Mockear el servicio - ValidarReglasNegocioOferta no usa el ORM, así que podemos pasar nil
	newOfertaService = func(o orm.Ormer) services.OfertaServiceInterface {
		return services.NewOfertaService(nil)
	}
}

func setupOfertaTest() (*OfertaController, *httptest.ResponseRecorder, *context.Context) {
	// Reset mocks
	mockOfertOrmer = &mockOfertaOrmer{}
	mockOfertQS = &mockOfertaQuerySeter{}

	// Crear controller
	controller := &OfertaController{}
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

func TestOfertaGetAll_Success(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	// Configurar mocks
	ofertas := []*models.Oferta{
		{PkIdOferta: 1, Titulo: "Oferta 1", TipoDescuento: "PORCENTAJE", ValorDescuento: 20, Activo: true},
		{PkIdOferta: 2, Titulo: "Oferta 2", TipoDescuento: "MONTO", ValorDescuento: 5000, Activo: true},
	}

	mockOfertOrmer.On("QueryTable", "oferta").Return(mockOfertQS)
	mockOfertQS.On("Count").Return(int64(2), nil)
	mockOfertQS.On("OrderBy", []string{"-pk_id_oferta"}).Return(mockOfertQS)
	mockOfertQS.On("Limit", 20).Return(mockOfertQS)
	mockOfertQS.On("Offset", int64(0)).Return(mockOfertQS)
	mockOfertQS.On("All", mock.AnythingOfType("*[]*models.Oferta"), []string(nil)).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*[]*models.Oferta)
		*arg = ofertas
	}).Return(int64(2), nil)

	// Ejecutar
	controller.GetAll()

	// Verificar
	assert.Equal(t, http.StatusOK, recorder.Code)
	mockOfertOrmer.AssertExpectations(t)
	mockOfertQS.AssertExpectations(t)
}

func TestOfertaGetAll_WithFilters(t *testing.T) {
	controller, recorder, ctx := setupOfertaTest()

	// Configurar query parameters
	ctx.Input.SetParam("activo", "true")
	ctx.Input.SetParam("restaurante_id", "1")
	ctx.Input.SetParam("titulo", "Oferta")
	ctx.Input.SetParam("limit", "10")
	ctx.Input.SetParam("offset", "5")

	// Configurar mocks
	mockOfertOrmer.On("QueryTable", "oferta").Return(mockOfertQS)
	mockOfertQS.On("Filter", "activo", []interface{}{true}).Return(mockOfertQS)
	mockOfertQS.On("Filter", "pk_id_restaurante", []interface{}{int64(1)}).Return(mockOfertQS)
	mockOfertQS.On("Filter", "titulo__icontains", []interface{}{"Oferta"}).Return(mockOfertQS)
	mockOfertQS.On("Count").Return(int64(1), nil)
	mockOfertQS.On("OrderBy", []string{"-pk_id_oferta"}).Return(mockOfertQS)
	mockOfertQS.On("Limit", 10).Return(mockOfertQS)
	mockOfertQS.On("Offset", int64(5)).Return(mockOfertQS)
	mockOfertQS.On("All", mock.AnythingOfType("*[]*models.Oferta"), []string(nil)).Return(int64(1), nil)

	// Ejecutar
	controller.GetAll()

	// Verificar
	assert.Equal(t, http.StatusOK, recorder.Code)
	mockOfertOrmer.AssertExpectations(t)
	mockOfertQS.AssertExpectations(t)
}

func TestOfertaGetAll_CountError(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	// Configurar mocks
	mockOfertOrmer.On("QueryTable", "oferta").Return(mockOfertQS)
	mockOfertQS.On("Count").Return(int64(0), fmt.Errorf("database error"))

	// Ejecutar
	controller.GetAll()

	// Verificar
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	mockOfertOrmer.AssertExpectations(t)
	mockOfertQS.AssertExpectations(t)
}

func TestOfertaPost_Success(t *testing.T) {
	controller, recorder, ctx := setupOfertaTest()

	// Preparar request body usando map para controlar formato de fechas
	payload := map[string]interface{}{
		"titulo":          "Nueva Oferta",
		"tipoDescuento":   "PORCENTAJE",
		"valorDescuento":  25,
		"fechaInicio":     "2025-01-01",
		"fechaFin":        "2025-12-31",
		"diasSemana":      []string{"Lunes", "Martes"},
		"activo":          true,
		"pkIdRestaurante": 1,
	}

	body, _ := json.Marshal(payload)
	ctx.Request = httptest.NewRequest("POST", "/ofertas", bytes.NewBuffer(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Input.RequestBody = body // IMPORTANTE: establecer RequestBody

	// Configurar mocks
	mockOfertOrmer.On("Insert", mock.AnythingOfType("*models.Oferta")).Return(int64(1), nil)

	// Ejecutar
	controller.Post()

	// Verificar
	assert.Equal(t, http.StatusCreated, recorder.Code)
	mockOfertOrmer.AssertExpectations(t)
}

func TestOfertaPost_InvalidJSON(t *testing.T) {
	controller, recorder, ctx := setupOfertaTest()

	// Request con JSON inválido
	body := []byte("invalid json")
	ctx.Request = httptest.NewRequest("POST", "/ofertas", bytes.NewBuffer(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Input.RequestBody = body

	// Ejecutar
	controller.Post()

	// Verificar
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestOfertaPost_ValidationError(t *testing.T) {
	controller, recorder, ctx := setupOfertaTest()

	// Oferta con datos inválidos
	payload := map[string]interface{}{
		"titulo":         "", // Título vacío
		"tipoDescuento":  "INVALID",
		"valorDescuento": 150, // Porcentaje inválido
		"fechaInicio":    "2025-01-01",
		"fechaFin":       "2024-12-31", // Fecha fin antes que inicio
		"activo":         true,
	}

	body, _ := json.Marshal(payload)
	ctx.Request = httptest.NewRequest("POST", "/ofertas", bytes.NewBuffer(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Input.RequestBody = body

	// Ejecutar
	controller.Post()

	// Verificar
	assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
}

func TestOfertaGetById_Success(t *testing.T) {
	controller, recorder, ctx := setupOfertaTest()

	// Configurar parámetro ID usando query string
	ctx.Request = httptest.NewRequest("GET", "/ofertas/search?id=1", nil)

	// Configurar mock
	oferta := models.Oferta{PkIdOferta: 1, Titulo: "Oferta Test", TipoDescuento: "PORCENTAJE"}
	mockOfertOrmer.On("Read", mock.AnythingOfType("*models.Oferta"), []string(nil)).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.Oferta)
		*arg = oferta
	}).Return(nil)

	// Ejecutar
	controller.GetById()

	// Verificar
	assert.Equal(t, http.StatusOK, recorder.Code)
	mockOfertOrmer.AssertExpectations(t)
}

func TestOfertaGetById_InvalidID(t *testing.T) {
	controller, recorder, ctx := setupOfertaTest()

	// ID inválido usando query string
	ctx.Request = httptest.NewRequest("GET", "/ofertas/search?id=invalid", nil)

	// Ejecutar
	controller.GetById()

	// Verificar
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestOfertaGetById_NotFound(t *testing.T) {
	controller, recorder, ctx := setupOfertaTest()

	// Configurar parámetro ID usando query string
	ctx.Request = httptest.NewRequest("GET", "/ofertas/search?id=999", nil)

	// Configurar mock para no encontrado
	mockOfertOrmer.On("Read", mock.AnythingOfType("*models.Oferta"), []string(nil)).Return(orm.ErrNoRows)

	// Ejecutar
	controller.GetById()

	// Verificar
	assert.Equal(t, http.StatusNotFound, recorder.Code)
	mockOfertOrmer.AssertExpectations(t)
}

func TestOfertaPut_Success(t *testing.T) {
	controller, recorder, ctx := setupOfertaTest()

	// Preparar request body usando map
	payload := map[string]interface{}{
		"titulo":          "Oferta Actualizada",
		"tipoDescuento":   "PORCENTAJE",
		"valorDescuento":  30,
		"fechaInicio":     "2025-01-01",
		"fechaFin":        "2025-12-31",
		"diasSemana":      []string{"Lunes"},
		"activo":          true,
		"pkIdRestaurante": 1,
	}

	body, _ := json.Marshal(payload)
	ctx.Request = httptest.NewRequest("PUT", "/ofertas?id=1", bytes.NewBuffer(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Input.RequestBody = body

	// Configurar mocks
	existingOferta := models.Oferta{PkIdOferta: 1, Titulo: "Oferta Vieja"}
	mockOfertOrmer.On("Read", mock.AnythingOfType("*models.Oferta"), []string(nil)).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.Oferta)
		*arg = existingOferta
	}).Return(nil)
	mockOfertOrmer.On("Update", mock.AnythingOfType("*models.Oferta"), []string(nil)).Return(int64(1), nil)

	// Ejecutar
	controller.Put()

	// Verificar
	assert.Equal(t, http.StatusOK, recorder.Code)
	mockOfertOrmer.AssertExpectations(t)
}

func TestOfertaDelete_Success(t *testing.T) {
	controller, recorder, ctx := setupOfertaTest()

	// Configurar parámetro ID usando query string
	ctx.Request = httptest.NewRequest("DELETE", "/ofertas?id=1", nil)

	// Configurar mocks
	oferta := models.Oferta{PkIdOferta: 1, Titulo: "Oferta Test", Activo: true}
	mockOfertOrmer.On("Read", mock.AnythingOfType("*models.Oferta"), []string(nil)).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.Oferta)
		*arg = oferta
	}).Return(nil)
	// Delete hace soft delete con Update
	mockOfertOrmer.On("Update", mock.AnythingOfType("*models.Oferta"), []string{"Activo"}).Return(int64(1), nil)

	// Ejecutar
	controller.Delete()

	// Verificar
	assert.Equal(t, http.StatusOK, recorder.Code)
	mockOfertOrmer.AssertExpectations(t)
}

func TestOfertaAsociarProducto_Success(t *testing.T) {
	controller, recorder, ctx := setupOfertaTest()

	// Preparar request body
	request := map[string]interface{}{
		"productoId": 1,
	}

	body, _ := json.Marshal(request)
	ctx.Request = httptest.NewRequest("POST", "/ofertas/productos?id=1", bytes.NewBuffer(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Input.RequestBody = body

	// Configurar mocks - el controller lee tanto la oferta como el producto
	oferta := models.Oferta{PkIdOferta: 1, Titulo: "Oferta Test"}
	producto := models.Producto{PK_ID_PRODUCTO: 1, NOMBRE: "Producto Test"}

	// Primera llamada a Read es para la oferta
	mockOfertOrmer.On("Read", mock.AnythingOfType("*models.Oferta"), []string(nil)).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.Oferta)
		*arg = oferta
	}).Return(nil).Once()

	// Segunda llamada a Read es para el producto
	mockOfertOrmer.On("Read", mock.AnythingOfType("*models.Producto"), []string(nil)).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.Producto)
		*arg = producto
	}).Return(nil).Once()

	mockOfertOrmer.On("Insert", mock.AnythingOfType("*models.OfertaProducto")).Return(int64(1), nil)

	// Ejecutar
	controller.AsociarProducto()

	// Verificar
	assert.Equal(t, http.StatusCreated, recorder.Code)
	mockOfertOrmer.AssertExpectations(t)
}

func TestOfertaAsociarProducto_InvalidID(t *testing.T) {
	controller, recorder, ctx := setupOfertaTest()

	// ID inválido usando query string
	ctx.Request = httptest.NewRequest("POST", "/ofertas/productos?id=invalid", nil)

	// Ejecutar
	controller.AsociarProducto()

	// Verificar
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestOfertaDesasociarProducto_Success(t *testing.T) {
	controller, recorder, ctx := setupOfertaTest()

	// Configurar parámetros usando query string
	ctx.Request = httptest.NewRequest("DELETE", "/ofertas/productos?id=1&producto_id=1", nil)

	// Configurar mocks - el controller usa QueryTable para buscar la asociación
	mockOfertOrmer.On("QueryTable", "oferta_producto").Return(mockOfertQS)
	mockOfertQS.On("Filter", "pk_id_oferta", mock.Anything).Return(mockOfertQS)
	mockOfertQS.On("Filter", "pk_id_producto", mock.Anything).Return(mockOfertQS)
	mockOfertQS.On("One", mock.AnythingOfType("*models.OfertaProducto")).Return(nil)
	mockOfertOrmer.On("Delete", mock.AnythingOfType("*models.OfertaProducto"), []string(nil)).Return(int64(1), nil)

	// Ejecutar
	controller.DesasociarProducto()

	// Verificar
	assert.Equal(t, http.StatusOK, recorder.Code)
	mockOfertOrmer.AssertExpectations(t)
	mockOfertQS.AssertExpectations(t)
}

func TestOfertaDesasociarProducto_InvalidIDs(t *testing.T) {
	controller, recorder, ctx := setupOfertaTest()

	// IDs inválidos usando query string
	ctx.Request = httptest.NewRequest("DELETE", "/ofertas/productos?id=invalid&producto_id=invalid", nil)

	// Ejecutar
	controller.DesasociarProducto()

	// Verificar
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

// Test para cobertura de las interfaces adaptadoras - eliminado porque no aporta valor real
