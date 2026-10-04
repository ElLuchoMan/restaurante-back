package subcategoria

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"restaurante/models"

	"github.com/beego/beego/v2/server/web/context"
	"github.com/stretchr/testify/assert"
)

func TestPost_JSONErrorWithLogging(t *testing.T) {

	r := httptest.NewRequest(http.MethodPost, "/subcategorias", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte("invalid json")

	c := &SubcategoriaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp models.ApiResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestPost_EmptyNombre(t *testing.T) {
	payload := map[string]interface{}{
		"nombre":      "",
		"categoriaId": 1,
	}
	body, _ := json.Marshal(payload)

	r := httptest.NewRequest(http.MethodPost, "/subcategorias", bytes.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = body

	c := &SubcategoriaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp models.ApiResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Contains(t, resp.Message, "requeridos")
}

func TestPost_ZeroCategoriaId(t *testing.T) {
	payload := map[string]interface{}{
		"nombre":      "Test Subcategoria",
		"categoriaId": 0,
	}
	body, _ := json.Marshal(payload)

	r := httptest.NewRequest(http.MethodPost, "/subcategorias", bytes.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = body

	c := &SubcategoriaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp models.ApiResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestPut_OnlyNombre(t *testing.T) {
	mockOrmer := &mockSubcategoriaOrmer{
		readFunc: func(v interface{}, cols ...string) error {
			if s, ok := v.(*models.Subcategoria); ok {
				s.NOMBRE = "Old Name"
				s.PK_ID_CATEGORIA = &models.Categoria{PK_ID_CATEGORIA: 1}
			}
			return nil
		},
		updateFunc: func(v interface{}, cols ...string) (int64, error) {

			if len(cols) != 1 || cols[0] != "NOMBRE" {
				t.Errorf("Expected only NOMBRE column, got %v", cols)
			}
			return 1, nil
		},
	}

	origOrmNew := subcatOrmNew
	defer func() { subcatOrmNew = origOrmNew }()
	subcatOrmNew = func() subcatOrmer { return mockOrmer }

	nombre := "New Name"
	payload := map[string]interface{}{
		"nombre": nombre,
	}
	body, _ := json.Marshal(payload)

	r := httptest.NewRequest(http.MethodPut, "/subcategorias?id=1", bytes.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = body

	c := &SubcategoriaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Put()

	assert.Equal(t, http.StatusOK, w.Code)
	var resp models.ApiResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestPut_OnlyCategoriaId(t *testing.T) {
	mockOrmer := &mockSubcategoriaOrmer{
		readFunc: func(v interface{}, cols ...string) error {
			if s, ok := v.(*models.Subcategoria); ok {
				s.NOMBRE = "Test Name"
				s.PK_ID_CATEGORIA = &models.Categoria{PK_ID_CATEGORIA: 1}
			}
			return nil
		},
		updateFunc: func(v interface{}, cols ...string) (int64, error) {

			if len(cols) != 1 || cols[0] != "PK_ID_CATEGORIA" {
				t.Errorf("Expected only PK_ID_CATEGORIA column, got %v", cols)
			}
			return 1, nil
		},
	}

	origOrmNew := subcatOrmNew
	defer func() { subcatOrmNew = origOrmNew }()
	subcatOrmNew = func() subcatOrmer { return mockOrmer }

	categoriaId := int64(2)
	payload := map[string]interface{}{
		"categoriaId": categoriaId,
	}
	body, _ := json.Marshal(payload)

	r := httptest.NewRequest(http.MethodPut, "/subcategorias?id=1", bytes.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = body

	c := &SubcategoriaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Put()

	assert.Equal(t, http.StatusOK, w.Code)
	var resp models.ApiResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestGetAll_WithInvalidCategoriaId(t *testing.T) {
	mockOrmer := &mockSubcategoriaOrmer{
		queryTableFunc: func(i interface{}) subcatQuerySeter {
			return &mockSubcategoriaQuerySeter{
				filterFunc: func(field string, args ...interface{}) subcatQuerySeter {
					t.Error("Filter should not be called with invalid categoria_id")
					return &mockSubcategoriaQuerySeter{}
				},
				allFunc: func(res interface{}, cols ...string) (int64, error) {
					return 0, nil
				},
			}
		},
	}

	origOrmNew := subcatOrmNew
	defer func() { subcatOrmNew = origOrmNew }()
	subcatOrmNew = func() subcatOrmer { return mockOrmer }

	r := httptest.NewRequest(http.MethodGet, "/subcategorias?categoria_id=invalid", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)

	c := &SubcategoriaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetAll()

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetAll_WithZeroCategoriaId(t *testing.T) {
	filterCalled := false

	mockOrmer := &mockSubcategoriaOrmer{
		queryTableFunc: func(i interface{}) subcatQuerySeter {
			return &mockSubcategoriaQuerySeter{
				filterFunc: func(field string, args ...interface{}) subcatQuerySeter {
					filterCalled = true
					return &mockSubcategoriaQuerySeter{}
				},
				allFunc: func(res interface{}, cols ...string) (int64, error) {
					return 0, nil
				},
			}
		},
	}

	origOrmNew := subcatOrmNew
	defer func() { subcatOrmNew = origOrmNew }()
	subcatOrmNew = func() subcatOrmer { return mockOrmer }

	r := httptest.NewRequest(http.MethodGet, "/subcategorias?categoria_id=0", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)

	c := &SubcategoriaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetAll()

	assert.False(t, filterCalled, "Filter should not be called when categoria_id is 0")
	assert.Equal(t, http.StatusOK, w.Code)
}

type mockSubcategoriaOrmer struct {
	queryTableFunc func(interface{}) subcatQuerySeter
	insertFunc     func(interface{}) (int64, error)
	readFunc       func(interface{}, ...string) error
	updateFunc     func(interface{}, ...string) (int64, error)
	deleteFunc     func(interface{}, ...string) (int64, error)
}

func (m *mockSubcategoriaOrmer) QueryTable(i interface{}) subcatQuerySeter {
	if m.queryTableFunc != nil {
		return m.queryTableFunc(i)
	}
	return &mockSubcategoriaQuerySeter{}
}

func (m *mockSubcategoriaOrmer) Insert(v interface{}) (int64, error) {
	if m.insertFunc != nil {
		return m.insertFunc(v)
	}
	return 1, nil
}

func (m *mockSubcategoriaOrmer) Read(v interface{}, cols ...string) error {
	if m.readFunc != nil {
		return m.readFunc(v, cols...)
	}
	return nil
}

func (m *mockSubcategoriaOrmer) Update(v interface{}, cols ...string) (int64, error) {
	if m.updateFunc != nil {
		return m.updateFunc(v, cols...)
	}
	return 1, nil
}

func (m *mockSubcategoriaOrmer) Delete(v interface{}, cols ...string) (int64, error) {
	if m.deleteFunc != nil {
		return m.deleteFunc(v, cols...)
	}
	return 1, nil
}

type mockSubcategoriaQuerySeter struct {
	allFunc    func(interface{}, ...string) (int64, error)
	filterFunc func(string, ...interface{}) subcatQuerySeter
}

func (m *mockSubcategoriaQuerySeter) All(res interface{}, cols ...string) (int64, error) {
	if m.allFunc != nil {
		return m.allFunc(res, cols...)
	}
	return 0, nil
}

func (m *mockSubcategoriaQuerySeter) Filter(field string, args ...interface{}) subcatQuerySeter {
	if m.filterFunc != nil {
		return m.filterFunc(field, args...)
	}
	return m
}
