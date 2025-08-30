package controllers

import (
	"encoding/json"
	"net/http"
	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

type productoPedidoOrmer interface {
	QueryTable(interface{}) orm.QuerySeter
	Insert(interface{}) (int64, error)
	Update(interface{}, ...string) (int64, error)
}

// productoPedidoNewOrm allows tests to stub orm.NewOrm.
var productoPedidoNewOrm = func() productoPedidoOrmer { return orm.NewOrm() }

type ProductoPedidoController struct {
	web.Controller
}

// Estructura para mapear las respuestas en camelCase
type ProductoPedidoResponse struct {
	PedidoID int64       `json:"pedidoId"`
	Detalles interface{} `json:"detalles"`
}

// @Title GetAll
// @Summary Obtener los productos de un pedido
// @Description Devuelve los productos consolidados en un pedido específico
// @Tags producto_pedido
// @Accept json
// @Produce json
// @Param pedido_id query int true "ID del pedido"
// @Success 200 {object} models.ApiResponse "Lista de productos del pedido"
// @Failure 404 {object} models.ApiResponse "No se encontraron productos asociados a este pedido"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /producto_pedido [get]
func (c *ProductoPedidoController) GetAll() {
	pedidoID, err := c.GetInt64("pedido_id")
	if err != nil || pedidoID == 0 {
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El parámetro 'pedido_id' es obligatorio y debe ser válido",
		}
		c.ServeJSON()
		return
	}

	o := productoPedidoNewOrm()
	var productoPedido models.ProductoPedido

	err = o.QueryTable(new(models.ProductoPedido)).
		Filter("PK_ID_PEDIDO", pedidoID).
		One(&productoPedido)

	if err == orm.ErrNoRows {
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "No se encontraron productos asociados a este pedido",
		}
		c.ServeJSON()
		return
	} else if err != nil {
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener los productos del pedido",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	var detalles []models.ProductoPedidoDetalle
	if _, err := o.QueryTable(new(models.ProductoPedidoDetalle)).
		Filter("PK_ID_PEDIDO", pedidoID).
		All(&detalles); err != nil {
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener los productos del pedido",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	if len(detalles) == 0 {
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "No se encontraron productos asociados a este pedido",
		}
		c.ServeJSON()
		return
	}

	response := map[string]interface{}{
		"pedidoId": productoPedido.PK_ID_PEDIDO,
		"detalles": detalles,
	}

	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Productos del pedido obtenidos exitosamente",
		Data:    response,
	}
	c.ServeJSON()
}

// @Title Post
// @Summary Crear un pedido con productos consolidados
// @Description Crea un registro de productos consolidados en un pedido
// @Tags producto_pedido
// @Accept json
// @Produce json
// @Param body body models.ProductoPedido true "Datos del pedido con productos"
// @Success 201 {object} models.ApiResponse "Pedido con productos agregado exitosamente"
// @Failure 400 {object} models.ApiResponse "Datos inválidos"
// @Failure 500 {object} models.ApiResponse "Error interno del servidor"
// @Security BearerAuth
// @Router /producto_pedido [post]
func (c *ProductoPedidoController) Post() {
	var input struct {
		PedidoId int64                          `json:"pedidoId"`
		Detalles []models.ProductoPedidoDetalle `json:"detalles"`
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &input); err != nil {
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Datos inválidos",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	if input.PedidoId == 0 || len(input.Detalles) == 0 {
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El pedido y los detalles de los productos son obligatorios",
		}
		c.ServeJSON()
		return
	}

	productoPedido := models.ProductoPedido{PK_ID_PEDIDO: input.PedidoId}
	o := productoPedidoNewOrm()
	if _, err := o.Insert(&productoPedido); err != nil {
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al crear el pedido con productos",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	for _, d := range input.Detalles {
		detalle := models.ProductoPedidoDetalle{
			PKIDPedido:   input.PedidoId,
			PKIDProducto: d.PKIDProducto,
			CANTIDAD:     d.CANTIDAD,
		}
		if _, err := o.Insert(&detalle); err != nil {
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusInternalServerError,
				Message: "Error al crear el pedido con productos",
				Cause:   err.Error(),
			}
			c.ServeJSON()
			return
		}
		productoPedido.Detalles = append(productoPedido.Detalles, detalle)
	}

	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusCreated,
		Message: "Pedido con productos agregado exitosamente",
		Data:    productoPedido,
	}
	c.ServeJSON()
}

// @Title Update
// @Summary Actualizar productos en un pedido consolidado
// @Description Permite agregar o modificar productos en un pedido consolidado
// @Tags producto_pedido
// @Accept json
// @Produce json
// @Param pedido_id query int true "ID del pedido a actualizar"
// @Param body body []map[string]interface{} true "Lista actualizada de productos"
// @Success 200 {object} models.ApiResponse "Productos actualizados exitosamente"
// @Failure 400 {object} models.ApiResponse "Datos inválidos"
// @Failure 404 {object} models.ApiResponse "Pedido no encontrado"
// @Failure 500 {object} models.ApiResponse "Error interno del servidor"
// @Security BearerAuth
// @Router /producto_pedido [put]
func (c *ProductoPedidoController) Update() {
	pedidoID, err := c.GetInt64("pedido_id")
	if err != nil || pedidoID == 0 {
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El parámetro 'pedido_id' es obligatorio y debe ser válido",
		}
		c.ServeJSON()
		return
	}

	// Parsear los datos del cuerpo de la solicitud
	var nuevosProductos []models.ProductoPedidoDetalle
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &nuevosProductos); err != nil {
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Datos inválidos",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	if len(nuevosProductos) == 0 {
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "La lista de productos no puede estar vacía",
		}
		c.ServeJSON()
		return
	}

	o := productoPedidoNewOrm()
	productoPedido := models.ProductoPedido{}

	// Verificar si existe el pedido en la base de datos
	err = o.QueryTable(new(models.ProductoPedido)).
		Filter("PK_ID_PEDIDO", pedidoID).
		One(&productoPedido)
	if err == orm.ErrNoRows {
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Pedido no encontrado",
		}
		c.ServeJSON()
		return
	} else if err != nil {
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al buscar el pedido",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	if _, err := o.QueryTable(new(models.ProductoPedidoDetalle)).
		Filter("PK_ID_PEDIDO", pedidoID).
		Delete(); err != nil {
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al actualizar los productos del pedido",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	for _, d := range nuevosProductos {
		detalle := models.ProductoPedidoDetalle{
			PKIDPedido:   pedidoID,
			PKIDProducto: d.PKIDProducto,
			CANTIDAD:     d.CANTIDAD,
		}
		if _, err := o.Insert(&detalle); err != nil {
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusInternalServerError,
				Message: "Error al actualizar los productos del pedido",
				Cause:   err.Error(),
			}
			c.ServeJSON()
			return
		}
		productoPedido.Detalles = append(productoPedido.Detalles, detalle)
	}

	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Productos del pedido actualizados exitosamente",
		Data:    nuevosProductos,
	}
	c.ServeJSON()
}
