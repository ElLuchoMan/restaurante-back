package controllers

import (
	stdctx "context"
	"database/sql/driver"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
)

func setupHTCtx(method, url, body string) (*HorarioTrabajadorController, *httptest.ResponseRecorder) {
	r := httptest.NewRequest(method, url, strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	if body != "" {
		ctx.Input.RequestBody = []byte(body)
	}
	c := &HorarioTrabajadorController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	return c, w
}

func TestHorarioTrabajadorPostInvalidDia(t *testing.T) {
	body := `{"documentoTrabajador":1,"dia":"noday","horaInicio":"08:00:00","horaFin":"17:00:00"}`
	c, w := setupHTCtx(http.MethodPost, "/horario_trabajador", body)
	c.Post()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Día inválido") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestHorarioTrabajadorPostInvalidHours(t *testing.T) {
	body := `{"documentoTrabajador":1,"dia":"LUNES","horaInicio":"17:00:00","horaFin":"08:00:00"}`
	c, w := setupHTCtx(http.MethodPost, "/horario_trabajador", body)
	c.Post()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "horaFin debe ser mayor que horaInicio") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestHorarioTrabajadorPostSuccess(t *testing.T) {
	MockExec = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Result, error) {
		return mockResult{}, nil
	}
	defer func() { MockExec = nil }()

	body := `{"documentoTrabajador":1,"dia":"LUNES","horaInicio":"08:00:00","horaFin":"17:00:00"}`
	c, w := setupHTCtx(http.MethodPost, "/horario_trabajador", body)
	c.Post()
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Horario creado correctamente") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestHorarioTrabajadorPutInvalidDia(t *testing.T) {
	body := `{"horaInicio":"09:00:00"}`
	c, w := setupHTCtx(http.MethodPut, "/horario_trabajador?documento=1&dia=foo", body)
	c.Put()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Parámetro 'dia' inválido") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestHorarioTrabajadorPutSuccess(t *testing.T) {
	MockQuery = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
		cols := []string{"pk_documento_trabajador", "dia", "hora_inicio", "hora_fin"}
		vals := [][]driver.Value{{int64(1), "LUNES", time.Date(2024, 1, 1, 8, 0, 0, 0, time.UTC), time.Date(2024, 1, 1, 17, 0, 0, 0, time.UTC)}}
		return &mockRows{columns: cols, values: vals}, nil
	}
	MockExec = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Result, error) {
		return mockResult{}, nil
	}
	defer func() { MockExec = nil; MockQuery = nil }()

	body := `{"horaFin":"18:00:00"}`
	c, w := setupHTCtx(http.MethodPut, "/horario_trabajador?documento=1&dia=LUNES", body)
	c.Put()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Horario actualizado correctamente") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestIsValidDia(t *testing.T) {
	valid := []string{"LUNES", "martes", "Miércoles", "JUEVES", "viernes", "sábado", "DOMINGO"}
	for _, d := range valid {
		if !isValidDia(d) {
			t.Errorf("expected %s to be valid", d)
		}
	}
	invalid := []string{"", "FUNDAY", "luness"}
	for _, d := range invalid {
		if isValidDia(d) {
			t.Errorf("expected %s to be invalid", d)
		}
	}
}

func TestDiaToDB(t *testing.T) {
	cases := map[string]string{
		"LUNES": "Lunes",
		"lunes": "Lunes",
		"l":     "L",
		"":      "",
	}
	for in, exp := range cases {
		if got := diaToDB(in); got != exp {
			t.Errorf("diaToDB(%s) = %s, want %s", in, got, exp)
		}
	}
}

func TestHorarioTrabajadorGetAllSuccess(t *testing.T) {
	var capturedQuery string
	var capturedArgs []driver.NamedValue
	MockQuery = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
		capturedQuery = query
		capturedArgs = args
		cols := []string{"pk_documento_trabajador", "dia", "hora_inicio", "hora_fin"}
		vals := [][]driver.Value{{int64(1), "Lunes", time.Date(2024, 1, 1, 8, 0, 0, 0, time.UTC), time.Date(2024, 1, 1, 17, 0, 0, 0, time.UTC)}}
		return &mockRows{columns: cols, values: vals}, nil
	}
	defer func() { MockQuery = nil }()

	c, w := setupHTCtx(http.MethodGet, "/horario_trabajador?documento=1&dia=LUNES", "")
	c.GetAll()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Horarios obtenidos correctamente") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
	if !strings.Contains(capturedQuery, "WHERE pk_documento_trabajador") || !strings.Contains(capturedQuery, "AND dia") {
		t.Errorf("unexpected query: %s", capturedQuery)
	}
	if len(capturedArgs) != 2 || capturedArgs[0].Value.(int64) != 1 || capturedArgs[1].Value.(string) != "Lunes" {
		t.Errorf("unexpected args: %#v", capturedArgs)
	}
}

func TestHorarioTrabajadorGetAllDBError(t *testing.T) {
	MockQuery = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
		return nil, errors.New("db error")
	}
	defer func() { MockQuery = nil }()

	c, w := setupHTCtx(http.MethodGet, "/horario_trabajador", "")
	c.GetAll()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al obtener horarios") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestHorarioTrabajadorDeleteInvalidDia(t *testing.T) {
	c, w := setupHTCtx(http.MethodDelete, "/horario_trabajador?documento=1&dia=foo", "")
	c.Delete()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Parámetro 'dia' inválido") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestHorarioTrabajadorDeleteNotFound(t *testing.T) {
	MockQuery = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
		return nil, orm.ErrNoRows
	}
	defer func() { MockQuery = nil }()

	c, w := setupHTCtx(http.MethodDelete, "/horario_trabajador?documento=1&dia=LUNES", "")
	c.Delete()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Horario no encontrado") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestHorarioTrabajadorDeleteSuccess(t *testing.T) {
	MockQuery = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
		cols := []string{"pk_documento_trabajador", "dia"}
		vals := [][]driver.Value{{int64(1), "Lunes"}}
		return &mockRows{columns: cols, values: vals}, nil
	}
	MockExec = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Result, error) {
		return mockResult{}, nil
	}
	defer func() { MockQuery = nil; MockExec = nil }()

	c, w := setupHTCtx(http.MethodDelete, "/horario_trabajador?documento=1&dia=LUNES", "")
	c.Delete()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Horario eliminado correctamente") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}
