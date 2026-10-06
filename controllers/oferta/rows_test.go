package oferta

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
