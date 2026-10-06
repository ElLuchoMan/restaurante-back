package pedido

import (
	"database/sql/driver"
	"net/http"
	"strings"
	"testing"
)

// montoRoutes programa las consultas con las que se calculaba el monto: el pago que ya tiene el
// pedido (0 = ninguno), el subtotal y el número de líneas del detalle.
func montoRoutes(pagoActual, subtotal, lineas int64) []rt {
	return []rt{
		{"FROM pedido WHERE", func() driver.Rows { return rowsOf([]string{"pago"}, []driver.Value{pagoActual}) }},
		{"SUM(precio * cantidad)", func() driver.Rows {
			return rowsOf([]string{"subtotal", "lineas"}, []driver.Value{subtotal, lineas})
		}},
		{"SUM(monto_descuento)", func() driver.Rows { return rowsOf([]string{"descuento"}, []driver.Value{int64(0)}) }},
	}
}

func rutasAsignar(pagoActual, subtotal, lineas int64) []rt {
	return append(montoRoutes(pagoActual, subtotal, lineas), count("pedido", 0), count("pago", 1), pedidoOK())
}

// Ataque: un cliente adivina el id de un pago sin asignar (el huérfano de un tercero, o el de otro
// pedido) y lo asigna a su pedido para apropiárselo o reescribirle el monto. Siempre 404, sin
// escribir nada y sin distinguir "huérfano" de "ligado a otro pedido".
func TestAssignPagoClienteNoReclamaPagosAjenos(t *testing.T) {
	defer resetFake()
	grabar(t)
	a := func(c *PedidoController) { c.AssignPago() }
	const url = "/pedidos/asignar-pago?pedido_id=10&pago_id=4&cambiar_estado=false"
	dueno(t)
	var escrito []string
	fakeExec = func(q string, _ []driver.NamedValue) (driver.Result, error) {
		escrito = append(escrito, q)
		return fakeResult{}, nil
	}
	// el pedido no tiene pago (el pago 4 es huérfano o de otro) / tiene otro pago distinto
	for _, actual := range []int64{0, 9} {
		serve(rutasAsignar(actual, 73000, 3)...)
		b := call(t, http.MethodPost, url, "", a, http.StatusNotFound)
		if !strings.Contains(b, "Pago no encontrado") {
			t.Fatalf("pagoActual=%d: %s", actual, b)
		}
	}
	if len(escrito) != 0 || fakeCommits != 0 || fakeRollbacks != 2 {
		t.Fatalf("no debe escribir: %v commits=%d rollbacks=%d", escrito, fakeCommits, fakeRollbacks)
	}
}

// El monto lo manda el servidor: se recalcula desde detalle_pedido con el pedido bloqueado.
func TestAssignPagoClienteRecalculaMonto(t *testing.T) {
	defer resetFake()
	grabar(t)
	a := func(c *PedidoController) { c.AssignPago() }
	const url = "/pedidos/asignar-pago?pedido_id=10&pago_id=4&cambiar_estado=false"
	dueno(t)

	var monto int64 = -1
	fakeExec = func(q string, args []driver.NamedValue) (driver.Result, error) {
		if strings.HasPrefix(q, "UPDATE pago SET monto") {
			monto = args[0].Value.(int64)
		}
		return fakeResult{}, nil
	}
	serve(rutasAsignar(4, 73000, 3)...)
	call(t, http.MethodPost, url, "", a, http.StatusOK)
	if monto != 73000 {
		t.Fatalf("el monto debe ser el calculado por el servidor: %d", monto)
	}
	// pedido sin productos: 409 y no se fija monto
	monto = -1
	serve(rutasAsignar(4, 0, 0)...)
	if b := call(t, http.MethodPost, url, "", a, http.StatusConflict); !strings.Contains(b, "no tiene productos") || monto != -1 {
		t.Fatalf("sin productos: %s (monto=%d)", b, monto)
	}
	// errores de base de datos al bloquear, calcular, fijar el monto, abrir la transacción y confirmar
	serve(rutasAsignar(4, 73000, 3)...)
	ok := fakeQuery
	fakeQuery = func(q string, a []driver.NamedValue) (driver.Rows, error) {
		if strings.Contains(q, "FROM pedido WHERE") {
			return nil, errBoom
		}
		return ok(q, a)
	}
	call(t, http.MethodPost, url, "", a, http.StatusInternalServerError)
	serve(rutasAsignar(4, 73000, 3)[1:]...) // sin la fila del pedido bloqueado: el cálculo ya no importa, falla el bloqueo
	call(t, http.MethodPost, url, "", a, http.StatusInternalServerError)
	serve(append(montoRoutes(4, 1, 1)[:1], count("pedido", 0), count("pago", 1), pedidoOK())...) // sin filas de suma: el cálculo falla
	call(t, http.MethodPost, url, "", a, http.StatusInternalServerError)
	serve(rutasAsignar(4, 73000, 3)...)
	fakeExec = func(string, []driver.NamedValue) (driver.Result, error) { return nil, errBoom }
	call(t, http.MethodPost, url, "", a, http.StatusInternalServerError)
	fakeExec = nil
	fakeCommitErr = errBoom
	call(t, http.MethodPost, url, "", a, http.StatusInternalServerError)
	fakeCommitErr = nil
	fakeBeginErr = errBoom
	call(t, http.MethodPost, url, "", a, http.StatusInternalServerError)
}
