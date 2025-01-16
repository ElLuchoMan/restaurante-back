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

// @Title GetAll
// @Summary Obtener pedidos con múltiples filtros
// @Description Devuelve pedidos filtrados según varios criterios: fecha, rango de fechas, usuario (cliente), tipo de método de pago, si tienen domicilio, etc.
// @Tags pedido
// @Accept json
// @Produce json
// @Param fecha query string false "Fecha específica en formato YYYY-MM-DD"
// @Param desde query string false "Fecha inicial del rango en formato YYYY-MM-DD"
// @Param hasta query string false "Fecha final del rango en formato YYYY-MM-DD"
// @Param mes query int false "Mes del año (1-12)"
// @Param anio query int false "Año para el filtro de mes"
// @Param cliente query int false "ID del cliente (PK_DOCUMENTO_CLIENTE)"
// @Param metodo_pago query string false "Tipo de método de pago (NEQUI, DAVIPLATA, EFECTIVO)"
// @Param domicilio query bool false "Indica si el pedido tiene domicilio (true/false)"
// @Success 200 {array} models.Pedido "Lista de pedidos filtrados"
// @Failure 400 {object} models.ApiResponse "Error en los parámetros de filtro"
// @Failure 500 {object} models.ApiResponse "Error al obtener los pedidos"
// @Security BearerAuth
// @Router /pedidos [get]
func (c *PedidoController) GetAll() {
	o := orm.NewOrm()

	// Construcción de la consulta SQL
	query := `
        SELECT p.* 
        FROM "PEDIDO" p
        LEFT JOIN "PEDIDO_CLIENTE" pc ON p."PK_ID_PEDIDO" = pc."PK_ID_PEDIDO"
        LEFT JOIN "PAGO" pa ON p."PK_ID_PAGO" = pa."PK_ID_PAGO"
        LEFT JOIN "METODO_PAGO" mp ON pa."PK_ID_METODO_PAGO" = mp."PK_ID_METODO_PAGO"
        WHERE 1 = 1
    `

	// Parámetros de filtro
	params := []interface{}{}
	fecha := c.GetString("fecha")
	desde := c.GetString("desde")
	hasta := c.GetString("hasta")
	mes, _ := c.GetInt("mes")
	anio, _ := c.GetInt("anio")
	cliente, _ := c.GetInt("cliente")
	metodoPago := c.GetString("metodo_pago")
	domicilio, errDomicilio := c.GetBool("domicilio")

	// Agregar filtros según los parámetros proporcionados
	if fecha != "" {
		query += ` AND p."FECHA" = ?`
		params = append(params, fecha)
	}

	if desde != "" && hasta != "" {
		query += ` AND p."FECHA" BETWEEN ? AND ?`
		params = append(params, desde, hasta)
	}

	if mes > 0 && mes <= 12 {
		query += ` AND EXTRACT(MONTH FROM p."FECHA") = ?`
		params = append(params, mes)
		if anio > 0 {
			query += ` AND EXTRACT(YEAR FROM p."FECHA") = ?`
			params = append(params, anio)
		}
	}

	if cliente > 0 {
		query += ` AND pc."PK_DOCUMENTO_CLIENTE" = ?`
		params = append(params, cliente)
	}

	if metodoPago != "" {
		query += ` AND mp."TIPO" ILIKE ?`
		params = append(params, metodoPago)
	}

	if errDomicilio == nil {
		if domicilio {
			query += ` AND p."PK_ID_DOMICILIO" IS NOT NULL`
		} else {
			query += ` AND p."PK_ID_DOMICILIO" IS NULL`
		}
	}

	// Ejecutar la consulta y obtener los resultados
	var pedidos []models.Pedido
	_, err := o.Raw(query, params...).QueryRows(&pedidos)
	if err != nil {
		c.CustomAbort(500, "Error al obtener los pedidos: "+err.Error())
		return
	}

	// Responder con los pedidos obtenidos
	c.Data["json"] = map[string]interface{}{
		"message": "Pedidos obtenidos exitosamente",
		"pedidos": pedidos,
	}
	c.ServeJSON()
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
// @Router /pedidos [post]
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
// @Router /pedidos/asignar-domicilio [post]
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
// @Router /pedidos/asignar-pago [post]
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
// @Router /pedidos/actualizar-estado [put]
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
