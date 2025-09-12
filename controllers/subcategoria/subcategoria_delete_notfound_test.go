package subcategoria

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
)

type nfSubOrm struct{}

func (nfSubOrm) QueryTable(interface{}) subcatQuerySeter      { return subQS{} }
func (nfSubOrm) Insert(interface{}) (int64, error)            { return 0, nil }
func (nfSubOrm) Read(interface{}, ...string) error            { return nil }
func (nfSubOrm) Update(interface{}, ...string) (int64, error) { return 0, nil }
func (nfSubOrm) Delete(interface{}, ...string) (int64, error) { return 0, orm.ErrNoRows }

func TestSubcategoriaDeleteNotFound(t *testing.T) {
	orig := subcatOrmNew
	subcatOrmNew = func() subcatOrmer { return nfSubOrm{} }
	t.Cleanup(func() { subcatOrmNew = orig })

	r := httptest.NewRequest(http.MethodDelete, "/subcategorias?id=99", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := &SubcategoriaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Delete()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
