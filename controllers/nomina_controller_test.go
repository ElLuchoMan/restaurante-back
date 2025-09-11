package controllers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
	"restaurante/models"
)

func setup() func() {
	origNewOrm := nominaNewOrm
	origQueryAll := nominaQueryAll
	origInsert := nominaInsert
	origQueryByID := nominaQueryByID
	origRead := nominaRead
	origUpdate := nominaUpdate
	return func() {
		nominaNewOrm = origNewOrm
		nominaQueryAll = origQueryAll
		nominaInsert = origInsert
		nominaQueryByID = origQueryByID
		nominaRead = origRead
		nominaUpdate = origUpdate
	}
}

func TestNominaGetAllWithoutDB(t *testing.T) {
	teardown := setup()
	defer teardown()
	nominaNewOrm = func() orm.Ormer { return nil }
	nominaQueryAll = func(o orm.Ormer) ([]models.Nomina, error) {
		return nil, errors.New("db error")
	}

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
	teardown := setup()
	defer teardown()
	nominaNewOrm = func() orm.Ormer { return nil }

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
	teardown := setup()
	defer teardown()
	nominaNewOrm = func() orm.Ormer { return nil }
	nominaInsert = func(o orm.Ormer, n *models.Nomina) error {
		return errors.New("insert error")
	}

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

func TestNominaGetAllNoResults(t *testing.T) {
	teardown := setup()
	defer teardown()
	nominaNewOrm = func() orm.Ormer { return nil }
	nominaQueryAll = func(o orm.Ormer) ([]models.Nomina, error) {
		return []models.Nomina{}, nil
	}

	r := httptest.NewRequest(http.MethodGet, "/nominas", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := NominaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetAll()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "No se encontraron nóminas") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestNominaGetAllSuccess(t *testing.T) {
	teardown := setup()
	defer teardown()
	nominaNewOrm = func() orm.Ormer { return nil }
	nominaQueryAll = func(o orm.Ormer) ([]models.Nomina, error) {
		return []models.Nomina{
			{FECHA: time.Date(2023, 10, 5, 0, 0, 0, 0, time.UTC)},
			{FECHA: time.Date(2023, 11, 5, 0, 0, 0, 0, time.UTC)},
			{FECHA: time.Date(2024, 10, 5, 0, 0, 0, 0, time.UTC)},
		}, nil
	}

	r := httptest.NewRequest(http.MethodGet, "/nominas?mes=10&anio=2023", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := NominaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetAll()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Nóminas obtenidas exitosamente") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestNominaGetAllFechaFilter(t *testing.T) {
	teardown := setup()
	defer teardown()
	nominaNewOrm = func() orm.Ormer { return nil }
	nominaQueryAll = func(o orm.Ormer) ([]models.Nomina, error) {
		return []models.Nomina{
			{FECHA: time.Date(2023, 10, 5, 0, 0, 0, 0, time.UTC)},
			{FECHA: time.Date(2023, 10, 6, 0, 0, 0, 0, time.UTC)},
		}, nil
	}

	r := httptest.NewRequest(http.MethodGet, "/nominas?fecha=2023-10-05", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := NominaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.GetAll()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Nóminas obtenidas exitosamente") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestNominaPostSuccess(t *testing.T) {
	teardown := setup()
	defer teardown()
	nominaNewOrm = func() orm.Ormer { return nil }
	nominaInsert = func(o orm.Ormer, n *models.Nomina) error {
		n.PK_ID_NOMINA = 1
		return nil
	}
	nominaQueryByID = func(o orm.Ormer, id int64) (models.Nomina, error) {
		return models.Nomina{PK_ID_NOMINA: id, ESTADO_NOMINA: "NO PAGO"}, nil
	}

	r := httptest.NewRequest(http.MethodPost, "/nominas", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(`{"estadoNomina":"INVALIDO"}`)
	c := NominaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Nómina creada correctamente") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestNominaPostVerifyError(t *testing.T) {
	teardown := setup()
	defer teardown()
	nominaNewOrm = func() orm.Ormer { return nil }
	nominaInsert = func(o orm.Ormer, n *models.Nomina) error { return nil }
	nominaQueryByID = func(o orm.Ormer, id int64) (models.Nomina, error) {
		return models.Nomina{}, errors.New("query error")
	}

	r := httptest.NewRequest(http.MethodPost, "/nominas", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(`{}`)
	c := NominaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Post()

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al verificar la nómina generada") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestNominaPutInvalidID(t *testing.T) {
	teardown := setup()
	defer teardown()
	nominaNewOrm = func() orm.Ormer { return nil }

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
	teardown := setup()
	defer teardown()
	nominaNewOrm = func() orm.Ormer { return nil }
	nominaRead = func(o orm.Ormer, n *models.Nomina, cols ...string) error { return nil }
	nominaUpdate = func(o orm.Ormer, n *models.Nomina, cols ...string) (int64, error) {
		return 0, errors.New("update error")
	}

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

func TestNominaPutNotFound(t *testing.T) {
	teardown := setup()
	defer teardown()
	nominaNewOrm = func() orm.Ormer { return nil }
	nominaRead = func(o orm.Ormer, n *models.Nomina, cols ...string) error {
		return orm.ErrNoRows
	}

	r := httptest.NewRequest(http.MethodPut, "/nominas?id=1", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := NominaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Put()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Nómina no encontrada") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestNominaPutAlreadyPago(t *testing.T) {
	teardown := setup()
	defer teardown()
	nominaNewOrm = func() orm.Ormer { return nil }
	nominaRead = func(o orm.Ormer, n *models.Nomina, cols ...string) error {
		n.ESTADO_NOMINA = "PAGO"
		return nil
	}

	r := httptest.NewRequest(http.MethodPut, "/nominas?id=1", nil)
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

func TestNominaPutSuccess(t *testing.T) {
	teardown := setup()
	defer teardown()
	nominaNewOrm = func() orm.Ormer { return nil }
	nominaRead = func(o orm.Ormer, n *models.Nomina, cols ...string) error {
		n.ESTADO_NOMINA = "NO PAGO"
		return nil
	}
	nominaUpdate = func(o orm.Ormer, n *models.Nomina, cols ...string) (int64, error) {
		return 1, nil
	}

	r := httptest.NewRequest(http.MethodPut, "/nominas?id=1", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := NominaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Put()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Estado de la nómina actualizado") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestNominaDeleteInvalidID(t *testing.T) {
	teardown := setup()
	defer teardown()
	nominaNewOrm = func() orm.Ormer { return nil }

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
	teardown := setup()
	defer teardown()
	nominaNewOrm = func() orm.Ormer { return nil }
	nominaRead = func(o orm.Ormer, n *models.Nomina, cols ...string) error {
		return errors.New("not found")
	}

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

func TestNominaDeleteUpdateError(t *testing.T) {
	teardown := setup()
	defer teardown()
	nominaNewOrm = func() orm.Ormer { return nil }
	nominaRead = func(o orm.Ormer, n *models.Nomina, cols ...string) error { return nil }
	nominaUpdate = func(o orm.Ormer, n *models.Nomina, cols ...string) (int64, error) {
		return 0, errors.New("update error")
	}

	r := httptest.NewRequest(http.MethodDelete, "/nominas?id=1", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := NominaController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Delete()

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al eliminar lógicamente la nómina") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestNominaDeleteSuccess(t *testing.T) {
	teardown := setup()
	defer teardown()
	nominaNewOrm = func() orm.Ormer { return nil }
	nominaRead = func(o orm.Ormer, n *models.Nomina, cols ...string) error { return nil }
	nominaUpdate = func(o orm.Ormer, n *models.Nomina, cols ...string) (int64, error) {
		return 1, nil
	}

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
	if !strings.Contains(w.Body.String(), "Nómina eliminada lógicamente") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}
