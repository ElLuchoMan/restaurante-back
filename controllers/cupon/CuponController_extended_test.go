package cupon

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

// Tests adicionales para aumentar cobertura

func TestGetAll_LimitExceedsMax(t *testing.T) {
	controller, recorder, ctx := setupTest()

	// Configurar query parameters con límite > 100
	ctx.Input.SetParam("limit", "200")

	// Configurar mocks
	mockOrmer.On("QueryTable", "cupon").Return(mockQS)
	mockQS.On("Count").Return(int64(10), nil)
	mockQS.On("OrderBy", []string{"-pk_id_cupon"}).Return(mockQS)
	mockQS.On("Limit", 100).Return(mockQS) // Debe limitar a 100
	mockQS.On("Offset", int64(0)).Return(mockQS)
	mockQS.On("All", mock.AnythingOfType("*[]*models.Cupon"), []string(nil)).Return(int64(10), nil)

	// Ejecutar
	controller.GetAll()

	// Verificar
	assert.Equal(t, http.StatusOK, recorder.Code)
	mockOrmer.AssertExpectations(t)
	mockQS.AssertExpectations(t)
}

func TestGetAll_WithInvalidActivoBool(t *testing.T) {
	controller, recorder, ctx := setupTest()

	// Configurar query parameter con valor inválido para activo
	ctx.Input.SetParam("activo", "invalid")

	// Configurar mocks (no debería filtrar por activo)
	mockOrmer.On("QueryTable", "cupon").Return(mockQS)
	mockQS.On("Count").Return(int64(1), nil)
	mockQS.On("OrderBy", []string{"-pk_id_cupon"}).Return(mockQS)
	mockQS.On("Limit", 20).Return(mockQS)
	mockQS.On("Offset", int64(0)).Return(mockQS)
	mockQS.On("All", mock.AnythingOfType("*[]*models.Cupon"), []string(nil)).Return(int64(1), nil)

	// Ejecutar
	controller.GetAll()

	// Verificar
	assert.Equal(t, http.StatusOK, recorder.Code)
	mockOrmer.AssertExpectations(t)
	mockQS.AssertExpectations(t)
}

func TestPost_InvalidScope(t *testing.T) {
	controller, recorder, ctx := setupTest()

	// Cupón con scope inválido
	cupon := map[string]interface{}{
		"codigo":         "TEST",
		"scope":          "INVALID_SCOPE",
		"tipoDescuento":  "PORCENTAJE",
		"valorDescuento": 10,
		"fechaInicio":    "2025-01-01",
		"fechaFin":       "2025-12-31",
	}

	body, _ := json.Marshal(cupon)
	ctx.Request = httptest.NewRequest("POST", "/cupones", bytes.NewBuffer(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Input.RequestBody = body

	// Ejecutar
	controller.Post()

	// Verificar
	assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
}

func TestPost_InvalidTipoDescuento(t *testing.T) {
	controller, recorder, ctx := setupTest()

	// Cupón con tipo descuento inválido
	cupon := map[string]interface{}{
		"codigo":         "TEST",
		"scope":          "GLOBAL",
		"tipoDescuento":  "INVALID_TYPE",
		"valorDescuento": 10,
		"fechaInicio":    "2025-01-01",
		"fechaFin":       "2025-12-31",
	}

	body, _ := json.Marshal(cupon)
	ctx.Request = httptest.NewRequest("POST", "/cupones", bytes.NewBuffer(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Input.RequestBody = body

	// Ejecutar
	controller.Post()

	// Verificar
	assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
}

func TestPost_InvalidFechaInicio(t *testing.T) {
	controller, recorder, ctx := setupTest()

	// Cupón con fecha de inicio inválida
	cupon := map[string]interface{}{
		"codigo":         "TEST",
		"scope":          "GLOBAL",
		"tipoDescuento":  "PORCENTAJE",
		"valorDescuento": 10,
		"fechaInicio":    "invalid-date",
		"fechaFin":       "2025-12-31",
	}

	body, _ := json.Marshal(cupon)
	ctx.Request = httptest.NewRequest("POST", "/cupones", bytes.NewBuffer(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Input.RequestBody = body

	// Ejecutar
	controller.Post()

	// Verificar
	assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
}

func TestPost_InvalidFechaFin(t *testing.T) {
	controller, recorder, ctx := setupTest()

	// Cupón con fecha de fin inválida
	cupon := map[string]interface{}{
		"codigo":         "TEST",
		"scope":          "GLOBAL",
		"tipoDescuento":  "PORCENTAJE",
		"valorDescuento": 10,
		"fechaInicio":    "2025-01-01",
		"fechaFin":       "invalid-date",
	}

	body, _ := json.Marshal(cupon)
	ctx.Request = httptest.NewRequest("POST", "/cupones", bytes.NewBuffer(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Input.RequestBody = body

	// Ejecutar
	controller.Post()

	// Verificar
	assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
}

func TestGetById_SearchByCodigo(t *testing.T) {
	controller, recorder, _ := setupTest()

	// Crear request con código (no numérico)
	req := httptest.NewRequest("GET", "/cupones?id=TESTCODE", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Configurar mock para búsqueda por código
	cupon := models.Cupon{PkIdCupon: 1, Codigo: "TESTCODE", Scope: "GLOBAL"}
	mockOrmer.On("QueryTable", "cupon").Return(mockQS)
	mockQS.On("Filter", "codigo", []interface{}{"TESTCODE"}).Return(mockQS)
	mockQS.On("One", mock.AnythingOfType("*models.Cupon")).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.Cupon)
		*arg = cupon
	}).Return(nil)

	// Ejecutar
	controller.GetById()

	// Verificar
	assert.Equal(t, http.StatusOK, recorder.Code)
	mockOrmer.AssertExpectations(t)
	mockQS.AssertExpectations(t)
}

func TestGetById_EmptyID(t *testing.T) {
	controller, recorder, _ := setupTest()

	// Crear request sin ID
	req := httptest.NewRequest("GET", "/cupones", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Ejecutar
	controller.GetById()

	// Verificar
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestGetById_ReadError(t *testing.T) {
	controller, recorder, _ := setupTest()

	// Crear request con ID válido
	req := httptest.NewRequest("GET", "/cupones?id=1", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Configurar mock para error de base de datos
	mockOrmer.On("Read", mock.AnythingOfType("*models.Cupon"), []string(nil)).Return(fmt.Errorf("database error"))

	// Ejecutar
	controller.GetById()

	// Verificar
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	mockOrmer.AssertExpectations(t)
}

func TestGetById_QueryErrorAfterNoRows(t *testing.T) {
	controller, recorder, _ := setupTest()

	// Crear request con ID válido
	req := httptest.NewRequest("GET", "/cupones?id=999", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Configurar mock: no encontrado por ID, error al buscar por código
	mockOrmer.On("Read", mock.AnythingOfType("*models.Cupon"), []string(nil)).Return(orm.ErrNoRows)
	mockOrmer.On("QueryTable", "cupon").Return(mockQS)
	mockQS.On("Filter", "codigo", []interface{}{"999"}).Return(mockQS)
	mockQS.On("One", mock.AnythingOfType("*models.Cupon")).Return(fmt.Errorf("query error"))

	// Ejecutar
	controller.GetById()

	// Verificar
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	mockOrmer.AssertExpectations(t)
	mockQS.AssertExpectations(t)
}

func TestPut_InvalidJSON(t *testing.T) {
	controller, recorder, _ := setupTest()

	// Request con JSON inválido
	req := httptest.NewRequest("PUT", "/cupones?id=1", bytes.NewBuffer([]byte("invalid json")))
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

func TestPut_InvalidScope(t *testing.T) {
	controller, recorder, _ := setupTest()

	// Cupón con scope inválido
	cupon := map[string]interface{}{
		"codigo":         "TEST",
		"scope":          "INVALID_SCOPE",
		"tipoDescuento":  "PORCENTAJE",
		"valorDescuento": 10,
		"fechaInicio":    "2025-01-01",
		"fechaFin":       "2025-12-31",
	}

	body, _ := json.Marshal(cupon)
	req := httptest.NewRequest("PUT", "/cupones?id=1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = body
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Configurar mock para leer cupón existente
	existingCupon := models.Cupon{PkIdCupon: 1, Codigo: "OLD", Scope: "GLOBAL"}
	mockOrmer.On("Read", mock.AnythingOfType("*models.Cupon"), []string(nil)).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.Cupon)
		*arg = existingCupon
	}).Return(nil)

	// Ejecutar
	controller.Put()

	// Verificar
	assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
	mockOrmer.AssertExpectations(t)
}

func TestPut_InvalidTipoDescuento(t *testing.T) {
	controller, recorder, _ := setupTest()

	// Cupón con tipo descuento inválido
	cupon := map[string]interface{}{
		"codigo":         "TEST",
		"scope":          "GLOBAL",
		"tipoDescuento":  "INVALID_TYPE",
		"valorDescuento": 10,
		"fechaInicio":    "2025-01-01",
		"fechaFin":       "2025-12-31",
	}

	body, _ := json.Marshal(cupon)
	req := httptest.NewRequest("PUT", "/cupones?id=1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = body
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Configurar mock para leer cupón existente
	existingCupon := models.Cupon{PkIdCupon: 1, Codigo: "OLD", Scope: "GLOBAL"}
	mockOrmer.On("Read", mock.AnythingOfType("*models.Cupon"), []string(nil)).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.Cupon)
		*arg = existingCupon
	}).Return(nil)

	// Ejecutar
	controller.Put()

	// Verificar
	assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
	mockOrmer.AssertExpectations(t)
}

func TestPut_InvalidFechaInicio(t *testing.T) {
	controller, recorder, _ := setupTest()

	// Cupón con fecha de inicio inválida
	cupon := map[string]interface{}{
		"codigo":         "TEST",
		"scope":          "GLOBAL",
		"tipoDescuento":  "PORCENTAJE",
		"valorDescuento": 10,
		"fechaInicio":    "invalid-date",
		"fechaFin":       "2025-12-31",
	}

	body, _ := json.Marshal(cupon)
	req := httptest.NewRequest("PUT", "/cupones?id=1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = body
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Configurar mock para leer cupón existente
	existingCupon := models.Cupon{PkIdCupon: 1, Codigo: "OLD", Scope: "GLOBAL"}
	mockOrmer.On("Read", mock.AnythingOfType("*models.Cupon"), []string(nil)).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.Cupon)
		*arg = existingCupon
	}).Return(nil)

	// Ejecutar
	controller.Put()

	// Verificar
	assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
	mockOrmer.AssertExpectations(t)
}

func TestPut_InvalidFechaFin(t *testing.T) {
	controller, recorder, _ := setupTest()

	// Cupón con fecha de fin inválida
	cupon := map[string]interface{}{
		"codigo":         "TEST",
		"scope":          "GLOBAL",
		"tipoDescuento":  "PORCENTAJE",
		"valorDescuento": 10,
		"fechaInicio":    "2025-01-01",
		"fechaFin":       "invalid-date",
	}

	body, _ := json.Marshal(cupon)
	req := httptest.NewRequest("PUT", "/cupones?id=1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = body
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Configurar mock para leer cupón existente
	existingCupon := models.Cupon{PkIdCupon: 1, Codigo: "OLD", Scope: "GLOBAL"}
	mockOrmer.On("Read", mock.AnythingOfType("*models.Cupon"), []string(nil)).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.Cupon)
		*arg = existingCupon
	}).Return(nil)

	// Ejecutar
	controller.Put()

	// Verificar
	assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
	mockOrmer.AssertExpectations(t)
}

func TestPut_ValidationError(t *testing.T) {
	controller, recorder, _ := setupTest()

	// Cupón con validación que fallará (porcentaje > 100)
	cupon := map[string]interface{}{
		"codigo":         "",
		"scope":          "GLOBAL",
		"tipoDescuento":  "PORCENTAJE",
		"valorDescuento": 150,
		"fechaInicio":    "2025-01-01",
		"fechaFin":       "2025-12-31",
	}

	body, _ := json.Marshal(cupon)
	req := httptest.NewRequest("PUT", "/cupones?id=1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = body
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Configurar mock para leer cupón existente
	existingCupon := models.Cupon{PkIdCupon: 1, Codigo: "OLD", Scope: "GLOBAL"}
	mockOrmer.On("Read", mock.AnythingOfType("*models.Cupon"), []string(nil)).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.Cupon)
		*arg = existingCupon
	}).Return(nil)

	// Ejecutar
	controller.Put()

	// Verificar
	assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
	mockOrmer.AssertExpectations(t)
}

func TestPut_UpdateError(t *testing.T) {
	controller, recorder, _ := setupTest()

	// Cupón válido
	cupon := map[string]interface{}{
		"codigo":         "UPDATED",
		"scope":          "GLOBAL",
		"tipoDescuento":  "PORCENTAJE",
		"valorDescuento": 20,
		"fechaInicio":    "2025-01-01",
		"fechaFin":       "2025-12-31",
	}

	body, _ := json.Marshal(cupon)
	req := httptest.NewRequest("PUT", "/cupones?id=1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = body
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Configurar mocks
	existingCupon := models.Cupon{PkIdCupon: 1, Codigo: "OLD", Scope: "GLOBAL"}
	mockOrmer.On("Read", mock.AnythingOfType("*models.Cupon"), []string(nil)).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.Cupon)
		*arg = existingCupon
	}).Return(nil)
	mockOrmer.On("Update", mock.AnythingOfType("*models.Cupon"), []string(nil)).Return(int64(0), fmt.Errorf("update error"))

	// Ejecutar
	controller.Put()

	// Verificar
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	mockOrmer.AssertExpectations(t)
}

func TestPut_ReadError(t *testing.T) {
	controller, recorder, _ := setupTest()

	// Cupón válido
	cupon := map[string]interface{}{
		"codigo": "TEST",
	}

	body, _ := json.Marshal(cupon)
	req := httptest.NewRequest("PUT", "/cupones?id=1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = body
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Configurar mock para error de lectura (no ErrNoRows)
	mockOrmer.On("Read", mock.AnythingOfType("*models.Cupon"), []string(nil)).Return(fmt.Errorf("database error"))

	// Ejecutar
	controller.Put()

	// Verificar
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	mockOrmer.AssertExpectations(t)
}

func TestDelete_NotFound(t *testing.T) {
	controller, recorder, _ := setupTest()

	// Crear request con ID
	req := httptest.NewRequest("DELETE", "/cupones?id=999", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Configurar mock para no encontrado
	mockOrmer.On("Read", mock.AnythingOfType("*models.Cupon"), []string(nil)).Return(orm.ErrNoRows)

	// Ejecutar
	controller.Delete()

	// Verificar
	assert.Equal(t, http.StatusNotFound, recorder.Code)
	mockOrmer.AssertExpectations(t)
}

func TestDelete_AlreadyInactive(t *testing.T) {
	controller, recorder, _ := setupTest()

	// Crear request con ID
	req := httptest.NewRequest("DELETE", "/cupones?id=1", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Configurar mock para cupón ya inactivo
	cupon := models.Cupon{PkIdCupon: 1, Codigo: "TEST1", Activo: false}
	mockOrmer.On("Read", mock.AnythingOfType("*models.Cupon"), []string(nil)).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.Cupon)
		*arg = cupon
	}).Return(nil)

	// Ejecutar
	controller.Delete()

	// Verificar
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	mockOrmer.AssertExpectations(t)
}

func TestDelete_ReadError(t *testing.T) {
	controller, recorder, _ := setupTest()

	// Crear request con ID
	req := httptest.NewRequest("DELETE", "/cupones?id=1", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Configurar mock para error de lectura
	mockOrmer.On("Read", mock.AnythingOfType("*models.Cupon"), []string(nil)).Return(fmt.Errorf("database error"))

	// Ejecutar
	controller.Delete()

	// Verificar
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	mockOrmer.AssertExpectations(t)
}

func TestDelete_UpdateError(t *testing.T) {
	controller, recorder, _ := setupTest()

	// Crear request con ID
	req := httptest.NewRequest("DELETE", "/cupones?id=1", nil)
	ctx := context.NewContext()
	ctx.Reset(recorder, req)
	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	// Configurar mocks
	cupon := models.Cupon{PkIdCupon: 1, Codigo: "TEST1", Activo: true}
	mockOrmer.On("Read", mock.AnythingOfType("*models.Cupon"), []string(nil)).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.Cupon)
		*arg = cupon
	}).Return(nil)
	mockOrmer.On("Update", mock.AnythingOfType("*models.Cupon"), []string{"Activo"}).Return(int64(0), fmt.Errorf("update error"))

	// Ejecutar
	controller.Delete()

	// Verificar
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	mockOrmer.AssertExpectations(t)
}

func TestListarRedenciones_WithCuponCodigoFilter(t *testing.T) {
	controller, recorder, ctx := setupTest()

	// Configurar query parameter
	ctx.Input.SetParam("cupon_codigo", "TEST1")

	// Configurar mocks
	cupon := models.Cupon{PkIdCupon: 1, Codigo: "TEST1"}
	mockOrmer.On("QueryTable", "cupon").Return(mockQS)
	mockQS.On("Filter", "codigo", []interface{}{"TEST1"}).Return(mockQS)
	mockQS.On("One", mock.AnythingOfType("*models.Cupon")).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.Cupon)
		*arg = cupon
	}).Return(nil)

	mockOrmer.On("QueryTable", "cupon_redencion").Return(mockQS)
	mockQS.On("Filter", "pk_id_cupon", []interface{}{int64(1)}).Return(mockQS)
	mockQS.On("Count").Return(int64(1), nil)
	mockQS.On("OrderBy", []string{"-created_at"}).Return(mockQS)
	mockQS.On("Limit", 20).Return(mockQS)
	mockQS.On("Offset", int64(0)).Return(mockQS)
	mockQS.On("All", mock.AnythingOfType("*[]*models.CuponRedencion"), []string(nil)).Return(int64(1), nil)

	// Ejecutar
	controller.ListarRedenciones()

	// Verificar
	assert.Equal(t, http.StatusOK, recorder.Code)
	mockOrmer.AssertExpectations(t)
	mockQS.AssertExpectations(t)
}

func TestListarRedenciones_CuponNotFound(t *testing.T) {
	controller, recorder, ctx := setupTest()

	// Configurar query parameter
	ctx.Input.SetParam("cupon_codigo", "NOTFOUND")

	// Configurar mocks - El controller llama primero a QueryTable("cupon_redencion"), luego busca el cupón
	mockOrmer.On("QueryTable", "cupon_redencion").Return(mockQS).Once()
	mockOrmer.On("QueryTable", "cupon").Return(mockQS).Once()
	mockQS.On("Filter", "codigo", []interface{}{"NOTFOUND"}).Return(mockQS).Once()
	mockQS.On("One", mock.AnythingOfType("*models.Cupon")).Return(orm.ErrNoRows).Once()

	// Ejecutar
	controller.ListarRedenciones()

	// Verificar - debería retornar lista vacía, no error
	assert.Equal(t, http.StatusOK, recorder.Code)
	mockOrmer.AssertExpectations(t)
	mockQS.AssertExpectations(t)
}

func TestListarRedenciones_WithCuponIdAndClienteIdFilters(t *testing.T) {
	controller, recorder, ctx := setupTest()

	// Configurar query parameters
	ctx.Input.SetParam("cupon_id", "1")
	ctx.Input.SetParam("cliente_id", "123")

	// Configurar mocks
	mockOrmer.On("QueryTable", "cupon_redencion").Return(mockQS)
	mockQS.On("Filter", "pk_id_cupon", []interface{}{int64(1)}).Return(mockQS)
	mockQS.On("Filter", "pk_documento_cliente", []interface{}{int64(123)}).Return(mockQS)
	mockQS.On("Count").Return(int64(1), nil)
	mockQS.On("OrderBy", []string{"-created_at"}).Return(mockQS)
	mockQS.On("Limit", 20).Return(mockQS)
	mockQS.On("Offset", int64(0)).Return(mockQS)
	mockQS.On("All", mock.AnythingOfType("*[]*models.CuponRedencion"), []string(nil)).Return(int64(1), nil)

	// Ejecutar
	controller.ListarRedenciones()

	// Verificar
	assert.Equal(t, http.StatusOK, recorder.Code)
	mockOrmer.AssertExpectations(t)
	mockQS.AssertExpectations(t)
}

func TestListarRedenciones_CountError(t *testing.T) {
	controller, recorder, _ := setupTest()

	// Configurar mocks
	mockOrmer.On("QueryTable", "cupon_redencion").Return(mockQS)
	mockQS.On("Count").Return(int64(0), fmt.Errorf("database error"))

	// Ejecutar
	controller.ListarRedenciones()

	// Verificar
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	mockOrmer.AssertExpectations(t)
	mockQS.AssertExpectations(t)
}

func TestListarRedenciones_QueryError(t *testing.T) {
	controller, recorder, _ := setupTest()

	// Configurar mocks
	mockOrmer.On("QueryTable", "cupon_redencion").Return(mockQS)
	mockQS.On("Count").Return(int64(1), nil)
	mockQS.On("OrderBy", []string{"-created_at"}).Return(mockQS)
	mockQS.On("Limit", 20).Return(mockQS)
	mockQS.On("Offset", int64(0)).Return(mockQS)
	mockQS.On("All", mock.AnythingOfType("*[]*models.CuponRedencion"), []string(nil)).Return(int64(0), fmt.Errorf("query error"))

	// Ejecutar
	controller.ListarRedenciones()

	// Verificar
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	mockOrmer.AssertExpectations(t)
	mockQS.AssertExpectations(t)
}

// Tests de RedimirCupon requieren refactorización del controller para inyectar el servicio completo
// Estos casos están cubiertos por los tests existentes de ValidarCupon y RedimirCupon en CuponController_test.go
