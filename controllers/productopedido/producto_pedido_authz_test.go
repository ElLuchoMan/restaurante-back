package productopedido

import (
	"database/sql/driver"
	"net/http"
	"strings"
	"testing"
)

const (
	urlDetalle = "/producto_pedido?pedido_id=10"
	bodyPost   = `{"pedidoId":10,"detalles":[{"productoId":1,"cantidad":2}]}`
	bodyPut    = `[{"productoId":1,"cantidad":2}]`
)

func TestSinTokenEs401(t *testing.T) {
	defer resetFake()
	newWorld()
	sinToken(t)
	call(t, http.MethodGet, urlDetalle, "", func(c *ProductoPedidoController) { c.GetAll() }, http.StatusUnauthorized)
	call(t, http.MethodPost, "/producto_pedido", bodyPost, func(c *ProductoPedidoController) { c.Post() }, http.StatusUnauthorized)
	call(t, http.MethodPut, urlDetalle, bodyPut, func(c *ProductoPedidoController) { c.Update() }, http.StatusUnauthorized)
}

func TestGetAllCliente(t *testing.T) {
	defer resetFake()
	g := func(c *ProductoPedidoController) { c.GetAll() }
	w := newWorld()
	w.actuales[1] = 2
	como(t, rolClienteT, 1001)
	call(t, http.MethodGet, urlDetalle, "", g, http.StatusOK)
	// pedido de otro cliente: igual que inexistente, sin revelar las líneas
	como(t, rolClienteT, 2002)
	if b := call(t, http.MethodGet, urlDetalle, "", g, http.StatusNotFound); !strings.Contains(b, "Pedido no encontrado") || strings.Contains(b, "precio") {
		t.Fatalf("ajeno debe verse como inexistente: %s", b)
	}
	como(t, rolClienteT, 0)
	call(t, http.MethodGet, urlDetalle, "", g, http.StatusNotFound)
	// error al comprobar el dueño
	como(t, rolClienteT, 1001)
	base := fakeQuery
	fakeQuery = func(q string, a []driver.NamedValue) (driver.Rows, error) {
		if strings.Contains(q, "COUNT(*)") {
			return nil, errBoom
		}
		return base(q, a)
	}
	call(t, http.MethodGet, urlDetalle, "", g, http.StatusInternalServerError)
	// el personal lee cualquiera
	como(t, rolMesero, 5)
	fakeQuery = base
	call(t, http.MethodGet, urlDetalle, "", g, http.StatusOK)
}

func TestEscrituraCliente(t *testing.T) {
	defer resetFake()
	p := func(c *ProductoPedidoController) { c.Post() }
	u := func(c *ProductoPedidoController) { c.Update() }
	w := newWorld()
	// su propio pedido
	como(t, rolClienteT, 1001)
	call(t, http.MethodPost, "/producto_pedido", bodyPost, p, http.StatusCreated)
	call(t, http.MethodPut, urlDetalle, bodyPut, u, http.StatusOK)
	// pedido ajeno: 404 sin tocar el inventario ni las líneas
	var escribio bool
	fakeExec = func(string, []driver.NamedValue) (driver.Result, error) {
		escribio = true
		return fakeResult{}, nil
	}
	como(t, rolClienteT, 2002)
	call(t, http.MethodPost, "/producto_pedido", bodyPost, p, http.StatusNotFound)
	call(t, http.MethodPut, urlDetalle, bodyPut, u, http.StatusNotFound)
	// un pedido sin cliente solo lo gestiona el personal
	w.dueno = 0
	como(t, rolClienteT, 0)
	call(t, http.MethodPost, "/producto_pedido", bodyPost, p, http.StatusNotFound)
	if escribio {
		t.Fatal("un pedido ajeno no debe modificarse")
	}
	como(t, rolMesero, 5)
	call(t, http.MethodPost, "/producto_pedido", bodyPost, p, http.StatusCreated)
	call(t, http.MethodPut, urlDetalle, bodyPut, u, http.StatusOK)
	if !escribio {
		t.Fatal("el personal sí debe poder escribir")
	}
}
