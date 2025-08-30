package controllers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"restaurante/models"
	"strconv"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

type DomicilioController struct {
	web.Controller
}

// @Title GetAll
// @Summary Obtener todos los domicilios con posibilidad de filtrar
// @Description Devuelve todos los domicilios registrados en la base de datos, filtrando según criterios específicos.
// @Tags domicilios
// @Accept json
// @Produce json
// @Param   direccion    query   string   false   "Filtrar por dirección"
// @Param   telefono     query   string   false   "Filtrar por teléfono"
// @Param   fecha        query   string   false   "Filtrar por fecha"
// @Param   updated_by   query   string   false   "Filtrar por usuario que realizó la última actualización"
// @Param   trabajador   query   int      false   "ID del domiciliario solicitante"
// @Success 200 {array} models.Domicilio "Lista de domicilios"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /domicilios [get]
func (c *DomicilioController) GetAll() {
	o := orm.NewOrm()
	qs := o.QueryTable(new(models.Domicilio))

	// Leer parámetros de la URL
	direccion := c.GetString("direccion")
	telefono := c.GetString("telefono")
	updatedBy := c.GetString("updated_by")
	fecha := c.GetString("fecha")
	trabajadorID, _ := c.GetInt("trabajador") // ID del domiciliario solicitante

	// Aplicar filtros opcionales SOLO si se proporcionan
	if direccion != "" {
		qs = qs.Filter("DIRECCION__icontains", direccion)
	}
	if telefono != "" {
		qs = qs.Filter("TELEFONO", telefono)
	}
	if updatedBy != "" {
		qs = qs.Filter("UPDATED_BY__icontains", updatedBy)
	}
	if fecha != "" {
		qs = qs.Filter("FECHA", fecha)
	}

	// Aplicar condición para que los domiciliarios solo vean pedidos que pueden tomar
	if trabajadorID != 0 {
		cond := orm.NewCondition().
			Or("PK_DOCUMENTO_TRABAJADOR__isnull", true).
			Or("PK_DOCUMENTO_TRABAJADOR", trabajadorID)

		qs = qs.Filter("ENTREGADO", false).SetCond(cond)
	}

	// Ejecutar consulta
	var domicilios []models.Domicilio
	count, err := qs.All(&domicilios)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener domicilios de la base de datos",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Si no hay domicilios, retornar mensaje informativo
	if count == 0 {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "No se encontraron domicilios que coincidan con los filtros proporcionados",
		}
		c.ServeJSON()
		return
	}

	// Responder con los domicilios filtrados
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Domicilios obtenidos exitosamente",
		Data:    domicilios,
	}
	c.ServeJSON()
}

// @Title GetById
// @Summary Obtener domicilio por ID (incluye cliente y pedido asociado si existen)
// @Description Devuelve un domicilio por ID y, si está asociado a un pedido, incluye documento/nombre del cliente y resumen del pedido (monto/productos).
// @Tags domicilios
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID del Domicilio"
// @Success 200 {object} models.ApiResponse "Domicilio encontrado (con cliente/pedido si aplica)"
// @Failure 400 {object} models.ApiResponse "Parámetro inválido"
// @Failure 404 {object} models.ApiResponse "Domicilio no encontrado"
// @Security BearerAuth
// @Router /domicilios/search [get]
func (c *DomicilioController) GetById() {
	o := orm.NewOrm()
	id, err := c.GetInt("id")
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

	// 1) Leer el domicilio
	domicilio := models.Domicilio{PK_ID_DOMICILIO: id}
	if err := o.Read(&domicilio); err == orm.ErrNoRows {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Domicilio no encontrado",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// 2) Cliente asociado (vía pedido) — usar UN struct para QueryRow
	type clienteRow struct {
		Documento int64  `orm:"column(documento)"`
		Nombre    string `orm:"column(nombre)"`
		Apellido  string `orm:"column(apellido)"`
	}
	var cli clienteRow

	qCliente := `
SELECT
  pc."PK_DOCUMENTO_CLIENTE" AS documento,
  c."NOMBRE"                AS nombre,
  c."APELLIDO"              AS apellido
FROM "PEDIDO" p
JOIN "PEDIDO_CLIENTE" pc ON pc."PK_ID_PEDIDO" = p."PK_ID_PEDIDO"
JOIN "CLIENTE" c        ON c."PK_DOCUMENTO_CLIENTE" = pc."PK_DOCUMENTO_CLIENTE"
WHERE p."PK_ID_DOMICILIO" = ?
ORDER BY p."PK_ID_PEDIDO" DESC
LIMIT 1;`

	cliErr := o.Raw(qCliente, id).QueryRow(&cli)

	// 3) Pedido asociado — también UN struct
	type pedidoRow struct {
		PedidoID          int64           `orm:"column(pedido_id)"`
		PagoID            sql.NullInt64   `orm:"column(pago_id)"`
		PagoMonto         sql.NullFloat64 `orm:"column(pago_monto)"`
		SubtotalProductos sql.NullFloat64 `orm:"column(subtotal_productos)"`
		Productos         string          `orm:"column(productos)"` // json string
	}
	var ped pedidoRow

	qPedido := `
SELECT
  p."PK_ID_PEDIDO" AS pedido_id,
  pa."PK_ID_PAGO"  AS pago_id,
  pa."MONTO"::numeric AS pago_monto,
  (
    SELECT COALESCE(SUM(d."SUBTOTAL"), 0)
    FROM "PRODUCTO_PEDIDO_DETALLE" d
    WHERE d."PK_ID_PEDIDO" = p."PK_ID_PEDIDO"
  ) AS subtotal_productos,
  (
    SELECT COALESCE(jsonb_agg(json_build_object(
      'PK_ID_PRODUCTO', d."PK_ID_PRODUCTO",
      'NOMBRE', pr."NOMBRE",
      'CANTIDAD', d."CANTIDAD",
      'PRECIO_UNITARIO', d."PRECIO_UNITARIO",
      'SUBTOTAL', d."SUBTOTAL"
    )), '[]'::jsonb)::text
    FROM "PRODUCTO_PEDIDO_DETALLE" d
    JOIN "PRODUCTO" pr ON pr."PK_ID_PRODUCTO" = d."PK_ID_PRODUCTO"
    WHERE d."PK_ID_PEDIDO" = p."PK_ID_PEDIDO"
  ) AS productos
FROM "PEDIDO" p
LEFT JOIN "PAGO" pa ON pa."PK_ID_PAGO" = p."PK_ID_PAGO"
WHERE p."PK_ID_DOMICILIO" = ?
ORDER BY p."PK_ID_PEDIDO" DESC
LIMIT 1;`

	pedErr := o.Raw(qPedido, id).QueryRow(&ped)

	// 4) Construir respuesta
	resp := map[string]interface{}{
		"domicilio": domicilio,
	}

	if cliErr == nil {
		resp["cliente"] = map[string]interface{}{
			"documento": cli.Documento,
			"nombre":    cli.Nombre,
			"apellido":  cli.Apellido,
		}
	}

	if pedErr == nil {
		// Parsear productos
		var productos []map[string]interface{}
		if ped.Productos != "" {
			_ = json.Unmarshal([]byte(ped.Productos), &productos)
		}

		// Total: prioriza pago; si no hay, usa subtotal de productos
		total := 0.0
		if ped.PagoMonto.Valid {
			total = ped.PagoMonto.Float64
		} else if ped.SubtotalProductos.Valid {
			total = ped.SubtotalProductos.Float64
		}

		var pagoIdPtr *int64
		if ped.PagoID.Valid {
			v := ped.PagoID.Int64
			pagoIdPtr = &v
		}

		resp["pedido"] = map[string]interface{}{
			"pedidoId":          ped.PedidoID,
			"pagoId":            pagoIdPtr,                     // puede ser null
			"montoPago":         ped.PagoMonto.Float64,         // 0 si null
			"subtotalProductos": ped.SubtotalProductos.Float64, // 0 si null
			"total":             total,
			"productos":         productos,
		}
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Domicilio encontrado",
		Data:    resp,
	}
	c.ServeJSON()
}

// @Title Create
// @Summary Crear un nuevo domicilio
// @Description Crea un nuevo domicilio en la base de datos.
// @Tags domicilios
// @Accept json
// @Produce json
// @Param   body  body   models.Domicilio true  "Datos del domicilio a crear"
// @Success 201 {object} models.Domicilio "Domicilio creado"
// @Failure 400 {object} models.ApiResponse "Error en la solicitud"
// @Security BearerAuth
// @Router /domicilios [post]
func (c *DomicilioController) Post() {
	o := orm.NewOrm()
	var input map[string]interface{}
	var domicilio models.Domicilio

	// Decodificar la solicitud
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &input); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Error al procesar la solicitud",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Validar y establecer los campos obligatorios
	if direccion, ok := input["direccion"].(string); ok && direccion != "" {
		domicilio.DIRECCION = direccion
	} else {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El campo 'DIRECCION' es obligatorio",
		}
		c.ServeJSON()
		return
	}
	// Validar fecha:
	if fechaStr, ok := input["fechaDomicilio"].(string); ok && fechaStr != "" {
		parsedDate, err := time.Parse("2006-01-02", fechaStr)
		if err != nil {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusBadRequest,
				Message: "Formato de fecha inválido",
				Cause:   err.Error(),
			}
			c.ServeJSON()
			return
		}
		domicilio.FECHA = parsedDate
	} else {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El campo FECHA no puede estar vacío",
		}
		c.ServeJSON()
		return
	}

	if telefono, ok := input["telefono"].(string); ok && telefono != "" {
		domicilio.TELEFONO = telefono
	} else {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El campo 'TELEFONO' es obligatorio",
		}
		c.ServeJSON()
		return
	}

	// Procesar campos opcionales
	if estadoPago, ok := input["estadoPago"].(string); ok {
		domicilio.ESTADO_PAGO = estadoPago
	}
	if entregado, ok := input["entregado"].(bool); ok {
		domicilio.ENTREGADO = entregado
	}
	if observaciones, ok := input["observaciones"].(string); ok {
		domicilio.OBSERVACIONES = observaciones
	}
	if createdBy, ok := input["createdBy"].(string); ok {
		domicilio.CREATED_BY = &createdBy
	}

	// Establecer valores automáticos
	domicilio.CREATED_AT = time.Now().UTC()
	domicilio.UPDATED_AT = time.Time{} // Inicializa vacío

	// Insertar en la base de datos
	_, err := o.Insert(&domicilio)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al crear el domicilio",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Responder con éxito
	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusCreated,
		Message: "Domicilio creado correctamente",
		Data:    domicilio,
	}
	c.ServeJSON()
}

// @Title Update
// @Summary Actualizar un domicilio
// @Description Actualiza los datos de un domicilio existente.
// @Tags domicilios
// @Accept json
// @Produce json
// @Param   id    query    int  true   "ID del Domicilio"
// @Param   body  body   models.Domicilio true  "Datos del domicilio a actualizar"
// @Success 200 {object} models.Domicilio "Domicilio actualizado"
// @Failure 404 {object} models.ApiResponse "Domicilio no encontrado"
// @Security BearerAuth
// @Router /domicilios [put]
func (c *DomicilioController) Put() {
	o := orm.NewOrm()

	// Obtener el ID del domicilio
	idStr := c.GetString("id")
	id, err := strconv.Atoi(idStr)
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

	// Buscar el domicilio por ID
	domicilio := models.Domicilio{PK_ID_DOMICILIO: id}
	if err := o.Read(&domicilio); err == orm.ErrNoRows {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Domicilio no encontrado",
		}
		c.ServeJSON()
		return
	}

	// Deserializar datos actualizados
	var input map[string]interface{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &input); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Error al procesar la solicitud",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Actualizar campos
	if direccion, ok := input["direccion"].(string); ok {
		domicilio.DIRECCION = direccion
	}
	if telefono, ok := input["telefono"].(string); ok {
		domicilio.TELEFONO = telefono
	}
	if estadoPago, ok := input["estadoPago"].(string); ok {
		domicilio.ESTADO_PAGO = estadoPago
	}
	if entregado, ok := input["entregado"].(bool); ok {
		domicilio.ENTREGADO = entregado
	}
	if updatedBy, ok := input["updatedBy"].(string); ok {
		domicilio.UPDATED_BY = &updatedBy
	}

	// Actualizar la fecha de modificación
	domicilio.UPDATED_AT = time.Now().UTC()

	// Guardar cambios
	if _, err := o.Update(&domicilio); err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al actualizar el domicilio",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Responder con éxito
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Domicilio actualizado correctamente",
		Data:    domicilio,
	}
	c.ServeJSON()
}

// @Title Delete
// @Summary Eliminar un domicilio
// @Description Elimina un domicilio de la base de datos.
// @Tags domicilios
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID del Domicilio"
// @Success 204 {object} nil "Domicilio eliminado"
// @Failure 404 {object} models.ApiResponse "Domicilio no encontrado"
// @Security BearerAuth
// @Router /domicilios [delete]
func (c *DomicilioController) Delete() {
	o := orm.NewOrm()

	idStr := c.GetString("id")
	id, err := strconv.Atoi(idStr)
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

	domicilio := models.Domicilio{PK_ID_DOMICILIO: id}

	if _, err := o.Delete(&domicilio); err == nil {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusOK,
			Message: "Domicilio eliminado",
		}
		c.ServeJSON()
	} else {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Domicilio no encontrado",
			Cause:   err.Error(),
		}
		c.ServeJSON()
	}
}

// @Title AsignarDomiciliario
// @Summary Asignar un domiciliario a un pedido de domicilio
// @Description Un domiciliario puede tomar un pedido si no ha sido asignado previamente
// @Tags domicilios
// @Accept json
// @Produce json
// @Param domicilio_id query int true "ID del domicilio"
// @Param trabajador_id query int true "ID del domiciliario que lo tomará"
// @Success 200 {object} models.ApiResponse "Domicilio asignado"
// @Failure 404 {object} models.ApiResponse "Domicilio no encontrado o ya asignado"
// @Failure 500 {object} models.ApiResponse "Error al asignar domicilio"
// @Security BearerAuth
// @Router /domicilios/asignar [post]
func (c *DomicilioController) AsignarDomiciliario() {
	domicilioID, _ := c.GetInt("domicilio_id")
	trabajadorID, _ := c.GetInt("trabajador_id")

	o := orm.NewOrm()

	// Buscar el domicilio
	domicilio := models.Domicilio{PK_ID_DOMICILIO: domicilioID}
	if err := o.Read(&domicilio); err != nil {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Domicilio no encontrado",
		}
		c.ServeJSON()
		return
	}

	// Verificar si ya está asignado
	if domicilio.PK_DOCUMENTO_TRABAJADOR != nil {
		c.Ctx.Output.SetStatus(http.StatusConflict)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusConflict,
			Message: "Este domicilio ya ha sido asignado",
		}
		c.ServeJSON()
		return
	}

	// Asignar el domiciliario
	domicilio.PK_DOCUMENTO_TRABAJADOR = &trabajadorID

	if _, err := o.Update(&domicilio, "PK_DOCUMENTO_TRABAJADOR"); err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al asignar domicilio",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Domicilio asignado correctamente",
		Data:    domicilio,
	}
	c.ServeJSON()
}
