package controllers

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
)

// TestHelper proporciona utilidades para tests de controladores
type TestHelper struct {
	t *testing.T
}

func NewTestHelper(t *testing.T) *TestHelper {
	return &TestHelper{t: t}
}

// CreateTestContext crea un contexto de prueba para Beego
func (h *TestHelper) CreateTestContext(method, url string, body interface{}) (*context.Context, *httptest.ResponseRecorder) {
	var reqBody []byte
	var err error

	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			h.t.Fatalf("Failed to marshal request body: %v", err)
		}
	}

	req := httptest.NewRequest(method, url, bytes.NewBuffer(reqBody))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Simular token JWT válido
	req.Header.Set("Authorization", "Bearer valid_test_token")

	recorder := httptest.NewRecorder()

	// Usar context.NewContext() y Reset() para inicializar correctamente
	ctx := context.NewContext()
	ctx.Reset(recorder, req)

	return ctx, recorder
}

// ParseJSONResponse parsea la respuesta JSON
func (h *TestHelper) ParseJSONResponse(recorder *httptest.ResponseRecorder, target interface{}) {
	if recorder.Body.Len() == 0 {
		h.t.Fatal("Empty response body")
	}

	err := json.Unmarshal(recorder.Body.Bytes(), target)
	if err != nil {
		h.t.Fatalf("Failed to parse JSON response: %v\nBody: %s", err, recorder.Body.String())
	}
}

// AssertStatusCode verifica el código de estado HTTP
func (h *TestHelper) AssertStatusCode(recorder *httptest.ResponseRecorder, expected int) {
	if recorder.Code != expected {
		h.t.Errorf("Expected status code %d, got %d\nBody: %s", expected, recorder.Code, recorder.Body.String())
	}
}

// AssertJSONField verifica que un campo JSON tenga el valor esperado
func (h *TestHelper) AssertJSONField(recorder *httptest.ResponseRecorder, field string, expected interface{}) {
	var response map[string]interface{}
	h.ParseJSONResponse(recorder, &response)

	actual, exists := response[field]
	if !exists {
		h.t.Errorf("Field '%s' not found in response", field)
		return
	}

	if actual != expected {
		h.t.Errorf("Field '%s': expected %v, got %v", field, expected, actual)
	}
}

// MockController proporciona funcionalidad base para mocks de controladores
type MockController struct {
	web.Controller
	TestContext *context.Context
}

func (c *MockController) Init(ctx *context.Context, controllerName, actionName string, app interface{}) {
	c.Controller.Init(ctx, controllerName, actionName, app)
	c.TestContext = ctx
}

// SetPathParam simula un parámetro de ruta
func (c *MockController) SetPathParam(key, value string) {
	if c.TestContext != nil && c.TestContext.Input != nil {
		c.TestContext.Input.SetParam(key, value)
	}
}

// SetQueryParam simula un parámetro de query
func (c *MockController) SetQueryParam(key, value string) {
	if c.TestContext != nil && c.TestContext.Request != nil {
		q := c.TestContext.Request.URL.Query()
		q.Set(key, value)
		c.TestContext.Request.URL.RawQuery = q.Encode()
	}
}

// Helper functions para crear punteros
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func int64Ptr(i int64) *int64 {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}
