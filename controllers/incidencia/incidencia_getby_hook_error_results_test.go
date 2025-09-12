package incidencia

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
)

type fakeIncidQSErrorRes struct{ orm.QuerySeter }

func (fakeIncidQSErrorRes) Filter(string, ...interface{}) orm.QuerySeter {
	return fakeIncidQSErrorRes{}
}
func (fakeIncidQSErrorRes) All(res interface{}, _ ...string) (int64, error) {
	dst := res.(*[]models.Incidencia)
	*dst = append(*dst, models.Incidencia{MOTIVO: "x"})
	return 1, errors.New("db")
}

type fakeIncidOrmErrRes struct{}

func (fakeIncidOrmErrRes) QueryTable(interface{}) orm.QuerySeter        { return fakeIncidQSErrorRes{} }
func (fakeIncidOrmErrRes) Insert(interface{}) (int64, error)            { return 0, nil }
func (fakeIncidOrmErrRes) Read(interface{}, ...string) error            { return nil }
func (fakeIncidOrmErrRes) Update(interface{}, ...string) (int64, error) { return 0, nil }
func (fakeIncidOrmErrRes) Delete(interface{}, ...string) (int64, error) { return 0, nil }

func TestIncidenciaGetByDocumentAndDate_HookErrWithResults(t *testing.T) {
	orig := incidenciaOrmNew
	incidenciaOrmNew = func() incidenciaOrmer { return fakeIncidOrmErrRes{} }
	t.Cleanup(func() { incidenciaOrmNew = orig })

	r := httptest.NewRequest(http.MethodGet, "/incidencias/search?documento=1&mes=1&anio=2024", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := IncidenciaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetByDocumentAndDate()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}
