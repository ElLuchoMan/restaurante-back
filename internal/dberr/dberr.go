// Package dberr clasifica errores de base de datos (PostgreSQL) para que los
// controladores respondan 409 (unicidad) o 400 (llave foránea inexistente) en
// lugar de 500.
package dberr

import (
	"errors"
	"strings"

	"github.com/lib/pq"
)

const (
	codeUnique     = "23505"
	codeForeignKey = "23503"
)

func hasCode(err error, code string) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return string(pqErr.Code) == code
	}
	return false
}

// IsUnique indica si err es una violación de restricción única.
func IsUnique(err error) bool {
	if err == nil {
		return false
	}
	if hasCode(err, codeUnique) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate key") || strings.Contains(msg, "unique constraint") || strings.Contains(msg, "23505")
}

// IsForeignKey indica si err es una violación de llave foránea.
func IsForeignKey(err error) bool {
	if err == nil {
		return false
	}
	if hasCode(err, codeForeignKey) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "foreign key") || strings.Contains(msg, "23503")
}

// Mentions indica (sin distinguir mayúsculas) si el mensaje de err contiene s.
// Sirve para saber qué columna/restricción originó una violación.
func Mentions(err error, s string) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), strings.ToLower(s))
}
