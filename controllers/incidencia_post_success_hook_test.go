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

type fakeIncidOrmInsertOK struct{}

func (fakeIncidOrmInsertOK) QueryTable(interface{}) orm.QuerySeter { return nil }
func (fakeIncidOrmInsertOK) Insert(v interface{}) (int64, error) {
	if inc, ok := v.(*models.Incidencia); ok {
		inc.PK_ID_INCIDENCIA = 123
	}
	return 1, nil
}
func (fakeIncidOrmInsertOK) Read(interface{}, ...string) error            { return nil }
func (fakeIncidOrmInsertOK) Update(interface{}, ...string) (int64, error) { return 1, nil }
func (fakeIncidOrmInsertOK) Delete(interface{}, ...string) (int64, error) { return 1, nil }

func TestIncidenciaPostSuccess_Hook(t *testing.T) {
	orig := incidenciaOrmNew
	incidenciaOrmNew = func() incidenciaOrmer { return fakeIncidOrmInsertOK{} }
	t.Cleanup(func() { incidenciaOrmNew = orig })

	body := `{"fechaIncidencia":"2024-01-01","monto":100,"resta":false,"motivo":"ok","documentoTrabajador":1}`
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
		t.Fatalf("expected 201, got %d. Body: %s", w.Code, w.Body.String())
	}
	var resp models.ApiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Code != http.StatusCreated {
		t.Fatalf("expected code 201, got %d", resp.Code)
	}
}
