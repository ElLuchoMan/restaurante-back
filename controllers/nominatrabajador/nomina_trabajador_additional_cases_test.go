package nominatrabajador

import (
	stdctx "context"
	"database/sql/driver"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beego/beego/v2/client/orm"
	beegoCtx "github.com/beego/beego/v2/server/web/context"
)

func TestNominaTrabajadorGetByTrabajadorInvalidParams(t *testing.T) {
	cases := []string{
		"/nomina_trabajador/search?documento=1&actual=x",
		"/nomina_trabajador/search?documento=1&pagas=x",
		"/nomina_trabajador/search?documento=1&no_pagas=x",
		"/nomina_trabajador/search?documento=1&mes=x",
		"/nomina_trabajador/search?documento=1&anio=x",
		"/nomina_trabajador/search?documento=0",
	}
	for _, u := range cases {
		r := httptest.NewRequest(http.MethodGet, u, nil)
		w := httptest.NewRecorder()
		ctx := beegoCtx.NewContext()
		ctx.Reset(w, r)
		c := NominaTrabajadorController{}
		c.Ctx = ctx
		c.Data = make(map[interface{}]interface{})
		c.GetByTrabajador()
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for %s, got %d", u, w.Code)
		}
	}
}

func TestNominaTrabajadorGetNominasByMesParseError(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/nomina_trabajador/mes?mes=bad&anio=2024", nil)
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	c := NominaTrabajadorController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.GetNominasByMes()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestNominaTrabajadorGetByTrabajadorDBError(t *testing.T) {
	orig := MockQuery
	MockQuery = func(_ stdctx.Context, _ string, _ []driver.NamedValue) (driver.Rows, error) {
		return nil, fmt.Errorf("db error")
	}
	t.Cleanup(func() { MockQuery = orig })

	r := httptest.NewRequest(http.MethodGet, "/nomina_trabajador/search?documento=1", nil)
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	c := NominaTrabajadorController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.GetByTrabajador()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestNominaTrabajadorGetByTrabajadorErrNoRows(t *testing.T) {
	orig := MockQuery
	MockQuery = func(_ stdctx.Context, _ string, _ []driver.NamedValue) (driver.Rows, error) {
		return nil, orm.ErrNoRows
	}
	t.Cleanup(func() { MockQuery = orig })

	r := httptest.NewRequest(http.MethodGet, "/nomina_trabajador/search?documento=1", nil)
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	c := NominaTrabajadorController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.GetByTrabajador()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "No se encontraron") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}
