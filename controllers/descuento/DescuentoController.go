package descuento

import (
	"encoding/json"
	"errors"
	"net/http"

	"restaurante/internal/authz"
	"restaurante/internal/httpx"
	"restaurante/logging"
	"restaurante/models"
	"restaurante/services"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

// DescuentoController gestiona /descuentos/pedidos. En las respuestas
// `pedidoId`, `cuponId` y `ofertaId` son los objetos relacionados (sin
// contraseñas); `cuponId` y `ofertaId` se omiten si no aplican.
type DescuentoController struct {
	web.Controller
}

const msgPedidoInvalido = "El parámetro 'pedido_id' es inválido o está ausente"

// @Title GetAll
// @Summary Obtener descuentos de un pedido
// @Description Cualquier usuario autenticado. Un Cliente solo puede consultar descuentos de sus propios pedidos (403 si el pedido es de otro cliente); un trabajador o administrador puede consultar cualquiera. Lista los descuentos aplicados al pedido. Si el pedido existe pero no tiene descuentos, `data` es una lista vacía `[]`; si no existe responde 404.
// @Tags descuentos
// @Accept json
// @Produce json
// @Param pedido_id query int true "ID del pedido (entero positivo)"
// @Success 200 {object} models.ApiResponse{data=[]models.PedidoDescuentoDoc} "Descuentos del pedido (puede ser vacío)"
// @Failure 400 {object} models.ApiResponse "pedido_id inválido o ausente"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 403 {object} models.ApiResponse "El pedido es de otro cliente"
// @Failure 404 {object} models.ApiResponse "Pedido no encontrado"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /descuentos/pedidos [get]
func (c *DescuentoController) GetAll() {
	claims, ok := authz.RequireAuth(&c.Controller)
	if !ok {
		return
	}
	pedidoID, err := httpx.PositiveInt64Param(&c.Controller, "pedido_id")
	if err != nil {
		logging.LogControllerError(c.Ctx, "descuentos.getall.bad_request", err, map[string]interface{}{"pedido_id": c.GetString("pedido_id")})
		httpx.Fail(&c.Controller, http.StatusBadRequest, msgPedidoInvalido, err)
		return
	}

	o := orm.NewOrm()
	pedido := &models.Pedido{PK_ID_PEDIDO: pedidoID}
	if err := o.Read(pedido); err != nil {
		if errors.Is(err, orm.ErrNoRows) {
			httpx.Fail(&c.Controller, http.StatusNotFound, "Pedido no encontrado", nil)
			return
		}
		logging.LogControllerError(c.Ctx, "descuentos.getall.read_error", err, map[string]interface{}{"pedido_id": pedidoID})
		httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error interno del servidor", err)
		return
	}
	if !claims.IsStaff() && (pedido.PK_DOCUMENTO_CLIENTE == nil || pedido.PK_DOCUMENTO_CLIENTE.PK_DOCUMENTO_CLIENTE != claims.Documento) {
		httpx.Fail(&c.Controller, http.StatusForbidden, "El pedido no pertenece al cliente", nil)
		return
	}

	descuentos, err := services.NewDescuentoService(o).ObtenerDescuentosPedido(c.Ctx.Request.Context(), pedidoID)
	if err != nil {
		logging.LogControllerError(c.Ctx, "descuentos.getall.service_error", err, map[string]interface{}{"pedido_id": pedidoID})
		httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error al obtener descuentos", err)
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, "Descuentos obtenidos exitosamente", httpx.List(descuentos))
}

// @Title Post
// @Summary Aplicar descuento a pedido
// @Description Cualquier usuario autenticado. Aplica un único descuento (cupón u oferta, exactamente uno) a un pedido y recalcula su total, todo en UNA transacción: el servidor valida (cupón: activo, vigencia, topes `maxUsos`/`limitePorCliente` bajo `SELECT ... FOR UPDATE`, monto mínimo, scope; oferta: activa, período, día, horario, restaurante y productos del pedido), CALCULA el monto con el detalle del pedido (el cliente nunca envía el monto), redime el cupón, registra el descuento y resta el monto del `monto` del pago del pedido (si tiene pago). Si algo falla no queda nada a medias. Un pedido admite un solo descuento y no debe estar cancelado, terminado ni pagado. El cliente SALE DEL TOKEN: un Cliente no envía `clienteId` (si lo envía y no coincide con su documento responde 403); un trabajador o administrador actúa en nombre de un cliente y debe indicar `clienteId`. El pedido debe pertenecer a ese cliente (403 si es de otro, 404 si no existe). `detalle` es opcional y debe ser un objeto JSON; se conserva y se le agregan los datos del cupón/oferta (`tipo`, `codigo`/`titulo`, `scope`), que prevalecen. La respuesta trae el descuento registrado y los importes: `subtotal` (suma del detalle), `montoDescuento` y `total` (lo que debe pagarse; si hay pago, es su nuevo `monto`, y `pagoId` lo identifica). Errores: 400 (pedido_id/ids/JSON inválidos o `clienteId` ausente para un trabajador), 403, 404 (pedido, cupón u oferta inexistente), 409 (el pedido ya tiene un descuento, ya está pagado, cancelado o terminado; cupón agotado, límite por cliente alcanzado o ya redimido en el pedido), 422 (no se indicó exactamente uno de cupón u oferta, detalle que no es objeto, cupón u oferta no aplicable).
// @Tags descuentos
// @Accept json
// @Produce json
// @Param pedido_id query int true "ID del pedido (entero positivo)"
// @Param body body models.AplicarDescuentoRequest true "Cupón u oferta a aplicar (sin monto: lo calcula el servidor)"
// @Success 201 {object} models.ApiResponse{data=models.DescuentoAplicadoDoc} "Descuento aplicado y total recalculado"
// @Failure 400 {object} models.ApiResponse "pedido_id, ids o JSON inválidos, o clienteId ausente para un trabajador"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 403 {object} models.ApiResponse "clienteId distinto del token o pedido de otro cliente"
// @Failure 404 {object} models.ApiResponse "Pedido, cupón u oferta no encontrado"
// @Failure 409 {object} models.ApiResponse "El pedido ya tiene descuento, ya está pagado/cerrado, o el cupón está agotado o ya redimido"
// @Failure 422 {object} models.ApiResponse "Solicitud inválida o cupón/oferta no aplicable"
// @Failure 500 {object} models.ApiResponse "Error al aplicar el descuento"
// @Security BearerAuth
// @Router /descuentos/pedidos [post]
func (c *DescuentoController) Post() {
	claims, ok := authz.RequireAuth(&c.Controller)
	if !ok {
		return
	}
	pedidoID, err := httpx.PositiveInt64Param(&c.Controller, "pedido_id")
	if err != nil {
		logging.LogControllerError(c.Ctx, "descuentos.post.bad_request", err, map[string]interface{}{"pedido_id": c.GetString("pedido_id")})
		httpx.Fail(&c.Controller, http.StatusBadRequest, msgPedidoInvalido, err)
		return
	}

	var req models.AplicarDescuentoRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		logging.LogControllerError(c.Ctx, "descuentos.post.bad_json", err, map[string]interface{}{"pedido_id": pedidoID})
		httpx.Fail(&c.Controller, http.StatusBadRequest, "JSON inválido", err)
		return
	}
	clienteID, ok := authz.ResolveCliente(&c.Controller, claims, req.ClienteId)
	if !ok {
		return
	}
	if (req.PkIdCupon != nil && *req.PkIdCupon <= 0) || (req.PkIdOferta != nil && *req.PkIdOferta <= 0) {
		httpx.Fail(&c.Controller, http.StatusBadRequest, "cuponId y ofertaId deben ser enteros positivos", nil)
		return
	}

	aplicado, err := services.NewDescuentoService(orm.NewOrm()).AplicarDescuento(c.Ctx.Request.Context(), pedidoID, clienteID, &req)
	if err != nil {
		c.aplicarError(pedidoID, err)
		return
	}
	httpx.Send(&c.Controller, http.StatusCreated, "Descuento aplicado exitosamente", aplicado)
}

// aplicarError traduce los errores tipados del servicio a su código HTTP.
func (c *DescuentoController) aplicarError(pedidoID int64, err error) {
	logging.LogControllerError(c.Ctx, "descuentos.post.service_error", err, map[string]interface{}{"pedido_id": pedidoID})
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, services.ErrPedidoNoEncontrado), errors.Is(err, services.ErrCuponNoEncontrado), errors.Is(err, services.ErrOfertaNoEncontrada):
		status = http.StatusNotFound
	case errors.Is(err, services.ErrPedidoAjeno):
		status = http.StatusForbidden
	case errors.Is(err, services.ErrDescuentoYaAplicado), errors.Is(err, services.ErrPedidoPagado), errors.Is(err, services.ErrPedidoCerrado), errors.Is(err, services.ErrCuponConflicto):
		status = http.StatusConflict
	case errors.Is(err, services.ErrDescuentoInvalido), errors.Is(err, services.ErrCuponNoAplicable), errors.Is(err, services.ErrOfertaNoAplicable):
		status = http.StatusUnprocessableEntity
	}
	httpx.Fail(&c.Controller, status, "Error al aplicar descuento", err)
}
