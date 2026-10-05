package oferta

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
	"os"
	"testing"

	"github.com/beego/beego/v2/client/orm"
)

// Driver SQL falso registrado como base "default" para poder usar orm.NewOrm()
// sin base de datos real.
var (
	fakeExec  func(query string, args []driver.NamedValue) (driver.Result, error)
	fakeQuery func(query string, args []driver.NamedValue) (driver.Rows, error)
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

func (fakeDriver) Open(string) (driver.Conn, error) { return &fakeConn{}, nil }

func (c *fakeConn) Prepare(q string) (driver.Stmt, error) { return &fakeStmt{query: q}, nil }
func (c *fakeConn) Close() error                          { return nil }
func (c *fakeConn) Begin() (driver.Tx, error)             { return fakeTx{}, nil }
func (c *fakeConn) Ping(context.Context) error            { return nil }

func (s *fakeStmt) Close() error  { return nil }
func (s *fakeStmt) NumInput() int { return -1 }

func fakeNamed(args []driver.Value) []driver.NamedValue {
	named := make([]driver.NamedValue, len(args))
	for i, v := range args {
		named[i] = driver.NamedValue{Ordinal: i + 1, Value: v}
	}
	return named
}

func (s *fakeStmt) Exec(args []driver.Value) (driver.Result, error) {
	if fakeExec != nil {
		return fakeExec(s.query, fakeNamed(args))
	}
	return fakeResult{}, nil
}

func (s *fakeStmt) Query(args []driver.Value) (driver.Rows, error) {
	if fakeQuery != nil {
		return fakeQuery(s.query, fakeNamed(args))
	}
	return &fakeRows{}, nil
}

func (fakeTx) Commit() error   { return nil }
func (fakeTx) Rollback() error { return nil }

func (fakeResult) LastInsertId() (int64, error) { return 7, nil }
func (fakeResult) RowsAffected() (int64, error) { return 1, nil }

func (r *fakeRows) Columns() []string { return r.columns }
func (r *fakeRows) Close() error      { return nil }
func (r *fakeRows) Next(dest []driver.Value) error {
	if r.idx >= len(r.values) {
		return io.EOF
	}
	copy(dest, r.values[r.idx])
	r.idx++
	return nil
}

func resetFakeDB() { fakeExec, fakeQuery = nil, nil }

func TestMain(m *testing.M) {
	sql.Register("fakeorm", fakeDriver{})
	orm.RegisterDriver("fakeorm", orm.DRPostgres)
	_ = orm.RegisterDataBase("default", "fakeorm", "")
	os.Exit(m.Run())
}
