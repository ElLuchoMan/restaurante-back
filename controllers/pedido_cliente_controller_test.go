package controllers

import (
	stdctx "context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
)

func TestPedidoClienteGetAllWithoutDB(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/pedido_clientes", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := PedidoClienteController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetAll()

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al obtener las relaciones de la base de datos") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestPedidoClienteGetAllSuccess(t *testing.T) {
	origNewOrm := newOrm
	origQueryAll := pcQueryAll
	defer func() {
		newOrm = origNewOrm
		pcQueryAll = origQueryAll
	}()
	newOrm = func() orm.Ormer { return nil }
	pcQueryAll = func(o orm.Ormer, relaciones *[]models.PedidoCliente) (int64, error) {
		*relaciones = []models.PedidoCliente{{PK_ID_PEDIDO_CLIENTE: 1}}
		return 1, nil
	}

	r := httptest.NewRequest(http.MethodGet, "/pedido_clientes", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := PedidoClienteController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetAll()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Relaciones obtenidas exitosamente") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestPedidoClientePostInvalidJSON(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/pedido_clientes", strings.NewReader("{"))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := PedidoClienteController{}
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

func TestPedidoClientePostMissingFields(t *testing.T) {
	body := `{"documentoCliente":0,"pedidoId":0}`
	r := httptest.NewRequest(http.MethodPost, "/pedido_clientes", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := PedidoClienteController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Faltan campos obligatorios") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestPedidoClientePostClienteNoEncontrado(t *testing.T) {
	body := `{"documentoCliente":1,"pedidoId":1}`
	r := httptest.NewRequest(http.MethodPost, "/pedido_clientes", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := PedidoClienteController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	origDoTx := pcDoTx
	origReadCliente := pcReadCliente
	defer func() {
		pcDoTx = origDoTx
		pcReadCliente = origReadCliente
	}()
	pcDoTx = func(o orm.Ormer, f func(stdctx.Context, orm.TxOrmer) error) error {
		return f(stdctx.Background(), nil)
	}
	pcReadCliente = func(txOrm orm.TxOrmer, cliente *models.Cliente) error { return errors.New("not found") }

	c.Post()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestPedidoClientePostPedidoNoEncontrado(t *testing.T) {
	body := `{"documentoCliente":1,"pedidoId":2}`
	r := httptest.NewRequest(http.MethodPost, "/pedido_clientes", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := PedidoClienteController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	origNewOrm := newOrm
	origDoTx := pcDoTx
	origReadCliente := pcReadCliente
	origReadPedido := pcReadPedido
	defer func() {
		newOrm = origNewOrm
		pcDoTx = origDoTx
		pcReadCliente = origReadCliente
		pcReadPedido = origReadPedido
	}()
	newOrm = func() orm.Ormer { return nil }
	pcDoTx = func(o orm.Ormer, f func(stdctx.Context, orm.TxOrmer) error) error {
		return f(stdctx.Background(), nil)
	}
	pcReadCliente = func(txOrm orm.TxOrmer, cliente *models.Cliente) error { return nil }
	pcReadPedido = func(txOrm orm.TxOrmer, pedido *models.Pedido) error { return errors.New("not found") }

	c.Post()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Pedido no encontrado") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestPedidoClientePostPedidoYaAsignado(t *testing.T) {
	body := `{"documentoCliente":1,"pedidoId":2}`
	r := httptest.NewRequest(http.MethodPost, "/pedido_clientes", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := PedidoClienteController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	origNewOrm := newOrm
	origDoTx := pcDoTx
	origReadCliente := pcReadCliente
	origReadPedido := pcReadPedido
	origCheck := pcCheckExistingPedidoCliente
	defer func() {
		newOrm = origNewOrm
		pcDoTx = origDoTx
		pcReadCliente = origReadCliente
		pcReadPedido = origReadPedido
		pcCheckExistingPedidoCliente = origCheck
	}()
	newOrm = func() orm.Ormer { return nil }
	pcDoTx = func(o orm.Ormer, f func(stdctx.Context, orm.TxOrmer) error) error {
		return f(stdctx.Background(), nil)
	}
	pcReadCliente = func(txOrm orm.TxOrmer, cliente *models.Cliente) error { return nil }
	pcReadPedido = func(txOrm orm.TxOrmer, pedido *models.Pedido) error { return nil }
	pcCheckExistingPedidoCliente = func(txOrm orm.TxOrmer, pedidoID int, rel *models.PedidoCliente) error { return nil }

	c.Post()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "El pedido ya pertenece a otro cliente") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestPedidoClientePostInsertError(t *testing.T) {
	body := `{"documentoCliente":1,"pedidoId":2}`
	r := httptest.NewRequest(http.MethodPost, "/pedido_clientes", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := PedidoClienteController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	origNewOrm := newOrm
	origDoTx := pcDoTx
	origReadCliente := pcReadCliente
	origReadPedido := pcReadPedido
	origCheck := pcCheckExistingPedidoCliente
	origInsert := pcInsertPedidoCliente
	defer func() {
		newOrm = origNewOrm
		pcDoTx = origDoTx
		pcReadCliente = origReadCliente
		pcReadPedido = origReadPedido
		pcCheckExistingPedidoCliente = origCheck
		pcInsertPedidoCliente = origInsert
	}()
	newOrm = func() orm.Ormer { return nil }
	pcDoTx = func(o orm.Ormer, f func(stdctx.Context, orm.TxOrmer) error) error {
		return f(stdctx.Background(), nil)
	}
	pcReadCliente = func(txOrm orm.TxOrmer, cliente *models.Cliente) error { return nil }
	pcReadPedido = func(txOrm orm.TxOrmer, pedido *models.Pedido) error { return nil }
	pcCheckExistingPedidoCliente = func(txOrm orm.TxOrmer, pedidoID int, rel *models.PedidoCliente) error {
		return errors.New("no existe")
	}
	pcInsertPedidoCliente = func(txOrm orm.TxOrmer, rel *models.PedidoCliente) (int64, error) {
		return 0, errors.New("insert error")
	}

	c.Post()

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al crear la relación") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestPedidoClientePostSuccess(t *testing.T) {
	body := `{"documentoCliente":1,"pedidoId":2}`
	r := httptest.NewRequest(http.MethodPost, "/pedido_clientes", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := PedidoClienteController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	origNewOrm := newOrm
	origDoTx := pcDoTx
	origReadCliente := pcReadCliente
	origReadPedido := pcReadPedido
	origCheck := pcCheckExistingPedidoCliente
	origInsert := pcInsertPedidoCliente
	defer func() {
		newOrm = origNewOrm
		pcDoTx = origDoTx
		pcReadCliente = origReadCliente
		pcReadPedido = origReadPedido
		pcCheckExistingPedidoCliente = origCheck
		pcInsertPedidoCliente = origInsert
	}()
	newOrm = func() orm.Ormer { return nil }
	pcDoTx = func(o orm.Ormer, f func(stdctx.Context, orm.TxOrmer) error) error {
		return f(stdctx.Background(), nil)
	}
	pcReadCliente = func(txOrm orm.TxOrmer, cliente *models.Cliente) error { return nil }
	pcReadPedido = func(txOrm orm.TxOrmer, pedido *models.Pedido) error { return nil }
	pcCheckExistingPedidoCliente = func(txOrm orm.TxOrmer, pedidoID int, rel *models.PedidoCliente) error {
		return errors.New("no existe")
	}
	pcInsertPedidoCliente = func(txOrm orm.TxOrmer, rel *models.PedidoCliente) (int64, error) {
		return 10, nil
	}

	c.Post()

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", w.Code)
	}
	var resp models.ApiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Code != http.StatusCreated {
		t.Fatalf("expected response code 201, got %d", resp.Code)
	}
	relacion := resp.Data.(map[string]interface{})
	if int(relacion["pedidoClienteId"].(float64)) != 10 {
		t.Errorf("expected id 10, got %v", relacion["pedidoClienteId"])
	}
}

func TestPedidoClienteHelperDefaults(t *testing.T) {
	o := orm.NewOrm()
	tx, _ := o.Begin()
	fns := []func(){
		func() { pcQueryAll(o, &[]models.PedidoCliente{}) },
		func() { pcDoTx(o, func(stdctx.Context, orm.TxOrmer) error { return nil }) },
		func() { pcReadCliente(tx, &models.Cliente{}) },
		func() { pcReadPedido(tx, &models.Pedido{}) },
		func() { pcCheckExistingPedidoCliente(tx, 0, &models.PedidoCliente{}) },
		func() { pcInsertPedidoCliente(tx, &models.PedidoCliente{}) },
	}
	for _, fn := range fns {
		func() {
			defer func() { _ = recover() }()
			fn()
		}()
	}
}
