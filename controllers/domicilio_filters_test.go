package controllers

import (
	stdctx "context"
	"database/sql/driver"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/beego/beego/v2/server/web/context"
)

func TestDomicilioGetAll_WithMultipleFilters(t *testing.T) {
    origQ := MockQuery
    MockQuery = func(_ stdctx.Context, q string, _ []driver.NamedValue) (driver.Rows, error) {
        // Devolver 1 fila válida tras aplicar filtros
        cols := []string{"pk_id_domicilio", "direccion", "telefono", "estado_domicilio", "entregado", "fecha", "observaciones", "created_at", "updated_at", "created_by", "updated_by", "pk_documento_trabajador"}
        vals := [][]driver.Value{{int64(1), "Calle X", "321", "PENDIENTE", false, time.Now(), nil, time.Now(), time.Now(), nil, nil, nil}}
        return &mockRows{columns: cols, values: vals}, nil
    }
    t.Cleanup(func(){ MockQuery = origQ })

    r := httptest.NewRequest(http.MethodGet, "/domicilios?direccion=x&telefono=321&updated_by=admin&fecha=2024-01-02&estado=PENDIENTE&trabajador=10", nil)
    w := httptest.NewRecorder()
    ctx := context.NewContext(); ctx.Reset(w, r)
    c := DomicilioController{}; c.Ctx = ctx; c.Data = make(map[interface{}]interface{})

    c.GetAll()
    if w.Code != http.StatusOK { t.Fatalf("expected 200, got %d", w.Code) }
}


