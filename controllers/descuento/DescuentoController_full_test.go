package descuento

import (
	"bytes"
	stdctx "context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	beegoctx "github.com/beego/beego/v2/server/web/context"
)

type stubDescuentoService struct {
	obtener func(stdctx.Context, int64) ([]*models.PedidoDescuentoAplicado, error)
	validar func(stdctx.Context, int64, *int64, *int64) error
	aplicar func(stdctx.Context, int64, *models.AplicarDescuentoRequest) (*models.PedidoDescuentoAplicado, error)
}

func (s stubDescuentoService) ObtenerDescuentosPedido(ctx stdctx.Context, pedidoId int64) ([]*models.PedidoDescuentoAplicado, error) {
	if s.obtener != nil {
		return s.obtener(ctx, pedidoId)
	}
	return nil, nil
}

func (s stubDescuentoService) ValidarExclusividadDescuento(ctx stdctx.Context, pedidoId int64, cuponId *int64, ofertaId *int64) error {
	if s.validar != nil {
		return s.validar(ctx, pedidoId, cuponId, ofertaId)
	}
	return nil
}

func (s stubDescuentoService) AplicarDescuento(ctx stdctx.Context, pedidoId int64, req *models.AplicarDescuentoRequest) (*models.PedidoDescuentoAplicado, error) {
	if s.aplicar != nil {
		return s.aplicar(ctx, pedidoId, req)
	}
	return nil, nil
}

type stubDescuentoOrmer struct {
	readFunc  func(interface{}, ...string) error
	readCalls int
}

func (s *stubDescuentoOrmer) Read(v interface{}, cols ...string) error {
	s.readCalls++
	if s.readFunc != nil {
		return s.readFunc(v, cols...)
	}
	return nil
}

func newTestController(t *testing.T, method, url string, body []byte) (*DescuentoController, *httptest.ResponseRecorder) {
	t.Helper()
	c := &DescuentoController{}
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(method, url, bytes.NewReader(body))
	ctx := beegoctx.NewContext()
	ctx.Reset(recorder, req)
	ctx.Input.RequestBody = body
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}
	return c, recorder
}

func decodeResponse(t *testing.T, recorder *httptest.ResponseRecorder) models.ApiResponse {
	t.Helper()
	var resp models.ApiResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unexpected error decoding response: %v", err)
	}
	return resp
}

func TestGetAll_InvalidPedidoId(t *testing.T) {
	controller, recorder := newTestController(t, http.MethodGet, "/descuentos/pedidos?pedido_id=invalid", nil)

	controller.GetAll()

	expectEqual(t, recorder.Code, http.StatusBadRequest)
}

func TestGetAll_MissingPedidoId(t *testing.T) {
	controller, recorder := newTestController(t, http.MethodGet, "/descuentos/pedidos", nil)

	controller.GetAll()

	expectEqual(t, recorder.Code, http.StatusBadRequest)
}

func TestGetAll_ZeroPedidoId(t *testing.T) {
	controller, recorder := newTestController(t, http.MethodGet, "/descuentos/pedidos?pedido_id=0", nil)

	controller.GetAll()

	expectEqual(t, recorder.Code, http.StatusBadRequest)
}

func TestGetAll_PedidoNotFound(t *testing.T) {
	stubOrm := &stubDescuentoOrmer{readFunc: func(interface{}, ...string) error { return orm.ErrNoRows }}
	originalFactory := descOrmFactory
	descOrmFactory = func() descuentoOrmer { return stubOrm }
	t.Cleanup(func() { descOrmFactory = originalFactory })

	controller, recorder := newTestController(t, http.MethodGet, "/descuentos/pedidos?pedido_id=10", nil)

	controller.GetAll()

	expectEqual(t, recorder.Code, http.StatusNotFound)
}

func TestGetAll_ReadError(t *testing.T) {
	stubOrm := &stubDescuentoOrmer{readFunc: func(interface{}, ...string) error { return errors.New("read failure") }}
	originalFactory := descOrmFactory
	descOrmFactory = func() descuentoOrmer { return stubOrm }
	t.Cleanup(func() { descOrmFactory = originalFactory })

	controller, recorder := newTestController(t, http.MethodGet, "/descuentos/pedidos?pedido_id=5", nil)

	controller.GetAll()

	expectEqual(t, recorder.Code, http.StatusInternalServerError)
	resp := decodeResponse(t, recorder)
	expectContains(t, resp.Cause, "read failure")
}

func TestGetAll_ServiceError(t *testing.T) {
	stubOrm := &stubDescuentoOrmer{}
	originalOrmFactory := descOrmFactory
	descOrmFactory = func() descuentoOrmer { return stubOrm }
	t.Cleanup(func() { descOrmFactory = originalOrmFactory })

	originalSvcFactory := descuentoServiceOrmFactory
	descuentoServiceOrmFactory = func() orm.Ormer { return nil }
	t.Cleanup(func() { descuentoServiceOrmFactory = originalSvcFactory })

	originalService := newDescuentoService
	newDescuentoService = func(orm.Ormer) descuentoService {
		return stubDescuentoService{
			obtener: func(stdctx.Context, int64) ([]*models.PedidoDescuentoAplicado, error) {
				return nil, errors.New("service failure")
			},
		}
	}
	t.Cleanup(func() { newDescuentoService = originalService })

	controller, recorder := newTestController(t, http.MethodGet, "/descuentos/pedidos?pedido_id=7", nil)

	controller.GetAll()

	expectEqual(t, recorder.Code, http.StatusInternalServerError)
	resp := decodeResponse(t, recorder)
	expectEqual(t, resp.Code, http.StatusInternalServerError)
	expectContains(t, resp.Cause, "service failure")
}

func TestGetAll_Success(t *testing.T) {
	descuentos := []*models.PedidoDescuentoAplicado{{MontoDescuento: 1500}}
	stubOrm := &stubDescuentoOrmer{}
	originalOrmFactory := descOrmFactory
	descOrmFactory = func() descuentoOrmer { return stubOrm }
	t.Cleanup(func() { descOrmFactory = originalOrmFactory })

	originalSvcFactory := descuentoServiceOrmFactory
	descuentoServiceOrmFactory = func() orm.Ormer { return nil }
	t.Cleanup(func() { descuentoServiceOrmFactory = originalSvcFactory })

	originalService := newDescuentoService
	newDescuentoService = func(orm.Ormer) descuentoService {
		return stubDescuentoService{
			obtener: func(stdctx.Context, int64) ([]*models.PedidoDescuentoAplicado, error) {
				return descuentos, nil
			},
		}
	}
	t.Cleanup(func() { newDescuentoService = originalService })

	controller, recorder := newTestController(t, http.MethodGet, "/descuentos/pedidos?pedido_id=3", nil)

	controller.GetAll()

	expectEqual(t, recorder.Code, http.StatusOK)
	resp := decodeResponse(t, recorder)
	expectEqual(t, resp.Code, http.StatusOK)
	data, ok := resp.Data.([]interface{})
	if !ok {
		t.Fatalf("expected data to be []interface{}, got %T", resp.Data)
	}
	if len(data) != 1 {
		t.Fatalf("expected one descuento, got %d", len(data))
	}
}

func TestPost_InvalidPedidoId(t *testing.T) {
	body := []byte(`{"cuponId":1,"montoDescuento":100}`)
	controller, recorder := newTestController(t, http.MethodPost, "/descuentos/pedidos?pedido_id=invalid", body)

	controller.Post()

	expectEqual(t, recorder.Code, http.StatusBadRequest)
}

func TestPost_MissingPedidoId(t *testing.T) {
	body := []byte(`{"cuponId":1,"montoDescuento":100}`)
	controller, recorder := newTestController(t, http.MethodPost, "/descuentos/pedidos", body)

	controller.Post()

	expectEqual(t, recorder.Code, http.StatusBadRequest)
}

func TestPost_ZeroPedidoId(t *testing.T) {
	body := []byte(`{"cuponId":1,"montoDescuento":100}`)
	controller, recorder := newTestController(t, http.MethodPost, "/descuentos/pedidos?pedido_id=0", body)

	controller.Post()

	expectEqual(t, recorder.Code, http.StatusBadRequest)
}

func TestPost_InvalidJSON(t *testing.T) {
	body := []byte("invalid json")
	controller, recorder := newTestController(t, http.MethodPost, "/descuentos/pedidos?pedido_id=1", body)

	controller.Post()

	expectEqual(t, recorder.Code, http.StatusBadRequest)
	resp := decodeResponse(t, recorder)
	expectEqual(t, resp.Code, http.StatusBadRequest)
}

func TestPost_BothCuponAndOferta(t *testing.T) {
	body := []byte(`{"cuponId":1,"ofertaId":2}`)
	controller, recorder := newTestController(t, http.MethodPost, "/descuentos/pedidos?pedido_id=1", body)

	controller.Post()

	expectEqual(t, recorder.Code, http.StatusUnprocessableEntity)
	resp := decodeResponse(t, recorder)
	expectEqual(t, resp.Code, http.StatusUnprocessableEntity)
}

func TestPost_NeitherCuponNorOferta(t *testing.T) {
	body := []byte(`{"montoDescuento":100}`)
	controller, recorder := newTestController(t, http.MethodPost, "/descuentos/pedidos?pedido_id=1", body)

	controller.Post()

	expectEqual(t, recorder.Code, http.StatusUnprocessableEntity)
}

func TestPost_ExclusivityError(t *testing.T) {
	originalSvcFactory := descuentoServiceOrmFactory
	descuentoServiceOrmFactory = func() orm.Ormer { return nil }
	t.Cleanup(func() { descuentoServiceOrmFactory = originalSvcFactory })

	originalService := newDescuentoService
	newDescuentoService = func(orm.Ormer) descuentoService {
		return stubDescuentoService{
			validar: func(stdctx.Context, int64, *int64, *int64) error {
				return errors.New("conflict")
			},
		}
	}
	t.Cleanup(func() { newDescuentoService = originalService })

	body := []byte(`{"cuponId":1,"montoDescuento":100}`)
	controller, recorder := newTestController(t, http.MethodPost, "/descuentos/pedidos?pedido_id=1", body)

	controller.Post()

	expectEqual(t, recorder.Code, http.StatusConflict)
	resp := decodeResponse(t, recorder)
	expectEqual(t, resp.Code, http.StatusConflict)
}

func TestPost_ServiceErrorMappings(t *testing.T) {
	cases := []struct {
		name            string
		errMsg          string
		expectedStatus  int
		expectedCode    int
		expectedMessage string
		expectedCause   string
	}{
		{name: "pedido", errMsg: "pedido no encontrado", expectedStatus: http.StatusOK, expectedCode: http.StatusNotFound, expectedMessage: "Pedido no encontrado", expectedCause: ""},
		{name: "cupon", errMsg: "cupón no encontrado", expectedStatus: http.StatusOK, expectedCode: http.StatusNotFound, expectedMessage: "cupón no encontrado", expectedCause: ""},
		{name: "oferta", errMsg: "oferta no encontrada", expectedStatus: http.StatusOK, expectedCode: http.StatusNotFound, expectedMessage: "oferta no encontrada", expectedCause: ""},
		{name: "exists", errMsg: "ya existe un descuento aplicado para este pedido", expectedStatus: http.StatusConflict, expectedCode: http.StatusConflict, expectedMessage: "ya existe un descuento aplicado para este pedido", expectedCause: ""},
		{name: "default", errMsg: "otro error", expectedStatus: http.StatusUnprocessableEntity, expectedCode: http.StatusUnprocessableEntity, expectedMessage: "Error al aplicar descuento", expectedCause: "otro error"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			originalSvcFactory := descuentoServiceOrmFactory
			descuentoServiceOrmFactory = func() orm.Ormer { return nil }
			t.Cleanup(func() { descuentoServiceOrmFactory = originalSvcFactory })

			originalService := newDescuentoService
			newDescuentoService = func(orm.Ormer) descuentoService {
				return stubDescuentoService{
					validar: func(stdctx.Context, int64, *int64, *int64) error { return nil },
					aplicar: func(stdctx.Context, int64, *models.AplicarDescuentoRequest) (*models.PedidoDescuentoAplicado, error) {
						return nil, errors.New(tc.errMsg)
					},
				}
			}
			t.Cleanup(func() { newDescuentoService = originalService })

			body := []byte(`{"cuponId":1,"montoDescuento":100}`)
			controller, recorder := newTestController(t, http.MethodPost, "/descuentos/pedidos?pedido_id=1", body)

			controller.Post()

			expectEqual(t, recorder.Code, tc.expectedStatus)
			resp := decodeResponse(t, recorder)
			expectEqual(t, resp.Code, tc.expectedCode)
			expectEqual(t, resp.Message, tc.expectedMessage)
			expectEqual(t, resp.Cause, tc.expectedCause)
		})
	}
}

func TestPost_Success(t *testing.T) {
	expected := &models.PedidoDescuentoAplicado{MontoDescuento: 200}

	originalSvcFactory := descuentoServiceOrmFactory
	descuentoServiceOrmFactory = func() orm.Ormer { return nil }
	t.Cleanup(func() { descuentoServiceOrmFactory = originalSvcFactory })

	originalService := newDescuentoService
	newDescuentoService = func(orm.Ormer) descuentoService {
		return stubDescuentoService{
			validar: func(stdctx.Context, int64, *int64, *int64) error { return nil },
			aplicar: func(stdctx.Context, int64, *models.AplicarDescuentoRequest) (*models.PedidoDescuentoAplicado, error) {
				return expected, nil
			},
		}
	}
	t.Cleanup(func() { newDescuentoService = originalService })

	body := []byte(`{"cuponId":1,"montoDescuento":100}`)
	controller, recorder := newTestController(t, http.MethodPost, "/descuentos/pedidos?pedido_id=1", body)

	controller.Post()

	expectEqual(t, recorder.Code, http.StatusCreated)
	resp := decodeResponse(t, recorder)
	expectEqual(t, resp.Code, http.StatusCreated)
}

func TestDescOrmAdapter_Read(t *testing.T) {
	called := false
	adapter := descOrmAdapter{readFn: func(interface{}, ...string) error {
		called = true
		return nil
	}}

	if err := adapter.Read(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatalf("expected read function to be called")
	}
}

func TestDescOrmFactory_UsesReadFuncFactory(t *testing.T) {
	called := false
	original := descBaseReadFunc
	descBaseReadFunc = func() func(interface{}, ...string) error {
		return func(interface{}, ...string) error {
			called = true
			return nil
		}
	}
	t.Cleanup(func() { descBaseReadFunc = original })

	adapter := descOrmFactory()
	if err := adapter.Read(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatalf("expected read func factory to be used")
	}
}

func TestDescOrmNew_UsesFactory(t *testing.T) {
	stubOrm := &stubDescuentoOrmer{}
	originalFactory := descOrmFactory
	descOrmFactory = func() descuentoOrmer { return stubOrm }
	t.Cleanup(func() { descOrmFactory = originalFactory })

	if descOrmNew() != stubOrm {
		t.Fatalf("expected stub ormer instance")
	}
}

func TestNewDescuentoService_UsesFactory(t *testing.T) {
	factoryCalled := false
	originalSvcFactory := descuentoServiceOrmFactory
	descuentoServiceOrmFactory = func() orm.Ormer {
		factoryCalled = true
		return nil
	}
	t.Cleanup(func() { descuentoServiceOrmFactory = originalSvcFactory })

	originalService := newDescuentoService
	newDescuentoService = func(o orm.Ormer) descuentoService {
		if o != nil {
			t.Fatalf("expected nil ormer, got %v", o)
		}
		return stubDescuentoService{
			validar: func(stdctx.Context, int64, *int64, *int64) error { return nil },
			aplicar: func(stdctx.Context, int64, *models.AplicarDescuentoRequest) (*models.PedidoDescuentoAplicado, error) {
				return &models.PedidoDescuentoAplicado{}, nil
			},
		}
	}
	t.Cleanup(func() { newDescuentoService = originalService })

	body := []byte(`{"cuponId":1,"montoDescuento":100}`)
	controller, recorder := newTestController(t, http.MethodPost, "/descuentos/pedidos?pedido_id=1", body)

	controller.Post()

	if !factoryCalled {
		t.Fatalf("expected service factory to be called")
	}
	expectEqual(t, recorder.Code, http.StatusCreated)
}

func TestDescReadFuncFactory_Default(t *testing.T) {
	called := false
	original := ormReadProvider
	ormReadProvider = func() func(interface{}, ...string) error {
		return func(interface{}, ...string) error {
			called = true
			return nil
		}
	}
	t.Cleanup(func() { ormReadProvider = original })

	readFn := descBaseReadFunc()
	if err := readFn(nil); err != nil {
		t.Fatalf("unexpected error invoking read func: %v", err)
	}
	if !called {
		t.Fatalf("expected base read func to be used")
	}
}

func TestDescuentoServiceOrmFactory_Default(t *testing.T) {
	called := false
	original := ormProvider
	ormProvider = func() orm.Ormer {
		called = true
		return nil
	}
	t.Cleanup(func() { ormProvider = original })

	if descuentoServiceOrmFactory() != nil {
		t.Fatalf("expected nil ormer from factory")
	}
	if !called {
		t.Fatalf("expected base ormer factory to run")
	}
}

func TestNewDescuentoService_Default(t *testing.T) {
	service := newDescuentoService(nil)
	if service == nil {
		t.Fatalf("expected default service instance")
	}
}

func expectEqual[T comparable](t *testing.T, got, want T) {
	t.Helper()
	if got != want {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func expectContains(t *testing.T, text, substr string) {
	t.Helper()
	if !strings.Contains(text, substr) {
		t.Fatalf("expected %q to contain %q", text, substr)
	}
}
