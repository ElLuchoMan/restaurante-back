package reserva

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"restaurante/models"

	"github.com/beego/beego/v2/server/web/context"
	"github.com/stretchr/testify/assert"
)

// ============================================================================
// TESTS DE VALIDACIÓN - GetByDocumento
// ============================================================================

func TestReservaGetByDocumento_DocumentoFaltante(t *testing.T) {
	ctrl := &ReservaController{}
	ctrl.Data = make(map[interface{}]interface{})

	r := httptest.NewRequest(http.MethodGet, "/reservas/documento", nil)
	recorder := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(recorder, r)
	ctrl.Ctx = ctx

	ctrl.GetByDocumento()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Contains(t, response.Message, "documento")
}

func TestReservaGetByDocumento_DocumentoInvalido(t *testing.T) {
	ctrl := &ReservaController{}
	ctrl.Data = make(map[interface{}]interface{})

	r := httptest.NewRequest(http.MethodGet, "/reservas/documento?documento=abc", nil)
	recorder := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(recorder, r)
	ctrl.Ctx = ctx

	ctrl.GetByDocumento()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Contains(t, response.Message, "documento")
}

func TestReservaGetByDocumento_FechaInvalida(t *testing.T) {
	ctrl := &ReservaController{}
	ctrl.Data = make(map[interface{}]interface{})

	r := httptest.NewRequest(http.MethodGet, "/reservas/documento?documento=123456&fecha=invalid", nil)
	recorder := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(recorder, r)
	ctrl.Ctx = ctx

	ctrl.GetByDocumento()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Contains(t, response.Message, "fecha")
}

// ============================================================================
// TESTS DE VALIDACIÓN - GetByDocumentoCliente
// ============================================================================

func TestReservaGetByDocumentoCliente_DocumentoFaltante(t *testing.T) {
	ctrl := &ReservaController{}
	ctrl.Data = make(map[interface{}]interface{})

	r := httptest.NewRequest(http.MethodGet, "/reservas/cliente", nil)
	recorder := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(recorder, r)
	ctrl.Ctx = ctx

	ctrl.GetByDocumentoCliente()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Contains(t, response.Message, "documentoCliente")
}

func TestReservaGetByDocumentoCliente_DocumentoInvalido(t *testing.T) {
	ctrl := &ReservaController{}
	ctrl.Data = make(map[interface{}]interface{})

	r := httptest.NewRequest(http.MethodGet, "/reservas/cliente?documentoCliente=abc", nil)
	recorder := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(recorder, r)
	ctrl.Ctx = ctx

	ctrl.GetByDocumentoCliente()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Contains(t, response.Message, "documentoCliente")
}

func TestReservaGetByDocumentoCliente_FechaInvalida(t *testing.T) {
	ctrl := &ReservaController{}
	ctrl.Data = make(map[interface{}]interface{})

	r := httptest.NewRequest(http.MethodGet, "/reservas/cliente?documentoCliente=123456&fecha=invalid", nil)
	recorder := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(recorder, r)
	ctrl.Ctx = ctx

	ctrl.GetByDocumentoCliente()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Contains(t, response.Message, "fecha")
}
