package domicilio

import (
	stdctx "context"
	"database/sql/driver"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	beegoCtx "github.com/beego/beego/v2/server/web/context"
)

func TestDomicilioGetById_InvalidProductosJSON(t *testing.T) {
	call := 0
	origQ := MockQuery
	MockQuery = func(_ stdctx.Context, _ string, _ []driver.NamedValue) (driver.Rows, error) {
		call++
		switch call {
		case 1:
			cols := []string{"pk_id_domicilio", "direccion", "telefono", "estado_domicilio", "entregado", "fecha", "observaciones", "created_at", "updated_at", "created_by", "updated_by", "pk_documento_trabajador"}
			vals := [][]driver.Value{{int64(1), "C", "1", "PENDIENTE", false, time.Now(), nil, time.Now(), time.Now(), nil, nil, nil}}
			return &mockRows{columns: cols, values: vals}, nil
		case 2:
			cols := []string{"documento", "nombre", "apellido"}
			vals := [][]driver.Value{{int64(10), "John", "Doe"}}
			return &mockRows{columns: cols, values: vals}, nil
		case 3:
			cols := []string{"pedido_id", "pago_id", "pago_monto", "subtotal_productos", "productos"}
			vals := [][]driver.Value{{int64(5), nil, float64(20), nil, "["}}
			return &mockRows{columns: cols, values: vals}, nil
		default:
			return &mockRows{columns: []string{"ok"}, values: [][]driver.Value{{int64(1)}}}, nil
		}
	}
	t.Cleanup(func() { MockQuery = origQ })

	r := httptest.NewRequest(http.MethodGet, "/domicilios/search?id=1", nil)
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	c := DomicilioController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetById()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}
