package controllers

import (
	stdctx "context"
	"database/sql/driver"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	beegoCtx "github.com/beego/beego/v2/server/web/context"
)

func TestPedidoGetAllWithoutDB(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/pedidos", nil)
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	c := PedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetAll()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al obtener los pedidos") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestPedidoPostDatabaseError(t *testing.T) {
	body := "{}"
	r := httptest.NewRequest(http.MethodPost, "/pedidos", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	c := PedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Ctx.Input.RequestBody = []byte(body)

	c.Post()

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al crear el pedido") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestPedidoAssignDomicilioNotFound(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/pedidos/asignar-domicilio?pedido_id=1&domicilio_id=1", nil)
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	c := PedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.AssignDomicilio()

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Pedido no encontrado") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestPedidoAssignPagoNotFound(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/pedidos/asignar-pago?pedido_id=1&pago_id=1", nil)
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	c := PedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.AssignPago()

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Pedido no encontrado") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestPedidoUpdateEstadoPedidoNotFound(t *testing.T) {
	r := httptest.NewRequest(http.MethodPut, "/pedidos/actualizar-estado?pedido_id=1&estado=TERMINADO", nil)
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	c := PedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.UpdateEstadoPedido()

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Pedido no encontrado") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestPedidoUpdateEstadoPedidoInvalidEstado(t *testing.T) {
	r := httptest.NewRequest(http.MethodPut, "/pedidos/actualizar-estado?pedido_id=1&estado=foo", nil)
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	c := PedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.UpdateEstadoPedido()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Estado inválido") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestPedidoGetPedidoDetailsMissingID(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/pedidos/detalles", nil)
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	c := PedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetPedidoDetails()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "El parámetro 'pedido_id' es obligatorio") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestPedidoGetPedidoDetailsDBError(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/pedidos/detalles?pedido_id=1", nil)
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	c := PedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetPedidoDetails()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al obtener los detalles del pedido") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestPedidoPostParseError(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/pedidos", strings.NewReader("a=%"))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	c := PedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Datos inválidos") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestPedidoPostDeliveryWithoutDomicilio(t *testing.T) {
	body := `{"delivery":true}`
	r := httptest.NewRequest(http.MethodPost, "/pedidos", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	c := PedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Ctx.Input.RequestBody = []byte(body)

	c.Post()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "domicilio es obligatorio") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestPedidoAssignDomicilioBadParam(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/pedidos/asignar-domicilio?pedido_id=1&domicilio_id=0", nil)
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	c := PedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.AssignDomicilio()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "debe ser un entero positivo") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestPedidoPostSuccess(t *testing.T) {
	MockExec = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Result, error) {
		return mockResult{}, nil
	}
	MockQuery = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
		cols := []string{"pk_id_pedido"}
		vals := [][]driver.Value{{int64(1)}}
		return &mockRows{columns: cols, values: vals}, nil
	}
	defer func() { MockExec = nil; MockQuery = nil }()

	body := "{}"
	r := httptest.NewRequest(http.MethodPost, "/pedidos", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	c := PedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Ctx.Input.RequestBody = []byte(body)

	c.Post()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Pedido creado exitosamente") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestPedidoPostDeliveryWithDomicilioAndRestaurante(t *testing.T) {
	MockExec = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Result, error) {
		return mockResult{}, nil
	}
	MockQuery = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
		cols := []string{"pk_id_pedido"}
		vals := [][]driver.Value{{int64(1)}}
		return &mockRows{columns: cols, values: vals}, nil
	}
	defer func() { MockExec = nil; MockQuery = nil }()

	body := `{"delivery":true,"pk_id_domicilio":1,"restauranteId":2}`
	r := httptest.NewRequest(http.MethodPost, "/pedidos", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	c := PedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Ctx.Input.RequestBody = []byte(body)

	c.Post()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Pedido creado exitosamente") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestPedidoGetAllWithFiltersNoResults(t *testing.T) {
	MockQuery = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
		cols := []string{"pk_id_pedido", "fecha", "hora", "delivery", "estado_pedido", "pk_id_domicilio", "pk_id_pago", "pk_id_restaurante", "updated_at", "updated_by"}
		return &mockRows{columns: cols, values: [][]driver.Value{}}, nil
	}
	defer func() { MockQuery = nil }()

	url := "/pedidos?fecha=2024-01-01&desde=2024-01-01&hasta=2024-02-01&mes=1&anio=2024&cliente=1&metodo_pago=NEQUI&domicilio=true"
	r := httptest.NewRequest(http.MethodGet, url, nil)
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	c := PedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetAll()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "No se encontraron pedidos") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestPedidoGetAllAnioOnly(t *testing.T) {
	MockQuery = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
		cols := []string{"pk_id_pedido", "fecha", "hora", "delivery", "estado_pedido", "pk_id_domicilio", "pk_id_pago", "pk_id_restaurante", "updated_at", "updated_by"}
		return &mockRows{columns: cols, values: [][]driver.Value{}}, nil
	}
	defer func() { MockQuery = nil }()

	r := httptest.NewRequest(http.MethodGet, "/pedidos?anio=2024", nil)
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	c := PedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetAll()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "No se encontraron pedidos") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestPedidoGetAllWithResults(t *testing.T) {
	MockQuery = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
		cols := []string{"pk_id_pedido", "fecha", "hora", "delivery", "estado_pedido", "pk_id_domicilio", "pk_id_pago", "pk_id_restaurante", "updated_at", "updated_by"}
		now := time.Now()
		vals := [][]driver.Value{{int64(1), now, now, false, "INICIADO", nil, nil, nil, now, "tester"}}
		return &mockRows{columns: cols, values: vals}, nil
	}
	defer func() { MockQuery = nil }()

	r := httptest.NewRequest(http.MethodGet, "/pedidos", nil)
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	c := PedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetAll()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Pedidos obtenidos exitosamente") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestPedidoGetAllDomicilioFalse(t *testing.T) {
	MockQuery = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
		cols := []string{"pk_id_pedido", "fecha", "hora", "delivery", "estado_pedido", "pk_id_domicilio", "pk_id_pago", "pk_id_restaurante", "updated_at", "updated_by"}
		return &mockRows{columns: cols, values: [][]driver.Value{}}, nil
	}
	defer func() { MockQuery = nil }()

	r := httptest.NewRequest(http.MethodGet, "/pedidos?domicilio=false", nil)
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	c := PedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetAll()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "No se encontraron pedidos") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestPedidoAssignDomicilioUpdateError(t *testing.T) {
	MockQuery = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
		cols := []string{"pk_id_pedido", "fecha", "hora", "delivery", "estado_pedido", "pk_id_domicilio", "pk_id_pago", "pk_id_restaurante", "updated_at", "updated_by"}
		now := time.Now()
		vals := [][]driver.Value{{int64(1), now, now, false, "INICIADO", nil, nil, nil, now, "tester"}}
		return &mockRows{columns: cols, values: vals}, nil
	}
	MockExec = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Result, error) {
		return nil, errors.New("update error")
	}
	defer func() { MockQuery = nil; MockExec = nil }()

	r := httptest.NewRequest(http.MethodPost, "/pedidos/asignar-domicilio?pedido_id=1&domicilio_id=1", nil)
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	c := PedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.AssignDomicilio()

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al asignar domicilio") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestPedidoAssignDomicilioSuccess(t *testing.T) {
	MockQuery = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
		cols := []string{"pk_id_pedido", "fecha", "hora", "delivery", "estado_pedido", "pk_id_domicilio", "pk_id_pago", "pk_id_restaurante", "updated_at", "updated_by"}
		now := time.Now()
		vals := [][]driver.Value{{int64(1), now, now, false, "INICIADO", nil, nil, nil, now, "tester"}}
		return &mockRows{columns: cols, values: vals}, nil
	}
	MockExec = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Result, error) {
		return mockResult{}, nil
	}
	defer func() { MockQuery = nil; MockExec = nil }()

	r := httptest.NewRequest(http.MethodPost, "/pedidos/asignar-domicilio?pedido_id=1&domicilio_id=1", nil)
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	c := PedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.AssignDomicilio()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Domicilio asignado correctamente") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestPedidoAssignPagoUpdateError(t *testing.T) {
	MockQuery = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
		cols := []string{"pk_id_pedido", "fecha", "hora", "delivery", "estado_pedido", "pk_id_domicilio", "pk_id_pago", "pk_id_restaurante", "updated_at", "updated_by"}
		now := time.Now()
		vals := [][]driver.Value{{int64(1), now, now, false, "INICIADO", nil, nil, nil, now, "tester"}}
		return &mockRows{columns: cols, values: vals}, nil
	}
	MockExec = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Result, error) {
		return nil, errors.New("update error")
	}
	defer func() { MockQuery = nil; MockExec = nil }()

	r := httptest.NewRequest(http.MethodPost, "/pedidos/asignar-pago?pedido_id=1&pago_id=1", nil)
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	c := PedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.AssignPago()

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al asignar pago") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestPedidoAssignPagoSuccess(t *testing.T) {
	qCount := 0
	MockQuery = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
		if qCount == 0 {
			qCount++
			cols := []string{"pk_id_pedido", "fecha", "hora", "delivery", "estado_pedido", "pk_id_domicilio", "pk_id_pago", "pk_id_restaurante", "updated_at", "updated_by"}
			now := time.Now()
			vals := [][]driver.Value{{int64(1), now, now, false, "INICIADO", nil, nil, nil, now, "tester"}}
			return &mockRows{columns: cols, values: vals}, nil
		}
		cols := []string{"pk_id_pago", "fecha", "hora", "monto", "estado_pago", "pk_id_metodo_pago", "updated_at", "updated_by"}
		now := time.Now()
		vals := [][]driver.Value{{int64(1), now, now, int64(100), "PENDIENTE", int64(1), now, "tester"}}
		return &mockRows{columns: cols, values: vals}, nil
	}
	MockExec = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Result, error) {
		return mockResult{}, nil
	}
	defer func() { MockQuery = nil; MockExec = nil }()

	r := httptest.NewRequest(http.MethodPost, "/pedidos/asignar-pago?pedido_id=1&pago_id=1", nil)
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	c := PedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.AssignPago()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Pago asignado correctamente") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestPedidoUpdateEstadoPedidoUpdateError(t *testing.T) {
	MockQuery = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
		cols := []string{"pk_id_pedido", "fecha", "hora", "delivery", "estado_pedido", "pk_id_domicilio", "pk_id_pago", "pk_id_restaurante", "updated_at", "updated_by"}
		now := time.Now()
		vals := [][]driver.Value{{int64(1), now, now, false, "INICIADO", nil, nil, nil, now, "tester"}}
		return &mockRows{columns: cols, values: vals}, nil
	}
	MockExec = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Result, error) {
		return nil, errors.New("update error")
	}
	defer func() { MockQuery = nil; MockExec = nil }()

	r := httptest.NewRequest(http.MethodPut, "/pedidos/actualizar-estado?pedido_id=1&estado=TERMINADO", nil)
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	c := PedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.UpdateEstadoPedido()

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al actualizar estado del pedido") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestPedidoUpdateEstadoPedidoSuccess(t *testing.T) {
	MockQuery = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
		cols := []string{"pk_id_pedido", "fecha", "hora", "delivery", "estado_pedido", "pk_id_domicilio", "pk_id_pago", "pk_id_restaurante", "updated_at", "updated_by"}
		now := time.Now()
		vals := [][]driver.Value{{int64(1), now, now, false, "INICIADO", nil, nil, nil, now, "tester"}}
		return &mockRows{columns: cols, values: vals}, nil
	}
	MockExec = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Result, error) {
		return mockResult{}, nil
	}
	defer func() { MockQuery = nil; MockExec = nil }()

	r := httptest.NewRequest(http.MethodPut, "/pedidos/actualizar-estado?pedido_id=1&estado=TERMINADO", nil)
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	c := PedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.UpdateEstadoPedido()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Estado del pedido actualizado correctamente") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestPedidoUpdateEstadoPedidoSuccess_EnPreparacion(t *testing.T) {
	MockQuery = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
		cols := []string{"pk_id_pedido", "fecha", "hora", "delivery", "estado_pedido", "pk_id_domicilio", "pk_id_pago", "pk_id_restaurante", "updated_at", "updated_by"}
		now := time.Now()
		vals := [][]driver.Value{{int64(1), now, now, false, "INICIADO", nil, nil, nil, now, "tester"}}
		return &mockRows{columns: cols, values: vals}, nil
	}
	MockExec = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Result, error) {
		return mockResult{}, nil
	}
	defer func() { MockQuery = nil; MockExec = nil }()

	r := httptest.NewRequest(http.MethodPut, "/pedidos/actualizar-estado?pedido_id=1&estado=EN_PREPARACION", nil)
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	c := PedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.UpdateEstadoPedido()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestPedidoUpdateEstadoPedidoSuccess_Listo(t *testing.T) {
	MockQuery = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
		cols := []string{"pk_id_pedido", "fecha", "hora", "delivery", "estado_pedido", "pk_id_domicilio", "pk_id_pago", "pk_id_restaurante", "updated_at", "updated_by"}
		now := time.Now()
		vals := [][]driver.Value{{int64(1), now, now, false, "INICIADO", nil, nil, nil, now, "tester"}}
		return &mockRows{columns: cols, values: vals}, nil
	}
	MockExec = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Result, error) {
		return mockResult{}, nil
	}
	defer func() { MockQuery = nil; MockExec = nil }()

	r := httptest.NewRequest(http.MethodPut, "/pedidos/actualizar-estado?pedido_id=1&estado=LISTO", nil)
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	c := PedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.UpdateEstadoPedido()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestPedidoGetPedidoDetailsSuccess(t *testing.T) {
	MockQuery = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
		cols := []string{"pk_id_pedido", "fecha", "hora", "delivery", "estado_pedido", "metodo_pago", "productos", "pago_id", "metodo_pago_id", "domicilio_id", "pk_documento_cliente"}
		vals := [][]driver.Value{{int64(1), "2024-01-01", "12:00:00", false, "TERMINADO", "NEQUI", "[]", int64(2), int64(3), int64(4), int64(5)}}
		return &mockRows{columns: cols, values: vals}, nil
	}
	defer func() { MockQuery = nil }()

	r := httptest.NewRequest(http.MethodGet, "/pedidos/detalles?pedido_id=1", nil)
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	c := PedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetPedidoDetails()
	body := w.Body.String()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	for _, field := range []string{
		"Detalles del pedido obtenidos exitosamente",
		"\"fechaPedido\":\"2024-01-01\"",
		"\"horaPedido\":\"12:00:00\"",
		"\"delivery\":false",
		"\"estadoPedido\":\"TERMINADO\"",
		"\"pagoId\":2",
		"\"metodoPagoId\":3",
		"\"domicilioId\":4",
		"\"documentoCliente\":5",
	} {
		if !strings.Contains(body, field) {
			t.Errorf("unexpected body: %s", body)
		}
	}
}
