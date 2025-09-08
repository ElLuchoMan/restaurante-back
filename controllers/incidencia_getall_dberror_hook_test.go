package controllers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
)

type fakeIncidQSAllErr struct{ orm.QuerySeter }

func (fakeIncidQSAllErr) All(interface{}, ...string) (int64, error) { return 0, errors.New("db") }

type fakeIncidOrmAllErr struct{}

func (fakeIncidOrmAllErr) QueryTable(interface{}) orm.QuerySeter        { return fakeIncidQSAllErr{} }
func (fakeIncidOrmAllErr) Insert(interface{}) (int64, error)            { return 0, nil }
func (fakeIncidOrmAllErr) Read(interface{}, ...string) error            { return nil }
func (fakeIncidOrmAllErr) Update(interface{}, ...string) (int64, error) { return 0, nil }
func (fakeIncidOrmAllErr) Delete(interface{}, ...string) (int64, error) { return 0, nil }

func TestIncidenciaGetAll_DBError_Hook(t *testing.T) {
	orig := incidenciaOrmNew
	incidenciaOrmNew = func() incidenciaOrmer { return fakeIncidOrmAllErr{} }
	t.Cleanup(func() { incidenciaOrmNew = orig })

	r := httptest.NewRequest(http.MethodGet, "/incidencias", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := IncidenciaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.GetAll()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}
