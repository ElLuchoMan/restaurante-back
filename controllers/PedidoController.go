package controllers

import (
	"restaurante/models" // Asegúrate de que la ruta del paquete sea la correcta
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
// @Success 200 {object} models.Pedido "Pedido creado"
// @Failure 400 {object} models.ApiResponse "Datos inválidos"
// @Failure 500 {object} models.ApiResponse "Error al crear el pedido"
// @Security BearerAuth
// @Router /pedido [post]
func (c *PedidoController) CreatePedido() {
	var pedido models.Pedido

	if err := c.ParseForm(&pedido); err != nil {
		c.CustomAbort(400, "Datos inválidos")
		return
	}

	pedido.FECHA = time.Now()
	pedido.ESTADO_PEDIDO = "INICIADO"
	o := orm.NewOrm()
	if _, err := o.Insert(&pedido); err != nil {
		c.CustomAbort(500, "Error al crear el pedido")
		return
	}

	c.Data["json"] = map[string]interface{}{"message": "Pedido creado", "pedido": pedido}
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
// @Success 200 {object} models.Pedido "Domicilio asignado al pedido"
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
		c.CustomAbort(404, "Pedido no encontrado")
		return
	}

	// Actualizar el domicilio y el estado del pedido
	pedido.PK_ID_DOMICILIO = &domicilioID
	pedido.ESTADO_PEDIDO = "EN CAMINO"

	if _, err := o.Update(&pedido, "PK_ID_DOMICILIO", "ESTADO_PEDIDO"); err != nil {
		c.CustomAbort(500, "Error al asignar domicilio")
		return
	}

	// Actualizar el estado del domicilio
	domicilio := models.Domicilio{PK_ID_DOMICILIO: domicilioID}
	if err := o.Read(&domicilio); err == nil {
		domicilio.ENTREGADO = false
		if _, err := o.Update(&domicilio, "ENTREGADO"); err != nil {
			c.CustomAbort(500, "Error al actualizar el domicilio")
			return
		}
	}

	c.Data["json"] = map[string]interface{}{"message": "Domicilio asignado", "pedido": pedido}
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
// @Success 200 {object} models.Pedido "Pago asignado al pedido"
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
		c.CustomAbort(404, "Pedido no encontrado")
		return
	}

	// Actualizar el pago y el estado del pedido
	pedido.PK_ID_PAGO = &pagoID
	pedido.ESTADO_PEDIDO = "PAGADO"

	if _, err := o.Update(&pedido, "PK_ID_PAGO", "ESTADO_PEDIDO"); err != nil {
		c.CustomAbort(500, "Error al asignar pago")
		return
	}

	// Actualizar el estado del pago
	pago := models.Pago{PK_ID_PAGO: pagoID}
	if err := o.Read(&pago); err == nil {
		pago.ESTADO_PAGO = "PAGADO"
		if _, err := o.Update(&pago, "ESTADO_PAGO"); err != nil {
			c.CustomAbort(500, "Error al actualizar el pago")
			return
		}
	}

	c.Data["json"] = map[string]interface{}{"message": "Pago asignado", "pedido": pedido}
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
// @Success 200 {object} models.Pedido "Estado actualizado"
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
		c.CustomAbort(404, "Pedido no encontrado")
		return
	}

	// Actualizar el estado del pedido
	pedido.ESTADO_PEDIDO = estado

	if _, err := o.Update(&pedido, "ESTADO_PEDIDO"); err != nil {
		c.CustomAbort(500, "Error al actualizar estado del pedido")
		return
	}

	c.Data["json"] = map[string]interface{}{"message": "Estado actualizado", "pedido": pedido}
	c.ServeJSON()
}
