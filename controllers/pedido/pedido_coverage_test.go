package pedido

import (
	stdctx "context"
	"database/sql/driver"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"restaurante/database"

	beegoCtx "github.com/beego/beego/v2/server/web/context"
)

func newPedidoCtrl(method, url, body string) (*PedidoController, *httptest.ResponseRecorder) {
	r := httptest.NewRequest(method, url, strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	c := &PedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Ctx.Input.RequestBody = []byte(body)
	return c, w
}

func pedidoRows() (driver.Rows, error) {
	cols := []string{"pk_id_pedido", "fecha", "hora", "delivery", "estado_pedido", "pk_id_domicilio", "pk_id_pago", "pk_id_restaurante", "updated_at", "updated_by"}
	now := time.Now()
	return &mockRows{columns: cols, values: [][]driver.Value{{int64(1), now, now, false, "INICIADO", nil, nil, nil, now, "tester"}}}, nil
}

func TestPedidoPostWithDocumentoClienteAndZoneFallback(t *testing.T) {
	origZone := database.BogotaZone
	origLoad := loadLocation
	database.BogotaZone = nil
	loadLocation = func(string) (*time.Location, error) { return nil, errors.New("sin tzdata") }
	t.Cleanup(func() { database.BogotaZone = origZone; loadLocation = origLoad })

	var insertArgs []driver.NamedValue
	MockExec = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Result, error) {
		return mockResult{}, nil
	}
	MockQuery = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
		insertArgs = args
		return &mockRows{columns: []string{"pk_id_pedido"}, values: [][]driver.Value{{int64(1)}}}, nil
	}
	t.Cleanup(func() { MockExec = nil; MockQuery = nil })

	c, w := newPedidoCtrl(http.MethodPost, "/pedidos", `{"documentoCliente":12345}`)
	c.Post()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "Pedido creado exitosamente") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
	found := false
	for _, a := range insertArgs {
		if v, ok := a.Value.(int64); ok && v == 12345 {
			found = true
		}
	}
	if !found {
		t.Errorf("el documento del cliente no se envio en el INSERT: %+v", insertArgs)
	}
}

func TestPedidoAssignPagoSinCambiarEstado(t *testing.T) {
	var execQueries []string
	MockQuery = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
		return pedidoRows()
	}
	MockExec = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Result, error) {
		execQueries = append(execQueries, query)
		return mockResult{}, nil
	}
	t.Cleanup(func() { MockQuery = nil; MockExec = nil })

	c, w := newPedidoCtrl(http.MethodPost, "/pedidos/asignar-pago?pedido_id=1&pago_id=2&cambiar_estado=false", "")
	c.AssignPago()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "Pago asignado correctamente") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "INICIADO") {
		t.Errorf("el estado actual debe conservarse: %s", w.Body.String())
	}
	for _, q := range execQueries {
		if strings.Contains(strings.ToLower(q), "update \"pago\"") || strings.Contains(strings.ToLower(q), "update pago") {
			t.Errorf("no debe actualizarse el pago cuando cambiar_estado=false: %s", q)
		}
	}
}

func TestPedidoAssignPagoGetEstadoError(t *testing.T) {
	qCount := 0
	MockQuery = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
		qCount++
		if qCount == 1 {
			return pedidoRows()
		}
		return nil, errors.New("estado error")
	}
	t.Cleanup(func() { MockQuery = nil })

	c, w := newPedidoCtrl(http.MethodPost, "/pedidos/asignar-pago?pedido_id=1&pago_id=2&cambiar_estado=false", "")
	c.AssignPago()

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al obtener estado del pedido") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestPedidoAssignPagoCambiarEstadoInvalidoUsaDefault(t *testing.T) {
	MockQuery = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
		return pedidoRows()
	}
	MockExec = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Result, error) {
		return mockResult{}, nil
	}
	t.Cleanup(func() { MockQuery = nil; MockExec = nil })

	c, w := newPedidoCtrl(http.MethodPost, "/pedidos/asignar-pago?pedido_id=1&pago_id=2&cambiar_estado=quizas", "")
	c.AssignPago()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "TERMINADO") {
		t.Errorf("con valor invalido se usa cambiar_estado=true (TERMINADO): %s", w.Body.String())
	}
}
