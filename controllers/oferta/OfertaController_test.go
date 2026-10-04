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

var mockOfertOrmer *mockOfertaOrmer
var mockOfertQS *mockOfertaQuerySeter

func init() {
	ofertOrmNew = func() ofertaOrmer {
		return mockOfertOrmer
	}

	newOfertaService = func(o orm.Ormer) services.OfertaServiceInterface {
		return services.NewOfertaService(nil)
	}
}

func setupOfertaTest() (*OfertaController, *httptest.ResponseRecorder, *context.Context) {

	mockOfertOrmer = &mockOfertaOrmer{}
	mockOfertQS = &mockOfertaQuerySeter{}

	controller := &OfertaController{}
	controller.Controller = web.Controller{}

	recorder := httptest.NewRecorder()

	req := httptest.NewRequest("GET", "/", nil)

	ctx := context.NewContext()
	ctx.Reset(recorder, req)

	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	return controller, recorder, ctx
}

func TestOfertaGetAll_Success(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

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

	controller.GetAll()

	assert.Equal(t, http.StatusOK, recorder.Code)
	mockOfertOrmer.AssertExpectations(t)
	mockOfertQS.AssertExpectations(t)
}

func TestOfertaGetAll_WithFilters(t *testing.T) {
	controller, recorder, ctx := setupOfertaTest()

	ctx.Input.SetParam("activo", "true")
	ctx.Input.SetParam("restaurante_id", "1")
	ctx.Input.SetParam("titulo", "Oferta")
	ctx.Input.SetParam("limit", "10")
	ctx.Input.SetParam("offset", "5")

	mockOfertOrmer.On("QueryTable", "oferta").Return(mockOfertQS)
	mockOfertQS.On("Filter", "activo", []interface{}{true}).Return(mockOfertQS)
	mockOfertQS.On("Filter", "pk_id_restaurante", []interface{}{int64(1)}).Return(mockOfertQS)
	mockOfertQS.On("Filter", "titulo__icontains", []interface{}{"Oferta"}).Return(mockOfertQS)
	mockOfertQS.On("Count").Return(int64(1), nil)
	mockOfertQS.On("OrderBy", []string{"-pk_id_oferta"}).Return(mockOfertQS)
	mockOfertQS.On("Limit", 10).Return(mockOfertQS)
	mockOfertQS.On("Offset", int64(5)).Return(mockOfertQS)
	mockOfertQS.On("All", mock.AnythingOfType("*[]*models.Oferta"), []string(nil)).Return(int64(1), nil)

	controller.GetAll()

	assert.Equal(t, http.StatusOK, recorder.Code)
	mockOfertOrmer.AssertExpectations(t)
	mockOfertQS.AssertExpectations(t)
}

func TestOfertaGetAll_CountError(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	mockOfertOrmer.On("QueryTable", "oferta").Return(mockOfertQS)
	mockOfertQS.On("Count").Return(int64(0), fmt.Errorf("database error"))

	controller.GetAll()

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	mockOfertOrmer.AssertExpectations(t)
	mockOfertQS.AssertExpectations(t)
}

func TestOfertaPost_Success(t *testing.T) {
	controller, recorder, ctx := setupOfertaTest()

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
	ctx.Input.RequestBody = body

	mockOfertOrmer.On("Insert", mock.AnythingOfType("*models.Oferta")).Return(int64(1), nil)

	controller.Post()

	assert.Equal(t, http.StatusCreated, recorder.Code)
	mockOfertOrmer.AssertExpectations(t)
}

func TestOfertaPost_InvalidJSON(t *testing.T) {
	controller, recorder, ctx := setupOfertaTest()

	body := []byte("invalid json")
	ctx.Request = httptest.NewRequest("POST", "/ofertas", bytes.NewBuffer(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Input.RequestBody = body

	controller.Post()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestOfertaPost_ValidationError(t *testing.T) {
	controller, recorder, ctx := setupOfertaTest()

	payload := map[string]interface{}{
		"titulo":         "",
		"tipoDescuento":  "INVALID",
		"valorDescuento": 150,
		"fechaInicio":    "2025-01-01",
		"fechaFin":       "2024-12-31",
		"activo":         true,
	}

	body, _ := json.Marshal(payload)
	ctx.Request = httptest.NewRequest("POST", "/ofertas", bytes.NewBuffer(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Input.RequestBody = body

	controller.Post()

	assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
}

func TestOfertaGetById_Success(t *testing.T) {
	controller, recorder, ctx := setupOfertaTest()

	ctx.Request = httptest.NewRequest("GET", "/ofertas/search?id=1", nil)

	oferta := models.Oferta{PkIdOferta: 1, Titulo: "Oferta Test", TipoDescuento: "PORCENTAJE"}
	mockOfertOrmer.On("Read", mock.AnythingOfType("*models.Oferta"), []string(nil)).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.Oferta)
		*arg = oferta
	}).Return(nil)

	controller.GetById()

	assert.Equal(t, http.StatusOK, recorder.Code)
	mockOfertOrmer.AssertExpectations(t)
}

func TestOfertaGetById_InvalidID(t *testing.T) {
	controller, recorder, ctx := setupOfertaTest()

	ctx.Request = httptest.NewRequest("GET", "/ofertas/search?id=invalid", nil)

	controller.GetById()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestOfertaGetById_NotFound(t *testing.T) {
	controller, recorder, ctx := setupOfertaTest()

	ctx.Request = httptest.NewRequest("GET", "/ofertas/search?id=999", nil)

	mockOfertOrmer.On("Read", mock.AnythingOfType("*models.Oferta"), []string(nil)).Return(orm.ErrNoRows)

	controller.GetById()

	assert.Equal(t, http.StatusNotFound, recorder.Code)
	mockOfertOrmer.AssertExpectations(t)
}

func TestOfertaPut_Success(t *testing.T) {
	controller, recorder, ctx := setupOfertaTest()

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

	existingOferta := models.Oferta{PkIdOferta: 1, Titulo: "Oferta Vieja"}
	mockOfertOrmer.On("Read", mock.AnythingOfType("*models.Oferta"), []string(nil)).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.Oferta)
		*arg = existingOferta
	}).Return(nil)
	mockOfertOrmer.On("Update", mock.AnythingOfType("*models.Oferta"), []string(nil)).Return(int64(1), nil)

	controller.Put()

	assert.Equal(t, http.StatusOK, recorder.Code)
	mockOfertOrmer.AssertExpectations(t)
}

func TestOfertaDelete_Success(t *testing.T) {
	controller, recorder, ctx := setupOfertaTest()

	ctx.Request = httptest.NewRequest("DELETE", "/ofertas?id=1", nil)

	oferta := models.Oferta{PkIdOferta: 1, Titulo: "Oferta Test", Activo: true}
	mockOfertOrmer.On("Read", mock.AnythingOfType("*models.Oferta"), []string(nil)).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.Oferta)
		*arg = oferta
	}).Return(nil)

	mockOfertOrmer.On("Update", mock.AnythingOfType("*models.Oferta"), []string{"Activo"}).Return(int64(1), nil)

	controller.Delete()

	assert.Equal(t, http.StatusOK, recorder.Code)
	mockOfertOrmer.AssertExpectations(t)
}

func TestOfertaAsociarProducto_Success(t *testing.T) {
	controller, recorder, ctx := setupOfertaTest()

	request := map[string]interface{}{
		"productoId": 1,
	}

	body, _ := json.Marshal(request)
	ctx.Request = httptest.NewRequest("POST", "/ofertas/productos?id=1", bytes.NewBuffer(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Input.RequestBody = body

	oferta := models.Oferta{PkIdOferta: 1, Titulo: "Oferta Test"}
	producto := models.Producto{PK_ID_PRODUCTO: 1, NOMBRE: "Producto Test"}

	mockOfertOrmer.On("Read", mock.AnythingOfType("*models.Oferta"), []string(nil)).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.Oferta)
		*arg = oferta
	}).Return(nil).Once()

	mockOfertOrmer.On("Read", mock.AnythingOfType("*models.Producto"), []string(nil)).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.Producto)
		*arg = producto
	}).Return(nil).Once()

	mockOfertOrmer.On("Insert", mock.AnythingOfType("*models.OfertaProducto")).Return(int64(1), nil)

	controller.AsociarProducto()

	assert.Equal(t, http.StatusCreated, recorder.Code)
	mockOfertOrmer.AssertExpectations(t)
}

func TestOfertaAsociarProducto_InvalidID(t *testing.T) {
	controller, recorder, ctx := setupOfertaTest()

	ctx.Request = httptest.NewRequest("POST", "/ofertas/productos?id=invalid", nil)

	controller.AsociarProducto()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestOfertaDesasociarProducto_Success(t *testing.T) {
	controller, recorder, ctx := setupOfertaTest()

	ctx.Request = httptest.NewRequest("DELETE", "/ofertas/productos?id=1&producto_id=1", nil)

	mockOfertOrmer.On("QueryTable", "oferta_producto").Return(mockOfertQS)
	mockOfertQS.On("Filter", "pk_id_oferta", mock.Anything).Return(mockOfertQS)
	mockOfertQS.On("Filter", "pk_id_producto", mock.Anything).Return(mockOfertQS)
	mockOfertQS.On("One", mock.AnythingOfType("*models.OfertaProducto")).Return(nil)
	mockOfertOrmer.On("Delete", mock.AnythingOfType("*models.OfertaProducto"), []string(nil)).Return(int64(1), nil)

	controller.DesasociarProducto()

	assert.Equal(t, http.StatusOK, recorder.Code)
	mockOfertOrmer.AssertExpectations(t)
	mockOfertQS.AssertExpectations(t)
}

func TestOfertaDesasociarProducto_InvalidIDs(t *testing.T) {
	controller, recorder, ctx := setupOfertaTest()

	ctx.Request = httptest.NewRequest("DELETE", "/ofertas/productos?id=invalid&producto_id=invalid", nil)

	controller.DesasociarProducto()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}
