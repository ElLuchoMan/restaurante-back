package services

import (
	"database/sql"
	"database/sql/driver"
	"io"
	"os"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/beego/beego/v2/client/orm"
)

// Driver SQL falso registrado como base "default": permite ejercitar los
// servicios con el ORM real (SQL generado, transacciones) sin base de datos.
// fk.install programa las respuestas por tabla; los bloqueos de fila
// (SELECT ... FOR UPDATE) y los efectos transaccionales se emulan por conexión
// con fkConn.holdLock / fkConn.onCommit, de modo que la prueba de concurrencia
// reproduce la semántica de PostgreSQL (bloqueo hasta COMMIT/ROLLBACK).

type fkHandler func(c *fkConn, q string, args []driver.NamedValue) (driver.Rows, error)

type fakeDB struct {
	mu sync.Mutex
	// rows responde los SELECT de cada tabla; counts los SELECT COUNT(*).
	rows   map[string]func(c *fkConn, q string, args []driver.NamedValue) ([]map[string]driver.Value, error)
	counts map[string]func(c *fkConn, q string, args []driver.NamedValue) (int64, error)
	// insertErr / updateErr fuerzan error en INSERT / UPDATE de una tabla.
	insertErr map[string]error
	updateErr map[string]error
	// onInsert observa cada INSERT (tabla, conexión) antes de confirmarlo.
	onInsert  func(c *fkConn, table string, args []driver.NamedValue)
	beginErr  error
	commitErr error
	log       []string
	nextConn  int
}

var fk = &fakeDB{}

func (d *fakeDB) reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.rows = map[string]func(*fkConn, string, []driver.NamedValue) ([]map[string]driver.Value, error){}
	d.counts = map[string]func(*fkConn, string, []driver.NamedValue) (int64, error){}
	d.insertErr = map[string]error{}
	d.updateErr = map[string]error{}
	d.onInsert = nil
	d.beginErr, d.commitErr = nil, nil
	d.log = nil
}

// sawQuery devuelve las sentencias registradas (en orden) de una conexión.
func (d *fakeDB) statements() []string {
	d.mu.Lock()
	defer d.mu.Unlock()
	return append([]string(nil), d.log...)
}

type fkDriver struct{}

type fkConn struct {
	db       *fakeDB
	id       int
	inTx     bool
	releases []func()
	commits  []func()
}

type fkStmt struct {
	c     *fkConn
	query string
}

type fkTx struct{ c *fkConn }

func (fkDriver) Open(string) (driver.Conn, error) {
	fk.mu.Lock()
	defer fk.mu.Unlock()
	fk.nextConn++
	return &fkConn{db: fk, id: fk.nextConn}, nil
}

func (c *fkConn) Prepare(q string) (driver.Stmt, error) { return &fkStmt{c: c, query: q}, nil }
func (c *fkConn) Close() error                          { return nil }
func (c *fkConn) Begin() (driver.Tx, error) {
	if c.db.beginErr != nil {
		return nil, c.db.beginErr
	}
	c.inTx = true
	return fkTx{c: c}, nil
}

// holdLock toma mu y lo conserva hasta que termine la transacción de la conexión.
func (c *fkConn) holdLock(mu *sync.Mutex) {
	mu.Lock()
	c.releases = append(c.releases, mu.Unlock)
}

// onCommit difiere f hasta el COMMIT (se descarta en ROLLBACK); fuera de una
// transacción se ejecuta de inmediato.
func (c *fkConn) onCommit(f func()) {
	if !c.inTx {
		f()
		return
	}
	c.commits = append(c.commits, f)
}

func (c *fkConn) finish(commit bool) {
	if commit {
		for _, f := range c.commits {
			f()
		}
	}
	for i := len(c.releases) - 1; i >= 0; i-- {
		c.releases[i]()
	}
	c.commits, c.releases, c.inTx = nil, nil, false
}

func (t fkTx) Commit() error {
	if err := t.c.db.commitErr; err != nil {
		t.c.finish(false)
		return err
	}
	t.c.finish(true)
	return nil
}

func (t fkTx) Rollback() error { t.c.finish(false); return nil }

func (s *fkStmt) Close() error  { return nil }
func (s *fkStmt) NumInput() int { return -1 }

func fkNamed(args []driver.Value) []driver.NamedValue {
	named := make([]driver.NamedValue, len(args))
	for i, v := range args {
		named[i] = driver.NamedValue{Ordinal: i + 1, Value: v}
	}
	return named
}

var (
	fkTableRe = regexp.MustCompile(`(?:FROM|INSERT INTO|UPDATE) "([a-z_]+)"`)
	fkColRe   = regexp.MustCompile(`(?:T(\d+)\.)?"([^"]+)"`)
)

func fkTable(q string) string {
	if m := fkTableRe.FindStringSubmatch(q); m != nil {
		return m[1]
	}
	return ""
}

func (s *fkStmt) record() {
	s.c.db.mu.Lock()
	s.c.db.log = append(s.c.db.log, s.query)
	s.c.db.mu.Unlock()
}

func (s *fkStmt) write(args []driver.Value) error {
	table := fkTable(s.query)
	if strings.HasPrefix(s.query, "UPDATE") {
		return s.c.db.updateErr[table]
	}
	if err := s.c.db.insertErr[table]; err != nil {
		return err
	}
	if h := s.c.db.onInsert; h != nil {
		h(s.c, table, fkNamed(args))
	}
	return nil
}

func (s *fkStmt) Exec(args []driver.Value) (driver.Result, error) {
	s.record()
	if err := s.write(args); err != nil {
		return nil, err
	}
	return fkResult{}, nil
}

func (s *fkStmt) Query(args []driver.Value) (driver.Rows, error) {
	s.record()
	named := fkNamed(args)
	table := fkTable(s.query)
	// INSERT ... RETURNING del ORM en PostgreSQL: devuelve el id generado.
	if strings.HasPrefix(s.query, "INSERT") {
		if err := s.write(args); err != nil {
			return nil, err
		}
		return &fkRows{columns: []string{"id"}, values: [][]driver.Value{{int64(7)}}}, nil
	}
	if strings.Contains(s.query, "COUNT(") {
		var n int64
		if h := s.c.db.counts[table]; h != nil {
			var err error
			if n, err = h(s.c, s.query, named); err != nil {
				return nil, err
			}
		}
		return &fkRows{columns: []string{"count"}, values: [][]driver.Value{{n}}}, nil
	}
	h := s.c.db.rows[table]
	if h == nil {
		return &fkRows{}, nil
	}
	rows, err := h(s.c, s.query, named)
	if err != nil {
		return nil, err
	}
	return fkRowsFor(s.query, rows), nil
}

// fkRowsFor arma un resultado con una columna por cada campo del SELECT (en
// orden); los valores salen de cada mapa por nombre de columna y el resto es NULL.
func fkRowsFor(q string, rows []map[string]driver.Value) driver.Rows {
	sel := q[len("SELECT "):strings.Index(q, " FROM ")]
	var cols []string
	for _, m := range fkColRe.FindAllStringSubmatch(sel, -1) {
		cols = append(cols, m[2])
	}
	out := &fkRows{columns: cols}
	for _, r := range rows {
		row := make([]driver.Value, len(cols))
		for i, c := range cols {
			row[i] = r[c]
		}
		out.values = append(out.values, row)
	}
	return out
}

type fkResult struct{}

func (fkResult) LastInsertId() (int64, error) { return 7, nil }
func (fkResult) RowsAffected() (int64, error) { return 1, nil }

type fkRows struct {
	columns []string
	values  [][]driver.Value
	idx     int
}

func (r *fkRows) Columns() []string { return r.columns }
func (r *fkRows) Close() error      { return nil }
func (r *fkRows) Next(dest []driver.Value) error {
	if r.idx >= len(r.values) {
		return io.EOF
	}
	copy(dest, r.values[r.idx])
	r.idx++
	return nil
}

func TestMain(m *testing.M) {
	sql.Register("fakeorm-services", fkDriver{})
	orm.RegisterDriver("fakeorm-services", orm.DRPostgres)
	_ = orm.RegisterDataBase("default", "fakeorm-services", "")
	fk.reset()
	os.Exit(m.Run())
}
