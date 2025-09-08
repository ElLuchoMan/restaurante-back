package controllers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
)

func TestRestauranteGetAllWithoutDB(t *testing.T) {
	orig := restOrmNew
	restOrmNew = func() restOrmer { return restFakeOrm{qsErr: errors.New("db")} }
	defer func() { restOrmNew = orig }()

	r := httptest.NewRequest(http.MethodGet, "/restaurante/v1/restaurantes", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := RestauranteController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetAll()

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al obtener restaurantes") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

// Mocks para RestauranteController
type restFakeQS struct{ err error }

func (f restFakeQS) All(res interface{}, _ ...string) (int64, error) { return 0, f.err }

type restFakeOrm struct {
	qsErr     error
	readErr   error
	insertErr error
	updateErr error
	deleteErr error
	updated   int
	deleted   int
}

func (f restFakeOrm) QueryTable(i interface{}) restQuerySeter { return restFakeQS{err: f.qsErr} }
func (f restFakeOrm) Read(v interface{}, _ ...string) error   { return f.readErr }
func (f restFakeOrm) Insert(v interface{}) (int64, error)     { return 1, f.insertErr }
func (f restFakeOrm) Update(v interface{}, _ ...string) (int64, error) {
	f.updated++
	return 1, f.updateErr
}
func (f restFakeOrm) Delete(v interface{}, _ ...string) (int64, error) {
	f.deleted++
	return 1, f.deleteErr
}

func TestRestauranteGetAllSuccess(t *testing.T) {
	orig := restOrmNew
	restOrmNew = func() restOrmer { return restFakeOrm{qsErr: nil} }
	defer func() { restOrmNew = orig }()

	r := httptest.NewRequest(http.MethodGet, "/restaurante/v1/restaurantes", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := RestauranteController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.GetAll()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestRestaurantePutSuccess(t *testing.T) {
	orig := restOrmNew
	restOrmNew = func() restOrmer { return restFakeOrm{readErr: nil} }
	defer func() { restOrmNew = orig }()

	body := `{"nombre":"X"}`
	r := httptest.NewRequest(http.MethodPut, "/restaurante/v1/restaurantes?id=1", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := RestauranteController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Put()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestRestauranteGetByIdInvalidID(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/restaurante/v1/restaurantes/search", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := RestauranteController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetById()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "El parámetro 'id' es inválido o está ausente") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestRestauranteGetByIdDBError(t *testing.T) {
	orig := restOrmNew
	restOrmNew = func() restOrmer { return restFakeOrm{readErr: errors.New("db")} }
	defer func() { restOrmNew = orig }()

	r := httptest.NewRequest(http.MethodGet, "/restaurante/v1/restaurantes/search?id=1", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := RestauranteController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetById()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Restaurante encontrado") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestRestauranteGetByIdSuccess(t *testing.T) {
	orig := restOrmNew
	restOrmNew = func() restOrmer { return restFakeOrm{} }
	defer func() { restOrmNew = orig }()

	r := httptest.NewRequest(http.MethodGet, "/restaurante/v1/restaurantes/search?id=1", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := RestauranteController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.GetById()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestRestaurantePostSuccess(t *testing.T) {
	orig := restOrmNew
	restOrmNew = func() restOrmer { return restFakeOrm{} }
	defer func() { restOrmNew = orig }()

	body := `{"nombre":"X"}`
	r := httptest.NewRequest(http.MethodPost, "/restaurante/v1/restaurantes", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := RestauranteController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Post()
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}

func TestRestaurantePostInvalidJSON(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/restaurante/v1/restaurantes", strings.NewReader("notjson"))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte("notjson")
	c := RestauranteController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error en la solicitud") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestRestaurantePostWithoutDB(t *testing.T) {
	orig := restOrmNew
	restOrmNew = func() restOrmer { return restFakeOrm{insertErr: errors.New("db")} }
	defer func() { restOrmNew = orig }()

	body := `{"restauranteId":1}`
	r := httptest.NewRequest(http.MethodPost, "/restaurante/v1/restaurantes", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := RestauranteController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al crear el restaurante") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestRestaurantePutInvalidID(t *testing.T) {
	r := httptest.NewRequest(http.MethodPut, "/restaurante/v1/restaurantes", strings.NewReader("{}"))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := RestauranteController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Put()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "El parámetro 'id' es inválido o está ausente") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestRestaurantePutNotFoundWithoutDB(t *testing.T) {
	orig := restOrmNew
	restOrmNew = func() restOrmer { return restFakeOrm{readErr: errors.New("db")} }
	defer func() { restOrmNew = orig }()

	r := httptest.NewRequest(http.MethodPut, "/restaurante/v1/restaurantes?id=1", strings.NewReader("{}"))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := RestauranteController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Put()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Restaurante no encontrado") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestRestauranteDeleteInvalidID(t *testing.T) {
	r := httptest.NewRequest(http.MethodDelete, "/restaurante/v1/restaurantes", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := RestauranteController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Delete()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "El parámetro 'id' es inválido o está ausente") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestRestauranteDeleteNotFoundWithoutDB(t *testing.T) {
	orig := restOrmNew
	restOrmNew = func() restOrmer { return restFakeOrm{deleteErr: errors.New("db")} }
	defer func() { restOrmNew = orig }()

	r := httptest.NewRequest(http.MethodDelete, "/restaurante/v1/restaurantes?id=1", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := RestauranteController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Delete()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Restaurante no encontrado") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestRestauranteGetByIdNotFound(t *testing.T) {
	orig := restOrmNew
	restOrmNew = func() restOrmer { return restFakeOrm{readErr: orm.ErrNoRows} }
	defer func() { restOrmNew = orig }()

	r := httptest.NewRequest(http.MethodGet, "/restaurante/v1/restaurantes/search?id=1", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := RestauranteController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetById()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Restaurante no encontrado") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestRestaurantePutInvalidJSON(t *testing.T) {
	orig := restOrmNew
	restOrmNew = func() restOrmer { return restFakeOrm{readErr: nil} }
	defer func() { restOrmNew = orig }()

	body := "badjson"
	r := httptest.NewRequest(http.MethodPut, "/restaurante/v1/restaurantes?id=1", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := RestauranteController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Put()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error en la solicitud") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestRestaurantePutUpdateError(t *testing.T) {
	orig := restOrmNew
	restOrmNew = func() restOrmer { return restFakeOrm{readErr: nil, updateErr: errors.New("db")} }
	defer func() { restOrmNew = orig }()

	body := `{"nombre":"X"}`
	r := httptest.NewRequest(http.MethodPut, "/restaurante/v1/restaurantes?id=1", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := RestauranteController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Put()

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al actualizar el restaurante") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestRestauranteDeleteSuccess(t *testing.T) {
	orig := restOrmNew
	restOrmNew = func() restOrmer { return restFakeOrm{} }
	defer func() { restOrmNew = orig }()

	r := httptest.NewRequest(http.MethodDelete, "/restaurante/v1/restaurantes?id=1", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := RestauranteController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Delete()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Restaurante eliminado") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}
