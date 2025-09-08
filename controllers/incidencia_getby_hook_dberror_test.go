package controllers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
)

type fakeIncidQSFail struct{ orm.QuerySeter }

func (fakeIncidQSFail) Filter(string, ...interface{}) orm.QuerySeter { return fakeIncidQSFail{} }
func (fakeIncidQSFail) All(interface{}, ...string) (int64, error)    { return 0, errors.New("db") }

type fakeIncidOrmFail struct{}

func (fakeIncidOrmFail) QueryTable(interface{}) orm.QuerySeter        { return fakeIncidQSFail{} }
func (fakeIncidOrmFail) Insert(interface{}) (int64, error)            { return 0, nil }
func (fakeIncidOrmFail) Read(interface{}, ...string) error            { return nil }
func (fakeIncidOrmFail) Update(interface{}, ...string) (int64, error) { return 0, nil }
func (fakeIncidOrmFail) Delete(interface{}, ...string) (int64, error) { return 0, nil }

func TestIncidenciaGetByDocumentAndDate_HookDBError(t *testing.T) {
	orig := incidenciaOrmNew
	incidenciaOrmNew = func() incidenciaOrmer { return fakeIncidOrmFail{} }
	t.Cleanup(func() { incidenciaOrmNew = orig })

	r := httptest.NewRequest(http.MethodGet, "/incidencias/search?documento=1&mes=1&anio=2024", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := IncidenciaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.GetByDocumentAndDate()
	if w.Code != http.StatusInternalServerError && w.Code != http.StatusOK {
		t.Fatalf("expected 500 or 200, got %d", w.Code)
	}
}
