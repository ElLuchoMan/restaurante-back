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
	}
}

func cuponVals() map[string]driver.Value {
	return map[string]driver.Value{
		"pk_id_cupon": int64(1), "codigo": "VERANO10", "scope": "GLOBAL", "tipo_descuento": "PORCENTAJE", "valor_descuento": int64(10),
		"fecha_inicio": time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), "fecha_fin": time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC), "activo": true,
	}
}

func ofertaVals() map[string]driver.Value {
	return map[string]driver.Value{
		"pk_id_oferta": int64(2), "titulo": "Martes", "tipo_descuento": "MONTO", "valor_descuento": int64(500),
		"fecha_inicio": time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), "fecha_fin": time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC), "dias_semana": "{Martes}", "activo": true,
	}
}

// setup: pedido 9 existente, cupón 1 y oferta 2 existentes, sin descuentos aplicados.
func setup() *db {
	d := newDB().install()
	d.rows["pedido"] = []map[string]driver.Value{pedidoVals()}
	d.rows["cupon"] = []map[string]driver.Value{cuponVals()}
	d.rows["oferta"] = []map[string]driver.Value{ofertaVals()}
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

func TestPost(t *testing.T) {
	defer resetFake()
	for _, q := range []string{"", "?pedido_id=x", "?pedido_id=0"} {
		call(t, http.MethodPost, "/descuentos/pedidos"+q, `{"cuponId":1}`, post, http.StatusBadRequest)
	}
	url := "/descuentos/pedidos?pedido_id=9"
	call(t, http.MethodPost, url, "", post, http.StatusBadRequest)
	call(t, http.MethodPost, url, "nojson", post, http.StatusBadRequest)
	call(t, http.MethodPost, url, `{"cuponId":0}`, post, http.StatusBadRequest)
	call(t, http.MethodPost, url, `{"ofertaId":-1}`, post, http.StatusBadRequest)
	for _, b := range []string{`{}`, `{"cuponId":1,"ofertaId":2}`, `{"cuponId":1,"montoDescuento":-5}`, `{"cuponId":1,"detalle":[1]}`, `{"cuponId":1,"detalle":"x"}`} {
		call(t, http.MethodPost, url, b, post, http.StatusUnprocessableEntity)
	}

	d := newDB().install()
	call(t, http.MethodPost, url, `{"cuponId":1,"montoDescuento":1000}`, post, http.StatusNotFound) // pedido
	d = setup()
	d.rows["cupon"] = nil
	call(t, http.MethodPost, url, `{"cuponId":1,"montoDescuento":1000}`, post, http.StatusNotFound)
	d.rows["oferta"] = nil
	call(t, http.MethodPost, url, `{"ofertaId":2,"montoDescuento":1000}`, post, http.StatusNotFound)

	d = setup()
	b := call(t, http.MethodPost, url, `{"cuponId":1,"montoDescuento":1000,"detalle":{"nota":"x"}}`, post, http.StatusCreated)
	contains(t, b, `"pedidoDescuentoId":7`, `"montoDescuento":1000`, `"tipo":"cupon"`, `"nota":"x"`, `"codigo":"VERANO10"`)
	b = call(t, http.MethodPost, url, `{"ofertaId":2,"montoDescuento":500}`, post, http.StatusCreated)
	contains(t, b, `"tipo":"oferta"`, `"titulo":"Martes"`)

	d.counts["pedido_descuento_aplicado"] = 1
	call(t, http.MethodPost, url, `{"cuponId":1,"montoDescuento":1000}`, post, http.StatusConflict)
	d.counts["pedido_descuento_aplicado"] = 0

	execErrOn("INSERT INTO", "duplicate key value (23505)")
	call(t, http.MethodPost, url, `{"cuponId":1,"montoDescuento":1000}`, post, http.StatusConflict)
	resetFake()
	d = setup()
	execErrOn("INSERT INTO", "boom")
	call(t, http.MethodPost, url, `{"cuponId":1,"montoDescuento":1000}`, post, http.StatusInternalServerError)
	d.errs["pedido"] = errors.New("boom")
	call(t, http.MethodPost, url, `{"cuponId":1,"montoDescuento":1000}`, post, http.StatusInternalServerError)
}
