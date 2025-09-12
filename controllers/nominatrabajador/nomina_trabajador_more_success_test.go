package nominatrabajador

import (
	stdctx "context"
	"database/sql/driver"
	"net/http"
	"net/http/httptest"
	"testing"

	beegoCtx "github.com/beego/beego/v2/server/web/context"
)

func TestNominaTrabajadorGetByTrabajadorSuccess(t *testing.T) {
	orig := MockQuery
	MockQuery = func(_ stdctx.Context, q string, _ []driver.NamedValue) (driver.Rows, error) {
		cols := []string{
			"pk_id_nomina_trabajador",
			"sueldo_base",
			"monto_incidencias",
			"detalles",
			"pk_documento_trabajador",
			"pk_id_nomina",
		}
		vals := [][]driver.Value{{int64(1), int64(1000), int64(0), "desc", int64(1), int64(1)}}
		return &mockRows{columns: cols, values: vals}, nil
	}
	t.Cleanup(func() { MockQuery = orig })

	r := httptest.NewRequest(http.MethodGet,
		"/nomina_trabajador/search?documento=1&actual=true&pagas=true&mes=1&anio=2024", nil)
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
}
