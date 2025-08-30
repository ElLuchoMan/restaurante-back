package controllers

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"

	"github.com/beego/beego/v2/client/orm"
)

type mockDriver struct{}

func (d mockDriver) Open(name string) (driver.Conn, error) {
	return &mockConn{}, nil
}

var (
	// MockExec and MockQuery allow tests to override database behaviour.
	MockExec  func(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error)
	MockQuery func(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error)
)

type mockConn struct{}

func (c *mockConn) Prepare(query string) (driver.Stmt, error) {
	return nil, errors.New("not implemented")
}
func (c *mockConn) Close() error              { return nil }
func (c *mockConn) Begin() (driver.Tx, error) { return nil, errors.New("not implemented") }
func (c *mockConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	if MockExec != nil {
		return MockExec(ctx, query, args)
	}
	return nil, errors.New("mock exec error")
}
func (c *mockConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if MockQuery != nil {
		return MockQuery(ctx, query, args)
	}
	return nil, errors.New("mock query error")
}
func (c *mockConn) Ping(ctx context.Context) error { return nil }

func init() {
	sql.Register("mock", mockDriver{})
	orm.RegisterDriver("mock", orm.DRPostgres)
	orm.RegisterDataBase("default", "mock", "")
}

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
