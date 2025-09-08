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

type fakeIncidOrmInsertOK3 struct{}

func (fakeIncidOrmInsertOK3) QueryTable(interface{}) orm.QuerySeter        { return nil }
func (fakeIncidOrmInsertOK3) Insert(v interface{}) (int64, error)          { return 1, nil }
func (fakeIncidOrmInsertOK3) Read(interface{}, ...string) error            { return nil }
func (fakeIncidOrmInsertOK3) Update(interface{}, ...string) (int64, error) { return 1, nil }
func (fakeIncidOrmInsertOK3) Delete(interface{}, ...string) (int64, error) { return 1, nil }

func TestIncidenciaPost_SetsResponseFields(t *testing.T) {
	orig := incidenciaOrmNew
	incidenciaOrmNew = func() incidenciaOrmer { return fakeIncidOrmInsertOK3{} }
	t.Cleanup(func() { incidenciaOrmNew = orig })

	motivo := strings.Repeat("m", 50)
	body := `{"fechaIncidencia":"2024-06-06","monto":150,"resta":false,"motivo":"` + motivo + `","documentoTrabajador":10}`
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
	if resp.Code != http.StatusCreated {
		t.Fatalf("expected code 201, got %d", resp.Code)
	}
}
