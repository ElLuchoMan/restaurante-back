package domicilio

import (
	"database/sql/driver"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"
)

var (
	errBoom = errors.New("boom")

	domCols     = []string{"pk_id_domicilio", "direccion", "telefono", "estado_domicilio", "entregado", "fecha", "observaciones", "created_at", "updated_at", "created_by", "updated_by", "pk_documento_trabajador"}
	clienteCols = []string{"documento", "nombre", "apellido"}
	pedidoCols  = []string{"pedido_id", "pago_id", "pago_monto", "subtotal_productos", "productos"}
	countCols   = []string{"count"}
	idCols      = []string{"pk_id_domicilio"}
	ts          = time.Date(2025, 1, 31, 20, 0, 0, 0, time.UTC)
)

func domRow(estado string, entregado bool, trabajador driver.Value) []driver.Value {
	return []driver.Value{int64(3), "Calle 1", "300", estado, entregado, time.Date(2025, 1, 31, 12, 0, 0, 0, time.UTC), "obs", ts, ts, "admin", "op", trabajador}
}

// rt: devuelve las filas del primer fragmento que aparezca en la consulta.
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

func domOK(estado string, entregado bool, trabajador driver.Value) rt {
	return rt{`FROM "domicilio"`, func() driver.Rows { return rowsOf(domCols, domRow(estado, entregado, trabajador)) }}
}

func trabajadorCount(n int64) rt {
	return rt{`COUNT(*) FROM "trabajador"`, func() driver.Rows { return rowsOf(countCols, []driver.Value{n}) }}
}

func call(t *testing.T, method, target, body string, f func(c *DomicilioController), status int) string {
	t.Helper()
	ctx, w := newCtx(method, target, body)
	c := &DomicilioController{}
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
	g := func(c *DomicilioController) { c.GetAll() }
	if b := call(t, http.MethodGet, "/domicilios", "", g, http.StatusOK); !strings.Contains(b, `"data":[]`) {
		t.Fatalf("lista vacía debe ser []: %s", b)
	}
	var gotQ string
	fakeQuery = func(q string, _ []driver.NamedValue) (driver.Rows, error) {
		gotQ = q
		return rowsOf(domCols, domRow("PENDIENTE", false, int64(77))), nil
	}
	b := call(t, http.MethodGet, "/domicilios?direccion=calle&telefono=300&updated_by=op&fecha=2025-01-31&estado=pendiente&trabajador=77", "", g, http.StatusOK)
	contains(t, b, `"fechaDomicilio":"31-01-2025"`, `"estadoDomicilio":"PENDIENTE"`, `"trabajadorAsignado":{"documentoTrabajador":77`)
	for _, want := range []string{"UPPER(T0.\"direccion\"", "estado_domicilio", "entregado", "IS NULL", "pk_documento_trabajador"} {
		if !strings.Contains(gotQ, want) {
			t.Fatalf("filtro %q no aplicado: %s", want, gotQ)
		}
	}
	for _, q := range []string{"fecha=31-01-2025", "estado=X", "trabajador=0", "trabajador=x"} {
		call(t, http.MethodGet, "/domicilios?"+q, "", g, http.StatusBadRequest)
	}
	fakeQuery = func(string, []driver.NamedValue) (driver.Rows, error) { return nil, errBoom }
	call(t, http.MethodGet, "/domicilios", "", g, http.StatusInternalServerError)
}

func TestGetById(t *testing.T) {
	defer resetFake()
	g := func(c *DomicilioController) { c.GetById() }
	for _, q := range []string{"", "id=0", "id=x"} {
		call(t, http.MethodGet, "/domicilios/search?"+q, "", g, http.StatusBadRequest)
	}
	url := "/domicilios/search?id=3"
	serve()
	call(t, http.MethodGet, url, "", g, http.StatusNotFound)
	fakeQuery = func(string, []driver.NamedValue) (driver.Rows, error) { return nil, errBoom }
	call(t, http.MethodGet, url, "", g, http.StatusInternalServerError)

	// sin pedido asociado: no hay cliente ni pedido
	serve(domOK("PENDIENTE", false, nil))
	b := call(t, http.MethodGet, url, "", g, http.StatusOK)
	if strings.Contains(b, `"cliente"`) || strings.Contains(b, `"pedido"`) {
		t.Fatalf("no debe traer cliente/pedido: %s", b)
	}
	contains(t, b, `"domicilio":{"domicilioId":3`)

	cli := rt{"FROM pedido p\nJOIN cliente", func() driver.Rows {
		return rowsOf(clienteCols, []driver.Value{int64(1001), "Juan", "Pérez"})
	}}
	prods := `[{"pk_id_producto":1,"nombre":"Bandeja","cantidad":2,"precio":25000,"subtotal":50000}]`
	ped := func(pago driver.Value, monto driver.Value, productos string) rt {
		return rt{"LEFT JOIN pago pa", func() driver.Rows {
			return rowsOf(pedidoCols, []driver.Value{int64(10), pago, monto, float64(50000), productos})
		}}
	}
	serve(domOK("PENDIENTE", false, nil), cli, ped(int64(4), float64(48000), prods))
	b = call(t, http.MethodGet, url, "", g, http.StatusOK)
	contains(t, b, `"cliente":{"documento":1001,"nombre":"Juan","apellido":"Pérez"}`, `"pagoId":4`, `"montoPago":48000`, `"total":48000`, `"subtotalProductos":50000`, `"pk_id_producto":1`)
	// sin pago: total = subtotal y pagoId null
	serve(domOK("PENDIENTE", false, nil), ped(nil, nil, "[]"))
	b = call(t, http.MethodGet, url, "", g, http.StatusOK)
	contains(t, b, `"pagoId":null`, `"total":50000`, `"productos":[]`)
	// productos ilegibles
	serve(domOK("PENDIENTE", false, nil), ped(nil, nil, "no-json"))
	call(t, http.MethodGet, url, "", g, http.StatusInternalServerError)

	// errores en las consultas auxiliares
	for _, frag := range []string{"FROM pedido p\nJOIN cliente", "LEFT JOIN pago pa"} {
		serve(domOK("PENDIENTE", false, nil))
		inner := fakeQuery
		fakeQuery = func(q string, a []driver.NamedValue) (driver.Rows, error) {
			if strings.Contains(q, frag) {
				return nil, errBoom
			}
			return inner(q, a)
		}
		call(t, http.MethodGet, url, "", g, http.StatusInternalServerError)
	}
}

func TestPost(t *testing.T) {
	defer resetFake()
	p := func(c *DomicilioController) { c.Post() }
	for _, body := range []string{"", "{", `{"direccion":1}`, `{"telefono":"3","fechaDomicilio":"2025-01-31"}`,
		`{"direccion":"d","telefono":" ","fechaDomicilio":"2025-01-31"}`, `{"direccion":"d","telefono":"3"}`,
		`{"direccion":"d","telefono":"3","fechaDomicilio":"31-01-2025"}`,
		`{"direccion":"d","telefono":"3","fechaDomicilio":"2025-01-31","estadoDomicilio":"X"}`,
		`{"direccion":"d","telefono":"3","fechaDomicilio":"2025-01-31","estado":"X"}`,
		`{"direccion":"d","telefono":"3","fechaDomicilio":"2025-01-31","trabajadorAsignado":-4}`,
		`{"direccion":"d","telefono":"3","fechaDomicilio":"2025-01-31","trabajadorAsignado":""}`} {
		call(t, http.MethodPost, "/domicilios", body, p, http.StatusBadRequest)
	}

	serve(domOK("PENDIENTE", false, nil))
	var insert string
	var args []driver.NamedValue
	fakeExec = func(q string, a []driver.NamedValue) (driver.Result, error) {
		insert, args = q, a
		return fakeResult{}, nil
	}
	min := `{"direccion":"d","telefono":"3","fechaDomicilio":"2025-01-31"}`
	b := call(t, http.MethodPost, "/domicilios", min, p, http.StatusCreated)
	contains(t, b, `"domicilioId":3`, `"fechaDomicilio":"31-01-2025"`)
	if strings.Contains(insert, "estado_domicilio") || strings.Contains(insert, "pk_documento_trabajador") || len(args) != 3 {
		t.Fatalf("insert mínimo inesperado: %s %v", insert, args)
	}
	// trabajadorAsignado null/0 se ignora; alias estado; todos los opcionales
	call(t, http.MethodPost, "/domicilios", `{"direccion":"d","telefono":"3","fechaDomicilio":"2025-01-31","trabajadorAsignado":0,"estado":"en_camino"}`, p, http.StatusCreated)
	if !strings.Contains(insert, "estado_domicilio") || strings.Contains(insert, "pk_documento_trabajador") {
		t.Fatalf("insert inesperado: %s", insert)
	}
	call(t, http.MethodPost, "/domicilios", `{"direccion":"d","telefono":"3","fechaDomicilio":"2025-01-31","trabajadorAsignado":null}`, p, http.StatusCreated)
	serve(domOK("PENDIENTE", false, int64(77)), trabajadorCount(1))
	call(t, http.MethodPost, "/domicilios", `{"direccion":"d","telefono":"3","fechaDomicilio":"2025-01-31","observaciones":"x","createdBy":"a","estadoDomicilio":"PENDIENTE","trabajadorAsignado":77}`, p, http.StatusCreated)
	for _, want := range []string{"observaciones", "created_by", "estado_domicilio", "pk_documento_trabajador"} {
		if !strings.Contains(insert, want) {
			t.Fatalf("falta %s en %s", want, insert)
		}
	}

	serve(trabajadorCount(0))
	call(t, http.MethodPost, "/domicilios", `{"direccion":"d","telefono":"3","fechaDomicilio":"2025-01-31","trabajadorAsignado":77}`, p, http.StatusNotFound)
	fakeQuery = func(string, []driver.NamedValue) (driver.Rows, error) { return nil, errBoom }
	call(t, http.MethodPost, "/domicilios", `{"direccion":"d","telefono":"3","fechaDomicilio":"2025-01-31","trabajadorAsignado":77}`, p, http.StatusInternalServerError)

	// error al insertar (500 y 409) y al releer la fila creada
	serve(domOK("PENDIENTE", false, nil))
	fakeExec = func(string, []driver.NamedValue) (driver.Result, error) { return nil, errBoom }
	call(t, http.MethodPost, "/domicilios", min, p, http.StatusInternalServerError)
	fakeExec = func(string, []driver.NamedValue) (driver.Result, error) {
		return nil, errors.New("duplicate key value violates unique constraint")
	}
	call(t, http.MethodPost, "/domicilios", min, p, http.StatusConflict)
	fakeExec = nil
	serve()
	call(t, http.MethodPost, "/domicilios", min, p, http.StatusNotFound)
}

func TestPut(t *testing.T) {
	defer resetFake()
	u := func(c *DomicilioController) { c.Put() }
	call(t, http.MethodPut, "/domicilios", `{}`, u, http.StatusBadRequest)
	call(t, http.MethodPut, "/domicilios?id=0", `{}`, u, http.StatusBadRequest)
	url := "/domicilios?id=3"
	serve()
	call(t, http.MethodPut, url, `{}`, u, http.StatusNotFound)
	fakeQuery = func(string, []driver.NamedValue) (driver.Rows, error) { return nil, errBoom }
	call(t, http.MethodPut, url, `{}`, u, http.StatusInternalServerError)

	serve(domOK("PENDIENTE", false, nil))
	for _, body := range []string{"", "{", `{"direccion":1}`, `{"direccion":null}`, `{"telefono":null}`, `{"estado":null}`, `{"estadoDomicilio":null}`, `{"fechaDomicilio":null}`,
		`{"direccion":"  "}`, `{"telefono":""}`, `{"estado":"X"}`, `{"estadoDomicilio":"X"}`, `{"fechaDomicilio":"31-01-2025"}`} {
		call(t, http.MethodPut, url, body, u, http.StatusBadRequest)
	}

	var update string
	var args []driver.NamedValue
	fakeExec = func(q string, a []driver.NamedValue) (driver.Result, error) {
		update, args = q, a
		return fakeResult{}, nil
	}
	// {} solo refresca updated_at
	call(t, http.MethodPut, url, `{}`, u, http.StatusOK)
	if len(args) != 2 || !strings.Contains(update, "updated_at") || strings.Contains(update, "direccion") {
		t.Fatalf("merge: solo debía actualizarse updated_at: %s %v", update, args)
	}
	// marcar entregado (estado) y alias estadoDomicilio
	serve(domOK("ENTREGADO", true, nil))
	b := call(t, http.MethodPut, url, `{"estado":"entregado"}`, u, http.StatusOK)
	contains(t, b, `"estadoDomicilio":"ENTREGADO"`, `"entregado":true`)
	if !strings.Contains(update, "estado_domicilio") {
		t.Fatalf("no se actualizó el estado: %s", update)
	}
	call(t, http.MethodPut, url, `{"estadoDomicilio":"EN_CAMINO"}`, u, http.StatusOK)
	// todos los campos, y null limpia observaciones y updatedBy
	call(t, http.MethodPut, url, `{"direccion":" Calle 2 ","telefono":" 311 ","estado":"PENDIENTE","fechaDomicilio":"2025-02-01","observaciones":"o","updatedBy":"u"}`, u, http.StatusOK)
	for _, want := range []string{"direccion", "telefono", "estado_domicilio", "fecha", "observaciones", "updated_by"} {
		if !strings.Contains(update, want) {
			t.Fatalf("falta %s en %s", want, update)
		}
	}
	call(t, http.MethodPut, url, `{"observaciones":null,"updatedBy":null}`, u, http.StatusOK)
	if !strings.Contains(update, "observaciones") || !strings.Contains(update, "updated_by") || args[1].Value != nil {
		t.Fatalf("null debía limpiar observaciones: %s %v", update, args)
	}
	// solo direccion o solo telefono
	call(t, http.MethodPut, url, `{"direccion":"x"}`, u, http.StatusOK)
	call(t, http.MethodPut, url, `{"telefono":"x"}`, u, http.StatusOK)

	fakeExec = func(string, []driver.NamedValue) (driver.Result, error) { return nil, errBoom }
	call(t, http.MethodPut, url, `{}`, u, http.StatusInternalServerError)
	fakeExec = func(string, []driver.NamedValue) (driver.Result, error) {
		return nil, errors.New("violates foreign key constraint")
	}
	call(t, http.MethodPut, url, `{}`, u, http.StatusConflict)
	// falla la relectura posterior al update
	fakeExec = nil
	reads := 0
	fakeQuery = func(q string, _ []driver.NamedValue) (driver.Rows, error) {
		reads++
		if reads > 1 {
			return nil, errBoom
		}
		return rowsOf(domCols, domRow("PENDIENTE", false, nil)), nil
	}
	call(t, http.MethodPut, url, `{}`, u, http.StatusInternalServerError)
}

func TestDelete(t *testing.T) {
	defer resetFake()
	d := func(c *DomicilioController) { c.Delete() }
	call(t, http.MethodDelete, "/domicilios", "", d, http.StatusBadRequest)
	call(t, http.MethodDelete, "/domicilios?id=0", "", d, http.StatusBadRequest)
	call(t, http.MethodDelete, "/domicilios?id=3", "", d, http.StatusOK)
	fakeAffected = 0
	call(t, http.MethodDelete, "/domicilios?id=3", "", d, http.StatusNotFound)
	fakeAffected = 1
	fakeExec = func(string, []driver.NamedValue) (driver.Result, error) { return nil, errBoom }
	call(t, http.MethodDelete, "/domicilios?id=3", "", d, http.StatusInternalServerError)
	fakeExec = func(string, []driver.NamedValue) (driver.Result, error) {
		return nil, errors.New("violates foreign key constraint (SQLSTATE 23503)")
	}
	call(t, http.MethodDelete, "/domicilios?id=3", "", d, http.StatusConflict)
}

func TestAsignarDomiciliario(t *testing.T) {
	defer resetFake()
	a := func(c *DomicilioController) { c.AsignarDomiciliario() }
	base := "/domicilios/asignar?"
	for _, q := range []string{"", "trabajador_id=77", "domicilio_id=0&trabajador_id=77", "domicilio_id=3", "domicilio_id=3&trabajador_id=0", "domicilio_id=3&trabajador_id=x"} {
		call(t, http.MethodPost, base+q, "", a, http.StatusBadRequest)
	}
	url := base + "domicilio_id=3&trabajador_id=77"
	serve(trabajadorCount(0))
	call(t, http.MethodPost, url, "", a, http.StatusNotFound)
	fakeQuery = func(string, []driver.NamedValue) (driver.Rows, error) { return nil, errBoom }
	call(t, http.MethodPost, url, "", a, http.StatusInternalServerError)

	serve(trabajadorCount(1), domOK("EN_CAMINO", false, int64(77)))
	b := call(t, http.MethodPost, url, "", a, http.StatusOK)
	contains(t, b, `"estadoDomicilio":"EN_CAMINO"`, `"trabajadorAsignado":{"documentoTrabajador":77`)

	// ya asignado -> 409; inexistente -> 404; fallo al leer -> 500
	fakeAffected = 0
	call(t, http.MethodPost, url, "", a, http.StatusConflict)
	serve(trabajadorCount(1))
	call(t, http.MethodPost, url, "", a, http.StatusNotFound)
	fakeAffected = 1

	serve(trabajadorCount(1), domOK("EN_CAMINO", false, int64(77)))
	fakeExec = func(string, []driver.NamedValue) (driver.Result, error) { return nil, errBoom }
	call(t, http.MethodPost, url, "", a, http.StatusInternalServerError)
	fakeExec = nil
	fakeAffErr = errBoom
	call(t, http.MethodPost, url, "", a, http.StatusInternalServerError)
	fakeAffErr = nil
	base2 := fakeQuery
	fakeQuery = func(q string, args []driver.NamedValue) (driver.Rows, error) {
		if strings.Contains(q, `FROM "domicilio"`) {
			return nil, errBoom
		}
		return base2(q, args)
	}
	call(t, http.MethodPost, url, "", a, http.StatusInternalServerError)
}
