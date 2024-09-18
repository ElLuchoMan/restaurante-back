package controllers

import (
	"encoding/json"
	"net/http"
	"restaurante/models"
	"strconv"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

type ItemPedidoController struct {
	web.Controller
}

// @Title GetAll
// @Summary Obtener todos los ítems de pedido
// @Description Devuelve todos los ítems de pedido registrados en la base de datos.
// @Tags item_pedidos
// @Accept json
// @Produce json
// @Success 200 {array} models.ItemPedido "Lista de ítems de pedido"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Router /item_pedidos [get]
func (c *ItemPedidoController) GetAll() {
	o := orm.NewOrm()
	var items []models.ItemPedido

	_, err := o.QueryTable(new(models.ItemPedido)).All(&items)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener los ítems de pedido de la base de datos",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Ítems de pedido obtenidos exitosamente",
		Data:    items,
	}
	c.ServeJSON()
}

// @Title GetById
// @Summary Obtener ítem de pedido por ID
// @Description Devuelve un ítem de pedido específico por ID.
// @Tags item_pedidos
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID del Ítem de Pedido"
// @Success 200 {object} models.ItemPedido "Ítem de pedido encontrado"
// @Failure 404 {object} models.ApiResponse "Ítem de pedido no encontrado"
// @Router /item_pedidos/search [get]
func (c *ItemPedidoController) GetById() {
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

	item := models.ItemPedido{PK_ID_ITEM_PEDIDO: id}

	err = o.Read(&item)
	if err == orm.ErrNoRows {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Ítem de pedido no encontrado",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Ítem de pedido encontrado",
		Data:    item,
	}
	c.ServeJSON()
}

// @Title Create
// @Summary Crear un nuevo ítem de pedido
// @Description Crea un nuevo ítem de pedido en la base de datos.
// @Tags item_pedidos
// @Accept json
// @Produce json
// @Param   body  body   models.ItemPedido true  "Datos del ítem de pedido a crear"
// @Success 201 {object} models.ItemPedido "Ítem de pedido creado"
// @Failure 400 {object} models.ApiResponse "Error en la solicitud"
// @Router /item_pedidos [post]
func (c *ItemPedidoController) Post() {
	o := orm.NewOrm()
	var item models.ItemPedido

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &item); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Error en la solicitud",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	id, err := o.Insert(&item)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al crear el ítem de pedido",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	item.PK_ID_ITEM_PEDIDO = id

	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusCreated,
		Message: "Ítem de pedido creado correctamente",
		Data:    item,
	}
	c.ServeJSON()
}

// @Title Update
// @Summary Actualizar un ítem de pedido
// @Description Actualiza los datos de un ítem de pedido existente.
// @Tags item_pedidos
// @Accept json
// @Produce json
// @Param   id    query    int  true   "ID del Ítem de Pedido"
// @Param   body  body   models.ItemPedido true  "Datos del ítem de pedido a actualizar"
// @Success 200 {object} models.ItemPedido "Ítem de pedido actualizado"
// @Failure 404 {object} models.ApiResponse "Ítem de pedido no encontrado"
// @Router /item_pedidos [put]
func (c *ItemPedidoController) Put() {
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

	item := models.ItemPedido{PK_ID_ITEM_PEDIDO: id}

	if o.Read(&item) == nil {
		var updatedItem models.ItemPedido
		if err := json.Unmarshal(c.Ctx.Input.RequestBody, &updatedItem); err != nil {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusBadRequest,
				Message: "Error en la solicitud",
				Cause:   err.Error(),
			}
			c.ServeJSON()
			return
		}

		updatedItem.PK_ID_ITEM_PEDIDO = id
		_, err := o.Update(&updatedItem)
		if err != nil {
			c.Ctx.Output.SetStatus(http.StatusInternalServerError)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusInternalServerError,
				Message: "Error al actualizar el ítem de pedido",
				Cause:   err.Error(),
			}
			c.ServeJSON()
			return
		}

		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusOK,
			Message: "Ítem de pedido actualizado",
			Data:    updatedItem,
		}
		c.ServeJSON()
	} else {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Ítem de pedido no encontrado",
		}
		c.ServeJSON()
	}
}

// @Title Delete
// @Summary Eliminar un ítem de pedido
// @Description Elimina un ítem de pedido de la base de datos.
// @Tags item_pedidos
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID del Ítem de Pedido"
// @Success 200 {object} models.ApiResponse "Ítem de pedido eliminado"
// @Failure 404 {object} models.ApiResponse "Ítem de pedido no encontrado"
// @Router /item_pedidos [delete]
func (c *ItemPedidoController) Delete() {
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

	item := models.ItemPedido{PK_ID_ITEM_PEDIDO: id}

	if _, err := o.Delete(&item); err == nil {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusOK,
			Message: "Ítem de pedido eliminado",
		}
		c.ServeJSON()
	} else {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Ítem de pedido no encontrado",
			Cause:   err.Error(),
		}
		c.ServeJSON()
	}
}
