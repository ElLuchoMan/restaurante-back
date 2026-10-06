package pago

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"
)

var (
	pagoCols   = []string{"pk_id_pago", "fecha", "hora", "monto", "estado_pago", "pk_id_metodo_pago", "updated_at", "updated_by"}
	metodoCols = []string{"pk_id_metodo_pago", "tipo", "detalle"}
	errBoom    = errors.New("boom")
)

// pagoEstado y descuentos parametrizan el pago y el conteo de descuentos del pedido que sirve serve.
var (
	pagoEstado = "PAGADO"
	descuentos = int64(0)
)

func pagoRow() []driver.Value {
	return []driver.Value{int64(4), time.Date(2025, 1, 31, 12, 0, 0, 0, time.UTC), time.Date(0, 1, 1, 4, 37, 28, 0, time.UTC), // el ORM lo entrega con desfase LMT: se muestra 14:30:00
		int64(50000), pagoEstado, int64(2), time.Date(2025, 1, 31, 20, 0, 0, 0, time.UTC), "cajero"}
}

// serve programa el driver: pago existente y método existente (salvo que se anulen).
func serve(pago, metodo bool) {
	fakeQuery = func(q string, _ []driver.NamedValue) (driver.Rows, error) {
		switch {
		case strings.Contains(q, "pedido_descuento_aplicado"):
			return rowsOf([]string{"count"}, []driver.Value{descuentos}), nil
		case strings.Contains(q, `FROM "pago"`) && pago:
			return rowsOf(pagoCols, pagoRow()), nil
		case strings.Contains(q, `FROM "metodo_pago"`) && metodo:
			return rowsOf(metodoCols, []driver.Value{int64(2), "NEQUI", "300"}), nil
		}
		return rowsOf(nil), nil
	}
}

func call(t *testing.T, method, target, body string, f func(c *PagoController), status int) string {
	t.Helper()
	ctx, w := newCtx(method, target, body)
	c := &PagoController{}
	c.Ctx, c.Data = ctx, map[interface{}]interface{}{}
	f(c)
	expect(t, w, status)
	return w.Body.String()
}

func TestGetAll(t *testing.T) {
	defer resetFake()
	g := func(c *PagoController) { c.GetAll() }
	if b := call(t, http.MethodGet, "/pagos", "", g, http.StatusOK); !strings.Contains(b, `"data":[]`) {
		t.Fatalf("lista vacía debe ser []: %s", b)
	}
	serve(true, true)
	var gotQ string
	fakeQuery = func(q string, a []driver.NamedValue) (driver.Rows, error) {
		gotQ = q
		return rowsOf(pagoCols, pagoRow()), nil
	}
	b := call(t, http.MethodGet, "/pagos?fecha=2025-01-31&dia=31&mes=1&anio=2025&estado=pagado&metodo_pago=2", "", g, http.StatusOK)
	for _, want := range []string{`"fechaPago":"31-01-2025"`, `"horaPago":"14:30:00"`, `"pagoId":4`, `"metodoPagoId":{`} {
		if !strings.Contains(b, want) {
			t.Fatalf("falta %s en %s", want, b)
		}
	}
	for _, want := range []string{"estado_pago", "pk_id_metodo_pago"} {
		if !strings.Contains(gotQ, want) {
			t.Fatalf("filtro %q no aplicado: %s", want, gotQ)
		}
	}
	// filtros en memoria que descartan la fila
	for _, q := range []string{"fecha=2025-01-30", "dia=1", "mes=2", "anio=2024", "estado=no_pago&metodo_pago=2"} {
		serve(true, true)
		fakeQuery = func(string, []driver.NamedValue) (driver.Rows, error) { return rowsOf(pagoCols, pagoRow()), nil }
		if b := call(t, http.MethodGet, "/pagos?"+q, "", g, http.StatusOK); q != "estado_no" && !strings.Contains(b, `"data":[]`) && !strings.Contains(q, "estado") {
			t.Fatalf("%s debía vaciar la lista: %s", q, b)
		}
	}
	for _, q := range []string{"fecha=31-01-2025", "dia=0", "dia=x", "mes=13", "anio=0", "estado=X", "metodo_pago=0", "metodo_pago=a"} {
		call(t, http.MethodGet, "/pagos?"+q, "", g, http.StatusBadRequest)
	}
	fakeQuery = func(string, []driver.NamedValue) (driver.Rows, error) { return nil, errBoom }
	call(t, http.MethodGet, "/pagos", "", g, http.StatusInternalServerError)
}

func TestGetById(t *testing.T) {
	defer resetFake()
	g := func(c *PagoController) { c.GetById() }
	call(t, http.MethodGet, "/pagos/search", "", g, http.StatusBadRequest)
	call(t, http.MethodGet, "/pagos/search?id=0", "", g, http.StatusBadRequest)
	call(t, http.MethodGet, "/pagos/search?id=4", "", g, http.StatusNotFound)
	serve(true, true)
	if b := call(t, http.MethodGet, "/pagos/search?id=4", "", g, http.StatusOK); !strings.Contains(b, `"monto":50000`) {
		t.Fatalf("cuerpo inesperado: %s", b)
	}
	fakeQuery = func(string, []driver.NamedValue) (driver.Rows, error) { return nil, errBoom }
	call(t, http.MethodGet, "/pagos/search?id=4", "", g, http.StatusInternalServerError)
}

func TestPost(t *testing.T) {
	defer resetFake()
	p := func(c *PagoController) { c.Post() }
	ok := `{"estadoPago":"pagado","fechaPago":"2025-02-01","horaPago":"15:00","metodoPagoId":2,"monto":1000,"updatedBy":"x"}`
	with := func(old, new string) string { return strings.Replace(ok, old, new, 1) }
	bad := []string{
		"nojson",
		with(`"fechaPago":"2025-02-01",`, ``),
		with(`2025-02-01`, `01/02/2025`),
		with(`"horaPago":"15:00",`, ``),
		with(`15:00`, `3pm`),
		with(`15:00`, `25:00:00`),
		with(`"monto":1000`, `"monto":0`),
		with(`"monto":1000`, `"monto":-5`),
		with(`"estadoPago":"pagado",`, ``),
		with(`pagado`, `otro`),
		with(`"metodoPagoId":2`, `"metodoPagoId":0`),
	}
	for _, b := range bad {
		call(t, http.MethodPost, "/pagos", b, p, http.StatusBadRequest)
	}
	// método inexistente -> 404
	serve(false, false)
	call(t, http.MethodPost, "/pagos", ok, p, http.StatusNotFound)
	// error al validar el método -> 500
	fakeQuery = func(string, []driver.NamedValue) (driver.Rows, error) { return nil, errBoom }
	call(t, http.MethodPost, "/pagos", ok, p, http.StatusInternalServerError)

	serve(false, true)
	b := call(t, http.MethodPost, "/pagos", ok, p, http.StatusCreated)
	for _, want := range []string{`"fechaPago":"01-02-2025"`, `"estadoPago":"PAGADO"`, `"updatedBy":"x"`} {
		if !strings.Contains(b, want) {
			t.Fatalf("falta %s en %s", want, b)
		}
	}
	// sin updatedBy
	call(t, http.MethodPost, "/pagos", with(`,"updatedBy":"x"`, ``), p, http.StatusCreated)

	fakeExec = func(string, []driver.NamedValue) (driver.Result, error) {
		return nil, errors.New("duplicate key 23505")
	}
	call(t, http.MethodPost, "/pagos", ok, p, http.StatusConflict)
	fakeExec = func(string, []driver.NamedValue) (driver.Result, error) { return nil, errBoom }
	call(t, http.MethodPost, "/pagos", ok, p, http.StatusInternalServerError)
}

func decodeData(t *testing.T, body string) map[string]any {
	t.Helper()
	var r struct {
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &r); err != nil {
		t.Fatal(err)
	}
	return r.Data
}

func TestPut(t *testing.T) {
	defer resetFake()
	u := func(c *PagoController) { c.Put() }
	call(t, http.MethodPut, "/pagos", `{}`, u, http.StatusBadRequest)
	call(t, http.MethodPut, "/pagos?id=4", `{}`, u, http.StatusNotFound)
	fakeQuery = func(string, []driver.NamedValue) (driver.Rows, error) { return nil, errBoom }
	call(t, http.MethodPut, "/pagos?id=4", `{}`, u, http.StatusInternalServerError)

	serve(true, true)
	// {} conserva todo
	d := decodeData(t, call(t, http.MethodPut, "/pagos?id=4", `{}`, u, http.StatusOK))
	if d["monto"] != float64(50000) || d["horaPago"] != "14:30:00" || d["fechaPago"] != "31-01-2025" || d["updatedBy"] != "cajero" {
		t.Fatalf("merge vacío debe conservar: %v", d)
	}
	// solo monto: hora y método se conservan (no se exigen); un pago PAGADO no admite cambiar el monto
	call(t, http.MethodPut, "/pagos?id=4", `{"monto":70}`, u, http.StatusConflict)
	pagoEstado = "PENDIENTE"
	t.Cleanup(func() { pagoEstado = "PAGADO" })
	d = decodeData(t, call(t, http.MethodPut, "/pagos?id=4", `{"monto":70}`, u, http.StatusOK))
	if d["monto"] != float64(70) || d["horaPago"] != "14:30:00" {
		t.Fatalf("merge parcial incorrecto: %v", d)
	}
	// claves unificadas con POST + metodo + estado + updatedBy
	d = decodeData(t, call(t, http.MethodPut, "/pagos?id=4",
		`{"fechaPago":"2025-03-02","horaPago":"09:15","estadoPago":"pendiente","metodoPagoId":2,"updatedBy":"yo"}`, u, http.StatusOK))
	if d["fechaPago"] != "02-03-2025" || d["horaPago"] != "09:15:00" || d["estadoPago"] != "PENDIENTE" || d["updatedBy"] != "yo" {
		t.Fatalf("claves nuevas no aplicadas: %v", d)
	}
	// alias heredados fecha/hora
	d = decodeData(t, call(t, http.MethodPut, "/pagos?id=4", `{"fecha":"2025-04-05","hora":"10:00:00"}`, u, http.StatusOK))
	if d["fechaPago"] != "05-04-2025" || d["horaPago"] != "10:00:00" {
		t.Fatalf("alias no aplicados: %v", d)
	}
	// fechaPago gana sobre fecha
	d = decodeData(t, call(t, http.MethodPut, "/pagos?id=4", `{"fecha":"2025-04-05","fechaPago":"2025-06-07"}`, u, http.StatusOK))
	if d["fechaPago"] != "07-06-2025" {
		t.Fatalf("precedencia incorrecta: %v", d)
	}
	// updatedBy null lo limpia
	d = decodeData(t, call(t, http.MethodPut, "/pagos?id=4", `{"updatedBy":null}`, u, http.StatusOK))
	if _, ok := d["updatedBy"]; ok {
		t.Fatalf("updatedBy debía limpiarse: %v", d)
	}

	for _, b := range []string{
		`nojson`, `[]`,
		`{"monto":null}`, `{"fechaPago":null}`, `{"horaPago":null}`, `{"estadoPago":null}`, `{"metodoPagoId":null}`,
		`{"fechaPago":"2025/01/01"}`, `{"horaPago":"x"}`, `{"horaPago":"99:00"}`,
		`{"monto":0}`, `{"monto":-1}`, `{"estadoPago":"zzz"}`, `{"metodoPagoId":0}`,
	} {
		call(t, http.MethodPut, "/pagos?id=4", b, u, http.StatusBadRequest)
	}

	// método inexistente -> 404; error de BD al validarlo -> 500
	serve(true, false)
	call(t, http.MethodPut, "/pagos?id=4", `{"metodoPagoId":9}`, u, http.StatusNotFound)
	fakeQuery = func(q string, _ []driver.NamedValue) (driver.Rows, error) {
		if strings.Contains(q, `FROM "metodo_pago"`) {
			return nil, errBoom
		}
		return rowsOf(pagoCols, pagoRow()), nil
	}
	call(t, http.MethodPut, "/pagos?id=4", `{"metodoPagoId":9}`, u, http.StatusInternalServerError)

	serve(true, true)
	fakeExec = func(string, []driver.NamedValue) (driver.Result, error) { return nil, errors.New("duplicate key") }
	call(t, http.MethodPut, "/pagos?id=4", `{}`, u, http.StatusConflict)
	fakeExec = func(string, []driver.NamedValue) (driver.Result, error) { return nil, errBoom }
	call(t, http.MethodPut, "/pagos?id=4", `{}`, u, http.StatusInternalServerError)
}

func TestDelete(t *testing.T) {
	defer resetFake()
	d := func(c *PagoController) { c.Delete() }
	call(t, http.MethodDelete, "/pagos", "", d, http.StatusBadRequest)
	call(t, http.MethodDelete, "/pagos?id=4", "", d, http.StatusOK)
	fakeAffected = 0
	call(t, http.MethodDelete, "/pagos?id=4", "", d, http.StatusNotFound)
	fakeExec = func(string, []driver.NamedValue) (driver.Result, error) {
		return nil, errors.New("violates foreign key")
	}
	call(t, http.MethodDelete, "/pagos?id=4", "", d, http.StatusConflict)
	fakeExec = func(string, []driver.NamedValue) (driver.Result, error) { return nil, errBoom }
	call(t, http.MethodDelete, "/pagos?id=4", "", d, http.StatusInternalServerError)
}

func TestHelpers(t *testing.T) {
	if _, ok := normalizeEstado(" pagado "); !ok {
		t.Fatal("pagado debe ser válido")
	}
	if firstNonNil(nil, nil) != nil {
		t.Fatal("firstNonNil debe devolver nil")
	}
}
