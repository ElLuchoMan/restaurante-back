package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
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

	for i := range reservas {
		adjustTimezoneFields(&reservas[i])
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

	adjustTimezoneFields(&reserva)

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
	var reserva models.Reserva

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

	// Validar y procesar los campos de entrada
	if err := validateAndParseFields(input, &reserva); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Error en los datos de entrada",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Establecer valores automáticos
	reserva.CREATED_AT = time.Now().UTC()
	reserva.UPDATED_AT = time.Time{}

	// Insertar en la base de datos
	if _, err := o.Insert(&reserva); err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al crear la reserva",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	adjustTimezoneFields(&reserva)

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

	// Guardar el valor original de CREATED_BY
	originalCreatedBy := reserva.CREATED_BY

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
	if err := validateAndParseFields(input, &reserva); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Error en los datos de entrada",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Restaurar el valor original de CREATED_BY
	reserva.CREATED_BY = originalCreatedBy

	// Actualizar la fecha de modificación
	reserva.UPDATED_AT = time.Now().UTC()

	// Actualizar los datos en la base de datos
	if _, err := o.Update(&reserva, "FECHA", "HORA", "PERSONAS", "ESTADO_RESERVA", "INDICACIONES", "UPDATED_AT", "UPDATED_BY"); err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al actualizar la reserva",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	adjustTimezoneFields(&reserva)

	// Responder con éxito
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Reserva actualizada correctamente",
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
	reserva.UPDATED_AT = time.Now().UTC()

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

	adjustTimezoneFields(&reserva)

	// Responder con éxito
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Reserva cancelada correctamente",
		Data:    reserva,
	}
	c.ServeJSON()
}

func adjustTimezoneFields(reserva *models.Reserva) {
	if !reserva.CREATED_AT.IsZero() {
		reserva.CREATED_AT = reserva.CREATED_AT.Local()
	}
	if !reserva.UPDATED_AT.IsZero() {
		reserva.UPDATED_AT = reserva.UPDATED_AT.Local()
	}
	if !reserva.FECHA.IsZero() {
		reserva.FECHA = reserva.FECHA.Local()
	}
	if len(reserva.HORA) >= 8 {
		reserva.HORA = reserva.HORA[:8] // Formato HH:MM:SS
	}
}

func validateAndParseFields(input map[string]interface{}, reserva *models.Reserva) error {
	// Validar y procesar FECHA
	if fechaStr, ok := input["FECHA"].(string); ok && fechaStr != "" {
		parsedDate, err := time.Parse("2006-01-02", fechaStr)
		if err != nil {
			return fmt.Errorf("formato de fecha inválido: %v", err)
		}
		reserva.FECHA = parsedDate
	}

	// Validar y procesar HORA
	if horaStr, ok := input["HORA"].(string); ok && horaStr != "" {
		_, err := time.Parse("15:04:05", horaStr)
		if err != nil {
			return fmt.Errorf("formato de hora inválido: %v", err)
		}
		reserva.HORA = horaStr
	}

	// Validar PERSONAS
	if personas, ok := input["PERSONAS"].(float64); ok {
		reserva.PERSONAS = int(personas)
	} else {
		return fmt.Errorf("el campo PERSONAS debe ser un número mayor a 0")
	}

	// Validar ESTADO_RESERVA
	if estado, ok := input["ESTADO_RESERVA"].(string); ok && estado != "" {
		if !estadosPermitidos[estado] {
			return fmt.Errorf("estado de reserva inválido")
		}
		reserva.ESTADO_RESERVA = &estado
	}

	// Procesar INDICACIONES
	if indicaciones, ok := input["INDICACIONES"].(string); ok {
		reserva.INDICACIONES = &indicaciones
	}

	// Procesar UPDATED_BY
	if updatedBy, ok := input["UPDATED_BY"].(string); ok {
		reserva.UPDATED_BY = &updatedBy
	}

	return nil
}
