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

	beegoCtx "github.com/beego/beego/v2/server/web/context"
)

func TestPedidoAssignPagoUpdatePagoError(t *testing.T) {
	MockQuery = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
		cols := []string{"pk_id_pedido", "fecha", "hora", "delivery", "estado_pedido", "pk_id_domicilio", "pk_id_pago", "pk_id_restaurante", "updated_at", "updated_by"}
		now := time.Now()
		vals := [][]driver.Value{{int64(1), now, now, false, "INICIADO", nil, nil, nil, now, "tester"}}
		return &mockRows{columns: cols, values: vals}, nil
	}
	execCount := 0
	MockExec = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Result, error) {
		execCount++
		if execCount == 1 {
			return mockResult{}, nil
		}
		return nil, errors.New("update pago error")
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

func TestPedidoGetAllInvalidDomicilio(t *testing.T) {
	MockQuery = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
		cols := []string{"pk_id_pedido", "fecha", "hora", "delivery", "estado_pedido", "pk_id_domicilio", "pk_id_pago", "pk_id_restaurante", "updated_at", "updated_by"}
		return &mockRows{columns: cols, values: [][]driver.Value{}}, nil
	}
	defer func() { MockQuery = nil }()

	r := httptest.NewRequest(http.MethodGet, "/pedidos?domicilio=maybe", nil)
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
