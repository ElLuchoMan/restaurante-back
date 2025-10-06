package descuento

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

// Tests adicionales para aumentar cobertura de DescuentoController

func TestDescuentoPost_MissingPedidoId(t *testing.T) {
	controller, recorder, _ := setupDescuentoTest()

	// Preparar request body válido pero sin pedido_id
	cuponId := int64(1)
	payload := map[string]interface{}{
		"cuponId": cuponId,
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/descuentos/pedidos", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = body
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Ejecutar
	controller.Post()

	// Verificar
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestDescuentoPost_InvalidPedidoId(t *testing.T) {
	controller, recorder, _ := setupDescuentoTest()

	// Preparar request body válido pero con pedido_id inválido
	cuponId := int64(1)
	payload := map[string]interface{}{
		"cuponId": cuponId,
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/descuentos/pedidos?pedido_id=invalid", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = body
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Ejecutar
	controller.Post()

	// Verificar
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestDescuentoPost_BothCuponAndOferta(t *testing.T) {
	controller, recorder, _ := setupDescuentoTest()

	// Preparar request body con AMBOS cupón y oferta (inválido)
	cuponId := int64(1)
	ofertaId := int64(2)
	payload := map[string]interface{}{
		"cuponId":  cuponId,
		"ofertaId": ofertaId,
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/descuentos/pedidos?pedido_id=1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = body
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Ejecutar
	controller.Post()

	// Verificar - debe retornar 422
	assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)

	// Verificar mensaje
	var response models.ApiResponse
	json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Contains(t, response.Message, "exactamente uno de cupón o oferta")
}

func TestDescuentoPost_NeitherCuponNorOferta(t *testing.T) {
	controller, recorder, _ := setupDescuentoTest()

	// Preparar request body sin cupón ni oferta (inválido)
	payload := map[string]interface{}{
		"montoDescuento": 5000,
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/descuentos/pedidos?pedido_id=1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = body
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Ejecutar
	controller.Post()

	// Verificar - debe retornar 422
	assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)

	// Verificar mensaje
	var response models.ApiResponse
	json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Contains(t, response.Message, "exactamente uno de cupón o oferta")
}

func TestDescuentoPost_ZeroPedidoId(t *testing.T) {
	controller, recorder, _ := setupDescuentoTest()

	// Preparar request body válido pero con pedido_id = 0
	cuponId := int64(1)
	payload := map[string]interface{}{
		"cuponId": cuponId,
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/descuentos/pedidos?pedido_id=0", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = body
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Ejecutar
	controller.Post()

	// Verificar
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestDescuentoGetAll_ZeroPedidoId(t *testing.T) {
	controller, recorder, _ := setupDescuentoTest()

	// Configurar query parameter con pedido_id = 0
	req := httptest.NewRequest("GET", "/descuentos/pedidos?pedido_id=0", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Ejecutar
	controller.GetAll()

	// Verificar
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}
