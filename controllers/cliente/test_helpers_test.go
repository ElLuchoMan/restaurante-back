package cliente

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

func (c *mockConn) Ping(ctx context.Context) error { return nil }

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
	copy(dest, row)
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

func TestMockDriver(t *testing.T) {
	execCalled, queryCalled := false, false
	MockExec = func(ctx context.Context, q string, args []driver.NamedValue) (driver.Result, error) {
		execCalled = true
		return mockResult{}, nil
	}
	MockQuery = func(ctx context.Context, q string, args []driver.NamedValue) (driver.Rows, error) {
		queryCalled = true
		return &mockRows{columns: []string{"col"}, values: [][]driver.Value{{1}}}, nil
	}
	t.Cleanup(func() { MockExec, MockQuery = nil, nil })

	conn := &mockConn{}
	if _, err := conn.ExecContext(context.Background(), "UPDATE", nil); err != nil {
		t.Fatalf("ExecContext failed: %v", err)
	}
	rows, err := conn.QueryContext(context.Background(), "SELECT", nil)
	if err != nil {
		t.Fatalf("QueryContext failed: %v", err)
	}
	dest := make([]driver.Value, 1)
	if err := rows.Next(dest); err != nil {
		t.Fatalf("Next failed: %v", err)
	}
	if dest[0] != 1 {
		t.Fatalf("unexpected value %v", dest[0])
	}
	if len(rows.Columns()) != 1 {
		t.Fatalf("expected one column")
	}
	if err := rows.Close(); err != nil {
		t.Fatalf("close failed: %v", err)
	}

	stmt, err := conn.Prepare("INSERT")
	if err != nil {
		t.Fatalf("prepare failed: %v", err)
	}
	if _, err = stmt.Exec([]driver.Value{1}); err != nil {
		t.Fatalf("stmt Exec failed: %v", err)
	}
	if _, err = stmt.Query([]driver.Value{1}); err != nil {
		t.Fatalf("stmt Query failed: %v", err)
	}
	if err = stmt.Close(); err != nil {
		t.Fatalf("stmt Close failed: %v", err)
	}

	tx, err := conn.Begin()
	if err != nil {
		t.Fatalf("begin failed: %v", err)
	}
	if err = tx.Commit(); err != nil {
		t.Fatalf("commit failed: %v", err)
	}
	if err = tx.Rollback(); err != nil {
		t.Fatalf("rollback failed: %v", err)
	}

	res := mockResult{}
	if id, _ := res.LastInsertId(); id != 1 {
		t.Fatalf("unexpected id %d", id)
	}
	if aff, _ := res.RowsAffected(); aff != 1 {
		t.Fatalf("unexpected rows %d", aff)
	}
	if !execCalled || !queryCalled {
		t.Fatalf("callbacks not invoked")
	}
}

func TestMockDriverNoCallbacks(t *testing.T) {
	conn := &mockConn{}

	if err := conn.Ping(context.Background()); err != nil {
		t.Fatalf("Ping failed: %v", err)
	}

	res, err := conn.ExecContext(context.Background(), "UPDATE", nil)
	if err != nil {
		t.Fatalf("ExecContext failed: %v", err)
	}
	if _, err = res.RowsAffected(); err != nil {
		t.Fatalf("RowsAffected failed: %v", err)
	}
	if _, err = res.LastInsertId(); err != nil {
		t.Fatalf("LastInsertId failed: %v", err)
	}

	rows, err := conn.QueryContext(context.Background(), "SELECT", nil)
	if err != nil {
		t.Fatalf("QueryContext failed: %v", err)
	}
	if len(rows.Columns()) != 0 {
		t.Fatalf("expected no columns")
	}
	dest := make([]driver.Value, 1)
	if err = rows.Next(dest); err != io.EOF {
		t.Fatalf("expected EOF, got %v", err)
	}
	if err = rows.Close(); err != nil {
		t.Fatalf("rows Close failed: %v", err)
	}

	stmt, err := conn.Prepare("INSERT")
	if err != nil {
		t.Fatalf("Prepare failed: %v", err)
	}
	if _, err = stmt.Exec([]driver.Value{1}); err != nil {
		t.Fatalf("stmt Exec failed: %v", err)
	}
	srows, err := stmt.Query([]driver.Value{1})
	if err != nil {
		t.Fatalf("stmt Query failed: %v", err)
	}
	dest2 := make([]driver.Value, 1)
	if err = srows.Next(dest2); err != io.EOF {
		t.Fatalf("expected EOF from stmt rows: %v", err)
	}
}
