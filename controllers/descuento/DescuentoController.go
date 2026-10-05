package descuento

import (
	"encoding/json"
	"errors"
	"net/http"

	"restaurante/internal/dberr"
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
// @Description Lista los descuentos aplicados al pedido. Si el pedido existe pero no tiene descuentos, `data` es una lista vacía `[]`; si no existe responde 404.
// @Tags descuentos
// @Accept json
// @Produce json
// @Param pedido_id query int true "ID del pedido (entero positivo)"
// @Success 200 {object} models.ApiResponse{data=[]models.PedidoDescuentoDoc} "Descuentos del pedido (puede ser vacío)"
// @Failure 400 {object} models.ApiResponse "pedido_id inválido o ausente"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 404 {object} models.ApiResponse "Pedido no encontrado"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /descuentos/pedidos [get]
func (c *DescuentoController) GetAll() {
	pedidoID, err := httpx.PositiveInt64Param(&c.Controller, "pedido_id")
	if err != nil {
		logging.LogControllerError(c.Ctx, "descuentos.getall.bad_request", err, map[string]interface{}{"pedido_id": c.GetString("pedido_id")})
		httpx.Fail(&c.Controller, http.StatusBadRequest, msgPedidoInvalido, err)
		return
	}

	o := orm.NewOrm()
	if err := o.Read(&models.Pedido{PK_ID_PEDIDO: pedidoID}); err != nil {
		if errors.Is(err, orm.ErrNoRows) {
			httpx.Fail(&c.Controller, http.StatusNotFound, "Pedido no encontrado", nil)
			return
		}
		logging.LogControllerError(c.Ctx, "descuentos.getall.read_error", err, map[string]interface{}{"pedido_id": pedidoID})
		httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error interno del servidor", err)
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
// @Description Registra un único descuento (cupón u oferta, exactamente uno) sobre un pedido; un pedido admite un solo descuento. No redime el cupón (para eso use POST /cupones/{codigo}/redimir) ni recalcula el monto: `montoDescuento` (>= 0) lo informa el cliente. `detalle` es opcional y debe ser un objeto JSON; se conserva y se le agregan los datos del cupón/oferta (`tipo`, `codigo`/`titulo`, `scope`), que prevalecen. Errores: 400 (pedido_id/ids/JSON inválidos), 404 (pedido, cupón u oferta inexistente), 409 (el pedido ya tiene un descuento), 422 (no se indicó exactamente uno de cupón u oferta, monto negativo o detalle que no es objeto).
// @Tags descuentos
// @Accept json
// @Produce json
// @Param pedido_id query int true "ID del pedido (entero positivo)"
// @Param body body models.AplicarDescuentoRequest true "Datos del descuento a aplicar"
// @Success 201 {object} models.ApiResponse{data=models.PedidoDescuentoDoc} "Descuento aplicado"
// @Failure 400 {object} models.ApiResponse "pedido_id, ids o JSON inválidos"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 404 {object} models.ApiResponse "Pedido, cupón u oferta no encontrado"
// @Failure 409 {object} models.ApiResponse "El pedido ya tiene un descuento aplicado"
// @Failure 422 {object} models.ApiResponse "Solicitud de descuento inválida"
// @Failure 500 {object} models.ApiResponse "Error al aplicar el descuento"
// @Security BearerAuth
// @Router /descuentos/pedidos [post]
func (c *DescuentoController) Post() {
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
	if (req.PkIdCupon != nil && *req.PkIdCupon <= 0) || (req.PkIdOferta != nil && *req.PkIdOferta <= 0) {
		httpx.Fail(&c.Controller, http.StatusBadRequest, "cuponId y ofertaId deben ser enteros positivos", nil)
		return
	}

	aplicado, err := services.NewDescuentoService(orm.NewOrm()).AplicarDescuento(c.Ctx.Request.Context(), pedidoID, &req)
	if err != nil {
		c.aplicarError(pedidoID, err)
		return
	}
	httpx.Send(&c.Controller, http.StatusCreated, "Descuento aplicado exitosamente", aplicado)
}

// aplicarError traduce los errores tipados del servicio a su código HTTP.
func (c *DescuentoController) aplicarError(pedidoID int64, err error) {
	logging.LogControllerError(c.Ctx, "descuentos.post.service_error", err, map[string]interface{}{"pedido_id": pedidoID})
	switch {
	case errors.Is(err, services.ErrPedidoNoEncontrado), errors.Is(err, services.ErrCuponNoEncontrado), errors.Is(err, services.ErrOfertaNoEncontrada):
		httpx.Fail(&c.Controller, http.StatusNotFound, err.Error(), nil)
	case errors.Is(err, services.ErrDescuentoYaAplicado), dberr.IsUnique(err):
		httpx.Fail(&c.Controller, http.StatusConflict, services.ErrDescuentoYaAplicado.Error(), nil)
	case errors.Is(err, services.ErrDescuentoInvalido):
		httpx.Fail(&c.Controller, http.StatusUnprocessableEntity, "Error al aplicar descuento", err)
	default:
		httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error al aplicar descuento", err)
	}
}
