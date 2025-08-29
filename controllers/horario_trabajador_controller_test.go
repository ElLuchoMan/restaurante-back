package controllers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
	"restaurante/models"
)

type fakeHorarioQuery struct {
	orm.QuerySeter
	all func(interface{}, ...string) (int64, error)
	one func(interface{}, ...string) error
	del func() (int64, error)
}

func (f fakeHorarioQuery) Filter(string, ...interface{}) orm.QuerySeter { return f }
func (f fakeHorarioQuery) All(res interface{}, cols ...string) (int64, error) {
	if f.all != nil {
		return f.all(res, cols...)
	}
	return 0, nil
}
func (f fakeHorarioQuery) One(res interface{}, cols ...string) error {
	if f.one != nil {
		return f.one(res, cols...)
	}
	return nil
}
func (f fakeHorarioQuery) Delete() (int64, error) {
	if f.del != nil {
		return f.del()
	}
	return 1, nil
}

type fakeHorarioOrm struct {
	query  func(interface{}) orm.QuerySeter
	insert func(interface{}) (int64, error)
	update func(interface{}, ...string) (int64, error)
}

func (f fakeHorarioOrm) QueryTable(i interface{}) orm.QuerySeter { return f.query(i) }
func (f fakeHorarioOrm) Insert(m interface{}) (int64, error) {
	if f.insert != nil {
		return f.insert(m)
	}
	return 1, nil
}
func (f fakeHorarioOrm) Update(m interface{}, cols ...string) (int64, error) {
	if f.update != nil {
		return f.update(m, cols...)
	}
	return 1, nil
}

func buildHorarioContext(method, url, body string) (*HorarioTrabajadorController, *httptest.ResponseRecorder) {
	r := httptest.NewRequest(method, url, strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := &HorarioTrabajadorController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	return c, w
}

func TestHorarioTrabajadorGetAllDBError(t *testing.T) {
	original := horarioTrabajadorNewOrm
	horarioTrabajadorNewOrm = func() horarioOrmer {
		return fakeHorarioOrm{query: func(interface{}) orm.QuerySeter {
			return fakeHorarioQuery{all: func(interface{}, ...string) (int64, error) {
				return 0, errors.New("db")
			}}
		}}
	}
	defer func() { horarioTrabajadorNewOrm = original }()

	c, w := buildHorarioContext(http.MethodGet, "/horario_trabajador?documento=1", "")
	c.GetAll()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("got %d", w.Code)
	}
}

func TestHorarioTrabajadorGetAllSuccess(t *testing.T) {
	original := horarioTrabajadorNewOrm
	horarioTrabajadorNewOrm = func() horarioOrmer {
		return fakeHorarioOrm{query: func(interface{}) orm.QuerySeter {
			return fakeHorarioQuery{all: func(res interface{}, cols ...string) (int64, error) {
				out := res.(*[]models.HorarioTrabajador)
				*out = []models.HorarioTrabajador{{PK_DOCUMENTO_TRABAJADOR: 1}}
				return 1, nil
			}}
		}}
	}
	defer func() { horarioTrabajadorNewOrm = original }()

	c, w := buildHorarioContext(http.MethodGet, "/horario_trabajador", "")
	c.GetAll()
	if w.Code != http.StatusOK {
		t.Fatalf("got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Horarios obtenidos") {
		t.Fatalf("unexpected %s", w.Body.String())
	}
}

func TestHorarioTrabajadorPostInvalidJSON(t *testing.T) {
	c, w := buildHorarioContext(http.MethodPost, "/horario_trabajador", "notjson")
	c.Post()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("got %d", w.Code)
	}
}

func TestHorarioTrabajadorPostInvalidTime(t *testing.T) {
	body := `{"documentoTrabajador":1,"dia":"Lunes","horaInicio":"bad","horaFin":"10:00:00"}`
	c, w := buildHorarioContext(http.MethodPost, "/horario_trabajador", body)
	c.Post()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("got %d", w.Code)
	}
}

func TestHorarioTrabajadorPostDBError(t *testing.T) {
	original := horarioTrabajadorNewOrm
	horarioTrabajadorNewOrm = func() horarioOrmer {
		return fakeHorarioOrm{insert: func(interface{}) (int64, error) { return 0, errors.New("db") }}
	}
	defer func() { horarioTrabajadorNewOrm = original }()
	body := `{"documentoTrabajador":1,"dia":"Lunes","horaInicio":"09:00:00","horaFin":"10:00:00"}`
	c, w := buildHorarioContext(http.MethodPost, "/horario_trabajador", body)
	c.Post()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("got %d", w.Code)
	}
}

func TestHorarioTrabajadorPostSuccess(t *testing.T) {
	original := horarioTrabajadorNewOrm
	horarioTrabajadorNewOrm = func() horarioOrmer { return fakeHorarioOrm{insert: func(interface{}) (int64, error) { return 1, nil }} }
	defer func() { horarioTrabajadorNewOrm = original }()
	body := `{"documentoTrabajador":1,"dia":"Lunes","horaInicio":"09:00:00","horaFin":"10:00:00"}`
	c, w := buildHorarioContext(http.MethodPost, "/horario_trabajador", body)
	c.Post()
	if w.Code != http.StatusCreated {
		t.Fatalf("got %d", w.Code)
	}
}

func TestHorarioTrabajadorPutInvalidParams(t *testing.T) {
	c, w := buildHorarioContext(http.MethodPut, "/horario_trabajador", "{}")
	c.Put()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("got %d", w.Code)
	}
}

func TestHorarioTrabajadorPutNotFound(t *testing.T) {
	original := horarioTrabajadorNewOrm
	horarioTrabajadorNewOrm = func() horarioOrmer {
		return fakeHorarioOrm{query: func(interface{}) orm.QuerySeter {
			return fakeHorarioQuery{one: func(interface{}, ...string) error { return orm.ErrNoRows }}
		}}
	}
	defer func() { horarioTrabajadorNewOrm = original }()

	c, w := buildHorarioContext(http.MethodPut, "/horario_trabajador?documento=1&dia=Lunes", "{}")
	c.Put()
	if w.Code != http.StatusOK {
		t.Fatalf("got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Horario no encontrado") {
		t.Fatalf("unexpected %s", w.Body.String())
	}
}

func TestHorarioTrabajadorPutInvalidJSON(t *testing.T) {
	original := horarioTrabajadorNewOrm
	horarioTrabajadorNewOrm = func() horarioOrmer {
		return fakeHorarioOrm{query: func(interface{}) orm.QuerySeter {
			return fakeHorarioQuery{one: func(interface{}, ...string) error { return nil }}
		}}
	}
	defer func() { horarioTrabajadorNewOrm = original }()

	c, w := buildHorarioContext(http.MethodPut, "/horario_trabajador?documento=1&dia=Lunes", "notjson")
	c.Put()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("got %d", w.Code)
	}
}

func TestHorarioTrabajadorPutInvalidTime(t *testing.T) {
	original := horarioTrabajadorNewOrm
	horarioTrabajadorNewOrm = func() horarioOrmer {
		return fakeHorarioOrm{query: func(interface{}) orm.QuerySeter {
			return fakeHorarioQuery{one: func(interface{}, ...string) error { return nil }}
		}}
	}
	defer func() { horarioTrabajadorNewOrm = original }()
	body := `{"horaInicio":"bad"}`
	c, w := buildHorarioContext(http.MethodPut, "/horario_trabajador?documento=1&dia=Lunes", body)
	c.Put()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("got %d", w.Code)
	}
}

func TestHorarioTrabajadorPutUpdateError(t *testing.T) {
	original := horarioTrabajadorNewOrm
	horarioTrabajadorNewOrm = func() horarioOrmer {
		return fakeHorarioOrm{query: func(interface{}) orm.QuerySeter {
			return fakeHorarioQuery{one: func(interface{}, ...string) error { return nil }}
		}, update: func(interface{}, ...string) (int64, error) { return 0, errors.New("db") }}
	}
	defer func() { horarioTrabajadorNewOrm = original }()
	body := `{"horaInicio":"09:00:00"}`
	c, w := buildHorarioContext(http.MethodPut, "/horario_trabajador?documento=1&dia=Lunes", body)
	c.Put()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("got %d", w.Code)
	}
}

func TestHorarioTrabajadorPutSuccess(t *testing.T) {
	original := horarioTrabajadorNewOrm
	horarioTrabajadorNewOrm = func() horarioOrmer {
		return fakeHorarioOrm{query: func(interface{}) orm.QuerySeter {
			return fakeHorarioQuery{one: func(interface{}, ...string) error { return nil }}
		}, update: func(interface{}, ...string) (int64, error) { return 1, nil }}
	}
	defer func() { horarioTrabajadorNewOrm = original }()
	body := `{"horaInicio":"09:00:00"}`
	c, w := buildHorarioContext(http.MethodPut, "/horario_trabajador?documento=1&dia=Lunes", body)
	c.Put()
	if w.Code != http.StatusOK {
		t.Fatalf("got %d", w.Code)
	}
}

func TestHorarioTrabajadorDeleteInvalidParams(t *testing.T) {
	c, w := buildHorarioContext(http.MethodDelete, "/horario_trabajador", "")
	c.Delete()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("got %d", w.Code)
	}
}

func TestHorarioTrabajadorDeleteDBError(t *testing.T) {
	original := horarioTrabajadorNewOrm
	horarioTrabajadorNewOrm = func() horarioOrmer {
		return fakeHorarioOrm{query: func(interface{}) orm.QuerySeter {
			return fakeHorarioQuery{del: func() (int64, error) { return 0, errors.New("db") }}
		}}
	}
	defer func() { horarioTrabajadorNewOrm = original }()

	c, w := buildHorarioContext(http.MethodDelete, "/horario_trabajador?documento=1&dia=Lunes", "")
	c.Delete()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("got %d", w.Code)
	}
}

func TestHorarioTrabajadorDeleteSuccess(t *testing.T) {
	original := horarioTrabajadorNewOrm
	horarioTrabajadorNewOrm = func() horarioOrmer {
		return fakeHorarioOrm{query: func(interface{}) orm.QuerySeter { return fakeHorarioQuery{} }}
	}
	defer func() { horarioTrabajadorNewOrm = original }()

	c, w := buildHorarioContext(http.MethodDelete, "/horario_trabajador?documento=1&dia=Lunes", "")
	c.Delete()
	if w.Code != http.StatusOK {
		t.Fatalf("got %d", w.Code)
	}
}
