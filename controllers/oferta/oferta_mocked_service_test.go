package oferta

import (
	stdcontext "context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"restaurante/models"
	"restaurante/services"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockOfertaServiceInterface struct {
	mock.Mock
}

func (m *MockOfertaServiceInterface) ObtenerOfertasActivas(ctx stdcontext.Context, restauranteId int64, fecha *time.Time, hora *time.Time, productoId *int64) ([]*models.OfertaActivaResponse, error) {
	args := m.Called(ctx, restauranteId, fecha, hora, productoId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.OfertaActivaResponse), args.Error(1)
}

func (m *MockOfertaServiceInterface) ValidarReglasNegocioOferta(oferta *models.Oferta) error {
	args := m.Called(oferta)
	return args.Error(0)
}

func (m *MockOfertaServiceInterface) CalcularDescuentoOferta(oferta *models.Oferta, items []models.ValidarCuponItemRequest) (int64, error) {
	args := m.Called(oferta, items)
	return args.Get(0).(int64), args.Error(1)
}

func TestOfertaObtenerOfertasActivas_ConServicioMockeado_Exitoso(t *testing.T) {

	originalNewOfertaService := newOfertaService
	originalOfertaServiceOrmFactory := ofertaServiceOrmFactory
	defer func() {
		newOfertaService = originalNewOfertaService
		ofertaServiceOrmFactory = originalOfertaServiceOrmFactory
	}()

	mockService := new(MockOfertaServiceInterface)
	expectedOfertas := []*models.OfertaActivaResponse{
		{
			OfertaId: 1,
			Titulo:   "Oferta de Prueba",
		},
	}
	mockService.On("ObtenerOfertasActivas", mock.Anything, int64(1), mock.Anything, mock.Anything, mock.Anything).
		Return(expectedOfertas, nil)

	newOfertaService = func(o orm.Ormer) services.OfertaServiceInterface {
		return mockService
	}
	ofertaServiceOrmFactory = func() orm.Ormer {
		return nil
	}

	ctrl := &OfertaController{}
	ctrl.Data = make(map[interface{}]interface{})

	r := httptest.NewRequest(http.MethodGet, "/ofertas/activas?restaurante_id=1", nil)
	recorder := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(recorder, r)
	ctrl.Ctx = ctx

	ctrl.ObtenerOfertasActivas()

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, "Ofertas activas obtenidas exitosamente", response.Message)

	mockService.AssertExpectations(t)
}

func TestOfertaObtenerOfertasActivas_ConServicioMockeado_RestauranteIdFaltante(t *testing.T) {
	ctrl := &OfertaController{}
	ctrl.Data = make(map[interface{}]interface{})

	r := httptest.NewRequest(http.MethodGet, "/ofertas/activas", nil)
	recorder := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(recorder, r)
	ctrl.Ctx = ctx

	ctrl.ObtenerOfertasActivas()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "restaurante_id es requerido y debe ser un número entero válido", response.Message)
}

func TestOfertaObtenerOfertasActivas_ConServicioMockeado_RestauranteIdInvalido(t *testing.T) {
	ctrl := &OfertaController{}
	ctrl.Data = make(map[interface{}]interface{})

	r := httptest.NewRequest(http.MethodGet, "/ofertas/activas?restaurante_id=abc", nil)
	recorder := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(recorder, r)
	ctrl.Ctx = ctx

	ctrl.ObtenerOfertasActivas()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "restaurante_id es requerido y debe ser un número entero válido", response.Message)
}

func TestOfertaObtenerOfertasActivas_ConServicioMockeado_FechaInvalida(t *testing.T) {
	ctrl := &OfertaController{}
	ctrl.Data = make(map[interface{}]interface{})

	r := httptest.NewRequest(http.MethodGet, "/ofertas/activas?restaurante_id=1&fecha=invalid", nil)
	recorder := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(recorder, r)
	ctrl.Ctx = ctx

	ctrl.ObtenerOfertasActivas()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "Fecha inválida - debe tener formato YYYY-MM-DD", response.Message)
}

func TestOfertaObtenerOfertasActivas_ConServicioMockeado_HoraInvalida(t *testing.T) {
	ctrl := &OfertaController{}
	ctrl.Data = make(map[interface{}]interface{})

	r := httptest.NewRequest(http.MethodGet, "/ofertas/activas?restaurante_id=1&hora=invalid", nil)
	recorder := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(recorder, r)
	ctrl.Ctx = ctx

	ctrl.ObtenerOfertasActivas()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "Hora inválida - debe tener formato HH:MM", response.Message)
}

func TestOfertaObtenerOfertasActivas_ConServicioMockeado_ProductoIdInvalido(t *testing.T) {

	originalNewOfertaService := newOfertaService
	originalOfertaServiceOrmFactory := ofertaServiceOrmFactory
	defer func() {
		newOfertaService = originalNewOfertaService
		ofertaServiceOrmFactory = originalOfertaServiceOrmFactory
	}()

	mockService := new(MockOfertaServiceInterface)
	mockService.On("ObtenerOfertasActivas", mock.Anything, int64(1), mock.Anything, mock.Anything, (*int64)(nil)).
		Return([]*models.OfertaActivaResponse{}, nil)

	newOfertaService = func(o orm.Ormer) services.OfertaServiceInterface {
		return mockService
	}
	ofertaServiceOrmFactory = func() orm.Ormer {
		return nil
	}

	ctrl := &OfertaController{}
	ctrl.Data = make(map[interface{}]interface{})

	r := httptest.NewRequest(http.MethodGet, "/ofertas/activas?restaurante_id=1&producto_id=abc", nil)
	recorder := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(recorder, r)
	ctrl.Ctx = ctx

	ctrl.ObtenerOfertasActivas()

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Equal(t, http.StatusOK, response.Code)

	mockService.AssertExpectations(t)
}

func TestOfertaObtenerOfertasActivas_ConServicioMockeado_ErrorDelServicio(t *testing.T) {

	originalNewOfertaService := newOfertaService
	originalOfertaServiceOrmFactory := ofertaServiceOrmFactory
	defer func() {
		newOfertaService = originalNewOfertaService
		ofertaServiceOrmFactory = originalOfertaServiceOrmFactory
	}()

	mockService := new(MockOfertaServiceInterface)
	mockService.On("ObtenerOfertasActivas", mock.Anything, int64(1), mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errors.New("error de base de datos"))

	newOfertaService = func(o orm.Ormer) services.OfertaServiceInterface {
		return mockService
	}
	ofertaServiceOrmFactory = func() orm.Ormer {
		return nil
	}

	ctrl := &OfertaController{}
	ctrl.Data = make(map[interface{}]interface{})

	r := httptest.NewRequest(http.MethodGet, "/ofertas/activas?restaurante_id=1", nil)
	recorder := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(recorder, r)
	ctrl.Ctx = ctx

	ctrl.ObtenerOfertasActivas()

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)

	var response models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.Equal(t, "Error al obtener ofertas activas", response.Message)

	mockService.AssertExpectations(t)
}

func TestOfertaObtenerOfertasActivas_ConServicioMockeado_SinResultados(t *testing.T) {

	originalNewOfertaService := newOfertaService
	originalOfertaServiceOrmFactory := ofertaServiceOrmFactory
	defer func() {
		newOfertaService = originalNewOfertaService
		ofertaServiceOrmFactory = originalOfertaServiceOrmFactory
	}()

	mockService := new(MockOfertaServiceInterface)
	mockService.On("ObtenerOfertasActivas", mock.Anything, int64(1), mock.Anything, mock.Anything, mock.Anything).
		Return([]*models.OfertaActivaResponse{}, nil)

	newOfertaService = func(o orm.Ormer) services.OfertaServiceInterface {
		return mockService
	}
	ofertaServiceOrmFactory = func() orm.Ormer {
		return nil
	}

	ctrl := &OfertaController{}
	ctrl.Data = make(map[interface{}]interface{})

	r := httptest.NewRequest(http.MethodGet, "/ofertas/activas?restaurante_id=1", nil)
	recorder := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(recorder, r)
	ctrl.Ctx = ctx

	ctrl.ObtenerOfertasActivas()

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, "Ofertas activas obtenidas exitosamente", response.Message)

	mockService.AssertExpectations(t)
}
