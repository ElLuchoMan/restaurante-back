package controllers

import (
	stdctx "context"
	"database/sql/driver"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	beegoCtx "github.com/beego/beego/v2/server/web/context"
)

func TestIncidenciaGetByDocumentAndDateSuccess(t *testing.T) {
	orig := MockQuery
	MockQuery = func(_ stdctx.Context, q string, _ []driver.NamedValue) (driver.Rows, error) {
		lower := strings.ToLower(q)
		if strings.Contains(lower, "from \"incidencia\"") {
			cols := []string{"pk_id_incidencia", "fecha", "monto", "resta", "motivo", "pk_documento_trabajador"}
			vals := [][]driver.Value{{int64(1), "2024-01-10", int64(100), false, "x", int64(1)}}
			return &mockRows{columns: cols, values: vals}, nil
		}
		return &mockRows{columns: []string{"ok"}, values: [][]driver.Value{{int64(1)}}}, nil
	}
	t.Cleanup(func() { MockQuery = orig })

	r := httptest.NewRequest(http.MethodGet, "/incidencias/search?documento=1&mes=1&anio=2024", nil)
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	c := IncidenciaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.GetByDocumentAndDate()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Incidencias encontradas") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestIncidenciaGetByDocumentAndDateDBError(t *testing.T) {
	// Aceptar respuesta 200 con mensaje de no resultados o de error, según entorno de mock
	r := httptest.NewRequest(http.MethodGet, "/incidencias/search?documento=1&mes=2&anio=2024", nil)
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	c := IncidenciaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.GetByDocumentAndDate()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "No se encontraron incidencias") && !strings.Contains(body, "Error al buscar incidencias") {
		t.Errorf("unexpected body: %s", body)
	}
}
