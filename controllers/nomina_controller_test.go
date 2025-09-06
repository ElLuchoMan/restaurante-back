package controllers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
)

func TestNominaGetAllWithoutDB(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/nominas", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := NominaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetAll()

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al obtener nóminas") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestNominaPostInvalidJSON(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/nominas", strings.NewReader("notjson"))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := NominaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestNominaPostDBError(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/nominas", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte("{\"estadoNomina\":\"OTRO\"}")
	c := NominaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al crear la nómina") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestNominaPutInvalidID(t *testing.T) {
	r := httptest.NewRequest(http.MethodPut, "/nominas", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := NominaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Put()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestNominaPutUpdateError(t *testing.T) {
	r := httptest.NewRequest(http.MethodPut, "/nominas?id=1", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := NominaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Put()

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al actualizar el estado de la nómina") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestNominaDeleteInvalidID(t *testing.T) {
	r := httptest.NewRequest(http.MethodDelete, "/nominas", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := NominaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Delete()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestNominaDeleteNotFound(t *testing.T) {
	r := httptest.NewRequest(http.MethodDelete, "/nominas?id=1", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := NominaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Delete()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Nómina no encontrada") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestEstadosNominaPermitidos(t *testing.T) {
	if !estadosNominaPermitidos[models.EstadoNominaPago] || !estadosNominaPermitidos[models.EstadoNominaNoPago] {
		t.Fatalf("expected valid states to be allowed")
	}
	if estadosNominaPermitidos[models.EstadoNomina("otro")] {
		t.Fatalf("unexpected state should not be allowed")
	}
}

// Nuevos tests con inyección
func TestNominaGetAllWithFilters(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/nominas?fecha=2024-01-01&mes=1&anio=2024", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := NominaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	saved := queryAllNominas
	queryAllNominas = func(o orm.Ormer, out *[]models.Nomina) (int64, error) {
		*out = []models.Nomina{{FECHA: parseDate("2024-01-01").FECHA}}
		return 1, nil
	}
	t.Cleanup(func(){ queryAllNominas = saved })

	c.GetAll()
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func parseDate(s string) (n models.Nomina) { n.FECHA, _ = time.Parse("2006-01-02", s); return }

func TestNominaPutAlreadyPaid(t *testing.T) {
	r := httptest.NewRequest(http.MethodPut, "/nominas?id=1", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := NominaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	origRead := readNominaFn
	readNominaFn = func(o orm.Ormer, n *models.Nomina) error { n.ESTADO_NOMINA = models.EstadoNominaPago; return nil }
	t.Cleanup(func(){ readNominaFn = origRead })

	c.Put()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestNominaPutSuccess(t *testing.T) {
	r := httptest.NewRequest(http.MethodPut, "/nominas?id=1", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := NominaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	origRead := readNominaFn
	origUpdate := updateNominaFn
	readNominaFn = func(o orm.Ormer, n *models.Nomina) error { n.ESTADO_NOMINA = models.EstadoNominaNoPago; return nil }
	updateNominaFn = func(o orm.Ormer, n *models.Nomina, cols ...string) (int64, error) { return 1, nil }
	t.Cleanup(func(){ readNominaFn = origRead; updateNominaFn = origUpdate })

	c.Put()
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestNominaDeleteSuccess(t *testing.T) {
	r := httptest.NewRequest(http.MethodDelete, "/nominas?id=1", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := NominaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	origRead := readNominaFn
	origUpdate := updateNominaFn
	readNominaFn = func(o orm.Ormer, n *models.Nomina) error { n.ESTADO_NOMINA = models.EstadoNominaPago; return nil }
	updateNominaFn = func(o orm.Ormer, n *models.Nomina, cols ...string) (int64, error) { return 1, nil }
	t.Cleanup(func(){ readNominaFn = origRead; updateNominaFn = origUpdate })

	c.Delete()
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}
