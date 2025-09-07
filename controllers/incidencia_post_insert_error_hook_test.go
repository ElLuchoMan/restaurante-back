package controllers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
)

type fakeIncidOrmInsertFail struct{}
func (fakeIncidOrmInsertFail) QueryTable(interface{}) orm.QuerySeter { return nil }
func (fakeIncidOrmInsertFail) Insert(interface{}) (int64, error) { return 0, orm.ErrTxDone }
func (fakeIncidOrmInsertFail) Read(interface{}, ...string) error { return nil }
func (fakeIncidOrmInsertFail) Update(interface{}, ...string) (int64, error) { return 0, nil }
func (fakeIncidOrmInsertFail) Delete(interface{}, ...string) (int64, error) { return 0, nil }

func TestIncidenciaPost_InsertError_Hook(t *testing.T) {
    orig := incidenciaOrmNew
    incidenciaOrmNew = func() incidenciaOrmer { return fakeIncidOrmInsertFail{} }
    t.Cleanup(func(){ incidenciaOrmNew = orig })

    body := `{"fechaIncidencia":"2024-01-01","monto":100,"resta":false,"motivo":"x","documentoTrabajador":1}`
    r := httptest.NewRequest(http.MethodPost, "/incidencias", strings.NewReader(body))
    w := httptest.NewRecorder(); ctx := context.NewContext(); ctx.Reset(w, r); ctx.Input.RequestBody = []byte(body)
    c := IncidenciaController{}; c.Ctx = ctx; c.Data = make(map[interface{}]interface{})

    c.Post()
    if w.Code != http.StatusInternalServerError { t.Fatalf("expected 500, got %d", w.Code) }
}


