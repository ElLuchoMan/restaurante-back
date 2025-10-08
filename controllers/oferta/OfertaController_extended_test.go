package oferta

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Tests adicionales para aumentar cobertura de OfertaController

func TestOfertaGetAll_QueryError(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	// Configurar mocks
	mockOfertOrmer.On("QueryTable", "oferta").Return(mockOfertQS)
	mockOfertQS.On("Count").Return(int64(10), nil)
	mockOfertQS.On("OrderBy", []string{"-pk_id_oferta"}).Return(mockOfertQS)
	mockOfertQS.On("Limit", 20).Return(mockOfertQS)
	mockOfertQS.On("Offset", int64(0)).Return(mockOfertQS)
	mockOfertQS.On("All", mock.AnythingOfType("*[]*models.Oferta"), []string(nil)).Return(int64(0), fmt.Errorf("query error"))

	// Ejecutar
	controller.GetAll()

	// Verificar
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	mockOfertOrmer.AssertExpectations(t)
	mockOfertQS.AssertExpectations(t)
}

func TestOfertaPost_DatabaseError(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	// Preparar request body válido
	oferta := map[string]interface{}{
		"nombre":            "Oferta Test",
		"descripcion":       "Descripción test",
		"tipoDescuento":     "PORCENTAJE",
		"valorDescuento":    10,
		"fechaInicio":       "2025-01-01",
		"fechaFin":          "2025-12-31",
		"tipoAplicacion":    "PEDIDO_COMPLETO",
		"montoMinimoPedido": 10000,
	}

	body, _ := json.Marshal(oferta)
	req := httptest.NewRequest("POST", "/ofertas", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = body
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Configurar mock para error de inserción
	mockOfertOrmer.On("Insert", mock.AnythingOfType("*models.Oferta")).Return(int64(0), fmt.Errorf("database error"))

	// Ejecutar
	controller.Post()

	// Verificar
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	mockOfertOrmer.AssertExpectations(t)
}

func TestOfertaPut_InvalidID(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	// Preparar request body válido
	oferta := map[string]interface{}{
		"nombre": "Oferta Updated",
	}

	body, _ := json.Marshal(oferta)
	req := httptest.NewRequest("PUT", "/ofertas?id=invalid", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = body
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Ejecutar
	controller.Put()

	// Verificar
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestOfertaPut_NotFound(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	// Preparar request body válido
	oferta := map[string]interface{}{
		"nombre":            "Oferta Updated",
		"descripcion":       "Desc",
		"tipoDescuento":     "PORCENTAJE",
		"valorDescuento":    10,
		"fechaInicio":       "2025-01-01",
		"fechaFin":          "2025-12-31",
		"tipoAplicacion":    "PEDIDO_COMPLETO",
		"montoMinimoPedido": 10000,
	}

	body, _ := json.Marshal(oferta)
	req := httptest.NewRequest("PUT", "/ofertas?id=999", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = body
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Configurar mock para no encontrado
	mockOfertOrmer.On("Read", mock.AnythingOfType("*models.Oferta"), []string(nil)).Return(orm.ErrNoRows)

	// Ejecutar
	controller.Put()

	// Verificar
	assert.Equal(t, http.StatusNotFound, recorder.Code)
	mockOfertOrmer.AssertExpectations(t)
}

func TestOfertaPut_InvalidJSON(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	// Request con JSON inválido
	req := httptest.NewRequest("PUT", "/ofertas?id=1", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = []byte("invalid json")
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Ejecutar
	controller.Put()

	// Verificar
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestOfertaDelete_InvalidID(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	// Request con ID inválido
	req := httptest.NewRequest("DELETE", "/ofertas?id=invalid", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Ejecutar
	controller.Delete()

	// Verificar
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestOfertaDelete_NotFound(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	// Request con ID válido
	req := httptest.NewRequest("DELETE", "/ofertas?id=999", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Configurar mock para no encontrado
	mockOfertOrmer.On("Read", mock.AnythingOfType("*models.Oferta"), []string(nil)).Return(orm.ErrNoRows)

	// Ejecutar
	controller.Delete()

	// Verificar
	assert.Equal(t, http.StatusNotFound, recorder.Code)
	mockOfertOrmer.AssertExpectations(t)
}

func TestOfertaDelete_AlreadyInactive(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	// Request con ID válido
	req := httptest.NewRequest("DELETE", "/ofertas?id=1", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Configurar mock para oferta ya inactiva
	oferta := models.Oferta{PkIdOferta: 1, Titulo: "Test", Activo: false}
	mockOfertOrmer.On("Read", mock.AnythingOfType("*models.Oferta"), []string(nil)).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.Oferta)
		*arg = oferta
	}).Return(nil)

	// Ejecutar
	controller.Delete()

	// Verificar
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	mockOfertOrmer.AssertExpectations(t)
}

func TestOfertaDelete_UpdateError(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	// Request con ID válido
	req := httptest.NewRequest("DELETE", "/ofertas?id=1", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Configurar mocks
	oferta := models.Oferta{PkIdOferta: 1, Titulo: "Test", Activo: true}
	mockOfertOrmer.On("Read", mock.AnythingOfType("*models.Oferta"), []string(nil)).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.Oferta)
		*arg = oferta
	}).Return(nil)
	mockOfertOrmer.On("Update", mock.AnythingOfType("*models.Oferta"), []string{"Activo"}).Return(int64(0), fmt.Errorf("update error"))

	// Ejecutar
	controller.Delete()

	// Verificar
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	mockOfertOrmer.AssertExpectations(t)
}

func TestOfertaAsociarProducto_MissingOfertaID(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	// Preparar request sin oferta_id
	payload := map[string]interface{}{
		"productoId": 1,
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/ofertas/asociar-producto", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = body
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Ejecutar
	controller.AsociarProducto()

	// Verificar
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestOfertaAsociarProducto_InvalidJSON(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	// Request con JSON inválido
	req := httptest.NewRequest("POST", "/ofertas/asociar-producto?oferta_id=1", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = []byte("invalid json")
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Ejecutar
	controller.AsociarProducto()

	// Verificar
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestOfertaDesasociarProducto_MissingOfertaID(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	// Request sin oferta_id
	req := httptest.NewRequest("DELETE", "/ofertas/desasociar-producto?producto_id=1", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Ejecutar
	controller.DesasociarProducto()

	// Verificar
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestOfertaDesasociarProducto_MissingProductoID(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	// Request sin producto_id
	req := httptest.NewRequest("DELETE", "/ofertas/desasociar-producto?oferta_id=1", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Ejecutar
	controller.DesasociarProducto()

	// Verificar
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestOfertaGetById_ReadError(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	// Request con ID válido
	req := httptest.NewRequest("GET", "/ofertas?id=1", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Configurar mock para error de lectura (no ErrNoRows)
	mockOfertOrmer.On("Read", mock.AnythingOfType("*models.Oferta"), []string(nil)).Return(fmt.Errorf("database error"))

	// Ejecutar
	controller.GetById()

	// Verificar
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	mockOfertOrmer.AssertExpectations(t)
}

func TestOfertaGetAll_LimitExceedsMax(t *testing.T) {
	controller, recorder, ctx := setupOfertaTest()

	// Configurar query parameters con límite > 100
	ctx.Input.SetParam("limit", "200")

	// Configurar mocks
	mockOfertOrmer.On("QueryTable", "oferta").Return(mockOfertQS)
	mockOfertQS.On("Count").Return(int64(10), nil)
	mockOfertQS.On("OrderBy", []string{"-pk_id_oferta"}).Return(mockOfertQS)
	mockOfertQS.On("Limit", 100).Return(mockOfertQS) // Debe limitar a 100
	mockOfertQS.On("Offset", int64(0)).Return(mockOfertQS)
	mockOfertQS.On("All", mock.AnythingOfType("*[]*models.Oferta"), []string(nil)).Return(int64(10), nil)

	// Ejecutar
	controller.GetAll()

	// Verificar
	assert.Equal(t, http.StatusOK, recorder.Code)
	mockOfertOrmer.AssertExpectations(t)
	mockOfertQS.AssertExpectations(t)
}
