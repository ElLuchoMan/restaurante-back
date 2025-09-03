package controllers

import (
	stdctx "context"
	"database/sql/driver"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/beego/beego/v2/server/web/context"
	"restaurante/models"
)

func TestPostInvalidEstado(t *testing.T) {
	body := `{"direccion":"Calle 1","fechaDomicilio":"2024-01-01","telefono":"123","estado":"otro"}`
	r := httptest.NewRequest(http.MethodPost, "/domicilios", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := DomicilioController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
	var resp models.ApiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if !strings.Contains(resp.Message, "estado") {
		t.Fatalf("expected error message about estado, got %s", resp.Message)
	}
}

func TestPostReturnsGeneratedEntregado(t *testing.T) {
	MockExec = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Result, error) {
		return mockResult{}, nil
	}
	MockQuery = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
		if strings.Contains(strings.ToUpper(query), "SELECT ENTREGADO") {
			cols := []string{"entregado", "created_at", "updated_at"}
			vals := [][]driver.Value{{false, time.Now(), time.Now()}}
			return &mockRows{columns: cols, values: vals}, nil
		}
		cols := []string{"pk_id_domicilio"}
		vals := [][]driver.Value{{int64(1)}}
		return &mockRows{columns: cols, values: vals}, nil
	}
	defer func() { MockExec = nil; MockQuery = nil }()

	body := `{"direccion":"Calle 1","fechaDomicilio":"2024-01-01","telefono":"123"}`
	r := httptest.NewRequest(http.MethodPost, "/domicilios", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := DomicilioController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", w.Code)
	}
	var resp models.ApiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %T", resp.Data)
	}
	if entregado, ok := data["entregado"].(bool); !ok || entregado {
		t.Fatalf("expected entregado false, got %v", data["entregado"])
	}
	if _, ok := data["createdAt"]; !ok {
		t.Fatalf("expected createdAt in response")
	}
	if _, ok := data["updatedAt"]; !ok {
		t.Fatalf("expected updatedAt in response")
	}
}

func TestAsignarDomiciliarioSuccess(t *testing.T) {
	MockExec = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Result, error) {
		if !strings.Contains(strings.ToUpper(query), "ESTADO_DOMICILIO='EN_CAMINO'") {
			t.Errorf("unexpected query: %s", query)
		}
		return mockResult{}, nil
	}
	defer func() { MockExec = nil }()

	r := httptest.NewRequest(http.MethodPost, "/domicilios/asignar?domicilio_id=1&trabajador_id=2", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := DomicilioController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.AsignarDomiciliario()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	var resp models.ApiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map data, got %T", resp.Data)
	}
	if estado, ok := data["estadoDomicilio"].(string); !ok || estado != string(models.EstadoDomicilioEnCamino) {
		t.Fatalf("expected estadoDomicilio %s, got %v", models.EstadoDomicilioEnCamino, data["estadoDomicilio"])
	}
}
