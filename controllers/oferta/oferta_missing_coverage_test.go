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
	"github.com/stretchr/testify/mock"
)

// ============================================================================
// TESTS PARA MEJORAR COBERTURA DE POST
// ============================================================================

func TestOfertaPost_HoraInicioInvalida(t *testing.T) {
	ctrl := &OfertaController{}
	ctrl.Data = make(map[interface{}]interface{})

	horaInvalida := "25:99"
	req := models.CrearOfertaRequest{
		Titulo:          "Oferta Test",
		TipoDescuento:   models.TipoDescuentoPorcentaje,
		ValorDescuento:  10,
		FechaInicio:     "2025-01-01",
		FechaFin:        "2025-12-31",
		PkIdRestaurante: 1,
		HoraInicio:      &horaInvalida,
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
	assert.Contains(t, response.Message, "Hora de inicio inválida")
}

func TestOfertaPost_HoraFinInvalida(t *testing.T) {
	ctrl := &OfertaController{}
	ctrl.Data = make(map[interface{}]interface{})

	horaInicio := "10:00"
	horaFinInvalida := "invalid"
	req := models.CrearOfertaRequest{
		Titulo:          "Oferta Test",
		TipoDescuento:   models.TipoDescuentoPorcentaje,
		ValorDescuento:  10,
		FechaInicio:     "2025-01-01",
		FechaFin:        "2025-12-31",
		PkIdRestaurante: 1,
		HoraInicio:      &horaInicio,
		HoraFin:         &horaFinInvalida,
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
	assert.Contains(t, response.Message, "Hora de fin inválida")
}

func TestOfertaPost_FechaFinInvalida(t *testing.T) {
	ctrl := &OfertaController{}
	ctrl.Data = make(map[interface{}]interface{})

	req := models.CrearOfertaRequest{
		Titulo:          "Oferta Test",
		TipoDescuento:   models.TipoDescuentoPorcentaje,
		ValorDescuento:  10,
		FechaInicio:     "2025-01-01",
		FechaFin:        "invalid-date",
		PkIdRestaurante: 1,
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
	assert.Contains(t, response.Message, "Fecha de fin inválida")
}

// ============================================================================
// TESTS PARA MEJORAR COBERTURA DE PUT
// ============================================================================

func TestOfertaPut_TipoDescuentoInvalido(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	oferta := &models.Oferta{
		PkIdOferta: 1,
		Titulo:     "Oferta Existente",
	}

	// Configurar mock para que la oferta exista
	mockOfertOrmer.On("Read", mock.AnythingOfType("*models.Oferta"), []string(nil)).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.Oferta)
		*arg = *oferta
	}).Return(nil)

	req := models.CrearOfertaRequest{
		Titulo:          "Oferta Updated",
		TipoDescuento:   "INVALID_TYPE",
		ValorDescuento:  15,
		FechaInicio:     "2025-02-01",
		FechaFin:        "2025-11-30",
		PkIdRestaurante: 1,
	}
	body, _ := json.Marshal(req)

	r := httptest.NewRequest(http.MethodPut, "/ofertas?id=1", bytes.NewReader(body))
	ctx := context.NewContext()
	ctx.Reset(recorder, r)
	ctx.Input.RequestBody = body
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	controller.Put()

	assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)

	var response models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Equal(t, http.StatusUnprocessableEntity, response.Code)
	assert.Contains(t, response.Message, "Tipo de descuento no válido")
}

func TestOfertaPut_FechaInicioInvalida(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	oferta := &models.Oferta{
		PkIdOferta: 1,
		Titulo:     "Oferta Existente",
	}

	// Configurar mock para que la oferta exista
	mockOfertOrmer.On("Read", mock.AnythingOfType("*models.Oferta"), []string(nil)).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.Oferta)
		*arg = *oferta
	}).Return(nil)

	req := models.CrearOfertaRequest{
		Titulo:          "Oferta Updated",
		TipoDescuento:   models.TipoDescuentoPorcentaje,
		ValorDescuento:  15,
		FechaInicio:     "invalid-date",
		FechaFin:        "2025-11-30",
		PkIdRestaurante: 1,
	}
	body, _ := json.Marshal(req)

	r := httptest.NewRequest(http.MethodPut, "/ofertas?id=1", bytes.NewReader(body))
	ctx := context.NewContext()
	ctx.Reset(recorder, r)
	ctx.Input.RequestBody = body
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	controller.Put()

	assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)

	var response models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Equal(t, http.StatusUnprocessableEntity, response.Code)
	assert.Contains(t, response.Message, "Fecha de inicio inválida")
}

func TestOfertaPut_FechaFinInvalida(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	oferta := &models.Oferta{
		PkIdOferta: 1,
		Titulo:     "Oferta Existente",
	}

	// Configurar mock para que la oferta exista
	mockOfertOrmer.On("Read", mock.AnythingOfType("*models.Oferta"), []string(nil)).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.Oferta)
		*arg = *oferta
	}).Return(nil)

	req := models.CrearOfertaRequest{
		Titulo:          "Oferta Updated",
		TipoDescuento:   models.TipoDescuentoPorcentaje,
		ValorDescuento:  15,
		FechaInicio:     "2025-02-01",
		FechaFin:        "not-a-date",
		PkIdRestaurante: 1,
	}
	body, _ := json.Marshal(req)

	r := httptest.NewRequest(http.MethodPut, "/ofertas?id=1", bytes.NewReader(body))
	ctx := context.NewContext()
	ctx.Reset(recorder, r)
	ctx.Input.RequestBody = body
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	controller.Put()

	assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)

	var response models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Equal(t, http.StatusUnprocessableEntity, response.Code)
	assert.Contains(t, response.Message, "Fecha de fin inválida")
}

func TestOfertaPut_HoraInicioInvalida(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	oferta := &models.Oferta{
		PkIdOferta: 1,
		Titulo:     "Oferta Existente",
	}

	// Configurar mock para que la oferta exista
	mockOfertOrmer.On("Read", mock.AnythingOfType("*models.Oferta"), []string(nil)).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.Oferta)
		*arg = *oferta
	}).Return(nil)

	horaInvalida := "99:99"
	req := models.CrearOfertaRequest{
		Titulo:          "Oferta Updated",
		TipoDescuento:   models.TipoDescuentoPorcentaje,
		ValorDescuento:  15,
		FechaInicio:     "2025-02-01",
		FechaFin:        "2025-11-30",
		PkIdRestaurante: 1,
		HoraInicio:      &horaInvalida,
	}
	body, _ := json.Marshal(req)

	r := httptest.NewRequest(http.MethodPut, "/ofertas?id=1", bytes.NewReader(body))
	ctx := context.NewContext()
	ctx.Reset(recorder, r)
	ctx.Input.RequestBody = body
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	controller.Put()

	assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)

	var response models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Equal(t, http.StatusUnprocessableEntity, response.Code)
	assert.Contains(t, response.Message, "Hora de inicio inválida")
}

func TestOfertaPut_HoraFinInvalida(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	oferta := &models.Oferta{
		PkIdOferta: 1,
		Titulo:     "Oferta Existente",
	}

	// Configurar mock para que la oferta exista
	mockOfertOrmer.On("Read", mock.AnythingOfType("*models.Oferta"), []string(nil)).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.Oferta)
		*arg = *oferta
	}).Return(nil)

	horaInicio := "10:00"
	horaFinInvalida := "30:00"
	req := models.CrearOfertaRequest{
		Titulo:          "Oferta Updated",
		TipoDescuento:   models.TipoDescuentoPorcentaje,
		ValorDescuento:  15,
		FechaInicio:     "2025-02-01",
		FechaFin:        "2025-11-30",
		PkIdRestaurante: 1,
		HoraInicio:      &horaInicio,
		HoraFin:         &horaFinInvalida,
	}
	body, _ := json.Marshal(req)

	r := httptest.NewRequest(http.MethodPut, "/ofertas?id=1", bytes.NewReader(body))
	ctx := context.NewContext()
	ctx.Reset(recorder, r)
	ctx.Input.RequestBody = body
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	controller.Put()

	assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)

	var response models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Equal(t, http.StatusUnprocessableEntity, response.Code)
	assert.Contains(t, response.Message, "Hora de fin inválida")
}

func TestOfertaPut_ReadErrorInternal(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	// Configurar mock para simular error de base de datos (no ErrNoRows)
	mockOfertOrmer.On("Read", mock.AnythingOfType("*models.Oferta"), []string(nil)).Return(assert.AnError)

	req := models.CrearOfertaRequest{
		Titulo:          "Oferta Updated",
		TipoDescuento:   models.TipoDescuentoPorcentaje,
		ValorDescuento:  15,
		FechaInicio:     "2025-02-01",
		FechaFin:        "2025-11-30",
		PkIdRestaurante: 1,
	}
	body, _ := json.Marshal(req)

	r := httptest.NewRequest(http.MethodPut, "/ofertas?id=1", bytes.NewReader(body))
	ctx := context.NewContext()
	ctx.Reset(recorder, r)
	ctx.Input.RequestBody = body
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	controller.Put()

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)

	var response models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.Contains(t, response.Message, "Error interno")
}
