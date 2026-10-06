// Package authz reúne las comprobaciones de autorización a nivel de controlador
// para las rutas de cupones, ofertas y descuentos: exigir sesión o rol
// Administrador y resolver el cliente que actúa (siempre desde el token cuando
// quien llama es un Cliente).
package authz

import (
	"net/http"

	"restaurante/controllers/login"
	"restaurante/internal/httpx"

	"github.com/beego/beego/v2/server/web"
)

// RequireAuth exige un access token válido: responde 401 y devuelve false si
// no hay token o es inválido.
func RequireAuth(c *web.Controller) (*login.Claims, bool) {
	claims := login.ClaimsFromContext(c.Ctx)
	if claims == nil {
		httpx.Fail(c, http.StatusUnauthorized, "Token ausente o inválido", nil)
		return nil, false
	}
	return claims, true
}

// RequireAdmin exige un token de Administrador: 401 sin token válido y 403 si
// el rol no es Administrador.
func RequireAdmin(c *web.Controller) (*login.Claims, bool) {
	claims, ok := RequireAuth(c)
	if !ok {
		return nil, false
	}
	if !claims.IsAdmin() {
		httpx.Fail(c, http.StatusForbidden, "Se requiere rol Administrador", nil)
		return nil, false
	}
	return claims, true
}

// RequireStaff exige un token de trabajador (cualquier rol salvo Cliente): 401
// sin token válido y 403 si el rol es Cliente.
func RequireStaff(c *web.Controller) (*login.Claims, bool) {
	claims, ok := RequireAuth(c)
	if !ok {
		return nil, false
	}
	if !claims.IsStaff() {
		httpx.Fail(c, http.StatusForbidden, "Se requiere un usuario trabajador", nil)
		return nil, false
	}
	return claims, true
}

// EsDuenio indica si quien llama puede acceder a los datos del cliente con
// el documento dado: el personal siempre; un Cliente solo si el documento de su
// token coincide (un token sin documento positivo nunca es dueño de nada).
func EsDuenio(claims *login.Claims, documentoCliente int64) bool {
	if claims.IsStaff() {
		return true
	}
	return claims.Documento > 0 && claims.Documento == documentoCliente
}

// ResolveCliente determina el documento del cliente que actúa.
//
//   - Cliente: el documento del token manda; un clienteId distinto en el body
//     responde 403 (0 o ausente se ignora).
//   - Trabajador/Administrador: actúa en nombre de un cliente, por lo que
//     `clienteId` del body es obligatorio (400 si falta o no es positivo).
func ResolveCliente(c *web.Controller, claims *login.Claims, bodyClienteID int64) (int64, bool) {
	if claims.IsStaff() {
		if bodyClienteID <= 0 {
			httpx.Fail(c, http.StatusBadRequest, "clienteId (positivo) es obligatorio para trabajadores", nil)
			return 0, false
		}
		return bodyClienteID, true
	}
	if claims.Documento <= 0 {
		httpx.Fail(c, http.StatusForbidden, "El token no identifica a un cliente", nil)
		return 0, false
	}
	if bodyClienteID != 0 && bodyClienteID != claims.Documento {
		httpx.Fail(c, http.StatusForbidden, "No puede actuar en nombre de otro cliente", nil)
		return 0, false
	}
	return claims.Documento, true
}
