package oferta

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"restaurante/models"

	"encoding/json"

	"github.com/beego/beego/v2/client/orm"
	beecontext "github.com/beego/beego/v2/server/web/context"
	"github.com/stretchr/testify/assert"
)

func TestObtenerOfertasActivas_RestauranteIdInvalido(t *testing.T) {
	ctrl := &OfertaController{}
	ctrl.Data = make(map[interface{}]interface{})
	req := httptest.NewRequest(http.MethodGet, "/ofertas/activas?restaurante_id=invalid", nil)
	recorder := httptest.NewRecorder()

	ctx := beecontext.NewContext()
	ctx.Reset(recorder, req)
	ctx.Request.URL.RawQuery = "restaurante_id=invalid"
	ctrl.Ctx = ctx

	ctrl.ObtenerOfertasActivas()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response models.ApiResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Contains(t, response.Message, "restaurante_id")
}

func TestObtenerOfertasActivas_RestauranteIdAusente(t *testing.T) {
	ctrl := &OfertaController{}
	ctrl.Data = make(map[interface{}]interface{})
	req := httptest.NewRequest(http.MethodGet, "/ofertas/activas", nil)
	recorder := httptest.NewRecorder()

	ctx := beecontext.NewContext()
	ctx.Reset(recorder, req)
	ctrl.Ctx = ctx

	ctrl.ObtenerOfertasActivas()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response models.ApiResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Contains(t, response.Message, "restaurante_id")
}

func TestObtenerOfertasActivas_FechaInvalida(t *testing.T) {
	ctrl := &OfertaController{}
	ctrl.Data = make(map[interface{}]interface{})
	req := httptest.NewRequest(http.MethodGet, "/ofertas/activas?restaurante_id=1&fecha=invalid-date", nil)
	recorder := httptest.NewRecorder()

	ctx := beecontext.NewContext()
	ctx.Reset(recorder, req)
	ctx.Request.URL.RawQuery = "restaurante_id=1&fecha=invalid-date"
	ctrl.Ctx = ctx

	ctrl.ObtenerOfertasActivas()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response models.ApiResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Contains(t, response.Message, "Fecha inválida")
	assert.Contains(t, response.Message, "YYYY-MM-DD")
}

func TestObtenerOfertasActivas_HoraInvalida(t *testing.T) {
	ctrl := &OfertaController{}
	ctrl.Data = make(map[interface{}]interface{})
	req := httptest.NewRequest(http.MethodGet, "/ofertas/activas?restaurante_id=1&hora=25:99", nil)
	recorder := httptest.NewRecorder()

	ctx := beecontext.NewContext()
	ctx.Reset(recorder, req)
	ctx.Request.URL.RawQuery = "restaurante_id=1&hora=25:99"
	ctrl.Ctx = ctx

	ctrl.ObtenerOfertasActivas()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response models.ApiResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Contains(t, response.Message, "Hora inválida")
	assert.Contains(t, response.Message, "HH:MM")
}

func TestOfertaServiceOrmFactory_Coverage(t *testing.T) {
	originalOrmProvider := ormProvider
	defer func() { ormProvider = originalOrmProvider }()

	called := false
	ormProvider = func() orm.Ormer {
		called = true
		return nil
	}

	result := ofertaServiceOrmFactory()
	assert.Nil(t, result)
	assert.True(t, called, "ormProvider debe ser llamado")
}

func TestOfertaServiceOrmBase_Coverage(t *testing.T) {
	originalOrmProvider := ormProvider
	defer func() { ormProvider = originalOrmProvider }()

	called := false
	ormProvider = func() orm.Ormer {
		called = true
		return nil
	}

	result := ofertaServiceOrmBase()
	assert.Nil(t, result)
	assert.True(t, called, "ormProvider debe ser llamado")
}
