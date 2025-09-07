package controllers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
)

type fakeIncidOrmPartial struct{}
func (fakeIncidOrmPartial) QueryTable(interface{}) orm.QuerySeter { return nil }
func (fakeIncidOrmPartial) Insert(interface{}) (int64, error) { return 0, nil }
func (fakeIncidOrmPartial) Read(interface{}, ...string) error { return nil }
func (fakeIncidOrmPartial) Update(interface{}, ...string) (int64, error) { return 1, nil }
func (fakeIncidOrmPartial) Delete(interface{}, ...string) (int64, error) { return 0, nil }

func TestIncidenciaPut_PartialMotivo(t *testing.T) {
    orig := incidenciaOrmNew
    incidenciaOrmNew = func() incidenciaOrmer { return fakeIncidOrmPartial{} }
    t.Cleanup(func(){ incidenciaOrmNew = orig })

    body := `{"motivo":"nuevo"}`
    r := httptest.NewRequest(http.MethodPut, "/incidencias?id=1", strings.NewReader(body))
    w := httptest.NewRecorder(); ctx := context.NewContext(); ctx.Reset(w, r); ctx.Input.RequestBody = []byte(body)
    c := IncidenciaController{}; c.Ctx = ctx; c.Data = make(map[interface{}]interface{})

    c.Put()
    if w.Code != http.StatusOK { t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String()) }
}


