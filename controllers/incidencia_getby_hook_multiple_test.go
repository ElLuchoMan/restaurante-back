package controllers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
)

type fakeIncidQSMulti struct{ orm.QuerySeter }

func (fakeIncidQSMulti) Filter(string, ...interface{}) orm.QuerySeter { return fakeIncidQSMulti{} }
func (fakeIncidQSMulti) All(res interface{}, _ ...string) (int64, error) {
	dst := res.(*[]models.Incidencia)
	*dst = append(*dst, models.Incidencia{MOTIVO: "a"}, models.Incidencia{MOTIVO: "b"})
	return 2, nil
}

type fakeIncidOrmMulti struct{}

func (fakeIncidOrmMulti) QueryTable(interface{}) orm.QuerySeter        { return fakeIncidQSMulti{} }
func (fakeIncidOrmMulti) Insert(interface{}) (int64, error)            { return 0, nil }
func (fakeIncidOrmMulti) Read(interface{}, ...string) error            { return nil }
func (fakeIncidOrmMulti) Update(interface{}, ...string) (int64, error) { return 0, nil }
func (fakeIncidOrmMulti) Delete(interface{}, ...string) (int64, error) { return 0, nil }

func TestIncidenciaGetByDocumentAndDate_HookMultipleRows(t *testing.T) {
	orig := incidenciaOrmNew
	incidenciaOrmNew = func() incidenciaOrmer { return fakeIncidOrmMulti{} }
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
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}
}
