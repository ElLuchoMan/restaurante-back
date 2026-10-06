package pedido

import (
	"database/sql/driver"
	"net/http"
	"strings"
	"testing"
)

// montoRoutes programa las consultas con las que AssignPago (Cliente) bloquea el
// pedido y calcula el monto: el pago que ya tiene el pedido (0 = ninguno), el
// subtotal y el número de líneas del detalle.
func montoRoutes(pagoActual, subtotal, lineas int64) []rt {
	return []rt{
		{"FROM pedido WHERE", func() driver.Rows { return rowsOf([]string{"pago"}, []driver.Value{pagoActual}) }},
		{"SUM(precio * cantidad)", func() driver.Rows {
			return rowsOf([]string{"subtotal", "lineas"}, []driver.Value{subtotal, lineas})
		}},
		{"SUM(monto_descuento)", func() driver.Rows { return rowsOf([]string{"descuento"}, []driver.Value{int64(0)}) }},
	}
}

func rutasAsignar(extra ...rt) []rt {
	return append(extra, count("pedido", 0), count("pago", 1), pedidoSinPago())
}

// Ataque: el cliente crea un pago con un monto falso (aquí 1) y lo asigna a su
// pedido; el servidor lo sobreescribe con lo calculado desde detalle_pedido.
func TestAssignPagoClienteFijaMontoDelServidor(t *testing.T) {
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
	serve(rutasAsignar(montoRoutes(0, 73000, 3)...)...)
	call(t, http.MethodPost, url, "", a, http.StatusOK)
	if monto != 73000 {
		t.Fatalf("el monto debe ser el calculado por el servidor, no el del pago: %d", monto)
	}

	// pedido sin productos: 409 y no se asigna ni se fija monto
	monto = -1
	var asignado bool
	fakeExec = func(q string, _ []driver.NamedValue) (driver.Result, error) {
		asignado = asignado || strings.HasPrefix(q, `UPDATE "pedido"`)
		return fakeResult{}, nil
	}
	serve(rutasAsignar(montoRoutes(0, 0, 0)...)...)
	if b := call(t, http.MethodPost, url, "", a, http.StatusConflict); !strings.Contains(b, "no tiene productos") || asignado {
		t.Fatalf("sin productos: %s (asignado=%v)", b, asignado)
	}

	// el pago no está PENDIENTE (o no existe en ese estado): 409
	fakeExec = nil
	fakeAffected = 0
	serve(rutasAsignar(montoRoutes(0, 73000, 3)...)...)
	if b := call(t, http.MethodPost, url, "", a, http.StatusConflict); !strings.Contains(b, "PENDIENTE") {
		t.Fatalf("pago no pendiente: %s", b)
	}
	fakeAffected = 1

	// errores de base de datos al bloquear, calcular y fijar el monto
	serve(rutasAsignar(montoRoutes(0, 73000, 3)...)...)
	ok := fakeQuery
	fakeQuery = func(q string, a []driver.NamedValue) (driver.Rows, error) {
		if strings.Contains(q, "FROM pedido WHERE") {
			return nil, errBoom
		}
		return ok(q, a)
	}
	call(t, http.MethodPost, url, "", a, http.StatusInternalServerError)
	serve(rutasAsignar(montoRoutes(0, 1, 1)[:1]...)...) // sin filas de suma: el cálculo falla
	call(t, http.MethodPost, url, "", a, http.StatusInternalServerError)
	serve(rutasAsignar(montoRoutes(0, 73000, 3)...)...)
	fakeExec = func(string, []driver.NamedValue) (driver.Result, error) { return nil, errBoom }
	call(t, http.MethodPost, url, "", a, http.StatusInternalServerError)
	fakeExec = nil
	fakeAffErr = errBoom // el error al leer las filas afectadas también aborta
	call(t, http.MethodPost, url, "", a, http.StatusInternalServerError)
}
