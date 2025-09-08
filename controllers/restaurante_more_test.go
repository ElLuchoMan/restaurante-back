package controllers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
)

type nfRestOrm struct{}

func (nfRestOrm) QueryTable(interface{}) restQuerySeter        { return restQSAdapter{qs: nil} }
func (nfRestOrm) Read(interface{}, ...string) error            { return orm.ErrNoRows }
func (nfRestOrm) Insert(interface{}) (int64, error)            { return 0, nil }
func (nfRestOrm) Update(interface{}, ...string) (int64, error) { return 0, nil }
func (nfRestOrm) Delete(interface{}, ...string) (int64, error) { return 0, nil }

type okRestOrm struct{}

func (okRestOrm) QueryTable(interface{}) restQuerySeter        { return restQSAdapter{qs: nil} }
func (okRestOrm) Read(interface{}, ...string) error            { return nil }
func (okRestOrm) Insert(interface{}) (int64, error)            { return 1, nil }
func (okRestOrm) Update(interface{}, ...string) (int64, error) { return 1, nil }
func (okRestOrm) Delete(interface{}, ...string) (int64, error) { return 1, nil }

func TestRestauranteGetById_NotFound(t *testing.T) {
	orig := restOrmNew
	restOrmNew = func() restOrmer { return nfRestOrm{} }
	t.Cleanup(func() { restOrmNew = orig })

	r := httptest.NewRequest(http.MethodGet, "/restaurante/v1/restaurantes/search?id=1", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := &RestauranteController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.GetById()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestRestaurantePut_InvalidJSON(t *testing.T) {
	orig := restOrmNew
	restOrmNew = func() restOrmer { return okRestOrm{} }
	t.Cleanup(func() { restOrmNew = orig })

	r := httptest.NewRequest(http.MethodPut, "/restaurante/v1/restaurantes?id=1", strings.NewReader("{"))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte("{")
	c := &RestauranteController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Put()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
