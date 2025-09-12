package horariotrabajador

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
	"os"
	"testing"

	"github.com/beego/beego/v2/client/orm"
)

var (
	MockExec  func(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error)
	MockQuery func(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error)
)

type mockDriver struct{}

type mockConn struct{}

type mockStmt struct{ query string }

type mockTx struct{}

func (d mockDriver) Open(name string) (driver.Conn, error) { return &mockConn{}, nil }

func (c *mockConn) Prepare(query string) (driver.Stmt, error) { return &mockStmt{query: query}, nil }
func (c *mockConn) Close() error                              { return nil }
func (c *mockConn) Begin() (driver.Tx, error)                 { return &mockTx{}, nil }
func (c *mockConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	if MockExec != nil {
		return MockExec(ctx, query, args)
	}
	return mockResult{}, nil
}
func (c *mockConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if MockQuery != nil {
		return MockQuery(ctx, query, args)
	}
	return &mockRows{}, nil
}

func (s *mockStmt) Close() error  { return nil }
func (s *mockStmt) NumInput() int { return -1 }
func (s *mockStmt) Exec(args []driver.Value) (driver.Result, error) {
	if MockExec != nil {
		named := make([]driver.NamedValue, len(args))
		for i, v := range args {
			named[i] = driver.NamedValue{Ordinal: i + 1, Value: v}
		}
		return MockExec(context.Background(), s.query, named)
	}
	return mockResult{}, nil
}
func (s *mockStmt) Query(args []driver.Value) (driver.Rows, error) {
	if MockQuery != nil {
		named := make([]driver.NamedValue, len(args))
		for i, v := range args {
			named[i] = driver.NamedValue{Ordinal: i + 1, Value: v}
		}
		return MockQuery(context.Background(), s.query, named)
	}
	return &mockRows{}, nil
}

func (mockTx) Commit() error   { return nil }
func (mockTx) Rollback() error { return nil }

type mockResult struct{}

func (mockResult) LastInsertId() (int64, error) { return 1, nil }
func (mockResult) RowsAffected() (int64, error) { return 1, nil }

type mockRows struct {
	columns []string
	values  [][]driver.Value
	idx     int
}

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
