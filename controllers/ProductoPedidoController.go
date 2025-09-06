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

// detallePedidoInput represents the payload for DetallePedido without precio.
type detallePedidoInput struct {
	PKIDProducto int64 `json:"productoId"`
	Cantidad     int   `json:"cantidad"`
}

// @Title GetAll
// @Summary Obtener los productos de un pedido
// @Description Devuelve los productos consolidados en un pedido específico
// @Tags producto_pedido
// @Accept json
// @Produce json
// @Param pedido_id query int true "ID del pedido"
// @Success 200 {object} models.ApiResponse{data=controllers.ProductoPedidoResponse} "Lista de productos del pedido"
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
	var detalles []models.DetallePedido
	if _, err := o.QueryTable(new(models.DetallePedido)).
		Filter("PKIDPedido", pedidoID).
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
		"pedidoId": pedidoID,
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
// @Param body body map[string]interface{} true "Datos del pedido con productos"
// @Success 201 {object} models.ApiResponse{data=controllers.ProductoPedidoResponse} "Pedido con productos agregado exitosamente"
// @Failure 400 {object} models.ApiResponse "Datos inválidos"
// @Failure 500 {object} models.ApiResponse "Error interno del servidor"
// @Security BearerAuth
// @Router /producto_pedido [post]
func (c *ProductoPedidoController) Post() {
	var input struct {
		PedidoId int64                `json:"pedidoId"`
		Detalles []detallePedidoInput `json:"detalles"`
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

	o := productoPedidoNewOrm()
	var detalles []models.DetallePedido
	for _, d := range input.Detalles {
		pedidoID := input.PedidoId
		productoID := d.PKIDProducto
		detalle := models.DetallePedido{
			PKIDPedido:   &models.Pedido{PK_ID_PEDIDO: pedidoID},
			PKIDProducto: &models.Producto{PK_ID_PRODUCTO: productoID},
			Cantidad:     d.Cantidad,
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

		// Reconsultar para obtener el precio definitivo
		var actualizado models.DetallePedido
		if err := o.QueryTable(new(models.DetallePedido)).
			Filter("PKIDPedido", *detalle.PKIDPedido).
			Filter("PKIDProducto", *detalle.PKIDProducto).
			One(&actualizado); err != nil {
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusInternalServerError,
				Message: "Error al obtener el precio del producto",
				Cause:   err.Error(),
			}
			c.ServeJSON()
			return
		}

		detalles = append(detalles, actualizado)
	}

	response := map[string]interface{}{
		"pedidoId": input.PedidoId,
		"detalles": detalles,
	}

	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusCreated,
		Message: "Pedido con productos agregado exitosamente",
		Data:    response,
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
// @Success 200 {object} models.ApiResponse{data=controllers.ProductoPedidoResponse} "Productos actualizados exitosamente"
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
	var nuevosProductos []detallePedidoInput
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
	qs := o.QueryTable(new(models.DetallePedido)).
		Filter("PKIDPedido", pedidoID)
	if _, err := qs.All(&[]models.DetallePedido{}); err != nil {
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al buscar los detalles del pedido",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	if _, err := qs.Delete(); err != nil {
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al actualizar los productos del pedido",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	var detalles []models.DetallePedido
	for _, d := range nuevosProductos {
		pid := pedidoID
		prodID := d.PKIDProducto
		detalle := models.DetallePedido{
			PKIDPedido:   &models.Pedido{PK_ID_PEDIDO: pid},
			PKIDProducto: &models.Producto{PK_ID_PRODUCTO: prodID},
			Cantidad:     d.Cantidad,
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

		// Reconsultar para obtener el precio definitivo
		var actualizado models.DetallePedido
		if err := o.QueryTable(new(models.DetallePedido)).
			Filter("PKIDPedido", *detalle.PKIDPedido).
			Filter("PKIDProducto", *detalle.PKIDProducto).
			One(&actualizado); err != nil {
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusInternalServerError,
				Message: "Error al obtener el precio del producto",
				Cause:   err.Error(),
			}
			c.ServeJSON()
			return
		}

		detalles = append(detalles, actualizado)
	}

	response := map[string]interface{}{
		"pedidoId": pedidoID,
		"detalles": detalles,
	}

	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Productos del pedido actualizados exitosamente",
		Data:    response,
	}
	c.ServeJSON()
}
