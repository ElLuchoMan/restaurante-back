package httpx

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/beego/beego/v2/server/web"
)

// PositiveInt64Param lee el query param key como entero positivo. Devuelve un
// error descriptivo (nunca nil con valor 0) si falta, no es entero o es <= 0.
func PositiveInt64Param(c *web.Controller, key string) (int64, error) {
	raw := strings.TrimSpace(c.GetString(key))
	if raw == "" {
		return 0, fmt.Errorf("el parámetro '%s' es obligatorio", key)
	}
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("el parámetro '%s' debe ser un entero: %w", key, err)
	}
	if v <= 0 {
		return 0, fmt.Errorf("el parámetro '%s' debe ser un entero positivo", key)
	}
	return v, nil
}

// IsPGConflict indica si err proviene de una violación de unicidad (23505) o de
// llave foránea (23503) de PostgreSQL; el controlador la traduce a HTTP 409.
func IsPGConflict(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "23505") || strings.Contains(msg, "23503") ||
		strings.Contains(msg, "duplicate key") || strings.Contains(msg, "violates foreign key")
}

// ErrNotFound es un centinela para flujos internos que terminan en HTTP 404.
var ErrNotFound = errors.New("no encontrado")
