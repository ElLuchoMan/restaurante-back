package productopedido

import (
	"database/sql/driver"
	"errors"
	"net/http"
	"strings"
	"testing"
)

var (
	errBoom     = errors.New("boom")
	detalleCols = []string{"pk_id_detalle", "pk_id_pedido", "pk_id_producto", "precio", "cantidad"}
	prodCols    = []string{"pk_id_producto", "cantidad"}
	countCols   = []string{"count"}
	lockCols    = []string{"documento", "pago"}
)

// stock programa el inventario por producto; actuales, las líneas que ya tiene
// el pedido (producto -> cantidad); pedido, si existe.
type world struct {
	pedido   bool
	dueno    int64 // documento del cliente del pedido (0 = sin cliente)
	stock    map[int64]int64
	actuales map[int64]int64
	lockErr  error
	listErr  error

	// pago asignado al pedido (0 = sin pago) y su estado; descuentos aplicados
	pagoID, descuentos int64
	pagoEstado         string
	// subtotal y lineas es lo que devuelve la suma del detalle (recálculo del monto)
	subtotal, lineas int64
	pagoErr          error // error al bloquear el pago
	descErr          error // error al contar descuentos
	montoErr         error // error al calcular el monto
}

func (w *world) serve() {
	fakeQuery = func(q string, args []driver.NamedValue) (driver.Rows, error) {
		switch {
		case strings.Contains(q, "FROM pedido WHERE") && strings.Contains(q, "FOR UPDATE"):
			if w.lockErr != nil {
				return nil, w.lockErr
			}
			if !w.pedido {
				return rowsOf(lockCols), nil
			}
			return rowsOf(lockCols, []driver.Value{w.dueno, w.pagoID}), nil
		case strings.Contains(q, "FROM pago WHERE"):
			if w.pagoErr != nil {
				return nil, w.pagoErr
			}
			return rowsOf([]string{"estado_pago"}, []driver.Value{w.pagoEstado}), nil
		case strings.Contains(q, "pedido_descuento_aplicado d JOIN"):
			if w.descErr != nil {
				return nil, w.descErr
			}
			return rowsOf(countCols, []driver.Value{w.descuentos}), nil
		case strings.Contains(q, "SUM(precio * cantidad)"):
			if w.montoErr != nil {
				return nil, w.montoErr
			}
			return rowsOf([]string{"subtotal", "lineas"}, []driver.Value{w.subtotal, w.lineas}), nil
		case strings.Contains(q, "SUM(monto_descuento)"):
			return rowsOf([]string{"descuento"}, []driver.Value{int64(0)}), nil
		case strings.Contains(q, "FROM producto"):
			var vals [][]driver.Value
			for _, a := range args {
				if s, ok := w.stock[a.Value.(int64)]; ok {
					vals = append(vals, []driver.Value{a.Value, s})
				}
			}
			return rowsOf(prodCols, vals...), nil
		case strings.Contains(q, "COUNT(*)"):
			// la comprobación de pertenencia de un Cliente trae el documento como segundo argumento
			if w.pedido && (len(args) < 2 || args[1].Value == w.dueno) {
				return rowsOf(countCols, []driver.Value{int64(1)}), nil
			}
			return rowsOf(countCols, []driver.Value{int64(0)}), nil
		case strings.Contains(q, `FROM "detalle_pedido"`):
			if w.listErr != nil {
				return nil, w.listErr
			}
			if len(args) == 2 { // relectura de la línea recién insertada
				return rowsOf(detalleCols, []driver.Value{int64(1), args[0].Value, args[1].Value, int64(25000), int64(2)}), nil
			}
			var vals [][]driver.Value
			for pid, qty := range w.actuales {
				vals = append(vals, []driver.Value{int64(pid), int64(10), pid, int64(25000), qty})
			}
			return rowsOf(detalleCols, vals...), nil
		}
		return rowsOf(nil), nil
	}
}

func newWorld() *world {
	w := &world{pedido: true, dueno: 1001, pagoEstado: "PENDIENTE", subtotal: 50000, lineas: 2, stock: map[int64]int64{1: 10, 2: 10, 3: 1}, actuales: map[int64]int64{}}
	w.serve()
	return w
}

func call(t *testing.T, method, target, body string, f func(c *ProductoPedidoController), status int) string {
	t.Helper()
	ctx, w := newCtx(method, target, body)
	c := &ProductoPedidoController{}
	c.Ctx, c.Data = ctx, map[interface{}]interface{}{}
	f(c)
	expect(t, w, status)
	return w.Body.String()
}

func contains(t *testing.T, body string, wants ...string) {
	t.Helper()
	for _, w := range wants {
		if !strings.Contains(body, w) {
			t.Fatalf("falta %s en %s", w, body)
		}
	}
}

func TestGetAll(t *testing.T) {
	defer resetFake()
	g := func(c *ProductoPedidoController) { c.GetAll() }
	for _, q := range []string{"", "pedido_id=0", "pedido_id=x"} {
		call(t, http.MethodGet, "/producto_pedido?"+q, "", g, http.StatusBadRequest)
	}
	w := newWorld()
	w.actuales[1] = 2
	b := call(t, http.MethodGet, "/producto_pedido?pedido_id=10", "", g, http.StatusOK)
	contains(t, b, `"pedidoId":10`, `"detalles":[{`, `"precio":25000`)
	w.actuales = map[int64]int64{}
	b = call(t, http.MethodGet, "/producto_pedido?pedido_id=10", "", g, http.StatusOK)
	contains(t, b, `"detalles":[]`)
	w.pedido = false
	call(t, http.MethodGet, "/producto_pedido?pedido_id=10", "", g, http.StatusNotFound)
	w.listErr = errBoom
	call(t, http.MethodGet, "/producto_pedido?pedido_id=10", "", g, http.StatusInternalServerError)
	w.listErr = nil
	base := fakeQuery
	fakeQuery = func(q string, a []driver.NamedValue) (driver.Rows, error) {
		if strings.Contains(q, "COUNT(*)") {
			return nil, errBoom
		}
		return base(q, a)
	}
	call(t, http.MethodGet, "/producto_pedido?pedido_id=10", "", g, http.StatusInternalServerError)
}

func TestPostValidation(t *testing.T) {
	defer resetFake()
	newWorld()
	p := func(c *ProductoPedidoController) { c.Post() }
	for _, body := range []string{"", "{", `{"pedidoId":"x"}`, `{}`, `{"pedidoId":0,"detalles":[{"productoId":1,"cantidad":1}]}`,
		`{"pedidoId":-1,"detalles":[{"productoId":1,"cantidad":1}]}`, `{"pedidoId":10,"detalles":[]}`,
		`{"pedidoId":10,"detalles":[{"productoId":0,"cantidad":1}]}`, `{"pedidoId":10,"detalles":[{"productoId":1,"cantidad":-1}]}`,
		`{"pedidoId":10,"detalles":[{"productoId":1,"cantidad":0}]}`} {
		call(t, http.MethodPost, "/producto_pedido", body, p, http.StatusBadRequest)
	}
}

func TestPostSuccess(t *testing.T) {
	defer resetFake()
	newWorld()
	p := func(c *ProductoPedidoController) { c.Post() }
	var stockUpdates []driver.NamedValue
	fakeExec = func(q string, a []driver.NamedValue) (driver.Result, error) {
		if strings.HasPrefix(q, "UPDATE producto") {
			stockUpdates = append(stockUpdates, a...)
		}
		return fakeResult{}, nil
	}
	// líneas repetidas se suman y las de cantidad 0 se ignoran
	b := call(t, http.MethodPost, "/producto_pedido", `{"pedidoId":10,"detalles":[{"productoId":2,"cantidad":1},{"productoId":1,"cantidad":2},{"productoId":2,"cantidad":2},{"productoId":3,"cantidad":0}]}`, p, http.StatusCreated)
	contains(t, b, `"pedidoId":10`, `"detalles":[{`, `"precio":25000`)
	if len(stockUpdates) != 4 || stockUpdates[0].Value != int64(2) || stockUpdates[2].Value != int64(3) {
		t.Fatalf("descuentos inesperados: %v", stockUpdates)
	}
}

func TestPostErrors(t *testing.T) {
	defer resetFake()
	w := newWorld()
	p := func(c *ProductoPedidoController) { c.Post() }
	body := `{"pedidoId":10,"detalles":[{"productoId":1,"cantidad":2}]}`

	fakeBeginErr = errBoom
	call(t, http.MethodPost, "/producto_pedido", body, p, http.StatusInternalServerError)
	fakeBeginErr = nil

	w.pedido = false
	call(t, http.MethodPost, "/producto_pedido", body, p, http.StatusNotFound)
	w.pedido = true
	w.lockErr = errBoom
	call(t, http.MethodPost, "/producto_pedido", body, p, http.StatusInternalServerError)
	w.lockErr = nil

	// producto inexistente -> 404; inventario insuficiente -> 409 con detalle
	call(t, http.MethodPost, "/producto_pedido", `{"pedidoId":10,"detalles":[{"productoId":99,"cantidad":1}]}`, p, http.StatusNotFound)
	b := call(t, http.MethodPost, "/producto_pedido", `{"pedidoId":10,"detalles":[{"productoId":3,"cantidad":5},{"productoId":1,"cantidad":1}]}`, p, http.StatusConflict)
	contains(t, b, `"productoId":3`, `"requerido":5`, `"disponible":1`)

	base := fakeQuery
	fakeQuery = func(q string, a []driver.NamedValue) (driver.Rows, error) {
		if strings.Contains(q, "FROM producto") {
			return nil, errBoom
		}
		return base(q, a)
	}
	call(t, http.MethodPost, "/producto_pedido", body, p, http.StatusInternalServerError)
	fakeQuery = base

	for _, prefix := range []string{"UPDATE producto", "INSERT"} {
		fakeExec = func(q string, _ []driver.NamedValue) (driver.Result, error) {
			if strings.HasPrefix(q, prefix) {
				return nil, errBoom
			}
			return fakeResult{}, nil
		}
		call(t, http.MethodPost, "/producto_pedido", body, p, http.StatusInternalServerError)
	}
	// producto ya presente en el pedido -> 409
	fakeExec = func(q string, _ []driver.NamedValue) (driver.Result, error) {
		if strings.HasPrefix(q, "INSERT") {
			return nil, errors.New(`duplicate key value violates unique constraint (SQLSTATE 23505)`)
		}
		return fakeResult{}, nil
	}
	call(t, http.MethodPost, "/producto_pedido", body, p, http.StatusConflict)
	fakeExec = nil

	// la relectura de la línea insertada falla
	fakeQuery = func(q string, a []driver.NamedValue) (driver.Rows, error) {
		if strings.Contains(q, `FROM "detalle_pedido"`) {
			return nil, errBoom
		}
		return base(q, a)
	}
	call(t, http.MethodPost, "/producto_pedido", body, p, http.StatusInternalServerError)
	fakeQuery = base

	fakeCommitErr = errBoom
	call(t, http.MethodPost, "/producto_pedido", body, p, http.StatusInternalServerError)
}

func TestUpdateValidation(t *testing.T) {
	defer resetFake()
	newWorld()
	u := func(c *ProductoPedidoController) { c.Update() }
	call(t, http.MethodPut, "/producto_pedido", `[{"productoId":1,"cantidad":1}]`, u, http.StatusBadRequest)
	call(t, http.MethodPut, "/producto_pedido?pedido_id=0", `[{"productoId":1,"cantidad":1}]`, u, http.StatusBadRequest)
	for _, body := range []string{"", "{", `{"productoId":1}`, `[]`, `[{"productoId":0,"cantidad":1}]`, `[{"productoId":1,"cantidad":-2}]`} {
		call(t, http.MethodPut, "/producto_pedido?pedido_id=10", body, u, http.StatusBadRequest)
	}
}

func TestUpdateSuccess(t *testing.T) {
	defer resetFake()
	w := newWorld()
	u := func(c *ProductoPedidoController) { c.Update() }
	// el pedido tiene 3 de producto 1 y 2 de producto 2; se pasa a 5 de 1 (descuenta 2),
	// se quita el 2 con cantidad 0 (devuelve 2) y se agrega el 3 (descuenta 1)
	w.actuales = map[int64]int64{1: 3, 2: 2}
	var updates [][2]int64 // {delta, producto}
	fakeExec = func(q string, a []driver.NamedValue) (driver.Result, error) {
		if strings.HasPrefix(q, "UPDATE producto") {
			updates = append(updates, [2]int64{a[0].Value.(int64), a[1].Value.(int64)})
		}
		return fakeResult{}, nil
	}
	b := call(t, http.MethodPut, "/producto_pedido?pedido_id=10", `[{"productoId":1,"cantidad":5},{"productoId":2,"cantidad":0},{"productoId":3,"cantidad":1}]`, u, http.StatusOK)
	contains(t, b, `"pedidoId":10`, `"detalles":[{`)
	if want := [][2]int64{{2, 1}, {-2, 2}, {1, 3}}; len(updates) != 3 || updates[0] != want[0] || updates[1] != want[1] || updates[2] != want[2] {
		t.Fatalf("ajustes de inventario inesperados: %v", updates)
	}
	// delta 0 (misma cantidad) no toca el inventario
	updates = nil
	w.actuales = map[int64]int64{1: 4}
	call(t, http.MethodPut, "/producto_pedido?pedido_id=10", `[{"productoId":1,"cantidad":1},{"productoId":1,"cantidad":3}]`, u, http.StatusOK)
	if len(updates) != 0 {
		t.Fatalf("no debía ajustar inventario: %v", updates)
	}
	// todas las líneas en 0: el pedido queda sin productos
	w.actuales = map[int64]int64{1: 4}
	b = call(t, http.MethodPut, "/producto_pedido?pedido_id=10", `[{"productoId":1,"cantidad":0}]`, u, http.StatusOK)
	contains(t, b, `"detalles":[]`)
}

func TestUpdateErrors(t *testing.T) {
	defer resetFake()
	w := newWorld()
	u := func(c *ProductoPedidoController) { c.Update() }
	url := "/producto_pedido?pedido_id=10"
	body := `[{"productoId":1,"cantidad":2}]`

	fakeBeginErr = errBoom
	call(t, http.MethodPut, url, body, u, http.StatusInternalServerError)
	fakeBeginErr = nil
	w.pedido = false
	call(t, http.MethodPut, url, body, u, http.StatusNotFound)
	w.pedido = true

	call(t, http.MethodPut, url, `[{"productoId":99,"cantidad":1}]`, u, http.StatusNotFound)
	b := call(t, http.MethodPut, url, `[{"productoId":3,"cantidad":4}]`, u, http.StatusConflict)
	contains(t, b, `"requerido":4`, `"disponible":1`)

	base := fakeQuery
	w.listErr = errBoom
	call(t, http.MethodPut, url, body, u, http.StatusInternalServerError)
	w.listErr = nil

	for _, prefix := range []string{"UPDATE producto", "DELETE", "INSERT"} {
		w.actuales = map[int64]int64{2: 1}
		fakeExec = func(q string, _ []driver.NamedValue) (driver.Result, error) {
			if strings.HasPrefix(q, prefix) {
				return nil, errBoom
			}
			return fakeResult{}, nil
		}
		call(t, http.MethodPut, url, body, u, http.StatusInternalServerError)
	}
	fakeExec = nil
	w.actuales = map[int64]int64{}

	fakeQuery = func(q string, a []driver.NamedValue) (driver.Rows, error) {
		if strings.Contains(q, `FROM "detalle_pedido"`) && len(a) == 2 {
			return nil, errBoom
		}
		return base(q, a)
	}
	call(t, http.MethodPut, url, body, u, http.StatusInternalServerError)
	fakeQuery = base
	fakeCommitErr = errBoom
	call(t, http.MethodPut, url, body, u, http.StatusInternalServerError)
}
