package horariotrabajador

import (
	"encoding/json"
	"net/http"
	"restaurante/logging"
	"restaurante/models"
	"strings"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

type HorarioTrabajadorController struct {
	web.Controller
}

func isValidDia(dia string) bool {
	db := diaToDB(dia)
	switch models.DiaSemana(db) {
	case models.DiaLunes, models.DiaMartes, models.DiaMiercoles,
		models.DiaJueves, models.DiaViernes, models.DiaSabado, models.DiaDomingo:
		return true
	default:
		return false
	}
}

func diaToDB(d string) string {
	if d == "" {
		return ""
	}
	d = strings.ToLower(d)
	if len(d) == 1 {
		return strings.ToUpper(d)
	}
	return strings.ToUpper(d[:1]) + d[1:]
}

// @Title GetAll
// @Summary Obtener horarios de trabajadores
// @Description Lista los horarios, opcionalmente filtrados por documento del trabajador.
// @Tags horarios_trabajador
// @Accept json
// @Produce json
// @Param   documento  query int false "Documento del trabajador"
// @Param   dia        query string false "Día a filtrar"
// @Success 200 {object} models.ApiResponse{data=[]models.HorarioTrabajador} "Lista de horarios"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /horario_trabajador [get]
func (c *HorarioTrabajadorController) GetAll() {
	o := orm.NewOrm()
	query := "SELECT pk_documento_trabajador, dia, hora_inicio, hora_fin FROM horario_trabajador"
	var args []interface{}
	var conds []string

	if doc, err := c.GetInt64("documento"); err == nil && doc != 0 {
		conds = append(conds, "pk_documento_trabajador = ?")
		args = append(args, doc)
	}
	if dia := c.GetString("dia"); dia != "" {
		if isValidDia(dia) {
			conds = append(conds, "dia = ?")
			args = append(args, diaToDB(dia))
		}
	}
	if len(conds) > 0 {
		query += " WHERE " + strings.Join(conds, " AND ")
	}

	var horarios []models.HorarioTrabajador
	if _, err := o.Raw(query, args...).QueryRows(&horarios); err != nil {
		logging.LogControllerError(c.Ctx, "horario_trabajador.getall.db_error", err, map[string]interface{}{"documento": c.GetString("documento"), "dia": c.GetString("dia")})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al obtener horarios", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{Code: http.StatusOK, Message: "Horarios obtenidos correctamente", Data: horarios}
	_ = c.ServeJSON()
}

// @Title Post
// @Summary Crear horario para trabajador
// @Description Crea un registro de horario para un trabajador.
// @Tags horarios_trabajador
// @Accept json
// @Produce json
// @Param   body  body   models.HorarioTrabajadorCreateRequest true  "Datos del horario (formato hora HH:MM:SS)"
// @Success 201 {object} models.ApiResponse{data=models.HorarioTrabajador} "Horario creado"
// @Failure 400 {object} models.ApiResponse "Solicitud inválida"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /horario_trabajador [post]
func (c *HorarioTrabajadorController) Post() {
	var input models.HorarioTrabajadorCreateRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &input); err != nil {
		logging.LogControllerError(c.Ctx, "horario_trabajador.post.bad_json", err, nil)
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Error al decodificar la solicitud", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}
	dbDia := diaToDB(input.Dia)
	if !isValidDia(dbDia) {
		logging.LogControllerError(c.Ctx, "horario_trabajador.post.validation_error", nil, map[string]interface{}{"dia": input.Dia})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Día inválido"}
		_ = c.ServeJSON()
		return
	}
	// Para este endpoint exigimos formato HH:MM:SS estrictamente
	horaInicio, err1 := time.Parse("15:04:05", input.HoraInicio)
	horaFin, err2 := time.Parse("15:04:05", input.HoraFin)
	if err1 != nil || err2 != nil {
		logging.LogControllerError(c.Ctx, "horario_trabajador.post.bad_time_format", nil, map[string]interface{}{"horaInicio": input.HoraInicio, "horaFin": input.HoraFin})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Formato de hora inválido"}
		_ = c.ServeJSON()
		return
	}
	// Normalizamos a la fecha base 0001-01-01 en UTC
	horaInicio = time.Date(1, 1, 1, horaInicio.Hour(), horaInicio.Minute(), horaInicio.Second(), 0, time.UTC)
	horaFin = time.Date(1, 1, 1, horaFin.Hour(), horaFin.Minute(), horaFin.Second(), 0, time.UTC)
	validCandidate := &models.HorarioTrabajador{HORA_INICIO: horaInicio, HORA_FIN: horaFin}
	if !validCandidate.ValidHours() {
		logging.LogControllerError(c.Ctx, "horario_trabajador.post.validation_error", nil, map[string]interface{}{"horaInicio": input.HoraInicio, "horaFin": input.HoraFin})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "horaFin debe ser mayor que horaInicio"}
		_ = c.ServeJSON()
		return
	}
	Horario := models.HorarioTrabajador{PK_DOCUMENTO_TRABAJADOR: &models.Trabajador{PK_DOCUMENTO_TRABAJADOR: input.DocumentoTrabajador}, DIA: dbDia, HORA_INICIO: horaInicio, HORA_FIN: horaFin}
	o := orm.NewOrm()
	if _, err := o.Raw("INSERT INTO horario_trabajador (pk_documento_trabajador, dia, hora_inicio, hora_fin) VALUES (?, ?, ?, ?)", Horario.PK_DOCUMENTO_TRABAJADOR.PK_DOCUMENTO_TRABAJADOR, Horario.DIA, Horario.HORA_INICIO, Horario.HORA_FIN).Exec(); err != nil {
		logging.LogControllerError(c.Ctx, "horario_trabajador.post.insert_error", err, map[string]interface{}{"documento": input.DocumentoTrabajador, "dia": dbDia})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al crear horario", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}
	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = models.ApiResponse{Code: http.StatusCreated, Message: "Horario creado correctamente", Data: Horario}
	_ = c.ServeJSON()
}

// @Title Put
// @Summary Actualizar horario de trabajador
// @Description Actualiza las horas de inicio o fin para un trabajador y día específico.
// @Tags horarios_trabajador
// @Accept json
// @Produce json
// @Param   documento query int true "Documento del trabajador"
// @Param   dia       query string true "Día del horario"
// @Param   body body models.HorarioTrabajadorUpdateRequest true "Horas a actualizar (formato HH:MM:SS)"
// @Success 200 {object} models.ApiResponse "Horario actualizado"
// @Failure 400 {object} models.ApiResponse "Solicitud inválida"
// @Failure 404 {object} models.ApiResponse "Horario no encontrado"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /horario_trabajador [put]
func (c *HorarioTrabajadorController) Put() {
	doc, err := c.GetInt64("documento")
	if err != nil || doc == 0 {
		logging.LogControllerError(c.Ctx, "horario_trabajador.put.bad_request", err, map[string]interface{}{"documento": c.GetString("documento")})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Parámetro 'documento' inválido"}
		_ = c.ServeJSON()
		return
	}
	dia := c.GetString("dia")
	if dia == "" || !isValidDia(dia) {
		logging.LogControllerError(c.Ctx, "horario_trabajador.put.bad_request", nil, map[string]interface{}{"dia": dia})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Parámetro 'dia' inválido"}
		_ = c.ServeJSON()
		return
	}
	dbDia := diaToDB(dia)
	var horario models.HorarioTrabajador
	o := orm.NewOrm()
	if err := o.Raw("SELECT pk_documento_trabajador, dia, hora_inicio, hora_fin FROM horario_trabajador WHERE pk_documento_trabajador = ? AND dia = ?", doc, dbDia).QueryRow(&horario); err == orm.ErrNoRows {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{Code: http.StatusNotFound, Message: "Horario no encontrado"}
		_ = c.ServeJSON()
		return
	} else if err != nil {
		logging.LogControllerError(c.Ctx, "horario_trabajador.put.db_error", err, map[string]interface{}{"documento": doc, "dia": dbDia})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al consultar horario", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}
	var input models.HorarioTrabajadorUpdateRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &input); err != nil {
		logging.LogControllerError(c.Ctx, "horario_trabajador.put.bad_json", err, map[string]interface{}{"documento": doc, "dia": dbDia})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Error al decodificar la solicitud", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}
	if input.HoraInicio != nil && *input.HoraInicio != "" {
		if t, err := models.ParseTimeToUTC(*input.HoraInicio); err == nil {
			horario.HORA_INICIO = t
		} else {
			logging.LogControllerError(c.Ctx, "horario_trabajador.put.bad_time_format", nil, map[string]interface{}{"horaInicio": *input.HoraInicio})
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Formato de horaInicio inválido"}
			_ = c.ServeJSON()
			return
		}
	}
	if input.HoraFin != nil && *input.HoraFin != "" {
		if t, err := models.ParseTimeToUTC(*input.HoraFin); err == nil {
			horario.HORA_FIN = t
		} else {
			logging.LogControllerError(c.Ctx, "horario_trabajador.put.bad_time_format", nil, map[string]interface{}{"horaFin": *input.HoraFin})
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Formato de horaFin inválido"}
			_ = c.ServeJSON()
			return
		}
	}
	// Normalizar ambas horas a fecha base para comparar solo por hora/min/seg
	horario.HORA_INICIO = time.Date(1, 1, 1, horario.HORA_INICIO.Hour(), horario.HORA_INICIO.Minute(), horario.HORA_INICIO.Second(), 0, time.UTC)
	horario.HORA_FIN = time.Date(1, 1, 1, horario.HORA_FIN.Hour(), horario.HORA_FIN.Minute(), horario.HORA_FIN.Second(), 0, time.UTC)
	if !horario.ValidHours() {
		logging.LogControllerError(c.Ctx, "horario_trabajador.put.validation_error", nil, map[string]interface{}{"horaInicio": input.HoraInicio, "horaFin": input.HoraFin})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "horaFin debe ser mayor que horaInicio"}
		_ = c.ServeJSON()
		return
	}
	if _, err := o.Raw("UPDATE horario_trabajador SET hora_inicio = ?, hora_fin = ? WHERE pk_documento_trabajador = ? AND dia = ?", horario.HORA_INICIO, horario.HORA_FIN, doc, dbDia).Exec(); err != nil {
		logging.LogControllerError(c.Ctx, "horario_trabajador.put.update_error", err, map[string]interface{}{"documento": doc, "dia": dbDia})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al actualizar horario", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{Code: http.StatusOK, Message: "Horario actualizado correctamente", Data: horario}
	_ = c.ServeJSON()
}

// @Title Delete
// @Summary Eliminar horario de trabajador
// @Description Elimina un horario de un trabajador y día específico.
// @Tags horarios_trabajador
// @Accept json
// @Produce json
// @Param   documento query int true "Documento del trabajador"
// @Param   dia       query string true "Día del horario"
// @Success 200 {object} models.ApiResponse "Horario eliminado"
// @Failure 400 {object} models.ApiResponse "Parámetros inválidos"
// @Failure 404 {object} models.ApiResponse "Horario no encontrado"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /horario_trabajador [delete]
func (c *HorarioTrabajadorController) Delete() {
	doc, err := c.GetInt64("documento")
	if err != nil || doc == 0 {
		logging.LogControllerError(c.Ctx, "horario_trabajador.delete.bad_request", err, map[string]interface{}{"documento": c.GetString("documento")})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Parámetro 'documento' inválido"}
		_ = c.ServeJSON()
		return
	}
	dia := c.GetString("dia")
	if dia == "" || !isValidDia(dia) {
		logging.LogControllerError(c.Ctx, "horario_trabajador.delete.bad_request", nil, map[string]interface{}{"dia": dia})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Parámetro 'dia' inválido"}
		_ = c.ServeJSON()
		return
	}
	dbDia := diaToDB(dia)
	o := orm.NewOrm()
	var horario models.HorarioTrabajador
	if err := o.Raw("SELECT pk_documento_trabajador, dia FROM horario_trabajador WHERE pk_documento_trabajador = ? AND dia = ?", doc, dbDia).QueryRow(&horario); err == orm.ErrNoRows {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{Code: http.StatusNotFound, Message: "Horario no encontrado"}
		_ = c.ServeJSON()
		return
	} else if err != nil {
		logging.LogControllerError(c.Ctx, "horario_trabajador.delete.db_query_error", err, map[string]interface{}{"documento": doc, "dia": dbDia})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al eliminar horario", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}
	if _, err := o.Raw("DELETE FROM horario_trabajador WHERE pk_documento_trabajador = ? AND dia = ?", doc, dbDia).Exec(); err != nil {
		logging.LogControllerError(c.Ctx, "horario_trabajador.delete.delete_error", err, map[string]interface{}{"documento": doc, "dia": dbDia})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al eliminar horario", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{Code: http.StatusOK, Message: "Horario eliminado correctamente"}
	_ = c.ServeJSON()
}
