package controllers

import (
	"restaurante/models" // Ajusta la ruta según tu proyecto
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

type PedidoController struct {
	web.Controller
}

// @Title CreatePedido
// @Summary Crear un nuevo pedido
// @Description Crea un nuevo pedido en el sistema sin domicilio ni pago asociados.
// @Tags pedido
// @Accept json
// @Produce json
// @Param body body models.Pedido true "Datos del pedido"
// @Success 200 {object} models.ApiResponse "Pedido creado"
// @Failure 400 {object} models.ApiResponse "Datos inválidos"
// @Failure 500 {object} models.ApiResponse "Error al crear el pedido"
// @Security BearerAuth
// @Router /pedido [post]
func (c *PedidoController) CreatePedido() {
	var pedido models.Pedido

	if err := c.ParseForm(&pedido); err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = models.ApiResponse{
			Code:    400,
			Message: "Datos inválidos",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	pedido.FECHA = time.Now()
	pedido.ESTADO_PEDIDO = "INICIADO"
	o := orm.NewOrm()
	if _, err := o.Insert(&pedido); err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = models.ApiResponse{
			Code:    500,
			Message: "Error al crear el pedido",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(200)
	c.Data["json"] = models.ApiResponse{
		Code:    200,
		Message: "Pedido creado exitosamente",
		Data:    pedido,
	}
	c.ServeJSON()
}

// @Title AssignDomicilio
// @Summary Asignar un domicilio a un pedido
// @Description Asigna un domicilio existente a un pedido y actualiza su estado a "EN CAMINO".
// @Tags pedido
// @Accept json
// @Produce json
// @Param pedido_id query int true "ID del pedido"
// @Param domicilio_id query int true "ID del domicilio"
// @Success 200 {object} models.ApiResponse "Domicilio asignado al pedido"
// @Failure 404 {object} models.ApiResponse "Pedido o domicilio no encontrado"
// @Failure 500 {object} models.ApiResponse "Error al asignar domicilio"
// @Security BearerAuth
// @Router /pedido/asignar-domicilio [post]
func (c *PedidoController) AssignDomicilio() {
	pedidoID, _ := c.GetInt("pedido_id")
	domicilioID, _ := c.GetInt("domicilio_id")

	o := orm.NewOrm()

	// Buscar el pedido
	pedido := models.Pedido{PK_ID_PEDIDO: pedidoID}
	if err := o.Read(&pedido); err != nil {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = models.ApiResponse{
			Code:    404,
			Message: "Pedido no encontrado",
		}
		c.ServeJSON()
		return
	}

	// Actualizar el domicilio y el estado del pedido
	pedido.PK_ID_DOMICILIO = &domicilioID
	pedido.ESTADO_PEDIDO = "EN CAMINO"

	if _, err := o.Update(&pedido, "PK_ID_DOMICILIO", "ESTADO_PEDIDO"); err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = models.ApiResponse{
			Code:    500,
			Message: "Error al asignar domicilio",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Actualizar el estado del domicilio
	domicilio := models.Domicilio{PK_ID_DOMICILIO: domicilioID}
	if err := o.Read(&domicilio); err == nil {
		domicilio.ENTREGADO = false
		if _, err := o.Update(&domicilio, "ENTREGADO"); err != nil {
			c.Ctx.Output.SetStatus(500)
			c.Data["json"] = models.ApiResponse{
				Code:    500,
				Message: "Error al actualizar el domicilio",
				Cause:   err.Error(),
			}
			c.ServeJSON()
			return
		}
	}

	c.Ctx.Output.SetStatus(200)
	c.Data["json"] = models.ApiResponse{
		Code:    200,
		Message: "Domicilio asignado correctamente",
		Data:    pedido,
	}
	c.ServeJSON()
}

// @Title AssignPago
// @Summary Asignar un pago a un pedido
// @Description Asigna un pago existente a un pedido y actualiza su estado a "PAGADO".
// @Tags pedido
// @Accept json
// @Produce json
// @Param pedido_id query int true "ID del pedido"
// @Param pago_id query int true "ID del pago"
// @Success 200 {object} models.ApiResponse "Pago asignado al pedido"
// @Failure 404 {object} models.ApiResponse "Pedido o pago no encontrado"
// @Failure 500 {object} models.ApiResponse "Error al asignar pago"
// @Security BearerAuth
// @Router /pedido/asignar-pago [post]
func (c *PedidoController) AssignPago() {
	pedidoID, _ := c.GetInt("pedido_id")
	pagoID, _ := c.GetInt("pago_id")

	o := orm.NewOrm()

	// Buscar el pedido
	pedido := models.Pedido{PK_ID_PEDIDO: pedidoID}
	if err := o.Read(&pedido); err != nil {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = models.ApiResponse{
			Code:    404,
			Message: "Pedido no encontrado",
		}
		c.ServeJSON()
		return
	}

	// Actualizar el pago y el estado del pedido
	pedido.PK_ID_PAGO = &pagoID
	pedido.ESTADO_PEDIDO = "PAGADO"

	if _, err := o.Update(&pedido, "PK_ID_PAGO", "ESTADO_PEDIDO"); err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = models.ApiResponse{
			Code:    500,
			Message: "Error al asignar pago",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Actualizar el estado del pago
	pago := models.Pago{PK_ID_PAGO: pagoID}
	if err := o.Read(&pago); err == nil {
		pago.ESTADO_PAGO = "PAGADO"
		if _, err := o.Update(&pago, "ESTADO_PAGO"); err != nil {
			c.Ctx.Output.SetStatus(500)
			c.Data["json"] = models.ApiResponse{
				Code:    500,
				Message: "Error al actualizar el pago",
				Cause:   err.Error(),
			}
			c.ServeJSON()
			return
		}
	}

	c.Ctx.Output.SetStatus(200)
	c.Data["json"] = models.ApiResponse{
		Code:    200,
		Message: "Pago asignado correctamente",
		Data:    pedido,
	}
	c.ServeJSON()
}

// @Title UpdateEstadoPedido
// @Summary Actualizar el estado de un pedido
// @Description Actualiza el estado de un pedido existente.
// @Tags pedido
// @Accept json
// @Produce json
// @Param pedido_id query int true "ID del pedido"
// @Param estado query string true "Nuevo estado del pedido"
// @Success 200 {object} models.ApiResponse "Estado actualizado"
// @Failure 404 {object} models.ApiResponse "Pedido no encontrado"
// @Failure 500 {object} models.ApiResponse "Error al actualizar estado del pedido"
// @Security BearerAuth
// @Router /pedido/actualizar-estado [put]
func (c *PedidoController) UpdateEstadoPedido() {
	pedidoID, _ := c.GetInt("pedido_id")
	estado := c.GetString("estado")

	o := orm.NewOrm()

	// Buscar el pedido
	pedido := models.Pedido{PK_ID_PEDIDO: pedidoID}
	if err := o.Read(&pedido); err != nil {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = models.ApiResponse{
			Code:    404,
			Message: "Pedido no encontrado",
		}
		c.ServeJSON()
		return
	}

	// Actualizar el estado del pedido
	pedido.ESTADO_PEDIDO = estado

	if _, err := o.Update(&pedido, "ESTADO_PEDIDO"); err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = models.ApiResponse{
			Code:    500,
			Message: "Error al actualizar estado del pedido",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(200)
	c.Data["json"] = models.ApiResponse{
		Code:    200,
		Message: "Estado del pedido actualizado correctamente",
		Data:    pedido,
	}
	c.ServeJSON()
}
