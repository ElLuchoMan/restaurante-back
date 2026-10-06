package pedido

import (
	"database/sql/driver"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"restaurante/controllers/login"
	"restaurante/internal/notify"
	"restaurante/models"
)

var domicilioCols = []string{"pk_id_domicilio", "direccion", "telefono", "estado_domicilio", "entregado", "fecha", "observaciones", "created_at", "updated_at", "created_by", "updated_by", "pk_documento_trabajador"}

// checkoutWorld programa el driver para POST /pedidos/checkout: existencia de las referencias,
// inventario, domicilio releído y suma del detalle.
type checkoutWorld struct {
	restaurante, metodo, cliente bool
	stock                        map[int64]int64
	subtotal, lineas, descuento  int64
	failQuery                    string // fragmento de una consulta que falla
	failExec                     string // fragmento de una escritura que falla (con failExecErr)
	failExecErr                  error
}

func newCheckoutWorld() *checkoutWorld {
	w := &checkoutWorld{restaurante: true, metodo: true, cliente: true, stock: map[int64]int64{1: 10, 2: 10, 3: 1}, subtotal: 50000, lineas: 2}
	w.serve()
	return w
}

func (w *checkoutWorld) serve() {
	existe := func(b bool) driver.Rows {
		if b {
			return rowsOf(countCols, []driver.Value{int64(1)})
		}
		return rowsOf(countCols, []driver.Value{int64(0)})
	}
	fakeQuery = func(q string, args []driver.NamedValue) (driver.Rows, error) {
		if w.failQuery != "" && strings.Contains(q, w.failQuery) {
			return nil, errBoom
		}
		switch {
		case strings.Contains(q, `COUNT(*) FROM "restaurante"`):
			return existe(w.restaurante), nil
		case strings.Contains(q, `COUNT(*) FROM "metodo_pago"`):
			return existe(w.metodo), nil
		case strings.Contains(q, `COUNT(*) FROM "cliente"`):
			return existe(w.cliente), nil
		case strings.Contains(q, "FROM producto WHERE"):
			var vals [][]driver.Value
			for _, a := range args {
				if s, ok := w.stock[a.Value.(int64)]; ok {
					vals = append(vals, []driver.Value{a.Value, s})
				}
			}
			return rowsOf([]string{"pk_id_producto", "cantidad"}, vals...), nil
		case strings.Contains(q, `FROM "domicilio"`):
			return rowsOf(domicilioCols, []driver.Value{int64(7), "Calle 1", "300", "PENDIENTE", false, fechaPedido, nil, updatedPedido, updatedPedido, "Usuario 1001", nil, nil}), nil
		case strings.Contains(q, "SUM(precio * cantidad)"):
			return rowsOf([]string{"subtotal", "lineas"}, []driver.Value{w.subtotal, w.lineas}), nil
		case strings.Contains(q, "SUM(monto_descuento)"):
			return rowsOf([]string{"descuento"}, []driver.Value{w.descuento}), nil
		}
		return rowsOf(nil), nil
	}
	fakeExec = func(q string, _ []driver.NamedValue) (driver.Result, error) {
		if w.failExec != "" && strings.Contains(q, w.failExec) {
			return nil, w.failExecErr
		}
		return fakeResult{}, nil
	}
}

const bodyCheckoutDomicilio = `{"restauranteId":1,"domicilio":{"direccion":" Calle 1 ","telefono":"300","observaciones":"timbre"},"productos":[{"productoId":1,"cantidad":2},{"productoId":2,"cantidad":1}],"pago":{"metodoPagoId":2,"monto":3}}`
const bodyCheckoutSimple = `{"productos":[{"productoId":1,"cantidad":2}],"pago":{"metodoPagoId":2}}`

func checkout(t *testing.T, body string, status int) string {
	t.Helper()
	return call(t, http.MethodPost, "/pedidos/checkout", body, func(c *PedidoController) { c.Checkout() }, status)
}

func contiene(l []fakeStmtLog, frag string) *fakeStmtLog {
	for i := range l {
		if strings.Contains(l[i].query, frag) {
			return &l[i]
		}
	}
	return nil
}

func tieneArg(l *fakeStmtLog, v interface{}) bool {
	for _, a := range l.args {
		if a.Value == v {
			return true
		}
	}
	return false
}

// nada comprueba que el checkout fallido no dejó nada: ni COMMIT ni notificaciones.
func nada(t *testing.T, g *grabador) {
	t.Helper()
	if fakeCommits != 0 || fakeRollbacks != 1 {
		t.Fatalf("debe deshacer sin confirmar: commits=%d rollbacks=%d", fakeCommits, fakeRollbacks)
	}
	if len(g.eventos) != 0 {
		t.Fatalf("no debe notificar: %+v", g.eventos)
	}
}

func TestCheckoutClienteConDomicilio(t *testing.T) {
	defer resetFake()
	g := grabar(t)
	como(t, rolClienteT, 1001)
	newCheckoutWorld()

	// El body trae monto=3 (ataque): se ignora y manda el calculado (50000).
	b := checkout(t, bodyCheckoutDomicilio, http.StatusCreated)
	contains(t, b, `"monto":50000`, `"delivery":true`, `"estadoPedido":"INICIADO"`, `"pagoId":{`, `"domicilioId":{`, `"domicilioId":7`, `"documentoCliente":{"documentoCliente":1001`, `"restauranteId":{`)
	w := escritos()
	pago := contiene(w, `INSERT INTO "pago"`)
	if pago == nil || !tieneArg(pago, int64(50000)) || tieneArg(pago, int64(3)) {
		t.Fatalf("el pago debe llevar el monto del servidor y no el del body: %+v", pago)
	}
	if !tieneArg(pago, "PENDIENTE") {
		t.Fatalf("estado por defecto PENDIENTE: %+v", pago)
	}
	dom := contiene(w, "INSERT INTO domicilio")
	if dom == nil || !tieneArg(dom, "Calle 1") || !tieneArg(dom, "timbre") || !tieneArg(dom, "Usuario 1001") {
		t.Fatalf("domicilio mal insertado: %+v", dom)
	}
	if d1, d2 := contiene(w, "UPDATE producto"), contiene(w, `INSERT INTO "detalle_pedido"`); d1 == nil || d2 == nil {
		t.Fatalf("falta descontar inventario o insertar detalles: %+v", w)
	}
	n := 0
	for _, l := range w {
		if strings.Contains(l.query, `INSERT INTO "detalle_pedido"`) {
			n++
		}
	}
	if n != 2 {
		t.Fatalf("esperaba 2 detalles, hay %d", n)
	}
	if contiene(w, `UPDATE "pedido"`) == nil {
		t.Fatal("el pago debe enlazarse al pedido")
	}
	if fakeCommits != 1 {
		t.Fatalf("debe confirmar una vez: %d", fakeCommits)
	}
	ev := unico(t, g)
	if ev.Tipo != notify.PedidoCreado || ev.PedidoID != 7 || ev.Cliente != 1001 || ev.DomicilioID != 7 {
		t.Fatalf("evento inesperado: %+v", ev)
	}
}

func TestCheckoutSinDomicilioNiRestaurante(t *testing.T) {
	defer resetFake()
	g := grabar(t)
	como(t, rolClienteT, 1001)
	newCheckoutWorld()
	b := checkout(t, bodyCheckoutSimple, http.StatusCreated)
	contains(t, b, `"delivery":false`, `"monto":50000`, `"restauranteId":null`)
	if strings.Contains(b, `"domicilioId":{`) {
		t.Fatalf("sin domicilio no debe haber domicilioId: %s", b)
	}
	if contiene(escritos(), "INSERT INTO domicilio") != nil {
		t.Fatal("no debe crear domicilio")
	}
	if ev := unico(t, g); ev.DomicilioID != 0 || ev.Cliente != 1001 {
		t.Fatalf("evento inesperado: %+v", ev)
	}
}

func TestCheckoutPersonalMostradorYMontoManual(t *testing.T) {
	defer resetFake()
	g := grabar(t)
	newCheckoutWorld()
	// personal sin cliente = mostrador (no consulta clientes); monto manual > 0 se respeta;
	// fecha, hora y estado explícitos (el estado en minúsculas se normaliza)
	body := `{"productos":[{"productoId":1,"cantidad":1},{"productoId":1,"cantidad":2}],"pago":{"metodoPagoId":2,"monto":777,"estadoPago":"pagado","fechaPago":"2025-02-01","horaPago":"10:30"}}`
	b := checkout(t, body, http.StatusCreated)
	contains(t, b, `"monto":777`, `"documentoCliente":null`)
	pago := contiene(escritos(), `INSERT INTO "pago"`)
	if !tieneArg(pago, int64(777)) || !tieneArg(pago, "PAGADO") {
		t.Fatalf("pago inesperado: %+v", pago)
	}
	if contiene(fakeLog, `FROM "cliente"`) != nil {
		t.Fatal("un pedido de mostrador no valida cliente")
	}
	// las líneas repetidas se suman: un solo UPDATE de inventario con 3 unidades
	if up := contiene(escritos(), "UPDATE producto"); up == nil || !tieneArg(up, int64(3)) {
		t.Fatalf("inventario: %+v", up)
	}
	if ev := unico(t, g); ev.Cliente != 0 {
		t.Fatalf("mostrador sin cliente: %+v", ev)
	}

	// monto 0 -> el calculado (con descuentos aplicados); personal a nombre de un cliente
	fakeLog, fakeCommits = nil, 0
	w := newCheckoutWorld()
	w.descuento = 8000
	b = checkout(t, `{"documentoCliente":1001,"productos":[{"productoId":1,"cantidad":1}],"pago":{"metodoPagoId":2,"horaPago":"10:30:15"}}`, http.StatusCreated)
	contains(t, b, `"monto":42000`)
}

func TestCheckoutInventarioInsuficienteNoPersisteNada(t *testing.T) {
	defer resetFake()
	g := grabar(t)
	como(t, rolClienteT, 1001)
	newCheckoutWorld()
	// producto 3 solo tiene 1 y 1 tiene 10: se piden 5 y 20
	body := `{"domicilio":{"direccion":"Calle 1","telefono":"300"},"productos":[{"productoId":3,"cantidad":5},{"productoId":1,"cantidad":20},{"productoId":2,"cantidad":1}],"pago":{"metodoPagoId":2}}`
	b := checkout(t, body, http.StatusConflict)
	contains(t, b, `"data":[{"productoId":1,"requerido":20,"disponible":10},{"productoId":3,"requerido":5,"disponible":1}]`, "Inventario insuficiente")
	if w := escritos(); len(w) != 0 {
		t.Fatalf("no debe escribir nada (ni domicilio, ni pedido, ni descuento): %+v", w)
	}
	nada(t, g)

	// producto inexistente: 404 y tampoco queda nada
	fakeLog, fakeRollbacks = nil, 0
	b = checkout(t, `{"productos":[{"productoId":99,"cantidad":1},{"productoId":98,"cantidad":1}],"pago":{"metodoPagoId":2}}`, http.StatusNotFound)
	contains(t, b, "Producto no encontrado: 98, 99")
	if len(escritos()) != 0 {
		t.Fatalf("escrituras inesperadas: %+v", escritos())
	}
	nada(t, g)
}

func TestCheckoutFalloTardioDeshaceTodo(t *testing.T) {
	// Cada fallo de BD en cualquier paso (incluso con el inventario ya descontado y el domicilio,
	// el pedido o el pago ya insertados) termina en 500, ROLLBACK, sin COMMIT y sin notificar.
	casos := []struct{ nombre, query, exec string }{
		{"restaurante", `COUNT(*) FROM "restaurante"`, ""},
		{"metodo", `COUNT(*) FROM "metodo_pago"`, ""},
		{"cliente", `COUNT(*) FROM "cliente"`, ""},
		{"bloqueo inventario", "FROM producto WHERE", ""},
		{"descuento inventario", "", "UPDATE producto"},
		{"insert domicilio", "", "INSERT INTO domicilio"},
		{"relectura domicilio", `FROM "domicilio"`, ""},
		{"insert pedido", "", `INSERT INTO "pedido"`},
		{"insert detalle", "", `INSERT INTO "detalle_pedido"`},
		{"monto", "SUM(precio * cantidad)", ""},
		{"insert pago", "", `INSERT INTO "pago"`},
		{"enlazar pago", "", `UPDATE "pedido"`},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			defer resetFake()
			g := grabar(t)
			como(t, rolClienteT, 1001)
			w := newCheckoutWorld()
			w.failQuery, w.failExec, w.failExecErr = c.query, c.exec, errBoom
			checkout(t, bodyCheckoutDomicilio, http.StatusInternalServerError)
			nada(t, g)
		})
	}
}

func TestCheckoutConflictoDeBDEs409(t *testing.T) {
	defer resetFake()
	g := grabar(t)
	como(t, rolClienteT, 1001)
	w := newCheckoutWorld()
	w.failExec, w.failExecErr = `INSERT INTO "pedido"`, errors.New("ERROR: duplicate key value (SQLSTATE 23505)")
	checkout(t, bodyCheckoutSimple, http.StatusConflict)
	nada(t, g)
}

func TestCheckoutFalloDeCommitNoNotifica(t *testing.T) {
	defer resetFake()
	g := grabar(t)
	como(t, rolClienteT, 1001)
	newCheckoutWorld()
	fakeCommitErr = errBoom
	checkout(t, bodyCheckoutDomicilio, http.StatusInternalServerError)
	if len(g.eventos) != 0 {
		t.Fatalf("un commit fallido no debe notificar: %+v", g.eventos)
	}
	if fakeCommits != 1 {
		t.Fatalf("se intentó confirmar una vez: %d", fakeCommits)
	}
}

func TestCheckoutNotificaSoloDespuesDelCommit(t *testing.T) {
	defer resetFake()
	como(t, rolClienteT, 1001)
	newCheckoutWorld()
	prev := notify.Default
	defer func() { notify.Default = prev }()
	var commitsAlNotificar = -1
	notify.Default = notificadorFunc(func(notify.Evento) { commitsAlNotificar = fakeCommits })
	checkout(t, bodyCheckoutSimple, http.StatusCreated)
	if commitsAlNotificar != 1 {
		t.Fatalf("la notificación debe salir después del commit (commits=%d)", commitsAlNotificar)
	}
}

type notificadorFunc func(notify.Evento)

func (f notificadorFunc) Notificar(ev notify.Evento) { f(ev) }

func TestCheckoutInicioDeTransaccionFalla(t *testing.T) {
	defer resetFake()
	g := grabar(t)
	como(t, rolClienteT, 1001)
	newCheckoutWorld()
	fakeBeginErr = errBoom
	checkout(t, bodyCheckoutSimple, http.StatusInternalServerError)
	if len(g.eventos) != 0 || len(fakeLog) != 0 {
		t.Fatalf("no debe haber actividad: %+v %+v", g.eventos, fakeLog)
	}
}

func TestCheckoutClienteNoActuaPorOtro(t *testing.T) {
	defer resetFake()
	g := grabar(t)
	newCheckoutWorld()
	conDoc := func(doc string) string {
		return `{"documentoCliente":` + doc + `,"productos":[{"productoId":1,"cantidad":1}],"pago":{"metodoPagoId":2}}`
	}
	como(t, rolClienteT, 1001)
	if b := checkout(t, conDoc("2002"), http.StatusForbidden); !strings.Contains(b, "otro cliente") {
		t.Fatalf("mensaje: %s", b)
	}
	// el mismo documento de su token sí es válido (y 0/ausente se ignora)
	checkout(t, conDoc("1001"), http.StatusCreated)
	checkout(t, conDoc("0"), http.StatusCreated)
	// un token de cliente sin documento no identifica a nadie
	como(t, rolClienteT, 0)
	checkout(t, bodyCheckoutSimple, http.StatusForbidden)
	// un Cliente solo crea pagos PENDIENTE
	como(t, rolClienteT, 1001)
	for _, e := range []string{"PAGADO", "NO_PAGO"} {
		checkout(t, `{"productos":[{"productoId":1,"cantidad":1}],"pago":{"metodoPagoId":2,"estadoPago":"`+e+`"}}`, http.StatusForbidden)
	}
	// nada se escribió en las respuestas 403: solo las dos compras válidas notificaron
	if len(g.eventos) != 2 {
		t.Fatalf("eventos: %+v", g.eventos)
	}
}

func TestCheckoutAutenticacionYCuerpo(t *testing.T) {
	defer resetFake()
	g := grabar(t)
	newCheckoutWorld()
	sinToken(t)
	checkout(t, bodyCheckoutSimple, http.StatusUnauthorized)
	como(t, rolMesero, 5)
	for nombre, body := range map[string]string{
		"vacío":           ``,
		"json roto":       `{`,
		"sin productos":   `{"productos":[],"pago":{"metodoPagoId":2}}`,
		"id no positivo":  `{"productos":[{"productoId":0,"cantidad":1}],"pago":{"metodoPagoId":2}}`,
		"cantidad cero":   `{"productos":[{"productoId":1,"cantidad":0}],"pago":{"metodoPagoId":2}}`,
		"cantidad neg":    `{"productos":[{"productoId":1,"cantidad":-1}],"pago":{"metodoPagoId":2}}`,
		"sin metodo":      `{"productos":[{"productoId":1,"cantidad":1}],"pago":{}}`,
		"restaurante neg": `{"restauranteId":-1,"productos":[{"productoId":1,"cantidad":1}],"pago":{"metodoPagoId":2}}`,
		"cliente neg":     `{"documentoCliente":-5,"productos":[{"productoId":1,"cantidad":1}],"pago":{"metodoPagoId":2}}`,
		"monto neg":       `{"productos":[{"productoId":1,"cantidad":1}],"pago":{"metodoPagoId":2,"monto":-1}}`,
		"fecha pago":      `{"productos":[{"productoId":1,"cantidad":1}],"pago":{"metodoPagoId":2,"fechaPago":"31/01/2025"}}`,
		"hora pago largo": `{"productos":[{"productoId":1,"cantidad":1}],"pago":{"metodoPagoId":2,"horaPago":"1030"}}`,
		"hora pago mala":  `{"productos":[{"productoId":1,"cantidad":1}],"pago":{"metodoPagoId":2,"horaPago":"25:00"}}`,
		"estado pago":     `{"productos":[{"productoId":1,"cantidad":1}],"pago":{"metodoPagoId":2,"estadoPago":"X"}}`,
		"dom sin dir":     `{"domicilio":{"direccion":"  ","telefono":"3"},"productos":[{"productoId":1,"cantidad":1}],"pago":{"metodoPagoId":2}}`,
		"dom sin tel":     `{"domicilio":{"direccion":"d","telefono":""},"productos":[{"productoId":1,"cantidad":1}],"pago":{"metodoPagoId":2}}`,
		"dom fecha":       `{"domicilio":{"direccion":"d","telefono":"3","fechaDomicilio":"hoy"},"productos":[{"productoId":1,"cantidad":1}],"pago":{"metodoPagoId":2}}`,
	} {
		t.Run(nombre, func(t *testing.T) { checkout(t, body, http.StatusBadRequest) })
	}
	if len(fakeLog) != 0 || len(g.eventos) != 0 {
		t.Fatalf("una petición inválida no debe tocar la BD: %+v", fakeLog)
	}
	// fecha del domicilio explícita
	checkout(t, `{"domicilio":{"direccion":"d","telefono":"3","fechaDomicilio":"2025-03-01"},"productos":[{"productoId":1,"cantidad":1}],"pago":{"metodoPagoId":2}}`, http.StatusCreated)
}

func TestCheckoutReferenciasInexistentesSon404(t *testing.T) {
	defer resetFake()
	g := grabar(t)
	como(t, rolClienteT, 1001)
	w := newCheckoutWorld()
	body := `{"restauranteId":1,"productos":[{"productoId":1,"cantidad":1}],"pago":{"metodoPagoId":2}}`
	w.restaurante = false
	contains(t, checkout(t, body, http.StatusNotFound), "Restaurante no encontrado")
	nada(t, g)
	w.restaurante, w.metodo = true, false
	fakeRollbacks = 0
	contains(t, checkout(t, body, http.StatusNotFound), "Método de pago no encontrado")
	w.metodo, w.cliente = true, false
	fakeRollbacks = 0
	contains(t, checkout(t, body, http.StatusNotFound), "Cliente no encontrado")
	if len(escritos()) != 0 {
		t.Fatalf("no debe escribir: %+v", escritos())
	}
}

func TestParseHoraPago(t *testing.T) {
	for _, s := range []string{"10:30", "10:30:15"} {
		h, err := parseHoraPago(s)
		if err != nil || h.Year() != 2000 {
			t.Fatalf("%s: %v %v", s, h, err)
		}
	}
	for _, s := range []string{"", "1030", "25:00", "10:30:99"} {
		if _, err := parseHoraPago(s); err == nil {
			t.Fatalf("%q debía fallar", s)
		}
	}
}

func TestPlanearCheckoutUsaBogota(t *testing.T) {
	// 23:30 del 31-ene en Bogotá: el pago y el pedido llevan esa fecha y hora (no la de UTC)
	bogota := time.FixedZone("UTC-5", -5*60*60)
	ahora := time.Date(2025, 1, 31, 23, 30, 5, 0, bogota)
	claims := &login.Claims{Documento: 1001, Rol: rolClienteT}
	plan, e := planearCheckout(claims, &models.CheckoutRequest{Productos: []models.ProductoPedidoItemInput{{ProductoId: 1, Cantidad: 1}}, Pago: models.CheckoutPago{MetodoPagoId: 2}}, ahora)
	if e != nil {
		t.Fatalf("%+v", e)
	}
	if plan.pago.fecha.Format("2006-01-02") != "2025-01-31" || plan.hoy != plan.pago.fecha || plan.pago.hora.Format("15:04:05") != "23:30:05" {
		t.Fatalf("pago: %+v hoy=%v", plan.pago, plan.hoy)
	}
}
