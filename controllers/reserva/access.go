package reserva

import (
	"net/http"
	"strings"

	loginc "restaurante/controllers/login"

	"github.com/beego/beego/v2/server/web/context"
)

// rutasSoloPersonal son los GET de /reservas con datos personales de todos los
// contactos (listado completo, por contacto/fecha y por documento).
var rutasSoloPersonal = []string{"/reservas", "/reservas/parameter", "/reservas/documento"}

// AccessFilter es el filtro previo de /reservas. Beego aplica el filtro de un
// namespace a todo su prefijo, así que el reparto por método y ruta se hace
// aquí: los listados completos exigen token de trabajador (401 sin token, 403 a
// clientes) y el resto pasa por ValidateToken, que deja públicos solo el POST de
// creación y GET /reservas/consulta (y exige token en search, cliente, PUT y DELETE).
func AccessFilter(ctx *context.Context) {
	if ctx.Input.Method() == http.MethodGet && esRutaSoloPersonal(ctx.Input.URL()) {
		loginc.ValidateStaff(ctx)
		return
	}
	loginc.ValidateToken(ctx)
}

func esRutaSoloPersonal(path string) bool {
	path = strings.TrimRight(path, "/")
	for _, r := range rutasSoloPersonal {
		if strings.HasSuffix(path, r) {
			return true
		}
	}
	return false
}
