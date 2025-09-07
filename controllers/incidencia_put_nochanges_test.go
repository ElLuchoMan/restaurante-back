package controllers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
)

// Fake ormer to simulate Read OK and Update returning zero affected rows (no-op)
type fakeIncidOrmNoChange struct{}
func (fakeIncidOrmNoChange) QueryTable(interface{}) orm.QuerySeter { return nil }
func (fakeIncidOrmNoChange) Insert(interface{}) (int64, error) { return 0, nil }
func (fakeIncidOrmNoChange) Read(v interface{}, _ ...string) error { return nil }
func (fakeIncidOrmNoChange) Update(v interface{}, _ ...string) (int64, error) { return 1, nil }
func (fakeIncidOrmNoChange) Delete(interface{}, ...string) (int64, error) { return 1, nil }

func TestIncidenciaPut_NoChanges_AcceptsOK(t *testing.T) {
    orig := incidenciaOrmNew
    incidenciaOrmNew = func() incidenciaOrmer { return fakeIncidOrmNoChange{} }
    t.Cleanup(func(){ incidenciaOrmNew = orig })

    body := `{}`
    r := httptest.NewRequest(http.MethodPut, "/incidencias?id=1", strings.NewReader(body))
    w := httptest.NewRecorder(); ctx := context.NewContext(); ctx.Reset(w, r); ctx.Input.RequestBody = []byte(body)
    c := IncidenciaController{}; c.Ctx = ctx; c.Data = make(map[interface{}]interface{})

    c.Put()
    if w.Code != http.StatusOK { t.Fatalf("expected 200, got %d", w.Code) }
}


