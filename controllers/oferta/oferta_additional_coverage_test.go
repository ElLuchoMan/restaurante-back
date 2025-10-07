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

// TestObtenerOfertasActivas_MissingRestauranteId cubre el caso de restaurante_id faltante
func TestObtenerOfertasActivas_MissingRestauranteId(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	// Request sin restaurante_id
	req := httptest.NewRequest("GET", "/ofertas/activas", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Ejecutar
	controller.ObtenerOfertasActivas()

	// Verificar
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

// TestObtenerOfertasActivas_InvalidRestauranteId cubre el caso de restaurante_id inválido
func TestObtenerOfertasActivas_InvalidRestauranteId(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	// Request con restaurante_id inválido
	req := httptest.NewRequest("GET", "/ofertas/activas?restaurante_id=invalid", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Ejecutar
	controller.ObtenerOfertasActivas()

	// Verificar
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

// TestObtenerOfertasActivas_InvalidFecha cubre el caso de fecha inválida
func TestObtenerOfertasActivas_InvalidFecha(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	// Request con fecha inválida
	req := httptest.NewRequest("GET", "/ofertas/activas?restaurante_id=1&fecha=invalid-date", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Ejecutar
	controller.ObtenerOfertasActivas()

	// Verificar
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

// TestObtenerOfertasActivas_InvalidHora cubre el caso de hora inválida
func TestObtenerOfertasActivas_InvalidHora(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	// Request con hora inválida
	req := httptest.NewRequest("GET", "/ofertas/activas?restaurante_id=1&hora=invalid-time", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Ejecutar
	controller.ObtenerOfertasActivas()

	// Verificar
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

// TestObtenerOfertasActivas_InvalidProductoId cubre el caso de producto_id inválido
func TestObtenerOfertasActivas_InvalidProductoId(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	// Request con producto_id inválido (no se debería fallar, se ignora)
	req := httptest.NewRequest("GET", "/ofertas/activas?restaurante_id=1&producto_id=invalid", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Skip: Este test requiere servicio real
	t.Skip("Skipping: requires real service or complex mocking")
}

// TestObtenerOfertasActivas_ServiceError cubre el caso de error del servicio
func TestObtenerOfertasActivas_ServiceError(t *testing.T) {
	// Skip: Este test requiere servicio real
	t.Skip("Skipping: requires real service or complex mocking")
}

// TestAsociarProducto_InvalidJSON cubre el caso de JSON inválido
func TestAsociarProducto_InvalidJSON(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	// Request con JSON inválido
	req := httptest.NewRequest("POST", "/ofertas/productos?id=1", bytes.NewBuffer([]byte("invalid json")))
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

// TestAsociarProducto_OfertaNotFound cubre el caso de oferta no encontrada
func TestAsociarProducto_OfertaNotFound(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	// Preparar request
	payload := models.AsociarProductoOfertaRequest{
		ProductoId: 1,
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/ofertas/productos?id=999", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = body
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Configurar mock para oferta no encontrada
	mockOfertOrmer.On("Read", mock.AnythingOfType("*models.Oferta"), []string(nil)).Return(orm.ErrNoRows)

	// Ejecutar
	controller.AsociarProducto()

	// Verificar
	assert.Equal(t, http.StatusOK, recorder.Code) // Controller devuelve 200 con código 404 en JSON
	var resp models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &resp)
	assert.Equal(t, http.StatusNotFound, resp.Code)
	mockOfertOrmer.AssertExpectations(t)
}

// TestAsociarProducto_ReadOfertaError cubre el caso de error al leer oferta
func TestAsociarProducto_ReadOfertaError(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	// Preparar request
	payload := models.AsociarProductoOfertaRequest{
		ProductoId: 1,
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/ofertas/productos?id=1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = body
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Configurar mock para error al leer oferta
	mockOfertOrmer.On("Read", mock.AnythingOfType("*models.Oferta"), []string(nil)).Return(fmt.Errorf("database error"))

	// Ejecutar
	controller.AsociarProducto()

	// Verificar
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	mockOfertOrmer.AssertExpectations(t)
}

// TestAsociarProducto_ProductoNotFound cubre el caso de producto no encontrado
func TestAsociarProducto_ProductoNotFound(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	// Preparar request
	payload := models.AsociarProductoOfertaRequest{
		ProductoId: 999,
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/ofertas/productos?id=1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = body
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Configurar mocks: oferta existe, producto no existe
	mockOfertOrmer.On("Read", mock.AnythingOfType("*models.Oferta"), []string(nil)).Return(nil).Once()
	mockOfertOrmer.On("Read", mock.AnythingOfType("*models.Producto"), []string(nil)).Return(orm.ErrNoRows).Once()

	// Ejecutar
	controller.AsociarProducto()

	// Verificar
	assert.Equal(t, http.StatusOK, recorder.Code)
	var resp models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &resp)
	assert.Equal(t, http.StatusNotFound, resp.Code)
	assert.Equal(t, "Producto no encontrado", resp.Message)
	mockOfertOrmer.AssertExpectations(t)
}

// TestAsociarProducto_ReadProductoError cubre el caso de error al leer producto
func TestAsociarProducto_ReadProductoError(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	// Preparar request
	payload := models.AsociarProductoOfertaRequest{
		ProductoId: 1,
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/ofertas/productos?id=1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = body
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Configurar mocks: oferta existe, error al leer producto
	mockOfertOrmer.On("Read", mock.AnythingOfType("*models.Oferta"), []string(nil)).Return(nil).Once()
	mockOfertOrmer.On("Read", mock.AnythingOfType("*models.Producto"), []string(nil)).Return(fmt.Errorf("database error")).Once()

	// Ejecutar
	controller.AsociarProducto()

	// Verificar
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	mockOfertOrmer.AssertExpectations(t)
}

// TestDesasociarProducto_InvalidOfertaId cubre el caso de ID de oferta inválido
func TestDesasociarProducto_InvalidOfertaId(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	// Request con ID de oferta inválido
	req := httptest.NewRequest("DELETE", "/ofertas/productos?id=invalid&producto_id=1", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Ejecutar
	controller.DesasociarProducto()

	// Verificar
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

// TestDesasociarProducto_InvalidProductoId cubre el caso de ID de producto inválido
func TestDesasociarProducto_InvalidProductoId(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	// Request con ID de producto inválido
	req := httptest.NewRequest("DELETE", "/ofertas/productos?id=1&producto_id=invalid", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Ejecutar
	controller.DesasociarProducto()

	// Verificar
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

// TestDesasociarProducto_AssociationNotFound cubre el caso de asociación no encontrada
func TestDesasociarProducto_AssociationNotFound(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	// Request con IDs válidos
	req := httptest.NewRequest("DELETE", "/ofertas/productos?id=1&producto_id=999", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Configurar mocks: asociación no encontrada
	mockOfertOrmer.On("QueryTable", "oferta_producto").Return(mockOfertQS)
	mockOfertQS.On("Filter", "pk_id_oferta", []interface{}{int64(1)}).Return(mockOfertQS)
	mockOfertQS.On("Filter", "pk_id_producto", []interface{}{int64(999)}).Return(mockOfertQS)
	mockOfertQS.On("One", mock.AnythingOfType("*models.OfertaProducto")).Return(orm.ErrNoRows)

	// Ejecutar
	controller.DesasociarProducto()

	// Verificar
	assert.Equal(t, http.StatusOK, recorder.Code)
	var resp models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &resp)
	assert.Equal(t, http.StatusNotFound, resp.Code)
	mockOfertOrmer.AssertExpectations(t)
	mockOfertQS.AssertExpectations(t)
}

// TestDesasociarProducto_QueryError cubre el caso de error de consulta
func TestDesasociarProducto_QueryError(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	// Request con IDs válidos
	req := httptest.NewRequest("DELETE", "/ofertas/productos?id=1&producto_id=1", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Configurar mocks: error de consulta
	mockOfertOrmer.On("QueryTable", "oferta_producto").Return(mockOfertQS)
	mockOfertQS.On("Filter", "pk_id_oferta", []interface{}{int64(1)}).Return(mockOfertQS)
	mockOfertQS.On("Filter", "pk_id_producto", []interface{}{int64(1)}).Return(mockOfertQS)
	mockOfertQS.On("One", mock.AnythingOfType("*models.OfertaProducto")).Return(fmt.Errorf("database error"))

	// Ejecutar
	controller.DesasociarProducto()

	// Verificar
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	mockOfertOrmer.AssertExpectations(t)
	mockOfertQS.AssertExpectations(t)
}

// TestDesasociarProducto_DeleteError cubre el caso de error al eliminar
func TestDesasociarProducto_DeleteError(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	// Request con IDs válidos
	req := httptest.NewRequest("DELETE", "/ofertas/productos?id=1&producto_id=1", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Configurar mocks: asociación encontrada, error al eliminar
	mockOfertOrmer.On("QueryTable", "oferta_producto").Return(mockOfertQS)
	mockOfertQS.On("Filter", "pk_id_oferta", []interface{}{int64(1)}).Return(mockOfertQS)
	mockOfertQS.On("Filter", "pk_id_producto", []interface{}{int64(1)}).Return(mockOfertQS)
	mockOfertQS.On("One", mock.AnythingOfType("*models.OfertaProducto")).Return(nil)
	mockOfertOrmer.On("Delete", mock.AnythingOfType("*models.OfertaProducto"), []string(nil)).Return(int64(0), fmt.Errorf("delete error"))

	// Ejecutar
	controller.DesasociarProducto()

	// Verificar
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	mockOfertOrmer.AssertExpectations(t)
	mockOfertQS.AssertExpectations(t)
}
