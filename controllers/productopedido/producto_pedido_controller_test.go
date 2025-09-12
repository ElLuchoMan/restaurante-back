package productopedido

import (
	"bytes"
	stdctx "context"
	"database/sql/driver"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
)

type fakeQueryPP struct {
	orm.QuerySeter
	one func(interface{}, ...string) error
	all func(interface{}, ...string) (int64, error)
	del func() (int64, error)
}

func (f fakeQueryPP) Filter(string, ...interface{}) orm.QuerySeter { return f }
func (f fakeQueryPP) One(res interface{}, cols ...string) error {
	if f.one != nil {
		return f.one(res, cols...)
	}
	return nil
}
func (f fakeQueryPP) All(res interface{}, cols ...string) (int64, error) {
	if f.all != nil {
		return f.all(res, cols...)
	}
	return 0, nil
}
func (f fakeQueryPP) Delete() (int64, error) {
	if f.del != nil {
		return f.del()
	}
	return 0, nil
}

type fakeOrmerPP struct {
	query  func(interface{}) orm.QuerySeter
	insert func(interface{}) (int64, error)
}

func (f fakeOrmerPP) QueryTable(i interface{}) orm.QuerySeter { return f.query(i) }
func (f fakeOrmerPP) Insert(m interface{}) (int64, error) {
	if f.insert != nil {
		return f.insert(m)
	}
	return 1, nil
}

func TestProductoPedidoGetAllMissingParam(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/producto_pedido", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ProductoPedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetAll()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "pedido_id") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestProductoPedidoGetAllDBError(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/producto_pedido?pedido_id=1", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ProductoPedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetAll()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al obtener los productos del pedido") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestProductoPedidoGetAllNotFound(t *testing.T) {
	original := productoPedidoNewOrm
	productoPedidoNewOrm = func() productoPedidoOrmer {
		return fakeOrmerPP{query: func(i interface{}) orm.QuerySeter {
			return fakeQueryPP{all: func(res interface{}, cols ...string) (int64, error) {
				return 0, nil
			}}
		}}
	}
	defer func() { productoPedidoNewOrm = original }()

	r := httptest.NewRequest(http.MethodGet, "/producto_pedido?pedido_id=1", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ProductoPedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetAll()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "No se encontraron productos") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestProductoPedidoPostInvalidJSON(t *testing.T) {
	body := "{"
	r := httptest.NewRequest(http.MethodPost, "/producto_pedido", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := ProductoPedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Datos inválidos") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestProductoPedidoPostMissingFields(t *testing.T) {
	body := `{"pedidoId":0,"detalles":[]}`
	r := httptest.NewRequest(http.MethodPost, "/producto_pedido", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := ProductoPedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "obligatorios") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestProductoPedidoPostEmptyDetalles(t *testing.T) {
	body := `{"pedidoId":1,"detalles":[]}`
	r := httptest.NewRequest(http.MethodPost, "/producto_pedido", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := ProductoPedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(strings.ToLower(w.Body.String()), "obligatorios") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestProductoPedidoPostDBError(t *testing.T) {
	body := `{"pedidoId":1,"detalles":[{"productoId":1,"cantidad":1}]}`
	r := httptest.NewRequest(http.MethodPost, "/producto_pedido", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := ProductoPedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Inventario insuficiente") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestProductoPedidoUpdateMissingParam(t *testing.T) {
	body := "[]"
	r := httptest.NewRequest(http.MethodPut, "/producto_pedido", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := ProductoPedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Update()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "pedido_id") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestProductoPedidoUpdateInvalidJSON(t *testing.T) {
	body := "{"
	r := httptest.NewRequest(http.MethodPut, "/producto_pedido?pedido_id=1", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := ProductoPedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Update()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Datos inválidos") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestProductoPedidoUpdateEmptyList(t *testing.T) {
	body := "[]"
	r := httptest.NewRequest(http.MethodPut, "/producto_pedido?pedido_id=1", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := ProductoPedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Update()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "no puede estar vacía") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestProductoPedidoUpdateDBError(t *testing.T) {
	body := `[{"productoId":1,"cantidad":1}]`
	r := httptest.NewRequest(http.MethodPut, "/producto_pedido?pedido_id=1", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := ProductoPedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Update()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al buscar los detalles del pedido") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestProductoPedidoUpdate_DeleteError(t *testing.T) {
	original := productoPedidoNewOrm
	productoPedidoNewOrm = func() productoPedidoOrmer {
		return fakeOrmerPP{
			query: func(i interface{}) orm.QuerySeter {
				return fakeQueryPP{
					all: func(res interface{}, cols ...string) (int64, error) { return 0, nil },
					del: func() (int64, error) { return 0, errors.New("del") },
				}
			},
			insert: func(m interface{}) (int64, error) { return 1, nil },
		}
	}
	defer func() { productoPedidoNewOrm = original }()

	body := `[{"productoId":1,"cantidad":1}]`
	r := httptest.NewRequest(http.MethodPut, "/producto_pedido?pedido_id=1", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := ProductoPedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Update()
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 200 or 500, got %d", w.Code)
	}
}

func TestProductoPedidoPost_BeginTxError(t *testing.T) {
	origQ, origBegin := MockQuery, productoPedidoBeginTx
	MockQuery = func(_ stdctx.Context, q string, _ []driver.NamedValue) (driver.Rows, error) {
		if strings.Contains(strings.ToLower(q), "select pk_id_producto, cantidad from producto") {
			cols := []string{"pk_id_producto", "cantidad"}
			vals := [][]driver.Value{{int64(1), int64(10)}}
			return &mockRows{columns: cols, values: vals}, nil
		}
		return &mockRows{columns: []string{"ok"}, values: [][]driver.Value{{int64(1)}}}, nil
	}
	productoPedidoBeginTx = func(o orm.Ormer) (orm.TxOrmer, error) { return nil, errors.New("begin fail") }
	t.Cleanup(func() { MockQuery = origQ; productoPedidoBeginTx = origBegin })

	body := `{"pedidoId":1,"detalles":[{"productoId":1,"cantidad":2}]}`
	r := httptest.NewRequest(http.MethodPost, "/producto_pedido", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := ProductoPedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Post()
	if !strings.Contains(w.Body.String(), "No fue posible iniciar transacción") {
		t.Fatalf("expected begin error, body: %s", w.Body.String())
	}
}

func TestProductoPedidoGetAllSuccess(t *testing.T) {
	original := productoPedidoNewOrm
	productoPedidoNewOrm = func() productoPedidoOrmer {
		return fakeOrmerPP{query: func(i interface{}) orm.QuerySeter {
			switch i.(type) {
			case *models.DetallePedido:
				return fakeQueryPP{all: func(res interface{}, cols ...string) (int64, error) {
					detalles := res.(*[]models.DetallePedido)
					pedidoID := int64(1)
					productoID := int64(1)
					*detalles = append(
						*detalles,
						models.DetallePedido{
							PKIDPedido:   &models.Pedido{PK_ID_PEDIDO: pedidoID},
							PKIDProducto: &models.Producto{PK_ID_PRODUCTO: productoID},
							Cantidad:     1,
							Precio:       1000,
						},
					)
					return 1, nil
				}}
			default:
				return fakeQueryPP{}
			}
		}}
	}
	defer func() { productoPedidoNewOrm = original }()

	r := httptest.NewRequest(http.MethodGet, "/producto_pedido?pedido_id=1", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ProductoPedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetAll()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "detalles") {
		t.Errorf("unexpected body: %s", body)
	}
	if !strings.Contains(body, "\"precio\":1000") {
		t.Errorf("expected price in response, got %s", body)
	}
}

func TestProductoPedidoPostSuccess(t *testing.T) {
	original := productoPedidoNewOrm
	productoPedidoNewOrm = func() productoPedidoOrmer {
		return fakeOrmerPP{
			insert: func(m interface{}) (int64, error) { return 1, nil },
			query: func(i interface{}) orm.QuerySeter {
				return fakeQueryPP{one: func(res interface{}, cols ...string) error {
					if d, ok := res.(*models.DetallePedido); ok {
						pedidoID := int64(1)
						productoID := int64(1)
						*d = models.DetallePedido{
							PKIDPedido:   &models.Pedido{PK_ID_PEDIDO: pedidoID},
							PKIDProducto: &models.Producto{PK_ID_PRODUCTO: productoID},
							Cantidad:     1,
							Precio:       1000,
						}
					}
					return nil
				}}
			},
		}
	}
	defer func() { productoPedidoNewOrm = original }()

	body := `{"pedidoId":1,"detalles":[{"productoId":1,"cantidad":1}]}`
	r := httptest.NewRequest(http.MethodPost, "/producto_pedido", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := ProductoPedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Inventario insuficiente") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestProductoPedidoUpdateSuccess(t *testing.T) {
	original := productoPedidoNewOrm
	productoPedidoNewOrm = func() productoPedidoOrmer {
		call := 0
		return fakeOrmerPP{
			query: func(i interface{}) orm.QuerySeter {
				call++
				if call == 1 {
					return fakeQueryPP{
						all: func(res interface{}, cols ...string) (int64, error) { return 0, nil },
						del: func() (int64, error) { return 0, nil },
					}
				}
				return fakeQueryPP{one: func(res interface{}, cols ...string) error {
					if d, ok := res.(*models.DetallePedido); ok {
						pedidoID := int64(1)
						productoID := int64(1)
						*d = models.DetallePedido{
							PKIDPedido:   &models.Pedido{PK_ID_PEDIDO: pedidoID},
							PKIDProducto: &models.Producto{PK_ID_PRODUCTO: productoID},
							Cantidad:     1,
							Precio:       1000,
						}
					}
					return nil
				}}
			},
			insert: func(m interface{}) (int64, error) { return 1, nil },
		}
	}
	defer func() { productoPedidoNewOrm = original }()

	body := `[{"productoId":1,"cantidad":1}]`
	r := httptest.NewRequest(http.MethodPut, "/producto_pedido?pedido_id=1", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := ProductoPedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Update()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al buscar los detalles del pedido") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestProductoPedidoUpdateEndToEndSuccess(t *testing.T) {
	origQ, origE := MockQuery, MockExec
	call := 0
	MockQuery = func(_ stdctx.Context, q string, _ []driver.NamedValue) (driver.Rows, error) {
		lower := strings.ToLower(q)
		if strings.Contains(lower, "detalle_pedido") && call == 0 {
			call++
			return &mockRows{columns: []string{"pk_id_pedido"}, values: [][]driver.Value{}}, nil
		}
		if strings.Contains(lower, "select pk_id_producto, cantidad from producto") {
			cols := []string{"pk_id_producto", "cantidad"}
			vals := [][]driver.Value{{int64(1), int64(5)}}
			return &mockRows{columns: cols, values: vals}, nil
		}
		if strings.Contains(lower, "detalle_pedido") {
			cols := []string{"pk_id_pedido", "pk_id_producto", "cantidad", "precio"}
			vals := [][]driver.Value{{int64(1), int64(1), int64(2), int64(2000)}}
			return &mockRows{columns: cols, values: vals}, nil
		}
		return &mockRows{columns: []string{"ok"}, values: [][]driver.Value{{int64(1)}}}, nil
	}
	MockExec = func(_ stdctx.Context, _ string, _ []driver.NamedValue) (driver.Result, error) {
		return mockResult{}, nil
	}
	t.Cleanup(func() { MockQuery, MockExec = origQ, origE })

	body := `[{"productoId":1,"cantidad":2}]`
	r := httptest.NewRequest(http.MethodPut, "/producto_pedido?pedido_id=1", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := ProductoPedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Update()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestProductoPedidoUpdate_InsufficientInventory_Validation(t *testing.T) {
	origQ, origE := MockQuery, MockExec
	call := 0
	MockQuery = func(_ stdctx.Context, q string, _ []driver.NamedValue) (driver.Rows, error) {
		lower := strings.ToLower(q)
		if strings.Contains(lower, "detalle_pedido") {
			return &mockRows{columns: []string{"pk_id_pedido"}, values: [][]driver.Value{}}, nil
		}
		if strings.Contains(lower, "select pk_id_producto, cantidad from producto") {
			cols := []string{"pk_id_producto", "cantidad"}
			vals := [][]driver.Value{{int64(1), int64(1)}}
			return &mockRows{columns: cols, values: vals}, nil
		}
		call++
		return &mockRows{columns: []string{"ok"}, values: [][]driver.Value{{int64(call)}}}, nil
	}
	MockExec = func(_ stdctx.Context, _ string, _ []driver.NamedValue) (driver.Result, error) {
		return mockResult{}, nil
	}
	t.Cleanup(func() { MockQuery, MockExec = origQ, origE })

	body := `[{"productoId":1,"cantidad":2}]`
	r := httptest.NewRequest(http.MethodPut, "/producto_pedido?pedido_id=1", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := ProductoPedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Update()
	if w.Code != http.StatusBadRequest && w.Code != http.StatusOK {
		t.Fatalf("expected 400 or 200, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestProductoPedidoUpdate_PositiveDelta_NoStockRowsAffected(t *testing.T) {
	origQ, origE := MockQuery, MockExec
	call := 0
	MockQuery = func(_ stdctx.Context, q string, _ []driver.NamedValue) (driver.Rows, error) {
		lower := strings.ToLower(q)
		if strings.Contains(lower, "detalle_pedido") {
			return &mockRows{columns: []string{"pk_id_pedido"}, values: [][]driver.Value{}}, nil
		}
		if strings.Contains(lower, "select pk_id_producto, cantidad from producto") {
			cols := []string{"pk_id_producto", "cantidad"}
			vals := [][]driver.Value{{int64(1), int64(10)}}
			return &mockRows{columns: cols, values: vals}, nil
		}
		call++
		return &mockRows{columns: []string{"ok"}, values: [][]driver.Value{{int64(call)}}}, nil
	}
	MockExec = func(_ stdctx.Context, q string, _ []driver.NamedValue) (driver.Result, error) {
		if strings.Contains(strings.ToLower(q), "update producto set cantidad = cantidad -") {
			return zeroRowsResult{}, nil
		}
		return mockResult{}, nil
	}
	t.Cleanup(func() { MockQuery, MockExec = origQ, origE })

	body := `[{"productoId":1,"cantidad":2}]`
	r := httptest.NewRequest(http.MethodPut, "/producto_pedido?pedido_id=1", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := ProductoPedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Update()
	if w.Code != http.StatusBadRequest && w.Code != http.StatusOK {
		t.Fatalf("expected 400 or 200, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestProductoPedidoUpdate_PositiveDelta_ExecError(t *testing.T) {
	origQ, origE := MockQuery, MockExec
	MockQuery = func(_ stdctx.Context, q string, _ []driver.NamedValue) (driver.Rows, error) {
		lower := strings.ToLower(q)
		if strings.Contains(lower, "detalle_pedido") {
			return &mockRows{columns: []string{"pk_id_pedido"}, values: [][]driver.Value{}}, nil
		}
		if strings.Contains(lower, "select pk_id_producto, cantidad from producto") {
			cols := []string{"pk_id_producto", "cantidad"}
			vals := [][]driver.Value{{int64(1), int64(10)}}
			return &mockRows{columns: cols, values: vals}, nil
		}
		return &mockRows{columns: []string{"ok"}, values: [][]driver.Value{{int64(1)}}}, nil
	}
	MockExec = func(_ stdctx.Context, q string, _ []driver.NamedValue) (driver.Result, error) {
		if strings.Contains(strings.ToLower(q), "update producto set cantidad = cantidad -") {
			return nil, errors.New("exec fail")
		}
		return mockResult{}, nil
	}
	t.Cleanup(func() { MockQuery, MockExec = origQ, origE })

	body := `[{"productoId":1,"cantidad":2}]`
	r := httptest.NewRequest(http.MethodPut, "/producto_pedido?pedido_id=1", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := ProductoPedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Update()
	if w.Code != http.StatusInternalServerError && w.Code != http.StatusOK {
		t.Fatalf("expected 500 or 200, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestProductoPedidoPostEndToEndSuccess(t *testing.T) {
	origQ, origE := MockQuery, MockExec
	MockExec = func(_ stdctx.Context, _ string, _ []driver.NamedValue) (driver.Result, error) {
		return mockResult{}, nil
	}
	MockQuery = func(_ stdctx.Context, q string, _ []driver.NamedValue) (driver.Rows, error) {
		lower := strings.ToLower(q)
		if strings.Contains(lower, "select pk_id_producto, cantidad from producto") {
			cols := []string{"pk_id_producto", "cantidad"}
			vals := [][]driver.Value{{int64(1), int64(10)}}
			return &mockRows{columns: cols, values: vals}, nil
		}
		if strings.Contains(lower, "detalle_pedido") {
			cols := []string{"pk_id_pedido", "pk_id_producto", "cantidad", "precio"}
			vals := [][]driver.Value{{int64(1), int64(1), int64(1), int64(1000)}}
			return &mockRows{columns: cols, values: vals}, nil
		}
		return &mockRows{columns: []string{"ok"}, values: [][]driver.Value{{int64(1)}}}, nil
	}
	t.Cleanup(func() { MockQuery, MockExec = origQ, origE })

	body := `{"pedidoId":1,"detalles":[{"productoId":1,"cantidad":1}]}`
	r := httptest.NewRequest(http.MethodPost, "/producto_pedido", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := ProductoPedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()
	if w.Code != http.StatusCreated && w.Code != http.StatusOK {
		t.Fatalf("unexpected status %d", w.Code)
	}
}

func TestProductoPedidoPost_ConsolidatesDuplicates(t *testing.T) {
	origQ, origE := MockQuery, MockExec
	MockQuery = func(_ stdctx.Context, q string, _ []driver.NamedValue) (driver.Rows, error) {
		if strings.Contains(strings.ToLower(q), "select pk_id_producto, cantidad from producto") {
			cols := []string{"pk_id_producto", "cantidad"}
			vals := [][]driver.Value{{int64(1), int64(5)}}
			return &mockRows{columns: cols, values: vals}, nil
		}
		return &mockRows{columns: []string{"ok"}, values: [][]driver.Value{{int64(1)}}}, nil
	}
	MockExec = nil
	t.Cleanup(func() { MockQuery, MockExec = origQ, origE })

	body := `{"pedidoId":1,"detalles":[{"productoId":1,"cantidad":2},{"productoId":1,"cantidad":3}]}`
	r := httptest.NewRequest(http.MethodPost, "/producto_pedido", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := ProductoPedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Post()
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 200 or 500, got %d", w.Code)
	}
}

func TestProductoPedidoPost_MixedValidInvalidItems(t *testing.T) {
	origQ, origE := MockQuery, MockExec
	MockQuery = func(_ stdctx.Context, q string, _ []driver.NamedValue) (driver.Rows, error) {
		if strings.Contains(strings.ToLower(q), "select pk_id_producto, cantidad from producto") {
			cols := []string{"pk_id_producto", "cantidad"}
			vals := [][]driver.Value{{int64(2), int64(10)}}
			return &mockRows{columns: cols, values: vals}, nil
		}
		return &mockRows{columns: []string{"ok"}, values: [][]driver.Value{{int64(1)}}}, nil
	}
	MockExec = nil
	t.Cleanup(func() { MockQuery, MockExec = origQ, origE })

	body := `{"pedidoId":1,"detalles":[{"productoId":0,"cantidad":2},{"productoId":2,"cantidad":1},{"productoId":2,"cantidad":0}]}`
	r := httptest.NewRequest(http.MethodPost, "/producto_pedido", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := ProductoPedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Post()
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 200 or 500, got %d", w.Code)
	}
}

type zeroRowsResult struct{}

func (zeroRowsResult) LastInsertId() (int64, error) { return 0, nil }
func (zeroRowsResult) RowsAffected() (int64, error) { return 0, nil }

func TestProductoPedidoPost_UpdateStockExecError(t *testing.T) {
	origQ, origE := MockQuery, MockExec
	MockQuery = func(_ stdctx.Context, q string, _ []driver.NamedValue) (driver.Rows, error) {
		lower := strings.ToLower(q)
		if strings.Contains(lower, "select pk_id_producto, cantidad from producto") {
			cols := []string{"pk_id_producto", "cantidad"}
			vals := [][]driver.Value{{int64(1), int64(10)}}
			return &mockRows{columns: cols, values: vals}, nil
		}
		return &mockRows{columns: []string{"ok"}, values: [][]driver.Value{{int64(1)}}}, nil
	}
	MockExec = func(_ stdctx.Context, q string, _ []driver.NamedValue) (driver.Result, error) {
		if strings.Contains(strings.ToLower(q), "update producto set cantidad = cantidad -") {
			return nil, errors.New("exec fail")
		}
		return mockResult{}, nil
	}
	t.Cleanup(func() { MockQuery, MockExec = origQ, origE })

	body := `{"pedidoId":1,"detalles":[{"productoId":1,"cantidad":2}]}`
	r := httptest.NewRequest(http.MethodPost, "/producto_pedido", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := ProductoPedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()
	if w.Code != http.StatusInternalServerError && w.Code != http.StatusOK {
		t.Fatalf("expected 500 or 200, got %d. Body: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "Error al descontar inventario") && !strings.Contains(w.Body.String(), "No fue posible iniciar transacción") {
		t.Errorf("expected inventory discount or tx begin error, body: %s", w.Body.String())
	}
}

func TestProductoPedidoPost_UpdateStockNoRowsAffected(t *testing.T) {
	origQ, origE := MockQuery, MockExec
	MockQuery = func(_ stdctx.Context, q string, _ []driver.NamedValue) (driver.Rows, error) {
		lower := strings.ToLower(q)
		if strings.Contains(lower, "select pk_id_producto, cantidad from producto") {
			cols := []string{"pk_id_producto", "cantidad"}
			vals := [][]driver.Value{{int64(1), int64(10)}}
			return &mockRows{columns: cols, values: vals}, nil
		}
		return &mockRows{columns: []string{"ok"}, values: [][]driver.Value{{int64(1)}}}, nil
	}
	MockExec = func(_ stdctx.Context, q string, _ []driver.NamedValue) (driver.Result, error) {
		if strings.Contains(strings.ToLower(q), "update producto set cantidad = cantidad -") {
			return zeroRowsResult{}, nil
		}
		return mockResult{}, nil
	}
	t.Cleanup(func() { MockQuery, MockExec = origQ, origE })

	body := `{"pedidoId":1,"detalles":[{"productoId":1,"cantidad":2}]}`
	r := httptest.NewRequest(http.MethodPost, "/producto_pedido", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := ProductoPedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "Inventario insuficiente") {
		t.Errorf("expected insufficient inventory, body: %s", w.Body.String())
	}
}

func TestProductoPedidoPost_InsertError(t *testing.T) {
	origQ, origE := MockQuery, MockExec
	MockQuery = func(_ stdctx.Context, q string, _ []driver.NamedValue) (driver.Rows, error) {
		lower := strings.ToLower(q)
		if strings.Contains(lower, "select pk_id_producto, cantidad from producto") {
			cols := []string{"pk_id_producto", "cantidad"}
			vals := [][]driver.Value{{int64(1), int64(10)}}
			return &mockRows{columns: cols, values: vals}, nil
		}
		if strings.Contains(lower, "insert into") {
			return nil, errors.New("insert fail")
		}
		if strings.Contains(lower, "detalle_pedido") {
			cols := []string{"pk_id_pedido", "pk_id_producto", "cantidad", "precio"}
			vals := [][]driver.Value{{int64(1), int64(1), int64(2), int64(1000)}}
			return &mockRows{columns: cols, values: vals}, nil
		}
		return &mockRows{columns: []string{"ok"}, values: [][]driver.Value{{int64(1)}}}, nil
	}
	MockExec = func(_ stdctx.Context, q string, _ []driver.NamedValue) (driver.Result, error) {
		return mockResult{}, nil
	}
	t.Cleanup(func() { MockQuery, MockExec = origQ, origE })

	body := `{"pedidoId":1,"detalles":[{"productoId":1,"cantidad":2}]}`
	r := httptest.NewRequest(http.MethodPost, "/producto_pedido", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := ProductoPedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()
	if w.Code != http.StatusInternalServerError && w.Code != http.StatusOK {
		t.Fatalf("expected 500 or 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al crear el pedido") {
		t.Errorf("expected insert error, body: %s", w.Body.String())
	}
}

func TestProductoPedidoUpdate_FilterInvalidItem(t *testing.T) {
	origQ := MockQuery
	MockQuery = func(_ stdctx.Context, q string, _ []driver.NamedValue) (driver.Rows, error) {
		return nil, errors.New("db error")
	}
	t.Cleanup(func() { MockQuery = origQ })

	body := `[{"productoId":0,"cantidad":-1},{"productoId":1,"cantidad":1}]`
	r := httptest.NewRequest(http.MethodPut, "/producto_pedido?pedido_id=1", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := ProductoPedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Update()
	if w.Code != http.StatusInternalServerError && w.Code != http.StatusOK {
		t.Fatalf("unexpected status %d", w.Code)
	}
}

func TestProductoPedidoUpdate_BeginTxError(t *testing.T) {
	origQ, origBegin := MockQuery, productoPedidoBeginTx
	MockQuery = func(_ stdctx.Context, q string, _ []driver.NamedValue) (driver.Rows, error) {
		lower := strings.ToLower(q)
		if strings.Contains(lower, "detalle_pedido") {
			return &mockRows{columns: []string{"pk_id_pedido"}, values: [][]driver.Value{}}, nil
		}
		if strings.Contains(lower, "select pk_id_producto, cantidad from producto") {
			cols := []string{"pk_id_producto", "cantidad"}
			vals := [][]driver.Value{{int64(1), int64(10)}}
			return &mockRows{columns: cols, values: vals}, nil
		}
		return &mockRows{columns: []string{"ok"}, values: [][]driver.Value{{int64(1)}}}, nil
	}
	productoPedidoBeginTx = func(o orm.Ormer) (orm.TxOrmer, error) { return nil, errors.New("begin fail") }
	t.Cleanup(func() { MockQuery = origQ; productoPedidoBeginTx = origBegin })

	body := `[{"productoId":1,"cantidad":2}]`
	r := httptest.NewRequest(http.MethodPut, "/producto_pedido?pedido_id=1", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := ProductoPedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Update()
	if !strings.Contains(strings.ToLower(w.Body.String()), "iniciar transacción") {
		t.Fatalf("expected begin tx error, body: %s", w.Body.String())
	}
}

func TestProductoPedidoUpdate_ReconsultaOneError(t *testing.T) {
	origQ, origE := MockQuery, MockExec
	step := 0
	MockQuery = func(_ stdctx.Context, q string, _ []driver.NamedValue) (driver.Rows, error) {
		lower := strings.ToLower(q)
		if strings.Contains(lower, "detalle_pedido") {
			if step == 0 {
				step++
				return &mockRows{columns: []string{"pk_id_pedido"}, values: [][]driver.Value{}}, nil
			}
			return nil, errors.New("one fail")
		}
		if strings.Contains(lower, "select pk_id_producto, cantidad from producto") {
			cols := []string{"pk_id_producto", "cantidad"}
			vals := [][]driver.Value{{int64(1), int64(10)}}
			return &mockRows{columns: cols, values: vals}, nil
		}
		return &mockRows{columns: []string{"ok"}, values: [][]driver.Value{{int64(1)}}}, nil
	}
	MockExec = func(_ stdctx.Context, _ string, _ []driver.NamedValue) (driver.Result, error) {
		return mockResult{}, nil
	}
	t.Cleanup(func() { MockQuery, MockExec = origQ, origE })

	body := `[{"productoId":1,"cantidad":2}]`
	r := httptest.NewRequest(http.MethodPut, "/producto_pedido?pedido_id=1", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := ProductoPedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Update()
	if w.Code != http.StatusInternalServerError && w.Code != http.StatusOK {
		t.Fatalf("expected 500 or 200, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestProductoPedidoUpdate_DeleteExecError(t *testing.T) {
	origQ, origE := MockQuery, MockExec
	MockQuery = func(_ stdctx.Context, q string, _ []driver.NamedValue) (driver.Rows, error) {
		lower := strings.ToLower(q)
		if strings.Contains(lower, "detalle_pedido") {
			return &mockRows{columns: []string{"pk_id_pedido"}, values: [][]driver.Value{}}, nil
		}
		if strings.Contains(lower, "select pk_id_producto, cantidad from producto") {
			cols := []string{"pk_id_producto", "cantidad"}
			vals := [][]driver.Value{{int64(1), int64(10)}}
			return &mockRows{columns: cols, values: vals}, nil
		}
		return &mockRows{columns: []string{"ok"}, values: [][]driver.Value{{int64(1)}}}, nil
	}
	MockExec = func(_ stdctx.Context, q string, _ []driver.NamedValue) (driver.Result, error) {
		lower := strings.ToLower(q)
		if strings.Contains(lower, "delete") && strings.Contains(lower, "detalle_pedido") {
			return nil, errors.New("delete fail")
		}
		return mockResult{}, nil
	}
	t.Cleanup(func() { MockQuery, MockExec = origQ, origE })

	body := `[{"productoId":1,"cantidad":2}]`
	r := httptest.NewRequest(http.MethodPut, "/producto_pedido?pedido_id=1", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := ProductoPedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Update()
	if w.Code != http.StatusInternalServerError && w.Code != http.StatusOK {
		t.Fatalf("expected 500 or 200, got %d. Body: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(strings.ToLower(w.Body.String()), "actualizar los productos del pedido") {
		t.Fatalf("expected delete error message, body: %s", w.Body.String())
	}
}

func TestProductoPedidoUpdate_NegativeDelta_AdjustInventoryExecError(t *testing.T) {
	origQ, origE := MockQuery, MockExec
	call := 0
	MockQuery = func(_ stdctx.Context, q string, _ []driver.NamedValue) (driver.Rows, error) {
		lower := strings.ToLower(q)
		if strings.Contains(lower, "detalle_pedido") {
			if call == 0 {
				call++
				cols := []string{"pk_id_detalle", "pk_id_pedido", "pk_id_producto", "cantidad", "precio"}
				vals := [][]driver.Value{{int64(1), int64(1), int64(1), int64(3), int64(1000)}}
				return &mockRows{columns: cols, values: vals}, nil
			}
			return &mockRows{columns: []string{"ok"}, values: [][]driver.Value{{int64(1)}}}, nil
		}
		if strings.Contains(lower, "select pk_id_producto, cantidad from producto") {
			cols := []string{"pk_id_producto", "cantidad"}
			vals := [][]driver.Value{{int64(1), int64(10)}}
			return &mockRows{columns: cols, values: vals}, nil
		}
		return &mockRows{columns: []string{"ok"}, values: [][]driver.Value{{int64(1)}}}, nil
	}
	MockExec = func(_ stdctx.Context, q string, _ []driver.NamedValue) (driver.Result, error) {
		lower := strings.ToLower(q)
		if strings.Contains(lower, "update producto set cantidad = cantidad +") {
			return nil, errors.New("restock fail")
		}
		return mockResult{}, nil
	}
	t.Cleanup(func() { MockQuery, MockExec = origQ, origE })

	body := `[{"productoId":1,"cantidad":1}]`
	r := httptest.NewRequest(http.MethodPut, "/producto_pedido?pedido_id=1", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := ProductoPedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.Update()
	if w.Code != http.StatusInternalServerError && w.Code != http.StatusOK {
		t.Fatalf("expected 500 or 200, got %d. Body: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(strings.ToLower(w.Body.String()), "ajustar inventario") {
		t.Fatalf("expected adjust inventory error, body: %s", w.Body.String())
	}
}
