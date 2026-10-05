// Package httpx reúne utilidades comunes de los controladores: respuestas con
// código HTTP real, listas siempre serializadas como [] y decodificación
// "merge" para PUT (los campos ausentes se conservan, null limpia el campo).
package httpx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"restaurante/models"

	"github.com/beego/beego/v2/server/web"
)

// Send fija el status HTTP y envía una ApiResponse cuyo `code` coincide con él.
func Send(c *web.Controller, status int, message string, data any) {
	c.Ctx.Output.SetStatus(status)
	c.Data["json"] = models.ApiResponse{Code: status, Message: message, Data: data}
	_ = c.ServeJSON()
}

// Fail envía un error con el status indicado; cause (opcional) es err.Error().
func Fail(c *web.Controller, status int, message string, err error) {
	resp := models.ApiResponse{Code: status, Message: message}
	if err != nil {
		resp.Cause = err.Error()
	}
	c.Ctx.Output.SetStatus(status)
	c.Data["json"] = resp
	_ = c.ServeJSON()
}

// List devuelve s o, si es nil, un slice vacío para que se serialice como [].
func List[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}

// DecodeMerge aplica el JSON de body sobre dst conservando los campos
// ausentes. Un campo enviado explícitamente como null solo se admite si su
// nombre json está en nullable (se limpia); en otro caso devuelve error.
func DecodeMerge(body []byte, dst any, nullable ...string) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return err
	}
	allowed := make(map[string]bool, len(nullable))
	for _, n := range nullable {
		allowed[n] = true
	}
	var bad []string
	for k, v := range raw {
		if bytes.Equal(bytes.TrimSpace(v), []byte("null")) && !allowed[k] {
			bad = append(bad, k)
		}
	}
	if len(bad) > 0 {
		sort.Strings(bad)
		return fmt.Errorf("los campos no admiten null: %s", strings.Join(bad, ", "))
	}
	return json.Unmarshal(body, dst)
}

// Present devuelve el conjunto de claves json presentes en body (null incluido).
func Present(body []byte) (map[string]bool, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}
	out := make(map[string]bool, len(raw))
	for k := range raw {
		out[k] = true
	}
	return out, nil
}

// IsNull indica si la clave vino en body con valor null explícito.
func IsNull(body []byte, key string) bool {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return false
	}
	v, ok := raw[key]
	return ok && bytes.Equal(bytes.TrimSpace(v), []byte("null"))
}
