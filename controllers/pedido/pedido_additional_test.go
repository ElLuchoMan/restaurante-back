package pedido

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	beegoCtx "github.com/beego/beego/v2/server/web/context"
)

func TestPedidoPost_TimezoneError(t *testing.T) {
	orig := loadLocationPedido
	loadLocationPedido = func(name string) (*time.Location, error) { return nil, errors.New("tz") }
	t.Cleanup(func() { loadLocationPedido = orig })

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
	if !strings.Contains(w.Body.String(), "No se pudo cargar zona horaria") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}
