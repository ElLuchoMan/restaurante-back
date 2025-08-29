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
	PedidoID          int64       `json:"pedidoId"`
	DetallesProductos interface{} `json:"detallesProductos"`
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

	// Convertir el JSONB a un formato de salida legible
	var detalles []map[string]interface{}
	if err := json.Unmarshal([]byte(productoPedido.DETALLES_PRODUCTOS), &detalles); err != nil {
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al procesar los detalles del pedido",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Transformar las claves de los detalles a camelCase
	var detallesCamelCase []map[string]interface{}
	for _, detalle := range detalles {
		camelCaseDetalle := map[string]interface{}{
			"cantidad":       detalle["CANTIDAD"],
			"nombre":         detalle["NOMBRE"],
			"productoId":     detalle["PK_ID_PRODUCTO"],
			"precioUnitario": detalle["PRECIO_UNITARIO"],
			"subtotal":       detalle["SUBTOTAL"],
		}
		detallesCamelCase = append(detallesCamelCase, camelCaseDetalle)
	}

	// Construir la respuesta
	response := map[string]interface{}{
		"pedidoId":          productoPedido.PK_ID_PEDIDO,
		"detallesProductos": detallesCamelCase,
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
		PedidoId          int64                    `json:"pedidoId"`
		DetallesProductos []map[string]interface{} `json:"detallesProductos"`
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

	// Validar que se proporcione el pedido y los detalles
	if input.PedidoId == 0 || len(input.DetallesProductos) == 0 {
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El pedido y los detalles de los productos son obligatorios",
		}
		c.ServeJSON()
		return
	}

	// Convertir los detalles a JSON
	detallesJSON, err := json.Marshal(input.DetallesProductos)
	if err != nil {
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al procesar los detalles del pedido",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	productoPedido := models.ProductoPedido{
		PK_ID_PEDIDO:       input.PedidoId,
		DETALLES_PRODUCTOS: string(detallesJSON),
	}

	o := productoPedidoNewOrm()
	_, err = o.Insert(&productoPedido)
	if err != nil {
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al crear el pedido con productos",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
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
	var nuevosProductos []map[string]interface{}
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

	// Convertir la nueva lista de productos a JSON
	nuevosDetallesJSON, err := json.Marshal(nuevosProductos)
	if err != nil {
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al procesar los detalles actualizados",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Actualizar los detalles en la base de datos
	productoPedido.DETALLES_PRODUCTOS = string(nuevosDetallesJSON)
	if _, err := o.Update(&productoPedido, "DETALLES_PRODUCTOS"); err != nil {
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al actualizar los productos del pedido",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Productos del pedido actualizados exitosamente",
		Data:    nuevosProductos,
	}
	c.ServeJSON()
}
