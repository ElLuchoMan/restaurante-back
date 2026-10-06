package pedido

import (
	"database/sql/driver"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"restaurante/database"
)

var (
	errBoom = errors.New("boom")

	pedidoCols    = []string{"pk_id_pedido", "fecha", "hora", "delivery", "estado_pedido", "pk_id_domicilio", "pk_id_pago", "pk_id_restaurante", "pk_documento_cliente", "updated_at", "updated_by"}
	detallesCols  = []string{"pk_id_pedido", "fecha", "hora", "delivery", "estado_pedido", "metodo_pago", "productos", "pago_id", "metodo_pago_id", "domicilio_id", "pk_documento_cliente"}
	countCols     = []string{"count"}
	fechaPedido   = time.Date(2025, 1, 31, 12, 0, 0, 0, time.UTC)
	updatedPedido = time.Date(2025, 1, 31, 20, 0, 0, 0, time.UTC)
)

func pedidoRow() []driver.Value {
	return []driver.Value{int64(10), fechaPedido, time.Date(2000, 1, 1, 18, 30, 0, 0, time.UTC), false, "INICIADO",
		nil, int64(4), int64(1), int64(1001), updatedPedido, "cajero"}
}

// rt programa el driver: devuelve las filas del primer fragmento que aparezca
// en la consulta; sin coincidencia, un resultado vacío.
type rt struct {
	frag string
	rows func() driver.Rows
}

func serve(routes ...rt) {
	fakeQuery = func(q string, _ []driver.NamedValue) (driver.Rows, error) {
		for _, r := range routes {
			if strings.Contains(q, r.frag) {
				return r.rows(), nil
			}
		}
		return rowsOf(nil), nil
	}
}

func pedidoOK() rt {
	return rt{`FROM "pedido"`, func() driver.Rows { return rowsOf(pedidoCols, pedidoRow()) }}
}
func count(frag string, n int64) rt {
	return rt{"COUNT(*) FROM \"" + frag + "\"", func() driver.Rows { return rowsOf(countCols, []driver.Value{n}) }}
}

func call(t *testing.T, method, target, body string, f func(c *PedidoController), status int) string {
	t.Helper()
	ctx, w := newCtx(method, target, body)
	c := &PedidoController{}
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
	g := func(c *PedidoController) { c.GetAll() }
	if b := call(t, http.MethodGet, "/pedidos", "", g, http.StatusOK); !strings.Contains(b, `"data":[]`) {
		t.Fatalf("lista vacía debe ser []: %s", b)
	}
	var gotQ string
	var gotArgs []driver.NamedValue
	fakeQuery = func(q string, a []driver.NamedValue) (driver.Rows, error) {
		gotQ, gotArgs = q, a
		return rowsOf(pedidoCols, pedidoRow()), nil
	}
	b := call(t, http.MethodGet, "/pedidos?fecha=2025-01-31&desde=2025-01-01&hasta=2025-01-31&mes=1&anio=2025&cliente=1001&metodo_pago=nequi&domicilio=true", "", g, http.StatusOK)
	contains(t, b, `"fechaPedido":"31-01-2025"`, `"horaPedido":"18:30:00"`, `"pedidoId":10`, `"pagoId":{`)
	for _, want := range []string{"p.fecha = $1", "BETWEEN", "MONTH", "YEAR", "pk_documento_cliente", "ILIKE", "IS NOT NULL"} {
		if !strings.Contains(gotQ, want) {
			t.Fatalf("filtro %q no aplicado: %s", want, gotQ)
		}
	}
	if len(gotArgs) != 7 {
		t.Fatalf("args inesperados: %v", gotArgs)
	}
	call(t, http.MethodGet, "/pedidos?domicilio=false", "", g, http.StatusOK)
	if !strings.Contains(gotQ, "pk_id_domicilio IS NULL") {
		t.Fatalf("falta IS NULL: %s", gotQ)
	}
	for _, q := range []string{"fecha=31-01-2025", "desde=x&hasta=2025-01-01", "desde=2025-01-01&hasta=x", "desde=2025-01-01", "hasta=2025-01-01",
		"desde=2025-02-01&hasta=2025-01-01", "mes=13", "mes=x", "anio=0", "cliente=0", "cliente=x", "domicilio=quizas"} {
		call(t, http.MethodGet, "/pedidos?"+q, "", g, http.StatusBadRequest)
	}
	fakeQuery = func(string, []driver.NamedValue) (driver.Rows, error) { return nil, errBoom }
	call(t, http.MethodGet, "/pedidos", "", g, http.StatusInternalServerError)
}

func TestPostValidation(t *testing.T) {
	defer resetFake()
	p := func(c *PedidoController) { c.Post() }
	for _, body := range []string{"", "{", `{"delivery":"si"}`} {
		call(t, http.MethodPost, "/pedidos", body, p, http.StatusBadRequest)
	}
	for _, body := range []string{`{"pk_id_domicilio":0}`, `{"documentoCliente":-1}`, `{"restauranteId":-1}`, `{"delivery":true}`} {
		call(t, http.MethodPost, "/pedidos", body, p, http.StatusBadRequest)
	}
}

func TestPostSuccessAndZones(t *testing.T) {
	defer resetFake()
	p := func(c *PedidoController) { c.Post() }
	serve(count("domicilio", 1), count("restaurante", 1), count("cliente", 1))
	b := call(t, http.MethodPost, "/pedidos", `{"delivery":true,"pk_id_domicilio":3,"restauranteId":1,"documentoCliente":1001}`, p, http.StatusCreated)
	contains(t, b, `"pedidoId":7`, `"estadoPedido":"INICIADO"`, `"delivery":true`)

	origZone, origLoad := database.BogotaZone, loadLocation
	t.Cleanup(func() { database.BogotaZone, loadLocation = origZone, origLoad })
	database.BogotaZone = nil
	loadLocation = time.LoadLocation
	call(t, http.MethodPost, "/pedidos", `{}`, p, http.StatusCreated)
	loadLocation = func(string) (*time.Location, error) { return nil, errBoom }
	call(t, http.MethodPost, "/pedidos", `{}`, p, http.StatusCreated)
}

func TestPostReferences(t *testing.T) {
	defer resetFake()
	p := func(c *PedidoController) { c.Post() }
	cases := []struct{ body, table string }{
		{`{"delivery":true,"pk_id_domicilio":3}`, "domicilio"},
		{`{"restauranteId":1}`, "restaurante"},
		{`{"documentoCliente":1001}`, "cliente"},
	}
	for _, tc := range cases {
		serve(count(tc.table, 0))
		call(t, http.MethodPost, "/pedidos", tc.body, p, http.StatusNotFound)
		fakeQuery = func(string, []driver.NamedValue) (driver.Rows, error) { return nil, errBoom }
		call(t, http.MethodPost, "/pedidos", tc.body, p, http.StatusInternalServerError)
	}
	serve(count("cliente", 1))
	fakeExec = func(string, []driver.NamedValue) (driver.Result, error) { return nil, errBoom }
	call(t, http.MethodPost, "/pedidos", `{"documentoCliente":1001}`, p, http.StatusInternalServerError)
	fakeExec = func(string, []driver.NamedValue) (driver.Result, error) {
		return nil, errors.New(`ERROR: insert or update violates foreign key constraint (SQLSTATE 23503)`)
	}
	call(t, http.MethodPost, "/pedidos", `{"documentoCliente":1001}`, p, http.StatusConflict)
}

func TestAssignDomicilio(t *testing.T) {
	defer resetFake()
	a := func(c *PedidoController) { c.AssignDomicilio() }
	for _, q := range []string{"", "pedido_id=10", "pedido_id=10&domicilio_id=0", "domicilio_id=3", "domicilio_id=3&pedido_id=x"} {
		call(t, http.MethodPost, "/pedidos/asignar-domicilio?"+q, "", a, http.StatusBadRequest)
	}
	u := "/pedidos/asignar-domicilio?pedido_id=10&domicilio_id=3"
	serve() // pedido inexistente
	call(t, http.MethodPost, u, "", a, http.StatusNotFound)
	fakeQuery = func(string, []driver.NamedValue) (driver.Rows, error) { return nil, errBoom }
	call(t, http.MethodPost, u, "", a, http.StatusInternalServerError)
	serve(pedidoOK(), count("domicilio", 0))
	call(t, http.MethodPost, u, "", a, http.StatusNotFound)
	serve(pedidoOK(), count("domicilio", 1))
	fakeExec = func(string, []driver.NamedValue) (driver.Result, error) { return nil, errBoom }
	call(t, http.MethodPost, u, "", a, http.StatusInternalServerError)
	fakeExec = nil
	b := call(t, http.MethodPost, u, "", a, http.StatusOK)
	contains(t, b, `"delivery":true`, `"domicilioId":{"domicilioId":3`, `"fechaPedido":"31-01-2025"`)
}

func TestAssignPago(t *testing.T) {
	defer resetFake()
	a := func(c *PedidoController) { c.AssignPago() }
	base := "/pedidos/asignar-pago?"
	for _, q := range []string{"pedido_id=10", "pedido_id=10&pago_id=0", "pedido_id=10&pago_id=4&cambiar_estado=tal", "pago_id=4&pedido_id=x"} {
		call(t, http.MethodPost, base+q, "", a, http.StatusBadRequest)
	}
	u := base + "pedido_id=10&pago_id=4"
	serve()
	call(t, http.MethodPost, u, "", a, http.StatusNotFound)
	serve(pedidoOK(), count("pago", 0))
	call(t, http.MethodPost, u, "", a, http.StatusNotFound)

	serve(pedidoOK(), count("pago", 1))
	fakeBeginErr = errBoom
	call(t, http.MethodPost, u, "", a, http.StatusInternalServerError)
	fakeBeginErr = nil

	// fallo al actualizar el pedido, luego al actualizar el pago, luego en commit
	var seen []string
	fakeExec = func(q string, _ []driver.NamedValue) (driver.Result, error) {
		seen = append(seen, q)
		if strings.HasPrefix(q, `UPDATE "pedido"`) {
			return nil, errBoom
		}
		return fakeResult{}, nil
	}
	call(t, http.MethodPost, u, "", a, http.StatusInternalServerError)
	fakeExec = func(q string, _ []driver.NamedValue) (driver.Result, error) {
		if strings.HasPrefix(q, `UPDATE "pago"`) {
			return nil, errBoom
		}
		return fakeResult{}, nil
	}
	call(t, http.MethodPost, u, "", a, http.StatusInternalServerError)
	fakeExec = nil
	fakeCommitErr = errBoom
	call(t, http.MethodPost, u, "", a, http.StatusInternalServerError)
	fakeCommitErr = nil

	b := call(t, http.MethodPost, u, "", a, http.StatusOK)
	contains(t, b, `"estadoPedido":"TERMINADO"`)
	seen = nil
	fakeExec = func(q string, _ []driver.NamedValue) (driver.Result, error) {
		seen = append(seen, q)
		return fakeResult{}, nil
	}
	b = call(t, http.MethodPost, u+"&cambiar_estado=false", "", a, http.StatusOK)
	contains(t, b, `"estadoPedido":"INICIADO"`)
	for _, q := range seen {
		if strings.HasPrefix(q, `UPDATE "pago"`) {
			t.Fatalf("no debe tocar el pago con cambiar_estado=false: %v", seen)
		}
	}
}

func TestUpdateEstado(t *testing.T) {
	defer resetFake()
	u := func(c *PedidoController) { c.UpdateEstadoPedido() }
	for _, q := range []string{"", "pedido_id=10", "pedido_id=10&estado=X", "estado=LISTO", "estado=LISTO&pedido_id=0"} {
		call(t, http.MethodPut, "/pedidos/actualizar-estado?"+q, "", u, http.StatusBadRequest)
	}
	url := "/pedidos/actualizar-estado?pedido_id=10&estado=listo"
	serve()
	call(t, http.MethodPut, url, "", u, http.StatusNotFound)
	fakeQuery = func(string, []driver.NamedValue) (driver.Rows, error) { return nil, errBoom }
	call(t, http.MethodPut, url, "", u, http.StatusInternalServerError)
	serve(pedidoOK())
	fakeExec = func(string, []driver.NamedValue) (driver.Result, error) { return nil, errBoom }
	call(t, http.MethodPut, url, "", u, http.StatusInternalServerError)
	fakeExec = nil
	b := call(t, http.MethodPut, url, "", u, http.StatusOK)
	contains(t, b, `"estadoPedido":"LISTO"`, `"fechaPedido":"31-01-2025"`)
	for _, e := range []string{"INICIADO", "EN_PREPARACION", "TERMINADO", "CANCELADO"} {
		call(t, http.MethodPut, "/pedidos/actualizar-estado?pedido_id=10&estado="+e, "", u, http.StatusOK)
	}
}

func TestDetalles(t *testing.T) {
	defer resetFake()
	d := func(c *PedidoController) { c.GetPedidoDetails() }
	for _, q := range []string{"", "pedido_id=0", "pedido_id=x", "pedido_id=-2"} {
		call(t, http.MethodGet, "/pedidos/detalles?"+q, "", d, http.StatusBadRequest)
	}
	call(t, http.MethodGet, "/pedidos/detalles?pedido_id=10", "", d, http.StatusNotFound)
	fakeQuery = func(string, []driver.NamedValue) (driver.Rows, error) { return nil, errBoom }
	call(t, http.MethodGet, "/pedidos/detalles?pedido_id=10", "", d, http.StatusInternalServerError)
	var gotQ string
	fakeQuery = func(q string, _ []driver.NamedValue) (driver.Rows, error) {
		gotQ = q
		return rowsOf(detallesCols, []driver.Value{int64(10), "31-01-2025", "18:30:00", false, "INICIADO", "NEQUI", "[]", int64(4), int64(1), int64(0), int64(1001)}), nil
	}
	b := call(t, http.MethodGet, "/pedidos/detalles?pedido_id=10", "", d, http.StatusOK)
	contains(t, b, `"fechaPedido":"31-01-2025"`, `"productos":"[]"`, `"metodoPago":"NEQUI"`)
	if !strings.Contains(gotQ, "'DD-MM-YYYY'") {
		t.Fatalf("la fecha debe consultarse como DD-MM-YYYY: %s", gotQ)
	}
}
