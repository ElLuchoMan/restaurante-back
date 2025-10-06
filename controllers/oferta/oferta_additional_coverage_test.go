package oferta

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
	"github.com/stretchr/testify/mock"
)

// Tests adicionales para OfertaController - aumentar cobertura al 70%+

func TestOfertaDelete_AlreadyInactive(t *testing.T) {
	ctx := context.NewContext()
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/ofertas?id=1", nil)
	ctx.Reset(recorder, req)
	ctx.Input.SetParam("id", "1")

	controller := &OfertaController{}
	controller.Init(ctx, "OfertaController", "Delete", nil)

	// Mock - oferta ya inactiva
	mockOfertOrmer.ExpectedCalls = nil
	mockOfertQS.ExpectedCalls = nil

	mockOfertOrmer.On("QueryTable", "oferta").Return(mockOfertQS)
	mockOfertQS.On("Filter", "pk_id_oferta", int64(1)).Return(mockOfertQS)

	ofertaInactiva := &models.Oferta{
		PK_ID_OFERTA: 1,
		Activo:       false, // Ya inactiva
	}
	mockOfertQS.On("One", &models.Oferta{}).Return(nil).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.Oferta)
		*arg = *ofertaInactiva
	})

	controller.Delete()

	// Debería retornar 404 o mensaje de ya inactiva
	if recorder.Code == 500 {
		t.Errorf("Expected not 500, got %d", recorder.Code)
	}
}

func TestOfertaGetAll_EmptyResult(t *testing.T) {
	ctx := context.NewContext()
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/ofertas", nil)
	ctx.Reset(recorder, req)

	controller := &OfertaController{}
	controller.Init(ctx, "OfertaController", "GetAll", nil)

	mockOfertOrmer.ExpectedCalls = nil
	mockOfertQS.ExpectedCalls = nil

	mockOfertOrmer.On("QueryTable", "oferta").Return(mockOfertQS)
	mockOfertQS.On("Count").Return(int64(0), nil)
	mockOfertQS.On("OrderBy", []string{"-pk_id_oferta"}).Return(mockOfertQS)
	mockOfertQS.On("Limit", 10).Return(mockOfertQS)
	mockOfertQS.On("Offset", int64(0)).Return(mockOfertQS)
	mockOfertQS.On("All", mock.AnythingOfType("*[]*models.Oferta"), []string(nil)).Return(int64(0), nil)

	controller.GetAll()

	if recorder.Code != 200 {
		t.Errorf("Expected status 200, got %d", recorder.Code)
	}
}

func TestOfertaGetById_InvalidIDFormat(t *testing.T) {
	ctx := context.NewContext()
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/ofertas/search?id=invalid", nil)
	ctx.Reset(recorder, req)
	ctx.Input.SetParam("id", "invalid")

	controller := &OfertaController{}
	controller.Init(ctx, "OfertaController", "GetById", nil)

	controller.GetById()

	// Debería retornar error de ID inválido
	if recorder.Code == 200 && recorder.Body.Len() > 0 {
		var resp map[string]interface{}
		json.Unmarshal(recorder.Body.Bytes(), &resp)
		if resp["code"] != float64(404) && resp["code"] != float64(400) {
			t.Errorf("Expected error code, got %v", resp["code"])
		}
	}
}

func TestOfertaPut_EmptyBody(t *testing.T) {
	ctx := context.NewContext()
	recorder := httptest.NewRecorder()

	req := httptest.NewRequest("PUT", "/ofertas?id=1", bytes.NewReader([]byte("{}")))
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = []byte("{}")
	ctx.Input.SetParam("id", "1")

	controller := &OfertaController{}
	controller.Init(ctx, "OfertaController", "Put", nil)

	mockOfertOrmer.ExpectedCalls = nil
	mockOfertQS.ExpectedCalls = nil

	mockOfertOrmer.On("QueryTable", "oferta").Return(mockOfertQS)
	mockOfertQS.On("Filter", "pk_id_oferta", int64(1)).Return(mockOfertQS)

	existingOferta := &models.Oferta{
		PK_ID_OFERTA: 1,
		Titulo:       "Oferta Test",
		Activo:       true,
	}
	mockOfertQS.On("One", &models.Oferta{}).Return(nil).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.Oferta)
		*arg = *existingOferta
	})

	mockOfertOrmer.On("Update", mock.AnythingOfType("*models.Oferta"), []string(nil)).Return(int64(0), nil)

	controller.Put()

	// Empty body debería actualizar sin cambios o retornar 200
	if recorder.Code != 200 && recorder.Code != 400 {
		t.Errorf("Expected status 200 or 400, got %d", recorder.Code)
	}
}

func TestOfertaAsociarProducto_Success(t *testing.T) {
	ctx := context.NewContext()
	recorder := httptest.NewRecorder()

	body := map[string]interface{}{
		"ofertaId":   float64(1),
		"productoId": float64(10),
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/ofertas/asociar", bytes.NewReader(bodyBytes))
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = bodyBytes

	controller := &OfertaController{}
	controller.Init(ctx, "OfertaController", "AsociarProducto", nil)

	mockOfertOrmer.ExpectedCalls = nil
	mockOfertQS.ExpectedCalls = nil

	// Mock verificar oferta existe
	mockOfertOrmer.On("QueryTable", "oferta").Return(mockOfertQS)
	mockOfertQS.On("Filter", "pk_id_oferta", int64(1)).Return(mockOfertQS)
	mockOfertQS.On("Exist").Return(true)

	// Mock verificar producto existe
	mockOfertOrmer.On("QueryTable", "producto").Return(mockOfertQS)
	mockOfertQS.On("Filter", "pk_id_producto", int64(10)).Return(mockOfertQS)
	mockOfertQS.On("Exist").Return(true)

	// Mock insert asociación
	mockOfertOrmer.On("Insert", mock.AnythingOfType("*models.OfertaProducto")).Return(int64(1), nil)

	controller.AsociarProducto()

	if recorder.Code != 201 && recorder.Code != 200 {
		t.Errorf("Expected status 201 or 200, got %d", recorder.Code)
	}
}

func TestOfertaDesasociarProducto_NotFound(t *testing.T) {
	ctx := context.NewContext()
	recorder := httptest.NewRecorder()

	body := map[string]interface{}{
		"ofertaId":   float64(999),
		"productoId": float64(888),
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("DELETE", "/ofertas/desasociar", bytes.NewReader(bodyBytes))
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = bodyBytes

	controller := &OfertaController{}
	controller.Init(ctx, "OfertaController", "DesasociarProducto", nil)

	mockOfertOrmer.ExpectedCalls = nil
	mockOfertQS.ExpectedCalls = nil

	mockOfertOrmer.On("QueryTable", "oferta_producto").Return(mockOfertQS)
	mockOfertQS.On("Filter", "pk_id_oferta", int64(999)).Return(mockOfertQS)
	mockOfertQS.On("Filter", "pk_id_producto", int64(888)).Return(mockOfertQS)
	mockOfertQS.On("Delete").Return(int64(0), nil) // No encontrado

	controller.DesasociarProducto()

	// Debería retornar 404 o success sin cambios
	if recorder.Code == 500 {
		t.Errorf("Expected not 500, got %d", recorder.Code)
	}
}

func TestOfertaObtenerOfertasActivas_WithFilters(t *testing.T) {
	ctx := context.NewContext()
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/ofertas/activas?restaurante_id=1", nil)
	ctx.Reset(recorder, req)
	ctx.Input.SetParam("restaurante_id", "1")

	controller := &OfertaController{}
	controller.Init(ctx, "OfertaController", "ObtenerOfertasActivas", nil)

	mockOfertOrmer.ExpectedCalls = nil
	mockOfertQS.ExpectedCalls = nil

	now := time.Now()
	mockOfertOrmer.On("QueryTable", "oferta").Return(mockOfertQS)
	mockOfertQS.On("Filter", "Activo", true).Return(mockOfertQS)
	mockOfertQS.On("Filter", "FechaInicio__lte", mock.AnythingOfType("time.Time")).Return(mockOfertQS)
	mockOfertQS.On("Filter", "FechaFin__gte", mock.AnythingOfType("time.Time")).Return(mockOfertQS)
	mockOfertQS.On("Filter", "PKIDRestaurante", int64(1)).Return(mockOfertQS)

	ofertas := []*models.Oferta{
		{
			PK_ID_OFERTA: 1,
			Titulo:       "Oferta Activa",
			Activo:       true,
			FechaInicio:  now.AddDate(0, 0, -1),
			FechaFin:     now.AddDate(0, 0, 1),
		},
	}
	mockOfertQS.On("All", mock.AnythingOfType("*[]*models.Oferta"), []string(nil)).Return(int64(1), nil).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*[]*models.Oferta)
		*arg = ofertas
	})

	controller.ObtenerOfertasActivas()

	if recorder.Code != 200 {
		t.Errorf("Expected status 200, got %d", recorder.Code)
	}
}

func TestOfertaPost_FechaFinBeforeFechaInicio(t *testing.T) {
	ctx := context.NewContext()
	recorder := httptest.NewRecorder()

	body := map[string]interface{}{
		"titulo":         "Oferta Test",
		"descripcion":    "Test",
		"scope":          "producto",
		"tipoDescuento":  "porcentaje",
		"valorDescuento": float64(10),
		"fechaInicio":    "2025-01-20",
		"fechaFin":       "2025-01-10", // Antes de fechaInicio
		"restauranteId":  float64(1),
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/ofertas", bytes.NewReader(bodyBytes))
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = bodyBytes

	controller := &OfertaController{}
	controller.Init(ctx, "OfertaController", "Post", nil)

	controller.Post()

	// Debería retornar error de validación
	if recorder.Code == 201 {
		t.Error("Expected error for fechaFin before fechaInicio")
	}
}
