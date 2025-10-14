package subcategoria

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/beego/beego/v2/server/web/context"
)

type ormFailRead struct{}

func (ormFailRead) QueryTable(interface{}) subcatQuerySeter { return nil }
func (ormFailRead) Insert(interface{}) (int64, error)       { return 0, nil }
func (ormFailRead) Read(interface{}, ...string) error {
	return fmt.Errorf("database connection failed")
}
func (ormFailRead) Update(interface{}, ...string) (int64, error) { return 0, nil }
func (ormFailRead) Delete(interface{}, ...string) (int64, error) { return 0, nil }

func TestSubcategoriaController_GetById_DBError(t *testing.T) {
	orig := subcatOrmNew
	subcatOrmNew = func() subcatOrmer { return ormFailRead{} }
	defer func() { subcatOrmNew = orig }()

	r := httptest.NewRequest(http.MethodGet, "/subcategorias/search?id=999", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)

	c := &SubcategoriaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.GetById()

	if w.Code != http.StatusOK {
		t.Errorf("Expected HTTP status %d, got %d", http.StatusOK, w.Code)
	}
}
