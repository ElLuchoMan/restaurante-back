package precioproductohist

import (
	"database/sql/driver"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"
)

var cols = []string{"pk_id_precio_hist", "pk_id_producto", "nombre", "estado_producto", "precio", "fecha_vigencia"}

func found(queries *[]string, args *[][]driver.NamedValue) {
	fakeQuery = func(q string, a []driver.NamedValue) (driver.Rows, error) {
		if queries != nil {
			*queries = append(*queries, q)
			*args = append(*args, a)
		}
		return rowsOf(cols, []driver.Value{int64(9), int64(4), "Bandeja", "DISPONIBLE", int64(25000), time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)}), nil
	}
}

func boom() {
	fakeQuery = func(string, []driver.NamedValue) (driver.Rows, error) { return nil, errors.New("boom") }
}

func call(t *testing.T, target string, f func(c *PrecioProductoHistController), status int) string {
	t.Helper()
	ctx, w := newCtx(http.MethodGet, target, "")
	c := &PrecioProductoHistController{}
	c.Ctx, c.Data = ctx, map[interface{}]interface{}{}
	f(c)
	expect(t, w, status)
	return w.Body.String()
}

var (
	getAll  = func(c *PrecioProductoHistController) { c.GetAll() }
	getByID = func(c *PrecioProductoHistController) { c.GetById() }
)

func TestGetAll(t *testing.T) {
	defer resetFake()
	if b := call(t, "/precio_producto_hist", getAll, http.StatusOK); !strings.Contains(b, `"data":[]`) {
		t.Fatalf("lista vacía debe ser []: %s", b)
	}
	var qs []string
	var as [][]driver.NamedValue
	found(&qs, &as)
	b := call(t, "/precio_producto_hist?producto_id=4&fecha=2025-01-31", getAll, http.StatusOK)
	for _, want := range []string{`"precioHistId":9`, `"productoId":4`, `"nombre":"Bandeja"`, `"estadoProducto":"DISPONIBLE"`, `"precio":25000`, `"fechaVigencia":"31-01-2025"`} {
		if !strings.Contains(b, want) {
			t.Fatalf("falta %s en %s", want, b)
		}
	}
	if len(qs) != 1 || !strings.Contains(qs[0], "pph.pk_id_producto = ") || !strings.Contains(qs[0], "pph.fecha_vigencia = ") || len(as[0]) != 2 {
		t.Fatalf("filtros no aplicados: %v %v", qs, as)
	}
	for _, q := range []string{"?producto_id=abc", "?producto_id=0", "?fecha=31-01-2025", "?fecha=nada"} {
		call(t, "/precio_producto_hist"+q, getAll, http.StatusBadRequest)
	}
	boom()
	call(t, "/precio_producto_hist", getAll, http.StatusInternalServerError)
}

func TestGetById(t *testing.T) {
	defer resetFake()
	for _, q := range []string{"", "?id=abc", "?id=0"} {
		call(t, "/precio_producto_hist/search"+q, getByID, http.StatusBadRequest)
	}
	call(t, "/precio_producto_hist/search?id=9", getByID, http.StatusNotFound)
	found(nil, nil)
	b := call(t, "/precio_producto_hist/search?id=9", getByID, http.StatusOK)
	if !strings.Contains(b, `"precioHistId":9`) || !strings.Contains(b, `"productoId":4`) {
		t.Fatalf("cuerpo inesperado %s", b)
	}
	boom()
	call(t, "/precio_producto_hist/search?id=9", getByID, http.StatusInternalServerError)
}
