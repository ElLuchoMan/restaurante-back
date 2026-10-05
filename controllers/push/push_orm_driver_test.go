//go:build !unit

package push

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/beego/beego/v2/client/orm"
)

// Driver SQL falso y determinista para ejercitar los adaptadores reales de
// beego orm sin base de datos.

var (
	mockQueryHook func(query string) (driver.Rows, error)
	mockExecHook  func(query string) (driver.Result, error)
)

type fakeDriver struct{}
type fakeConn struct{}
type fakeStmt struct{ query string }
type fakeTx struct{}
type fakeResult struct{}
type fakeRows struct {
	columns []string
	values  [][]driver.Value
	idx     int
}

func (fakeDriver) Open(string) (driver.Conn, error)    { return fakeConn{}, nil }
func (fakeConn) Prepare(q string) (driver.Stmt, error) { return fakeStmt{query: q}, nil }
func (fakeConn) Close() error                          { return nil }
func (fakeConn) Begin() (driver.Tx, error)             { return fakeTx{}, nil }
func (fakeConn) Ping(context.Context) error            { return nil }
func (fakeStmt) Close() error                          { return nil }
func (fakeStmt) NumInput() int                         { return -1 }
func (fakeTx) Commit() error                           { return nil }
func (fakeTx) Rollback() error                         { return nil }
func (fakeResult) LastInsertId() (int64, error)        { return 1, nil }
func (fakeResult) RowsAffected() (int64, error)        { return 1, nil }
func (r *fakeRows) Columns() []string                  { return r.columns }
func (r *fakeRows) Close() error                       { return nil }
func (s fakeStmt) Exec([]driver.Value) (driver.Result, error) {
	if mockExecHook != nil {
		return mockExecHook(s.query)
	}
	return fakeResult{}, nil
}
func (s fakeStmt) Query([]driver.Value) (driver.Rows, error) {
	if mockQueryHook != nil {
		return mockQueryHook(s.query)
	}
	return defaultFakeRows(s.query), nil
}
func (r *fakeRows) Next(dest []driver.Value) error {
	if r.idx >= len(r.values) {
		return io.EOF
	}
	copy(dest, r.values[r.idx])
	r.idx++
	return nil
}

// defaultFakeRows devuelve una fila por consulta con tantas columnas como
// seleccione la sentencia (todas NULL salvo COUNT y RETURNING).
func defaultFakeRows(query string) driver.Rows {
	upper := strings.ToUpper(query)
	if strings.Contains(upper, "COUNT(") || strings.Contains(upper, "RETURNING") {
		return &fakeRows{columns: []string{"v"}, values: [][]driver.Value{{int64(3)}}}
	}
	if !strings.Contains(upper, "SELECT ") || !strings.Contains(upper, " FROM ") {
		// p. ej. detección de zona horaria durante RegisterDataBase
		return &fakeRows{columns: []string{"v"}, values: [][]driver.Value{{"UTC"}}}
	}
	start := strings.Index(upper, "SELECT ") + len("SELECT ")
	end := strings.Index(upper, " FROM ")
	n := strings.Count(query[start:end], ",") + 1
	cols := make([]string, n)
	row := make([]driver.Value, n)
	for i := range cols {
		cols[i] = "c"
	}
	return &fakeRows{columns: cols, values: [][]driver.Value{row}}
}

func TestMain(m *testing.M) {
	sql.Register("pushfake", fakeDriver{})
	orm.RegisterDriver("pushfake", orm.DRPostgres)
	_ = orm.RegisterDataBase("default", "pushfake", "")
	os.Exit(m.Run())
}
