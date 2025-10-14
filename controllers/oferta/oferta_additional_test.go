package oferta

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"restaurante/models"

	"github.com/beego/beego/v2/server/web/context"
	"github.com/stretchr/testify/assert"
)

func TestOfertaPost_JSONInvalido(t *testing.T) {
	ctrl := &OfertaController{}
	ctrl.Data = make(map[interface{}]interface{})

	body := []byte(`{"invalid json}`)

	r := httptest.NewRequest(http.MethodPost, "/ofertas", bytes.NewReader(body))
	recorder := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(recorder, r)
	ctx.Input.RequestBody = body
	ctrl.Ctx = ctx

	ctrl.Post()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Contains(t, response.Message, "JSON")
}

func TestOfertaPost_TituloVacio(t *testing.T) {
	ctrl := &OfertaController{}
	ctrl.Data = make(map[interface{}]interface{})

	req := models.CrearOfertaRequest{
		Titulo: "",
	}
	body, _ := json.Marshal(req)

	r := httptest.NewRequest(http.MethodPost, "/ofertas", bytes.NewReader(body))
	recorder := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(recorder, r)
	ctx.Input.RequestBody = body
	ctrl.Ctx = ctx

	ctrl.Post()

	assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)

	var response models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Equal(t, http.StatusUnprocessableEntity, response.Code)
}

func TestOfertaPost_FechaInicioInvalida(t *testing.T) {
	ctrl := &OfertaController{}
	ctrl.Data = make(map[interface{}]interface{})

	body := []byte(`{
		"titulo": "Test Oferta",
		"tipo_descuento": "PORCENTAJE",
		"valor_descuento": 10,
		"fecha_inicio": "invalid-date",
		"fecha_fin": "2025-12-31",
		"pk_id_restaurante": 1
	}`)

	r := httptest.NewRequest(http.MethodPost, "/ofertas", bytes.NewReader(body))
	recorder := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(recorder, r)
	ctx.Input.RequestBody = body
	ctrl.Ctx = ctx

	ctrl.Post()

	assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)

	var response models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Equal(t, http.StatusUnprocessableEntity, response.Code)
}

func TestOfertaPut_IDFaltante(t *testing.T) {
	ctrl := &OfertaController{}
	ctrl.Data = make(map[interface{}]interface{})

	req := models.CrearOfertaRequest{
		Titulo: "Test",
	}
	body, _ := json.Marshal(req)

	r := httptest.NewRequest(http.MethodPut, "/ofertas", bytes.NewReader(body))
	recorder := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(recorder, r)
	ctx.Input.RequestBody = body
	ctrl.Ctx = ctx

	ctrl.Put()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Contains(t, response.Message, "ID")
}

func TestOfertaPut_IDInvalido(t *testing.T) {
	ctrl := &OfertaController{}
	ctrl.Data = make(map[interface{}]interface{})

	req := models.CrearOfertaRequest{
		Titulo: "Test",
	}
	body, _ := json.Marshal(req)

	r := httptest.NewRequest(http.MethodPut, "/ofertas?id=abc", bytes.NewReader(body))
	recorder := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(recorder, r)
	ctx.Input.RequestBody = body
	ctrl.Ctx = ctx

	ctrl.Put()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Contains(t, response.Message, "ID")
}

func TestOfertaPut_JSONInvalido(t *testing.T) {
	ctrl := &OfertaController{}
	ctrl.Data = make(map[interface{}]interface{})

	body := []byte(`{"invalid json}`)

	r := httptest.NewRequest(http.MethodPut, "/ofertas?id=1", bytes.NewReader(body))
	recorder := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(recorder, r)
	ctx.Input.RequestBody = body
	ctrl.Ctx = ctx

	ctrl.Put()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Contains(t, response.Message, "JSON")
}

func TestOfertaDelete_IDFaltante(t *testing.T) {
	ctrl := &OfertaController{}
	ctrl.Data = make(map[interface{}]interface{})

	r := httptest.NewRequest(http.MethodDelete, "/ofertas", nil)
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

func TestOfertaDelete_IDInvalido(t *testing.T) {
	ctrl := &OfertaController{}
	ctrl.Data = make(map[interface{}]interface{})

	r := httptest.NewRequest(http.MethodDelete, "/ofertas?id=abc", nil)
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

func TestOfertaDelete_IDCero(t *testing.T) {
	ctrl := &OfertaController{}
	ctrl.Data = make(map[interface{}]interface{})

	r := httptest.NewRequest(http.MethodDelete, "/ofertas?id=0", nil)
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

func TestOfertaGetById_IDFaltante(t *testing.T) {
	ctrl := &OfertaController{}
	ctrl.Data = make(map[interface{}]interface{})

	r := httptest.NewRequest(http.MethodGet, "/ofertas/search", nil)
	recorder := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(recorder, r)
	ctrl.Ctx = ctx

	ctrl.GetById()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Contains(t, response.Message, "ID")
}

func TestOfertaGetById_IDInvalido(t *testing.T) {
	ctrl := &OfertaController{}
	ctrl.Data = make(map[interface{}]interface{})

	r := httptest.NewRequest(http.MethodGet, "/ofertas/search?id=abc", nil)
	recorder := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(recorder, r)
	ctrl.Ctx = ctx

	ctrl.GetById()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Contains(t, response.Message, "ID")
}

func TestOfertaGetById_IDCero(t *testing.T) {
	ctrl := &OfertaController{}
	ctrl.Data = make(map[interface{}]interface{})

	r := httptest.NewRequest(http.MethodGet, "/ofertas/search?id=0", nil)
	recorder := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(recorder, r)
	ctrl.Ctx = ctx

	ctrl.GetById()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Contains(t, response.Message, "ID")
}
