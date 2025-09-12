package domicilio

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

func TestDomicilioPostTrabajadorVariations(t *testing.T) {
	cases := []struct{ name, body string }{
		{"Nil", `{"direccion":"c","telefono":"1","fechaDomicilio":"2024-01-01","trabajadorAsignado":null}`},
		{"Zero", `{"direccion":"c","telefono":"1","fechaDomicilio":"2024-01-01","trabajadorAsignado":0}`},
		{"Positive", `{"direccion":"c","telefono":"1","fechaDomicilio":"2024-01-01","trabajadorAsignado":5}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			origQ := MockQuery
			MockQuery = func(_ stdctx.Context, q string, _ []driver.NamedValue) (driver.Rows, error) {
				lower := strings.ToLower(q)
				if strings.Contains(lower, "insert into domicilio") {
					cols := []string{"pk_id_domicilio"}
					vals := [][]driver.Value{{int64(1)}}
					return &mockRows{columns: cols, values: vals}, nil
				}
				if strings.Contains(lower, "select entregado, created_at, updated_at") {
					cols := []string{"entregado", "created_at", "updated_at"}
					vals := [][]driver.Value{{false, time.Now(), time.Now()}}
					return &mockRows{columns: cols, values: vals}, nil
				}
				return &mockRows{columns: []string{"ok"}, values: [][]driver.Value{{int64(1)}}}, nil
			}
			t.Cleanup(func() { MockQuery = origQ })

			r := httptest.NewRequest(http.MethodPost, "/domicilios", strings.NewReader(tc.body))
			w := httptest.NewRecorder()
			ctx := beegoCtx.NewContext()
			ctx.Reset(w, r)
			ctx.Input.RequestBody = []byte(tc.body)
			c := DomicilioController{}
			c.Ctx = ctx
			c.Data = make(map[interface{}]interface{})
			c.Post()
			if w.Code != http.StatusCreated {
				t.Fatalf("unexpected status %d", w.Code)
			}
		})
	}
}

func TestDomicilioPostUnmarshalError(t *testing.T) {
	body := `{"direccion":"c","telefono":123,"fechaDomicilio":"2024-01-01"}`
	r := httptest.NewRequest(http.MethodPost, "/domicilios", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := DomicilioController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Post()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestDomicilioPostMarshalError(t *testing.T) {
	origMarshal := jsonMarshal
	jsonMarshal = func(interface{}) ([]byte, error) { return nil, errors.New("marshal") }
	defer func() { jsonMarshal = origMarshal }()

	body := `{"direccion":"c","telefono":"1","fechaDomicilio":"2024-01-01"}`
	r := httptest.NewRequest(http.MethodPost, "/domicilios", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := DomicilioController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Post()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestDomicilioPutWithUpdatedBy(t *testing.T) {
	call := 0
	origQ := MockQuery
	MockQuery = func(_ stdctx.Context, q string, _ []driver.NamedValue) (driver.Rows, error) {
		call++
		switch call {
		case 1:
			cols := []string{"pk_id_domicilio", "direccion", "telefono", "estado_domicilio", "entregado", "fecha", "observaciones", "created_at", "updated_at", "created_by", "updated_by", "pk_documento_trabajador"}
			vals := [][]driver.Value{{int64(1), "Calle", "123", "PENDIENTE", false, time.Now(), nil, time.Now(), time.Now(), nil, nil, nil}}
			return &mockRows{columns: cols, values: vals}, nil
		case 2:
			cols := []string{"entregado", "created_at", "updated_at"}
			vals := [][]driver.Value{{false, time.Now(), time.Now()}}
			return &mockRows{columns: cols, values: vals}, nil
		default:
			return nil, errors.New("unexpected query")
		}
	}
	origE := MockExec
	MockExec = func(_ stdctx.Context, _ string, _ []driver.NamedValue) (driver.Result, error) {
		return mockResult{}, nil
	}
	t.Cleanup(func() { MockQuery = origQ; MockExec = origE })

	body := `{"direccion":"nueva","telefono":"456","updatedBy":"user"}`
	r := httptest.NewRequest(http.MethodPut, "/domicilios?id=1", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
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

func TestDomicilioGetByIdSubtotalFallback(t *testing.T) {
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
			return &mockRows{columns: []string{"documento"}, values: [][]driver.Value{}}, nil
		case 3:
			cols := []string{"pedido_id", "pago_id", "pago_monto", "subtotal_productos", "productos"}
			vals := [][]driver.Value{{int64(5), nil, nil, float64(15), ""}}
			return &mockRows{columns: cols, values: vals}, nil
		default:
			return nil, errors.New("unexpected query")
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
	if !strings.Contains(w.Body.String(), "\"total\":15") {
		t.Fatalf("expected total 15, got %s", w.Body.String())
	}
}

func TestAsignarDomiciliarioNotFound(t *testing.T) {
	origE := MockExec
	MockExec = func(_ stdctx.Context, _ string, _ []driver.NamedValue) (driver.Result, error) {
		return zeroResult{}, nil
	}
	origQ := MockQuery
	MockQuery = func(_ stdctx.Context, _ string, _ []driver.NamedValue) (driver.Rows, error) {
		cols := []string{"count"}
		vals := [][]driver.Value{{0}}
		return &mockRows{columns: cols, values: vals}, nil
	}
	t.Cleanup(func() { MockExec = origE; MockQuery = origQ })

	r := httptest.NewRequest(http.MethodPost, "/domicilios/asignar?domicilio_id=1&trabajador_id=2", nil)
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	c := DomicilioController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.AsignarDomiciliario()
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestDomicilioGetAllInvalidTrabajador(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/domicilios?trabajador=abc", nil)
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	c := DomicilioController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.GetAll()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestDomicilioPutUpdateError(t *testing.T) {
	origQ := MockQuery
	MockQuery = func(_ stdctx.Context, _ string, _ []driver.NamedValue) (driver.Rows, error) {
		cols := []string{"pk_id_domicilio", "direccion", "telefono", "estado_domicilio", "entregado", "fecha", "observaciones", "created_at", "updated_at", "created_by", "updated_by", "pk_documento_trabajador"}
		vals := [][]driver.Value{{int64(1), "C", "1", "PENDIENTE", false, time.Now(), nil, time.Now(), time.Now(), nil, nil, nil}}
		return &mockRows{columns: cols, values: vals}, nil
	}
	origE := MockExec
	MockExec = func(_ stdctx.Context, _ string, _ []driver.NamedValue) (driver.Result, error) {
		return nil, errors.New("update fail")
	}
	t.Cleanup(func() { MockQuery = origQ; MockExec = origE })

	body := `{"direccion":"n","telefono":"1"}`
	r := httptest.NewRequest(http.MethodPut, "/domicilios?id=1", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := DomicilioController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Put()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestDomicilioGetByIdQueryErrors(t *testing.T) {
	call := 0
	origQ := MockQuery
	MockQuery = func(_ stdctx.Context, _ string, _ []driver.NamedValue) (driver.Rows, error) {
		call++
		switch call {
		case 1:
			cols := []string{"pk_id_domicilio", "direccion", "telefono", "estado_domicilio", "entregado", "fecha", "observaciones", "created_at", "updated_at", "created_by", "updated_by", "pk_documento_trabajador"}
			vals := [][]driver.Value{{int64(1), "C", "1", "PENDIENTE", false, time.Now(), nil, time.Now(), time.Now(), nil, nil, nil}}
			return &mockRows{columns: cols, values: vals}, nil
		default:
			return nil, errors.New("db error")
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
	body := w.Body.String()
	if strings.Contains(body, "\"cliente\"") || strings.Contains(body, "\"pedido\"") {
		t.Fatalf("unexpected related data: %s", body)
	}
}
