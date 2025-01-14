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

type CambiosHorarioController struct {
	web.Controller
}

// @Title GetAll
// @Summary Obtener todos los cambios de horario
// @Description Obtiene un listado de todos los cambios de horario registrados en la base de datos
// @Tags cambios_horario
// @Accept json
// @Produce json
// @Success 200 {array} map[string]interface{} "Listado de cambios de horario"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /cambios_horario [get]
func (c *CambiosHorarioController) GetAll() {
	o := orm.NewOrm()
	var horarios []models.CambiosHorario

	_, err := o.QueryTable(new(models.CambiosHorario)).All(&horarios)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener cambios de horario",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Preparar respuesta con horas formateadas
	var response []map[string]interface{}
	for _, horario := range horarios {
		h := map[string]interface{}{
			"PK_ID_CAMBIO_HORARIO": horario.PK_ID_CAMBIO_HORARIO,
			"FECHA":                horario.FECHA.Format("2006-01-02"),
			"ABIERTO":              horario.ABIERTO,
		}
		if horario.HORA_APERTURA != nil {
			h["HORA_APERTURA"] = horario.HORA_APERTURA.Format("15:04:05")
		}
		if horario.HORA_CIERRE != nil {
			h["HORA_CIERRE"] = horario.HORA_CIERRE.Format("15:04:05")
		}
		response = append(response, h)
	}

	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Cambios de horario obtenidos correctamente",
		Data:    response,
	}
	c.ServeJSON()
}

// @Title GetByCurrentDate
// @Summary Consultar cambios de horario para la fecha actual
// @Description Obtiene el cambio de horario que aplica para la fecha actual, si existe.
// @Tags cambios_horario
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Cambio de horario para la fecha actual"
// @Failure 404 {object} models.ApiResponse "No hay cambios de horario para la fecha actual"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /cambios_horario/actual [get]
func (c *CambiosHorarioController) GetByCurrentDate() {
	o := orm.NewOrm()
	var cambioHorario models.CambiosHorario

	// Obtener la fecha actual
	currentDate := time.Now().In(database.BogotaZone)

	// Consultar si hay un cambio de horario para la fecha actual
	err := o.QueryTable(new(models.CambiosHorario)).
		Filter("FECHA", currentDate.Format("2006-01-02")).
		One(&cambioHorario)

	if err == orm.ErrNoRows {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "No hay cambios de horario para la fecha actual",
		}
		c.ServeJSON()
		return
	} else if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al consultar cambios de horario",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Preparar la respuesta con horas formateadas
	response := map[string]interface{}{
		"PK_ID_CAMBIO_HORARIO": cambioHorario.PK_ID_CAMBIO_HORARIO,
		"FECHA":                cambioHorario.FECHA.Format("2006-01-02"),
		"ABIERTO":              cambioHorario.ABIERTO,
	}
	if cambioHorario.HORA_APERTURA != nil {
		response["HORA_APERTURA"] = cambioHorario.HORA_APERTURA.Format("15:04:05")
	}
	if cambioHorario.HORA_CIERRE != nil {
		response["HORA_CIERRE"] = cambioHorario.HORA_CIERRE.Format("15:04:05")
	}

	// Respuesta con el cambio de horario encontrado
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Cambio de horario encontrado para la fecha actual",
		Data:    response,
	}
	c.ServeJSON()
}

// @Title Post
// @Summary Crear un nuevo cambio de horario
// @Description Crea un nuevo cambio de horario en la base de datos.
// @Tags cambios_horario
// @Accept json
// @Produce json
// @Param body body models.CambiosHorario true "Datos del cambio de horario"
// @Success 201 {object} map[string]interface{} "Cambio de horario creado"
// @Failure 400 {object} models.ApiResponse "Error en la solicitud"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /cambios_horario [post]
func (c *CambiosHorarioController) Post() {
	o := orm.NewOrm()
	var input map[string]interface{}
	var horario models.CambiosHorario

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

	// Validar y procesar FECHA
	if fechaStr, ok := input["FECHA"].(string); ok && fechaStr != "" {
		parsedDate, err := time.Parse("2006-01-02", fechaStr)
		if err != nil {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusBadRequest,
				Message: "Formato de fecha inválido para FECHA",
				Cause:   err.Error(),
			}
			c.ServeJSON()
			return
		}
		horario.FECHA = parsedDate
	} else {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El campo FECHA es obligatorio",
		}
		c.ServeJSON()
		return
	}

	// Validar ABIERTO
	if abierto, ok := input["ABIERTO"].(bool); ok {
		horario.ABIERTO = abierto
	} else {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El campo ABIERTO es obligatorio",
		}
		c.ServeJSON()
		return
	}

	// Configuración automática para días cerrados
	if !horario.ABIERTO {
		// Si el restaurante no abre, establecer horas predeterminadas
		horaApertura, _ := time.Parse("15:04:05", "00:00:00")
		horaCierre, _ := time.Parse("15:04:05", "23:59:59")
		horario.HORA_APERTURA = &horaApertura
		horario.HORA_CIERRE = &horaCierre
	} else {
		// Validar y procesar HORA_APERTURA (opcional)
		if horaAperturaStr, ok := input["HORA_APERTURA"].(string); ok && horaAperturaStr != "" {
			parsedHora, err := time.Parse("15:04:05", horaAperturaStr)
			if err != nil {
				c.Ctx.Output.SetStatus(http.StatusBadRequest)
				c.Data["json"] = models.ApiResponse{
					Code:    http.StatusBadRequest,
					Message: "Formato de hora inválido para HORA_APERTURA",
					Cause:   err.Error(),
				}
				c.ServeJSON()
				return
			}
			horario.HORA_APERTURA = &parsedHora
		} else {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusBadRequest,
				Message: "El campo HORA_APERTURA es obligatorio cuando ABIERTO es true",
			}
			c.ServeJSON()
			return
		}

		// Validar y procesar HORA_CIERRE
		if horaCierreStr, ok := input["HORA_CIERRE"].(string); ok && horaCierreStr != "" {
			parsedHora, err := time.Parse("15:04:05", horaCierreStr)
			if err != nil {
				c.Ctx.Output.SetStatus(http.StatusBadRequest)
				c.Data["json"] = models.ApiResponse{
					Code:    http.StatusBadRequest,
					Message: "Formato de hora inválido para HORA_CIERRE",
					Cause:   err.Error(),
				}
				c.ServeJSON()
				return
			}
			horario.HORA_CIERRE = &parsedHora
		} else {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusBadRequest,
				Message: "El campo HORA_CIERRE es obligatorio cuando ABIERTO es true",
			}
			c.ServeJSON()
			return
		}
	}

	// Insertar en la base de datos
	_, err := o.Insert(&horario)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al crear el cambio de horario",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Preparar la respuesta con el formato deseado
	response := map[string]interface{}{
		"PK_ID_CAMBIO_HORARIO": horario.PK_ID_CAMBIO_HORARIO,
		"FECHA":                horario.FECHA.Format("2006-01-02"),
		"ABIERTO":              horario.ABIERTO,
	}
	if horario.HORA_APERTURA != nil {
		response["HORA_APERTURA"] = horario.HORA_APERTURA.Format("15:04:05")
	}
	if horario.HORA_CIERRE != nil {
		response["HORA_CIERRE"] = horario.HORA_CIERRE.Format("15:04:05")
	}

	// Responder con éxito
	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusCreated,
		Message: "Cambio de horario creado correctamente",
		Data:    response,
	}
	c.ServeJSON()
}

// @Title Update
// @Summary Actualizar un cambio de horario
// @Description Actualiza los datos de un cambio de horario existente.
// @Tags cambios_horario
// @Accept json
// @Produce json
// @Param id query int true "ID del cambio de horario"
// @Param body body models.CambiosHorario true "Datos del cambio de horario a actualizar"
// @Success 200 {object} map[string]interface{} "Cambio de horario actualizado"
// @Failure 400 {object} models.ApiResponse "Error en la solicitud"
// @Failure 404 {object} models.ApiResponse "Cambio de horario no encontrado"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /cambios_horario [put]
func (c *CambiosHorarioController) Put() {
	o := orm.NewOrm()
	id, err := c.GetInt64("id")
	if err != nil || id == 0 {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "ID inválido o ausente",
		}
		c.ServeJSON()
		return
	}

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

	// Buscar el cambio de horario por ID
	var horario models.CambiosHorario
	if err := o.QueryTable(new(models.CambiosHorario)).Filter("PK_ID_CAMBIO_HORARIO", id).One(&horario); err == orm.ErrNoRows {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Cambio de horario no encontrado",
		}
		c.ServeJSON()
		return
	} else if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al buscar el cambio de horario",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Validar y actualizar campos
	if fechaStr, ok := input["FECHA"].(string); ok && fechaStr != "" {
		parsedDate, err := time.Parse("2006-01-02", fechaStr)
		if err != nil {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusBadRequest,
				Message: "Formato de fecha inválido para FECHA",
				Cause:   err.Error(),
			}
			c.ServeJSON()
			return
		}
		horario.FECHA = parsedDate
	}

	if abierto, ok := input["ABIERTO"].(bool); ok {
		horario.ABIERTO = abierto
		if !abierto {
			horaApertura, _ := time.Parse("15:04:05", "00:00:00")
			horaCierre, _ := time.Parse("15:04:05", "23:59:59")
			horario.HORA_APERTURA = &horaApertura
			horario.HORA_CIERRE = &horaCierre
		}
	}

	if horario.ABIERTO {
		if horaAperturaStr, ok := input["HORA_APERTURA"].(string); ok && horaAperturaStr != "" {
			parsedHora, err := time.Parse("15:04:05", horaAperturaStr)
			if err != nil {
				c.Ctx.Output.SetStatus(http.StatusBadRequest)
				c.Data["json"] = models.ApiResponse{
					Code:    http.StatusBadRequest,
					Message: "Formato de hora inválido para HORA_APERTURA",
					Cause:   err.Error(),
				}
				c.ServeJSON()
				return
			}
			horario.HORA_APERTURA = &parsedHora
		}

		if horaCierreStr, ok := input["HORA_CIERRE"].(string); ok && horaCierreStr != "" {
			parsedHora, err := time.Parse("15:04:05", horaCierreStr)
			if err != nil {
				c.Ctx.Output.SetStatus(http.StatusBadRequest)
				c.Data["json"] = models.ApiResponse{
					Code:    http.StatusBadRequest,
					Message: "Formato de hora inválido para HORA_CIERRE",
					Cause:   err.Error(),
				}
				c.ServeJSON()
				return
			}
			horario.HORA_CIERRE = &parsedHora
		}
	}

	// Guardar los cambios
	if _, err := o.Update(&horario); err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al actualizar el cambio de horario",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Preparar la respuesta
	response := map[string]interface{}{
		"PK_ID_CAMBIO_HORARIO": horario.PK_ID_CAMBIO_HORARIO,
		"FECHA":                horario.FECHA.Format("2006-01-02"),
		"ABIERTO":              horario.ABIERTO,
	}
	if horario.HORA_APERTURA != nil {
		response["HORA_APERTURA"] = horario.HORA_APERTURA.Format("15:04:05")
	}
	if horario.HORA_CIERRE != nil {
		response["HORA_CIERRE"] = horario.HORA_CIERRE.Format("15:04:05")
	}

	// Responder con éxito
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Cambio de horario actualizado correctamente",
		Data:    response,
	}
	c.ServeJSON()
}

// @Title Delete
// @Summary Eliminar un cambio de horario
// @Description Elimina un cambio de horario de la base de datos.
// @Tags cambios_horario
// @Accept json
// @Produce json
// @Param id query int true "ID del cambio de horario"
// @Success 200 {object} models.ApiResponse "Cambio de horario eliminado"
// @Failure 404 {object} models.ApiResponse "Cambio de horario no encontrado"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /cambios_horario [delete]
func (c *CambiosHorarioController) Delete() {
	o := orm.NewOrm()
	id, err := c.GetInt64("id")
	if err != nil || id == 0 {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "ID inválido o ausente",
		}
		c.ServeJSON()
		return
	}

	// Eliminar el cambio de horario
	if num, err := o.QueryTable(new(models.CambiosHorario)).
		Filter("PK_ID_CAMBIO_HORARIO", id).
		Delete(); err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al eliminar el cambio de horario",
			Cause:   err.Error(),
		}
		c.ServeJSON()
	} else if num == 0 {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Cambio de horario no encontrado",
		}
		c.ServeJSON()
	} else {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusOK,
			Message: "Cambio de horario eliminado correctamente",
		}
		c.ServeJSON()
	}
}
