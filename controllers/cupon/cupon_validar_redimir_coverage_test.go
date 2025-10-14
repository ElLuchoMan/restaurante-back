package cupon

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	beecontext "github.com/beego/beego/v2/server/web/context"
	"github.com/stretchr/testify/assert"
)

func TestValidarCupon_JSONInvalido(t *testing.T) {
	ctrl := &CuponController{}
	ctrl.Data = make(map[interface{}]interface{})
	req := httptest.NewRequest(http.MethodPost, "/cupones/validar", bytes.NewReader([]byte("{invalid json")))
	recorder := httptest.NewRecorder()

	ctx := beecontext.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = []byte("{invalid json")
	ctrl.Ctx = ctx

	ctrl.ValidarCupon()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response models.ApiResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Contains(t, response.Message, "JSON inválido")
}

func TestValidarCupon_ErrorDelServicio(t *testing.T) {

	originalOrmProvider := ormProvider
	defer func() { ormProvider = originalOrmProvider }()

	ormProvider = func() orm.Ormer {
		return nil
	}

	ctrl := &CuponController{}
	ctrl.Data = make(map[interface{}]interface{})

	reqBody := models.ValidarCuponRequest{
		Codigo:    "TEST",
		ClienteId: 123,
		Items: []models.ValidarCuponItemRequest{
			{ProductoId: 1, Cantidad: 1, Precio: 1000},
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/cupones/validar", bytes.NewReader(bodyBytes))
	recorder := httptest.NewRecorder()

	ctx := beecontext.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = bodyBytes
	ctrl.Ctx = ctx

	ctrl.ValidarCupon()

	assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)

	var response models.ApiResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, response.Code)
}

func TestValidarCupon_Exitoso(t *testing.T) {

	originalOrmProvider := ormProvider
	defer func() { ormProvider = originalOrmProvider }()

	ormProvider = func() orm.Ormer {
		return nil
	}

	ctrl := &CuponController{}
	ctrl.Data = make(map[interface{}]interface{})

	reqBody := models.ValidarCuponRequest{
		Codigo:    "VALIDO",
		ClienteId: 123,
		Items: []models.ValidarCuponItemRequest{
			{ProductoId: 1, Cantidad: 1, Precio: 1000},
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/cupones/validar", bytes.NewReader(bodyBytes))
	recorder := httptest.NewRecorder()

	ctx := beecontext.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = bodyBytes
	ctrl.Ctx = ctx

	ctrl.ValidarCupon()

	var response models.ApiResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
}

func TestRedimirCupon_JSONInvalido(t *testing.T) {
	ctrl := &CuponController{}
	ctrl.Data = make(map[interface{}]interface{})
	req := httptest.NewRequest(http.MethodPost, "/cupones/TEST123/redimir", bytes.NewReader([]byte("{invalid json")))
	recorder := httptest.NewRecorder()

	ctx := beecontext.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = []byte("{invalid json")
	ctx.Input.SetParam(":codigo", "TEST123")
	ctrl.Ctx = ctx

	ctrl.RedimirCupon()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response models.ApiResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Contains(t, response.Message, "JSON inválido")
}

func TestRedimirCupon_CuponNoEncontrado(t *testing.T) {
	originalOrmProvider := ormProvider
	defer func() { ormProvider = originalOrmProvider }()

	ormProvider = func() orm.Ormer {
		return nil
	}

	ctrl := &CuponController{}
	ctrl.Data = make(map[interface{}]interface{})

	reqBody := models.RedimirCuponRequest{
		ClienteId: 123,
		PedidoId:  nil,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/cupones/NOEXISTE/redimir", bytes.NewReader(bodyBytes))
	recorder := httptest.NewRecorder()

	ctx := beecontext.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = bodyBytes
	ctx.Input.SetParam(":codigo", "NOEXISTE")
	ctrl.Ctx = ctx

	ctrl.RedimirCupon()

	var response models.ApiResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
}

func TestRedimirCupon_CuponNoAplicable(t *testing.T) {
	originalOrmProvider := ormProvider
	defer func() { ormProvider = originalOrmProvider }()

	ormProvider = func() orm.Ormer {
		return nil
	}

	ctrl := &CuponController{}
	ctrl.Data = make(map[interface{}]interface{})

	reqBody := models.RedimirCuponRequest{
		ClienteId: 123,
		PedidoId:  nil,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/cupones/EXPIRADO/redimir", bytes.NewReader(bodyBytes))
	recorder := httptest.NewRecorder()

	ctx := beecontext.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = bodyBytes
	ctx.Input.SetParam(":codigo", "EXPIRADO")
	ctrl.Ctx = ctx

	ctrl.RedimirCupon()

	var response models.ApiResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
}

func TestRedimirCupon_ErrorGenerico(t *testing.T) {
	originalOrmProvider := ormProvider
	defer func() { ormProvider = originalOrmProvider }()

	ormProvider = func() orm.Ormer {
		return nil
	}

	ctrl := &CuponController{}
	ctrl.Data = make(map[interface{}]interface{})

	reqBody := models.RedimirCuponRequest{
		ClienteId: 123,
		PedidoId:  nil,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/cupones/ERROR/redimir", bytes.NewReader(bodyBytes))
	recorder := httptest.NewRecorder()

	ctx := beecontext.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = bodyBytes
	ctx.Input.SetParam(":codigo", "ERROR")
	ctrl.Ctx = ctx

	ctrl.RedimirCupon()

	var response models.ApiResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
}

func TestRedimirCupon_Exitoso(t *testing.T) {
	originalOrmProvider := ormProvider
	defer func() { ormProvider = originalOrmProvider }()

	ormProvider = func() orm.Ormer {
		return nil
	}

	ctrl := &CuponController{}
	ctrl.Data = make(map[interface{}]interface{})

	reqBody := models.RedimirCuponRequest{
		ClienteId: 123,
		PedidoId:  nil,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/cupones/VALIDO/redimir", bytes.NewReader(bodyBytes))
	recorder := httptest.NewRecorder()

	ctx := beecontext.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = bodyBytes
	ctx.Input.SetParam(":codigo", "VALIDO")
	ctrl.Ctx = ctx

	ctrl.RedimirCupon()

	var response models.ApiResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
}

func TestCuponServiceOrmFactory_Coverage(t *testing.T) {
	originalOrmProvider := ormProvider
	defer func() { ormProvider = originalOrmProvider }()

	called := false
	ormProvider = func() orm.Ormer {
		called = true
		return nil
	}

	result := cuponServiceOrmFactory()
	assert.Nil(t, result)
	assert.True(t, called, "ormProvider debe ser llamado")
}

func TestCupServiceOrmBase_Coverage(t *testing.T) {
	originalOrmProvider := ormProvider
	defer func() { ormProvider = originalOrmProvider }()

	called := false
	ormProvider = func() orm.Ormer {
		called = true
		return nil
	}

	result := cupServiceOrmBase()
	assert.Nil(t, result)
	assert.True(t, called, "ormProvider debe ser llamado")
}

func TestRedimirCupon_ConPedidoId(t *testing.T) {
	originalOrmProvider := ormProvider
	defer func() { ormProvider = originalOrmProvider }()

	ormProvider = func() orm.Ormer {
		return nil
	}

	ctrl := &CuponController{}
	ctrl.Data = make(map[interface{}]interface{})

	pedidoId := int64(123)
	reqBody := models.RedimirCuponRequest{
		ClienteId: 123,
		PedidoId:  &pedidoId,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/cupones/TEST/redimir", bytes.NewReader(bodyBytes))
	recorder := httptest.NewRecorder()

	ctx := beecontext.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = bodyBytes
	ctx.Input.SetParam(":codigo", "TEST")
	ctrl.Ctx = ctx

	ctrl.RedimirCupon()

	var response models.ApiResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
}
