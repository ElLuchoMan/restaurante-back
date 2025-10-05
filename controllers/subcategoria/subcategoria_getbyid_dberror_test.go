package subcategoria

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"restaurante/models"
	"testing"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
)

// mockOrmGetByIdDBError es un mock que retorna error de DB en Read
type mockOrmGetByIdDBError struct{}

func (mockOrmGetByIdDBError) QueryTable(_ interface{}) subcatQuerySeter {
	return badQSSub{}
}
func (mockOrmGetByIdDBError) Insert(interface{}) (int64, error) { return 0, nil }
func (mockOrmGetByIdDBError) Read(interface{}, ...string) error {
	return errors.New("database connection error")
}
func (mockOrmGetByIdDBError) Update(interface{}, ...string) (int64, error) { return 0, nil }
func (mockOrmGetByIdDBError) Delete(interface{}, ...string) (int64, error) { return 0, nil }

// TestSubcategoriaController_GetById_DBErrorNotNoRows verifica el manejo de errores de DB distintos a NoRows
func TestSubcategoriaController_GetById_DBErrorNotNoRows(t *testing.T) {
	orig := subcatOrmNew
	subcatOrmNew = func() subcatOrmer { return mockOrmGetByIdDBError{} }
	defer func() { subcatOrmNew = orig }()

	r := httptest.NewRequest(http.MethodGet, "/subcategorias/search?id=5", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)

	c := &SubcategoriaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.GetById()

	// Debería retornar 404 de todas formas, pero el error se loguea
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp models.ApiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Code != http.StatusNotFound {
		t.Errorf("Expected response code %d, got %d", http.StatusNotFound, resp.Code)
	}
}

// badQSSubForGetById es un QuerySeter que retorna error
type badQSSubForGetById struct{}

func (badQSSubForGetById) All(res interface{}, _ ...string) (int64, error) {
	return 0, orm.ErrNoRows
}
func (badQSSubForGetById) Filter(_ string, _ ...interface{}) subcatQuerySeter {
	return badQSSubForGetById{}
}
