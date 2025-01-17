package controllers

import (
	"encoding/json"
	"net/http"
	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

type PedidoClienteController struct {
	web.Controller
}

// @Title GetAll
// @Summary Obtener todas las relaciones pedido-cliente
// @Description Devuelve todas las relaciones entre pedidos y clientes.
// @Tags pedido_clientes
// @Accept json
// @Produce json
// @Success 200 {array} models.PedidoCliente "Lista de relaciones"
// @Security BearerAuth
// @Failure 500 {object} models.ApiResponse "Error interno del servidor"
// @Router /pedido_clientes [get]
func (c *PedidoClienteController) GetAll() {
	o := orm.NewOrm()
	var relaciones []models.PedidoCliente

	_, err := o.QueryTable(new(models.PedidoCliente)).All(&relaciones)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener las relaciones de la base de datos",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Relaciones obtenidas exitosamente",
		Data:    relaciones,
	}
	c.ServeJSON()
}

// @Title Post
// @Summary Crear una nueva relación pedido-cliente
// @Description Crea una nueva relación entre un pedido y un cliente después de validar su existencia.
// @Tags pedido_clientes
// @Accept json
// @Produce json
// @Param body body models.PedidoCliente true "Datos de la relación a crear"
// @Success 201 {object} models.PedidoCliente "Relación creada"
// @Failure 400 {object} models.ApiResponse "Datos inválidos o relación ya existente"
// @Failure 404 {object} models.ApiResponse "Cliente o pedido no encontrado"
// @Failure 500 {object} models.ApiResponse "Error interno del servidor"
// @Security BearerAuth
// @Router /pedido_clientes [post]
func (c *PedidoClienteController) Post() {
	o := orm.NewOrm()
	var relacion models.PedidoCliente

	// Parsear el cuerpo de la solicitud
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &relacion); err != nil {
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Error en la solicitud",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Validar que el cliente existe
	cliente := models.Cliente{PK_DOCUMENTO_CLIENTE: int(*relacion.PK_DOCUMENTO_CLIENTE)}
	if err := o.Read(&cliente); err != nil {
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Cliente no encontrado",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Validar que el pedido existe
	pedido := models.Pedido{PK_ID_PEDIDO: *relacion.PK_ID_PEDIDO}
	if err := o.Read(&pedido); err != nil {
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Pedido no encontrado",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Crear la relación
	id, err := o.Insert(&relacion)
	if err != nil {
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al crear la relación",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	relacion.PK_ID_PEDIDO_CLIENTE = id
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusCreated,
		Message: "Relación creada correctamente",
		Data:    relacion,
	}
	c.ServeJSON()
}
