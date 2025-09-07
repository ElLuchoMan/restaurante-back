package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
)

// Mocks para SubcategoriaController
type subQS struct{ list []models.Subcategoria }
func (m subQS) All(res interface{}, _ ...string) (int64, error) { dst := res.(*[]models.Subcategoria); *dst = append(*dst, m.list...); return int64(len(m.list)), nil }
func (m subQS) Filter(_ string, _ ...interface{}) subcatQuerySeter { return m }

type subMockOrm struct{ items map[int64]models.Subcategoria; next int64; qs subQS }
func newSubMockOrm() *subMockOrm { return &subMockOrm{items: map[int64]models.Subcategoria{}, next: 1} }
func (m *subMockOrm) QueryTable(_ interface{}) subcatQuerySeter { return m.qs }
func (m *subMockOrm) Insert(v interface{}) (int64, error) { s := v.(*models.Subcategoria); s.PK_ID_SUBCATEGORIA = m.next; m.items[m.next] = *s; m.next++; return 1, nil }
func (m *subMockOrm) Read(v interface{}, _ ...string) error { s := v.(*models.Subcategoria); it, ok := m.items[s.PK_ID_SUBCATEGORIA]; if !ok { return orm.ErrNoRows }; *s = it; return nil }
func (m *subMockOrm) Update(v interface{}, _ ...string) (int64, error) { s := v.(*models.Subcategoria); if _, ok := m.items[s.PK_ID_SUBCATEGORIA]; !ok { return 0, orm.ErrNoRows }; m.items[s.PK_ID_SUBCATEGORIA] = *s; return 1, nil }
func (m *subMockOrm) Delete(v interface{}, _ ...string) (int64, error) { s := v.(*models.Subcategoria); if _, ok := m.items[s.PK_ID_SUBCATEGORIA]; !ok { return 0, orm.ErrNoRows }; delete(m.items, s.PK_ID_SUBCATEGORIA); return 1, nil }

func TestSubcategoriaController_FullCoverage(t *testing.T) {
    m := newSubMockOrm()
    m.qs = subQS{list: []models.Subcategoria{{PK_ID_SUBCATEGORIA: 7, NOMBRE: "Gaseosas"}}}
    orig := subcatOrmNew
    subcatOrmNew = func() subcatOrmer { return m }
    defer func() { subcatOrmNew = orig }()

    // GetAll
    r := httptest.NewRequest(http.MethodGet, "/subcategorias", nil)
    w := httptest.NewRecorder(); ctx := context.NewContext(); ctx.Reset(w, r)
    c := &SubcategoriaController{}; c.Ctx = ctx; c.Data = make(map[interface{}]interface{})
    c.GetAll(); if w.Code != http.StatusOK { t.Fatalf("getAll %d", w.Code) }

    // GetById not found
    r = httptest.NewRequest(http.MethodGet, "/subcategorias/search?id=5", nil)
    w = httptest.NewRecorder(); ctx = context.NewContext(); ctx.Reset(w, r)
    c = &SubcategoriaController{}; c.Ctx = ctx; c.Data = make(map[interface{}]interface{})
    c.GetById(); var resp models.ApiResponse; _ = json.Unmarshal(w.Body.Bytes(), &resp); if resp.Code != http.StatusNotFound { t.Fatalf("expected 404, got %d", resp.Code) }

    // Post invalid
    r = httptest.NewRequest(http.MethodPost, "/subcategorias", strings.NewReader("bad"))
    w = httptest.NewRecorder(); ctx = context.NewContext(); ctx.Reset(w, r); ctx.Input.RequestBody = []byte("bad")
    c = &SubcategoriaController{}; c.Ctx = ctx; c.Data = make(map[interface{}]interface{})
    c.Post(); if w.Code != http.StatusBadRequest { t.Fatalf("post bad %d", w.Code) }

    // Post ok
    body := `{"nombre":"Jugos","categoriaId":1}`
    r = httptest.NewRequest(http.MethodPost, "/subcategorias", strings.NewReader(body))
    w = httptest.NewRecorder(); ctx = context.NewContext(); ctx.Reset(w, r); ctx.Input.RequestBody = []byte(body)
    c = &SubcategoriaController{}; c.Ctx = ctx; c.Data = make(map[interface{}]interface{})
    c.Post(); if w.Code != http.StatusCreated { t.Fatalf("post ok %d", w.Code) }

    // Put not found
    r = httptest.NewRequest(http.MethodPut, "/subcategorias?id=9", strings.NewReader(`{"nombre":"Refrescos"}`))
    w = httptest.NewRecorder(); ctx = context.NewContext(); ctx.Reset(w, r); ctx.Input.RequestBody = []byte(`{"nombre":"Refrescos"}`)
    c = &SubcategoriaController{}; c.Ctx = ctx; c.Data = make(map[interface{}]interface{})
    c.Put(); if w.Code != http.StatusOK { t.Fatalf("put nf %d", w.Code) }

    // Put ok
    sub := models.Subcategoria{NOMBRE: "Sodas"}; m.Insert(&sub)
    r = httptest.NewRequest(http.MethodPut, "/subcategorias?id=1", strings.NewReader(`{"nombre":"Sodas light"}`))
    w = httptest.NewRecorder(); ctx = context.NewContext(); ctx.Reset(w, r); ctx.Input.RequestBody = []byte(`{"nombre":"Sodas light"}`)
    c = &SubcategoriaController{}; c.Ctx = ctx; c.Data = make(map[interface{}]interface{})
    c.Put(); if w.Code != http.StatusOK { t.Fatalf("put ok %d", w.Code) }

    // GetById ok
    r = httptest.NewRequest(http.MethodGet, "/subcategorias/search?id=1", nil)
    w = httptest.NewRecorder(); ctx = context.NewContext(); ctx.Reset(w, r)
    c = &SubcategoriaController{}; c.Ctx = ctx; c.Data = make(map[interface{}]interface{})
    c.GetById(); _ = json.Unmarshal(w.Body.Bytes(), &resp); if resp.Code != http.StatusOK { t.Fatalf("get id ok %d", resp.Code) }

    // Delete ok
    r = httptest.NewRequest(http.MethodDelete, "/subcategorias?id=1", nil)
    w = httptest.NewRecorder(); ctx = context.NewContext(); ctx.Reset(w, r)
    c = &SubcategoriaController{}; c.Ctx = ctx; c.Data = make(map[interface{}]interface{})
    c.Delete(); if w.Code != http.StatusOK { t.Fatalf("del ok %d", w.Code) }
}

// Tipos de prueba para error en GetAll de SubcategoriaController
type badQSSub struct{}

func (badQSSub) All(res interface{}, _ ...string) (int64, error) { return 0, orm.ErrNoRows }
func (badQSSub) Filter(_ string, _ ...interface{}) subcatQuerySeter { return badQSSub{} }

type badOrmSub struct{}

func (badOrmSub) QueryTable(_ interface{}) subcatQuerySeter { return badQSSub{} }
func (badOrmSub) Insert(interface{}) (int64, error) { return 0, nil }
func (badOrmSub) Read(interface{}, ...string) error { return nil }
func (badOrmSub) Update(interface{}, ...string) (int64, error) { return 1, nil }
func (badOrmSub) Delete(interface{}, ...string) (int64, error) { return 1, nil }

func TestSubcategoriaController_AllError(t *testing.T) {
    orig := subcatOrmNew
    subcatOrmNew = func() subcatOrmer { return badOrmSub{} }
    defer func() { subcatOrmNew = orig }()

    r := httptest.NewRequest(http.MethodGet, "/subcategorias", nil)
    w := httptest.NewRecorder(); ctx := context.NewContext(); ctx.Reset(w, r)
    c := &SubcategoriaController{}; c.Ctx = ctx; c.Data = make(map[interface{}]interface{})
    c.GetAll()
    if w.Code != http.StatusInternalServerError { t.Fatalf("expected 500, got %d", w.Code) }
}


