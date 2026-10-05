package descuento

import (
	"database/sql/driver"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"
)

func pedidoVals() map[string]driver.Value {
	return map[string]driver.Value{
		"pk_id_pedido": int64(9), "fecha": time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC), "hora": time.Date(2000, 1, 1, 18, 30, 0, 0, time.UTC),
		"delivery": false, "estado_pedido": "INICIADO", "updated_at": time.Date(2025, 1, 31, 18, 30, 0, 0, time.UTC),
		"pk_documento_cliente": int64(7),
	}
}

func cuponVals() map[string]driver.Value {
	return map[string]driver.Value{
		"pk_id_cupon": int64(1), "codigo": "VERANO10", "scope": "GLOBAL", "tipo_descuento": "PORCENTAJE", "valor_descuento": int64(10),
		"fecha_inicio": time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), "fecha_fin": time.Date(2999, 12, 31, 0, 0, 0, 0, time.UTC), "activo": true,
	}
}

func ofertaVals() map[string]driver.Value {
	return map[string]driver.Value{
		"pk_id_oferta": int64(2), "titulo": "Martes", "tipo_descuento": "MONTO", "valor_descuento": int64(500),
		"fecha_inicio": time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), "fecha_fin": time.Date(2999, 12, 31, 0, 0, 0, 0, time.UTC), "dias_semana": "", "activo": true,
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

// setup: pedido 9 existente, cupón 1 y oferta 2 existentes, sin descuentos aplicados.
func setup() *db {
	d := newDB().install()
	d.rows["pedido"] = []map[string]driver.Value{pedidoVals()}
	d.rows["cupon"] = []map[string]driver.Value{cuponVals()}
	d.rows["oferta"] = []map[string]driver.Value{ofertaVals()}
	d.rows["oferta_producto"] = []map[string]driver.Value{{"pk_id_oferta": int64(2), "pk_id_producto": int64(3)}}
	d.rows["detalle_pedido"] = []map[string]driver.Value{{"pk_id_producto": int64(3), "cantidad": int64(2), "precio": int64(5000)}}
	return d
}

func execErrOn(sub, msg string) {
	fakeExec = func(q string, _ []driver.NamedValue) (driver.Result, error) {
		if strings.Contains(q, sub) {
			return nil, errors.New(msg)
		}
		return fakeResult{}, nil
	}
}

func call(t *testing.T, method, target, body string, f func(c *DescuentoController), status int) string {
	t.Helper()
	ctx, w := newCtx(method, target, body)
	c := &DescuentoController{}
	c.Ctx, c.Data = ctx, map[interface{}]interface{}{}
	f(c)
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
	getAll = func(c *DescuentoController) { c.GetAll() }
	post   = func(c *DescuentoController) { c.Post() }
)

func TestGetAll(t *testing.T) {
	defer resetFake()
	defer asAdmin()
	for _, q := range []string{"", "?pedido_id=x", "?pedido_id=0"} {
		call(t, http.MethodGet, "/descuentos/pedidos"+q, "", getAll, http.StatusBadRequest)
	}
	d := newDB().install()
	call(t, http.MethodGet, "/descuentos/pedidos?pedido_id=9", "", getAll, http.StatusNotFound)
	d.errs["pedido"] = errors.New("boom")
	call(t, http.MethodGet, "/descuentos/pedidos?pedido_id=9", "", getAll, http.StatusInternalServerError)

	d = setup()
	if b := call(t, http.MethodGet, "/descuentos/pedidos?pedido_id=9", "", getAll, http.StatusOK); !strings.Contains(b, `"data":[]`) {
		t.Fatalf("sin descuentos debe ser []: %s", b)
	}
	d.rows["pedido_descuento_aplicado"] = []map[string]driver.Value{
		{"pk_id_pedido_descuento": int64(5), "T0.pk_id_pedido": int64(9), "T0.pk_id_cupon": int64(1), "monto_descuento": int64(1000),
			"detalle": `{"tipo":"cupon"}`, "created_at": time.Date(2025, 1, 31, 18, 30, 0, 0, time.UTC),
			"T1.pk_id_pedido": int64(9), "T1.fecha": time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC), "T1.hora": time.Date(2000, 1, 1, 18, 30, 0, 0, time.UTC),
			"T1.delivery": false, "T1.estado_pedido": "INICIADO", "T1.updated_at": time.Date(2025, 1, 31, 18, 30, 0, 0, time.UTC),
			"T2.pk_id_cupon": int64(1), "T2.codigo": "VERANO10", "T2.scope": "GLOBAL", "T2.tipo_descuento": "PORCENTAJE", "T2.valor_descuento": int64(10),
			"T2.fecha_inicio": time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), "T2.fecha_fin": time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC), "T2.activo": true,
		},
		{"pk_id_pedido_descuento": int64(6), "T0.pk_id_oferta": int64(2), "monto_descuento": int64(500), "created_at": time.Date(2025, 1, 31, 18, 30, 0, 0, time.UTC),
			"T3.pk_id_oferta": int64(2), "T3.titulo": "Martes", "T3.tipo_descuento": "MONTO", "T3.valor_descuento": int64(500), "T3.dias_semana": "{Martes}",
			"T3.fecha_inicio": time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), "T3.fecha_fin": time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC), "T3.activo": true,
		},
	}
	b := call(t, http.MethodGet, "/descuentos/pedidos?pedido_id=9", "", getAll, http.StatusOK)
	contains(t, b, `"pedidoDescuentoId":5`, `"codigo":"VERANO10"`, `"detalle":{"tipo":"cupon"}`, `"pedidoDescuentoId":6`, `"titulo":"Martes"`, `"diasSemana":["Martes"]`)
	d.errs["pedido_descuento_aplicado"] = errors.New("boom")
	call(t, http.MethodGet, "/descuentos/pedidos?pedido_id=9", "", getAll, http.StatusInternalServerError)
}

func TestGetAll_Permisos(t *testing.T) {
	defer resetFake()
	defer asAdmin()
	setup()
	const url = "/descuentos/pedidos?pedido_id=9"
	as(7, "Cliente") // dueño del pedido
	call(t, http.MethodGet, url, "", getAll, http.StatusOK)
	as(8, "Cliente") // otro cliente
	call(t, http.MethodGet, url, "", getAll, http.StatusForbidden)
	as(3, "Mesero") // trabajador: cualquier pedido
	call(t, http.MethodGet, url, "", getAll, http.StatusOK)
	// Pedido sin cliente: un cliente no lo ve.
	d := setup()
	d.rows["pedido"][0]["pk_documento_cliente"] = nil
	as(7, "Cliente")
	call(t, http.MethodGet, url, "", getAll, http.StatusForbidden)
	actor.doc = 0
	call(t, http.MethodGet, url, "", getAll, http.StatusUnauthorized)
}

func TestPost(t *testing.T) {
	defer resetFake()
	defer asAdmin()
	for _, q := range []string{"", "?pedido_id=x", "?pedido_id=0"} {
		call(t, http.MethodPost, "/descuentos/pedidos"+q, `{"clienteId":7,"cuponId":1}`, post, http.StatusBadRequest)
	}
	url := "/descuentos/pedidos?pedido_id=9"
	call(t, http.MethodPost, url, "", post, http.StatusBadRequest)
	call(t, http.MethodPost, url, "nojson", post, http.StatusBadRequest)
	call(t, http.MethodPost, url, `{"clienteId":0,"cuponId":1}`, post, http.StatusBadRequest) // trabajador sin clienteId
	call(t, http.MethodPost, url, `{"clienteId":7,"cuponId":0}`, post, http.StatusBadRequest)
	call(t, http.MethodPost, url, `{"clienteId":7,"ofertaId":-1}`, post, http.StatusBadRequest)
	for _, b := range []string{`{"clienteId":7}`, `{"clienteId":7,"cuponId":1,"ofertaId":2}`, `{"clienteId":7,"cuponId":1,"detalle":[1]}`, `{"clienteId":7,"cuponId":1,"detalle":"x"}`} {
		call(t, http.MethodPost, url, b, post, http.StatusUnprocessableEntity)
	}

	d := newDB().install()
	call(t, http.MethodPost, url, `{"clienteId":7,"cuponId":1}`, post, http.StatusNotFound) // pedido
	d = setup()
	d.rows["cupon"] = nil
	call(t, http.MethodPost, url, `{"clienteId":7,"cuponId":1}`, post, http.StatusNotFound)
	d.rows["oferta"] = nil
	call(t, http.MethodPost, url, `{"clienteId":7,"ofertaId":2}`, post, http.StatusNotFound)

	// Éxito: el monto lo calcula el servidor (10 % de 2 x 5000), aunque el cliente envíe otro.
	d = setup()
	b := call(t, http.MethodPost, url, `{"clienteId":7,"cuponId":1,"montoDescuento":99999,"detalle":{"nota":"x"}}`, post, http.StatusCreated)
	contains(t, b, `"pedidoDescuentoId":7`, `"montoDescuento":1000`, `"subtotal":10000`, `"total":9000`, `"tipo":"cupon"`, `"nota":"x"`, `"codigo":"VERANO10"`)
	if strings.Contains(b, "99999") {
		t.Fatalf("no debe usar el monto del cliente: %s", b)
	}
	b = call(t, http.MethodPost, url, `{"clienteId":7,"ofertaId":2}`, post, http.StatusCreated)
	contains(t, b, `"tipo":"oferta"`, `"titulo":"Martes"`, `"montoDescuento":500`, `"total":9500`)

	// Conflictos y no aplicables.
	d.counts["pedido_descuento_aplicado"] = 1
	call(t, http.MethodPost, url, `{"clienteId":7,"cuponId":1}`, post, http.StatusConflict)
	d.counts["pedido_descuento_aplicado"] = 0
	d.counts["cupon_redencion"] = 1
	call(t, http.MethodPost, url, `{"clienteId":7,"cuponId":1}`, post, http.StatusConflict)
	d.counts["cupon_redencion"] = 0
	d.rows["cupon"] = []map[string]driver.Value{with(cuponVals(), "activo", false)}
	call(t, http.MethodPost, url, `{"clienteId":7,"cuponId":1}`, post, http.StatusUnprocessableEntity)
	d.rows["oferta"] = []map[string]driver.Value{with(ofertaVals(), "activo", false)}
	call(t, http.MethodPost, url, `{"clienteId":7,"ofertaId":2}`, post, http.StatusUnprocessableEntity)
	d.rows["cupon"] = []map[string]driver.Value{cuponVals()}
	for estado, status := range map[string]int{"CANCELADO": http.StatusConflict, "TERMINADO": http.StatusConflict} {
		d.rows["pedido"][0]["estado_pedido"] = estado
		call(t, http.MethodPost, url, `{"clienteId":7,"cuponId":1}`, post, status)
	}
	d.rows["pedido"][0]["estado_pedido"] = "INICIADO"

	// Pedido ya pagado: 409.
	d.rows["pedido"][0]["pk_id_pago"] = int64(4)
	d.rows["pago"] = []map[string]driver.Value{{"pk_id_pago": int64(4), "monto": int64(10000), "estado_pago": "PAGADO"}}
	call(t, http.MethodPost, url, `{"clienteId":7,"cuponId":1}`, post, http.StatusConflict)
	d.rows["pago"][0]["estado_pago"] = "PENDIENTE"
	b = call(t, http.MethodPost, url, `{"clienteId":7,"cuponId":1}`, post, http.StatusCreated)
	contains(t, b, `"pagoId":4`, `"total":9000`)

	execErrOn("INSERT INTO", "duplicate key value (23505)")
	call(t, http.MethodPost, url, `{"clienteId":7,"cuponId":1}`, post, http.StatusConflict)
	resetFake()
	d = setup()
	execErrOn("INSERT INTO", "boom")
	call(t, http.MethodPost, url, `{"clienteId":7,"cuponId":1}`, post, http.StatusInternalServerError)
	d.errs["pedido"] = errors.New("boom")
	call(t, http.MethodPost, url, `{"clienteId":7,"cuponId":1}`, post, http.StatusInternalServerError)
}

func TestPost_ClienteDelToken(t *testing.T) {
	defer resetFake()
	defer asAdmin()
	setup()
	const url = "/descuentos/pedidos?pedido_id=9"
	as(7, "Cliente")
	// Sin clienteId: manda el documento del token.
	call(t, http.MethodPost, url, `{"cuponId":1}`, post, http.StatusCreated)
	// clienteId distinto al del token: 403, aunque sea el dueño real del pedido.
	call(t, http.MethodPost, url, `{"clienteId":8,"cuponId":1}`, post, http.StatusForbidden)
	// Pedido de otro cliente: 403.
	as(8, "Cliente")
	call(t, http.MethodPost, url, `{"cuponId":1}`, post, http.StatusForbidden)
	// Un trabajador puede actuar en nombre del dueño, pero no de otro cliente.
	as(3, "Mesero")
	call(t, http.MethodPost, url, `{"clienteId":7,"cuponId":1}`, post, http.StatusCreated)
	call(t, http.MethodPost, url, `{"clienteId":8,"cuponId":1}`, post, http.StatusForbidden)
	// Sin token: 401.
	actor.doc = 0
	call(t, http.MethodPost, url, `{"cuponId":1}`, post, http.StatusUnauthorized)
}
