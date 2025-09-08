package controllers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
)

type fakeIncidQS struct{ orm.QuerySeter }

func (fakeIncidQS) Filter(string, ...interface{}) orm.QuerySeter    { return fakeIncidQS{} }
func (fakeIncidQS) All(res interface{}, _ ...string) (int64, error) { return 0, nil }

type fakeIncidOrmQuery struct{}

func (fakeIncidOrmQuery) QueryTable(interface{}) orm.QuerySeter        { return fakeIncidQS{} }
func (fakeIncidOrmQuery) Insert(interface{}) (int64, error)            { return 0, nil }
func (fakeIncidOrmQuery) Read(interface{}, ...string) error            { return nil }
func (fakeIncidOrmQuery) Update(interface{}, ...string) (int64, error) { return 0, nil }
func (fakeIncidOrmQuery) Delete(interface{}, ...string) (int64, error) { return 0, nil }

func TestIncidenciaGetByDocumentAndDate_HookSuccessNoRows(t *testing.T) {
	orig := incidenciaOrmNew
	incidenciaOrmNew = func() incidenciaOrmer { return fakeIncidOrmQuery{} }
	t.Cleanup(func() { incidenciaOrmNew = orig })

	r := httptest.NewRequest(http.MethodGet, "/incidencias/search?documento=1&mes=1&anio=2024", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := IncidenciaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.GetByDocumentAndDate()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
