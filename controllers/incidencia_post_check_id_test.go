package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
)

type fakeIncidOrmInsertWithID struct{}

func (fakeIncidOrmInsertWithID) QueryTable(interface{}) orm.QuerySeter { return nil }
func (fakeIncidOrmInsertWithID) Insert(v interface{}) (int64, error) {
	if inc, ok := v.(*models.Incidencia); ok {
		inc.PK_ID_INCIDENCIA = 999
	}
	return 1, nil
}
func (fakeIncidOrmInsertWithID) Read(interface{}, ...string) error            { return nil }
func (fakeIncidOrmInsertWithID) Update(interface{}, ...string) (int64, error) { return 1, nil }
func (fakeIncidOrmInsertWithID) Delete(interface{}, ...string) (int64, error) { return 1, nil }

func TestIncidenciaPost_ResponseContainsId(t *testing.T) {
	orig := incidenciaOrmNew
	incidenciaOrmNew = func() incidenciaOrmer { return fakeIncidOrmInsertWithID{} }
	t.Cleanup(func() { incidenciaOrmNew = orig })

	body := `{"fechaIncidencia":"2024-07-07","monto":10,"resta":false,"motivo":"ok","documentoTrabajador":1}`
	r := httptest.NewRequest(http.MethodPost, "/incidencias", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := IncidenciaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	var resp models.ApiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	m := resp.Data.(map[string]interface{})
	if m["incidenciaId"].(float64) == 0 {
		t.Fatalf("expected incidenciaId set, got 0")
	}
}
