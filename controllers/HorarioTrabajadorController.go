package controllers

import (
	"encoding/json"
	"net/http"
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
	switch models.DiaSemana(strings.ToUpper(dia)) {
	case models.DiaLunes, models.DiaMartes, models.DiaMiercoles,
		models.DiaJueves, models.DiaViernes, models.DiaSabado, models.DiaDomingo:
		return true
	default:
		return false
	}
}

// @Title GetAll
// @Summary Obtener horarios de trabajadores
// @Description Lista los horarios, opcionalmente filtrados por documento del trabajador.
// @Tags horarios_trabajador
// @Accept json
// @Produce json
// @Param   documento  query int false "Documento del trabajador"
// @Param   dia        query string false "Día a filtrar"
// @Success 200 {array} models.HorarioTrabajador "Lista de horarios"
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
	if dia := strings.ToUpper(c.GetString("dia")); dia != "" {
		conds = append(conds, "dia = ?")
		args = append(args, dia)
	}
	if len(conds) > 0 {
		query += " WHERE " + strings.Join(conds, " AND ")
	}

	var horarios []models.HorarioTrabajador
	if _, err := o.Raw(query, args...).QueryRows(&horarios); err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener horarios",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Horarios obtenidos correctamente",
		Data:    horarios,
	}
	c.ServeJSON()
}

// @Title Post
// @Summary Crear horario para trabajador
// @Description Crea un registro de horario para un trabajador.
// @Tags horarios_trabajador
// @Accept json
// @Produce json
// @Param   body  body   models.HorarioTrabajador true  "Datos del horario"
// @Success 201 {object} models.ApiResponse "Horario creado"
// @Failure 400 {object} models.ApiResponse "Solicitud inválida"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /horario_trabajador [post]
func (c *HorarioTrabajadorController) Post() {
	var input struct {
		PK_DOCUMENTO_TRABAJADOR int64  `json:"documentoTrabajador"`
		DIA                     string `json:"dia"`
		HORA_INICIO             string `json:"horaInicio"`
		HORA_FIN                string `json:"horaFin"`
	}

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

	dia := strings.ToUpper(input.DIA)
	if !isValidDia(dia) {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Día inválido"}
		c.ServeJSON()
		return
	}

	horaInicio, err1 := time.Parse("15:04:05", input.HORA_INICIO)
	horaFin, err2 := time.Parse("15:04:05", input.HORA_FIN)
	if err1 != nil || err2 != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Formato de hora inválido",
		}
		c.ServeJSON()
		return
	}

	horario := models.HorarioTrabajador{
		PK_DOCUMENTO_TRABAJADOR: input.PK_DOCUMENTO_TRABAJADOR,
		DIA:                     dia,
		HORA_INICIO:             horaInicio,
		HORA_FIN:                horaFin,
	}

	if !horario.ValidHours() {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "horaFin debe ser mayor que horaInicio"}
		c.ServeJSON()
		return
	}

	o := orm.NewOrm()
	if _, err := o.Raw(
		"INSERT INTO horario_trabajador (pk_documento_trabajador, dia, hora_inicio, hora_fin) VALUES (?, ?, ?, ?)",
		horario.PK_DOCUMENTO_TRABAJADOR, horario.DIA, horario.HORA_INICIO, horario.HORA_FIN,
	).Exec(); err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al crear horario",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusCreated,
		Message: "Horario creado correctamente",
		Data:    horario,
	}
	c.ServeJSON()
}

// @Title Put
// @Summary Actualizar horario de trabajador
// @Description Actualiza las horas de inicio o fin para un trabajador y día específico.
// @Tags horarios_trabajador
// @Accept json
// @Produce json
// @Param   documento query int true "Documento del trabajador"
// @Param   dia       query string true "Día del horario"
// @Param   body body map[string]string true "Horas a actualizar"
// @Success 200 {object} models.ApiResponse "Horario actualizado"
// @Failure 400 {object} models.ApiResponse "Solicitud inválida"
// @Failure 404 {object} models.ApiResponse "Horario no encontrado"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /horario_trabajador [put]
func (c *HorarioTrabajadorController) Put() {
	doc, err := c.GetInt64("documento")
	if err != nil || doc == 0 {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Parámetro 'documento' inválido"}
		c.ServeJSON()
		return
	}
	dia := strings.ToUpper(c.GetString("dia"))
	if dia == "" || !isValidDia(dia) {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Parámetro 'dia' inválido"}
		c.ServeJSON()
		return
	}

	var horario models.HorarioTrabajador
	o := orm.NewOrm()
	if err := o.Raw(
		"SELECT pk_documento_trabajador, dia, hora_inicio, hora_fin FROM horario_trabajador WHERE pk_documento_trabajador = ? AND dia = ?",
		doc, dia,
	).QueryRow(&horario); err == orm.ErrNoRows {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{Code: http.StatusNotFound, Message: "Horario no encontrado"}
		c.ServeJSON()
		return
	} else if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al consultar horario", Cause: err.Error()}
		c.ServeJSON()
		return
	}

	var input map[string]string
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &input); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Error al decodificar la solicitud", Cause: err.Error()}
		c.ServeJSON()
		return
	}

	if v, ok := input["horaInicio"]; ok && v != "" {
		if t, err := time.Parse("15:04:05", v); err == nil {
			horario.HORA_INICIO = t
		} else {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Formato de horaInicio inválido"}
			c.ServeJSON()
			return
		}
	}
	if v, ok := input["horaFin"]; ok && v != "" {
		if t, err := time.Parse("15:04:05", v); err == nil {
			horario.HORA_FIN = t
		} else {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Formato de horaFin inválido"}
			c.ServeJSON()
			return
		}
	}

	horario.HORA_INICIO = time.Date(0, 1, 1, horario.HORA_INICIO.Hour(), horario.HORA_INICIO.Minute(), horario.HORA_INICIO.Second(), 0, time.UTC)
	horario.HORA_FIN = time.Date(0, 1, 1, horario.HORA_FIN.Hour(), horario.HORA_FIN.Minute(), horario.HORA_FIN.Second(), 0, time.UTC)
	if !horario.ValidHours() {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "horaFin debe ser mayor que horaInicio"}
		c.ServeJSON()
		return
	}

	if _, err := o.Raw(
		"UPDATE horario_trabajador SET hora_inicio = ?, hora_fin = ? WHERE pk_documento_trabajador = ? AND dia = ?",
		horario.HORA_INICIO, horario.HORA_FIN, doc, dia,
	).Exec(); err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al actualizar horario", Cause: err.Error()}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{Code: http.StatusOK, Message: "Horario actualizado correctamente", Data: horario}
	c.ServeJSON()
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
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Parámetro 'documento' inválido"}
		c.ServeJSON()
		return
	}
	dia := c.GetString("dia")
	if dia == "" {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Parámetro 'dia' inválido"}
		c.ServeJSON()
		return
	}

	o := orm.NewOrm()
	var horario models.HorarioTrabajador
	if err := o.Raw(
		"SELECT pk_documento_trabajador, dia FROM horario_trabajador WHERE pk_documento_trabajador = ? AND dia = ?",
		doc, dia,
	).QueryRow(&horario); err == orm.ErrNoRows {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{Code: http.StatusNotFound, Message: "Horario no encontrado"}
		c.ServeJSON()
		return
	} else if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al eliminar horario", Cause: err.Error()}
		c.ServeJSON()
		return
	}

	if _, err := o.Raw(
		"DELETE FROM horario_trabajador WHERE pk_documento_trabajador = ? AND dia = ?",
		doc, dia,
	).Exec(); err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al eliminar horario", Cause: err.Error()}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{Code: http.StatusOK, Message: "Horario eliminado correctamente"}
	c.ServeJSON()
}
