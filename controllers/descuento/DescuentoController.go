package descuento

import (
	"context"
	"encoding/json"
	"net/http"

	"restaurante/logging"
	"restaurante/models"
	"restaurante/services"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

type descuentoService interface {
	ObtenerDescuentosPedido(ctx context.Context, pedidoId int64) ([]*models.PedidoDescuentoAplicado, error)
	ValidarExclusividadDescuento(ctx context.Context, pedidoId int64, cuponId *int64, ofertaId *int64) error
	AplicarDescuento(ctx context.Context, pedidoId int64, req *models.AplicarDescuentoRequest) (*models.PedidoDescuentoAplicado, error)
}

// Interfaces para testing
type descuentoOrmer interface {
	Read(interface{}, ...string) error
}

type descOrmAdapter struct {
	readFn func(interface{}, ...string) error
}

func (a descOrmAdapter) Read(v interface{}, cols ...string) error { return a.readFn(v, cols...) }

var ormReadProvider = defaultOrmReadProvider

var descBaseReadFunc = func() func(interface{}, ...string) error { return ormReadProvider() }

var descReadFuncFactory = func() func(interface{}, ...string) error { return descBaseReadFunc() }

var descOrmFactory = func() descuentoOrmer { return descOrmAdapter{readFn: descReadFuncFactory()} }

var descOrmNew = func() descuentoOrmer { return descOrmFactory() }

// Variable mockeable para tests
var newDescuentoService = func(o orm.Ormer) descuentoService {
	return services.NewDescuentoService(o)
}

var ormProvider = defaultOrmProvider

var descServiceOrmBase = func() orm.Ormer { return ormProvider() }

var descuentoServiceOrmFactory = func() orm.Ormer { return descServiceOrmBase() }

type DescuentoController struct {
	web.Controller
}

// @Title GetAll
// @Summary Obtener descuentos de pedido
// @Tags descuentos
// @Accept json
// @Produce json
// @Param pedido_id query int true "ID del pedido"
// @Success 200 {object} models.ApiResponse{data=[]models.PedidoDescuentoAplicado}
// @Failure 400 {object} models.ApiResponse
// @Failure 404 {object} models.ApiResponse
// @Failure 500 {object} models.ApiResponse
// @Router /descuentos/pedidos [get]
func (c *DescuentoController) GetAll() {
	pedidoId, err := c.GetInt64("pedido_id")
	if err != nil || pedidoId == 0 {
		logging.LogControllerError(c.Ctx, "descuentos.getall.bad_request", err, map[string]interface{}{"pedido_id": c.GetString("pedido_id")})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "ID de pedido inválido o ausente",
		}
		_ = c.ServeJSON()
		return
	}

	o := descOrmNew()

	// Verificar que el pedido existe
	pedido := &models.Pedido{PK_ID_PEDIDO: pedidoId}
	err = o.Read(pedido)
	if err != nil {
		if err == orm.ErrNoRows {
			c.Ctx.Output.SetStatus(http.StatusNotFound)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusNotFound,
				Message: "Pedido no encontrado",
			}
			_ = c.ServeJSON()
			return
		}
		logging.LogControllerError(c.Ctx, "descuentos.getall.read_error", err, map[string]interface{}{"pedido_id": pedidoId})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error interno del servidor",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	descuentoService := newDescuentoService(descuentoServiceOrmFactory())
	descuentos, err := descuentoService.ObtenerDescuentosPedido(c.Ctx.Request.Context(), pedidoId)
	if err != nil {
		logging.LogControllerError(c.Ctx, "descuentos.getall.service_error", err, map[string]interface{}{"pedido_id": pedidoId})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener descuentos",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Descuentos obtenidos exitosamente",
		Data:    descuentos,
	}
	_ = c.ServeJSON()
}

// @Title Post
// @Summary Aplicar descuento a pedido
// @Tags descuentos
// @Accept json
// @Produce json
// @Param pedido_id query int true "ID del pedido"
// @Param body body models.AplicarDescuentoRequest true "Datos del descuento a aplicar"
// @Success 201 {object} models.ApiResponse{data=models.PedidoDescuentoAplicado}
// @Failure 400 {object} models.ApiResponse
// @Failure 404 {object} models.ApiResponse
// @Failure 409 {object} models.ApiResponse
// @Failure 422 {object} models.ApiResponse
// @Router /descuentos/pedidos [post]
func (c *DescuentoController) Post() {
	pedidoId, err := c.GetInt64("pedido_id")
	if err != nil || pedidoId == 0 {
		logging.LogControllerError(c.Ctx, "descuentos.post.bad_request", err, map[string]interface{}{"pedido_id": c.GetString("pedido_id")})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "ID de pedido inválido o ausente",
		}
		_ = c.ServeJSON()
		return
	}

	var req models.AplicarDescuentoRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		logging.LogControllerError(c.Ctx, "descuentos.post.bad_json", err, map[string]interface{}{"pedido_id": pedidoId})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "JSON inválido",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	// Validar que se especifique exactamente uno de cupón o oferta
	if (req.PkIdCupon == nil && req.PkIdOferta == nil) || (req.PkIdCupon != nil && req.PkIdOferta != nil) {
		logging.LogControllerError(c.Ctx, "descuentos.post.invalid_request", nil, map[string]interface{}{"pedido_id": pedidoId, "cupon": req.PkIdCupon, "oferta": req.PkIdOferta})
		c.Ctx.Output.SetStatus(http.StatusUnprocessableEntity)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusUnprocessableEntity,
			Message: "Debe especificar exactamente uno de cupón o oferta",
		}
		_ = c.ServeJSON()
		return
	}

	descuentoService := newDescuentoService(descuentoServiceOrmFactory())

	// Validar exclusividad antes de aplicar
	err = descuentoService.ValidarExclusividadDescuento(c.Ctx.Request.Context(), pedidoId, req.PkIdCupon, req.PkIdOferta)
	if err != nil {
		logging.LogControllerError(c.Ctx, "descuentos.post.exclusivity_error", err, map[string]interface{}{"pedido_id": pedidoId})
		c.Ctx.Output.SetStatus(http.StatusConflict)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusConflict,
			Message: err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	descuentoAplicado, err := descuentoService.AplicarDescuento(c.Ctx.Request.Context(), pedidoId, &req)
	if err != nil {
		logging.LogControllerError(c.Ctx, "descuentos.post.service_error", err, map[string]interface{}{"pedido_id": pedidoId})

		// Determinar el tipo de error
		errorMsg := err.Error()
		switch errorMsg {
		case "pedido no encontrado":
			c.Ctx.Output.SetStatus(http.StatusOK)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusNotFound,
				Message: "Pedido no encontrado",
			}
		case "cupón no encontrado", "oferta no encontrada":
			c.Ctx.Output.SetStatus(http.StatusOK)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusNotFound,
				Message: errorMsg,
			}
		case "ya existe un descuento aplicado para este pedido":
			c.Ctx.Output.SetStatus(http.StatusConflict)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusConflict,
				Message: errorMsg,
			}
		default:
			c.Ctx.Output.SetStatus(http.StatusUnprocessableEntity)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusUnprocessableEntity,
				Message: "Error al aplicar descuento",
				Cause:   errorMsg,
			}
		}
		_ = c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusCreated,
		Message: "Descuento aplicado exitosamente",
		Data:    descuentoAplicado,
	}
	_ = c.ServeJSON()
}
