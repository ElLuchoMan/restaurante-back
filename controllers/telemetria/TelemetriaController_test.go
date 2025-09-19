package telemetria

import (
	"net/http"
	"net/http/httptest"
	"testing"

	loginc "restaurante/controllers/login"

	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
)

func TestTelemetriaController_GetDashboard_Unauthorized(t *testing.T) {
	// Configurar el controlador
	controller := &TelemetriaController{}

	// Crear request sin token de autorización
	req := httptest.NewRequest("GET", "/telemetria/dashboard", nil)
	w := httptest.NewRecorder()

	// Crear contexto de Beego
	ctx := context.NewContext()
	ctx.Reset(w, req)

	// Configurar el controlador con el contexto
	controller.Controller.Ctx = ctx
	controller.Controller.Data = make(map[interface{}]interface{})

	// Ejecutar el método
	controller.GetDashboard()

	// Verificar que retorna 401 Unauthorized
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestTelemetriaController_GetSales_Unauthorized(t *testing.T) {
	// Configurar el controlador
	controller := &TelemetriaController{}

	// Crear request sin token de autorización
	req := httptest.NewRequest("GET", "/telemetria/sales", nil)
	w := httptest.NewRecorder()

	// Crear contexto de Beego
	ctx := context.NewContext()
	ctx.Reset(w, req)

	// Configurar el controlador con el contexto
	controller.Controller.Ctx = ctx
	controller.Controller.Data = make(map[interface{}]interface{})

	// Ejecutar el método
	controller.GetSales()

	// Verificar que retorna 401 Unauthorized
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestTelemetriaController_GetProducts_Unauthorized(t *testing.T) {
	// Configurar el controlador
	controller := &TelemetriaController{}

	// Crear request sin token de autorización
	req := httptest.NewRequest("GET", "/telemetria/products", nil)
	w := httptest.NewRecorder()

	// Crear contexto de Beego
	ctx := context.NewContext()
	ctx.Reset(w, req)

	// Configurar el controlador con el contexto
	controller.Controller.Ctx = ctx
	controller.Controller.Data = make(map[interface{}]interface{})

	// Ejecutar el método
	controller.GetProducts()

	// Verificar que retorna 401 Unauthorized
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestTelemetriaController_GetUsers_Unauthorized(t *testing.T) {
	// Configurar el controlador
	controller := &TelemetriaController{}

	// Crear request sin token de autorización
	req := httptest.NewRequest("GET", "/telemetria/users", nil)
	w := httptest.NewRecorder()

	// Crear contexto de Beego
	ctx := context.NewContext()
	ctx.Reset(w, req)

	// Configurar el controlador con el contexto
	controller.Controller.Ctx = ctx
	controller.Controller.Data = make(map[interface{}]interface{})

	// Ejecutar el método
	controller.GetUsers()

	// Verificar que retorna 401 Unauthorized
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestTelemetriaController_GetTimeAnalysis_Unauthorized(t *testing.T) {
	// Configurar el controlador
	controller := &TelemetriaController{}

	// Crear request sin token de autorización
	req := httptest.NewRequest("GET", "/telemetria/time-analysis", nil)
	w := httptest.NewRecorder()

	// Crear contexto de Beego
	ctx := context.NewContext()
	ctx.Reset(w, req)

	// Configurar el controlador con el contexto
	controller.Controller.Ctx = ctx
	controller.Controller.Data = make(map[interface{}]interface{})

	// Ejecutar el método
	controller.GetTimeAnalysis()

	// Verificar que retorna 401 Unauthorized
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestGetJWTSecret(t *testing.T) {
	// Configurar modo de desarrollo para evitar panic
	web.BConfig.RunMode = "dev"

	// Ejecutar función del LoginController
	secret := loginc.GetJWTSecret()

	// Verificar que retorna un secreto válido
	if len(secret) == 0 {
		t.Error("Expected non-empty JWT secret")
	}
}

// TestValidateAdminRole_InvalidToken prueba la validación con token inválido
func TestValidateAdminRole_InvalidToken(t *testing.T) {
	// Configurar el controlador
	controller := &TelemetriaController{}

	// Crear request con token inválido
	req := httptest.NewRequest("GET", "/telemetria/dashboard", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()

	// Crear contexto de Beego
	ctx := context.NewContext()
	ctx.Reset(w, req)

	// Configurar el controlador con el contexto
	controller.Controller.Ctx = ctx
	controller.Controller.Data = make(map[interface{}]interface{})

	// Ejecutar validación
	_, valid := controller.validateAdminRole()

	// Verificar que la validación falla
	if valid {
		t.Error("Expected validation to fail with invalid token")
	}

	// Verificar que retorna 401 Unauthorized
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}
