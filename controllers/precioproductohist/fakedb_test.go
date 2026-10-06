package precioproductohist

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	beecontext "github.com/beego/beego/v2/server/web/context"
)

// Driver SQL falso registrado como base "default": permite usar orm.NewOrm()
// sin base de datos real. Los tests programan fakeExec/fakeQuery.
var (
	fakeExec      func(query string, args []driver.NamedValue) (driver.Result, error)
	fakeQuery     func(query string, args []driver.NamedValue) (driver.Rows, error)
	fakeBeginErr  error
	fakeCommitErr error
	fakeAffected  int64 = 1
	fakeAffErr    error
	fakeLastID    int64 = 7
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
func (c *fakeConn) Begin() (driver.Tx, error) {
	if fakeBeginErr != nil {
		return nil, fakeBeginErr
	}
	return fakeTx{}, nil
}
func (c *fakeConn) Ping(context.Context) error { return nil }

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
	// En PostgreSQL el ORM inserta con INSERT ... RETURNING (pasa por Query):
	// se trata como un Exec que devuelve el id generado.
	if strings.HasPrefix(s.query, "INSERT") && strings.Contains(s.query, "RETURNING") {
		if fakeExec != nil {
			if _, err := fakeExec(s.query, fakeNamed(args)); err != nil {
				return nil, err
			}
		}
		return &fakeRows{columns: []string{"id"}, values: [][]driver.Value{{fakeLastID}}}, nil
	}
	if fakeQuery != nil {
		return fakeQuery(s.query, fakeNamed(args))
	}
	return &fakeRows{}, nil
}

func (fakeTx) Commit() error   { return fakeCommitErr }
func (fakeTx) Rollback() error { return nil }

func (fakeResult) LastInsertId() (int64, error) { return fakeLastID, nil }
func (fakeResult) RowsAffected() (int64, error) { return fakeAffected, fakeAffErr }

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

// rowsOf arma un resultado posicional (las columnas se leen por orden).
func rowsOf(cols []string, vals ...[]driver.Value) driver.Rows {
	return &fakeRows{columns: cols, values: vals}
}

func resetFake() {
	fakeExec, fakeQuery = nil, nil
	fakeBeginErr, fakeCommitErr, fakeAffErr = nil, nil, nil
	fakeAffected, fakeLastID = 1, 7
}

func TestMain(m *testing.M) {
	sql.Register("fakeorm", fakeDriver{})
	orm.RegisterDriver("fakeorm", orm.DRPostgres)
	_ = orm.RegisterDataBase("default", "fakeorm", "")
	os.Exit(m.Run())
}

// newCtx crea un contexto Beego con la petición dada.
func newCtx(method, target, body string) (*beecontext.Context, *httptest.ResponseRecorder) {
	r := httptest.NewRequest(method, target, strings.NewReader(body))
	w := httptest.NewRecorder()
	ctx := beecontext.NewContext()
	ctx.Reset(w, r)
	ctx.Input.RequestBody = []byte(body)
	return ctx, w
}

// decode lee el envoltorio ApiResponse y comprueba que el status HTTP
// coincida con `code` del cuerpo.
func decode(t *testing.T, w *httptest.ResponseRecorder) models.ApiResponse {
	t.Helper()
	var resp models.ApiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("respuesta no es JSON: %v (%s)", err, w.Body.String())
	}
	if resp.Code != w.Code {
		t.Fatalf("status HTTP %d distinto de code %d: %s", w.Code, resp.Code, w.Body.String())
	}
	return resp
}

// expect valida status y, si pasa, devuelve la respuesta.
func expect(t *testing.T, w *httptest.ResponseRecorder, status int) models.ApiResponse {
	t.Helper()
	if w.Code != status {
		t.Fatalf("esperaba HTTP %d, obtuve %d: %s", status, w.Code, w.Body.String())
	}
	return decode(t, w)
}

var _ = http.StatusOK
