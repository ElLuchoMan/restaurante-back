package trabajador

import (
	"database/sql"
	"database/sql/driver"
	"io"
	"os"
	"testing"

	"github.com/beego/beego/v2/client/orm"
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

func (s *mockStmt) Close() error                                    { return nil }
func (s *mockStmt) NumInput() int                                   { return -1 }
func (s *mockStmt) Exec(args []driver.Value) (driver.Result, error) { return mockResult{}, nil }
func (s *mockStmt) Query(args []driver.Value) (driver.Rows, error) {
	return &mockRows{columns: []string{"ok"}, values: [][]driver.Value{{int64(1)}}}, nil
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
	sql.Register("mock", mockDriver{})
	orm.RegisterDriver("mock", orm.DRPostgres)
	_ = orm.RegisterDataBase("default", "mock", "")
	os.Exit(m.Run())
}
