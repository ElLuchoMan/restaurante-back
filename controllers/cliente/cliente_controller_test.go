package cliente

import (
	stdctx "context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
	"golang.org/x/crypto/bcrypt"
)

func resetMocks() {
	ormNew = orm.NewOrm
	queryAllClientes = func(o orm.Ormer, clientes *[]models.Cliente) (int64, error) {
		return o.QueryTable(new(models.Cliente)).All(clientes)
	}
	readCliente = func(o orm.Ormer, c *models.Cliente) error { return o.Read(c) }
	insertCliente = func(o orm.Ormer, c *models.Cliente) (int64, error) { return o.Insert(c) }
	updateCliente = func(o orm.Ormer, c *models.Cliente) (int64, error) { return o.Update(c) }
	deleteCliente = func(o orm.Ormer, c *models.Cliente) (int64, error) { return o.Delete(c) }
	bcryptGenerate = bcrypt.GenerateFromPassword
}

func TestNormalizeEmail(t *testing.T) {
	got := normalizeEmail("  Foo@Example.COM  ")
	if got != "foo@example.com" {
		t.Errorf("expected foo@example.com, got %s", got)
	}
}

func TestIsUniqueEmailErr(t *testing.T) {
	uniqueErr := errors.New("uq_cliente_correo")
	if !isUniqueEmailErr(uniqueErr) {
		t.Errorf("expected true for unique email error")
	}
	uniqueMsgErr := errors.New("unique constraint on correo")
	if !isUniqueEmailErr(uniqueMsgErr) {
		t.Errorf("expected true for unique correo message")
	}
	otherErr := errors.New("other")
	if isUniqueEmailErr(otherErr) {
		t.Errorf("expected false for non unique email error")
	}
	if isUniqueEmailErr(nil) {
		t.Errorf("expected false for nil error")
	}
}

func TestClienteGetAllSuccess(t *testing.T) {
	db := map[int64]models.Cliente{1: {PK_DOCUMENTO_CLIENTE: 1, NOMBRE: "Foo", APELLIDO: "Bar", TELEFONO: "123", PASSWORD: "pwd"}}
	ormNew = func() orm.Ormer { return nil }
	queryAllClientes = func(o orm.Ormer, clientes *[]models.Cliente) (int64, error) {
		for _, c := range db {
			*clientes = append(*clientes, c)
		}
		return int64(len(db)), nil
	}
	t.Cleanup(resetMocks)
	r := httptest.NewRequest(http.MethodGet, "/clientes", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ClienteController{}
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}
	c.GetAll()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if strings.Contains(w.Body.String(), "pwd") {
		t.Fatalf("password should be removed")
	}
}

func TestClienteGetAllFiltered(t *testing.T) {
	db := map[int64]models.Cliente{1: {PK_DOCUMENTO_CLIENTE: 1, NOMBRE: "Foo", APELLIDO: "Bar", TELEFONO: "123", PASSWORD: "pwd"}}
	ormNew = func() orm.Ormer { return nil }
	queryAllClientes = func(o orm.Ormer, clientes *[]models.Cliente) (int64, error) {
		for _, c := range db {
			*clientes = append(*clientes, c)
		}
		return 1, nil
	}
	t.Cleanup(resetMocks)
	r := httptest.NewRequest(http.MethodGet, "/clientes?fields=nombre_completo_telefono", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ClienteController{}
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}
	c.GetAll()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "nombre_completo") {
		t.Fatalf("expected filtered response")
	}
}

func TestClienteGetAllDBError(t *testing.T) {
	ormNew = func() orm.Ormer { return nil }
	queryAllClientes = func(o orm.Ormer, clientes *[]models.Cliente) (int64, error) {
		return 0, errors.New("db error")
	}
	t.Cleanup(resetMocks)
	r := httptest.NewRequest(http.MethodGet, "/clientes", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ClienteController{}
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}
	c.GetAll()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestClienteGetAllLimitOffsetSuccess(t *testing.T) {
	ormNew = orm.NewOrm
	MockQuery = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
		cols := []string{"pk_documento_cliente", "nombre", "apellido", "correo", "direccion", "telefono", "observaciones", "password"}
		vals := [][]driver.Value{{int64(1), "Foo", "Bar", "foo@bar.com", "Dir", "123", nil, "pwd"}}
		return &mockRows{columns: cols, values: vals}, nil
	}
	t.Cleanup(func() {
		MockQuery = nil
		resetMocks()
	})
	r := httptest.NewRequest(http.MethodGet, "/clientes?limit=0&offset=0", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ClienteController{}
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}
	c.GetAll()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if strings.Contains(w.Body.String(), "pwd") {
		t.Fatalf("password should be removed")
	}
}

func TestClienteGetAllLimitOnlySuccess(t *testing.T) {
	ormNew = orm.NewOrm
	MockQuery = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
		cols := []string{"pk_documento_cliente", "nombre", "apellido", "correo", "direccion", "telefono", "observaciones", "password"}
		vals := [][]driver.Value{{int64(1), "Foo", "Bar", "foo@bar.com", "Dir", "123", nil, "pwd"}}
		return &mockRows{columns: cols, values: vals}, nil
	}
	t.Cleanup(func() {
		MockQuery = nil
		resetMocks()
	})
	r := httptest.NewRequest(http.MethodGet, "/clientes?limit=5", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ClienteController{}
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}
	c.GetAll()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if strings.Contains(w.Body.String(), "pwd") {
		t.Fatalf("password should be removed")
	}
}

func TestClienteGetAllOffsetOnlySuccess(t *testing.T) {
	ormNew = orm.NewOrm
	MockQuery = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
		cols := []string{"pk_documento_cliente", "nombre", "apellido", "correo", "direccion", "telefono", "observaciones", "password"}
		vals := [][]driver.Value{{int64(1), "Foo", "Bar", "foo@bar.com", "Dir", "123", nil, "pwd"}}
		return &mockRows{columns: cols, values: vals}, nil
	}
	t.Cleanup(func() {
		MockQuery = nil
		resetMocks()
	})
	r := httptest.NewRequest(http.MethodGet, "/clientes?offset=5", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ClienteController{}
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}
	c.GetAll()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if strings.Contains(w.Body.String(), "pwd") {
		t.Fatalf("password should be removed")
	}
}

func TestClienteGetAllLimitOffsetError(t *testing.T) {
	ormNew = orm.NewOrm
	MockQuery = func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
		return nil, errors.New("db error")
	}
	t.Cleanup(func() {
		MockQuery = nil
		resetMocks()
	})
	r := httptest.NewRequest(http.MethodGet, "/clientes?limit=5&offset=1", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ClienteController{}
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}
	c.GetAll()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestClienteGetAllLimitParseError(t *testing.T) {
	ormNew = func() orm.Ormer { return nil }
	t.Cleanup(resetMocks)
	r := httptest.NewRequest(http.MethodGet, "/clientes?limit=abc", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ClienteController{}
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}
	c.GetAll()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestClienteGetAllOffsetParseError(t *testing.T) {
	ormNew = func() orm.Ormer { return nil }
	t.Cleanup(resetMocks)
	r := httptest.NewRequest(http.MethodGet, "/clientes?offset=abc", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ClienteController{}
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}
	c.GetAll()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestClienteGetByIdScenarios(t *testing.T) {
	db := map[int64]models.Cliente{1: {PK_DOCUMENTO_CLIENTE: 1, NOMBRE: "Foo", PASSWORD: "pwd"}}
	ormNew = func() orm.Ormer { return nil }
	var readErr error
	readCliente = func(o orm.Ormer, c *models.Cliente) error {
		if readErr != nil {
			return readErr
		}
		cli, ok := db[c.PK_DOCUMENTO_CLIENTE]
		if !ok {
			return orm.ErrNoRows
		}
		*c = cli
		return nil
	}
	t.Cleanup(resetMocks)
	r := httptest.NewRequest(http.MethodGet, "/clientes/search", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ClienteController{}
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}
	c.GetById()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	r = httptest.NewRequest(http.MethodGet, "/clientes/search?id=2", nil)
	w = httptest.NewRecorder()
	ctx = context.NewContext()
	ctx.Reset(w, r)
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}
	c.GetById()
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "Cliente no encontrado") {
		t.Fatalf("not found case failed")
	}
	readErr = errors.New("db error")
	r = httptest.NewRequest(http.MethodGet, "/clientes/search?id=1", nil)
	w = httptest.NewRecorder()
	ctx = context.NewContext()
	ctx.Reset(w, r)
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}
	c.GetById()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
	readErr = nil
	r = httptest.NewRequest(http.MethodGet, "/clientes/search?id=1", nil)
	w = httptest.NewRecorder()
	ctx = context.NewContext()
	ctx.Reset(w, r)
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}
	c.GetById()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if strings.Contains(w.Body.String(), "pwd") {
		t.Fatalf("password should be removed")
	}
}

func TestClientePostScenarios(t *testing.T) {
	db := make(map[int64]models.Cliente)
	ormNew = func() orm.Ormer { return nil }
	insertCliente = func(o orm.Ormer, c *models.Cliente) (int64, error) {
		if _, ok := db[c.PK_DOCUMENTO_CLIENTE]; ok {
			return 0, errors.New("unique correo")
		}
		db[c.PK_DOCUMENTO_CLIENTE] = *c
		return 1, nil
	}
	t.Cleanup(resetMocks)
	r := httptest.NewRequest(http.MethodPost, "/clientes", strings.NewReader("notjson"))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.CopyBody(1 << 20)
	c := ClienteController{}
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}
	c.Post()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	bodyNoCorreo := `{"documentoCliente":1,"nombre":"Foo","apellido":"B","direccion":"C","telefono":"1","password":"pass"}`
	r = httptest.NewRequest(http.MethodPost, "/clientes", strings.NewReader(bodyNoCorreo))
	w = httptest.NewRecorder()
	ctx = context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.CopyBody(1 << 20)
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}
	c.Post()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	insertCliente = func(o orm.Ormer, c *models.Cliente) (int64, error) { return 0, errors.New("db error") }
	body := `{"documentoCliente":1,"nombre":"Foo","apellido":"B","direccion":"C","telefono":"1","password":"pass","correo":" TeSt@Email.com "}`
	r = httptest.NewRequest(http.MethodPost, "/clientes", strings.NewReader(body))
	w = httptest.NewRecorder()
	ctx = context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.CopyBody(1 << 20)
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}
	c.Post()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
	insertCliente = func(o orm.Ormer, c *models.Cliente) (int64, error) { return 0, errors.New("unique correo") }
	r = httptest.NewRequest(http.MethodPost, "/clientes", strings.NewReader(body))
	w = httptest.NewRecorder()
	ctx = context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.CopyBody(1 << 20)
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}
	c.Post()
	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w.Code)
	}
	bcryptGenerate = func([]byte, int) ([]byte, error) { return nil, errors.New("hash") }
	r = httptest.NewRequest(http.MethodPost, "/clientes", strings.NewReader(body))
	w = httptest.NewRecorder()
	ctx = context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.CopyBody(1 << 20)
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}
	c.Post()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
	bcryptGenerate = bcrypt.GenerateFromPassword
	insertCliente = func(o orm.Ormer, c *models.Cliente) (int64, error) {
		db[c.PK_DOCUMENTO_CLIENTE] = *c
		return 1, nil
	}
	r = httptest.NewRequest(http.MethodPost, "/clientes", strings.NewReader(body))
	w = httptest.NewRecorder()
	ctx = context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.CopyBody(1 << 20)
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}
	c.Post()
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	var resp struct {
		Data models.Cliente `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if resp.Data.PASSWORD != "" {
		t.Fatalf("password should be empty")
	}
}

func TestClientePutScenarios(t *testing.T) {
	db := map[int64]models.Cliente{1: {PK_DOCUMENTO_CLIENTE: 1, NOMBRE: "Foo", PASSWORD: "old"}, 2: {PK_DOCUMENTO_CLIENTE: 2, NOMBRE: "Bar", CORREO: "a@a.com", PASSWORD: "x"}}
	ormNew = func() orm.Ormer { return nil }
	readCliente = func(o orm.Ormer, c *models.Cliente) error {
		cli, ok := db[c.PK_DOCUMENTO_CLIENTE]
		if !ok {
			return orm.ErrNoRows
		}
		*c = cli
		return nil
	}
	updateCliente = func(o orm.Ormer, c *models.Cliente) (int64, error) {
		if c.NOMBRE == "fail" {
			return 0, errors.New("db error")
		}
		for id, cli := range db {
			if id != c.PK_DOCUMENTO_CLIENTE && cli.CORREO != "" && c.CORREO != "" && cli.CORREO == c.CORREO {
				return 0, errors.New("unique correo")
			}
		}
		db[c.PK_DOCUMENTO_CLIENTE] = *c
		return 1, nil
	}
	t.Cleanup(resetMocks)
	r := httptest.NewRequest(http.MethodPut, "/clientes", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ClienteController{}
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}
	c.Put()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	readCliente = func(o orm.Ormer, c *models.Cliente) error { return errors.New("db") }
	r = httptest.NewRequest(http.MethodPut, "/clientes?id=1", strings.NewReader(`{"nombre":"Foo"}`))
	w = httptest.NewRecorder()
	ctx = context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.CopyBody(1 << 20)
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}
	c.Put()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
	readCliente = func(o orm.Ormer, c *models.Cliente) error {
		cli, ok := db[c.PK_DOCUMENTO_CLIENTE]
		if !ok {
			return orm.ErrNoRows
		}
		*c = cli
		return nil
	}
	r = httptest.NewRequest(http.MethodPut, "/clientes?id=3", strings.NewReader(`{"nombre":"Foo"}`))
	w = httptest.NewRecorder()
	ctx = context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.CopyBody(1 << 20)
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}
	c.Put()
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "Cliente no encontrado") {
		t.Fatalf("not found failed")
	}
	r = httptest.NewRequest(http.MethodPut, "/clientes?id=1", strings.NewReader("notjson"))
	w = httptest.NewRecorder()
	ctx = context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.CopyBody(1 << 20)
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}
	c.Put()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	r = httptest.NewRequest(http.MethodPut, "/clientes?id=1", strings.NewReader(`{"correo":"a@a.com"}`))
	w = httptest.NewRecorder()
	ctx = context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.CopyBody(1 << 20)
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}
	c.Put()
	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w.Code)
	}
	r = httptest.NewRequest(http.MethodPut, "/clientes?id=1", strings.NewReader(`{"nombre":"fail"}`))
	w = httptest.NewRecorder()
	ctx = context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.CopyBody(1 << 20)
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}
	c.Put()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
	bcryptGenerate = func([]byte, int) ([]byte, error) { return nil, errors.New("hash") }
	r = httptest.NewRequest(http.MethodPut, "/clientes?id=1", strings.NewReader(`{"password":"n"}`))
	w = httptest.NewRecorder()
	ctx = context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.CopyBody(1 << 20)
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}
	c.Put()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
	bcryptGenerate = bcrypt.GenerateFromPassword
	r = httptest.NewRequest(http.MethodPut, "/clientes?id=1", strings.NewReader(`{"nombre":"New","password":"newpass"}`))
	w = httptest.NewRecorder()
	ctx = context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.CopyBody(1 << 20)
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}
	c.Put()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if strings.Contains(w.Body.String(), "newpass") {
		t.Fatalf("password should be hidden")
	}
	if db[1].PASSWORD == "newpass" {
		t.Fatalf("password should be hashed")
	}
	r = httptest.NewRequest(http.MethodPut, "/clientes?id=1", strings.NewReader(`{"nombre":"Other"}`))
	w = httptest.NewRecorder()
	ctx = context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.CopyBody(1 << 20)
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}
	c.Put()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if db[1].PASSWORD == "" {
		t.Fatalf("password should remain")
	}
}

func TestClienteDeleteScenarios(t *testing.T) {
	db := map[int64]models.Cliente{1: {PK_DOCUMENTO_CLIENTE: 1}}
	ormNew = func() orm.Ormer { return nil }
	deleteCliente = func(o orm.Ormer, c *models.Cliente) (int64, error) {
		if _, ok := db[c.PK_DOCUMENTO_CLIENTE]; ok {
			delete(db, c.PK_DOCUMENTO_CLIENTE)
			return 1, nil
		}
		return 0, errors.New("not found")
	}
	t.Cleanup(resetMocks)
	r := httptest.NewRequest(http.MethodDelete, "/clientes", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := ClienteController{}
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}
	c.Delete()
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	r = httptest.NewRequest(http.MethodDelete, "/clientes?id=2", nil)
	w = httptest.NewRecorder()
	ctx = context.NewContext()
	ctx.Reset(w, r)
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}
	c.Delete()
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "Cliente no encontrado") {
		t.Fatalf("expected not found response")
	}
	r = httptest.NewRequest(http.MethodDelete, "/clientes?id=1", nil)
	w = httptest.NewRecorder()
	ctx = context.NewContext()
	ctx.Reset(w, r)
	c.Ctx = ctx
	c.Data = map[interface{}]interface{}{}
	c.Delete()
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "Cliente eliminado") {
		t.Fatalf("expected success")
	}
}

func TestDefaultWrappersCoverage(t *testing.T) {
	useDefaultClienteWrappers()
	t.Cleanup(resetMocks)
	o := ormNew()

	var list []models.Cliente
	if _, err := queryAllClientes(o, &list); err != nil {
		t.Fatalf("queryAllClientes error: %v", err)
	}

	c := &models.Cliente{PK_DOCUMENTO_CLIENTE: 1}
	_ = readCliente(o, c)

	_, _ = insertCliente(o, &models.Cliente{PK_DOCUMENTO_CLIENTE: 99, NOMBRE: "X"})

	_, _ = updateCliente(o, &models.Cliente{PK_DOCUMENTO_CLIENTE: 99, NOMBRE: "Y"})

	_, _ = deleteCliente(o, &models.Cliente{PK_DOCUMENTO_CLIENTE: 99})
}
