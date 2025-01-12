package controllers

import (
	"encoding/json"
	"net/http"
	"restaurante/models"
	"strconv"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

type IncidenciaController struct {
	web.Controller
}

// @Title GetAll
// @Summary Obtener todas las incidencias
// @Description Devuelve una lista de todas las incidencias registradas en la base de datos.
// @Tags incidencias
// @Accept json
// @Produce json
// @Success 200 {array} models.Incidencia "Lista de incidencias"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /incidencias [get]
func (c *IncidenciaController) GetAll() {
	o := orm.NewOrm()
	var incidencias []models.Incidencia

	_, err := o.QueryTable(new(models.Incidencia)).All(&incidencias)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener incidencias",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Incidencias obtenidas correctamente",
		Data:    incidencias,
	}
	c.ServeJSON()
}

// @Title GetByDocumentAndDate
// @Summary Obtener incidencias por documento del trabajador y fecha
// @Description Devuelve las incidencias de un trabajador en un mes y año específico.
// @Tags incidencias
// @Accept json
// @Produce json
// @Param   documento     query    int     true   "Documento del Trabajador"
// @Param   mes           query    int     true   "Mes de la Incidencia (1-12)"
// @Param   anio          query    int     true   "Año de la Incidencia"
// @Success 200 {array} models.Incidencia "Lista de incidencias encontradas"
// @Failure 400 {object} models.ApiResponse "Error en la solicitud"
// @Failure 404 {object} models.ApiResponse "No se encontraron incidencias"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /incidencias/search [get]
func (c *IncidenciaController) GetByDocumentAndDate() {
	o := orm.NewOrm()

	// Obtener parámetros de la consulta
	documento, err := c.GetInt64("documento")
	if err != nil || documento == 0 {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El parámetro 'documento' es inválido o ausente",
		}
		c.ServeJSON()
		return
	}

	mes, err := c.GetInt("mes")
	if err != nil || mes < 1 || mes > 12 {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El parámetro 'mes' es inválido. Debe estar entre 1 y 12",
		}
		c.ServeJSON()
		return
	}

	anio, err := c.GetInt("anio")
	if err != nil || anio < 1900 || anio > time.Now().Year() {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El parámetro 'anio' es inválido o ausente",
		}
		c.ServeJSON()
		return
	}

	// Calcular rango de fechas
	fechaInicio := time.Date(anio, time.Month(mes), 1, 0, 0, 0, 0, time.UTC)
	fechaFin := fechaInicio.AddDate(0, 1, 0).Add(-time.Second) // Fin del mes

	// Filtrar las incidencias
	var incidencias []models.Incidencia
	_, err = o.QueryTable(new(models.Incidencia)).
		Filter("PK_DOCUMENTO_TRABAJADOR", documento).
		Filter("FECHA__gte", fechaInicio.Format("2006-01-02")).
		Filter("FECHA__lte", fechaFin.Format("2006-01-02")).
		All(&incidencias)

	if err == orm.ErrNoRows || len(incidencias) == 0 {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "No se encontraron incidencias para los parámetros proporcionados",
		}
		c.ServeJSON()
		return
	} else if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al buscar incidencias",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Responder con las incidencias encontradas
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Incidencias encontradas",
		Data:    incidencias,
	}
	c.ServeJSON()
}

// @Title Post
// @Summary Crear una nueva incidencia
// @Description Crea una nueva incidencia en la base de datos.
// @Tags incidencias
// @Accept json
// @Produce json
// @Param body body models.Incidencia true "Datos de la incidencia"
// @Success 201 {object} map[string]interface{} "Incidencia creada"
// @Failure 400 {object} models.ApiResponse "Error en la solicitud"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /incidencias [post]
func (c *IncidenciaController) Post() {
	o := orm.NewOrm()
	var input map[string]interface{}
	var incidencia models.Incidencia

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
		incidencia.FECHA = parsedDate
	} else {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El campo FECHA es obligatorio",
		}
		c.ServeJSON()
		return
	}

	// Validar y procesar MONTO
	if monto, ok := input["MONTO"].(float64); ok {
		incidencia.MONTO = int64(monto)
	} else {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El campo MONTO es obligatorio y debe ser un número",
		}
		c.ServeJSON()
		return
	}

	// Validar y procesar RESTA
	if resta, ok := input["RESTA"].(bool); ok {
		incidencia.RESTA = resta
	} else {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El campo RESTA es obligatorio",
		}
		c.ServeJSON()
		return
	}

	// Validar y procesar MOTIVO
	if motivo, ok := input["MOTIVO"].(string); ok && motivo != "" {
		incidencia.MOTIVO = motivo
	} else {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El campo MOTIVO es obligatorio",
		}
		c.ServeJSON()
		return
	}

	// Procesar PK_DOCUMENTO_TRABAJADOR (opcional)
	if documento, ok := input["PK_DOCUMENTO_TRABAJADOR"].(float64); ok {
		doc := int64(documento)
		incidencia.PK_DOCUMENTO_TRABAJADOR = &doc
	}

	// Insertar en la base de datos
	_, err := o.Insert(&incidencia)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al crear la incidencia",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Preparar la respuesta con el formato deseado
	response := map[string]interface{}{
		"PK_ID_INCIDENCIA":        incidencia.PK_ID_INCIDENCIA,
		"FECHA":                   incidencia.FECHA.Format("2006-01-02"),
		"MONTO":                   incidencia.MONTO,
		"RESTA":                   incidencia.RESTA,
		"MOTIVO":                  incidencia.MOTIVO,
		"PK_DOCUMENTO_TRABAJADOR": incidencia.PK_DOCUMENTO_TRABAJADOR,
	}

	// Responder con éxito
	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusCreated,
		Message: "Incidencia creada correctamente",
		Data:    response,
	}
	c.ServeJSON()
}

// @Title Update
// @Summary Actualizar una incidencia
// @Description Actualiza los datos de una incidencia existente en la base de datos.
// @Tags incidencias
// @Accept json
// @Produce json
// @Param id query int true "ID de la Incidencia"
// @Param body body models.Incidencia true "Datos de la incidencia a actualizar"
// @Success 200 {object} map[string]interface{} "Incidencia actualizada"
// @Failure 400 {object} models.ApiResponse "Error en la solicitud"
// @Failure 404 {object} models.ApiResponse "Incidencia no encontrada"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /incidencias [put]
func (c *IncidenciaController) Put() {
	o := orm.NewOrm()

	// Obtener el ID de la incidencia desde los parámetros
	idStr := c.GetString("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id == 0 {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El parámetro 'id' es inválido o está ausente",
		}
		c.ServeJSON()
		return
	}

	// Buscar la incidencia por ID
	incidencia := models.Incidencia{PK_ID_INCIDENCIA: int64(id)}
	if err := o.Read(&incidencia); err == orm.ErrNoRows {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Incidencia no encontrada",
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
				Message: "Formato de fecha inválido para FECHA",
				Cause:   err.Error(),
			}
			c.ServeJSON()
			return
		}
		incidencia.FECHA = parsedDate
	}

	if monto, ok := input["MONTO"].(float64); ok {
		incidencia.MONTO = int64(monto)
	}

	if resta, ok := input["RESTA"].(bool); ok {
		incidencia.RESTA = resta
	}

	if motivo, ok := input["MOTIVO"].(string); ok && motivo != "" {
		incidencia.MOTIVO = motivo
	}

	if documento, ok := input["PK_DOCUMENTO_TRABAJADOR"].(float64); ok {
		doc := int64(documento)
		incidencia.PK_DOCUMENTO_TRABAJADOR = &doc
	}

	// Guardar los cambios en la base de datos
	if _, err := o.Update(&incidencia); err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al actualizar la incidencia",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Preparar la respuesta con el formato deseado
	response := map[string]interface{}{
		"PK_ID_INCIDENCIA":        incidencia.PK_ID_INCIDENCIA,
		"FECHA":                   incidencia.FECHA.Format("2006-01-02"),
		"MONTO":                   incidencia.MONTO,
		"RESTA":                   incidencia.RESTA,
		"MOTIVO":                  incidencia.MOTIVO,
		"PK_DOCUMENTO_TRABAJADOR": incidencia.PK_DOCUMENTO_TRABAJADOR,
	}

	// Responder con éxito
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Incidencia actualizada correctamente",
		Data:    response,
	}
	c.ServeJSON()
}

// @Title Delete
// @Summary Eliminar una incidencia
// @Description Elimina una incidencia de la base de datos.
// @Tags incidencias
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID de la incidencia"
// @Success 200 {object} models.ApiResponse "Incidencia eliminada"
// @Failure 404 {object} models.ApiResponse "Incidencia no encontrada"
// @Security BearerAuth
// @Router /incidencias [delete]
func (c *IncidenciaController) Delete() {
	o := orm.NewOrm()
	id, err := c.GetInt64("id")
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "ID inválido o ausente",
		}
		c.ServeJSON()
		return
	}

	_, err = o.Delete(&models.Incidencia{PK_ID_INCIDENCIA: id})
	if err == orm.ErrNoRows {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Incidencia no encontrada",
		}
		c.ServeJSON()
		return
	} else if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al eliminar la incidencia",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Incidencia eliminada correctamente",
	}
	c.ServeJSON()
}
