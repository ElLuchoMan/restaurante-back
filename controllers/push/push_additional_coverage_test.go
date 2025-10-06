package push

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"restaurante/models"

	"github.com/beego/beego/v2/server/web/context"
	"github.com/stretchr/testify/mock"
)

// Tests adicionales para PushController - aumentar cobertura al 65%+

func TestPushGetAll_WithOffset(t *testing.T) {
	ctx := context.NewContext()
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/push?page=2&limit=5", nil)
	ctx.Reset(recorder, req)
	ctx.Input.SetParam("page", "2")
	ctx.Input.SetParam("limit", "5")

	controller := &PushController{}
	controller.Init(ctx, "PushController", "GetAll", nil)

	mockPushOrm.ExpectedCalls = nil
	mockPushQS.ExpectedCalls = nil

	mockPushOrm.On("QueryTable", "push_dispositivo").Return(mockPushQS)
	mockPushQS.On("Count").Return(int64(20), nil)
	mockPushQS.On("OrderBy", []string{"-updated_at"}).Return(mockPushQS)
	mockPushQS.On("Limit", 5).Return(mockPushQS)
	mockPushQS.On("Offset", int64(5)).Return(mockPushQS) // offset = (page-1) * limit = 1*5 = 5
	mockPushQS.On("All", mock.AnythingOfType("*[]*models.PushDispositivo"), []string(nil)).Return(int64(5), nil)

	controller.GetAll()

	if recorder.Code != 200 {
		t.Errorf("Expected status 200, got %d", recorder.Code)
	}
}

func TestPushGetById_NotFound(t *testing.T) {
	ctx := context.NewContext()
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/push/search?id=999", nil)
	ctx.Reset(recorder, req)
	ctx.Input.SetParam("id", "999")

	controller := &PushController{}
	controller.Init(ctx, "PushController", "GetById", nil)

	mockPushOrm.ExpectedCalls = nil
	mockPushQS.ExpectedCalls = nil

	mockPushOrm.On("QueryTable", "push_dispositivo").Return(mockPushQS)
	mockPushQS.On("Filter", "pk_id_dispositivo", int64(999)).Return(mockPushQS)
	mockPushQS.On("One", &models.PushDispositivo{}).Return(mock.AnythingOfType("*errors.errorString"))

	controller.GetById()

	// Debería retornar 404
	if recorder.Code == 200 {
		t.Error("Expected non-200 for not found")
	}
}

func TestPushPut_UpdatePlataforma(t *testing.T) {
	ctx := context.NewContext()
	recorder := httptest.NewRecorder()

	body := map[string]interface{}{
		"plataforma": "ios",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("PUT", "/push?id=1", bytes.NewReader(bodyBytes))
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = bodyBytes
	ctx.Input.SetParam("id", "1")

	controller := &PushController{}
	controller.Init(ctx, "PushController", "Put", nil)

	mockPushOrm.ExpectedCalls = nil
	mockPushQS.ExpectedCalls = nil

	existingDispositivo := &models.PushDispositivo{
		PK_ID_DISPOSITIVO: 1,
		Plataforma:        models.PlataformaAndroid,
	}

	mockPushOrm.On("QueryTable", "push_dispositivo").Return(mockPushQS)
	mockPushQS.On("Filter", "pk_id_dispositivo", int64(1)).Return(mockPushQS)
	mockPushQS.On("One", &models.PushDispositivo{}).Return(nil).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.PushDispositivo)
		*arg = *existingDispositivo
	})

	mockPushOrm.On("Update", mock.AnythingOfType("*models.PushDispositivo"), []string{"Plataforma"}).Return(int64(1), nil)

	controller.Put()

	if recorder.Code != 200 {
		t.Errorf("Expected status 200, got %d", recorder.Code)
	}
}

func TestPushDelete_Success(t *testing.T) {
	ctx := context.NewContext()
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/push?id=1", nil)
	ctx.Reset(recorder, req)
	ctx.Input.SetParam("id", "1")

	controller := &PushController{}
	controller.Init(ctx, "PushController", "Delete", nil)

	mockPushOrm.ExpectedCalls = nil
	mockPushQS.ExpectedCalls = nil

	mockPushOrm.On("QueryTable", "push_dispositivo").Return(mockPushQS)
	mockPushQS.On("Filter", "pk_id_dispositivo", int64(1)).Return(mockPushQS)
	mockPushQS.On("Delete").Return(int64(1), nil)

	controller.Delete()

	if recorder.Code != 200 {
		t.Errorf("Expected status 200, got %d", recorder.Code)
	}
}

func TestPushListarEnvios_WithFiltros(t *testing.T) {
	ctx := context.NewContext()
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/push/envios?proveedor=WEB_PUSH&estado=enviado", nil)
	ctx.Reset(recorder, req)
	ctx.Input.SetParam("proveedor", "WEB_PUSH")
	ctx.Input.SetParam("estado", "enviado")

	controller := &PushController{}
	controller.Init(ctx, "PushController", "ListarEnvios", nil)

	mockPushOrm.ExpectedCalls = nil
	mockPushQS.ExpectedCalls = nil

	mockPushOrm.On("QueryTable", "push_envio").Return(mockPushQS)
	mockPushQS.On("Filter", "proveedor", "WEB_PUSH").Return(mockPushQS)
	mockPushQS.On("Filter", "estado", "enviado").Return(mockPushQS)
	mockPushQS.On("Count").Return(int64(5), nil)
	mockPushQS.On("OrderBy", []string{"-sent_at"}).Return(mockPushQS)
	mockPushQS.On("Limit", 10).Return(mockPushQS)
	mockPushQS.On("Offset", int64(0)).Return(mockPushQS)

	envios := []*models.PushEnvio{
		{
			PK_ID_ENVIO: 1,
			Proveedor:   models.ProveedorWebPush,
			Estado:      models.EstadoEnviadoPush,
			SentAt:      time.Now(),
		},
	}
	mockPushQS.On("All", mock.AnythingOfType("*[]*models.PushEnvio"), []string(nil)).Return(int64(1), nil).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*[]*models.PushEnvio)
		*arg = envios
	})

	controller.ListarEnvios()

	if recorder.Code != 200 {
		t.Errorf("Expected status 200, got %d", recorder.Code)
	}
}

func TestPushPost_InvalidPlataforma(t *testing.T) {
	ctx := context.NewContext()
	recorder := httptest.NewRecorder()

	body := map[string]interface{}{
		"token":      "test_token_123",
		"plataforma": "INVALID_PLATFORM",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/push", bytes.NewReader(bodyBytes))
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = bodyBytes

	controller := &PushController{}
	controller.Init(ctx, "PushController", "Post", nil)

	controller.Post()

	// Debería retornar error de validación
	if recorder.Code == 201 {
		t.Error("Expected error for invalid plataforma")
	}
}

func TestPushGetAll_EmptyDatabase(t *testing.T) {
	ctx := context.NewContext()
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/push", nil)
	ctx.Reset(recorder, req)

	controller := &PushController{}
	controller.Init(ctx, "PushController", "GetAll", nil)

	mockPushOrm.ExpectedCalls = nil
	mockPushQS.ExpectedCalls = nil

	mockPushOrm.On("QueryTable", "push_dispositivo").Return(mockPushQS)
	mockPushQS.On("Count").Return(int64(0), nil)
	mockPushQS.On("OrderBy", []string{"-updated_at"}).Return(mockPushQS)
	mockPushQS.On("Limit", 10).Return(mockPushQS)
	mockPushQS.On("Offset", int64(0)).Return(mockPushQS)
	mockPushQS.On("All", mock.AnythingOfType("*[]*models.PushDispositivo"), []string(nil)).Return(int64(0), nil)

	controller.GetAll()

	if recorder.Code != 200 {
		t.Errorf("Expected status 200, got %d", recorder.Code)
	}
}
