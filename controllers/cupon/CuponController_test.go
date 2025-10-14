package cupon

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"restaurante/models"
	"restaurante/services"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockCuponQuerySeter struct {
	mock.Mock
}

func (m *mockCuponQuerySeter) All(container interface{}, cols ...string) (int64, error) {
	args := m.Called(container, cols)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockCuponQuerySeter) Filter(expr string, args ...interface{}) cuponQuerySeter {
	m.Called(expr, args)
	return m
}

func (m *mockCuponQuerySeter) OrderBy(exprs ...string) cuponQuerySeter {
	m.Called(exprs)
	return m
}

func (m *mockCuponQuerySeter) Limit(limit int) cuponQuerySeter {
	m.Called(limit)
	return m
}

func (m *mockCuponQuerySeter) Offset(offset int64) cuponQuerySeter {
	m.Called(offset)
	return m
}

func (m *mockCuponQuerySeter) Count() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockCuponQuerySeter) One(container interface{}) error {
	args := m.Called(container)
	return args.Error(0)
}

type mockCuponOrmer struct {
	mock.Mock
}

func (m *mockCuponOrmer) QueryTable(ptrStructOrTableName interface{}) cuponQuerySeter {
	args := m.Called(ptrStructOrTableName)
	return args.Get(0).(cuponQuerySeter)
}

func (m *mockCuponOrmer) Insert(md interface{}) (int64, error) {
	args := m.Called(md)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockCuponOrmer) Read(md interface{}, cols ...string) error {
	args := m.Called(md, cols)
	return args.Error(0)
}

func (m *mockCuponOrmer) Update(md interface{}, cols ...string) (int64, error) {
	args := m.Called(md, cols)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockCuponOrmer) Delete(md interface{}, cols ...string) (int64, error) {
	args := m.Called(md, cols)
	return args.Get(0).(int64), args.Error(1)
}

var mockOrmer *mockCuponOrmer
var mockQS *mockCuponQuerySeter

func init() {
	cupOrmNew = func() cuponOrmer {
		return mockOrmer
	}

	newCuponService = func(o orm.Ormer) *services.CuponService {
		return services.NewCuponService(nil)
	}
}

func setupTest() (*CuponController, *httptest.ResponseRecorder, *context.Context) {

	mockOrmer = &mockCuponOrmer{}
	mockQS = &mockCuponQuerySeter{}

	controller := &CuponController{}
	controller.Controller = web.Controller{}

	recorder := httptest.NewRecorder()

	req := httptest.NewRequest("GET", "/", nil)

	ctx := context.NewContext()
	ctx.Reset(recorder, req)

	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	return controller, recorder, ctx
}

func TestGetAll_Success(t *testing.T) {
	controller, recorder, _ := setupTest()

	cupones := []models.Cupon{
		{PkIdCupon: 1, Codigo: "TEST1", Scope: "GLOBAL", TipoDescuento: "PORCENTAJE", ValorDescuento: 10, Activo: true},
		{PkIdCupon: 2, Codigo: "TEST2", Scope: "PRODUCTO", TipoDescuento: "MONTO", ValorDescuento: 5000, Activo: true},
	}

	mockOrmer.On("QueryTable", "cupon").Return(mockQS)
	mockQS.On("Count").Return(int64(2), nil)
	mockQS.On("OrderBy", []string{"-pk_id_cupon"}).Return(mockQS)
	mockQS.On("Limit", 20).Return(mockQS)
	mockQS.On("Offset", int64(0)).Return(mockQS)
	mockQS.On("All", mock.AnythingOfType("*[]*models.Cupon"), []string(nil)).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*[]*models.Cupon)
		*arg = []*models.Cupon{&cupones[0], &cupones[1]}
	}).Return(int64(2), nil)

	controller.GetAll()

	assert.Equal(t, http.StatusOK, recorder.Code)
	mockOrmer.AssertExpectations(t)
	mockQS.AssertExpectations(t)
}

func TestGetAll_WithFilters(t *testing.T) {
	controller, recorder, ctx := setupTest()

	ctx.Input.SetParam("activo", "true")
	ctx.Input.SetParam("codigo", "TEST")
	ctx.Input.SetParam("scope", "GLOBAL")
	ctx.Input.SetParam("fecha_desde", "2025-01-01")
	ctx.Input.SetParam("fecha_hasta", "2025-12-31")
	ctx.Input.SetParam("limit", "10")
	ctx.Input.SetParam("offset", "5")

	mockOrmer.On("QueryTable", "cupon").Return(mockQS)
	mockQS.On("Filter", "activo", []interface{}{true}).Return(mockQS)
	mockQS.On("Filter", "codigo__icontains", []interface{}{"TEST"}).Return(mockQS)
	mockQS.On("Filter", "scope", []interface{}{"GLOBAL"}).Return(mockQS)
	mockQS.On("Filter", "fecha_inicio__gte", mock.Anything).Return(mockQS)
	mockQS.On("Filter", "fecha_fin__lte", mock.Anything).Return(mockQS)
	mockQS.On("Count").Return(int64(1), nil)
	mockQS.On("OrderBy", []string{"-pk_id_cupon"}).Return(mockQS)
	mockQS.On("Limit", 10).Return(mockQS)
	mockQS.On("Offset", int64(5)).Return(mockQS)
	mockQS.On("All", mock.AnythingOfType("*[]*models.Cupon"), []string(nil)).Return(int64(1), nil)

	controller.GetAll()

	assert.Equal(t, http.StatusOK, recorder.Code)
	mockOrmer.AssertExpectations(t)
	mockQS.AssertExpectations(t)
}

func TestGetAll_CountError(t *testing.T) {
	controller, recorder, _ := setupTest()

	mockOrmer.On("QueryTable", "cupon").Return(mockQS)
	mockQS.On("Count").Return(int64(0), fmt.Errorf("database error"))

	controller.GetAll()

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	mockOrmer.AssertExpectations(t)
	mockQS.AssertExpectations(t)
}

func TestGetAll_QueryError(t *testing.T) {
	controller, recorder, _ := setupTest()

	mockOrmer.On("QueryTable", "cupon").Return(mockQS)
	mockQS.On("Count").Return(int64(2), nil)
	mockQS.On("OrderBy", []string{"-pk_id_cupon"}).Return(mockQS)
	mockQS.On("Limit", 20).Return(mockQS)
	mockQS.On("Offset", int64(0)).Return(mockQS)
	mockQS.On("All", mock.AnythingOfType("*[]*models.Cupon"), []string(nil)).Return(int64(0), fmt.Errorf("query error"))

	controller.GetAll()

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	mockOrmer.AssertExpectations(t)
	mockQS.AssertExpectations(t)
}

func TestPost_Success(t *testing.T) {
	controller, recorder, ctx := setupTest()

	cupon := map[string]interface{}{
		"codigo":         "NEWTEST",
		"scope":          "GLOBAL",
		"tipoDescuento":  "PORCENTAJE",
		"valorDescuento": 15,
		"fechaInicio":    "2025-01-01",
		"fechaFin":       "2025-12-31",
		"activo":         true,
	}

	body, _ := json.Marshal(cupon)
	ctx.Request = httptest.NewRequest("POST", "/cupones", bytes.NewBuffer(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Input.RequestBody = body

	mockOrmer.On("Insert", mock.AnythingOfType("*models.Cupon")).Return(int64(1), nil)

	controller.Post()

	assert.Equal(t, http.StatusCreated, recorder.Code)
	mockOrmer.AssertExpectations(t)
}

func TestPost_InvalidJSON(t *testing.T) {
	controller, recorder, ctx := setupTest()

	ctx.Request = httptest.NewRequest("POST", "/cupones", bytes.NewBuffer([]byte("invalid json")))
	ctx.Request.Header.Set("Content-Type", "application/json")

	controller.Post()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestPost_ValidationError(t *testing.T) {
	controller, recorder, ctx := setupTest()

	fechaInicio, _ := time.Parse("2006-01-02", "2025-01-01")
	fechaFin, _ := time.Parse("2006-01-02", "2024-12-31")
	cupon := models.Cupon{
		Codigo:         "",
		Scope:          "INVALID",
		TipoDescuento:  "PORCENTAJE",
		ValorDescuento: 150,
		FechaInicio:    fechaInicio,
		FechaFin:       fechaFin,
		Activo:         true,
	}

	body, _ := json.Marshal(cupon)
	ctx.Request = httptest.NewRequest("POST", "/cupones", bytes.NewBuffer(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Input.RequestBody = body

	controller.Post()

	assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
}

func TestPost_DatabaseError(t *testing.T) {
	controller, recorder, ctx := setupTest()

	cupon := map[string]interface{}{
		"codigo":         "NEWTEST",
		"scope":          "GLOBAL",
		"tipoDescuento":  "PORCENTAJE",
		"valorDescuento": 15,
		"fechaInicio":    "2025-01-01",
		"fechaFin":       "2025-12-31",
		"activo":         true,
	}

	body, _ := json.Marshal(cupon)
	ctx.Request = httptest.NewRequest("POST", "/cupones", bytes.NewBuffer(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Input.RequestBody = body

	mockOrmer.On("Insert", mock.AnythingOfType("*models.Cupon")).Return(int64(0), fmt.Errorf("database error"))

	controller.Post()

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	mockOrmer.AssertExpectations(t)
}

func TestGetById_Success(t *testing.T) {
	controller, recorder, _ := setupTest()

	req := httptest.NewRequest("GET", "/cupones?id=1", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	cupon := models.Cupon{PkIdCupon: 1, Codigo: "TEST1", Scope: "GLOBAL"}
	mockOrmer.On("Read", mock.AnythingOfType("*models.Cupon"), []string(nil)).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.Cupon)
		*arg = cupon
	}).Return(nil)

	controller.GetById()

	assert.Equal(t, http.StatusOK, recorder.Code)
	mockOrmer.AssertExpectations(t)
}

func TestGetById_InvalidID(t *testing.T) {
	controller, recorder, ctx := setupTest()

	ctx.Input.SetParam(":id", "invalid")

	controller.GetById()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestGetById_NotFound(t *testing.T) {
	controller, recorder, _ := setupTest()

	req := httptest.NewRequest("GET", "/cupones?id=999", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	mockOrmer.On("Read", mock.AnythingOfType("*models.Cupon"), []string(nil)).Return(orm.ErrNoRows)

	mockOrmer.On("QueryTable", "cupon").Return(mockQS)
	mockQS.On("Filter", "codigo", []interface{}{"999"}).Return(mockQS)
	mockQS.On("One", mock.AnythingOfType("*models.Cupon")).Return(orm.ErrNoRows)

	controller.GetById()

	assert.Equal(t, http.StatusNotFound, recorder.Code)
	mockOrmer.AssertExpectations(t)
	mockQS.AssertExpectations(t)
}

func TestPut_Success(t *testing.T) {
	controller, recorder, _ := setupTest()

	cupon := map[string]interface{}{
		"codigo":         "UPDATED",
		"scope":          "GLOBAL",
		"tipoDescuento":  "PORCENTAJE",
		"valorDescuento": 20,
		"fechaInicio":    "2025-01-01",
		"fechaFin":       "2025-12-31",
		"activo":         true,
	}

	body, _ := json.Marshal(cupon)
	req := httptest.NewRequest("PUT", "/cupones?id=1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = body
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	existingCupon := models.Cupon{PkIdCupon: 1, Codigo: "OLD", Scope: "GLOBAL"}
	mockOrmer.On("Read", mock.AnythingOfType("*models.Cupon"), []string(nil)).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.Cupon)
		*arg = existingCupon
	}).Return(nil)
	mockOrmer.On("Update", mock.AnythingOfType("*models.Cupon"), []string(nil)).Return(int64(1), nil)

	controller.Put()

	assert.Equal(t, http.StatusOK, recorder.Code)
	mockOrmer.AssertExpectations(t)
}

func TestPut_InvalidID(t *testing.T) {
	controller, recorder, ctx := setupTest()

	ctx.Input.SetParam(":id", "invalid")

	controller.Put()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestPut_NotFound(t *testing.T) {
	controller, recorder, _ := setupTest()

	body, _ := json.Marshal(map[string]interface{}{"codigo": "TEST"})
	req := httptest.NewRequest("PUT", "/cupones?id=999", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = body
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	mockOrmer.On("Read", mock.AnythingOfType("*models.Cupon"), []string(nil)).Return(orm.ErrNoRows)

	controller.Put()

	assert.Equal(t, http.StatusNotFound, recorder.Code)
	mockOrmer.AssertExpectations(t)
}

func TestDelete_Success(t *testing.T) {
	controller, recorder, _ := setupTest()

	req := httptest.NewRequest("DELETE", "/cupones?id=1", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	cupon := models.Cupon{PkIdCupon: 1, Codigo: "TEST1", Activo: true}
	mockOrmer.On("Read", mock.AnythingOfType("*models.Cupon"), []string(nil)).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.Cupon)
		*arg = cupon
	}).Return(nil)
	mockOrmer.On("Update", mock.AnythingOfType("*models.Cupon"), []string{"Activo"}).Return(int64(1), nil)

	controller.Delete()

	assert.Equal(t, http.StatusOK, recorder.Code)
	mockOrmer.AssertExpectations(t)
}

func TestDelete_InvalidID(t *testing.T) {
	controller, recorder, ctx := setupTest()

	ctx.Input.SetParam(":id", "invalid")

	controller.Delete()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestValidarCupon_InvalidJSON(t *testing.T) {
	controller, recorder, ctx := setupTest()

	ctx.Request = httptest.NewRequest("POST", "/cupones/validar", bytes.NewBuffer([]byte("invalid json")))
	ctx.Request.Header.Set("Content-Type", "application/json")

	controller.ValidarCupon()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestRedimirCupon_InvalidJSON(t *testing.T) {
	controller, recorder, ctx := setupTest()

	ctx.Input.SetParam(":codigo", "TEST1")

	ctx.Request = httptest.NewRequest("POST", "/cupones/TEST1/redimir", bytes.NewBuffer([]byte("invalid json")))
	ctx.Request.Header.Set("Content-Type", "application/json")

	controller.RedimirCupon()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestListarRedenciones_Success(t *testing.T) {
	controller, recorder, _ := setupTest()

	cupon1 := &models.Cupon{PkIdCupon: 1}
	cupon2 := &models.Cupon{PkIdCupon: 2}
	cliente1 := &models.Cliente{PK_DOCUMENTO_CLIENTE: 123}
	cliente2 := &models.Cliente{PK_DOCUMENTO_CLIENTE: 456}
	redenciones := []models.CuponRedencion{
		{PkIdCuponRedencion: 1, PkIdCupon: cupon1, PkDocumentoCliente: cliente1},
		{PkIdCuponRedencion: 2, PkIdCupon: cupon2, PkDocumentoCliente: cliente2},
	}

	mockOrmer.On("QueryTable", "cupon_redencion").Return(mockQS)
	mockQS.On("Count").Return(int64(2), nil)
	mockQS.On("OrderBy", []string{"-created_at"}).Return(mockQS)
	mockQS.On("Limit", 20).Return(mockQS)
	mockQS.On("Offset", int64(0)).Return(mockQS)
	mockQS.On("All", mock.AnythingOfType("*[]*models.CuponRedencion"), []string(nil)).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*[]*models.CuponRedencion)
		*arg = []*models.CuponRedencion{&redenciones[0], &redenciones[1]}
	}).Return(int64(2), nil)

	controller.ListarRedenciones()

	assert.Equal(t, http.StatusOK, recorder.Code)
	mockOrmer.AssertExpectations(t)
	mockQS.AssertExpectations(t)
}
