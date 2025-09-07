package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
)

func TestMetodoPagoGetByIdInvalidID(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/metodos_pago/search", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := MetodoPagoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetById()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestMetodoPagoPostInvalidJSON(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/metodos_pago", strings.NewReader("notjson"))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := MetodoPagoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestMetodoPagoGetAllWithoutDB(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/metodos_pago", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := MetodoPagoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetAll()

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestMetodoPagoGetAllSuccess(t *testing.T) {
	m := newMockMetodoPagoOrmer()
	original := getOrm
	getOrm = func() metodoPagoOrmer { return m }
	defer func() { getOrm = original }()

	// poblar datos
	_, _ = m.Insert(&models.MetodoPago{TIPO: "efectivo"})
	_, _ = m.Insert(&models.MetodoPago{TIPO: "tarjeta"})

	r := httptest.NewRequest(http.MethodGet, "/metodos_pago", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := MetodoPagoController{}
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
		t.Fatalf("expected code 200, got %d", resp.Code)
	}
}

func TestMetodoPagoPutInvalidID(t *testing.T) {
	r := httptest.NewRequest(http.MethodPut, "/metodos_pago", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := MetodoPagoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Put()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestMetodoPagoPutNotFound(t *testing.T) {
	r := httptest.NewRequest(http.MethodPut, "/metodos_pago?id=1", strings.NewReader("{}"))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := MetodoPagoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Put()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestMetodoPagoDeleteInvalidID(t *testing.T) {
	r := httptest.NewRequest(http.MethodDelete, "/metodos_pago", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := MetodoPagoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Delete()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestMetodoPagoDeleteNotFound(t *testing.T) {
	r := httptest.NewRequest(http.MethodDelete, "/metodos_pago?id=1", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := MetodoPagoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Delete()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

// mockMetodoPagoOrmer provides an in-memory implementation of metodoPagoOrmer
// used to test the controller without touching a real database.
type mockMetodoPagoOrmer struct {
	data   map[int64]models.MetodoPago
	nextID int64
}

func newMockMetodoPagoOrmer() *mockMetodoPagoOrmer {
	return &mockMetodoPagoOrmer{data: make(map[int64]models.MetodoPago), nextID: 1}
}

func (m *mockMetodoPagoOrmer) QueryTable(_ interface{}) metodoPagoQuerySeter {
	return m
}

func (m *mockMetodoPagoOrmer) All(res interface{}, _ ...string) (int64, error) {
	slice, ok := res.(*[]models.MetodoPago)
	if !ok {
		return 0, errors.New("invalid result type")
	}
	for _, v := range m.data {
		*slice = append(*slice, v)
	}
	return int64(len(m.data)), nil
}

func (m *mockMetodoPagoOrmer) Read(v interface{}, _ ...string) error {
	mp := v.(*models.MetodoPago)
	item, ok := m.data[mp.PK_ID_METODO_PAGO]
	if !ok {
		return orm.ErrNoRows
	}
	*mp = item
	return nil
}

func (m *mockMetodoPagoOrmer) Insert(v interface{}) (int64, error) {
	mp := v.(*models.MetodoPago)
	mp.PK_ID_METODO_PAGO = m.nextID
	m.nextID++
	m.data[mp.PK_ID_METODO_PAGO] = *mp
	return 1, nil
}

func (m *mockMetodoPagoOrmer) Update(v interface{}, _ ...string) (int64, error) {
	mp := v.(*models.MetodoPago)
	if _, ok := m.data[mp.PK_ID_METODO_PAGO]; !ok {
		return 0, orm.ErrNoRows
	}
	m.data[mp.PK_ID_METODO_PAGO] = *mp
	return 1, nil
}

func (m *mockMetodoPagoOrmer) Delete(v interface{}, _ ...string) (int64, error) {
	mp := v.(*models.MetodoPago)
	if _, ok := m.data[mp.PK_ID_METODO_PAGO]; !ok {
		return 0, orm.ErrNoRows
	}
	delete(m.data, mp.PK_ID_METODO_PAGO)
	return 1, nil
}

func TestMetodoPagoGetByIdSuccess(t *testing.T) {
	m := newMockMetodoPagoOrmer()
	original := getOrm
	getOrm = func() metodoPagoOrmer { return m }
	defer func() { getOrm = original }()

	mp := models.MetodoPago{TIPO: "efectivo"}
	if _, err := m.Insert(&mp); err != nil {
		t.Fatalf("insert: %v", err)
	}

	r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/metodos_pago/search?id=%d", mp.PK_ID_METODO_PAGO), nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := MetodoPagoController{}
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
	if resp.Code != http.StatusOK {
		t.Fatalf("expected code 200, got %d", resp.Code)
	}
}

func TestMetodoPagoGetByIdNotFound(t *testing.T) {
	m := newMockMetodoPagoOrmer()
	original := getOrm
	getOrm = func() metodoPagoOrmer { return m }
	defer func() { getOrm = original }()

	r := httptest.NewRequest(http.MethodGet, "/metodos_pago/search?id=1", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := MetodoPagoController{}
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
		t.Fatalf("expected code 404, got %d", resp.Code)
	}
}

func TestMetodoPagoPostSuccess(t *testing.T) {
	m := newMockMetodoPagoOrmer()
	original := getOrm
	getOrm = func() metodoPagoOrmer { return m }
	defer func() { getOrm = original }()

	body := `{"tipo":"tarjeta","detalle":"visa"}`
	r := httptest.NewRequest(http.MethodPost, "/metodos_pago", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := MetodoPagoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", w.Code)
	}
	if len(m.data) != 1 {
		t.Fatalf("expected 1 record, got %d", len(m.data))
	}
}

// failInsertOrmer envuelve un mock y fuerza error en Insert para cubrir la rama 500

type failInsertOrmer struct{ *mockMetodoPagoOrmer }

func (f failInsertOrmer) Insert(v interface{}) (int64, error) { return 0, errors.New("db") }

func TestMetodoPagoPostInsertError(t *testing.T) {
	original := getOrm
	getOrm = func() metodoPagoOrmer { return failInsertOrmer{newMockMetodoPagoOrmer()} }
	defer func() { getOrm = original }()

	body := `{"tipo":"tarjeta"}`
	r := httptest.NewRequest(http.MethodPost, "/metodos_pago", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := MetodoPagoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestMetodoPagoPutSuccess(t *testing.T) {
	m := newMockMetodoPagoOrmer()
	original := getOrm
	getOrm = func() metodoPagoOrmer { return m }
	defer func() { getOrm = original }()

	mp := models.MetodoPago{TIPO: "efectivo"}
	if _, err := m.Insert(&mp); err != nil {
		t.Fatalf("insert: %v", err)
	}

	body := `{"tipo":"tarjeta"}`
	r := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/metodos_pago?id=%d", mp.PK_ID_METODO_PAGO), strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := MetodoPagoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Put()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if m.data[mp.PK_ID_METODO_PAGO].TIPO != "tarjeta" {
		t.Fatalf("expected tipo updated to tarjeta, got %s", m.data[mp.PK_ID_METODO_PAGO].TIPO)
	}
}

// failUpdateOrmer fuerza error en Update para cubrir rama 500 del Put
type failUpdateOrmer struct{ *mockMetodoPagoOrmer }

func (f failUpdateOrmer) Read(v interface{}, _ ...string) error { return nil }
func (f failUpdateOrmer) Update(v interface{}, _ ...string) (int64, error) { return 0, errors.New("db") }

func TestMetodoPagoPutUpdateError(t *testing.T) {
    original := getOrm
    getOrm = func() metodoPagoOrmer { return failUpdateOrmer{newMockMetodoPagoOrmer()} }
    defer func() { getOrm = original }()

    body := `{"tipo":"x"}`
    r := httptest.NewRequest(http.MethodPut, "/metodos_pago?id=1", strings.NewReader(body))
    w := httptest.NewRecorder()
    ctx := context.NewContext(); ctx.Reset(w, r)
    ctx.Input.RequestBody = []byte(body)
    c := MetodoPagoController{}; c.Ctx = ctx; c.Data = make(map[interface{}]interface{})

    c.Put()

    if w.Code != http.StatusInternalServerError {
        t.Fatalf("expected 500, got %d", w.Code)
    }
}

func TestMetodoPagoPutInvalidJSONWithRecord(t *testing.T) {
	m := newMockMetodoPagoOrmer()
	original := getOrm
	getOrm = func() metodoPagoOrmer { return m }
	defer func() { getOrm = original }()

	mp := models.MetodoPago{TIPO: "efectivo"}
	if _, err := m.Insert(&mp); err != nil {
		t.Fatalf("insert: %v", err)
	}

	r := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/metodos_pago?id=%d", mp.PK_ID_METODO_PAGO), strings.NewReader("notjson"))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte("notjson")
	c := MetodoPagoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Put()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestMetodoPagoDeleteSuccess(t *testing.T) {
	m := newMockMetodoPagoOrmer()
	original := getOrm
	getOrm = func() metodoPagoOrmer { return m }
	defer func() { getOrm = original }()

	mp := models.MetodoPago{TIPO: "efectivo"}
	if _, err := m.Insert(&mp); err != nil {
		t.Fatalf("insert: %v", err)
	}

	r := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/metodos_pago?id=%d", mp.PK_ID_METODO_PAGO), nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := MetodoPagoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Delete()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if len(m.data) != 0 {
		t.Fatalf("expected 0 records, got %d", len(m.data))
	}
}
