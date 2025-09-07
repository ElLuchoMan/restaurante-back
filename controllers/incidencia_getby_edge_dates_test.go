package controllers

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
)

type fakeIncidQSEdge struct{ orm.QuerySeter }
func (fakeIncidQSEdge) Filter(string, ...interface{}) orm.QuerySeter { return fakeIncidQSEdge{} }
func (fakeIncidQSEdge) All(res interface{}, _ ...string) (int64, error) {
    dst := res.(*[]models.Incidencia)
    *dst = append(*dst, models.Incidencia{MOTIVO: "edge"})
    return 1, nil
}

type fakeIncidOrmEdge struct{}
func (fakeIncidOrmEdge) QueryTable(interface{}) orm.QuerySeter { return fakeIncidQSEdge{} }
func (fakeIncidOrmEdge) Insert(interface{}) (int64, error) { return 0, nil }
func (fakeIncidOrmEdge) Read(interface{}, ...string) error { return nil }
func (fakeIncidOrmEdge) Update(interface{}, ...string) (int64, error) { return 0, nil }
func (fakeIncidOrmEdge) Delete(interface{}, ...string) (int64, error) { return 0, nil }

func TestIncidenciaGetByDocumentAndDate_EdgeMonthYear(t *testing.T) {
    orig := incidenciaOrmNew
    incidenciaOrmNew = func() incidenciaOrmer { return fakeIncidOrmEdge{} }
    t.Cleanup(func(){ incidenciaOrmNew = orig })

    year := time.Now().Year()
    r := httptest.NewRequest(http.MethodGet, "/incidencias/search?documento=1&mes=12&anio="+strconv.Itoa(year), nil)
    w := httptest.NewRecorder(); ctx := context.NewContext(); ctx.Reset(w, r)
    c := IncidenciaController{}; c.Ctx = ctx; c.Data = make(map[interface{}]interface{})

    c.GetByDocumentAndDate()
    if w.Code != http.StatusOK { t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String()) }
}


