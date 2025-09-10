package categoria

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
	"testing"

	"github.com/beego/beego/v2/client/orm"
	"restaurante/models"
)

// stub implementations to exercise adapter wrappers

type stubQS struct{ orm.QuerySeter }

func (stubQS) All(res interface{}, cols ...string) (int64, error) {
	dst := res.(*[]int)
	*dst = append(*dst, 1, 2)
	return 2, nil
}

type stubOrm struct{ orm.Ormer }

func (stubOrm) QueryTable(i interface{}) orm.QuerySeter             { return stubQS{} }
func (stubOrm) Insert(v interface{}) (int64, error)                 { *v.(*int) = 1; return 1, nil }
func (stubOrm) Read(v interface{}, cols ...string) error            { *v.(*int) = 2; return nil }
func (stubOrm) Update(v interface{}, cols ...string) (int64, error) { *v.(*int) = 3; return 1, nil }
func (stubOrm) Delete(v interface{}, cols ...string) (int64, error) { return 1, nil }

func TestCatOrmAdapterAndQSAdapter(t *testing.T) {
	a := catOrmAdapter{o: stubOrm{}}
	qs := a.QueryTable(new(models.Categoria))
	var list []int
	if _, err := qs.All(&list); err != nil || len(list) != 2 {
		t.Fatalf("expected 2 items, got %v, err %v", list, err)
	}
	var n int
	if _, err := a.Insert(&n); err != nil || n != 1 {
		t.Fatalf("insert: n=%d err=%v", n, err)
	}
	if err := a.Read(&n); err != nil || n != 2 {
		t.Fatalf("read: n=%d err=%v", n, err)
	}
	if _, err := a.Update(&n); err != nil || n != 3 {
		t.Fatalf("update: n=%d err=%v", n, err)
	}
	if _, err := a.Delete(&n); err != nil {
		t.Fatalf("delete: err=%v", err)
	}
}

// Minimal mock driver to allow calling catOrmNew without hitting a real DB

type dummyDriver struct{}

type dummyConn struct{}

type dummyStmt struct{}

type dummyTx struct{}

type dummyResult struct{}

type dummyRows struct{}

func (dummyDriver) Open(string) (driver.Conn, error) { return dummyConn{}, nil }

func (dummyConn) Prepare(string) (driver.Stmt, error) { return dummyStmt{}, nil }
func (dummyConn) Close() error                        { return nil }
func (dummyConn) Begin() (driver.Tx, error)           { return dummyTx{}, nil }
func (dummyConn) ExecContext(context.Context, string, []driver.NamedValue) (driver.Result, error) {
	return dummyResult{}, nil
}
func (dummyConn) QueryContext(context.Context, string, []driver.NamedValue) (driver.Rows, error) {
	return dummyRows{}, nil
}
func (dummyConn) Ping(context.Context) error { return nil }

func (dummyStmt) Close() error                               { return nil }
func (dummyStmt) NumInput() int                              { return -1 }
func (dummyStmt) Exec([]driver.Value) (driver.Result, error) { return dummyResult{}, nil }
func (dummyStmt) Query([]driver.Value) (driver.Rows, error)  { return dummyRows{}, nil }

func (dummyTx) Commit() error   { return nil }
func (dummyTx) Rollback() error { return nil }

func (dummyResult) LastInsertId() (int64, error) { return 0, nil }
func (dummyResult) RowsAffected() (int64, error) { return 0, nil }

func (dummyRows) Columns() []string         { return nil }
func (dummyRows) Close() error              { return nil }
func (dummyRows) Next([]driver.Value) error { return io.EOF }

func TestCatOrmNew(t *testing.T) {
	sql.Register("dummy", dummyDriver{})
	orm.RegisterDriver("dummy", orm.DRPostgres)
	if err := orm.RegisterDataBase("default", "dummy", ""); err != nil {
		t.Fatalf("register db: %v", err)
	}
	if ormer := catOrmNew(); ormer == nil {
		t.Fatalf("expected ormer from catOrmNew")
	}
}
