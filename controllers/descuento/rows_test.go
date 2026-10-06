package descuento

import (
	"database/sql/driver"
	"regexp"
	"strings"
)

var colRe = regexp.MustCompile(`(?:T(\d+)\.)?"([^"]+)"`)

// rowFor arma una fila para la consulta q: una columna por cada campo del
// SELECT (en orden). Los valores salen de vals por "T<n>.<columna>" o por
// "<columna>"; el resto queda NULL. Para SELECT COUNT(*) usa vals["count"].
func rowFor(q string, vals map[string]driver.Value) driver.Rows {
	if strings.Contains(q, "COUNT(") {
		return rowsOf([]string{"count"}, []driver.Value{vals["count"]})
	}
	sel := q[len("SELECT "):strings.Index(q, " FROM ")]
	var cols []string
	var row []driver.Value
	for _, m := range colRe.FindAllStringSubmatch(sel, -1) {
		name := m[2]
		cols = append(cols, name)
		if v, ok := vals["T"+m[1]+"."+name]; ok && m[1] != "" {
			row = append(row, v)
		} else if v, ok := vals[name]; ok {
			row = append(row, v)
		} else {
			row = append(row, nil)
		}
	}
	return rowsOf(cols, row)
}

// queryFn despacha fakeQuery por fragmento de SQL: la primera clave contenida
// en la consulta decide la respuesta (nil devuelve un resultado vacío).
func queryFn(routes map[string]func(q string) (driver.Rows, error)) func(string, []driver.NamedValue) (driver.Rows, error) {
	return func(q string, _ []driver.NamedValue) (driver.Rows, error) {
		for frag, f := range routes {
			if strings.Contains(q, frag) {
				return f(q)
			}
		}
		return &fakeRows{}, nil
	}
}

var tableRe = regexp.MustCompile(`FROM "([a-z_]+)"`)

// db simula la base: counts[tabla] responde los COUNT(*), rows[tabla] las
// filas de los SELECT y errs[tabla] fuerza un error de consulta.
type db struct {
	counts map[string]int64
	rows   map[string][]map[string]driver.Value
	errs   map[string]error
	seen   []string
}

func newDB() *db {
	return &db{counts: map[string]int64{}, rows: map[string][]map[string]driver.Value{}, errs: map[string]error{}}
}

func (d *db) install() *db {
	fakeQuery = func(q string, _ []driver.NamedValue) (driver.Rows, error) {
		m := tableRe.FindStringSubmatch(q)
		table := ""
		if m != nil {
			table = m[1]
		}
		d.seen = append(d.seen, table)
		if err := d.errs[table]; err != nil {
			return nil, err
		}
		if strings.Contains(q, "COUNT(") {
			return rowFor(q, map[string]driver.Value{"count": d.counts[table]}), nil
		}
		rows := d.rows[table]
		if len(rows) == 0 {
			return &fakeRows{}, nil
		}
		first := rowFor(q, rows[0]).(*fakeRows)
		for _, r := range rows[1:] {
			first.values = append(first.values, rowFor(q, r).(*fakeRows).values[0])
		}
		return first, nil
	}
	return d
}
