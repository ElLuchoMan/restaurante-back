package categoria

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
)

// fake implementations

type fakeCatQS struct{ orm.QuerySeter }

func (f fakeCatQS) All(res interface{}, cols ...string) (int64, error) { return 0, nil }

type fakeCatOrm struct {
	readErr error
	delErr  error
}

func (f fakeCatOrm) QueryTable(interface{}) categoriaQuerySeter   { return fakeCatQS{} }
func (f fakeCatOrm) Insert(interface{}) (int64, error)            { return 0, nil }
func (f fakeCatOrm) Read(v interface{}, cols ...string) error     { return f.readErr }
func (f fakeCatOrm) Update(interface{}, ...string) (int64, error) { return 1, nil }
func (f fakeCatOrm) Delete(interface{}, ...string) (int64, error) { return 1, f.delErr }

func TestCategoria_GetById_BadRequest(t *testing.T) {
	orig := catOrmNew
	catOrmNew = func() categoriaOrmer { return fakeCatOrm{} }
	t.Cleanup(func() { catOrmNew = orig })

	r := httptest.NewRequest(http.MethodGet, "/categorias/search", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := &CategoriaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetById()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("esperaba 400, got %d", w.Code)
	}
}

func TestCategoria_GetById_NotFound(t *testing.T) {
	orig := catOrmNew
	catOrmNew = func() categoriaOrmer { return fakeCatOrm{readErr: errors.New("nf")} }
	t.Cleanup(func() { catOrmNew = orig })

	r := httptest.NewRequest(http.MethodGet, "/categorias/search?id=1", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := &CategoriaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetById()
	if w.Code != http.StatusOK {
		t.Fatalf("esperaba 200, got %d", w.Code)
	}
	var resp models.ApiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json: %v", err)
	}
	if resp.Code != http.StatusNotFound {
		t.Fatalf("esperaba ApiCode 404, got %d", resp.Code)
	}
}

func TestCategoria_GetById_Success(t *testing.T) {
	orig := catOrmNew
	catOrmNew = func() categoriaOrmer { return fakeCatOrm{readErr: nil} }
	t.Cleanup(func() { catOrmNew = orig })

	r := httptest.NewRequest(http.MethodGet, "/categorias/search?id=2", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := &CategoriaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetById()
	if w.Code != http.StatusOK {
		t.Fatalf("esperaba 200, got %d", w.Code)
	}
	var resp models.ApiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json: %v", err)
	}
	if resp.Code != http.StatusOK {
		t.Fatalf("esperaba ApiCode 200, got %d", resp.Code)
	}
}

func TestCategoria_Delete_BadRequest(t *testing.T) {
	orig := catOrmNew
	catOrmNew = func() categoriaOrmer { return fakeCatOrm{} }
	t.Cleanup(func() { catOrmNew = orig })

	r := httptest.NewRequest(http.MethodDelete, "/categorias", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := &CategoriaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Delete()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("esperaba 400, got %d", w.Code)
	}
}

func TestCategoria_Delete_NotFound(t *testing.T) {
	orig := catOrmNew
	catOrmNew = func() categoriaOrmer { return fakeCatOrm{delErr: errors.New("nf")} }
	t.Cleanup(func() { catOrmNew = orig })

	r := httptest.NewRequest(http.MethodDelete, "/categorias?id=1", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := &CategoriaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Delete()
	if w.Code != http.StatusOK {
		t.Fatalf("esperaba 200, got %d", w.Code)
	}
	var resp models.ApiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json: %v", err)
	}
	if resp.Code != http.StatusNotFound {
		t.Fatalf("esperaba ApiCode 404, got %d", resp.Code)
	}
}

func TestCategoria_Delete_Success(t *testing.T) {
	orig := catOrmNew
	catOrmNew = func() categoriaOrmer { return fakeCatOrm{} }
	t.Cleanup(func() { catOrmNew = orig })

	r := httptest.NewRequest(http.MethodDelete, "/categorias?id=2", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := &CategoriaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Delete()
	if w.Code != http.StatusOK {
		t.Fatalf("esperaba 200, got %d", w.Code)
	}
	var resp models.ApiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json: %v", err)
	}
	if resp.Code != http.StatusOK {
		t.Fatalf("esperaba ApiCode 200, got %d", resp.Code)
	}
}
