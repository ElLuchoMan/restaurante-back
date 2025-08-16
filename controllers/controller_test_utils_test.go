package controllers

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"

	"github.com/beego/beego/v2/client/orm"
)

type mockDriver struct{}

func (d mockDriver) Open(name string) (driver.Conn, error) {
	return &mockConn{}, nil
}

type mockConn struct{}

func (c *mockConn) Prepare(query string) (driver.Stmt, error) {
	return nil, errors.New("not implemented")
}
func (c *mockConn) Close() error              { return nil }
func (c *mockConn) Begin() (driver.Tx, error) { return nil, errors.New("not implemented") }
func (c *mockConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	return nil, errors.New("mock exec error")
}
func (c *mockConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	return nil, errors.New("mock query error")
}
func (c *mockConn) Ping(ctx context.Context) error { return nil }

func init() {
	sql.Register("mock", mockDriver{})
	orm.RegisterDriver("mock", orm.DRPostgres)
	orm.RegisterDataBase("default", "mock", "")
}
