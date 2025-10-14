package telemetria

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"restaurante/models"

	"github.com/beego/beego/v2/server/web/context"
)

func TestProductosPopularesData_Structure(t *testing.T) {
	data := ProductosPopularesData{
		ProductosPopulares: []models.ProductoVendido{},
	}

	if data.ProductosPopulares == nil {
		t.Error("Expected non-nil ProductosPopulares")
	}
}

func TestValidateAdminRole_NoBearer(t *testing.T) {

	controller := &TelemetriaController{}

	req := httptest.NewRequest("GET", "/telemetria/dashboard", nil)
	req.Header.Set("Authorization", "some-token-without-bearer")
	w := httptest.NewRecorder()

	ctx := context.NewContext()
	ctx.Reset(w, req)

	controller.Controller.Ctx = ctx
	controller.Controller.Data = make(map[interface{}]interface{})

	_, valid := controller.validateAdminRole()

	if valid {
		t.Error("Expected validation to fail with invalid token")
	}

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestGetRentabilidad_Unauthorized(t *testing.T) {
	controller := &TelemetriaController{}

	req := httptest.NewRequest("GET", "/telemetria/rentabilidad", nil)
	w := httptest.NewRecorder()

	ctx := context.NewContext()
	ctx.Reset(w, req)

	controller.Controller.Ctx = ctx
	controller.Controller.Data = make(map[interface{}]interface{})

	controller.GetRentabilidad()

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestGetSegmentacion_Unauthorized(t *testing.T) {
	controller := &TelemetriaController{}

	req := httptest.NewRequest("GET", "/telemetria/segmentacion", nil)
	w := httptest.NewRecorder()

	ctx := context.NewContext()
	ctx.Reset(w, req)

	controller.Controller.Ctx = ctx
	controller.Controller.Data = make(map[interface{}]interface{})

	controller.GetSegmentacion()

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestGetEficiencia_Unauthorized(t *testing.T) {
	controller := &TelemetriaController{}

	req := httptest.NewRequest("GET", "/telemetria/eficiencia", nil)
	w := httptest.NewRecorder()

	ctx := context.NewContext()
	ctx.Reset(w, req)

	controller.Controller.Ctx = ctx
	controller.Controller.Data = make(map[interface{}]interface{})

	controller.GetEficiencia()

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestGetReservasAnalisis_Unauthorized(t *testing.T) {
	controller := &TelemetriaController{}

	req := httptest.NewRequest("GET", "/telemetria/reservas-analisis", nil)
	w := httptest.NewRecorder()

	ctx := context.NewContext()
	ctx.Reset(w, req)

	controller.Controller.Ctx = ctx
	controller.Controller.Data = make(map[interface{}]interface{})

	controller.GetReservasAnalisis()

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestGetPedidosAnalisis_Unauthorized(t *testing.T) {
	controller := &TelemetriaController{}

	req := httptest.NewRequest("GET", "/telemetria/pedidos-analisis", nil)
	w := httptest.NewRecorder()

	ctx := context.NewContext()
	ctx.Reset(w, req)

	controller.Controller.Ctx = ctx
	controller.Controller.Data = make(map[interface{}]interface{})

	controller.GetPedidosAnalisis()

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}
