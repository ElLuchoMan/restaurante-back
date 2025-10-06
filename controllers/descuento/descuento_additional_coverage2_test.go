package descuento

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

// Tests adicionales para DescuentoController - aumentar cobertura al 60%+

func TestDescuentoGetAll_Success(t *testing.T) {
	ctx := context.NewContext()
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/descuentos", nil)
	ctx.Reset(recorder, req)

	controller := &DescuentoController{}
	controller.Init(ctx, "DescuentoController", "GetAll", nil)

	// Mock ORM
	origOrm := newOrmForDesc
	defer func() { newOrmForDesc = origOrm }()

	newOrmForDesc = func() orm.Ormer {
		return &mockDescOrmer{
			queryTableFunc: func(interface{}) orm.QuerySeter {
				return &mockDescQS{
					allFunc: func(interface{}, ...string) (int64, error) {
						return 0, nil
					},
				}
			},
		}
	}

	controller.GetAll()

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestDescuentoPost_CuponSuccess(t *testing.T) {
	ctx := context.NewContext()
	recorder := httptest.NewRecorder()

	body := map[string]interface{}{
		"pedidoId": float64(1),
		"cuponId":  float64(10),
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/descuentos", bytes.NewReader(bodyBytes))
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = bodyBytes

	controller := &DescuentoController{}
	controller.Init(ctx, "DescuentoController", "Post", nil)

	// Mock service
	origService := newDescuentoService
	defer func() { newDescuentoService = origService }()

	newDescuentoService = func(orm.Ormer) *descuentoService {
		return &descuentoService{
			aplicarCuponFunc: func(int64, int64) (*models.PedidoDescuentoAplicado, error) {
				return &models.PedidoDescuentoAplicado{
					PK_ID_DESCUENTO: 1,
					MontoDescuento:  500,
				}, nil
			},
		}
	}

	controller.Post()

	// Debería retornar 201
	if recorder.Code == http.StatusCreated || recorder.Code == http.StatusOK {
		// Success
	} else {
		t.Errorf("Expected success status, got %d", recorder.Code)
	}
}

func TestDescuentoPost_OfertaSuccess(t *testing.T) {
	ctx := context.NewContext()
	recorder := httptest.NewRecorder()

	body := map[string]interface{}{
		"pedidoId":  float64(1),
		"ofertaId":  float64(20),
		"productos": []interface{}{float64(5)},
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/descuentos", bytes.NewReader(bodyBytes))
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = bodyBytes

	controller := &DescuentoController{}
	controller.Init(ctx, "DescuentoController", "Post", nil)

	// Mock service
	origService := newDescuentoService
	defer func() { newDescuentoService = origService }()

	newDescuentoService = func(orm.Ormer) *descuentoService {
		return &descuentoService{
			aplicarOfertaFunc: func(int64, int64, []int64) ([]*models.PedidoDescuentoAplicado, error) {
				return []*models.PedidoDescuentoAplicado{
					{
						PK_ID_DESCUENTO: 2,
						MontoDescuento:  300,
					},
				}, nil
			},
		}
	}

	controller.Post()

	// Debería retornar 201
	if recorder.Code == http.StatusCreated || recorder.Code == http.StatusOK {
		// Success
	} else {
		t.Errorf("Expected success status, got %d", recorder.Code)
	}
}

func TestDescuentoPost_InvalidJSONMissingFields(t *testing.T) {
	ctx := context.NewContext()
	recorder := httptest.NewRecorder()

	// Body sin pedidoId
	body := map[string]interface{}{
		"cuponId": float64(10),
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/descuentos", bytes.NewReader(bodyBytes))
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = bodyBytes

	controller := &DescuentoController{}
	controller.Init(ctx, "DescuentoController", "Post", nil)

	controller.Post()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestDescuentoPost_NoCuponNorOferta(t *testing.T) {
	ctx := context.NewContext()
	recorder := httptest.NewRecorder()

	// Body sin cuponId ni ofertaId
	body := map[string]interface{}{
		"pedidoId": float64(1),
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/descuentos", bytes.NewReader(bodyBytes))
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = bodyBytes

	controller := &DescuentoController{}
	controller.Init(ctx, "DescuentoController", "Post", nil)

	controller.Post()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestDescuentoGetByPedido_Success(t *testing.T) {
	ctx := context.NewContext()
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/descuentos/pedido?pedido_id=1", nil)
	ctx.Reset(recorder, req)
	ctx.Input.SetParam("pedido_id", "1")

	controller := &DescuentoController{}
	controller.Init(ctx, "DescuentoController", "GetByPedido", nil)

	// Mock ORM
	origOrm := newOrmForDesc
	defer func() { newOrmForDesc = origOrm }()

	newOrmForDesc = func() orm.Ormer {
		return &mockDescOrmer{
			queryTableFunc: func(interface{}) orm.QuerySeter {
				return &mockDescQS{
					filterFunc: func(string, ...interface{}) orm.QuerySeter {
						return &mockDescQS{
							allFunc: func(interface{}, ...string) (int64, error) {
								return 2, nil
							},
						}
					},
				}
			},
		}
	}

	controller.GetByPedido()

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestDescuentoGetByPedido_InvalidPedidoID(t *testing.T) {
	ctx := context.NewContext()
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/descuentos/pedido?pedido_id=invalid", nil)
	ctx.Reset(recorder, req)
	ctx.Input.SetParam("pedido_id", "invalid")

	controller := &DescuentoController{}
	controller.Init(ctx, "DescuentoController", "GetByPedido", nil)

	controller.GetByPedido()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestDescuentoDelete_Success(t *testing.T) {
	ctx := context.NewContext()
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/descuentos?id=1", nil)
	ctx.Reset(recorder, req)
	ctx.Input.SetParam("id", "1")

	controller := &DescuentoController{}
	controller.Init(ctx, "DescuentoController", "Delete", nil)

	// Mock ORM
	origOrm := newOrmForDesc
	defer func() { newOrmForDesc = origOrm }()

	newOrmForDesc = func() orm.Ormer {
		return &mockDescOrmer{
			deleteFunc: func(interface{}, ...string) (int64, error) {
				return 1, nil
			},
		}
	}

	controller.Delete()

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestDescuentoDelete_InvalidID(t *testing.T) {
	ctx := context.NewContext()
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/descuentos?id=invalid", nil)
	ctx.Reset(recorder, req)
	ctx.Input.SetParam("id", "invalid")

	controller := &DescuentoController{}
	controller.Init(ctx, "DescuentoController", "Delete", nil)

	controller.Delete()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestDescuentoDelete_NotFound(t *testing.T) {
	ctx := context.NewContext()
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/descuentos?id=999", nil)
	ctx.Reset(recorder, req)
	ctx.Input.SetParam("id", "999")

	controller := &DescuentoController{}
	controller.Init(ctx, "DescuentoController", "Delete", nil)

	// Mock ORM
	origOrm := newOrmForDesc
	defer func() { newOrmForDesc = origOrm }()

	newOrmForDesc = func() orm.Ormer {
		return &mockDescOrmer{
			deleteFunc: func(interface{}, ...string) (int64, error) {
				return 0, fmt.Errorf("not found")
			},
		}
	}

	controller.Delete()

	assert.NotEqual(t, http.StatusOK, recorder.Code)
}

// Mock structs for descuento tests
type descuentoService struct {
	aplicarCuponFunc  func(int64, int64) (*models.PedidoDescuentoAplicado, error)
	aplicarOfertaFunc func(int64, int64, []int64) ([]*models.PedidoDescuentoAplicado, error)
}

type mockDescOrmer struct {
	orm.Ormer
	queryTableFunc func(interface{}) orm.QuerySeter
	deleteFunc     func(interface{}, ...string) (int64, error)
}

func (m *mockDescOrmer) QueryTable(i interface{}) orm.QuerySeter {
	if m.queryTableFunc != nil {
		return m.queryTableFunc(i)
	}
	return nil
}

func (m *mockDescOrmer) Delete(md interface{}, cols ...string) (int64, error) {
	if m.deleteFunc != nil {
		return m.deleteFunc(md, cols...)
	}
	return 0, nil
}

type mockDescQS struct {
	orm.QuerySeter
	allFunc    func(interface{}, ...string) (int64, error)
	filterFunc func(string, ...interface{}) orm.QuerySeter
}

func (m *mockDescQS) All(container interface{}, cols ...string) (int64, error) {
	if m.allFunc != nil {
		return m.allFunc(container, cols...)
	}
	return 0, nil
}

func (m *mockDescQS) Filter(expr string, args ...interface{}) orm.QuerySeter {
	if m.filterFunc != nil {
		return m.filterFunc(expr, args...)
	}
	return m
}
