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

func TestOfertaGetAll_QueryError(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	mockOfertOrmer.On("QueryTable", "oferta").Return(mockOfertQS)
	mockOfertQS.On("Count").Return(int64(10), nil)
	mockOfertQS.On("OrderBy", []string{"-pk_id_oferta"}).Return(mockOfertQS)
	mockOfertQS.On("Limit", 20).Return(mockOfertQS)
	mockOfertQS.On("Offset", int64(0)).Return(mockOfertQS)
	mockOfertQS.On("All", mock.AnythingOfType("*[]*models.Oferta"), []string(nil)).Return(int64(0), fmt.Errorf("query error"))

	controller.GetAll()

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	mockOfertOrmer.AssertExpectations(t)
	mockOfertQS.AssertExpectations(t)
}

func TestOfertaPost_DatabaseError(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

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

	mockOfertOrmer.On("Insert", mock.AnythingOfType("*models.Oferta")).Return(int64(0), fmt.Errorf("database error"))

	controller.Post()

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	mockOfertOrmer.AssertExpectations(t)
}

func TestOfertaPut_InvalidID(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

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

	controller.Put()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestOfertaPut_NotFound(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

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

	mockOfertOrmer.On("Read", mock.AnythingOfType("*models.Oferta"), []string(nil)).Return(orm.ErrNoRows)

	controller.Put()

	assert.Equal(t, http.StatusNotFound, recorder.Code)
	mockOfertOrmer.AssertExpectations(t)
}

func TestOfertaPut_InvalidJSON(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	req := httptest.NewRequest("PUT", "/ofertas?id=1", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = []byte("invalid json")
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	controller.Put()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestOfertaDelete_InvalidID(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	req := httptest.NewRequest("DELETE", "/ofertas?id=invalid", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	controller.Delete()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestOfertaDelete_NotFound(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	req := httptest.NewRequest("DELETE", "/ofertas?id=999", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	mockOfertOrmer.On("Read", mock.AnythingOfType("*models.Oferta"), []string(nil)).Return(orm.ErrNoRows)

	controller.Delete()

	assert.Equal(t, http.StatusNotFound, recorder.Code)
	mockOfertOrmer.AssertExpectations(t)
}

func TestOfertaDelete_AlreadyInactive(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	req := httptest.NewRequest("DELETE", "/ofertas?id=1", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	oferta := models.Oferta{PkIdOferta: 1, Titulo: "Test", Activo: false}
	mockOfertOrmer.On("Read", mock.AnythingOfType("*models.Oferta"), []string(nil)).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.Oferta)
		*arg = oferta
	}).Return(nil)

	controller.Delete()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	mockOfertOrmer.AssertExpectations(t)
}

func TestOfertaDelete_UpdateError(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	req := httptest.NewRequest("DELETE", "/ofertas?id=1", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	oferta := models.Oferta{PkIdOferta: 1, Titulo: "Test", Activo: true}
	mockOfertOrmer.On("Read", mock.AnythingOfType("*models.Oferta"), []string(nil)).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.Oferta)
		*arg = oferta
	}).Return(nil)
	mockOfertOrmer.On("Update", mock.AnythingOfType("*models.Oferta"), []string{"Activo"}).Return(int64(0), fmt.Errorf("update error"))

	controller.Delete()

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	mockOfertOrmer.AssertExpectations(t)
}

func TestOfertaAsociarProducto_MissingOfertaID(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

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

	controller.AsociarProducto()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestOfertaAsociarProducto_InvalidJSON(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	req := httptest.NewRequest("POST", "/ofertas/asociar-producto?oferta_id=1", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = []byte("invalid json")
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	controller.AsociarProducto()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestOfertaDesasociarProducto_MissingOfertaID(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	req := httptest.NewRequest("DELETE", "/ofertas/desasociar-producto?producto_id=1", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	controller.DesasociarProducto()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestOfertaDesasociarProducto_MissingProductoID(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	req := httptest.NewRequest("DELETE", "/ofertas/desasociar-producto?oferta_id=1", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	controller.DesasociarProducto()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestOfertaGetById_ReadError(t *testing.T) {
	controller, recorder, _ := setupOfertaTest()

	req := httptest.NewRequest("GET", "/ofertas?id=1", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	mockOfertOrmer.On("Read", mock.AnythingOfType("*models.Oferta"), []string(nil)).Return(fmt.Errorf("database error"))

	controller.GetById()

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	mockOfertOrmer.AssertExpectations(t)
}

func TestOfertaGetAll_LimitExceedsMax(t *testing.T) {
	controller, recorder, ctx := setupOfertaTest()

	ctx.Input.SetParam("limit", "200")

	mockOfertOrmer.On("QueryTable", "oferta").Return(mockOfertQS)
	mockOfertQS.On("Count").Return(int64(10), nil)
	mockOfertQS.On("OrderBy", []string{"-pk_id_oferta"}).Return(mockOfertQS)
	mockOfertQS.On("Limit", 100).Return(mockOfertQS)
	mockOfertQS.On("Offset", int64(0)).Return(mockOfertQS)
	mockOfertQS.On("All", mock.AnythingOfType("*[]*models.Oferta"), []string(nil)).Return(int64(10), nil)

	controller.GetAll()

	assert.Equal(t, http.StatusOK, recorder.Code)
	mockOfertOrmer.AssertExpectations(t)
	mockOfertQS.AssertExpectations(t)
}
