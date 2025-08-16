package controllers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beego/beego/v2/server/web/context"
)

func TestPagoGetByIdInvalidID(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/pagos/search", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := PagoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetById()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestPagoPostInvalidJSON(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/pagos", strings.NewReader("notjson"))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := PagoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestPagoGetAllWithoutDB(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/pagos", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := PagoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetAll()

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}
func TestPagoPostMissingFecha(t *testing.T) {
	body := `{"HORA":"10:00:00","MONTO":1000,"ESTADO_PAGO":"PAGADO","PK_ID_METODO_PAGO":1}`
	r := httptest.NewRequest(http.MethodPost, "/pagos", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := PagoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestPagoPostInvalidHora(t *testing.T) {
	body := `{"FECHA":"2024-01-01","HORA":"25:00","MONTO":1000,"ESTADO_PAGO":"PAGADO","PK_ID_METODO_PAGO":1}`
	r := httptest.NewRequest(http.MethodPost, "/pagos", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := PagoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestPagoPostInvalidEstado(t *testing.T) {
	body := `{"FECHA":"2024-01-01","HORA":"10:00:00","MONTO":1000,"ESTADO_PAGO":"INVALIDO","PK_ID_METODO_PAGO":1}`
	r := httptest.NewRequest(http.MethodPost, "/pagos", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := PagoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestPagoPostMissingMetodoPago(t *testing.T) {
	body := `{"FECHA":"2024-01-01","HORA":"10:00:00","MONTO":1000,"ESTADO_PAGO":"PAGADO"}`
	r := httptest.NewRequest(http.MethodPost, "/pagos", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := PagoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestPagoPutInvalidID(t *testing.T) {
	r := httptest.NewRequest(http.MethodPut, "/pagos", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := PagoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Put()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestPagoPutMissingHora(t *testing.T) {
	body := `{"FECHA":"2024-01-01"}`
	r := httptest.NewRequest(http.MethodPut, "/pagos?id=1", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := PagoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Put()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestPagoPutMissingMetodoPago(t *testing.T) {
	body := `{"FECHA":"2024-01-01","HORA":"10:00:00"}`
	r := httptest.NewRequest(http.MethodPut, "/pagos?id=1", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := PagoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Put()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestPagoDeleteInvalidID(t *testing.T) {
	r := httptest.NewRequest(http.MethodDelete, "/pagos", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := PagoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Delete()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestPagoDeleteNotFound(t *testing.T) {
	r := httptest.NewRequest(http.MethodDelete, "/pagos?id=1", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := PagoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Delete()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}
