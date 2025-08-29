package controllers

import (
	"encoding/json"
	"net/http"
	"restaurante/models"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

type horarioOrmer interface {
	QueryTable(interface{}) orm.QuerySeter
	Insert(interface{}) (int64, error)
	Update(interface{}, ...string) (int64, error)
}

// horarioTrabajadorNewOrm allows tests to stub orm.NewOrm.
var horarioTrabajadorNewOrm = func() horarioOrmer { return orm.NewOrm() }

type HorarioTrabajadorController struct {
	web.Controller
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
	o := horarioTrabajadorNewOrm()
	qs := o.QueryTable(new(models.HorarioTrabajador))

	if doc, err := c.GetInt64("documento"); err == nil && doc != 0 {
		qs = qs.Filter("PK_DOCUMENTO_TRABAJADOR", doc)
	}
	if dia := c.GetString("dia"); dia != "" {
		qs = qs.Filter("DIA", dia)
	}

	var horarios []models.HorarioTrabajador
	if _, err := qs.All(&horarios); err != nil {
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
		DIA:                     input.DIA,
		HORA_INICIO:             horaInicio,
		HORA_FIN:                horaFin,
	}

	o := horarioTrabajadorNewOrm()
	if _, err := o.Insert(&horario); err != nil {
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
	dia := c.GetString("dia")
	if dia == "" {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Parámetro 'dia' inválido"}
		c.ServeJSON()
		return
	}

	var horario models.HorarioTrabajador
	o := horarioTrabajadorNewOrm()
	if err := o.QueryTable(new(models.HorarioTrabajador)).
		Filter("PK_DOCUMENTO_TRABAJADOR", doc).
		Filter("DIA", dia).
		One(&horario); err == orm.ErrNoRows {
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

	if _, err := o.Update(&horario); err != nil {
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

	o := horarioTrabajadorNewOrm()
	if _, err := o.QueryTable(new(models.HorarioTrabajador)).
		Filter("PK_DOCUMENTO_TRABAJADOR", doc).
		Filter("DIA", dia).
		Delete(); err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al eliminar horario", Cause: err.Error()}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{Code: http.StatusOK, Message: "Horario eliminado correctamente"}
	c.ServeJSON()
}
