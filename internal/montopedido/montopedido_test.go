package montopedido

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/beego/beego/v2/client/orm"
)

// Driver SQL falso: cada consulta la responde fakeQuery según su texto.
var fakeQuery func(q string) (driver.Rows, error)

type fakeDriver struct{}
type fakeConn struct{}
type fakeStmt struct{ query string }
type fakeRows struct {
	vals []driver.Value
	done bool
}

func (fakeDriver) Open(string) (driver.Conn, error)         { return fakeConn{}, nil }
func (fakeConn) Prepare(q string) (driver.Stmt, error)      { return fakeStmt{q}, nil }
func (fakeConn) Close() error                               { return nil }
func (fakeConn) Begin() (driver.Tx, error)                  { return nil, errors.New("sin tx") }
func (fakeConn) Ping(context.Context) error                 { return nil }
func (fakeStmt) Close() error                               { return nil }
func (fakeStmt) NumInput() int                              { return -1 }
func (fakeStmt) Exec([]driver.Value) (driver.Result, error) { return nil, errors.New("sin exec") }
func (s fakeStmt) Query([]driver.Value) (driver.Rows, error) {
	if fakeQuery == nil { // consulta de zona horaria al registrar la base
		return &fakeRows{done: true}, nil
	}
	return fakeQuery(s.query)
}
func (r *fakeRows) Columns() []string {
	cols := make([]string, len(r.vals))
	for i := range cols {
		cols[i] = "c"
	}
	return cols
}
func (r *fakeRows) Close() error { return nil }
func (r *fakeRows) Next(dest []driver.Value) error {
	if r.done {
		return io.EOF
	}
	r.done = true
	copy(dest, r.vals)
	return nil
}

func TestMain(m *testing.M) {
	sql.Register("fakemonto", fakeDriver{})
	orm.RegisterDriver("fakemonto", orm.DRPostgres)
	_ = orm.RegisterDataBase("default", "fakemonto", "")
	os.Exit(m.Run())
}

func sirve(subtotal, lineas, descuento int64) {
	fakeQuery = func(q string) (driver.Rows, error) {
		if strings.Contains(q, "pedido_descuento_aplicado") {
			return &fakeRows{vals: []driver.Value{descuento}}, nil
		}
		return &fakeRows{vals: []driver.Value{subtotal, lineas}}, nil
	}
}

func TestCalcular(t *testing.T) {
	defer func() { fakeQuery = nil }()
	o := orm.NewOrm()

	sirve(50000, 2, 0)
	m, err := Calcular(o, 1)
	if err != nil || m != (Monto{Lineas: 2, Subtotal: 50000, Total: 50000}) {
		t.Fatalf("sin descuento: %+v, %v", m, err)
	}

	sirve(50000, 2, 8000)
	if m, _ = Calcular(o, 1); m.Total != 42000 || m.Descuento != 8000 {
		t.Fatalf("con descuento: %+v", m)
	}

	// el descuento nunca deja un total negativo
	sirve(5000, 1, 9000)
	if m, _ = Calcular(o, 1); m.Total != 0 {
		t.Fatalf("total negativo: %+v", m)
	}

	// pedido sin productos
	sirve(0, 0, 0)
	if m, _ = Calcular(o, 1); m.Lineas != 0 || m.Total != 0 {
		t.Fatalf("sin productos: %+v", m)
	}

	// errores de base de datos en cada consulta
	boom := errors.New("boom")
	fakeQuery = func(string) (driver.Rows, error) { return nil, boom }
	if _, err := Calcular(o, 1); !errors.Is(err, boom) {
		t.Fatalf("error del detalle: %v", err)
	}
	fakeQuery = func(q string) (driver.Rows, error) {
		if strings.Contains(q, "pedido_descuento_aplicado") {
			return nil, boom
		}
		return &fakeRows{vals: []driver.Value{int64(1), int64(1)}}, nil
	}
	if _, err := Calcular(o, 1); !errors.Is(err, boom) {
		t.Fatalf("error de descuentos: %v", err)
	}
}
