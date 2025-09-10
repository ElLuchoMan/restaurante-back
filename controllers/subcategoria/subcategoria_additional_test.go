package subcategoria

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

type fakeSubQS struct{ orm.QuerySeter }

func (f fakeSubQS) All(res interface{}, cols ...string) (int64, error) { return 0, nil }
func (f fakeSubQS) Filter(string, ...interface{}) subcatQuerySeter     { return f }

type fakeSubOrm struct {
	readErr error
	delErr  error
}

func (f fakeSubOrm) QueryTable(interface{}) subcatQuerySeter      { return fakeSubQS{} }
func (f fakeSubOrm) Insert(interface{}) (int64, error)            { return 0, nil }
func (f fakeSubOrm) Read(interface{}, ...string) error            { return f.readErr }
func (f fakeSubOrm) Update(interface{}, ...string) (int64, error) { return 1, nil }
func (f fakeSubOrm) Delete(interface{}, ...string) (int64, error) { return 1, f.delErr }

func TestSubcategoria_GetById_DBError(t *testing.T) {
	orig := subcatOrmNew
	subcatOrmNew = func() subcatOrmer { return fakeSubOrm{readErr: errors.New("db")} }
	t.Cleanup(func() { subcatOrmNew = orig })

	r := httptest.NewRequest(http.MethodGet, "/subcategorias/search?id=1", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := &SubcategoriaController{}
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
		t.Fatalf("esperaba 404 code, got %d", resp.Code)
	}
}

func TestSubcategoria_GetById_Success(t *testing.T) {
	orig := subcatOrmNew
	subcatOrmNew = func() subcatOrmer { return fakeSubOrm{readErr: nil} }
	t.Cleanup(func() { subcatOrmNew = orig })

	r := httptest.NewRequest(http.MethodGet, "/subcategorias/search?id=2", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := &SubcategoriaController{}
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
		t.Fatalf("esperaba 200 code, got %d", resp.Code)
	}
}
