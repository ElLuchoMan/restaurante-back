package login

import (
	"net/http"
	"strings"

	"restaurante/models"

	"github.com/beego/beego/v2/server/web/context"
)

// RolCliente es el valor del claim `rol` de los clientes.
const RolCliente = "Cliente"

// ClaimsFromContext devuelve los claims del access token de la petición sin
// escribir ninguna respuesta; nil si no hay token o es inválido. Sirve a los
// controladores de rutas públicas que ofrecen más datos a usuarios autenticados.
func ClaimsFromContext(ctx *context.Context) *Claims {
	h := ctx.Input.Header("Authorization")
	if h == "" {
		return nil
	}
	h = strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
	claims, err := ParseTokenClaims(h)
	if err != nil {
		return nil
	}
	return claims
}

// IsStaff indica si el rol es de trabajador (cualquiera distinto de Cliente).
func (c *Claims) IsStaff() bool { return c != nil && c.Rol != "" && c.Rol != RolCliente }

// IsAdmin indica si el rol es Administrador.
func (c *Claims) IsAdmin() bool { return c != nil && c.Rol == string(models.RolAdministrador) }

// ValidateStaff exige un access token de un trabajador (cualquier rol salvo
// Cliente): 401 sin token o inválido, 403 para clientes.
func ValidateStaff(ctx *context.Context) {
	if ctx.Input.Method() == http.MethodOptions {
		ctx.Output.Status = http.StatusOK
		return
	}
	claims := authenticate(ctx)
	if claims == nil {
		return
	}
	if !claims.IsStaff() {
		writeAuthError(ctx, http.StatusForbidden, "Se requiere un usuario trabajador")
	}
}
