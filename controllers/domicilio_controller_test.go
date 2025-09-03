package controllers

import (
	stdctx "context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/beego/beego/v2/server/web/context"
	"restaurante/models"
)

type zeroResult struct{}

func (zeroResult) LastInsertId() (int64, error) { return 0, nil }
func (zeroResult) RowsAffected() (int64, error) { return 0, nil }

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

func TestIsValidEstadoDomicilio(t *testing.T) {
	if !isValidEstadoDomicilio("pendiente") {
		t.Fatalf("expected pendiente to be valid")
	}
	if isValidEstadoDomicilio("otro") {
		t.Fatalf("expected other value to be invalid")
	}
}

func TestDomicilioGetAllDBError(t *testing.T) {
	MockQuery = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
		return nil, errors.New("db error")
	}
	defer func() { MockQuery = nil }()

	r := httptest.NewRequest(http.MethodGet, "/domicilios", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := DomicilioController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetAll()

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestDomicilioGetAllNoResults(t *testing.T) {
	MockQuery = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
		cols := []string{"pk_id_domicilio", "direccion", "telefono", "estado_domicilio", "entregado", "fecha", "observaciones", "created_at", "updated_at", "created_by", "updated_by", "pk_documento_trabajador"}
		return &mockRows{columns: cols, values: [][]driver.Value{}}, nil
	}
	defer func() { MockQuery = nil }()

	r := httptest.NewRequest(http.MethodGet, "/domicilios", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := DomicilioController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetAll()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	var resp models.ApiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Code != http.StatusNotFound {
		t.Fatalf("expected response code 404, got %d", resp.Code)
	}
}

func TestDomicilioGetAllSuccess(t *testing.T) {
	MockQuery = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
		cols := []string{"pk_id_domicilio", "direccion", "telefono", "estado_domicilio", "entregado", "fecha", "observaciones", "created_at", "updated_at", "created_by", "updated_by", "pk_documento_trabajador"}
		vals := [][]driver.Value{{int64(1), "Calle 1", "123", "PENDIENTE", false, time.Now(), nil, time.Now(), time.Now(), nil, nil, nil}}
		return &mockRows{columns: cols, values: vals}, nil
	}
	defer func() { MockQuery = nil }()

	r := httptest.NewRequest(http.MethodGet, "/domicilios", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := DomicilioController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetAll()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	var resp models.ApiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Code != http.StatusOK {
		t.Fatalf("expected response code 200, got %d", resp.Code)
	}
	if resp.Data == nil {
		t.Fatalf("expected data in response")
	}
}

func TestDomicilioGetByIdInvalidID(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/domicilios/search", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := DomicilioController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetById()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestDomicilioGetByIdNotFound(t *testing.T) {
	MockQuery = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
		cols := []string{"pk_id_domicilio", "direccion"}
		return &mockRows{columns: cols, values: [][]driver.Value{}}, nil
	}
	defer func() { MockQuery = nil }()

	r := httptest.NewRequest(http.MethodGet, "/domicilios/search?id=1", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := DomicilioController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetById()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	var resp models.ApiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Code != http.StatusNotFound {
		t.Fatalf("expected response code 404, got %d", resp.Code)
	}
}

func TestDomicilioGetByIdSuccess(t *testing.T) {
	call := 0
	MockQuery = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
		call++
		switch call {
		case 1:
			cols := []string{"pk_id_domicilio", "direccion", "telefono", "estado_domicilio", "entregado", "fecha", "observaciones", "created_at", "updated_at", "created_by", "updated_by", "pk_documento_trabajador"}
			vals := [][]driver.Value{{int64(1), "Calle 1", "123", "PENDIENTE", false, time.Now(), nil, time.Now(), time.Now(), nil, nil, nil}}
			return &mockRows{columns: cols, values: vals}, nil
		case 2:
			cols := []string{"documento", "nombre", "apellido"}
			vals := [][]driver.Value{{int64(10), "John", "Doe"}}
			return &mockRows{columns: cols, values: vals}, nil
		case 3:
			cols := []string{"pedido_id", "pago_id", "pago_monto", "subtotal_productos", "productos"}
			vals := [][]driver.Value{{int64(5), int64(2), float64(20), float64(15), "[]"}}
			return &mockRows{columns: cols, values: vals}, nil
		default:
			return nil, errors.New("unexpected query")
		}
	}
	defer func() { MockQuery = nil }()

	r := httptest.NewRequest(http.MethodGet, "/domicilios/search?id=1", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := DomicilioController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetById()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	var resp models.ApiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map data, got %T", resp.Data)
	}
	if data["domicilio"] == nil || data["cliente"] == nil || data["pedido"] == nil {
		t.Fatalf("expected domicilio, cliente and pedido data")
	}
}

func TestDomicilioPutInvalidID(t *testing.T) {
	r := httptest.NewRequest(http.MethodPut, "/domicilios", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := DomicilioController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Put()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestDomicilioPutNotFound(t *testing.T) {
	MockQuery = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
		cols := []string{"pk_id_domicilio"}
		return &mockRows{columns: cols, values: [][]driver.Value{}}, nil
	}
	defer func() { MockQuery = nil }()

	r := httptest.NewRequest(http.MethodPut, "/domicilios?id=1", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := DomicilioController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Put()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	var resp models.ApiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Code != http.StatusNotFound {
		t.Fatalf("expected response code 404, got %d", resp.Code)
	}
}

func TestDomicilioPutInvalidJSON(t *testing.T) {
	MockQuery = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
		cols := []string{"pk_id_domicilio", "direccion"}
		vals := [][]driver.Value{{int64(1), "Dir"}}
		return &mockRows{columns: cols, values: vals}, nil
	}
	defer func() { MockQuery = nil }()

	r := httptest.NewRequest(http.MethodPut, "/domicilios?id=1", strings.NewReader("{invalid"))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte("{invalid")
	c := DomicilioController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Put()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestDomicilioPutSuccess(t *testing.T) {
	call := 0
	MockQuery = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
		call++
		if call == 1 {
			cols := []string{"pk_id_domicilio", "direccion", "telefono", "estado_domicilio", "entregado", "fecha", "observaciones", "created_at", "updated_at", "created_by", "updated_by", "pk_documento_trabajador"}
			vals := [][]driver.Value{{int64(1), "Calle 1", "123", "PENDIENTE", false, time.Now(), nil, time.Now(), time.Now(), nil, nil, nil}}
			return &mockRows{columns: cols, values: vals}, nil
		}
		cols := []string{"entregado", "created_at", "updated_at"}
		vals := [][]driver.Value{{false, time.Now(), time.Now()}}
		return &mockRows{columns: cols, values: vals}, nil
	}
	MockExec = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Result, error) {
		return mockResult{}, nil
	}
	defer func() { MockQuery = nil; MockExec = nil }()

	body := `{"direccion":"Nueva","telefono":"456"}`
	r := httptest.NewRequest(http.MethodPut, "/domicilios?id=1", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := DomicilioController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Put()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Domicilio actualizado correctamente") {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
}

func TestDomicilioDeleteInvalidID(t *testing.T) {
	r := httptest.NewRequest(http.MethodDelete, "/domicilios", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := DomicilioController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Delete()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestDomicilioDeleteSuccess(t *testing.T) {
	MockExec = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Result, error) {
		return mockResult{}, nil
	}
	MockQuery = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
		return &mockRows{columns: []string{}, values: [][]driver.Value{}}, nil
	}
	defer func() { MockExec = nil; MockQuery = nil }()

	r := httptest.NewRequest(http.MethodDelete, "/domicilios?id=1", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := DomicilioController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Delete()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Domicilio eliminado") {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
}

func TestDomicilioDeleteError(t *testing.T) {
	MockExec = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Result, error) {
		return nil, errors.New("db")
	}
	MockQuery = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
		return &mockRows{columns: []string{}, values: [][]driver.Value{}}, nil
	}
	defer func() { MockExec = nil; MockQuery = nil }()

	r := httptest.NewRequest(http.MethodDelete, "/domicilios?id=1", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := DomicilioController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Delete()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Domicilio no encontrado") {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
}

func TestAsignarDomiciliarioConflict(t *testing.T) {
	MockExec = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Result, error) {
		return zeroResult{}, nil
	}
	MockQuery = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
		cols := []string{"count"}
		vals := [][]driver.Value{{1}}
		return &mockRows{columns: cols, values: vals}, nil
	}
	defer func() { MockExec = nil; MockQuery = nil }()

	r := httptest.NewRequest(http.MethodPost, "/domicilios/asignar?domicilio_id=1&trabajador_id=2", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := DomicilioController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.AsignarDomiciliario()

	if w.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d", w.Code)
	}
}
