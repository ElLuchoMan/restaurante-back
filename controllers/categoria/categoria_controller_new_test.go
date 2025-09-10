package categoria

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

// Mocks para CategoriaController
type catQS struct{ list []models.Categoria }

func (m catQS) All(res interface{}, _ ...string) (int64, error) {
	dst := res.(*[]models.Categoria)
	*dst = append(*dst, m.list...)
	return int64(len(m.list)), nil
}

type catMockOrm struct {
	items map[int64]models.Categoria
	next  int64
	qs    catQS
}

func newCatMockOrm() *catMockOrm                                   { return &catMockOrm{items: map[int64]models.Categoria{}, next: 1} }
func (m *catMockOrm) QueryTable(_ interface{}) categoriaQuerySeter { return m.qs }
func (m *catMockOrm) Insert(v interface{}) (int64, error) {
	c := v.(*models.Categoria)
	c.PK_ID_CATEGORIA = m.next
	m.items[m.next] = *c
	m.next++
	return 1, nil
}
func (m *catMockOrm) Read(v interface{}, _ ...string) error {
	c := v.(*models.Categoria)
	it, ok := m.items[c.PK_ID_CATEGORIA]
	if !ok {
		return orm.ErrNoRows
	}
	*c = it
	return nil
}
func (m *catMockOrm) Update(v interface{}, _ ...string) (int64, error) {
	c := v.(*models.Categoria)
	if _, ok := m.items[c.PK_ID_CATEGORIA]; !ok {
		return 0, orm.ErrNoRows
	}
	m.items[c.PK_ID_CATEGORIA] = *c
	return 1, nil
}
func (m *catMockOrm) Delete(v interface{}, _ ...string) (int64, error) {
	c := v.(*models.Categoria)
	if _, ok := m.items[c.PK_ID_CATEGORIA]; !ok {
		return 0, orm.ErrNoRows
	}
	delete(m.items, c.PK_ID_CATEGORIA)
	return 1, nil
}

func TestCategoriaController_FullCoverage(t *testing.T) {
	m := newCatMockOrm()
	m.qs = catQS{list: []models.Categoria{{PK_ID_CATEGORIA: 10, NOMBRE: "Bebidas"}}}
	orig := catOrmNew
	catOrmNew = func() categoriaOrmer { return m }
	defer func() { catOrmNew = orig }()

	// GetAll
	r := httptest.NewRequest(http.MethodGet, "/categorias", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := &CategoriaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.GetAll()
	if w.Code != http.StatusOK {
		t.Fatalf("GetAll status %d", w.Code)
	}

	// GetById not found
	r = httptest.NewRequest(http.MethodGet, "/categorias/search?id=99", nil)
	w = httptest.NewRecorder()
	ctx = context.NewContext()
	ctx.Reset(w, r)
	c = &CategoriaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.GetById()
	var resp models.ApiResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.Code)
	}

	// Post invalid
	r = httptest.NewRequest(http.MethodPost, "/categorias", strings.NewReader("bad"))
	w = httptest.NewRecorder()
	ctx = context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte("bad")
	c = &CategoriaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Post()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("post bad %d", w.Code)
	}

	// Post ok
	body := `{"nombre":"Postres"}`
	r = httptest.NewRequest(http.MethodPost, "/categorias", strings.NewReader(body))
	w = httptest.NewRecorder()
	ctx = context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c = &CategoriaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Post()
	if w.Code != http.StatusCreated {
		t.Fatalf("post ok %d", w.Code)
	}

	// Put not found
	r = httptest.NewRequest(http.MethodPut, "/categorias?id=99", strings.NewReader(`{"nombre":"Snacks"}`))
	w = httptest.NewRecorder()
	ctx = context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(`{"nombre":"Snacks"}`)
	c = &CategoriaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Put()
	if w.Code != http.StatusOK {
		t.Fatalf("put nf %d", w.Code)
	}

	// Put ok
	cat := models.Categoria{NOMBRE: "Cafe"}
	m.Insert(&cat)
	r = httptest.NewRequest(http.MethodPut, "/categorias?id=1", strings.NewReader(`{"nombre":"Café"}`))
	w = httptest.NewRecorder()
	ctx = context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(`{"nombre":"Café"}`)
	c = &CategoriaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Put()
	if w.Code != http.StatusOK {
		t.Fatalf("put ok %d", w.Code)
	}

	// GetById ok
	r = httptest.NewRequest(http.MethodGet, "/categorias/search?id=1", nil)
	w = httptest.NewRecorder()
	ctx = context.NewContext()
	ctx.Reset(w, r)
	c = &CategoriaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.GetById()
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != http.StatusOK {
		t.Fatalf("get id ok %d", resp.Code)
	}

	// Delete ok
	r = httptest.NewRequest(http.MethodDelete, "/categorias?id=1", nil)
	w = httptest.NewRecorder()
	ctx = context.NewContext()
	ctx.Reset(w, r)
	c = &CategoriaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Delete()
	if w.Code != http.StatusOK {
		t.Fatalf("del ok %d", w.Code)
	}
}

// Tipos de prueba para forzar errores en CategoriaController
type badQSCat struct{}

func (badQSCat) All(res interface{}, _ ...string) (int64, error) { return 0, orm.ErrNoRows }

type badOrmCat struct{ *catMockOrm }

func (b badOrmCat) QueryTable(_ interface{}) categoriaQuerySeter     { return badQSCat{} }
func (b badOrmCat) Insert(v interface{}) (int64, error)              { return 0, orm.ErrTxDone }
func (b badOrmCat) Delete(v interface{}, _ ...string) (int64, error) { return 0, orm.ErrNoRows }

func TestCategoriaController_AllError_InsertError_DeleteNotFound(t *testing.T) {
	orig := catOrmNew
	catOrmNew = func() categoriaOrmer { return badOrmCat{newCatMockOrm()} }
	defer func() { catOrmNew = orig }()

	// GetAll -> error
	r := httptest.NewRequest(http.MethodGet, "/categorias", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := &CategoriaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.GetAll()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}

	// Post -> insert error
	r = httptest.NewRequest(http.MethodPost, "/categorias", strings.NewReader(`{"nombre":"X"}`))
	w = httptest.NewRecorder()
	ctx = context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(`{"nombre":"X"}`)
	c = &CategoriaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Post()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}

	// Delete -> not found
	r = httptest.NewRequest(http.MethodDelete, "/categorias?id=1", nil)
	w = httptest.NewRecorder()
	ctx = context.NewContext()
	ctx.Reset(w, r)
	c = &CategoriaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Delete()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

// Ormer que falla en Update para probar la rama 500 en Put
type updErrOrm struct{ *catMockOrm }

func (m updErrOrm) QueryTable(_ interface{}) categoriaQuerySeter     { return m.catMockOrm.qs }
func (m updErrOrm) Insert(v interface{}) (int64, error)              { return m.catMockOrm.Insert(v) }
func (m updErrOrm) Read(v interface{}, _ ...string) error            { return nil }
func (m updErrOrm) Update(v interface{}, _ ...string) (int64, error) { return 0, orm.ErrTxDone }
func (m updErrOrm) Delete(v interface{}, _ ...string) (int64, error) { return 0, orm.ErrNoRows }

func TestCategoriaController_PutUpdateError(t *testing.T) {
	// Mock que lee ok y falla al actualizar

	orig := catOrmNew
	catOrmNew = func() categoriaOrmer { return updErrOrm{newCatMockOrm()} }
	defer func() { catOrmNew = orig }()

	body := `{"nombre":"Z"}`
	r := httptest.NewRequest(http.MethodPut, "/categorias?id=1", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := &CategoriaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Put()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}
