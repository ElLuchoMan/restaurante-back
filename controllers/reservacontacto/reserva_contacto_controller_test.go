package reservacontacto

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
)

type fakeResContactoQS struct {
	orm.QuerySeter
	allErr error
	oneErr error
	data   []models.ReservaContacto
}

func (f fakeResContactoQS) Filter(string, ...interface{}) orm.QuerySeter { return f }
func (f fakeResContactoQS) All(out interface{}, cols ...string) (int64, error) {
	if f.allErr != nil {
		return 0, f.allErr
	}
	if list, ok := out.(*[]models.ReservaContacto); ok {
		*list = f.data
		return int64(len(f.data)), nil
	}
	return 0, nil
}
func (f fakeResContactoQS) One(out interface{}, cols ...string) error {
	if f.oneErr != nil {
		return f.oneErr
	}
	if row, ok := out.(*models.ReservaContacto); ok && len(f.data) > 0 {
		*row = f.data[0]
	}
	return nil
}

type fakeResContactoOrm struct{ qs orm.QuerySeter }

func (f fakeResContactoOrm) QueryTable(interface{}) orm.QuerySeter { return f.qs }

func TestResContactoOrmNew_Default(t *testing.T) {
	if o := resContactoOrmNew(); o == nil {
		t.Fatal("expected a valid orm instance")
	}
}

func TestReservaContacto_GetAll_Success(t *testing.T) {
	orig := resContactoOrmNew
	resContactoOrmNew = func() resContactoOrmer {
		return fakeResContactoOrm{qs: fakeResContactoQS{data: []models.ReservaContacto{{PKIDContacto: 1}}}}
	}
	t.Cleanup(func() { resContactoOrmNew = orig })

	r := httptest.NewRequest(http.MethodGet, "/reserva_contacto?documento_contacto=1&documento_cliente=2", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := &ReservaContactoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.GetAll()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestReservaContacto_GetAll_Error(t *testing.T) {
	orig := resContactoOrmNew
	resContactoOrmNew = func() resContactoOrmer {
		return fakeResContactoOrm{qs: fakeResContactoQS{allErr: errors.New("db")}}
	}
	t.Cleanup(func() { resContactoOrmNew = orig })

	r := httptest.NewRequest(http.MethodGet, "/reserva_contacto", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := &ReservaContactoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.GetAll()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestReservaContacto_GetAll_InvalidParams(t *testing.T) {
	orig := resContactoOrmNew
	resContactoOrmNew = func() resContactoOrmer {
		return fakeResContactoOrm{qs: fakeResContactoQS{data: []models.ReservaContacto{{PKIDContacto: 1}}}}
	}
	t.Cleanup(func() { resContactoOrmNew = orig })

	r := httptest.NewRequest(http.MethodGet, "/reserva_contacto?documento_contacto=abc&documento_cliente=xyz", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := &ReservaContactoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.GetAll()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestReservaContacto_GetById_NotFound(t *testing.T) {
	orig := resContactoOrmNew
	resContactoOrmNew = func() resContactoOrmer {
		return fakeResContactoOrm{qs: fakeResContactoQS{oneErr: errors.New("nf")}}
	}
	t.Cleanup(func() { resContactoOrmNew = orig })

	r := httptest.NewRequest(http.MethodGet, "/reserva_contacto/search?id=1", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := &ReservaContactoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.GetById()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestReservaContacto_GetById_NoRows(t *testing.T) {
	orig := resContactoOrmNew
	resContactoOrmNew = func() resContactoOrmer {
		return fakeResContactoOrm{qs: fakeResContactoQS{oneErr: orm.ErrNoRows}}
	}
	t.Cleanup(func() { resContactoOrmNew = orig })

	r := httptest.NewRequest(http.MethodGet, "/reserva_contacto/search?id=1", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := &ReservaContactoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.GetById()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestReservaContacto_GetById_Success(t *testing.T) {
	orig := resContactoOrmNew
	resContactoOrmNew = func() resContactoOrmer {
		return fakeResContactoOrm{qs: fakeResContactoQS{data: []models.ReservaContacto{{PKIDContacto: 1}}}}
	}
	t.Cleanup(func() { resContactoOrmNew = orig })

	r := httptest.NewRequest(http.MethodGet, "/reserva_contacto/search?id=1", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := &ReservaContactoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.GetById()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestReservaContacto_GetById_InvalidID(t *testing.T) {
	orig := resContactoOrmNew
	resContactoOrmNew = func() resContactoOrmer {
		return fakeResContactoOrm{qs: fakeResContactoQS{oneErr: orm.ErrNoRows}}
	}
	t.Cleanup(func() { resContactoOrmNew = orig })

	r := httptest.NewRequest(http.MethodGet, "/reserva_contacto/search?id=abc", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := &ReservaContactoController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.GetById()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
