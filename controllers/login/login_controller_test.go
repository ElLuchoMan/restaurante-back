package login

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
	"github.com/golang-jwt/jwt/v5"
)

// mockLoginOrmer implements only the Read method of orm.Ormer, allowing tests
// to customize the behaviour while satisfying the interface. Other methods are
// promoted from the embedded orm.Ormer and will panic if used since the field
// is nil, but our tests only rely on Read.
type mockLoginOrmer struct {
	orm.Ormer
	ReadFunc func(interface{}, ...string) error
}

func (m mockLoginOrmer) Read(v interface{}, cols ...string) error {
	return m.ReadFunc(v, cols...)
}

func TestGenerateJWT(t *testing.T) {
	os.Setenv("JWT_SECRET", "testsecret")
	defer os.Unsetenv("JWT_SECRET")
	jwtSecret = []byte(os.Getenv("JWT_SECRET"))

	r := httptest.NewRequest(http.MethodPost, "/login", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	c := LoginController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	generateJWT(&c, 123, "Admin", "Foo Bar")

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp models.ApiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Code != http.StatusOK {
		t.Errorf("expected response code 200, got %d", resp.Code)
	}
	data := resp.Data.(map[string]interface{})
	if data["token"] == "" {
		t.Errorf("expected token to be set")
	}
	if data["nombre"] != "Foo Bar" {
		t.Errorf("expected nombre Foo Bar, got %v", data["nombre"])
	}
}

func TestLoginInvalidJSON(t *testing.T) {
	os.Setenv("JWT_SECRET", "testsecret")
	defer os.Unsetenv("JWT_SECRET")
	r := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader("notjson"))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte("notjson")
	c := LoginController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Login()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestLoginTrabajadorSuccess(t *testing.T) {
	os.Setenv("JWT_SECRET", "testsecret")
	defer os.Unsetenv("JWT_SECRET")
	jwtSecret = []byte(os.Getenv("JWT_SECRET"))

	origNewOrm := newOrm
	defer func() { newOrm = origNewOrm }()
	newOrm = func() orm.Ormer {
		return &mockLoginOrmer{ReadFunc: func(v interface{}, cols ...string) error {
			if trab, ok := v.(*models.Trabajador); ok {
				trab.NOMBRE = "Foo"
				trab.APELLIDO = "Bar"
				trab.ROL = "Admin"
				trab.PASSWORD = "hashed"
				return nil
			}
			return errors.New("not found")
		}}
	}

	origCompare := compareHashAndPassword
	defer func() { compareHashAndPassword = origCompare }()
	compareHashAndPassword = func([]byte, []byte) error { return nil }

	body := `{"documento":123,"password":"secret"}`
	r := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := LoginController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Login()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	var resp models.ApiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Code != http.StatusOK {
		t.Fatalf("expected response code 200, got %d", resp.Code)
	}
	data := resp.Data.(map[string]interface{})
	if data["token"].(string) == "" {
		t.Errorf("expected token to be set")
	}
	if data["nombre"].(string) != "Foo Bar" {
		t.Errorf("expected nombre Foo Bar, got %v", data["nombre"])
	}
}

func TestLoginClienteSuccess(t *testing.T) {
	os.Setenv("JWT_SECRET", "testsecret")
	defer os.Unsetenv("JWT_SECRET")
	jwtSecret = []byte(os.Getenv("JWT_SECRET"))

	origNewOrm := newOrm
	defer func() { newOrm = origNewOrm }()
	newOrm = func() orm.Ormer {
		return &mockLoginOrmer{ReadFunc: func(v interface{}, cols ...string) error {
			switch val := v.(type) {
			case *models.Trabajador:
				return errors.New("not found")
			case *models.Cliente:
				val.NOMBRE = "Jane"
				val.APELLIDO = "Doe"
				val.PASSWORD = "hashed"
				return nil
			}
			return errors.New("unknown type")
		}}
	}

	origCompare := compareHashAndPassword
	defer func() { compareHashAndPassword = origCompare }()
	compareHashAndPassword = func([]byte, []byte) error { return nil }

	body := `{"documento":123,"password":"secret"}`
	r := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := LoginController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Login()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	var resp models.ApiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Code != http.StatusOK {
		t.Fatalf("expected response code 200, got %d", resp.Code)
	}
	data := resp.Data.(map[string]interface{})
	if data["token"].(string) == "" {
		t.Errorf("expected token to be set")
	}
	if data["nombre"].(string) != "Jane Doe" {
		t.Errorf("expected nombre Jane Doe, got %v", data["nombre"])
	}
}

func TestLoginTrabajadorInvalidPassword(t *testing.T) {
	origNewOrm := newOrm
	defer func() { newOrm = origNewOrm }()
	newOrm = func() orm.Ormer {
		return &mockLoginOrmer{ReadFunc: func(v interface{}, cols ...string) error {
			if trab, ok := v.(*models.Trabajador); ok {
				trab.PASSWORD = "hashed"
				return nil
			}
			return errors.New("not found")
		}}
	}

	origCompare := compareHashAndPassword
	defer func() { compareHashAndPassword = origCompare }()
	compareHashAndPassword = func([]byte, []byte) error { return errors.New("mismatch") }

	body := `{"documento":123,"password":"bad"}`
	r := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := LoginController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Login()

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestLoginClienteInvalidPassword(t *testing.T) {
	origNewOrm := newOrm
	defer func() { newOrm = origNewOrm }()
	newOrm = func() orm.Ormer {
		return &mockLoginOrmer{ReadFunc: func(v interface{}, cols ...string) error {
			switch val := v.(type) {
			case *models.Trabajador:
				return errors.New("not found")
			case *models.Cliente:
				val.PASSWORD = "hashed"
				return nil
			}
			return errors.New("unknown type")
		}}
	}

	origCompare := compareHashAndPassword
	defer func() { compareHashAndPassword = origCompare }()
	compareHashAndPassword = func([]byte, []byte) error { return errors.New("mismatch") }

	body := `{"documento":123,"password":"bad"}`
	r := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := LoginController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Login()

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestLoginUserNotFound(t *testing.T) {
	origNewOrm := newOrm
	defer func() { newOrm = origNewOrm }()
	newOrm = func() orm.Ormer {
		return &mockLoginOrmer{ReadFunc: func(v interface{}, cols ...string) error {
			return errors.New("not found")
		}}
	}

	body := `{"documento":123,"password":"secret"}`
	r := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	c := LoginController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})

	c.Login()

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestValidateTokenMissing(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)

	ValidateToken(ctx)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestValidateTokenInvalid(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/protected", nil)
	r.Header.Set("Authorization", "Bearer invalid")
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)

	ValidateToken(ctx)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestValidateTokenValid(t *testing.T) {
	os.Setenv("JWT_SECRET", "testsecret")
	defer os.Unsetenv("JWT_SECRET")
	jwtSecret = []byte(os.Getenv("JWT_SECRET"))

	claims := &Claims{Documento: 1, Rol: "Admin", Nombre: "Tester"}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(jwtSecret)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	r := httptest.NewRequest(http.MethodGet, "/protected", nil)
	r.Header.Set("Authorization", "Bearer "+tokenStr)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)

	ValidateToken(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestValidateTokenWithoutBearerPrefix(t *testing.T) {
	os.Setenv("JWT_SECRET", "testsecret")
	defer os.Unsetenv("JWT_SECRET")
	jwtSecret = []byte(os.Getenv("JWT_SECRET"))

	claims := &Claims{Documento: 1, Rol: "Admin", Nombre: "Tester"}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(jwtSecret)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	r := httptest.NewRequest(http.MethodGet, "/protected", nil)
	r.Header.Set("Authorization", tokenStr)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)

	ValidateToken(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestValidateTokenOptions(t *testing.T) {
	r := httptest.NewRequest(http.MethodOptions, "/any", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)

	ValidateToken(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestValidateTokenPublicCliente(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/restaurante/v1/clientes", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)

	ValidateToken(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestValidateTokenSwaggerRefererDev(t *testing.T) {
	orig := web.BConfig.RunMode
	web.BConfig.RunMode = "dev"
	t.Cleanup(func() { web.BConfig.RunMode = orig })

	r := httptest.NewRequest(http.MethodGet, "/any", nil)
	r.Header.Set("Referer", "http://localhost/swagger/")
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)

	ValidateToken(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestValidateTokenSwaggerURLDev(t *testing.T) {
	orig := web.BConfig.RunMode
	web.BConfig.RunMode = "dev"
	t.Cleanup(func() { web.BConfig.RunMode = orig })

	r := httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)

	ValidateToken(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}
