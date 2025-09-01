package controllers

import (
	"encoding/json"
	"net/http"
	"restaurante/database"
	"restaurante/models"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

type ReservaController struct {
	web.Controller
}

// Estados permitidos para la reserva
var estadosPermitidos = map[models.EstadoReserva]bool{
	models.EstadoReservaPendiente:  true,
	models.EstadoReservaConfirmada: true,
	models.EstadoReservaCancelada:  true,
	models.EstadoReservaCumplida:   true,
}

var queryAllReservas = func(o orm.Ormer, reservas *[]models.Reserva) (int64, error) {
	return o.QueryTable(new(models.Reserva)).All(reservas)
}

var readReserva = func(o orm.Ormer, r *models.Reserva) error {
	return o.Read(r)
}

var insertReserva = func(o orm.Ormer, r *models.Reserva) (int64, error) {
	return o.Insert(r)
}

var updateReserva = func(o orm.Ormer, r *models.Reserva, cols ...string) (int64, error) {
	return o.Update(r, cols...)
}

var queryReservasByParam = func(o orm.Ormer, contactoID int64, fecha time.Time, useContacto, useFecha bool, reservas *[]models.Reserva) (int64, error) {
	qs := o.QueryTable(new(models.Reserva))
	if useContacto {
		qs = qs.Filter("PK_ID_CONTACTO", contactoID)
	}
	if useFecha {
		qs = qs.Filter("FECHA", fecha)
	}
	return qs.All(reservas)
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
	o := ormNew()
	var reservas []models.Reserva

	_, err := queryAllReservas(o, &reservas)
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
		reservas[i].FECHA = reservas[i].FECHA.UTC()
		reservas[i].HORA = reservas[i].HORA.UTC()
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
	o := ormNew()
	id, err := c.GetInt64("id")

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

	err = readReserva(o, &reserva)
	if err == orm.ErrNoRows {
		c.Ctx.Output.SetStatus(http.StatusOK)
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
	reserva.HORA = reserva.HORA.UTC()

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
	o := ormNew()
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
	if fechaStr, ok := input["fechaReserva"].(string); ok && fechaStr != "" {
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
	if horaStr, ok := input["horaReserva"].(string); ok && horaStr != "" {
		parsedHora, err := time.Parse("15:04:05", horaStr)
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
		reserva.HORA = parsedHora
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
	if personas, ok := input["personas"].(float64); ok {
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
	if estadoStr, ok := input["estadoReserva"].(string); ok {
		estado := models.EstadoReserva(estadoStr)
		if estado != "" {
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
			reserva.ESTADO_RESERVA = estado
		}
	}
	if reserva.ESTADO_RESERVA == "" {
		reserva.ESTADO_RESERVA = models.EstadoReservaPendiente
	}

	// Procesar INDICACIONES si existe
	if indicaciones, ok := input["indicaciones"].(string); ok {
		reserva.INDICACIONES = &indicaciones
	}

	// Procesar CREATED_BY si existe
	if createdBy, ok := input["createdBy"].(string); ok {
		reserva.CREATED_BY = &createdBy
	}

	if contactoID, ok := input["contactoId"].(float64); ok {
		reserva.PK_ID_CONTACTO = int64(contactoID)
	} else {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El campo PK_ID_CONTACTO debe ser un número",
		}
		c.ServeJSON()
		return
	}
	if restauranteID, ok := input["restauranteId"].(float64); ok {
		reserva.PK_ID_RESTAURANTE = int64(restauranteID)
	} else {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El campo PK_ID_RESTAURANTE debe ser un número",
		}
		c.ServeJSON()
		return
	}

	// Establecer valores automáticos
	reserva.CREATED_AT = time.Now().UTC()
	reserva.UPDATED_AT = time.Time{}

	// Insertar en la base de datos
	_, err := insertReserva(o, &reserva)
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
	o := ormNew()

	// Obtener el ID de la reserva desde los parámetros
	id, err := c.GetInt64("id")
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
	if err := readReserva(o, &reserva); err != nil {
		c.Ctx.Output.SetStatus(http.StatusOK)
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
	if fechaStr, ok := input["fechaReserva"].(string); ok && fechaStr != "" {
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

	if horaStr, ok := input["horaReserva"].(string); ok && horaStr != "" {
		parsedHora, err := time.Parse("15:04:05", horaStr)
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
		reserva.HORA = parsedHora
	}

	if personas, ok := input["personas"].(float64); ok {
		reserva.PERSONAS = int(personas)
	}

	if estadoStr, ok := input["estadoReserva"].(string); ok {
		estado := models.EstadoReserva(estadoStr)
		if estadosPermitidos[estado] {
			reserva.ESTADO_RESERVA = estado
		}
	}

	if indicaciones, ok := input["indicaciones"].(string); ok {
		reserva.INDICACIONES = &indicaciones
	}

	if updatedBy, ok := input["updatedBy"].(string); ok {
		reserva.UPDATED_BY = &updatedBy
	}

	if contactoID, ok := input["contactoId"].(float64); ok {
		reserva.PK_ID_CONTACTO = int64(contactoID)
	} else {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El campo PK_ID_CONTACTO debe ser un número",
		}
		c.ServeJSON()
		return
	}
	if restauranteID, ok := input["restauranteId"].(float64); ok {
		reserva.PK_ID_RESTAURANTE = int64(restauranteID)
	} else {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El campo PK_ID_RESTAURANTE debe ser un número",
		}
		c.ServeJSON()
		return
	}

	// Actualizar la fecha de modificación
	reserva.UPDATED_AT = time.Now().UTC()

	// Actualizar los datos en la base de datos
	if _, err := updateReserva(o, &reserva); err != nil {
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

// @Title GetByCliente
// @Summary Obtener reservas por contacto y/o fecha
// @Description Devuelve las reservas asociadas a un contacto en una fecha específica, todas sus reservas si no se especifica la fecha, o todas las reservas en una fecha específica si no se especifica el contacto.
// @Tags reservas
// @Accept json
// @Produce json
// @Param contactoId query int false "ID del Contacto (Opcional)"
// @Param fecha query string false "Fecha de la reserva (YYYY-MM-DD) (Opcional)"
// @Success 200 {array} models.Reserva "Lista de reservas encontradas"
// @Failure 400 {object} models.ApiResponse "Error en los parámetros"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Router /reservas/parameter [get]
func (c *ReservaController) GetByParameter() {
	o := ormNew()
	var reservas []models.Reserva

	contactoID, errContacto := c.GetInt64("contactoId")
	fechaReserva := c.GetString("fecha")

	useContacto := errContacto == nil && contactoID != 0
	var parsedDate time.Time
	useFecha := false
	if fechaReserva != "" {
		var err error
		parsedDate, err = time.Parse("2006-01-02", fechaReserva)
		if err != nil {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusBadRequest,
				Message: "El parámetro 'fecha' debe tener el formato YYYY-MM-DD",
			}
			c.ServeJSON()
			return
		}
		useFecha = true
	}

	_, err := queryReservasByParam(o, contactoID, parsedDate, useContacto, useFecha, &reservas)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener reservas",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Ajustar fechas y formato de hora antes de responder
	for i := range reservas {
		reservas[i].CREATED_AT = reservas[i].CREATED_AT.In(database.BogotaZone)
		reservas[i].UPDATED_AT = reservas[i].UPDATED_AT.In(database.BogotaZone)
		reservas[i].FECHA = reservas[i].FECHA.UTC()
		reservas[i].HORA = reservas[i].HORA.UTC()
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Reservas obtenidas exitosamente",
		Data:    reservas,
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
	o := ormNew()

	// Obtener el ID de la reserva desde los parámetros
	id, err := c.GetInt64("id")
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
	if err := readReserva(o, &reserva); err != nil {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Reserva no encontrada",
		}
		c.ServeJSON()
		return
	}

	// Actualizar el estado a CANCELADA
	estadoCancelada := models.EstadoReservaCancelada
	reserva.ESTADO_RESERVA = estadoCancelada
	reserva.UPDATED_AT = time.Now() // Actualizar la fecha de modificación

	// Guardar los cambios en la base de datos
	if _, err := updateReserva(o, &reserva, "estadoReserva", "updatedAt"); err != nil {
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
