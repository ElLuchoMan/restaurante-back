package incidencia

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
)

type fakeIncidOrmNotFound struct{}

func (fakeIncidOrmNotFound) QueryTable(interface{}) orm.QuerySeter        { return nil }
func (fakeIncidOrmNotFound) Insert(interface{}) (int64, error)            { return 0, nil }
func (fakeIncidOrmNotFound) Read(interface{}, ...string) error            { return nil }
func (fakeIncidOrmNotFound) Update(interface{}, ...string) (int64, error) { return 0, nil }
func (fakeIncidOrmNotFound) Delete(interface{}, ...string) (int64, error) { return 0, orm.ErrNoRows }

func TestIncidenciaDeleteNotFound_Hook(t *testing.T) {
	orig := incidenciaOrmNew
	incidenciaOrmNew = func() incidenciaOrmer { return fakeIncidOrmNotFound{} }
	t.Cleanup(func() { incidenciaOrmNew = orig })

	r := httptest.NewRequest(http.MethodDelete, "/incidencias?id=1", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := IncidenciaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Delete()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 wrapper, got %d", w.Code)
	}
}
