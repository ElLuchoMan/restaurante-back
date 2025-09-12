package subcategoria

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beego/beego/v2/server/web/context"
)

type errAllQS struct{}

func (errAllQS) All(res interface{}, cols ...string) (int64, error) { return 0, errors.New("db") }
func (errAllQS) Filter(string, ...interface{}) subcatQuerySeter     { return errAllQS{} }

type getAllErrOrm struct{}

func (getAllErrOrm) QueryTable(interface{}) subcatQuerySeter      { return errAllQS{} }
func (getAllErrOrm) Insert(interface{}) (int64, error)            { return 0, nil }
func (getAllErrOrm) Read(interface{}, ...string) error            { return nil }
func (getAllErrOrm) Update(interface{}, ...string) (int64, error) { return 0, nil }
func (getAllErrOrm) Delete(interface{}, ...string) (int64, error) { return 0, nil }

type insertErrOrm struct{}

func (insertErrOrm) QueryTable(interface{}) subcatQuerySeter      { return errAllQS{} }
func (insertErrOrm) Insert(interface{}) (int64, error)            { return 0, errors.New("db") }
func (insertErrOrm) Read(interface{}, ...string) error            { return nil }
func (insertErrOrm) Update(interface{}, ...string) (int64, error) { return 0, nil }
func (insertErrOrm) Delete(interface{}, ...string) (int64, error) { return 0, nil }

func TestSubcategoriaController_GetAll_AllError(t *testing.T) {
	orig := subcatOrmNew
	subcatOrmNew = func() subcatOrmer { return getAllErrOrm{} }
	t.Cleanup(func() { subcatOrmNew = orig })

	r := httptest.NewRequest(http.MethodGet, "/subcategorias", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := &SubcategoriaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.GetAll()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestSubcategoriaController_Post_InsertError_DB(t *testing.T) {
	orig := subcatOrmNew
	subcatOrmNew = func() subcatOrmer { return insertErrOrm{} }
	t.Cleanup(func() { subcatOrmNew = orig })

	body := `{"nombre":"T","categoriaId":1}`
	r := httptest.NewRequest(http.MethodPost, "/subcategorias", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := &SubcategoriaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Post()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}
