package pago

import (
	"database/sql"
	"database/sql/driver"
	"os"
	"testing"

	"github.com/beego/beego/v2/client/orm"
)

type mockDriver struct{}

type mockConn struct{}

type mockStmt struct{}

type mockTx struct{}

func (d mockDriver) Open(name string) (driver.Conn, error) { return &mockConn{}, nil }

func (c *mockConn) Prepare(query string) (driver.Stmt, error) { return &mockStmt{}, nil }
func (c *mockConn) Close() error                              { return nil }
func (c *mockConn) Begin() (driver.Tx, error)                 { return &mockTx{}, nil }

func (s *mockStmt) Close() error  { return nil }
func (s *mockStmt) NumInput() int { return -1 }
func (s *mockStmt) Exec(args []driver.Value) (driver.Result, error) {
	return driver.RowsAffected(0), nil
}
func (s *mockStmt) Query(args []driver.Value) (driver.Rows, error) { return &emptyRows{}, nil }

func (mockTx) Commit() error   { return nil }
func (mockTx) Rollback() error { return nil }

type emptyRows struct{}

func (r *emptyRows) Columns() []string              { return []string{} }
func (r *emptyRows) Close() error                   { return nil }
func (r *emptyRows) Next(dest []driver.Value) error { return driver.ErrBadConn }

func TestMain(m *testing.M) {
	if os.Getenv("JWT_SECRET") == "" {
		_ = os.Setenv("JWT_SECRET", "testsecret")
	}
	sql.Register("mock", mockDriver{})
	orm.RegisterDriver("mock", orm.DRPostgres)
	_ = orm.RegisterDataBase("default", "mock", "")
	os.Exit(m.Run())
}
