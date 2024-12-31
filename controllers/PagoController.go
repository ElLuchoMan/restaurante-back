package controllers

import (
	"encoding/json"
	"net/http"
	"restaurante/database"
	"restaurante/models"
	"strconv"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

type PagoController struct {
	web.Controller
}

// Estados permitidos para los pagos
var estadosPagoPermitidos = map[string]bool{
	"PAGADO":    true,
	"PENDIENTE": true,
	"NO PAGO":   true,
}

// @Title GetAll
// @Summary Obtener todos los pagos con filtros
// @Description Devuelve todos los pagos registrados en la base de datos, con opción de filtrar por fecha exacta, mes, año y estado.
// @Tags pagos
// @Accept json
// @Produce json
// @Param   fecha    query   string   false   "Filtrar por fecha exacta (YYYY-MM-DD)"
// @Param   dia      query   int      false   "Filtrar por dia (1-31)"
// @Param   mes      query   int      false   "Filtrar por mes (1-12)"
// @Param   anio     query   int      false   "Filtrar por año (YYYY)"
// @Param   estado   query   string   false   "Filtrar por estado del pago (PAGADO, PENDIENTE, NO PAGO)"
// @Param   metodo_pago     query   int      false   "Filtrar por metodo de pago"
// @Success 200 {array} models.Pago "Lista de pagos"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /pagos [get]
func (c *PagoController) GetAll() {
	o := orm.NewOrm()
	var pagos []models.Pago

	_, err := o.QueryTable(new(models.Pago)).All(&pagos)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener pagos de la base de datos",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Ajustar fechas y hora al formato correcto
	for i := range pagos {
		pagos[i].UPDATED_AT = pagos[i].UPDATED_AT.In(database.BogotaZone)
		pagos[i].FECHA = pagos[i].FECHA.In(database.BogotaZone)

		// Formatear HORA si es necesario
		if len(pagos[i].HORA) >= 19 {
			pagos[i].HORA = pagos[i].HORA[11:19] // Solo toma HH:mm:ss
		}
	}

	// Leer parámetros de la URL
	fecha := c.GetString("fecha")
	dia, _ := c.GetInt("dia")
	mes, _ := c.GetInt("mes")
	anio, _ := c.GetInt("anio")
	estado := c.GetString("estado")
	metodo_pago, _ := c.GetInt("metodo_pago")

	// Filtrar los pagos según los parámetros proporcionados
	var filteredPagos []models.Pago
	for _, pago := range pagos {
		if fecha != "" && pago.FECHA.Format("2006-01-02") != fecha {
			continue
		}
		if dia > 0 && dia <= 31 && pago.FECHA.Day() != dia {
			continue
		}
		if mes > 0 && mes <= 12 && int(pago.FECHA.Month()) != mes {
			continue
		}
		if anio > 0 && pago.FECHA.Year() != anio {
			continue
		}
		if estado != "" && pago.ESTADO_PAGO != estado {
			continue
		}
		if metodo_pago > 0 && pago.PK_ID_METODO_PAGO != metodo_pago {
			continue
		}

		filteredPagos = append(filteredPagos, pago)
	}

	// Si no hay resultados
	if len(filteredPagos) == 0 {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "No se encontraron pagos que coincidan con los filtros proporcionados",
		}
		c.ServeJSON()
		return
	}

	// Respuesta con los pagos filtrados
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Pagos obtenidos exitosamente",
		Data:    filteredPagos,
	}
	c.ServeJSON()
}

// @Title GetById
// @Summary Obtener pago por ID
// @Description Devuelve un pago específico por ID.
// @Tags pagos
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID del Pago"
// @Success 200 {object} models.Pago "Pago encontrado"
// @Failure 404 {object} models.ApiResponse "Pago no encontrado"
// @Security BearerAuth
// @Router /pagos/search [get]
func (c *PagoController) GetById() {
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

	pago := models.Pago{PK_ID_PAGO: id}
	err = o.Read(&pago)
	if err == orm.ErrNoRows {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Pago no encontrado",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Ajustar fechas y hora
	pago.FECHA = pago.FECHA.In(database.BogotaZone)
	pago.UPDATED_AT = pago.UPDATED_AT.In(database.BogotaZone)

	// Formatear HORA
	if len(pago.HORA) >= 19 {
		pago.HORA = pago.HORA[11:19] // Formato HH:mm:ss
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Pago encontrado",
		Data:    pago,
	}
	c.ServeJSON()
}

// @Title Create
// @Summary Crear un nuevo pago
// @Description Crea un nuevo pago en la base de datos.
// @Tags pagos
// @Accept json
// @Produce json
// @Param   body  body   models.Pago true  "Datos del pago a crear"
// @Success 201 {object} models.Pago "Pago creado"
// @Failure 400 {object} models.ApiResponse "Error en la solicitud"
// @Security BearerAuth
// @Router /pagos [post]
func (c *PagoController) Post() {
	o := orm.NewOrm()
	var input map[string]interface{}

	// Decodificar la solicitud
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &input); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Error al decodificar la solicitud",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Validar y procesar los campos requeridos
	var pago models.Pago

	// Procesar FECHA
	if fechaStr, ok := input["FECHA"].(string); ok && fechaStr != "" {
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
		pago.FECHA = parsedDate
	} else {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El campo FECHA no puede estar vacío",
		}
		c.ServeJSON()
		return
	}

	// Procesar HORA
	if horaStr, ok := input["HORA"].(string); ok && horaStr != "" {
		// Validar el formato de HORA
		if _, err := time.Parse("15:04:05", horaStr); err != nil {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusBadRequest,
				Message: "Formato de hora inválido, debe ser HH:mm:ss",
				Cause:   err.Error(),
			}
			c.ServeJSON()
			return
		}
		pago.HORA = horaStr
	} else {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El campo HORA no puede estar vacío",
		}
		c.ServeJSON()
		return
	}

	// Validar y procesar MONTO
	if monto, ok := input["MONTO"].(float64); ok {
		pago.MONTO = int64(monto)
	} else {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El campo MONTO es obligatorio y debe ser un número",
		}
		c.ServeJSON()
		return
	}

	// Validar y procesar ESTADO_PAGO
	if estado, ok := input["ESTADO_PAGO"].(string); ok && estado != "" {
		if !estadosPagoPermitidos[estado] {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusBadRequest,
				Message: "Estado de pago inválido",
				Cause:   "El estado debe ser 'PAGADO', 'PENDIENTE' o 'NO PAGO'",
			}
			c.ServeJSON()
			return
		}
		pago.ESTADO_PAGO = estado
	}

	if pkMetodoPago, ok := input["PK_ID_METODO_PAGO"].(float64); ok {
		valorMetodoPago := int(pkMetodoPago)     // Convertir a int
		pago.PK_ID_METODO_PAGO = valorMetodoPago // Asignar el valor directamente
	} else {
		// Opcional: Manejo de errores o acciones si el campo es obligatorio
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El campo PK_ID_METODO_PAGO es obligatorio y debe ser un número válido",
		}
		c.ServeJSON()
		return
	}

	// Insertar en la base de datos
	_, err := o.Insert(&pago)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al crear el pago",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusCreated,
		Message: "Pago creado correctamente",
		Data:    pago,
	}
	c.ServeJSON()
}

// @Title Update
// @Summary Actualizar un pago
// @Description Actualiza los datos de un pago existente.
// @Tags pagos
// @Accept json
// @Produce json
// @Param   id    query    int  true   "ID del Pago"
// @Param   body  body   models.Pago true  "Datos del pago a actualizar"
// @Success 200 {object} models.Pago "Pago actualizado"
// @Failure 404 {object} models.ApiResponse "Pago no encontrado"
// @Security BearerAuth
// @Router /pagos [put]
func (c *PagoController) Put() {
	o := orm.NewOrm()

	// Obtener el ID del pago desde los parámetros
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

	// Buscar el pago por ID
	pago := models.Pago{PK_ID_PAGO: id}
	if err := o.Read(&pago); err == orm.ErrNoRows {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Pago no encontrado",
		}
		c.ServeJSON()
		return
	}

	// Deserializar los datos actualizados desde el cuerpo de la solicitud
	var input map[string]interface{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &input); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Error al decodificar la solicitud",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Validar y actualizar los campos
	if fechaStr, ok := input["FECHA"].(string); ok && fechaStr != "" {
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
		pago.FECHA = parsedDate
	}

	// Procesar HORA
	if horaStr, ok := input["HORA"].(string); ok && horaStr != "" {
		// Validar el formato de HORA
		if _, err := time.Parse("15:04:05", horaStr); err != nil {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusBadRequest,
				Message: "Formato de hora inválido, debe ser HH:mm:ss",
				Cause:   err.Error(),
			}
			c.ServeJSON()
			return
		}
		pago.HORA = horaStr
	} else {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El campo HORA no puede estar vacío",
		}
		c.ServeJSON()
		return
	}

	if monto, ok := input["MONTO"].(float64); ok {
		pago.MONTO = int64(monto)
	}

	if estado, ok := input["ESTADO_PAGO"].(string); ok && estado != "" {
		if !estadosPagoPermitidos[estado] {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusBadRequest,
				Message: "Estado de pago inválido. Debe ser 'PAGADO', 'PENDIENTE' o 'NO PAGO'",
			}
			c.ServeJSON()
			return
		}
		pago.ESTADO_PAGO = estado
	}

	if updatedBy, ok := input["UPDATED_BY"].(string); ok {
		pago.UPDATED_BY = updatedBy
	}

	// Actualizar la fecha de modificación
	pago.UPDATED_AT = time.Now().UTC()

	if pkMetodoPago, ok := input["PK_ID_METODO_PAGO"].(float64); ok {
		valorMetodoPago := int(pkMetodoPago)     // Convertir a int
		pago.PK_ID_METODO_PAGO = valorMetodoPago // Asignar al puntero
	} else {
		// Opcional: Manejo de errores o acciones si el campo es obligatorio
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El campo PK_ID_METODO_PAGO es obligatorio y debe ser un número válido",
		}
		c.ServeJSON()
		return
	}

	// Guardar los cambios en la base de datos
	if _, err := o.Update(&pago); err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al actualizar el pago",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Responder con los datos actualizados
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Pago actualizado correctamente",
		Data:    pago,
	}
	c.ServeJSON()
}

// @Title Delete
// @Summary Eliminar un pago
// @Description Elimina un pago de la base de datos.
// @Tags pagos
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID del Pago"
// @Success 200 {object} models.ApiResponse "Pago eliminado"
// @Failure 404 {object} models.ApiResponse "Pago no encontrado"
// @Security BearerAuth
// @Router /pagos [delete]
func (c *PagoController) Delete() {
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

	pago := models.Pago{PK_ID_PAGO: id}

	if _, err := o.Delete(&pago); err == nil {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusOK,
			Message: "Pago eliminado",
		}
		c.ServeJSON()
	} else {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Pago no encontrado",
			Cause:   err.Error(),
		}
		c.ServeJSON()
	}
}
