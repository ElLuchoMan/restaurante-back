package cupon

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
	"github.com/stretchr/testify/assert"
)

// TestValidarCupon_BadJSON cubre el caso de JSON inválido en ValidarCupon
func TestValidarCupon_BadJSON(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/cupones/validar", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte("invalid json")

	c := &CuponController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.ValidarCupon()

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp models.ApiResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

// TestRedimirCupon_BadJSON cubre el caso de JSON inválido en RedimirCupon
func TestRedimirCupon_BadJSON(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/cupones/TEST/redimir", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte("invalid json")
	ctx.Input.SetParam(":codigo", "TEST")

	c := &CuponController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.RedimirCupon()

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp models.ApiResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

// Skipping RedimirCupon tests with service mocking - require full service implementation or real DB

// TestListarRedenciones_CuponCodigoNoEncontrado cubre el caso de cupón no encontrado por código
func TestListarRedenciones_CuponCodigoNoEncontrado(t *testing.T) {
	// Guardar y restaurar la función original
	origCupOrmNew := cupOrmNew
	defer func() { cupOrmNew = origCupOrmNew }()

	// Mock del ORM que retorna error al buscar por código
	mockOrm := &mockCuponOrmerForListar{cuponNotFound: true}
	cupOrmNew = func() cuponOrmer {
		return mockOrm
	}

	r := httptest.NewRequest(http.MethodGet, "/cupones/redenciones?cupon_codigo=NOTFOUND", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)

	c := &CuponController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.ListarRedenciones()

	assert.Equal(t, http.StatusOK, w.Code)
	var resp models.ApiResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, http.StatusOK, resp.Code)

	// Verificar que la respuesta tenga data vacía
	data := resp.Data.(map[string]interface{})
	assert.Equal(t, float64(0), data["total"])
}

// TestListarRedenciones_InvalidCuponId cubre el caso de cupon_id inválido
func TestListarRedenciones_InvalidCuponId(t *testing.T) {
	origCupOrmNew := cupOrmNew
	defer func() { cupOrmNew = origCupOrmNew }()

	mockOrm := &mockCuponOrmerForListar{}
	cupOrmNew = func() cuponOrmer {
		return mockOrm
	}

	r := httptest.NewRequest(http.MethodGet, "/cupones/redenciones?cupon_id=invalid", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)

	c := &CuponController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.ListarRedenciones()

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestListarRedenciones_InvalidClienteId cubre el caso de cliente_id inválido
func TestListarRedenciones_InvalidClienteId(t *testing.T) {
	origCupOrmNew := cupOrmNew
	defer func() { cupOrmNew = origCupOrmNew }()

	mockOrm := &mockCuponOrmerForListar{}
	cupOrmNew = func() cuponOrmer {
		return mockOrm
	}

	r := httptest.NewRequest(http.MethodGet, "/cupones/redenciones?cliente_id=invalid", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)

	c := &CuponController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.ListarRedenciones()

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestListarRedenciones_LimitExceeds100 cubre el caso de límite > 100
func TestListarRedenciones_LimitExceeds100(t *testing.T) {
	origCupOrmNew := cupOrmNew
	defer func() { cupOrmNew = origCupOrmNew }()

	mockOrm := &mockCuponOrmerForListar{}
	cupOrmNew = func() cuponOrmer {
		return mockOrm
	}

	r := httptest.NewRequest(http.MethodGet, "/cupones/redenciones?limit=200", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)

	c := &CuponController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.ListarRedenciones()

	assert.Equal(t, http.StatusOK, w.Code)
}

// Mock types

// mockCuponOrmerForListar mock para ListarRedenciones
type mockCuponOrmerForListar struct {
	cuponNotFound bool
	countError    bool
	queryError    bool
}

func (m *mockCuponOrmerForListar) QueryTable(name interface{}) cuponQuerySeter {
	return &mockCuponQuerySeterForListar{
		cuponNotFound: m.cuponNotFound,
		countError:    m.countError,
		queryError:    m.queryError,
	}
}

func (m *mockCuponOrmerForListar) Insert(interface{}) (int64, error) {
	return 0, nil
}

func (m *mockCuponOrmerForListar) Read(interface{}, ...string) error {
	return nil
}

func (m *mockCuponOrmerForListar) Update(interface{}, ...string) (int64, error) {
	return 0, nil
}

func (m *mockCuponOrmerForListar) Delete(interface{}, ...string) (int64, error) {
	return 0, nil
}

// mockCuponQuerySeterForListar mock para QuerySeter de ListarRedenciones
type mockCuponQuerySeterForListar struct {
	cuponNotFound bool
	countError    bool
	queryError    bool
	filters       map[string]interface{}
}

func (m *mockCuponQuerySeterForListar) One(container interface{}) error {
	if m.cuponNotFound {
		return orm.ErrNoRows
	}
	// Llenar el contenedor con datos de prueba
	if cupon, ok := container.(*models.Cupon); ok {
		cupon.PkIdCupon = 1
		cupon.Codigo = "TESTCODE"
	}
	return nil
}

func (m *mockCuponQuerySeterForListar) Filter(expr string, args ...interface{}) cuponQuerySeter {
	if m.filters == nil {
		m.filters = make(map[string]interface{})
	}
	m.filters[expr] = args
	return m
}

func (m *mockCuponQuerySeterForListar) OrderBy(exprs ...string) cuponQuerySeter {
	return m
}

func (m *mockCuponQuerySeterForListar) Limit(limit int) cuponQuerySeter {
	return m
}

func (m *mockCuponQuerySeterForListar) Offset(offset int64) cuponQuerySeter {
	return m
}

func (m *mockCuponQuerySeterForListar) Count() (int64, error) {
	if m.countError {
		return 0, fmt.Errorf("count error")
	}
	return 0, nil
}

func (m *mockCuponQuerySeterForListar) All(container interface{}, cols ...string) (int64, error) {
	if m.queryError {
		return 0, fmt.Errorf("query error")
	}
	return 0, nil
}
