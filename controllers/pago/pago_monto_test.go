package pago

import (
	"database/sql/driver"
	"errors"
	"net/http"
	"strings"
	"testing"
)

// pedidoFake programa el pedido que lee POST /pagos para calcular el monto.
type pedidoFake struct {
	existe             bool
	dueno, pagoID      int64
	subtotal, lineas   int64
	descuento          int64
	errPedido, errCalc error
}

func (p *pedidoFake) serve() {
	serve(false, true)
	base := fakeQuery
	fakeQuery = func(q string, a []driver.NamedValue) (driver.Rows, error) {
		switch {
		case strings.Contains(q, "FROM pedido WHERE") && strings.Contains(q, "FOR UPDATE"):
			if p.errPedido != nil {
				return nil, p.errPedido
			}
			if !p.existe {
				return rowsOf([]string{"documento", "pago"}), nil
			}
			return rowsOf([]string{"documento", "pago"}, []driver.Value{p.dueno, p.pagoID}), nil
		case strings.Contains(q, "FROM pedido WHERE"):
			if p.errPedido != nil {
				return nil, p.errPedido
			}
			if !p.existe {
				return rowsOf([]string{"id"}), nil
			}
			return rowsOf([]string{"id"}, []driver.Value{int64(10)}), nil
		case strings.Contains(q, "SUM(precio * cantidad)"):
			if p.errCalc != nil {
				return nil, p.errCalc
			}
			return rowsOf([]string{"subtotal", "lineas"}, []driver.Value{p.subtotal, p.lineas}), nil
		case strings.Contains(q, "SUM(monto_descuento)"):
			return rowsOf([]string{"descuento"}, []driver.Value{p.descuento}), nil
		}
		return base(q, a)
	}
}

func nuevoPedidoFake() *pedidoFake {
	p := &pedidoFake{existe: true, dueno: 1001, subtotal: 50000, lineas: 2}
	p.serve()
	return p
}

const bodyPagoCliente = `{"fechaPago":"2025-02-01","horaPago":"10:00","estadoPago":"PENDIENTE","metodoPagoId":2,"pedidoId":10,"monto":%d}`

func cuerpo(monto string) string { return strings.Replace(bodyPagoCliente, "%d", monto, 1) }

// Ataque: el cliente manda monto 1 (o negativo, o nada): se ignora y manda el calculado. El pago
// se crea y se liga al pedido en UNA transacción (INSERT pago + UPDATE pedido, un COMMIT).
func TestPostClienteCreaYLigaConMontoDelServidor(t *testing.T) {
	defer resetFake()
	p := func(c *PagoController) { c.Post() }
	como(t, rolClienteT, 1001)
	ped := nuevoPedidoFake()

	for _, m := range []string{"1", "-50", "0", "999999999"} {
		resetLog()
		b := call(t, http.MethodPost, "/pagos", cuerpo(m), p, http.StatusCreated)
		if !strings.Contains(b, `"monto":50000`) {
			t.Fatalf("monto %s debía ignorarse y quedar 50000: %s", m, b)
		}
		w := escritos()
		if len(w) != 2 || !strings.HasPrefix(w[0], "INSERT") || !strings.HasPrefix(w[1], "UPDATE pedido SET pk_id_pago") || fakeCommits != 1 || fakeRollbacks != 0 {
			t.Fatalf("debía crear y ligar en una sola transacción: %v commits=%d rollbacks=%d", w, fakeCommits, fakeRollbacks)
		}
	}
	b := call(t, http.MethodPost, "/pagos", strings.Replace(cuerpo("1"), `,"monto":1`, ``, 1), p, http.StatusCreated)
	if !strings.Contains(b, `"monto":50000`) {
		t.Fatalf("sin monto: %s", b)
	}
	ped.descuento = 8000
	if b = call(t, http.MethodPost, "/pagos", cuerpo("1"), p, http.StatusCreated); !strings.Contains(b, `"monto":42000`) {
		t.Fatalf("con descuento: %s", b)
	}
	ped.descuento = 90000
	if b = call(t, http.MethodPost, "/pagos", cuerpo("1"), p, http.StatusCreated); !strings.Contains(b, `"monto":0`) {
		t.Fatalf("descuento mayor al subtotal: %s", b)
	}
}

func resetLog() { fakeCommits, fakeRollbacks, fakeLog = 0, 0, nil }

// Nada queda escrito (rollback, ningún INSERT/UPDATE confirmado) cuando el pago no puede crearse.
func TestPostClienteRechazosNoEscribenNada(t *testing.T) {
	defer resetFake()
	p := func(c *PagoController) { c.Post() }
	como(t, rolClienteT, 1001)
	ped := nuevoPedidoFake()
	nada := func(t *testing.T) {
		t.Helper()
		if fakeCommits != 0 || fakeRollbacks != 1 {
			t.Fatalf("debía deshacer: commits=%d rollbacks=%d", fakeCommits, fakeRollbacks)
		}
		for _, w := range escritos() {
			if strings.HasPrefix(w, "UPDATE pedido") {
				t.Fatalf("no debía ligar nada: %v", escritos())
			}
		}
	}
	// sin pedidoId, y estados distintos de PENDIENTE: 400/403 antes de abrir la transacción
	call(t, http.MethodPost, "/pagos", strings.Replace(cuerpo("1"), `"pedidoId":10,`, ``, 1), p, http.StatusBadRequest)
	call(t, http.MethodPost, "/pagos", strings.Replace(cuerpo("1"), "PENDIENTE", "PAGADO", 1), p, http.StatusForbidden)
	call(t, http.MethodPost, "/pagos", strings.Replace(cuerpo("1"), "PENDIENTE", "NO_PAGO", 1), p, http.StatusForbidden)
	if fakeCommits+fakeRollbacks != 0 || len(escritos()) != 0 {
		t.Fatalf("no debía tocar la BD: %v", fakeLog)
	}
	// pedido sin productos: 409
	ped.lineas, ped.subtotal = 0, 0
	contains409(t, call(t, http.MethodPost, "/pagos", cuerpo("1"), p, http.StatusConflict), "no tiene productos")
	nada(t)
	ped.lineas, ped.subtotal = 2, 50000
	// pedido que ya tiene pago: 409 (y el pago huérfano no llega a crearse)
	resetLog()
	ped.pagoID = 4
	contains409(t, call(t, http.MethodPost, "/pagos", cuerpo("1"), p, http.StatusConflict), "ya tiene un pago")
	nada(t)
	for _, w := range escritos() {
		t.Fatalf("no debía insertar nada: %v", w)
	}
	ped.pagoID = 0
	// pedido ajeno o inexistente: 404 igual
	resetLog()
	como(t, rolClienteT, 2002)
	call(t, http.MethodPost, "/pagos", cuerpo("1"), p, http.StatusNotFound)
	nada(t)
	como(t, rolClienteT, 1001)
	resetLog()
	ped.existe = false
	call(t, http.MethodPost, "/pagos", cuerpo("1"), p, http.StatusNotFound)
	nada(t)
	ped.existe = true
	// errores de base de datos al bloquear el pedido y al calcular el monto
	resetLog()
	ped.errPedido = errBoom
	call(t, http.MethodPost, "/pagos", cuerpo("1"), p, http.StatusInternalServerError)
	nada(t)
	resetLog()
	ped.errPedido, ped.errCalc = nil, errBoom
	call(t, http.MethodPost, "/pagos", cuerpo("1"), p, http.StatusInternalServerError)
	nada(t)
	ped.errCalc = nil
	// fallo a mitad: el pago se insertó pero ligarlo falla -> rollback, no queda huérfano
	resetLog()
	fakeExec = func(q string, _ []driver.NamedValue) (driver.Result, error) {
		if strings.HasPrefix(q, "UPDATE pedido SET pk_id_pago") {
			return nil, errBoom
		}
		return fakeResult{}, nil
	}
	call(t, http.MethodPost, "/pagos", cuerpo("1"), p, http.StatusInternalServerError)
	if fakeCommits != 0 || fakeRollbacks != 1 || len(escritos()) != 2 {
		t.Fatalf("fallo al ligar debía deshacer el INSERT: commits=%d rollbacks=%d %v", fakeCommits, fakeRollbacks, escritos())
	}
	// fallo al insertar (500 y conflicto 409), al abrir la transacción y al confirmar
	resetLog()
	fakeExec = func(string, []driver.NamedValue) (driver.Result, error) { return nil, errBoom }
	call(t, http.MethodPost, "/pagos", cuerpo("1"), p, http.StatusInternalServerError)
	nada(t)
	fakeExec = func(string, []driver.NamedValue) (driver.Result, error) {
		return nil, errors.New("duplicate key 23505")
	}
	call(t, http.MethodPost, "/pagos", cuerpo("1"), p, http.StatusConflict)
	fakeExec = nil
	fakeBeginErr = errBoom
	call(t, http.MethodPost, "/pagos", cuerpo("1"), p, http.StatusInternalServerError)
	fakeBeginErr = nil
	fakeCommitErr = errBoom
	call(t, http.MethodPost, "/pagos", cuerpo("1"), p, http.StatusInternalServerError)
}

func TestPostPersonalMontoDelServidor(t *testing.T) {
	defer resetFake()
	p := func(c *PagoController) { c.Post() }
	como(t, rolMesero, 5)
	ped := nuevoPedidoFake()

	// sin monto (o 0) con pedido: manda el calculado
	for _, m := range []string{"0"} {
		b := call(t, http.MethodPost, "/pagos", cuerpo(m), p, http.StatusCreated)
		if !strings.Contains(b, `"monto":50000`) {
			t.Fatalf("monto %s debía quedar 50000: %s", m, b)
		}
	}
	b := call(t, http.MethodPost, "/pagos", strings.Replace(cuerpo("1"), `,"monto":1`, ``, 1), p, http.StatusCreated)
	if !strings.Contains(b, `"monto":50000`) {
		t.Fatalf("sin monto: %s", b)
	}

	// con descuentos ya aplicados al pedido, el monto los refleja (y nunca baja de 0)
	ped.descuento = 8000
	if b = call(t, http.MethodPost, "/pagos", cuerpo("0"), p, http.StatusCreated); !strings.Contains(b, `"monto":42000`) {
		t.Fatalf("con descuento: %s", b)
	}
	ped.descuento = 90000
	if b = call(t, http.MethodPost, "/pagos", cuerpo("0"), p, http.StatusCreated); !strings.Contains(b, `"monto":0`) {
		t.Fatalf("descuento mayor al subtotal: %s", b)
	}
	ped.descuento = 0

	// pedido sin productos: 409
	ped.lineas, ped.subtotal = 0, 0
	contains409(t, call(t, http.MethodPost, "/pagos", cuerpo("0"), p, http.StatusConflict), "no tiene productos")
	ped.lineas, ped.subtotal = 2, 50000
	// pedido inexistente: 404
	ped.existe = false
	call(t, http.MethodPost, "/pagos", cuerpo("0"), p, http.StatusNotFound)
	ped.existe = true
	// errores de base de datos
	ped.errPedido = errBoom
	call(t, http.MethodPost, "/pagos", cuerpo("0"), p, http.StatusInternalServerError)
	ped.errPedido, ped.errCalc = nil, errBoom
	call(t, http.MethodPost, "/pagos", cuerpo("0"), p, http.StatusInternalServerError)
}

func contains409(t *testing.T, body, want string) {
	t.Helper()
	if !strings.Contains(body, want) {
		t.Fatalf("falta %q en %s", want, body)
	}
}

func TestPostPersonalMontoManualOCalculado(t *testing.T) {
	defer resetFake()
	p := func(c *PagoController) { c.Post() }
	como(t, rolMesero, 5)
	ped := nuevoPedidoFake()

	// monto manual > 0 (ajuste de mostrador) se respeta, con o sin pedido
	if b := call(t, http.MethodPost, "/pagos", cuerpo("12345"), p, http.StatusCreated); !strings.Contains(b, `"monto":12345`) {
		t.Fatalf("manual con pedido: %s", b)
	}
	if b := call(t, http.MethodPost, "/pagos", strings.Replace(cuerpo("777"), `"pedidoId":10,`, ``, 1), p, http.StatusCreated); !strings.Contains(b, `"monto":777`) {
		t.Fatalf("manual sin pedido: %s", b)
	}
	// con pedido y monto 0/omitido usa el calculado, aunque el pedido ya tenga pago; no liga nada
	ped.pagoID = 4
	resetLog()
	if b := call(t, http.MethodPost, "/pagos", cuerpo("0"), p, http.StatusCreated); !strings.Contains(b, `"monto":50000`) {
		t.Fatalf("calculado: %s", b)
	}
	if w := escritos(); len(w) != 1 || !strings.HasPrefix(w[0], "INSERT") || fakeCommits != 0 {
		t.Fatalf("el personal solo inserta el pago (sin ligarlo ni abrir transacción): %v", w)
	}
	ped.pagoID = 0
	// negativo -> 400; sin pedido y monto 0 -> 400; pedido ajeno de otro cliente: el personal lo ve
	call(t, http.MethodPost, "/pagos", cuerpo("-1"), p, http.StatusBadRequest)
	call(t, http.MethodPost, "/pagos", strings.Replace(cuerpo("0"), `"pedidoId":10,`, ``, 1), p, http.StatusBadRequest)
	// el pedido sigue necesitando productos para usar el calculado
	ped.lineas = 0
	call(t, http.MethodPost, "/pagos", cuerpo("0"), p, http.StatusConflict)
	// pedido inexistente
	ped.existe = false
	call(t, http.MethodPost, "/pagos", cuerpo("100"), p, http.StatusNotFound)
}
