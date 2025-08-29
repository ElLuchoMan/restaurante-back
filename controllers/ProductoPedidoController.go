package controllers

import (
	"encoding/json"
	"net/http"
	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

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
	o := orm.NewOrm()

	query := `
SELECT
    pp."PK_ID_PEDIDO" AS pedido_id,
    COALESCE(
        json_agg(
            json_build_object(
                'cantidad', ppd."CANTIDAD",
                'nombre', pr."NOMBRE",
                'productoId', ppd."PK_ID_PRODUCTO",
                'precioUnitario', ppd."PRECIO_UNITARIO",
                'subtotal', ppd."SUBTOTAL"
            )
        ) FILTER (WHERE ppd."PK_ID_PRODUCTO_PEDIDO_DETALLE" IS NOT NULL), '[]'
    )::text AS detalles
FROM "PRODUCTO_PEDIDO" pp
LEFT JOIN "PRODUCTO_PEDIDO_DETALLE" ppd ON ppd."PK_ID_PRODUCTO_PEDIDO" = pp."PK_ID_PRODUCTO_PEDIDO"
LEFT JOIN "PRODUCTO" pr ON pr."PK_ID_PRODUCTO" = ppd."PK_ID_PRODUCTO"
WHERE pp."PK_ID_PEDIDO" = ?
GROUP BY pp."PK_ID_PEDIDO";`

	var row struct {
		PedidoID int64  `orm:"column(pedido_id)"`
		Detalles string `orm:"column(detalles)"`
	}

	err = o.Raw(query, pedidoID).QueryRow(&row)
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

	var detalles []map[string]interface{}
	if row.Detalles != "" {
		_ = json.Unmarshal([]byte(row.Detalles), &detalles)
	}

	response := map[string]interface{}{
		"pedidoId":          row.PedidoID,
		"detallesProductos": detalles,
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

	o := orm.NewOrm()
	tx, err := o.Begin()
	if err != nil {
		c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "No se pudo iniciar la transacción", Cause: err.Error()}
		c.ServeJSON()
		return
	}

	productoPedido := models.ProductoPedido{PK_ID_PEDIDO: input.PedidoId}
	if _, err := tx.Insert(&productoPedido); err != nil {
		tx.Rollback()
		c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al crear el pedido con productos", Cause: err.Error()}
		c.ServeJSON()
		return
	}

	for _, det := range input.DetallesProductos {
		productoID, _ := det["productoId"].(float64)
		cantidad, _ := det["cantidad"].(float64)
		precio, _ := det["precioUnitario"].(float64)
		subtotal, ok := det["subtotal"].(float64)
		if !ok {
			subtotal = precio * cantidad
		}

		detalle := models.ProductoPedidoDetalle{
			PK_ID_PRODUCTO_PEDIDO: productoPedido.PK_ID_PRODUCTO_PEDIDO,
			PK_ID_PRODUCTO:        int64(productoID),
			CANTIDAD:              int(cantidad),
			PRECIO_UNITARIO:       precio,
			SUBTOTAL:              subtotal,
		}
		if _, err := tx.Insert(&detalle); err != nil {
			tx.Rollback()
			c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al registrar los productos del pedido", Cause: err.Error()}
			c.ServeJSON()
			return
		}
	}

	if err := tx.Commit(); err != nil {
		c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al confirmar la transacción", Cause: err.Error()}
		c.ServeJSON()
		return
	}

	response := map[string]interface{}{
		"pedidoId":          productoPedido.PK_ID_PEDIDO,
		"detallesProductos": input.DetallesProductos,
	}

	c.Data["json"] = models.ApiResponse{Code: http.StatusCreated, Message: "Pedido con productos agregado exitosamente", Data: response}
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

	o := orm.NewOrm()
	tx, err := o.Begin()
	if err != nil {
		c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "No se pudo iniciar la transacción", Cause: err.Error()}
		c.ServeJSON()
		return
	}

	var productoPedido models.ProductoPedido
	err = tx.QueryTable(new(models.ProductoPedido)).Filter("PK_ID_PEDIDO", pedidoID).One(&productoPedido)
	if err == orm.ErrNoRows {
		tx.Rollback()
		c.Data["json"] = models.ApiResponse{Code: http.StatusNotFound, Message: "Pedido no encontrado"}
		c.ServeJSON()
		return
	} else if err != nil {
		tx.Rollback()
		c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al buscar el pedido", Cause: err.Error()}
		c.ServeJSON()
		return
	}

	if _, err := tx.QueryTable(new(models.ProductoPedidoDetalle)).Filter("PK_ID_PRODUCTO_PEDIDO", productoPedido.PK_ID_PRODUCTO_PEDIDO).Delete(); err != nil {
		tx.Rollback()
		c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al limpiar los productos del pedido", Cause: err.Error()}
		c.ServeJSON()
		return
	}

	for _, det := range nuevosProductos {
		productoID, _ := det["productoId"].(float64)
		cantidad, _ := det["cantidad"].(float64)
		precio, _ := det["precioUnitario"].(float64)
		subtotal, ok := det["subtotal"].(float64)
		if !ok {
			subtotal = precio * cantidad
		}

		detalle := models.ProductoPedidoDetalle{
			PK_ID_PRODUCTO_PEDIDO: productoPedido.PK_ID_PRODUCTO_PEDIDO,
			PK_ID_PRODUCTO:        int64(productoID),
			CANTIDAD:              int(cantidad),
			PRECIO_UNITARIO:       precio,
			SUBTOTAL:              subtotal,
		}
		if _, err := tx.Insert(&detalle); err != nil {
			tx.Rollback()
			c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al actualizar los productos del pedido", Cause: err.Error()}
			c.ServeJSON()
			return
		}
	}

	if err := tx.Commit(); err != nil {
		c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al confirmar la transacción", Cause: err.Error()}
		c.ServeJSON()
		return
	}

	c.Data["json"] = models.ApiResponse{Code: http.StatusOK, Message: "Productos del pedido actualizados exitosamente", Data: nuevosProductos}
	c.ServeJSON()
}
