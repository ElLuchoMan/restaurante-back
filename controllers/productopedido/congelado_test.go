package productopedido

import (
	"database/sql/driver"
	"net/http"
	"strings"
	"testing"
)

// escrituras registra las sentencias de escritura y, de UPDATE pago, el monto.
type escrituras struct {
	total      int
	pagoUpdate bool
	monto      int64
}

func registrar(t *testing.T) *escrituras {
	t.Helper()
	e := &escrituras{}
	fakeExec = func(q string, a []driver.NamedValue) (driver.Result, error) {
		e.total++
		if strings.HasPrefix(q, "UPDATE pago") {
			e.pagoUpdate = true
			e.monto = a[0].Value.(int64)
		}
		return fakeResult{}, nil
	}
	return e
}

// Ataque: un Cliente que ya asignó su pago (monto fijado por el servidor) intenta
// cambiar los productos para bajar o subir lo que ya pagó: 409 sin tocar nada.
func TestClienteNoModificaTrasAsignarPago(t *testing.T) {
	defer resetFake()
	p := func(c *ProductoPedidoController) { c.Post() }
	u := func(c *ProductoPedidoController) { c.Update() }
	for _, estado := range []string{"PENDIENTE", "PAGADO"} {
		w := newWorld()
		w.pagoID, w.pagoEstado = 4, estado
		como(t, rolClienteT, 1001)
		e := registrar(t)
		b := call(t, http.MethodPost, "/producto_pedido", bodyPost, p, http.StatusConflict)
		contains(t, b, "pago asignado")
		call(t, http.MethodPut, urlDetalle, bodyPut, u, http.StatusConflict)
		call(t, http.MethodPut, urlDetalle, `[{"productoId":1,"cantidad":0}]`, u, http.StatusConflict)
		if e.total != 0 {
			t.Fatalf("un pedido congelado no debe escribir nada (%d escrituras)", e.total)
		}
	}
	// el pedido ajeno sigue siendo 404 (no revela que está congelado)
	w := newWorld()
	w.pagoID = 4
	como(t, rolClienteT, 2002)
	call(t, http.MethodPost, "/producto_pedido", bodyPost, p, http.StatusNotFound)
}

func TestPersonalRecalculaMontoDelPago(t *testing.T) {
	defer resetFake()
	p := func(c *ProductoPedidoController) { c.Post() }
	u := func(c *ProductoPedidoController) { c.Update() }
	for _, estado := range []string{"PENDIENTE", "NO_PAGO"} {
		w := newWorld()
		w.pagoID, w.pagoEstado, w.subtotal, w.lineas = 4, estado, 75000, 3
		como(t, rolMesero, 5)
		e := registrar(t)
		call(t, http.MethodPost, "/producto_pedido", bodyPost, p, http.StatusCreated)
		if !e.pagoUpdate || e.monto != 75000 {
			t.Fatalf("POST debe recalcular el monto del pago (monto=%d, update=%v)", e.monto, e.pagoUpdate)
		}
		w.subtotal, w.lineas = 25000, 1
		e = registrar(t)
		call(t, http.MethodPut, urlDetalle, bodyPut, u, http.StatusOK)
		if !e.pagoUpdate || e.monto != 25000 {
			t.Fatalf("PUT debe recalcular el monto del pago (monto=%d)", e.monto)
		}
	}
	// sin pago asignado no hay nada que recalcular
	newWorld()
	e := registrar(t)
	call(t, http.MethodPut, urlDetalle, bodyPut, u, http.StatusOK)
	if e.pagoUpdate {
		t.Fatal("sin pago no se actualiza ningún pago")
	}
}

func TestPersonalBloqueadoConPagoPagadoODescuento(t *testing.T) {
	defer resetFake()
	p := func(c *ProductoPedidoController) { c.Post() }
	u := func(c *ProductoPedidoController) { c.Update() }
	como(t, rolMesero, 5)

	w := newWorld()
	w.pagoID, w.pagoEstado = 4, "PAGADO"
	e := registrar(t)
	contains(t, call(t, http.MethodPost, "/producto_pedido", bodyPost, p, http.StatusConflict), "PAGADO")
	call(t, http.MethodPut, urlDetalle, bodyPut, u, http.StatusConflict)

	w.pagoEstado, w.descuentos = "PENDIENTE", 1
	contains(t, call(t, http.MethodPost, "/producto_pedido", bodyPost, p, http.StatusConflict), "descuento")
	call(t, http.MethodPut, urlDetalle, bodyPut, u, http.StatusConflict)
	if e.total != 0 {
		t.Fatalf("no debe escribir nada (%d escrituras)", e.total)
	}
}

func TestErroresAlValidarOReCalcularElPago(t *testing.T) {
	defer resetFake()
	p := func(c *ProductoPedidoController) { c.Post() }
	como(t, rolMesero, 5)
	for name, set := range map[string]func(w *world){
		"lock pago":  func(w *world) { w.pagoErr = errBoom },
		"descuentos": func(w *world) { w.descErr = errBoom },
		"monto":      func(w *world) { w.montoErr = errBoom },
	} {
		w := newWorld()
		w.pagoID, w.pagoEstado = 4, "PENDIENTE"
		set(w)
		t.Log(name)
		call(t, http.MethodPost, "/producto_pedido", bodyPost, p, http.StatusInternalServerError)
	}
	// falla el UPDATE del pago
	w := newWorld()
	w.pagoID, w.pagoEstado = 4, "PENDIENTE"
	fakeExec = func(q string, _ []driver.NamedValue) (driver.Result, error) {
		if strings.HasPrefix(q, "UPDATE pago") {
			return nil, errBoom
		}
		return fakeResult{}, nil
	}
	call(t, http.MethodPost, "/producto_pedido", bodyPost, p, http.StatusInternalServerError)
}
