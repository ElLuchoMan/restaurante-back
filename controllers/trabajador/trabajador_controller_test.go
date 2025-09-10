package trabajador

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/beego/beego/v2/client/orm"
	context2 "github.com/beego/beego/v2/server/web/context"
	"restaurante/database"
	"restaurante/models"
)

// mock implementations

type mockQuery struct {
	orm.QuerySeter
	trabajadores []models.Trabajador
	err          error
}

func (mq *mockQuery) Filter(string, ...interface{}) orm.QuerySeter { return mq }
func (mq *mockQuery) All(result interface{}, cols ...string) (int64, error) {
	if mq.err != nil {
		return 0, mq.err
	}
	switch out := result.(type) {
	case *[]models.Trabajador:
		*out = mq.trabajadores
		return int64(len(mq.trabajadores)), nil
	case *[]models.HorarioTrabajador:
		*out = []models.HorarioTrabajador{}
		return 0, nil
	}
	return 0, nil
}

type mockOrm struct {
	orm.Ormer
	query      *mockQuery
	readErr    error
	insertErr  error
	updateErr  error
	trabajador models.Trabajador
}

func (m *mockOrm) QueryTable(interface{}) orm.QuerySeter {
	if m.query != nil {
		return m.query
	}
	return &mockQuery{}
}
func (m *mockOrm) Read(model interface{}, cols ...string) error {
	if m.readErr != nil {
		return m.readErr
	}
	if t, ok := model.(*models.Trabajador); ok {
		*t = m.trabajador
	}
	return nil
}
func (m *mockOrm) Insert(model interface{}) (int64, error) {
	if m.insertErr != nil {
		return 0, m.insertErr
	}
	if t, ok := model.(*models.Trabajador); ok {
		m.trabajador = *t
	}
	return 1, nil
}
func (m *mockOrm) Update(model interface{}, cols ...string) (int64, error) {
	if m.updateErr != nil {
		return 0, m.updateErr
	}
	if t, ok := model.(*models.Trabajador); ok {
		m.trabajador = *t
	}
	return 1, nil
}

// utility to build controller
func buildContext(method, url, body string) (*TrabajadorController, *httptest.ResponseRecorder) {
	r := httptest.NewRequest(method, url, strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context2.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := &TrabajadorController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	return c, w
}

func TestHashPassword(t *testing.T) {
	hashed, err := hashPassword("secret")
	if err != nil {
		t.Fatalf("err %v", err)
	}
	if hashed == "secret" {
		t.Fatalf("not hashed")
	}
}

func TestValidateDates(t *testing.T) {
	ingreso := time.Now()
	retiroAntes := ingreso.Add(-time.Hour)
	if validateDates(&ingreso, &retiroAntes) == nil {
		t.Fatalf("expected error")
	}
	retiroDespues := ingreso.Add(time.Hour)
	if err := validateDates(&ingreso, &retiroDespues); err != nil {
		t.Fatalf("unexpected %v", err)
	}
}

// GetAll tests
func TestGetAllError(t *testing.T) {
	database.BogotaZone = time.UTC
	original := newTrabajadorOrm
	newTrabajadorOrm = func() orm.Ormer { return &mockOrm{query: &mockQuery{err: errors.New("db")}} }
	defer func() { newTrabajadorOrm = original }()
	c, w := buildContext(http.MethodGet, "/trabajadores", "")
	c.GetAll()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("got %d", w.Code)
	}
}

func TestGetAllNoResults(t *testing.T) {
	database.BogotaZone = time.UTC
	mq := &mockQuery{trabajadores: []models.Trabajador{}}
	original := newTrabajadorOrm
	newTrabajadorOrm = func() orm.Ormer { return &mockOrm{query: mq} }
	defer func() { newTrabajadorOrm = original }()
	c, w := buildContext(http.MethodGet, "/trabajadores?fecha_ingreso=2023-01-01&rol=Admin", "")
	c.GetAll()
	if w.Code != http.StatusOK {
		t.Fatalf("got %d", w.Code)
	}
}

func TestGetAllSoloRetirados(t *testing.T) {
	database.BogotaZone = time.UTC
	f := time.Now()
	mq := &mockQuery{trabajadores: []models.Trabajador{{FECHA_RETIRO: &f, FECHA_NACIMIENTO: &f, FECHA_INGRESO: f}}}
	original := newTrabajadorOrm
	newTrabajadorOrm = func() orm.Ormer { return &mockOrm{query: mq} }
	defer func() { newTrabajadorOrm = original }()
	c, w := buildContext(http.MethodGet, "/trabajadores?solo_retirados=true", "")
	c.GetAll()
	if w.Code != http.StatusOK {
		t.Fatalf("got %d", w.Code)
	}
}

func TestGetAllInvalidDate(t *testing.T) {
	database.BogotaZone = time.UTC
	mq := &mockQuery{trabajadores: []models.Trabajador{}}
	original := newTrabajadorOrm
	newTrabajadorOrm = func() orm.Ormer { return &mockOrm{query: mq} }
	defer func() { newTrabajadorOrm = original }()
	c, w := buildContext(http.MethodGet, "/trabajadores?fecha_ingreso=bad", "")
	c.GetAll()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("got %d", w.Code)
	}
}

// GetById tests
func TestGetByIdInvalid(t *testing.T) {
	c, w := buildContext(http.MethodGet, "/trabajadores/search", "")
	c.GetById()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("%d", w.Code)
	}
}

func TestGetByIdNotFound(t *testing.T) {
	original := newTrabajadorOrm
	newTrabajadorOrm = func() orm.Ormer { return &mockOrm{readErr: orm.ErrNoRows} }
	defer func() { newTrabajadorOrm = original }()
	c, w := buildContext(http.MethodGet, "/trabajadores/search?id=1", "")
	c.GetById()
	if w.Code != http.StatusOK {
		t.Fatalf("%d", w.Code)
	}
}

func TestGetByIdSuccess(t *testing.T) {
	original := newTrabajadorOrm
	newTrabajadorOrm = func() orm.Ormer {
		return &mockOrm{trabajador: models.Trabajador{PK_DOCUMENTO_TRABAJADOR: 1, PASSWORD: "x"}}
	}
	defer func() { newTrabajadorOrm = original }()
	c, w := buildContext(http.MethodGet, "/trabajadores/search?id=1", "")
	c.GetById()
	if w.Code != http.StatusOK {
		t.Fatalf("%d", w.Code)
	}
}

func TestGetByIdDBError(t *testing.T) {
	original := newTrabajadorOrm
	newTrabajadorOrm = func() orm.Ormer { return &mockOrm{readErr: errors.New("db")} }
	defer func() { newTrabajadorOrm = original }()
	c, w := buildContext(http.MethodGet, "/trabajadores/search?id=1", "")
	c.GetById()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("%d", w.Code)
	}
}

// Post tests
func TestPostInvalidJSON(t *testing.T) {
	c, w := buildContext(http.MethodPost, "/trabajadores", "notjson")
	c.Post()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("%d", w.Code)
	}
}

func TestPostMissingFields(t *testing.T) {
	cases := []struct{ body, substr string }{
		{`{"nombre":"a"}`, "documentoTrabajador"},
		{`{"documentoTrabajador":1}`, "nombre"},
		{`{"documentoTrabajador":1,"nombre":"a"}`, "apellido"},
		{`{"documentoTrabajador":1,"nombre":"a","apellido":"b"}`, "rol"},
		{`{"documentoTrabajador":1,"nombre":"a","apellido":"b","rol":"Administrador"}`, "fechaIngreso"},
		{`{"documentoTrabajador":1,"nombre":"a","apellido":"b","rol":"Administrador","fechaIngreso":"2023-01-01"}`, "sueldo"},
		{`{"documentoTrabajador":1,"nombre":"a","apellido":"b","rol":"Administrador","fechaIngreso":"2023-01-01","sueldo":1}`, "password"},
	}
	for _, tc := range cases {
		c, w := buildContext(http.MethodPost, "/trabajadores", tc.body)
		c.Post()
		if w.Code != http.StatusBadRequest {
			t.Fatalf("case %s got %d", tc.substr, w.Code)
		}
		if !strings.Contains(w.Body.String(), tc.substr) {
			t.Fatalf("missing %s", tc.substr)
		}
	}
}

func TestPostInvalidDates(t *testing.T) {
	body := `{"documentoTrabajador":1,"nombre":"a","apellido":"b","rol":"Administrador","fechaIngreso":"bad","sueldo":1,"password":"p"}`
	c, w := buildContext(http.MethodPost, "/trabajadores", body)
	c.Post()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("%d", w.Code)
	}

	body2 := `{"documentoTrabajador":1,"nombre":"a","apellido":"b","rol":"Administrador","fechaIngreso":"2023-01-01","sueldo":1,"password":"p","fechaNacimiento":"bad"}`
	c2, w2 := buildContext(http.MethodPost, "/trabajadores", body2)
	c2.Post()
	if w2.Code != http.StatusBadRequest {
		t.Fatalf("%d", w2.Code)
	}
}

func TestPostHashError(t *testing.T) {
	originalHash := hashPassword
	hashPassword = func(string) (string, error) { return "", errors.New("hash") }
	defer func() { hashPassword = originalHash }()
	body := `{"documentoTrabajador":1,"nombre":"a","apellido":"b","rol":"Administrador","fechaIngreso":"2023-01-01","sueldo":1,"password":"p"}`
	c, w := buildContext(http.MethodPost, "/trabajadores", body)
	c.Post()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("%d", w.Code)
	}
}

func TestPostInsertError(t *testing.T) {
	original := newTrabajadorOrm
	newTrabajadorOrm = func() orm.Ormer { return &mockOrm{insertErr: errors.New("db")} }
	defer func() { newTrabajadorOrm = original }()
	body := `{"documentoTrabajador":1,"nombre":"a","apellido":"b","rol":"Administrador","fechaIngreso":"2023-01-01","sueldo":1,"password":"p"}`
	c, w := buildContext(http.MethodPost, "/trabajadores", body)
	c.Post()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("%d", w.Code)
	}
}

func TestPostSuccess(t *testing.T) {
	original := newTrabajadorOrm
	newTrabajadorOrm = func() orm.Ormer { return &mockOrm{} }
	defer func() { newTrabajadorOrm = original }()
	body := `{"documentoTrabajador":1,"nombre":"a","apellido":"b","rol":"c","fechaIngreso":"2023-01-01","sueldo":1,"password":"p","telefono":"1","restauranteId":1,"fechaNacimiento":"2023-01-02"}`
	c, w := buildContext(http.MethodPost, "/trabajadores", body)
	c.Post()
	if w.Code != http.StatusCreated {
		t.Fatalf("%d", w.Code)
	}
}

// Put tests
func TestPutInvalidID(t *testing.T) {
	c, w := buildContext(http.MethodPut, "/trabajadores", "{}")
	c.Put()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("%d", w.Code)
	}
}

func TestPutNotFound(t *testing.T) {
	original := newTrabajadorOrm
	newTrabajadorOrm = func() orm.Ormer { return &mockOrm{readErr: orm.ErrNoRows} }
	defer func() { newTrabajadorOrm = original }()
	c, w := buildContext(http.MethodPut, "/trabajadores?id=1", "{}")
	c.Put()
	if w.Code != http.StatusOK {
		t.Fatalf("%d", w.Code)
	}
}

func TestPutDecodeError(t *testing.T) {
	original := newTrabajadorOrm
	newTrabajadorOrm = func() orm.Ormer { return &mockOrm{} }
	defer func() { newTrabajadorOrm = original }()
	c, w := buildContext(http.MethodPut, "/trabajadores?id=1", "notjson")
	c.Put()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("%d", w.Code)
	}
}

func TestPutInvalidDates(t *testing.T) {
	original := newTrabajadorOrm
	newTrabajadorOrm = func() orm.Ormer { return &mockOrm{} }
	defer func() { newTrabajadorOrm = original }()
	body := `{"fechaIngreso":"bad"}`
	c, w := buildContext(http.MethodPut, "/trabajadores?id=1", body)
	c.Put()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("%d", w.Code)
	}

	body2 := `{"fechaRetiro":"bad"}`
	c2, w2 := buildContext(http.MethodPut, "/trabajadores?id=1", body2)
	c2.Put()
	if w2.Code != http.StatusBadRequest {
		t.Fatalf("%d", w2.Code)
	}

	body3 := `{"fechaNacimiento":"bad"}`
	c3, w3 := buildContext(http.MethodPut, "/trabajadores?id=1", body3)
	c3.Put()
	if w3.Code != http.StatusBadRequest {
		t.Fatalf("%d", w3.Code)
	}
}

func TestPutHashError(t *testing.T) {
	originalOrm := newTrabajadorOrm
	newTrabajadorOrm = func() orm.Ormer { return &mockOrm{} }
	defer func() { newTrabajadorOrm = originalOrm }()
	originalHash := hashPassword
	hashPassword = func(string) (string, error) { return "", errors.New("hash") }
	defer func() { hashPassword = originalHash }()
	body := `{"password":"p"}`
	c, w := buildContext(http.MethodPut, "/trabajadores?id=1", body)
	c.Put()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("%d", w.Code)
	}
}

func TestPutValidateDatesError(t *testing.T) {
	original := newTrabajadorOrm
	newTrabajadorOrm = func() orm.Ormer { return &mockOrm{} }
	defer func() { newTrabajadorOrm = original }()
	body := `{"fechaIngreso":"2023-01-02","fechaRetiro":"2023-01-01"}`
	c, w := buildContext(http.MethodPut, "/trabajadores?id=1", body)
	c.Put()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("%d", w.Code)
	}
}

func TestPutUpdateError(t *testing.T) {
	original := newTrabajadorOrm
	newTrabajadorOrm = func() orm.Ormer { return &mockOrm{updateErr: errors.New("db")} }
	defer func() { newTrabajadorOrm = original }()
	body := `{"nombre":"a"}`
	c, w := buildContext(http.MethodPut, "/trabajadores?id=1", body)
	c.Put()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("%d", w.Code)
	}
}

func TestPutSuccess(t *testing.T) {
	original := newTrabajadorOrm
	newTrabajadorOrm = func() orm.Ormer { return &mockOrm{} }
	defer func() { newTrabajadorOrm = original }()
	body := `{"nombre":"a","apellido":"b","rol":"c","sueldo":1,"nuevo":true,"telefono":"1","fechaIngreso":"2023-01-01","fechaRetiro":"2023-01-02","fechaNacimiento":"2023-01-03","password":"p"}`
	c, w := buildContext(http.MethodPut, "/trabajadores?id=1", body)
	c.Put()
	if w.Code != http.StatusOK {
		t.Fatalf("%d", w.Code)
	}
}

func TestPutTelefonoUpdated(t *testing.T) {
	m := &mockOrm{}
	original := newTrabajadorOrm
	newTrabajadorOrm = func() orm.Ormer { return m }
	defer func() { newTrabajadorOrm = original }()
	body := `{"telefono":"123"}`
	c, w := buildContext(http.MethodPut, "/trabajadores?id=1", body)
	c.Put()
	if w.Code != http.StatusOK {
		t.Fatalf("%d", w.Code)
	}
	if m.trabajador.TELEFONO == nil || *m.trabajador.TELEFONO != "123" {
		t.Fatalf("telefono not updated")
	}
}

// Delete tests
func TestDeleteInvalidID(t *testing.T) {
	c, w := buildContext(http.MethodDelete, "/trabajadores", "")
	c.Delete()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("%d", w.Code)
	}
}

func TestDeleteNotFound(t *testing.T) {
	original := newTrabajadorOrm
	newTrabajadorOrm = func() orm.Ormer { return &mockOrm{readErr: orm.ErrNoRows} }
	defer func() { newTrabajadorOrm = original }()
	c, w := buildContext(http.MethodDelete, "/trabajadores?id=1", "")
	c.Delete()
	if w.Code != http.StatusOK {
		t.Fatalf("%d", w.Code)
	}
}

func TestDeleteReadError(t *testing.T) {
	original := newTrabajadorOrm
	newTrabajadorOrm = func() orm.Ormer { return &mockOrm{readErr: errors.New("db")} }
	defer func() { newTrabajadorOrm = original }()
	c, w := buildContext(http.MethodDelete, "/trabajadores?id=1", "")
	c.Delete()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("%d", w.Code)
	}
}

func TestDeleteUpdateError(t *testing.T) {
	database.BogotaZone = time.UTC
	original := newTrabajadorOrm
	newTrabajadorOrm = func() orm.Ormer { return &mockOrm{updateErr: errors.New("db")} }
	defer func() { newTrabajadorOrm = original }()
	c, w := buildContext(http.MethodDelete, "/trabajadores?id=1", "")
	c.Delete()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("%d", w.Code)
	}
}

func TestDeleteSuccess(t *testing.T) {
	database.BogotaZone = time.UTC
	original := newTrabajadorOrm
	newTrabajadorOrm = func() orm.Ormer { return &mockOrm{} }
	defer func() { newTrabajadorOrm = original }()
	c, w := buildContext(http.MethodDelete, "/trabajadores?id=1", "")
	c.Delete()
	if w.Code != http.StatusOK {
		t.Fatalf("%d", w.Code)
	}
}
