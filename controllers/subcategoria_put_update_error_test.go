package controllers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
)

// Orm wrapper que falla en Update pero delega el resto a subMockOrm
type updFailOrm struct{ base *subMockOrm }

func (u updFailOrm) QueryTable(i interface{}) subcatQuerySeter           { return u.base.QueryTable(i) }
func (u updFailOrm) Insert(v interface{}) (int64, error)                 { return u.base.Insert(v) }
func (u updFailOrm) Read(v interface{}, cols ...string) error            { return u.base.Read(v, cols...) }
func (u updFailOrm) Update(v interface{}, cols ...string) (int64, error) { return 0, orm.ErrTxDone }
func (u updFailOrm) Delete(v interface{}, cols ...string) (int64, error) {
	return u.base.Delete(v, cols...)
}

// Cubre la rama de error en Update dentro de Put de SubcategoriaController
func TestSubcategoriaController_Put_UpdateFails(t *testing.T) {
	s := newSubMockOrm()
	// Insertar uno para que Read sea OK
	s.Insert(&models.Subcategoria{NOMBRE: "X"})

	orig := subcatOrmNew
	subcatOrmNew = func() subcatOrmer { return updFailOrm{base: s} }
	defer func() { subcatOrmNew = orig }()

	r := httptest.NewRequest(http.MethodPut, "/subcategorias?id=1", strings.NewReader(`{"nombre":"Y"}`))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(`{"nombre":"Y"}`)
	c := &SubcategoriaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Put()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}
