package controllers

import (
	"encoding/json"
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
// @Success 200 {object} models.ApiResponse{data=[]models.Pedido} "Pedidos obtenidos exitosamente, cada uno con pagoId, metodoPagoId, domicilioId y documentoCliente cuando apliquen"
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
		c.Data["json"] = models.ApiResponse{
			Code:    500,
			Message: "Error al obtener los pedidos",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Validar si no se encontraron resultados
	if len(pedidos) == 0 {
		c.Data["json"] = models.ApiResponse{
			Code:    404,
			Message: "No se encontraron pedidos que coincidan con los filtros proporcionados",
		}
		c.ServeJSON()
		return
	}

	// Responder con los pedidos obtenidos
	c.Data["json"] = models.ApiResponse{
		Code:    200,
		Message: "Pedidos obtenidos exitosamente",
		Data:    pedidos,
	}
	c.ServeJSON()
}

// @Title PostPedido
// @Summary Crear un nuevo pedido
// @Description Crea un nuevo pedido. Fuerza FECHA/HORA (Bogotá) y ESTADO_PEDIDO=INICIADO. Lee 'delivery' del JSON.
// @Tags pedido
// @Accept json
// @Produce json
// @Param body body models.Pedido true "Datos del pedido (sólo se respeta 'delivery')"
// @Success 200 {object} models.ApiResponse{data=models.Pedido} "Pedido creado con identificadores pagoId, metodoPagoId, domicilioId y documentoCliente cuando existan"
// @Failure 400 {object} models.ApiResponse "Datos inválidos"
// @Failure 500 {object} models.ApiResponse "Error al crear el pedido"
// @Security BearerAuth
// @Router /pedidos [post]
func (c *PedidoController) Post() {
	// Estructura mínima para leer el JSON de entrada sin acoplarse al modelo
	var in struct {
		Delivery *bool `json:"delivery"`
		// Ignoramos 'pagoId', 'estadoPedido', etc. por contrato actual
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &in); err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = models.ApiResponse{
			Code:    400,
			Message: "Datos inválidos (JSON mal formado)",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	loc, _ := time.LoadLocation("America/Bogota")
	now := time.Now().In(loc)

	// Construimos el pedido conforme a las reglas
	var pedido models.Pedido
	pedido.FECHA = now
	pedido.HORA = now.Format("15:04:05") // string HH:mm:ss
	pedido.ESTADO_PEDIDO = "INICIADO"

	// Si el cliente mandó delivery en el body, lo respetamos; si no, false
	if in.Delivery != nil {
		pedido.DELIVERY = *in.Delivery
	} else {
		pedido.DELIVERY = false
	}

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
// @Description Asigna un domicilio existente a un pedido (sólo setea PK_ID_DOMICILIO).
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

	// Leer pedido
	pedido := models.Pedido{PK_ID_PEDIDO: pedidoID}
	if err := o.Read(&pedido); err != nil {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = models.ApiResponse{Code: 404, Message: "Pedido no encontrado"}
		c.ServeJSON()
		return
	}

	// Sólo actualizamos la FK al domicilio (por contrato actual)
	pedido.PK_ID_DOMICILIO = &domicilioID
	if _, err := o.Update(&pedido, "PK_ID_DOMICILIO"); err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = models.ApiResponse{Code: 500, Message: "Error al asignar domicilio", Cause: err.Error()}
		c.ServeJSON()
		return
	}

	c.Data["json"] = models.ApiResponse{Code: 200, Message: "Domicilio asignado correctamente", Data: pedido}
	c.ServeJSON()
}

// @Title AssignPago
// @Summary Asignar un pago a un pedido
// @Description Asigna un pago existente a un pedido y actualiza su estado a "TERMINADO" y el pago a "PAGADO".
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

	// Leer pedido
	pedido := models.Pedido{PK_ID_PEDIDO: pedidoID}
	if err := o.Read(&pedido); err != nil {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = models.ApiResponse{Code: 404, Message: "Pedido no encontrado"}
		c.ServeJSON()
		return
	}

	// Actualizamos la FK al pago y marcamos la orden como terminada
	pedido.PK_ID_PAGO = &pagoID
	pedido.ESTADO_PEDIDO = "TERMINADO"
	if _, err := o.Update(&pedido, "PK_ID_PAGO", "ESTADO_PEDIDO"); err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = models.ApiResponse{Code: 500, Message: "Error al asignar pago", Cause: err.Error()}
		c.ServeJSON()
		return
	}

	// También cambiamos el estado en la tabla PAGO
	pago := models.Pago{PK_ID_PAGO: pagoID}
	if err := o.Read(&pago); err == nil {
		pago.ESTADO_PAGO = "PAGADO"
		o.Update(&pago, "ESTADO_PAGO")
	}

	c.Data["json"] = models.ApiResponse{Code: 200, Message: "Pago asignado correctamente", Data: pedido}
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

// @Title GetPedidoDetails
// @Summary Obtener detalles completos de un pedido
// @Description Devuelve la información del pedido, tipo de pago y los productos asociados.
// @Tags pedido
// @Accept json
// @Produce json
// @Param pedido_id query int false "ID del pedido (filtrar por pedido específico)"
// @Success 200 {object} models.ApiResponse "Detalles del pedido obtenidos exitosamente"
// @Failure 400 {object} models.ApiResponse "Error en los parámetros de filtro"
// @Failure 404 {object} models.ApiResponse "Pedido no encontrado"
// @Failure 500 {object} models.ApiResponse "Error al obtener los detalles del pedido"
// @Security BearerAuth
// @Router /pedidos/detalles [get]
func (c *PedidoController) GetPedidoDetails() {
	o := orm.NewOrm()

	// Parámetro
	pedidoID, _ := c.GetInt64("pedido_id")
	if pedidoID == 0 {
		c.Data["json"] = models.ApiResponse{
			Code:    400,
			Message: "El parámetro 'pedido_id' es obligatorio.",
		}
		c.ServeJSON()
		return
	}

	// Consulta para obtener detalles del pedido
	query := `
SELECT
    p."PK_ID_PEDIDO"                                   AS pk_id_pedido,
    COALESCE(TO_CHAR(p."FECHA", 'YYYY-MM-DD'), '')     AS fecha,
    COALESCE(TO_CHAR(p."HORA",  'HH24:MI:SS'), '')     AS hora,
    COALESCE(p."DELIVERY", false)                      AS delivery,
    COALESCE(p."ESTADO_PEDIDO", '')                    AS estado_pedido,
    COALESCE(mp."TIPO", '')                            AS metodo_pago,
    COALESCE((
        SELECT jsonb_agg(json_build_object(
            'PK_ID_PRODUCTO', d."PK_ID_PRODUCTO",
            'NOMBRE', pr."NOMBRE",
            'CANTIDAD', d."CANTIDAD",
            'PRECIO_UNITARIO', d."PRECIO_UNITARIO",
            'SUBTOTAL', d."SUBTOTAL"
        ))::text
        FROM "PRODUCTO_PEDIDO_DETALLE" d
        JOIN "PRODUCTO" pr ON pr."PK_ID_PRODUCTO" = d."PK_ID_PRODUCTO"
        WHERE d."PK_ID_PEDIDO" = p."PK_ID_PEDIDO"
    ), '[]')                                           AS productos,
    COALESCE(p."PK_ID_PAGO", 0)                        AS pago_id,
    COALESCE(pa."PK_ID_METODO_PAGO", 0)                AS metodo_pago_id,
    COALESCE(p."PK_ID_DOMICILIO", 0)                   AS domicilio_id,
    COALESCE(pc."PK_DOCUMENTO_CLIENTE", 0)             AS documento_cliente
FROM "PEDIDO" p
LEFT JOIN "PAGO" pa        ON p."PK_ID_PAGO" = pa."PK_ID_PAGO"
LEFT JOIN "METODO_PAGO" mp ON pa."PK_ID_METODO_PAGO" = mp."PK_ID_METODO_PAGO"
LEFT JOIN "PEDIDO_CLIENTE" pc ON p."PK_ID_PEDIDO" = pc."PK_ID_PEDIDO"
WHERE p."PK_ID_PEDIDO" = ?;
    `

	var details models.PedidoDetails

	// Ejecutar consulta
	err := o.Raw(query, pedidoID).QueryRow(&details)
	if err != nil {
		c.Data["json"] = models.ApiResponse{
			Code:    500,
			Message: "Error al obtener los detalles del pedido",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Respuesta exitosa
	c.Data["json"] = models.ApiResponse{
		Code:    200,
		Message: "Detalles del pedido obtenidos exitosamente",
		Data:    details,
	}
	c.ServeJSON()
}
