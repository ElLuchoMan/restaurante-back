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

func TestObtenerOfertasActivas_MissingRestauranteId(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	req := httptest.NewRequest("GET", "/ofertas/activas", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	controller.ObtenerOfertasActivas()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestObtenerOfertasActivas_InvalidRestauranteId(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	req := httptest.NewRequest("GET", "/ofertas/activas?restaurante_id=invalid", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	controller.ObtenerOfertasActivas()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestObtenerOfertasActivas_InvalidFecha(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	req := httptest.NewRequest("GET", "/ofertas/activas?restaurante_id=1&fecha=invalid-date", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	controller.ObtenerOfertasActivas()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestObtenerOfertasActivas_InvalidHora(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	req := httptest.NewRequest("GET", "/ofertas/activas?restaurante_id=1&hora=invalid-time", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	controller.ObtenerOfertasActivas()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestAsociarProducto_InvalidJSON(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	req := httptest.NewRequest("POST", "/ofertas/productos?id=1", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = []byte("invalid json")
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	controller.AsociarProducto()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestAsociarProducto_OfertaNotFound(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

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

	mockOfertOrmer.On("Read", mock.AnythingOfType("*models.Oferta"), []string(nil)).Return(orm.ErrNoRows)

	controller.AsociarProducto()

	assert.Equal(t, http.StatusOK, recorder.Code)
	var resp models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &resp)
	assert.Equal(t, http.StatusNotFound, resp.Code)
	mockOfertOrmer.AssertExpectations(t)
}

func TestAsociarProducto_ReadOfertaError(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

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

	mockOfertOrmer.On("Read", mock.AnythingOfType("*models.Oferta"), []string(nil)).Return(fmt.Errorf("database error"))

	controller.AsociarProducto()

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	mockOfertOrmer.AssertExpectations(t)
}

func TestAsociarProducto_ProductoNotFound(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

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

	mockOfertOrmer.On("Read", mock.AnythingOfType("*models.Oferta"), []string(nil)).Return(nil).Once()
	mockOfertOrmer.On("Read", mock.AnythingOfType("*models.Producto"), []string(nil)).Return(orm.ErrNoRows).Once()

	controller.AsociarProducto()

	assert.Equal(t, http.StatusOK, recorder.Code)
	var resp models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &resp)
	assert.Equal(t, http.StatusNotFound, resp.Code)
	assert.Equal(t, "Producto no encontrado", resp.Message)
	mockOfertOrmer.AssertExpectations(t)
}

func TestAsociarProducto_ReadProductoError(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

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

	mockOfertOrmer.On("Read", mock.AnythingOfType("*models.Oferta"), []string(nil)).Return(nil).Once()
	mockOfertOrmer.On("Read", mock.AnythingOfType("*models.Producto"), []string(nil)).Return(fmt.Errorf("database error")).Once()

	controller.AsociarProducto()

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	mockOfertOrmer.AssertExpectations(t)
}

func TestDesasociarProducto_InvalidOfertaId(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	req := httptest.NewRequest("DELETE", "/ofertas/productos?id=invalid&producto_id=1", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	controller.DesasociarProducto()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestDesasociarProducto_InvalidProductoId(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	req := httptest.NewRequest("DELETE", "/ofertas/productos?id=1&producto_id=invalid", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	controller.DesasociarProducto()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestDesasociarProducto_AssociationNotFound(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	req := httptest.NewRequest("DELETE", "/ofertas/productos?id=1&producto_id=999", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	mockOfertOrmer.On("QueryTable", "oferta_producto").Return(mockOfertQS)
	mockOfertQS.On("Filter", "pk_id_oferta", []interface{}{int64(1)}).Return(mockOfertQS)
	mockOfertQS.On("Filter", "pk_id_producto", []interface{}{int64(999)}).Return(mockOfertQS)
	mockOfertQS.On("One", mock.AnythingOfType("*models.OfertaProducto")).Return(orm.ErrNoRows)

	controller.DesasociarProducto()

	assert.Equal(t, http.StatusOK, recorder.Code)
	var resp models.ApiResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &resp)
	assert.Equal(t, http.StatusNotFound, resp.Code)
	mockOfertOrmer.AssertExpectations(t)
	mockOfertQS.AssertExpectations(t)
}

func TestDesasociarProducto_QueryError(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	req := httptest.NewRequest("DELETE", "/ofertas/productos?id=1&producto_id=1", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	mockOfertOrmer.On("QueryTable", "oferta_producto").Return(mockOfertQS)
	mockOfertQS.On("Filter", "pk_id_oferta", []interface{}{int64(1)}).Return(mockOfertQS)
	mockOfertQS.On("Filter", "pk_id_producto", []interface{}{int64(1)}).Return(mockOfertQS)
	mockOfertQS.On("One", mock.AnythingOfType("*models.OfertaProducto")).Return(fmt.Errorf("database error"))

	controller.DesasociarProducto()

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	mockOfertOrmer.AssertExpectations(t)
	mockOfertQS.AssertExpectations(t)
}

func TestDesasociarProducto_DeleteError(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	req := httptest.NewRequest("DELETE", "/ofertas/productos?id=1&producto_id=1", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	mockOfertOrmer.On("QueryTable", "oferta_producto").Return(mockOfertQS)
	mockOfertQS.On("Filter", "pk_id_oferta", []interface{}{int64(1)}).Return(mockOfertQS)
	mockOfertQS.On("Filter", "pk_id_producto", []interface{}{int64(1)}).Return(mockOfertQS)
	mockOfertQS.On("One", mock.AnythingOfType("*models.OfertaProducto")).Return(nil)
	mockOfertOrmer.On("Delete", mock.AnythingOfType("*models.OfertaProducto"), []string(nil)).Return(int64(0), fmt.Errorf("delete error"))

	controller.DesasociarProducto()

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	mockOfertOrmer.AssertExpectations(t)
	mockOfertQS.AssertExpectations(t)
}
