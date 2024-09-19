package controllers

import (
	"encoding/json"
	"net/http"
	"restaurante/models"
	"strconv"

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
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
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

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Relaciones obtenidas exitosamente",
		Data:    relaciones,
	}
	c.ServeJSON()
}

// @Title GetById
// @Summary Obtener relación por ID
// @Description Devuelve una relación específica por ID.
// @Tags pedido_clientes
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID de la Relación"
// @Success 200 {object} models.PedidoCliente "Relación encontrada"
// @Failure 404 {object} models.ApiResponse "Relación no encontrada"
// @Security BearerAuth
// @Router /pedido_clientes/search [get]
func (c *PedidoClienteController) GetById() {
	o := orm.NewOrm()
	id, err := c.GetInt64("id")

	if err != nil || id == 0 {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El parámetro 'id' es inválido o está ausente",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	relacion := models.PedidoCliente{PK_ID_PEDIDO_CLIENTE: id}

	err = o.Read(&relacion)
	if err == orm.ErrNoRows {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Relación no encontrada",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Relación encontrada",
		Data:    relacion,
	}
	c.ServeJSON()
}

// @Title Create
// @Summary Crear una nueva relación pedido-cliente
// @Description Crea una nueva relación entre un pedido y un cliente.
// @Tags pedido_clientes
// @Accept json
// @Produce json
// @Param   body  body   models.PedidoCliente true  "Datos de la relación a crear"
// @Success 201 {object} models.PedidoCliente "Relación creada"
// @Failure 400 {object} models.ApiResponse "Error en la solicitud"
// @Security BearerAuth
// @Router /pedido_clientes [post]
func (c *PedidoClienteController) Post() {
	o := orm.NewOrm()
	var relacion models.PedidoCliente

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &relacion); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Error en la solicitud",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	id, err := o.Insert(&relacion)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al crear la relación",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	relacion.PK_ID_PEDIDO_CLIENTE = id

	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusCreated,
		Message: "Relación creada correctamente",
		Data:    relacion,
	}
	c.ServeJSON()
}

// @Title Update
// @Summary Actualizar una relación pedido-cliente
// @Description Actualiza los datos de una relación existente.
// @Tags pedido_clientes
// @Accept json
// @Produce json
// @Param   id    query    int  true   "ID de la Relación"
// @Param   body  body   models.PedidoCliente true  "Datos de la relación a actualizar"
// @Success 200 {object} models.PedidoCliente "Relación actualizada"
// @Failure 404 {object} models.ApiResponse "Relación no encontrada"
// @Security BearerAuth
// @Router /pedido_clientes [put]
func (c *PedidoClienteController) Put() {
	o := orm.NewOrm()

	idStr := c.GetString("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id == 0 {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El parámetro 'id' es inválido o está ausente",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	relacion := models.PedidoCliente{PK_ID_PEDIDO_CLIENTE: id}

	if o.Read(&relacion) == nil {
		var updatedRelacion models.PedidoCliente
		if err := json.Unmarshal(c.Ctx.Input.RequestBody, &updatedRelacion); err != nil {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusBadRequest,
				Message: "Error en la solicitud",
				Cause:   err.Error(),
			}
			c.ServeJSON()
			return
		}

		updatedRelacion.PK_ID_PEDIDO_CLIENTE = id
		_, err := o.Update(&updatedRelacion)
		if err != nil {
			c.Ctx.Output.SetStatus(http.StatusInternalServerError)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusInternalServerError,
				Message: "Error al actualizar la relación",
				Cause:   err.Error(),
			}
			c.ServeJSON()
			return
		}

		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusOK,
			Message: "Relación actualizada",
			Data:    updatedRelacion,
		}
		c.ServeJSON()
	} else {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Relación no encontrada",
		}
		c.ServeJSON()
	}
}

// @Title Delete
// @Summary Eliminar una relación pedido-cliente
// @Description Elimina una relación de la base de datos.
// @Tags pedido_clientes
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID de la Relación"
// @Success 200 {object} models.ApiResponse "Relación eliminada"
// @Failure 404 {object} models.ApiResponse "Relación no encontrada"
// @Security BearerAuth
// @Router /pedido_clientes [delete]
func (c *PedidoClienteController) Delete() {
	o := orm.NewOrm()

	idStr := c.GetString("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id == 0 {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El parámetro 'id' es inválido o está ausente",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	relacion := models.PedidoCliente{PK_ID_PEDIDO_CLIENTE: id}

	if _, err := o.Delete(&relacion); err == nil {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusOK,
			Message: "Relación eliminada",
		}
		c.ServeJSON()
	} else {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Relación no encontrada",
			Cause:   err.Error(),
		}
		c.ServeJSON()
	}
}
