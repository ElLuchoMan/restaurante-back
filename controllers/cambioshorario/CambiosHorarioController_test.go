package cambioshorario

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
)

// Driver/mock mínimo para inicializar orm.NewOrm() sin DB real
type mockDriver struct{}

type mockConn struct{}

type mockStmt struct{ query string }

type mockTx struct{}

func (d mockDriver) Open(name string) (driver.Conn, error) { return &mockConn{}, nil }

func (c *mockConn) Prepare(query string) (driver.Stmt, error) { return &mockStmt{query: query}, nil }
func (c *mockConn) Close() error                              { return nil }
func (c *mockConn) Begin() (driver.Tx, error)                 { return &mockTx{}, nil }

func (s *mockStmt) Close() error                                    { return nil }
func (s *mockStmt) NumInput() int                                   { return -1 }
func (s *mockStmt) Exec(args []driver.Value) (driver.Result, error) { return mockResult{}, nil }
func (s *mockStmt) Query(args []driver.Value) (driver.Rows, error)  { return &mockRows{}, nil }

func (mockTx) Commit() error   { return nil }
func (mockTx) Rollback() error { return nil }

type mockResult struct{}

func (mockResult) LastInsertId() (int64, error) { return 1, nil }
func (mockResult) RowsAffected() (int64, error) { return 1, nil }

type mockRows struct{}

func (r *mockRows) Columns() []string              { return []string{} }
func (r *mockRows) Close() error                   { return nil }
func (r *mockRows) Next(dest []driver.Value) error { return io.EOF }

func TestMain(m *testing.M) {
	// Evitar pánico por JWT en inits de otros paquetes
	_ = os.Setenv("JWT_SECRET", "testsecret")
	// Registrar driver y base por defecto para que orm.NewOrm() no falle
	sql.Register("mock", mockDriver{})
	orm.RegisterDriver("mock", orm.DRPostgres)
	_ = orm.RegisterDataBase("default", "mock", "")
	os.Exit(m.Run())
}

func setupCtx(method, url string, body string) (*CambiosHorarioController, *httptest.ResponseRecorder) {
	r := httptest.NewRequest(method, url, strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, r)
	if body != "" {
		ctx.Input.RequestBody = []byte(body)
	}
	c := &CambiosHorarioController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	return c, w
}

// Copia íntegra de los tests existentes adaptados al nuevo paquete
// (contenido idéntico al archivo original en controllers/ con package actualizado)

func TestCambiosHorario_GetAll_DBError(t *testing.T) {
	orig := queryAllCambiosHorario
	queryAllCambiosHorario = func(o orm.Ormer, horarios *[]models.CambiosHorario) (int64, error) {
		return 0, errors.New("db fail")
	}
	defer func() { queryAllCambiosHorario = orig }()

	c, w := setupCtx(http.MethodGet, "/cambios_horario", "")
	c.GetAll()

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error al obtener cambios de horario") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

// ... existing code ...
