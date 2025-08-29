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

	type detalleRow struct {
		ProductoID     int64   `orm:"column(producto_id)"`
		Nombre         string  `orm:"column(nombre)"`
		Cantidad       int     `orm:"column(cantidad)"`
		PrecioUnitario float64 `orm:"column(precio_unitario)"`
		Subtotal       float64 `orm:"column(subtotal)"`
	}

	var rows []detalleRow
	count, err := o.Raw(`
SELECT pr."PK_ID_PRODUCTO" AS producto_id,
       pr."NOMBRE" AS nombre,
       ppd."CANTIDAD" AS cantidad,
       ppd."PRECIO_UNITARIO" AS precio_unitario,
       ppd."SUBTOTAL" AS subtotal
FROM "PRODUCTO_PEDIDO" pp
JOIN "PRODUCTO_PEDIDO_DETALLE" ppd ON pp."PK_ID_PEDIDO" = ppd."PK_ID_PEDIDO"
JOIN "PRODUCTO" pr ON pr."PK_ID_PRODUCTO" = ppd."PK_ID_PRODUCTO"
WHERE pp."PK_ID_PEDIDO" = ?`, pedidoID).QueryRows(&rows)
	if err != nil {
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener los productos del pedido",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}
	if count == 0 {
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "No se encontraron productos asociados a este pedido",
		}
		c.ServeJSON()
		return
	}

	detalles := make([]map[string]interface{}, 0, len(rows))
	for _, r := range rows {
		detalles = append(detalles, map[string]interface{}{
			"productoId":     r.ProductoID,
			"nombre":         r.Nombre,
			"cantidad":       r.Cantidad,
			"precioUnitario": r.PrecioUnitario,
			"subtotal":       r.Subtotal,
		})
	}

	response := map[string]interface{}{
		"pedidoId":          pedidoID,
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
// @Description Crea registros de productos asociados a un pedido
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
		PedidoId          int64                          `json:"pedidoId"`
		DetallesProductos []models.ProductoPedidoDetalle `json:"detallesProductos"`
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
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al iniciar la transacción",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	pp := models.ProductoPedido{PK_ID_PEDIDO: input.PedidoId, DETALLES_PRODUCTOS: "[]"}
	if _, err := tx.Insert(&pp); err != nil {
		_ = tx.Rollback()
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al crear el pedido con productos",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	for _, d := range input.DetallesProductos {
		d.PKIDPedido = input.PedidoId
		if _, err := tx.Insert(&d); err != nil {
			_ = tx.Rollback()
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusInternalServerError,
				Message: "Error al crear el pedido con productos",
				Cause:   err.Error(),
			}
			c.ServeJSON()
			return
		}
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
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
		Data:    map[string]interface{}{"pedidoId": input.PedidoId},
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

	o := orm.NewOrm()
	// Verificar existencia del pedido
	var productoPedido models.ProductoPedido
	if err := o.QueryTable(new(models.ProductoPedido)).Filter("PK_ID_PEDIDO", pedidoID).One(&productoPedido); err != nil {
		if err == orm.ErrNoRows {
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusNotFound,
				Message: "Pedido no encontrado",
			}
		} else {
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusInternalServerError,
				Message: "Error al buscar el pedido",
				Cause:   err.Error(),
			}
		}
		c.ServeJSON()
		return
	}

	tx, err := o.Begin()
	if err != nil {
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al iniciar la transacción",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	if _, err := tx.Raw(`DELETE FROM "PRODUCTO_PEDIDO_DETALLE" WHERE "PK_ID_PEDIDO" = ?`, pedidoID).Exec(); err != nil {
		_ = tx.Rollback()
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al actualizar los productos del pedido",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	for _, d := range nuevosProductos {
		d.PKIDPedido = pedidoID
		if _, err := tx.Insert(&d); err != nil {
			_ = tx.Rollback()
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusInternalServerError,
				Message: "Error al actualizar los productos del pedido",
				Cause:   err.Error(),
			}
			c.ServeJSON()
			return
		}
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
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
