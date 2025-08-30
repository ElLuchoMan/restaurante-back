package controllers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
	"restaurante/models"
)

type fakeQueryPP struct {
	orm.QuerySeter
	one func(interface{}, ...string) error
	all func(interface{}, ...string) (int64, error)
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
	if !strings.Contains(w.Body.String(), "Error al crear el pedido con productos") {
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

func TestProductoPedidoGetAllSuccess(t *testing.T) {
	original := productoPedidoNewOrm
	productoPedidoNewOrm = func() productoPedidoOrmer {
		return fakeOrmerPP{query: func(i interface{}) orm.QuerySeter {
			switch i.(type) {
			case *models.DetallePedido:
				return fakeQueryPP{all: func(res interface{}, cols ...string) (int64, error) {
					detalles := res.(*[]models.DetallePedido)
					*detalles = append(*detalles, models.DetallePedido{PKIDPedido: 1, PKIDProducto: 1, Cantidad: 1, Precio: 1000})
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
	if !strings.Contains(w.Body.String(), "detalles") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestProductoPedidoPostSuccess(t *testing.T) {
	original := productoPedidoNewOrm
	productoPedidoNewOrm = func() productoPedidoOrmer {
		return fakeOrmerPP{insert: func(m interface{}) (int64, error) {
			if d, ok := m.(*models.DetallePedido); ok {
				d.Precio = 1000
			}
			return 1, nil
		}}
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
	if !strings.Contains(w.Body.String(), "\"precio\":1000") {
		t.Errorf("expected price in response, got %s", w.Body.String())
	}
}
