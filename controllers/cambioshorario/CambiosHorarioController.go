package cambioshorario

import (
	"encoding/json"
	"errors"
	"net/http"
	"restaurante/database"
	"restaurante/logging"
	"restaurante/models"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

type CambiosHorarioController struct {
	web.Controller
}

var queryCambioHorarioByDate = func(o orm.Ormer, date string, ch *models.CambiosHorario) error {
	if o == nil {
		return errors.New("nil ormer")
	}
	return o.QueryTable(new(models.CambiosHorario)).Filter("FECHA", date).One(ch)
}

var (
	queryAllCambiosHorario = func(o orm.Ormer, horarios *[]models.CambiosHorario) (int64, error) {
		if o == nil {
			return 0, errors.New("nil ormer")
		}
		return o.QueryTable(new(models.CambiosHorario)).All(horarios)
	}
	insertCambioHorario = func(o orm.Ormer, horario *models.CambiosHorario) (int64, error) {
		if o == nil {
			return 0, errors.New("nil ormer")
		}
		return o.Insert(horario)
	}
	queryCambioHorarioByID = func(o orm.Ormer, id int64, horario *models.CambiosHorario) error {
		if o == nil {
			return errors.New("nil ormer")
		}
		return o.QueryTable(new(models.CambiosHorario)).Filter("PK_ID_CAMBIO_HORARIO", id).One(horario)
	}
	updateCambioHorario = func(o orm.Ormer, horario *models.CambiosHorario) (int64, error) {
		if o == nil {
			return 0, errors.New("nil ormer")
		}
		return o.Update(horario)
	}
	deleteCambioHorarioByID = func(o orm.Ormer, id int64) (int64, error) {
		if o == nil {
			return 0, errors.New("nil ormer")
		}
		return o.QueryTable(new(models.CambiosHorario)).Filter("PK_ID_CAMBIO_HORARIO", id).Delete()
	}
)

// @Title GetAll
// @Summary Obtener todos los cambios de horario
// @Description Obtiene un listado de todos los cambios de horario registrados en la base de datos
// @Tags cambios_horario
// @Accept json
// @Produce json
// @Success 200 {object} models.ApiResponse{data=[]map[string]interface{}} "Listado de cambios de horario"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /cambios_horario [get]
func (c *CambiosHorarioController) GetAll() {
	o := orm.NewOrm()
	var horarios []models.CambiosHorario

	if _, err := queryAllCambiosHorario(o, &horarios); err != nil {
		logging.LogControllerError(c.Ctx, "cambios_horario.getall.db_error", err, nil)
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener cambios de horario",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	var response []map[string]interface{}
	for _, horario := range horarios {
		h := map[string]interface{}{
			"cambioHorarioId":    horario.PK_ID_CAMBIO_HORARIO,
			"fechaCambioHorario": horario.FECHA.Format("2006-01-02"),
			"abierto":            horario.ABIERTO,
		}
		if horario.HORA_APERTURA != nil {
			h["horaApertura"] = horario.HORA_APERTURA.Format("15:04:05")
		} else {
			h["horaApertura"] = nil
		}
		if !horario.HORA_CIERRE.IsZero() {
			h["horaCierre"] = horario.HORA_CIERRE.Format("15:04:05")
		} else {
			h["horaCierre"] = nil
		}
		response = append(response, h)
	}

	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Cambios de horario obtenidos correctamente",
		Data:    response,
	}
	_ = c.ServeJSON()
}

// @Title GetByCurrentDate
// @Summary Consultar cambios de horario para la fecha actual
// @Description Obtiene el cambio de horario que aplica para la fecha actual, si existe.
// @Tags cambios_horario
// @Accept json
// @Produce json
// @Success 200 {object} models.ApiResponse{data=map[string]interface{}} "Cambio de horario para la fecha actual"
// @Failure 200 {object} models.ApiResponse "No hay cambios de horario para la fecha actual"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Router /cambios_horario/actual [get]
func (c *CambiosHorarioController) GetByCurrentDate() {
	o := orm.NewOrm()
	var cambioHorario models.CambiosHorario

	currentDate := time.Now().In(database.BogotaZone)
	dateStr := currentDate.Format("2006-01-02")

	if err := queryCambioHorarioByDate(o, dateStr, &cambioHorario); err != nil {
		if err == orm.ErrNoRows {
			c.Ctx.Output.SetStatus(http.StatusOK)
			c.Data["json"] = models.ApiResponse{Code: http.StatusNotFound, Message: "No hay cambios de horario para la fecha actual"}
			_ = c.ServeJSON()
			return
		}
		logging.LogControllerError(c.Ctx, "cambios_horario.actual.db_error", err, map[string]interface{}{"date": dateStr})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al consultar cambios de horario", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}

	response := map[string]interface{}{
		"cambioHorarioId":    cambioHorario.PK_ID_CAMBIO_HORARIO,
		"fechaCambioHorario": cambioHorario.FECHA.Format("2006-01-02"),
		"abierto":            cambioHorario.ABIERTO,
	}
	if cambioHorario.HORA_APERTURA != nil {
		response["horaApertura"] = cambioHorario.HORA_APERTURA.Format("15:04:05")
	}
	if !cambioHorario.HORA_CIERRE.IsZero() {
		response["horaCierre"] = cambioHorario.HORA_CIERRE.Format("15:04:05")
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{Code: http.StatusOK, Message: "Cambio de horario encontrado para la fecha actual", Data: response}
	_ = c.ServeJSON()
}

// @Title Post
// @Summary Crear un nuevo cambio de horario
// @Description Crea un nuevo cambio de horario en la base de datos.
// @Tags cambios_horario
// @Accept json
// @Produce json
// @Param body body models.CambiosHorarioCreateRequest true "Datos del cambio de horario (fecha YYYY-MM-DD, horas HH:MM:SS)"
// @Success 201 {object} models.ApiResponse{data=map[string]interface{}} "Cambio de horario creado"
// @Failure 400 {object} models.ApiResponse "Error en la solicitud"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /cambios_horario [post]
func (c *CambiosHorarioController) Post() {
	o := orm.NewOrm()
	var input map[string]interface{}
	var horario models.CambiosHorario

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &input); err != nil {
		logging.LogControllerError(c.Ctx, "cambios_horario.post.bad_json", err, nil)
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Error al procesar la solicitud",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	if fechaStr, ok := input["fechaCambioHorario"].(string); ok && fechaStr != "" {
		parsedDate, err := time.Parse("2006-01-02", fechaStr)
		if err != nil {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusBadRequest,
				Message: "Formato de fecha inválido para FECHA",
				Cause:   err.Error(),
			}
			_ = c.ServeJSON()
			return
		}
		// FECHA estable a mediodía UTC
		horario.FECHA = time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 12, 0, 0, 0, time.UTC)
	} else {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El campo FECHA es obligatorio",
		}
		_ = c.ServeJSON()
		return
	}

	if abierto, ok := input["abierto"].(bool); ok {
		horario.ABIERTO = abierto
	} else {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El campo ABIERTO es obligatorio",
		}
		_ = c.ServeJSON()
		return
	}

	if !horario.ABIERTO {
		if ha, err := time.Parse("15:04:05", "00:00:00"); err == nil {
			horario.HORA_APERTURA = &ha
		}
		if hc, err := time.Parse("15:04:05", "23:59:59"); err == nil {
			horario.HORA_CIERRE = hc
		}
	} else {
		if horaAperturaStr, ok := input["horaApertura"].(string); ok && horaAperturaStr != "" {
			parsedHora, err := time.Parse("15:04:05", horaAperturaStr)
			if err != nil {
				c.Ctx.Output.SetStatus(http.StatusBadRequest)
				c.Data["json"] = models.ApiResponse{
					Code:    http.StatusBadRequest,
					Message: "Formato de hora inválido para HORA_APERTURA",
					Cause:   err.Error(),
				}
				_ = c.ServeJSON()
				return
			}
			horario.HORA_APERTURA = &parsedHora
		} else {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusBadRequest,
				Message: "El campo HORA_APERTURA es obligatorio cuando ABIERTO es true",
			}
			_ = c.ServeJSON()
			return
		}

		if horaCierreStr, ok := input["horaCierre"].(string); ok && horaCierreStr != "" {
			parsedHora, err := time.Parse("15:04:05", horaCierreStr)
			if err != nil {
				c.Ctx.Output.SetStatus(http.StatusBadRequest)
				c.Data["json"] = models.ApiResponse{
					Code:    http.StatusBadRequest,
					Message: "Formato de hora inválido para HORA_CIERRE",
					Cause:   err.Error(),
				}
				_ = c.ServeJSON()
				return
			}
			horario.HORA_CIERRE = parsedHora
		} else {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusBadRequest,
				Message: "El campo HORA_CIERRE es obligatorio cuando ABIERTO es true",
			}
			_ = c.ServeJSON()
			return
		}
	}

	if _, err := insertCambioHorario(o, &horario); err != nil {
		logging.LogControllerError(c.Ctx, "cambios_horario.post.db_error", err, nil)
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al crear el cambio de horario",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	response := map[string]interface{}{
		"cambioHorarioId":    horario.PK_ID_CAMBIO_HORARIO,
		"fechaCambioHorario": horario.FECHA.Format("2006-01-02"),
		"abierto":            horario.ABIERTO,
	}
	if horario.HORA_APERTURA != nil {
		response["horaApertura"] = horario.HORA_APERTURA.Format("15:04:05")
	}
	if !horario.HORA_CIERRE.IsZero() {
		response["horaCierre"] = horario.HORA_CIERRE.Format("15:04:05")
	}

	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusCreated,
		Message: "Cambio de horario creado correctamente",
		Data:    response,
	}
	_ = c.ServeJSON()
}

// @Title Update
// @Summary Actualizar un cambio de horario
// @Description Actualiza los datos de un cambio de horario existente.
// @Tags cambios_horario
// @Accept json
// @Produce json
// @Param id query int true "ID del cambio de horario"
// @Param body body models.CambiosHorarioUpdateRequest true "Datos del cambio de horario a actualizar (sólo campos a modificar)"
// @Success 200 {object} models.ApiResponse{data=map[string]interface{}} "Cambio de horario actualizado"
// @Failure 400 {object} models.ApiResponse "Error en la solicitud"
// @Failure 404 {object} models.ApiResponse "Cambio de horario no encontrado"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /cambios_horario [put]
func (c *CambiosHorarioController) Put() {
	o := orm.NewOrm()
	id, err := c.GetInt64("id")
	if err != nil || id == 0 {
		logging.LogControllerError(c.Ctx, "cambios_horario.put.bad_request", err, map[string]interface{}{"id": c.GetString("id")})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "ID inválido o ausente",
		}
		_ = c.ServeJSON()
		return
	}

	var input map[string]interface{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &input); err != nil {
		logging.LogControllerError(c.Ctx, "cambios_horario.put.bad_json", err, nil)
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Error al procesar la solicitud",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	var horario models.CambiosHorario
	if err := queryCambioHorarioByID(o, id, &horario); err == orm.ErrNoRows {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Cambio de horario no encontrado",
		}
		_ = c.ServeJSON()
		return
	} else if err != nil {
		logging.LogControllerError(c.Ctx, "cambios_horario.put.db_error", err, map[string]interface{}{"id": id})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al buscar el cambio de horario",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	if fechaStr, ok := input["fechaCambioHorario"].(string); ok && fechaStr != "" {
		parsedDate, err := time.Parse("2006-01-02", fechaStr)
		if err != nil {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusBadRequest,
				Message: "Formato de fecha inválido para FECHA",
				Cause:   err.Error(),
			}
			_ = c.ServeJSON()
			return
		}
		horario.FECHA = time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 12, 0, 0, 0, time.UTC)
	}

	if abierto, ok := input["abierto"].(bool); ok {
		horario.ABIERTO = abierto
		if !abierto {
			if ha, err := time.Parse("15:04:05", "00:00:00"); err == nil {
				horario.HORA_APERTURA = &ha
			}
			if hc, err := time.Parse("15:04:05", "23:59:59"); err == nil {
				horario.HORA_CIERRE = hc
			}
		}
	}

	if horario.ABIERTO {
		if horaAperturaStr, ok := input["horaApertura"].(string); ok && horaAperturaStr != "" {
			parsedHora, err := time.Parse("15:04:05", horaAperturaStr)
			if err != nil {
				c.Ctx.Output.SetStatus(http.StatusBadRequest)
				c.Data["json"] = models.ApiResponse{
					Code:    http.StatusBadRequest,
					Message: "Formato de hora inválido para HORA_APERTURA",
					Cause:   err.Error(),
				}
				_ = c.ServeJSON()
				return
			}
			horario.HORA_APERTURA = &parsedHora
		}

		if horaCierreStr, ok := input["horaCierre"].(string); ok && horaCierreStr != "" {
			parsedHora, err := time.Parse("15:04:05", horaCierreStr)
			if err != nil {
				c.Ctx.Output.SetStatus(http.StatusBadRequest)
				c.Data["json"] = models.ApiResponse{
					Code:    http.StatusBadRequest,
					Message: "Formato de hora inválido para HORA_CIERRE",
					Cause:   err.Error(),
				}
				_ = c.ServeJSON()
				return
			}
			horario.HORA_CIERRE = parsedHora
		}
	}

	if _, err := updateCambioHorario(o, &horario); err != nil {
		logging.LogControllerError(c.Ctx, "cambios_horario.put.db_error", err, map[string]interface{}{"id": id})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al actualizar el cambio de horario",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	response := map[string]interface{}{
		"cambioHorarioId": horario.PK_ID_CAMBIO_HORARIO,
		"fecha":           horario.FECHA.Format("2006-01-02"),
		"abierto":         horario.ABIERTO,
	}
	if horario.HORA_APERTURA != nil {
		response["horaApertura"] = horario.HORA_APERTURA.Format("15:04:05")
	}
	if !horario.HORA_CIERRE.IsZero() {
		response["horaCierre"] = horario.HORA_CIERRE.Format("15:04:05")
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Cambio de horario actualizado correctamente",
		Data:    response,
	}
	_ = c.ServeJSON()
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
		logging.LogControllerError(c.Ctx, "cambios_horario.delete.bad_request", err, map[string]interface{}{"id": c.GetString("id")})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "ID inválido o ausente",
		}
		_ = c.ServeJSON()
		return
	}

	if num, err := deleteCambioHorarioByID(o, id); err != nil {
		logging.LogControllerError(c.Ctx, "cambios_horario.delete.db_error", err, map[string]interface{}{"id": id})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al eliminar el cambio de horario",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
	} else if num == 0 {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Cambio de horario no encontrado",
		}
		_ = c.ServeJSON()
	} else {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusOK,
			Message: "Cambio de horario eliminado correctamente",
		}
		_ = c.ServeJSON()
	}
}
