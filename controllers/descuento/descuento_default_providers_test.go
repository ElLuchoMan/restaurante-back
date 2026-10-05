//go:build !unit

package descuento

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Driver SQL mínimo: responde la detección de zona horaria al registrar la BD
// y falla toda consulta posterior con un error controlado.

var errFakeQuery = errors.New("fake query failure")

type fakeDriver struct{}
type fakeConn struct{}
type fakeStmt struct{ query string }
type fakeTx struct{}
type fakeRows struct{ done bool }

func (fakeDriver) Open(string) (driver.Conn, error)    { return fakeConn{}, nil }
func (fakeConn) Prepare(q string) (driver.Stmt, error) { return fakeStmt{query: q}, nil }
func (fakeConn) Close() error                          { return nil }
func (fakeConn) Begin() (driver.Tx, error)             { return fakeTx{}, nil }
func (fakeConn) Ping(context.Context) error            { return nil }
func (fakeStmt) Close() error                          { return nil }
func (fakeStmt) NumInput() int                         { return -1 }
func (fakeStmt) Exec([]driver.Value) (driver.Result, error) {
	return nil, errFakeQuery
}
func (s fakeStmt) Query([]driver.Value) (driver.Rows, error) {
	if strings.Contains(strings.ToUpper(s.query), " FROM ") {
		return nil, errFakeQuery
	}
	return &fakeRows{}, nil
}
func (fakeTx) Commit() error          { return nil }
func (fakeTx) Rollback() error        { return nil }
func (r *fakeRows) Columns() []string { return []string{"v"} }
func (r *fakeRows) Close() error      { return nil }
func (r *fakeRows) Next(dest []driver.Value) error {
	if r.done {
		return io.EOF
	}
	r.done = true
	dest[0] = "UTC"
	return nil
}

func TestMain(m *testing.M) {
	sql.Register("descuentofake", fakeDriver{})
	orm.RegisterDriver("descuentofake", orm.DRPostgres)
	_ = orm.RegisterDataBase("default", "descuentofake", "")
	os.Exit(m.Run())
}

func TestDefaultOrmReadProviderReadsThroughRealOrm(t *testing.T) {
	read := defaultOrmReadProvider()
	require.NotNil(t, read)

	err := read(&models.Cupon{PkIdCupon: 1})
	require.Error(t, err)
	assert.ErrorIs(t, err, errFakeQuery)
}

func TestDefaultOrmProviderReturnsRealOrm(t *testing.T) {
	o := defaultOrmProvider()
	require.NotNil(t, o)

	err := o.Read(&models.Cupon{PkIdCupon: 1})
	require.Error(t, err)
	assert.ErrorIs(t, err, errFakeQuery)
}
