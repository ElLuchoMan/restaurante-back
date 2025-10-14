package subcategoria

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"restaurante/models"
	"strings"
	"testing"

	"github.com/beego/beego/v2/server/web/context"
)

func TestSubcategoriaController_Post_EmptyNombre(t *testing.T) {
	m := newSubMockOrm()
	orig := subcatOrmNew
	subcatOrmNew = func() subcatOrmer { return m }
	defer func() { subcatOrmNew = orig }()

	body := `{"nombre":"","categoriaId":1}`
	r := httptest.NewRequest(http.MethodPost, "/subcategorias", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)

	c := &SubcategoriaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Post()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var resp models.ApiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected response code %d, got %d", http.StatusBadRequest, resp.Code)
	}
}

func TestSubcategoriaController_Post_ZeroCategoriaId(t *testing.T) {
	m := newSubMockOrm()
	orig := subcatOrmNew
	subcatOrmNew = func() subcatOrmer { return m }
	defer func() { subcatOrmNew = orig }()

	body := `{"nombre":"Test","categoriaId":0}`
	r := httptest.NewRequest(http.MethodPost, "/subcategorias", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)

	c := &SubcategoriaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Post()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var resp models.ApiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected response code %d, got %d", http.StatusBadRequest, resp.Code)
	}
}

func TestSubcategoriaController_Post_MissingCategoriaId(t *testing.T) {
	m := newSubMockOrm()
	orig := subcatOrmNew
	subcatOrmNew = func() subcatOrmer { return m }
	defer func() { subcatOrmNew = orig }()

	body := `{"nombre":"Test"}`
	r := httptest.NewRequest(http.MethodPost, "/subcategorias", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)

	c := &SubcategoriaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Post()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var resp models.ApiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected response code %d, got %d", http.StatusBadRequest, resp.Code)
	}
}
