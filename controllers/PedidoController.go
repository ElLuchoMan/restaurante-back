package controllers

import (
	"encoding/json"
	"net/http"
	"restaurante/models"
	"strconv"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

type PedidoController struct {
	web.Controller
}

// @Title GetAll
// @Summary Obtener todos los pedidos
// @Description Devuelve todos los pedidos registrados en la base de datos.
// @Tags pedidos
// @Accept json
// @Produce json
// @Success 200 {array} models.Pedido "Lista de pedidos"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Router /pedidos [get]
func (c *PedidoController) GetAll() {
	o := orm.NewOrm()
	var pedidos []models.Pedido

	_, err := o.QueryTable(new(models.Pedido)).All(&pedidos)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener pedidos de la base de datos",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Pedidos obtenidos exitosamente",
		Data:    pedidos,
	}
	c.ServeJSON()
}

// @Title GetById
// @Summary Obtener pedido por ID
// @Description Devuelve un pedido específico por ID utilizando query parameters.
// @Tags pedidos
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID del Pedido"
// @Success 200 {object} models.Pedido "Pedido encontrado"
// @Failure 404 {object} models.ApiResponse "Pedido no encontrado"
// @Router /pedidos/search [get]
func (c *PedidoController) GetById() {
	o := orm.NewOrm()
	id, err := c.GetInt("id")

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

	pedido := models.Pedido{PK_ID_PEDIDO: id}

	err = o.Read(&pedido)
	if err == orm.ErrNoRows {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Pedido no encontrado",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Pedido encontrado",
		Data:    pedido,
	}
	c.ServeJSON()
}

// @Title Create
// @Summary Crear un nuevo pedido
// @Description Crea un nuevo pedido en la base de datos.
// @Tags pedidos
// @Accept json
// @Produce json
// @Param   body  body   models.Pedido true  "Datos del pedido a crear"
// @Success 201 {object} models.Pedido "Pedido creado"
// @Failure 400 {object} models.ApiResponse "Error en la solicitud"
// @Router /pedidos [post]
func (c *PedidoController) Post() {
	o := orm.NewOrm()
	var pedido models.Pedido

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &pedido); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Error en la solicitud",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	_, err := o.Insert(&pedido)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al crear el pedido",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusCreated,
		Message: "Pedido creado correctamente",
		Data:    pedido,
	}
	c.ServeJSON()
}

// @Title Update
// @Summary Actualizar un pedido
// @Description Actualiza los datos de un pedido existente.
// @Tags pedidos
// @Accept json
// @Produce json
// @Param   id    query    int  true   "ID del Pedido"
// @Param   body  body   models.Pedido true  "Datos del pedido a actualizar"
// @Success 200 {object} models.Pedido "Pedido actualizado"
// @Failure 404 {object} models.ApiResponse "Pedido no encontrado"
// @Router /pedidos [put]
func (c *PedidoController) Put() {
	o := orm.NewOrm()

	idStr := c.GetString("id")
	id, err := strconv.Atoi(idStr)
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

	pedido := models.Pedido{PK_ID_PEDIDO: id}

	if o.Read(&pedido) == nil {
		var updatedPedido models.Pedido
		if err := json.Unmarshal(c.Ctx.Input.RequestBody, &updatedPedido); err != nil {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusBadRequest,
				Message: "Error en la solicitud",
				Cause:   err.Error(),
			}
			c.ServeJSON()
			return
		}

		updatedPedido.PK_ID_PEDIDO = id
		_, err := o.Update(&updatedPedido)
		if err != nil {
			c.Ctx.Output.SetStatus(http.StatusInternalServerError)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusInternalServerError,
				Message: "Error al actualizar el pedido",
				Cause:   err.Error(),
			}
			c.ServeJSON()
			return
		}

		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusOK,
			Message: "Pedido actualizado",
			Data:    updatedPedido,
		}
		c.ServeJSON()
	} else {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Pedido no encontrado",
		}
		c.ServeJSON()
	}
}

// @Title Delete
// @Summary Eliminar un pedido
// @Description Elimina un pedido de la base de datos.
// @Tags pedidos
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID del Pedido"
// @Success 204 {object} nil "Pedido eliminado"
// @Failure 404 {object} models.ApiResponse "Pedido no encontrado"
// @Router /pedidos [delete]
func (c *PedidoController) Delete() {
	o := orm.NewOrm()

	idStr := c.GetString("id")
	id, err := strconv.Atoi(idStr)
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

	pedido := models.Pedido{PK_ID_PEDIDO: id}

	if _, err := o.Delete(&pedido); err == nil {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusOK,
			Message: "Pedido eliminado",
		}
		c.ServeJSON()
	} else {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Pedido no encontrado",
			Cause:   err.Error(),
		}
		c.ServeJSON()
	}
}
