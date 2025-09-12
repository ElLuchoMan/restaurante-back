package precioproductohist

import (
	stdctx "context"
	"database/sql"
	"database/sql/driver"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/beego/beego/v2/client/orm"
	beegoCtx "github.com/beego/beego/v2/server/web/context"
)

var (
	MockExec  func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Result, error)
	MockQuery func(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Rows, error)
)

type mockDriver struct{}

type mockConn struct{}

type mockStmt struct{ query string }

type mockTx struct{}

type mockResult struct{}

type mockRows struct {
	columns []string
	values  [][]driver.Value
	idx     int
}

func (d mockDriver) Open(name string) (driver.Conn, error) { return &mockConn{}, nil }

func (c *mockConn) Prepare(query string) (driver.Stmt, error) { return &mockStmt{query: query}, nil }
func (c *mockConn) Close() error                              { return nil }
func (c *mockConn) Begin() (driver.Tx, error)                 { return &mockTx{}, nil }
func (c *mockConn) ExecContext(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	if MockExec != nil {
		return MockExec(ctx, query, args)
	}
	return mockResult{}, nil
}
func (c *mockConn) QueryContext(ctx stdctx.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if MockQuery != nil {
		return MockQuery(ctx, query, args)
	}
	return nil, driver.ErrBadConn
}

func (s *mockStmt) Close() error  { return nil }
func (s *mockStmt) NumInput() int { return -1 }
func (s *mockStmt) Exec(args []driver.Value) (driver.Result, error) {
	if MockExec != nil {
		nv := make([]driver.NamedValue, len(args))
		for i, v := range args {
			nv[i] = driver.NamedValue{Ordinal: i + 1, Value: v}
		}
		return MockExec(stdctx.Background(), s.query, nv)
	}
	return mockResult{}, nil
}
func (s *mockStmt) Query(args []driver.Value) (driver.Rows, error) {
	if MockQuery != nil {
		nv := make([]driver.NamedValue, len(args))
		for i, v := range args {
			nv[i] = driver.NamedValue{Ordinal: i + 1, Value: v}
		}
		return MockQuery(stdctx.Background(), s.query, nv)
	}
	return nil, driver.ErrBadConn
}

func (mockTx) Commit() error   { return nil }
func (mockTx) Rollback() error { return nil }

func (mockResult) LastInsertId() (int64, error) { return 1, nil }
func (mockResult) RowsAffected() (int64, error) { return 1, nil }

func (r *mockRows) Columns() []string { return r.columns }
func (r *mockRows) Close() error      { return nil }
func (r *mockRows) Next(dest []driver.Value) error {
	if r.idx >= len(r.values) {
		return io.EOF
	}
	row := r.values[r.idx]
	for i, v := range row {
		dest[i] = v
	}
	r.idx++
	return nil
}

func TestMain(m *testing.M) {
	if os.Getenv("JWT_SECRET") == "" {
		_ = os.Setenv("JWT_SECRET", "testsecret")
	}
	sql.Register("mock", mockDriver{})
	orm.RegisterDriver("mock", orm.DRPostgres)
	_ = orm.RegisterDataBase("default", "mock", "")
	os.Exit(m.Run())
}

func TestPPH_GetAll_Success(t *testing.T) {
	orig := MockQuery
	MockQuery = func(_ stdctx.Context, _ string, _ []driver.NamedValue) (driver.Rows, error) {
		cols := []string{"nombre", "estado_producto", "precio", "fecha_vigencia"}
		vals := [][]driver.Value{{"Cafe", "DISPONIBLE", int64(1000), nil}}
		return &mockRows{columns: cols, values: vals}, nil
	}
	t.Cleanup(func() { MockQuery = orig })

	r := httptest.NewRequest(http.MethodGet, "/precio_producto_hist?producto_id=1", nil)
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	c := &PrecioProductoHistController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.GetAll()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestPPH_GetById_Success(t *testing.T) {
	orig := MockQuery
	MockQuery = func(_ stdctx.Context, _ string, _ []driver.NamedValue) (driver.Rows, error) {
		cols := []string{"nombre", "estado_producto", "precio", "fecha_vigencia"}
		vals := [][]driver.Value{{"Te", "DISPONIBLE", int64(2000), nil}}
		return &mockRows{columns: cols, values: vals}, nil
	}
	t.Cleanup(func() { MockQuery = orig })

	r := httptest.NewRequest(http.MethodGet, "/precio_producto_hist/search?id=1", nil)
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	c := &PrecioProductoHistController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.GetById()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestPPH_GetAll_Error(t *testing.T) {
	orig := MockQuery
	MockQuery = func(_ stdctx.Context, _ string, _ []driver.NamedValue) (driver.Rows, error) {
		return nil, driver.ErrBadConn
	}
	t.Cleanup(func() { MockQuery = orig })

	r := httptest.NewRequest(http.MethodGet, "/precio_producto_hist", nil)
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	c := &PrecioProductoHistController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.GetAll()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestPPH_GetAll_SuccessWithFecha(t *testing.T) {
	orig := MockQuery
	MockQuery = func(_ stdctx.Context, q string, _ []driver.NamedValue) (driver.Rows, error) {
		cols := []string{"nombre", "estado_producto", "precio", "fecha_vigencia"}
		vals := [][]driver.Value{{"Cafe", "DISPONIBLE", int64(1000), nil}}
		return &mockRows{columns: cols, values: vals}, nil
	}
	t.Cleanup(func() { MockQuery = orig })

	r := httptest.NewRequest(http.MethodGet, "/precio_producto_hist?fecha=2024-01-01", nil)
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	c := &PrecioProductoHistController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.GetAll()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestPPH_GetAll_InvalidFecha(t *testing.T) {
	orig := MockQuery
	MockQuery = func(_ stdctx.Context, _ string, _ []driver.NamedValue) (driver.Rows, error) {
		cols := []string{"nombre", "estado_producto", "precio", "fecha_vigencia"}
		vals := [][]driver.Value{{"Cafe", "DISPONIBLE", int64(1000), nil}}
		return &mockRows{columns: cols, values: vals}, nil
	}
	t.Cleanup(func() { MockQuery = orig })

	r := httptest.NewRequest(http.MethodGet, "/precio_producto_hist?fecha=invalid", nil)
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	c := &PrecioProductoHistController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.GetAll()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestPPH_GetById_DBError(t *testing.T) {
	orig := MockQuery
	MockQuery = func(_ stdctx.Context, _ string, _ []driver.NamedValue) (driver.Rows, error) {
		return nil, driver.ErrBadConn
	}
	t.Cleanup(func() { MockQuery = orig })

	r := httptest.NewRequest(http.MethodGet, "/precio_producto_hist/search?id=2", nil)
	w := httptest.NewRecorder()
	ctx := beegoCtx.NewContext()
	ctx.Reset(w, r)
	c := &PrecioProductoHistController{}
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.GetById()
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
