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

type ReservaController struct {
	web.Controller
}

// Estados permitidos para la reserva
var estadosPermitidos = map[string]bool{
	"PENDIENTE":  true,
	"CONFIRMADA": true,
	"CANCELADA":  true,
	"CUMPLIDA":   true,
}

var location, _ = time.LoadLocation("America/Lima")

// @Title GetAll
// @Summary Obtener todas las reservas
// @Description Devuelve todas las reservas registradas en la base de datos.
// @Tags reservas
// @Accept json
// @Produce json
// @Success 200 {array} models.Reserva "Lista de reservas"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Router /reservas [get]
func (c *ReservaController) GetAll() {
	o := orm.NewOrm()
	var reservas []models.Reserva

	_, err := o.QueryTable(new(models.Reserva)).All(&reservas)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener reservas de la base de datos",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Ajustar las fechas y horas al formato y zona horaria correcta
	for i := range reservas {
		reservas[i].CREATED_AT = reservas[i].CREATED_AT.In(database.BogotaZone)
		reservas[i].UPDATED_AT = reservas[i].UPDATED_AT.In(database.BogotaZone)
		reservas[i].FECHA = reservas[i].FECHA.In(database.BogotaZone)

		if len(reservas[i].HORA) >= 8 {
			reservas[i].HORA = reservas[i].HORA[:8] // Asegurar formato HH:MM:SS
		}
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Reservas obtenidas exitosamente",
		Data:    reservas,
	}
	c.ServeJSON()
}

// @Title GetById
// @Summary Obtener reserva por ID
// @Description Devuelve una reserva específica por ID utilizando query parameters.
// @Tags reservas
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID de la Reserva"
// @Success 200 {object} models.Reserva "Reserva encontrada"
// @Failure 404 {object} models.ApiResponse "Reserva no encontrada"
// @Router /reservas/search [get]
func (c *ReservaController) GetById() {
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

	reserva := models.Reserva{PK_ID_RESERVA: id}

	err = o.Read(&reserva)
	if err == orm.ErrNoRows {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Reserva no encontrada",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}
	reserva.FECHA = reserva.FECHA.In(database.BogotaZone)
	reserva.CREATED_AT = reserva.CREATED_AT.In(database.BogotaZone)
	reserva.UPDATED_AT = reserva.UPDATED_AT.In(database.BogotaZone)
	if len(reserva.HORA) >= 8 {
		reserva.HORA = reserva.HORA[:8] // Formato HH:MM:SS
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Reserva encontrada",
		Data:    reserva,
	}
	c.ServeJSON()
}

// @Title Create
// @Summary Crear una nueva reserva
// @Description Crea una nueva reserva en la base de datos.
// @Tags reservas
// @Accept json
// @Produce json
// @Param   body  body   models.Reserva true  "Datos de la reserva a crear"
// @Success 201 {object} models.Reserva "Reserva creada"
// @Failure 400 {object} models.ApiResponse "Error en la solicitud"
// @Router /reservas [post]
func (c *ReservaController) Post() {
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
	var reserva models.Reserva

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
		reserva.FECHA = parsedDate
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
		_, err := time.Parse("15:04:05", horaStr)
		if err != nil {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusBadRequest,
				Message: "Formato de hora inválido",
				Cause:   err.Error(),
			}
			c.ServeJSON()
			return
		}
		reserva.HORA = horaStr
	} else {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El campo HORA no puede estar vacío",
		}
		c.ServeJSON()
		return
	}

	// Procesar PERSONAS
	if personas, ok := input["PERSONAS"].(float64); ok {
		reserva.PERSONAS = int(personas)
	} else {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El campo PERSONAS debe ser un número mayor a 0",
		}
		c.ServeJSON()
		return
	}

	// Procesar ESTADO_RESERVA si existe
	if estado, ok := input["ESTADO_RESERVA"].(string); ok && estado != "" {
		if !estadosPermitidos[estado] {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusBadRequest,
				Message: "Estado de reserva inválido",
				Cause:   "El estado debe ser uno de los siguientes: PENDIENTE, CONFIRMADA, CANCELADA, CUMPLIDA",
			}
			c.ServeJSON()
			return
		}
		reserva.ESTADO_RESERVA = &estado
	}

	// Procesar INDICACIONES si existe
	if indicaciones, ok := input["INDICACIONES"].(string); ok {
		reserva.INDICACIONES = &indicaciones
	}

	// Procesar CREATED_BY si existe
	if createdBy, ok := input["CREATED_BY"].(string); ok {
		reserva.CREATED_BY = &createdBy
	}

	// Establecer valores automáticos
	reserva.CREATED_AT = time.Now().UTC()
	reserva.UPDATED_AT = time.Time{}

	// Insertar en la base de datos
	_, err := o.Insert(&reserva)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al crear la reserva",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Responder con éxito
	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusCreated,
		Message: "Reserva creada correctamente",
		Data:    reserva,
	}
	c.ServeJSON()
}

// @Title Update
// @Summary Actualizar una reserva
// @Description Actualiza los datos de una reserva existente.
// @Tags reservas
// @Accept json
// @Produce json
// @Param   id    query    int  true   "ID de la Reserva"
// @Param   body  body   models.Reserva true  "Datos de la reserva a actualizar"
// @Success 200 {object} models.Reserva "Reserva actualizada"
// @Failure 404 {object} models.ApiResponse "Reserva no encontrada"
// @Router /reservas [put]
func (c *ReservaController) Put() {
	o := orm.NewOrm()

	// Obtener el ID de la reserva desde los parámetros
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

	// Buscar la reserva por ID
	reserva := models.Reserva{PK_ID_RESERVA: id}
	if err := o.Read(&reserva); err != nil {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Reserva no encontrada",
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

	// Validar y actualizar los campos que pueden cambiar
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
		reserva.FECHA = parsedDate
	}

	if horaStr, ok := input["HORA"].(string); ok && horaStr != "" {
		_, err := time.Parse("15:04:05", horaStr)
		if err != nil {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusBadRequest,
				Message: "Formato de hora inválido",
				Cause:   err.Error(),
			}
			c.ServeJSON()
			return
		}
		reserva.HORA = horaStr
	}

	if personas, ok := input["PERSONAS"].(float64); ok {
		reserva.PERSONAS = int(personas)
	}

	if estado, ok := input["ESTADO_RESERVA"].(string); ok && estadosPermitidos[estado] {
		reserva.ESTADO_RESERVA = &estado
	}

	if indicaciones, ok := input["INDICACIONES"].(string); ok {
		reserva.INDICACIONES = &indicaciones
	}

	if updatedBy, ok := input["UPDATED_BY"].(string); ok {
		reserva.UPDATED_BY = &updatedBy
	}

	// Actualizar la fecha de modificación
	reserva.UPDATED_AT = time.Now().UTC()

	// Actualizar los datos en la base de datos
	if _, err := o.Update(&reserva); err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al actualizar la reserva",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Responder con los datos actualizados
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Reserva actualizada",
		Data:    reserva,
	}
	c.ServeJSON()
}

// @Title Delete
// @Summary Cancelar una reserva
// @Description Actualiza el estado de una reserva a "CANCELADA".
// @Tags reservas
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID de la Reserva"
// @Success 200 {object} models.ApiResponse "Reserva cancelada"
// @Failure 404 {object} models.ApiResponse "Reserva no encontrada"
// @Router /reservas [delete]
func (c *ReservaController) Delete() {
	o := orm.NewOrm()

	// Obtener el ID de la reserva desde los parámetros
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

	// Buscar la reserva por ID
	reserva := models.Reserva{PK_ID_RESERVA: id}
	if err := o.Read(&reserva); err != nil {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Reserva no encontrada",
		}
		c.ServeJSON()
		return
	}

	// Actualizar el estado a CANCELADA
	estadoCancelada := "CANCELADA"
	reserva.ESTADO_RESERVA = &estadoCancelada
	reserva.UPDATED_AT = time.Now() // Actualizar la fecha de modificación

	// Guardar los cambios en la base de datos
	if _, err := o.Update(&reserva, "ESTADO_RESERVA", "UPDATED_AT"); err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al cancelar la reserva",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Responder con éxito
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Reserva cancelada correctamente",
		Data:    reserva,
	}
	c.ServeJSON()
}
