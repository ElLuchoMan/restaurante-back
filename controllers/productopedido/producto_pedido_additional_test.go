package productopedido

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"restaurante/models"

	"github.com/beego/beego/v2/server/web/context"
)

func TestProductoPedido_Post_BadJSON(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/producto_pedido", strings.NewReader("notjson"))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte("notjson")
	c := ProductoPedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()
	// Puede devolver 400 directo o 200 con ApiResponse{Code:400}
	if w.Code == http.StatusBadRequest {
		return
	}
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200/400, got %d", w.Code)
	}
	var resp models.ApiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid json response: %v", err)
	}
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected ApiResponse.Code=400, got %d", resp.Code)
	}
}

func TestProductoPedido_Post_ValidationError(t *testing.T) {
	body := `{"pedidoId":1,"detalles":[]}`
	r := httptest.NewRequest(http.MethodPost, "/producto_pedido", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := ProductoPedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()
	if w.Code == http.StatusBadRequest {
		return
	}
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200/400, got %d", w.Code)
	}
	var resp models.ApiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid json response: %v", err)
	}
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected ApiResponse.Code=400, got %d", resp.Code)
	}
}

func TestProductoPedido_Update_BadRequest_NoPedidoID(t *testing.T) {
	r := httptest.NewRequest(http.MethodPut, "/producto_pedido", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ProductoPedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Update()
	if w.Code == http.StatusBadRequest {
		return
	}
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200/400, got %d", w.Code)
	}
	var resp models.ApiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid json response: %v", err)
	}
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected ApiResponse.Code=400, got %d", resp.Code)
	}
}

func TestProductoPedido_Update_BadJSON(t *testing.T) {
	r := httptest.NewRequest(http.MethodPut, "/producto_pedido?pedido_id=1", strings.NewReader("notjson"))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte("notjson")
	c := ProductoPedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Update()
	if w.Code == http.StatusBadRequest {
		return
	}
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200/400, got %d", w.Code)
	}
	var resp models.ApiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid json response: %v", err)
	}
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected ApiResponse.Code=400, got %d", resp.Code)
	}
}

func TestProductoPedido_Update_EmptyList(t *testing.T) {
	body := `[]`
	r := httptest.NewRequest(http.MethodPut, "/producto_pedido?pedido_id=1", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := ProductoPedidoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Update()
	if w.Code == http.StatusBadRequest {
		return
	}
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200/400, got %d", w.Code)
	}
	var resp models.ApiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid json response: %v", err)
	}
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected ApiResponse.Code=400, got %d", resp.Code)
	}
}
