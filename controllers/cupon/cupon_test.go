package cupon

import (
	"database/sql/driver"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"
)

func day(y int, m time.Month, d int) time.Time { return time.Date(y, m, d, 0, 0, 0, 0, time.UTC) }

// vigente es una fecha de fin lejana para que el cupón esté siempre vigente.
func cuponVals() map[string]driver.Value {
	return map[string]driver.Value{
		"pk_id_cupon":     int64(1),
		"codigo":          "VERANO10",
		"scope":           "GLOBAL",
		"tipo_descuento":  "PORCENTAJE",
		"valor_descuento": int64(10),
		"fecha_inicio":    day(2020, 1, 1),
		"fecha_fin":       day(2999, 12, 31),
		"activo":          true,
	}
}

func with(base map[string]driver.Value, kv ...interface{}) map[string]driver.Value {
	out := map[string]driver.Value{}
	for k, v := range base {
		out[k] = v
	}
	for i := 0; i < len(kv); i += 2 {
		out[kv[i].(string)] = kv[i+1]
	}
	return out
}

// setup devuelve una base con un cupón GLOBAL vigente, el cliente 7, el pedido 9
// y un detalle de pedido (producto 2, 2 x 5000).
func setup() (*db, *[]string, *[]driver.NamedValue) {
	d := newDB().install()
	d.rows["cupon"] = []map[string]driver.Value{cuponVals()}
	d.rows["pedido"] = []map[string]driver.Value{{"pk_id_pedido": int64(9), "estado_pedido": "INICIADO", "pk_documento_cliente": int64(7)}}
	d.rows["detalle_pedido"] = []map[string]driver.Value{{"pk_id_producto": int64(2), "cantidad": int64(2), "precio": int64(5000)}}
	var execs []string
	var last []driver.NamedValue
	fakeExec = func(q string, a []driver.NamedValue) (driver.Result, error) {
		execs = append(execs, q)
		last = a
		return fakeResult{}, nil
	}
	return d, &execs, &last
}

func execErrOn(sub, msg string) {
	prev := fakeExec
	fakeExec = func(q string, a []driver.NamedValue) (driver.Result, error) {
		if strings.Contains(q, sub) {
			return nil, errors.New(msg)
		}
		return prev(q, a)
	}
}

func call(t *testing.T, method, target, body string, f func(c *CuponController), status int) string {
	t.Helper()
	ctx, w := newCtx(method, target, body)
	c := &CuponController{}
	c.Ctx, c.Data = ctx, map[interface{}]interface{}{}
	f(c)
	expect(t, w, status)
	return w.Body.String()
}

func callCodigo(t *testing.T, codigo, body string, status int) string {
	t.Helper()
	ctx, w := newCtx(http.MethodPost, "/cupones/x/redimir", body)
	ctx.Input.SetParam(":codigo", codigo)
	c := &CuponController{}
	c.Ctx, c.Data = ctx, map[interface{}]interface{}{}
	c.RedimirCupon()
	expect(t, w, status)
	return w.Body.String()
}

func contains(t *testing.T, b string, parts ...string) {
	t.Helper()
	for _, p := range parts {
		if !strings.Contains(b, p) {
			t.Fatalf("falta %s en %s", p, b)
		}
	}
}

var (
	getAll   = func(c *CuponController) { c.GetAll() }
	getByID  = func(c *CuponController) { c.GetById() }
	post     = func(c *CuponController) { c.Post() }
	put      = func(c *CuponController) { c.Put() }
	del      = func(c *CuponController) { c.Delete() }
	validar  = func(c *CuponController) { c.ValidarCupon() }
	redencio = func(c *CuponController) { c.ListarRedenciones() }
)

func TestGetAll(t *testing.T) {
	defer resetFake()
	newDB().install()
	if b := call(t, http.MethodGet, "/cupones", "", getAll, http.StatusOK); !strings.Contains(b, `"data":[]`) {
		t.Fatalf("lista vacía debe ser []: %s", b)
	}
	d, _, _ := setup()
	d.counts["cupon"] = 1
	d.rows["cupon"][0] = with(cuponVals(), "scope", "PRODUCTO", "T0.pk_id_producto", int64(2), "T1.pk_id_producto", int64(2), "T1.nombre", "Bandeja", "T1.imagen", []byte("img"))
	b := call(t, http.MethodGet, "/cupones?activo=true&codigo=ver&scope=PRODUCTO&fecha_desde=2020-01-01&fecha_hasta=2999-12-31&limit=500", "", getAll, http.StatusOK)
	contains(t, b, `"codigo":"VERANO10"`, `"nombre":"Bandeja"`, `"productoId":2`, `"pageSize":100`, `"total":1`)
	if strings.Contains(b, "aW1n") {
		t.Fatalf("no debe incluir la imagen del producto: %s", b)
	}
	for _, q := range []string{"?activo=quizas", "?scope=otro", "?fecha_desde=x", "?fecha_hasta=x", "?limit=0", "?limit=x", "?offset=-1", "?offset=x"} {
		call(t, http.MethodGet, "/cupones"+q, "", getAll, http.StatusBadRequest)
	}
	d.errs["cupon"] = errors.New("boom")
	call(t, http.MethodGet, "/cupones", "", getAll, http.StatusInternalServerError)
	delete(d.errs, "cupon")
	prev := fakeQuery
	fakeQuery = func(q string, a []driver.NamedValue) (driver.Rows, error) {
		if !strings.Contains(q, "COUNT(") {
			return nil, errors.New("boom")
		}
		return prev(q, a)
	}
	call(t, http.MethodGet, "/cupones", "", getAll, http.StatusInternalServerError)
}

const okPost = `{"codigo":" VERANO10 ","scope":"GLOBAL","tipoDescuento":"PORCENTAJE","valorDescuento":10,"fechaInicio":"2025-01-01","fechaFin":"2025-12-31","maxUsos":100,"limitePorCliente":1,"montoMinimo":20000}`

func TestPost(t *testing.T) {
	defer resetFake()
	call(t, http.MethodPost, "/cupones", "", post, http.StatusBadRequest)
	call(t, http.MethodPost, "/cupones", "nojson", post, http.StatusBadRequest)
	mut := func(old, new string) string { return strings.Replace(okPost, old, new, 1) }
	for _, b := range []string{
		mut(`"GLOBAL"`, `"OTRO"`),
		mut(`"PORCENTAJE"`, `"OTRO"`),
		mut(`"2025-01-01"`, `"x"`),
		mut(`"2025-12-31"`, `"x"`),
		mut(`" VERANO10 "`, `"ab"`),
		mut(`" VERANO10 "`, `"`+strings.Repeat("x", 51)+`"`),
		mut(`"valorDescuento":10`, `"valorDescuento":150`),
		mut(`"scope":"GLOBAL"`, `"scope":"PRODUCTO"`),
		mut(`"maxUsos":100`, `"maxUsos":100,"productoId":2`),
	} {
		call(t, http.MethodPost, "/cupones", b, post, http.StatusUnprocessableEntity)
	}

	d, _, args := setup()
	b := call(t, http.MethodPost, "/cupones", okPost, post, http.StatusCreated)
	contains(t, b, `"cuponId":1`, `"codigo":"VERANO10"`)
	for _, a := range *args {
		if a.Value == " VERANO10 " {
			t.Fatalf("el código debe guardarse sin espacios: %v", *args)
		}
	}
	// Scopes con relación.
	call(t, http.MethodPost, "/cupones", `{"codigo":"PROD1","scope":"PRODUCTO","tipoDescuento":"MONTO","valorDescuento":500,"fechaInicio":"2025-01-01","fechaFin":"2025-12-31","productoId":2}`, post, http.StatusCreated)
	call(t, http.MethodPost, "/cupones", `{"codigo":"CAT1","scope":"CATEGORIA","tipoDescuento":"MONTO","valorDescuento":500,"fechaInicio":"2025-01-01","fechaFin":"2025-12-31","categoriaId":2}`, post, http.StatusCreated)
	call(t, http.MethodPost, "/cupones", `{"codigo":"CLI1","scope":"CLIENTE","tipoDescuento":"MONTO","valorDescuento":500,"fechaInicio":"2025-01-01","fechaFin":"2025-12-31","documentoCliente":7}`, post, http.StatusCreated)

	execErrOn("INSERT INTO", "duplicate key value (23505)")
	call(t, http.MethodPost, "/cupones", okPost, post, http.StatusConflict)
	resetFake()
	setup()
	execErrOn("INSERT INTO", "violates foreign key constraint (23503)")
	call(t, http.MethodPost, "/cupones", okPost, post, http.StatusBadRequest)
	resetFake()
	setup()
	execErrOn("INSERT INTO", "boom")
	call(t, http.MethodPost, "/cupones", okPost, post, http.StatusInternalServerError)
	resetFake()
	d, _, _ = setup()
	d.rows["cupon"] = nil
	call(t, http.MethodPost, "/cupones", okPost, post, http.StatusNotFound) // no se puede releer
}

func TestGetById(t *testing.T) {
	defer resetFake()
	for _, q := range []string{"", "?id=%20"} {
		call(t, http.MethodGet, "/cupones/search"+q, "", getByID, http.StatusBadRequest)
	}
	d := newDB().install()
	call(t, http.MethodGet, "/cupones/search?id=1", "", getByID, http.StatusNotFound)
	call(t, http.MethodGet, "/cupones/search?id=0", "", getByID, http.StatusNotFound)
	d, _, _ = setup()
	contains(t, call(t, http.MethodGet, "/cupones/search?id=1", "", getByID, http.StatusOK), `"cuponId":1`)
	contains(t, call(t, http.MethodGet, "/cupones/search?id=VERANO10", "", getByID, http.StatusOK), `"codigo":"VERANO10"`)
	d.rows["cupon"][0] = with(cuponVals(), "scope", "CLIENTE", "T0.pk_documento_cliente", int64(7), "T3.pk_documento_cliente", int64(7), "T3.nombre", "Ana", "T3.password", "hash-secreto")
	b := call(t, http.MethodGet, "/cupones/search?id=1", "", getByID, http.StatusOK)
	contains(t, b, `"documentoCliente":{`, `"nombre":"Ana"`)
	if strings.Contains(strings.ToLower(b), "password") || strings.Contains(b, "hash-secreto") {
		t.Fatalf("no debe filtrar la contraseña: %s", b)
	}
	d.errs["cupon"] = errors.New("boom")
	call(t, http.MethodGet, "/cupones/search?id=1", "", getByID, http.StatusInternalServerError)
	call(t, http.MethodGet, "/cupones/search?id=VERANO10", "", getByID, http.StatusInternalServerError)
}

func TestPut(t *testing.T) {
	defer resetFake()
	call(t, http.MethodPut, "/cupones", `{}`, put, http.StatusBadRequest)
	d := newDB().install()
	call(t, http.MethodPut, "/cupones?id=1", `{}`, put, http.StatusNotFound)
	d.errs["cupon"] = errors.New("boom")
	call(t, http.MethodPut, "/cupones?id=1", `{}`, put, http.StatusInternalServerError)

	d, execs, _ := setup()
	for _, b := range []string{"", "nojson", `{"codigo":null}`, `{"scope":null}`, `{"valorDescuento":null}`, `{"fechaInicio":null}`, `{"activo":null}`, `{"maxUsos":"x"}`} {
		call(t, http.MethodPut, "/cupones?id=1", b, put, http.StatusBadRequest)
	}
	for _, b := range []string{
		`{"scope":"OTRO"}`, `{"tipoDescuento":"OTRO"}`, `{"fechaInicio":"x"}`, `{"fechaFin":"x"}`,
		`{"codigo":"ab"}`, `{"valorDescuento":500}`, `{"scope":"PRODUCTO"}`, `{"productoId":2}`, `{"fechaFin":"2000-01-01"}`,
	} {
		call(t, http.MethodPut, "/cupones?id=1", b, put, http.StatusUnprocessableEntity)
	}

	contains(t, call(t, http.MethodPut, "/cupones?id=1", `{}`, put, http.StatusOK), `"codigo":"VERANO10"`)
	if len(*execs) == 0 {
		t.Fatal("debe actualizar")
	}
	call(t, http.MethodPut, "/cupones?id=1", `{"codigo":" NUEVO ","tipoDescuento":"MONTO","valorDescuento":500,"fechaInicio":"2025-01-01","fechaFin":"2025-12-31","maxUsos":5,"limitePorCliente":1,"montoMinimo":10,"activo":false}`, put, http.StatusOK)
	// null limpia los campos anulables; cambio de scope con su relación.
	d.rows["cupon"][0] = with(cuponVals(), "max_usos", int64(5), "limite_por_cliente", int64(1), "monto_minimo", int64(10), "scope", "PRODUCTO", "T0.pk_id_producto", int64(2))
	call(t, http.MethodPut, "/cupones?id=1", `{"maxUsos":null,"limitePorCliente":null,"montoMinimo":null,"scope":"GLOBAL","productoId":null,"categoriaId":null,"documentoCliente":null}`, put, http.StatusOK)
	call(t, http.MethodPut, "/cupones?id=1", `{"productoId":3}`, put, http.StatusOK)
	call(t, http.MethodPut, "/cupones?id=1", `{"scope":"CATEGORIA","productoId":null,"categoriaId":4}`, put, http.StatusOK)
	call(t, http.MethodPut, "/cupones?id=1", `{"scope":"CLIENTE","productoId":null,"documentoCliente":7}`, put, http.StatusOK)

	execErrOn(`UPDATE "cupon"`, "duplicate key value (23505)")
	call(t, http.MethodPut, "/cupones?id=1", `{"codigo":"OTRO1"}`, put, http.StatusConflict)
	resetFake()
	setup()
	execErrOn(`UPDATE "cupon"`, "violates foreign key constraint (23503)")
	call(t, http.MethodPut, "/cupones?id=1", `{"codigo":"OTRO1"}`, put, http.StatusBadRequest)
	resetFake()
	setup()
	execErrOn(`UPDATE "cupon"`, "boom")
	call(t, http.MethodPut, "/cupones?id=1", `{"codigo":"OTRO1"}`, put, http.StatusInternalServerError)

	resetFake()
	setup()
	n := 0
	prev := fakeQuery
	fakeQuery = func(q string, a []driver.NamedValue) (driver.Rows, error) {
		n++
		if n > 1 {
			return nil, errors.New("boom")
		}
		return prev(q, a)
	}
	call(t, http.MethodPut, "/cupones?id=1", `{"codigo":"OTRO1"}`, put, http.StatusInternalServerError)
}

func TestDelete(t *testing.T) {
	defer resetFake()
	for _, q := range []string{"", "?id=x", "?id=0"} {
		call(t, http.MethodDelete, "/cupones"+q, "", del, http.StatusBadRequest)
	}
	d := newDB().install()
	call(t, http.MethodDelete, "/cupones?id=1", "", del, http.StatusNotFound)
	d.errs["cupon"] = errors.New("boom")
	call(t, http.MethodDelete, "/cupones?id=1", "", del, http.StatusInternalServerError)
	_, execs, _ := setup()
	call(t, http.MethodDelete, "/cupones?id=1", "", del, http.StatusOK)
	if len(*execs) != 1 || !strings.Contains((*execs)[0], `UPDATE "cupon"`) {
		t.Fatalf("debe desactivar: %v", *execs)
	}
	execErrOn(`UPDATE "cupon"`, "boom")
	call(t, http.MethodDelete, "/cupones?id=1", "", del, http.StatusInternalServerError)
	d, execs, _ = setup()
	d.rows["cupon"][0] = with(cuponVals(), "activo", false)
	call(t, http.MethodDelete, "/cupones?id=1", "", del, http.StatusBadRequest)
	if len(*execs) != 0 {
		t.Fatalf("no debe escribir: %v", *execs)
	}
}

const okValidar = `{"codigo":"VERANO10","clienteId":7,"items":[{"productoId":2,"cantidad":2,"precio":5000}]}`

func TestValidar(t *testing.T) {
	defer resetFake()
	defer asAdmin()
	// Un trabajador debe indicar clienteId; items obligatorios sin pedidoId.
	for _, b := range []string{"", "nojson", `{}`,
		`{"codigo":"X","clienteId":0,"items":[{"productoId":2,"cantidad":1,"precio":1}]}`,
		`{"codigo":"X","clienteId":7,"items":[]}`,
		`{"codigo":"X","clienteId":7,"pedidoId":0}`,
		`{"codigo":" ","clienteId":7,"items":[{"productoId":2,"cantidad":1,"precio":1}]}`,
		`{"codigo":"X","clienteId":7,"items":[{"productoId":0,"cantidad":1,"precio":1}]}`,
		`{"codigo":"X","clienteId":7,"items":[{"productoId":2,"cantidad":0,"precio":1}]}`,
		`{"codigo":"X","clienteId":7,"items":[{"productoId":2,"cantidad":1,"precio":-1}]}`,
	} {
		call(t, http.MethodPost, "/cupones/validar", b, validar, http.StatusBadRequest)
	}
	d, _, _ := setup()
	contains(t, call(t, http.MethodPost, "/cupones/validar", okValidar, validar, http.StatusOK), `"aplicable":true`, `"montoDescuento":1000`)
	d.rows["cupon"] = nil
	contains(t, call(t, http.MethodPost, "/cupones/validar", okValidar, validar, http.StatusOK), `"aplicable":false`, `"motivo":"Cupón no encontrado"`)
	d.errs["cupon"] = errors.New("boom")
	call(t, http.MethodPost, "/cupones/validar", okValidar, validar, http.StatusInternalServerError)
}

func TestValidar_ClienteDelToken(t *testing.T) {
	defer resetFake()
	defer asAdmin()
	setup()
	as(7, "Cliente")
	// Sin clienteId: manda el documento del token.
	body := `{"codigo":"VERANO10","items":[{"productoId":2,"cantidad":2,"precio":5000}]}`
	contains(t, call(t, http.MethodPost, "/cupones/validar", body, validar, http.StatusOK), `"aplicable":true`)
	// Igual al del token: permitido.
	contains(t, call(t, http.MethodPost, "/cupones/validar", okValidar, validar, http.StatusOK), `"aplicable":true`)
	// Distinto al del token: 403.
	call(t, http.MethodPost, "/cupones/validar", strings.Replace(okValidar, `"clienteId":7`, `"clienteId":8`, 1), validar, http.StatusForbidden)
	// Sin token: 401.
	asAnon()
	call(t, http.MethodPost, "/cupones/validar", okValidar, validar, http.StatusUnauthorized)
}

func TestValidar_ConPedido(t *testing.T) {
	defer resetFake()
	defer asAdmin()
	d, _, _ := setup()
	as(7, "Cliente")
	// Con pedidoId los ítems del cliente se ignoran y se usa el detalle real (2 x 5000).
	body := `{"codigo":"VERANO10","pedidoId":9,"items":[{"productoId":2,"cantidad":1,"precio":1}]}`
	contains(t, call(t, http.MethodPost, "/cupones/validar", body, validar, http.StatusOK), `"montoDescuento":1000`)
	// Pedido de otro cliente: 403. Inexistente: 404.
	as(8, "Cliente")
	call(t, http.MethodPost, "/cupones/validar", body, validar, http.StatusForbidden)
	as(7, "Cliente")
	d.rows["pedido"] = nil
	call(t, http.MethodPost, "/cupones/validar", body, validar, http.StatusNotFound)
}

func TestRedimir(t *testing.T) {
	defer resetFake()
	defer asAdmin()
	callCodigo(t, "VERANO10", "nojson", http.StatusBadRequest)
	callCodigo(t, "VERANO10", `{"clienteId":0,"pedidoId":9}`, http.StatusBadRequest) // trabajador sin clienteId
	callCodigo(t, "VERANO10", `{"clienteId":7,"pedidoId":0}`, http.StatusBadRequest)
	callCodigo(t, " ", `{"clienteId":7,"pedidoId":9}`, http.StatusBadRequest)
	callCodigo(t, "VERANO10", `{"clienteId":7}`, http.StatusBadRequest) // pedidoId requerido

	d, execs, _ := setup()
	b := callCodigo(t, "VERANO10", `{"clienteId":7,"pedidoId":9}`, http.StatusCreated)
	contains(t, b, `"montoDescuento":1000`, `"cuponRedencionId":7`)
	if len(*execs) == 0 || !d.saw("detalle_pedido") {
		t.Fatalf("debe leer el detalle e insertar la redención: %v %v", *execs, d.seen)
	}

	// Inexistentes: 404.
	saved := d.rows["cupon"]
	d.rows["cupon"] = nil
	callCodigo(t, "NOPE", `{"clienteId":7,"pedidoId":9}`, http.StatusNotFound)
	d.rows["cupon"] = saved
	savedPedido := d.rows["pedido"]
	d.rows["pedido"] = nil
	callCodigo(t, "VERANO10", `{"clienteId":7,"pedidoId":9}`, http.StatusNotFound)
	// Pedido de otro cliente (un trabajador indica un clienteId que no es el dueño): 403.
	d.rows["pedido"] = savedPedido
	callCodigo(t, "VERANO10", `{"clienteId":8,"pedidoId":9}`, http.StatusForbidden)
	// Pedido cerrado: 409.
	d.rows["pedido"] = []map[string]driver.Value{{"pk_id_pedido": int64(9), "estado_pedido": "CANCELADO", "pk_documento_cliente": int64(7)}}
	callCodigo(t, "VERANO10", `{"clienteId":7,"pedidoId":9}`, http.StatusConflict)
	d.rows["pedido"] = savedPedido

	// Conflictos: ya redimido en el pedido / usos agotados / límite por cliente.
	d.counts["cupon_redencion"] = 1
	callCodigo(t, "VERANO10", `{"clienteId":7,"pedidoId":9}`, http.StatusConflict)
	d.counts["cupon_redencion"] = 0
	d.rows["cupon"] = []map[string]driver.Value{with(cuponVals(), "max_usos", int64(0))}
	callCodigo(t, "VERANO10", `{"clienteId":7,"pedidoId":9}`, http.StatusConflict)
	d.rows["cupon"] = []map[string]driver.Value{with(cuponVals(), "limite_por_cliente", int64(0))}
	callCodigo(t, "VERANO10", `{"clienteId":7,"pedidoId":9}`, http.StatusConflict)

	// No aplicable: inactivo, vencido, monto mínimo, cliente distinto.
	for _, v := range []map[string]driver.Value{
		with(cuponVals(), "activo", false),
		with(cuponVals(), "fecha_fin", day(2020, 2, 1)),
		with(cuponVals(), "monto_minimo", int64(999999)),
		with(cuponVals(), "scope", "CLIENTE", "T0.pk_documento_cliente", int64(8)),
	} {
		d.rows["cupon"] = []map[string]driver.Value{v}
		callCodigo(t, "VERANO10", `{"clienteId":7,"pedidoId":9}`, http.StatusUnprocessableEntity)
	}
	d.rows["cupon"] = []map[string]driver.Value{cuponVals()}

	// Errores internos.
	execErrOn("INSERT INTO", "duplicate key value (23505)")
	callCodigo(t, "VERANO10", `{"clienteId":7,"pedidoId":9}`, http.StatusConflict)
	resetFake()
	d, _, _ = setup()
	execErrOn("INSERT INTO", "boom")
	callCodigo(t, "VERANO10", `{"clienteId":7,"pedidoId":9}`, http.StatusInternalServerError)
	d.errs["detalle_pedido"] = errors.New("boom")
	callCodigo(t, "VERANO10", `{"clienteId":7,"pedidoId":9}`, http.StatusInternalServerError)
	// Falla al abrir la transacción: 500.
	resetFake()
	setup()
	fakeBeginErr = errors.New("begin boom")
	callCodigo(t, "VERANO10", `{"clienteId":7,"pedidoId":9}`, http.StatusInternalServerError)
}

func TestRedimir_ClienteDelToken(t *testing.T) {
	defer resetFake()
	defer asAdmin()
	setup()
	as(7, "Cliente")
	// Sin clienteId en el body: se usa el documento del token.
	callCodigo(t, "VERANO10", `{"pedidoId":9}`, http.StatusCreated)
	// Un clienteId distinto al del token se rechaza aunque el pedido sea de ese otro cliente.
	callCodigo(t, "VERANO10", `{"clienteId":8,"pedidoId":9}`, http.StatusForbidden)
	// Un cliente no puede redimir sobre un pedido ajeno: 403.
	as(8, "Cliente")
	callCodigo(t, "VERANO10", `{"pedidoId":9}`, http.StatusForbidden)
	// Sin token: 401.
	asAnon()
	callCodigo(t, "VERANO10", `{"pedidoId":9}`, http.StatusUnauthorized)
}

// Solo el Administrador gestiona y consulta cupones; los demás roles reciben 403
// (sin token, 401) sin tocar la base.
func TestPermisosAdministrador(t *testing.T) {
	defer resetFake()
	defer asAdmin()
	d, execs, _ := setup()
	handlers := []struct {
		name   string
		method string
		f      func(c *CuponController)
		body   string
	}{
		{"GetAll", http.MethodGet, getAll, ""},
		{"Post", http.MethodPost, post, okPost},
		{"GetById", http.MethodGet, getByID, ""},
		{"Put", http.MethodPut, put, `{"activo":false}`},
		{"Delete", http.MethodDelete, del, ""},
		{"ListarRedenciones", http.MethodGet, redencio, ""},
	}
	for _, h := range handlers {
		for _, quien := range []struct {
			doc    int64
			rol    string
			status int
		}{{0, "", http.StatusUnauthorized}, {7, "Cliente", http.StatusForbidden}, {3, "Mesero", http.StatusForbidden}} {
			as(quien.doc, quien.rol)
			if quien.doc == 0 {
				asAnon()
			}
			call(t, h.method, "/cupones?id=1&codigo=VERANO10", h.body, h.f, quien.status)
		}
	}
	if len(*execs) != 0 || len(d.seen) != 0 {
		t.Fatalf("sin permisos no debe tocar la base: %v %v", *execs, d.seen)
	}
}

func TestRedenciones(t *testing.T) {
	defer resetFake()
	for _, q := range []string{"?limit=0", "?offset=-1", "?cupon_id=x", "?cupon_id=0", "?cliente_id=x", "?cliente_id=0"} {
		call(t, http.MethodGet, "/cupones/redenciones"+q, "", redencio, http.StatusBadRequest)
	}
	d := newDB().install()
	if b := call(t, http.MethodGet, "/cupones/redenciones", "", redencio, http.StatusOK); !strings.Contains(b, `"data":[]`) {
		t.Fatalf("debe ser []: %s", b)
	}
	// Código desconocido: página vacía.
	if b := call(t, http.MethodGet, "/cupones/redenciones?cupon_codigo=NOPE", "", redencio, http.StatusOK); !strings.Contains(b, `"data":[]`) {
		t.Fatalf("debe ser []: %s", b)
	}
	d.rows["cupon"] = []map[string]driver.Value{cuponVals()}
	d.counts["cupon_redencion"] = 1
	d.rows["cupon_redencion"] = []map[string]driver.Value{{
		"pk_id_cupon_redencion": int64(4), "T0.pk_id_cupon": int64(1), "T0.pk_documento_cliente": int64(7), "T0.pk_id_pedido": int64(9),
		"monto_descuento": int64(1000), "created_at": time.Date(2025, 1, 31, 18, 30, 0, 0, time.UTC),
		"T1.pk_id_cupon": int64(1), "T2.pk_documento_cliente": int64(7), "T2.nombre": "Ana", "T2.password": "hash-secreto", "T3.pk_id_pedido": int64(9),
	}}
	b := call(t, http.MethodGet, "/cupones/redenciones?cupon_codigo=VERANO10&cupon_id=1&cliente_id=7", "", redencio, http.StatusOK)
	contains(t, b, `"cuponRedencionId":4`, `"montoDescuento":1000`, `"nombre":"Ana"`)
	if strings.Contains(b, "hash-secreto") || strings.Contains(strings.ToLower(b), "password") {
		t.Fatalf("no debe filtrar la contraseña: %s", b)
	}
	d.errs["cupon"] = errors.New("boom")
	call(t, http.MethodGet, "/cupones/redenciones?cupon_codigo=VERANO10", "", redencio, http.StatusInternalServerError)
	delete(d.errs, "cupon")
	d.errs["cupon_redencion"] = errors.New("boom")
	call(t, http.MethodGet, "/cupones/redenciones", "", redencio, http.StatusInternalServerError)
	delete(d.errs, "cupon_redencion")
	prev := fakeQuery
	fakeQuery = func(q string, a []driver.NamedValue) (driver.Rows, error) {
		if !strings.Contains(q, "COUNT(") {
			return nil, errors.New("boom")
		}
		return prev(q, a)
	}
	call(t, http.MethodGet, "/cupones/redenciones", "", redencio, http.StatusInternalServerError)
}
