package controllers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
	"restaurante/database"
	"restaurante/models"
)

type fakeQuery struct {
	orm.QuerySeter
	all func(interface{}, ...string) (int64, error)
}

func (q fakeQuery) All(container interface{}, cols ...string) (int64, error) {
	return q.all(container, cols...)
}

type fakeOrmer struct {
	queryAll func(interface{}, ...string) (int64, error)
	read     func(interface{}, ...string) error
	insert   func(interface{}) (int64, error)
	update   func(interface{}, ...string) (int64, error)
	delete   func(interface{}, ...string) (int64, error)
}

func (f fakeOrmer) QueryTable(i interface{}) orm.QuerySeter {
	return fakeQuery{all: f.queryAll}
}

func (f fakeOrmer) Read(m interface{}, cols ...string) error {
	if f.read != nil {
		return f.read(m, cols...)
	}
	return nil
}

func (f fakeOrmer) Insert(m interface{}) (int64, error) {
	if f.insert != nil {
		return f.insert(m)
	}
	return 0, nil
}

func (f fakeOrmer) Update(m interface{}, cols ...string) (int64, error) {
	if f.update != nil {
		return f.update(m, cols...)
	}
	return 0, nil
}

func (f fakeOrmer) Delete(m interface{}, cols ...string) (int64, error) {
	if f.delete != nil {
		return f.delete(m, cols...)
	}
	return 0, nil
}

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
	body := `{"horaPago":"10:00:00","monto":1000,"estadoPago":"pagado","metodoPagoId":1}`
	r := httptest.NewRequest(http.MethodPost, "/pagos", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := PagoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Ctx.Input.RequestBody = []byte(body)

	c.Post()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestPagoPostInvalidHora(t *testing.T) {
	body := `{"fechaPago":"2024-01-01","horaPago":"25:00","monto":1000,"estadoPago":"pagado","metodoPagoId":1}`
	r := httptest.NewRequest(http.MethodPost, "/pagos", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := PagoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Ctx.Input.RequestBody = []byte(body)

	c.Post()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestPagoPostInvalidEstado(t *testing.T) {
	body := `{"fechaPago":"2024-01-01","horaPago":"10:00:00","monto":1000,"estadoPago":"INVALIDO","metodoPagoId":1}`
	r := httptest.NewRequest(http.MethodPost, "/pagos", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := PagoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Ctx.Input.RequestBody = []byte(body)

	c.Post()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestPagoPostMissingMetodoPago(t *testing.T) {
	body := `{"fechaPago":"2024-01-01","horaPago":"10:00:00","monto":1000,"estadoPago":"pagado"}`
	r := httptest.NewRequest(http.MethodPost, "/pagos", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := PagoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Ctx.Input.RequestBody = []byte(body)
	c.Post()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestPagoPostMissingHora(t *testing.T) {
	body := `{"fechaPago":"2024-01-01","monto":1000,"estadoPago":"pagado","metodoPagoId":1}`
	r := httptest.NewRequest(http.MethodPost, "/pagos", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := PagoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Ctx.Input.RequestBody = []byte(body)
	c.Post()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestPagoPostSuccess(t *testing.T) {
	body := `{"estadoPago":"pagado","fechaPago":"2024-01-01","horaPago":"10:00:00","metodoPagoId":1,"monto":1000}`
	r := httptest.NewRequest(http.MethodPost, "/pagos", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)

	original := pagoNewOrm
	pagoNewOrm = func() ormer { return fakeOrmer{insert: func(interface{}) (int64, error) { return 1, nil }} }
	defer func() { pagoNewOrm = original }()

	c := PagoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Pago creado") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestPagoPostMissingMonto(t *testing.T) {
	body := `{"fechaPago":"2024-01-01","horaPago":"10:00:00","estadoPago":"pagado","metodoPagoId":1}`
	r := httptest.NewRequest(http.MethodPost, "/pagos", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := PagoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Ctx.Input.RequestBody = []byte(body)
	c.Post()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestPagoPostInvalidFecha(t *testing.T) {
	body := `{"fechaPago":"2024-13-01","horaPago":"10:00:00","monto":1000,"estadoPago":"pagado","metodoPagoId":1}`
	r := httptest.NewRequest(http.MethodPost, "/pagos", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := PagoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Ctx.Input.RequestBody = []byte(body)
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

func TestPagoPutInvalidJSON(t *testing.T) {
	orig := pagoNewOrm
	pagoNewOrm = func() ormer { return fakeOrmer{read: func(m interface{}, cols ...string) error { return nil }} }
	t.Cleanup(func() { pagoNewOrm = orig })
	r := httptest.NewRequest(http.MethodPut, "/pagos?id=1", strings.NewReader("notjson"))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := PagoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Ctx.Input.RequestBody = []byte("notjson")
	c.Put()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestPagoPutMissingHora(t *testing.T) {
	body := `{"fechaPago":"2024-01-01"}`
	r := httptest.NewRequest(http.MethodPut, "/pagos?id=1", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := PagoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Ctx.Input.RequestBody = []byte(body)

	c.Put()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestPagoPutMissingMetodoPago(t *testing.T) {
	orig := pagoNewOrm
	pagoNewOrm = func() ormer { return fakeOrmer{read: func(m interface{}, cols ...string) error { return nil }} }
	t.Cleanup(func() { pagoNewOrm = orig })
	body := `{"fechaPago":"2024-01-01","horaPago":"10:00:00"}`
	r := httptest.NewRequest(http.MethodPut, "/pagos?id=1", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := PagoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Ctx.Input.RequestBody = []byte(body)
	c.Put()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestPagoPutInvalidFecha(t *testing.T) {
	orig := pagoNewOrm
	pagoNewOrm = func() ormer {
		return fakeOrmer{read: func(m interface{}, cols ...string) error { return nil }}
	}
	t.Cleanup(func() { pagoNewOrm = orig })
	body := `{"fechaPago":"2024-13-01","horaPago":"10:00:00","metodoPagoId":1}`
	r := httptest.NewRequest(http.MethodPut, "/pagos?id=1", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := PagoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Ctx.Input.RequestBody = []byte(body)
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

func TestPagoGetAllSuccess(t *testing.T) {
	orig := pagoNewOrm
	pagoNewOrm = func() ormer {
		pagos := []models.Pago{{PK_ID_PAGO: int64(1), FECHA: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), HORA: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC), ESTADO_PAGO: "pagado", PK_ID_METODO_PAGO: int64(1)}}
		return fakeOrmer{queryAll: func(res interface{}, cols ...string) (int64, error) {
			ptr := res.(*[]models.Pago)
			*ptr = pagos
			return int64(len(pagos)), nil
		}}
	}
	t.Cleanup(func() { pagoNewOrm = orig })
	r := httptest.NewRequest(http.MethodGet, "/pagos", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := PagoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.GetAll()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestPagoGetAllFilterFecha(t *testing.T) {
	orig := pagoNewOrm
	pagoNewOrm = func() ormer {
		pagos := []models.Pago{
			{PK_ID_PAGO: int64(1), FECHA: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), HORA: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)},
			{PK_ID_PAGO: int64(2), FECHA: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), HORA: time.Date(2024, 1, 2, 11, 0, 0, 0, time.UTC)},
		}
		return fakeOrmer{queryAll: func(res interface{}, cols ...string) (int64, error) {
			ptr := res.(*[]models.Pago)
			*ptr = pagos
			return int64(len(pagos)), nil
		}}
	}
	t.Cleanup(func() { pagoNewOrm = orig })
	r := httptest.NewRequest(http.MethodGet, "/pagos?fecha=2024-01-01", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := PagoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.GetAll()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestPagoGetAllFilterOthers(t *testing.T) {
	orig := pagoNewOrm
	pagoNewOrm = func() ormer {
		pagos := []models.Pago{
			{FECHA: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), HORA: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC), ESTADO_PAGO: "pagado", PK_ID_METODO_PAGO: int64(1)},
			{FECHA: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), HORA: time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC), ESTADO_PAGO: "pagado", PK_ID_METODO_PAGO: int64(1)},
			{FECHA: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC), HORA: time.Date(2024, 2, 1, 10, 0, 0, 0, time.UTC), ESTADO_PAGO: "pagado", PK_ID_METODO_PAGO: int64(1)},
			{FECHA: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), HORA: time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC), ESTADO_PAGO: "pagado", PK_ID_METODO_PAGO: int64(1)},
			{FECHA: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), HORA: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC), ESTADO_PAGO: "no pago", PK_ID_METODO_PAGO: int64(1)},
			{FECHA: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), HORA: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC), ESTADO_PAGO: "pagado", PK_ID_METODO_PAGO: int64(2)},
		}
		return fakeOrmer{queryAll: func(res interface{}, cols ...string) (int64, error) {
			ptr := res.(*[]models.Pago)
			*ptr = pagos
			return int64(len(pagos)), nil
		}}
	}
	t.Cleanup(func() { pagoNewOrm = orig })
	r := httptest.NewRequest(http.MethodGet, "/pagos?dia=1&mes=1&anio=2024&estado=pagado&metodo_pago=1", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := PagoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.GetAll()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestPagoGetAllNoResults(t *testing.T) {
	orig := pagoNewOrm
	pagoNewOrm = func() ormer {
		pagos := []models.Pago{{FECHA: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), HORA: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC), ESTADO_PAGO: "pagado", PK_ID_METODO_PAGO: int64(1)}}
		return fakeOrmer{queryAll: func(res interface{}, cols ...string) (int64, error) {
			ptr := res.(*[]models.Pago)
			*ptr = pagos
			return int64(len(pagos)), nil
		}}
	}
	t.Cleanup(func() { pagoNewOrm = orig })
	r := httptest.NewRequest(http.MethodGet, "/pagos?estado=no%20pago", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := PagoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.GetAll()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestPagoGetByIdNotFound(t *testing.T) {
	orig := pagoNewOrm
	pagoNewOrm = func() ormer {
		return fakeOrmer{read: func(m interface{}, cols ...string) error { return orm.ErrNoRows }}
	}
	t.Cleanup(func() { pagoNewOrm = orig })
	r := httptest.NewRequest(http.MethodGet, "/pagos/search?id=1", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := PagoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.GetById()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestPagoGetByIdSuccess(t *testing.T) {
	orig := pagoNewOrm
	pagoNewOrm = func() ormer {
		return fakeOrmer{read: func(m interface{}, cols ...string) error {
			p := m.(*models.Pago)
			*p = models.Pago{PK_ID_PAGO: int64(1), FECHA: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), HORA: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC), UPDATED_AT: time.Now()}
			return nil
		}}
	}
	t.Cleanup(func() { pagoNewOrm = orig })
	database.BogotaZone = time.UTC
	r := httptest.NewRequest(http.MethodGet, "/pagos/search?id=1", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := PagoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.GetById()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestPagoPostInsertError(t *testing.T) {
	orig := pagoNewOrm
	pagoNewOrm = func() ormer {
		return fakeOrmer{insert: func(m interface{}) (int64, error) { return 0, fmt.Errorf("fail") }}
	}
	t.Cleanup(func() { pagoNewOrm = orig })
	body := `{"estadoPago":"pagado","fechaPago":"2024-01-01","horaPago":"10:00:00","metodoPagoId":1,"monto":1000}`
	r := httptest.NewRequest(http.MethodPost, "/pagos", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := PagoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Ctx.Input.RequestBody = []byte(body)
	c.Post()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestPagoPutSuccess(t *testing.T) {
	orig := pagoNewOrm
	pagoNewOrm = func() ormer {
		return fakeOrmer{
			read:   func(m interface{}, cols ...string) error { return nil },
			update: func(m interface{}, cols ...string) (int64, error) { return 1, nil },
		}
	}
	t.Cleanup(func() { pagoNewOrm = orig })
	body := `{"FECHA":"2024-02-02","HORA":"11:00:00","MONTO":2000,"ESTADO_PAGO":"pendiente","UPDATED_BY":"me","PK_ID_METODO_PAGO":1}`
	r := httptest.NewRequest(http.MethodPut, "/pagos?id=1", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := PagoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Ctx.Input.RequestBody = []byte(body)
	c.Put()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body %s", w.Code, w.Body.String())
	}
}

func TestPagoPutSuccessWithoutUpdatedBy(t *testing.T) {
	orig := pagoNewOrm
	pagoNewOrm = func() ormer {
		return fakeOrmer{
			read:   func(m interface{}, cols ...string) error { return nil },
			update: func(m interface{}, cols ...string) (int64, error) { return 1, nil },
		}
	}
	t.Cleanup(func() { pagoNewOrm = orig })
	body := `{"FECHA":"2024-02-02","HORA":"11:00:00","MONTO":2000,"ESTADO_PAGO":"pendiente","PK_ID_METODO_PAGO":1}`
	r := httptest.NewRequest(http.MethodPut, "/pagos?id=1", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := PagoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Ctx.Input.RequestBody = []byte(body)
	c.Put()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body %s", w.Code, w.Body.String())
	}
	if strings.Contains(w.Body.String(), "updatedBy") {
		t.Fatalf("response should not include updatedBy: %s", w.Body.String())
	}
}

func TestPagoPutInvalidHora(t *testing.T) {
	orig := pagoNewOrm
	pagoNewOrm = func() ormer {
		return fakeOrmer{read: func(m interface{}, cols ...string) error { return nil }}
	}
	t.Cleanup(func() { pagoNewOrm = orig })
	body := `{"fechaPago":"2024-02-02","horaPago":"25:00:00","metodoPagoId":1}`
	r := httptest.NewRequest(http.MethodPut, "/pagos?id=1", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := PagoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Ctx.Input.RequestBody = []byte(body)
	c.Put()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestPagoPutInvalidEstado(t *testing.T) {
	orig := pagoNewOrm
	pagoNewOrm = func() ormer {
		return fakeOrmer{read: func(m interface{}, cols ...string) error { return nil }}
	}
	t.Cleanup(func() { pagoNewOrm = orig })
	body := `{"fechaPago":"2024-02-02","horaPago":"11:00:00","estadoPago":"MALO","metodoPagoId":1}`
	r := httptest.NewRequest(http.MethodPut, "/pagos?id=1", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := PagoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Ctx.Input.RequestBody = []byte(body)
	c.Put()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestPagoPutNotFound(t *testing.T) {
	orig := pagoNewOrm
	pagoNewOrm = func() ormer {
		return fakeOrmer{read: func(m interface{}, cols ...string) error { return orm.ErrNoRows }}
	}
	t.Cleanup(func() { pagoNewOrm = orig })
	body := `{"FECHA":"2024-01-01","HORA":"10:00:00","PK_ID_METODO_PAGO":1}`
	r := httptest.NewRequest(http.MethodPut, "/pagos?id=1", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := PagoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Ctx.Input.RequestBody = []byte(body)
	c.Put()
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestPagoPutUpdateError(t *testing.T) {
	orig := pagoNewOrm
	pagoNewOrm = func() ormer {
		return fakeOrmer{
			read:   func(m interface{}, cols ...string) error { return nil },
			update: func(m interface{}, cols ...string) (int64, error) { return 0, fmt.Errorf("fail") },
		}
	}
	t.Cleanup(func() { pagoNewOrm = orig })
	body := `{"FECHA":"2024-01-01","HORA":"10:00:00","PK_ID_METODO_PAGO":1}`
	r := httptest.NewRequest(http.MethodPut, "/pagos?id=1", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := PagoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Ctx.Input.RequestBody = []byte(body)
	c.Put()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestPagoDeleteSuccess(t *testing.T) {
	orig := pagoNewOrm
	pagoNewOrm = func() ormer {
		return fakeOrmer{delete: func(m interface{}, cols ...string) (int64, error) { return 1, nil }}
	}
	t.Cleanup(func() { pagoNewOrm = orig })
	r := httptest.NewRequest(http.MethodDelete, "/pagos?id=1", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := PagoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Delete()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
