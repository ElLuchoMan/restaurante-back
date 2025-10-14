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

func TestSubcategoriaController_Put_NoColumnsToUpdate(t *testing.T) {
	m := newSubMockOrm()

	sub := models.Subcategoria{NOMBRE: "Existing"}
	m.Insert(&sub)

	orig := subcatOrmNew
	subcatOrmNew = func() subcatOrmer { return m }
	defer func() { subcatOrmNew = orig }()

	body := `{}`
	r := httptest.NewRequest(http.MethodPut, "/subcategorias?id=1", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)

	c := &SubcategoriaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Put()

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp models.ApiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Code != http.StatusOK {
		t.Errorf("Expected response code %d, got %d", http.StatusOK, resp.Code)
	}
}

func TestSubcategoriaController_Put_OnlyNombre(t *testing.T) {
	m := newSubMockOrm()

	sub := models.Subcategoria{NOMBRE: "Old Name"}
	m.Insert(&sub)

	orig := subcatOrmNew
	subcatOrmNew = func() subcatOrmer { return m }
	defer func() { subcatOrmNew = orig }()

	body := `{"nombre":"New Name"}`
	r := httptest.NewRequest(http.MethodPut, "/subcategorias?id=1", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)

	c := &SubcategoriaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Put()

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp models.ApiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Code != http.StatusOK {
		t.Errorf("Expected response code %d, got %d", http.StatusOK, resp.Code)
	}
}

func TestSubcategoriaController_Put_OnlyCategoriaId(t *testing.T) {
	m := newSubMockOrm()

	sub := models.Subcategoria{NOMBRE: "Name"}
	m.Insert(&sub)

	orig := subcatOrmNew
	subcatOrmNew = func() subcatOrmer { return m }
	defer func() { subcatOrmNew = orig }()

	body := `{"categoriaId":5}`
	r := httptest.NewRequest(http.MethodPut, "/subcategorias?id=1", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)

	c := &SubcategoriaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Put()

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp models.ApiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Code != http.StatusOK {
		t.Errorf("Expected response code %d, got %d", http.StatusOK, resp.Code)
	}
}
