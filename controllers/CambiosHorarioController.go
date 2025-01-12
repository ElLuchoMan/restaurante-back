package controllers

import (
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
// @Success 200 {array} models.CambiosHorario "Listado de cambios de horario"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /cambios_horario [get]
func (c *CambiosHorarioController) GetAll() {
	o := orm.NewOrm()
	var horarios []models.CambiosHorario

	// Obtener todos los registros de la tabla
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

	// Ajustar fechas y horas al timezone de Bogotá y manejar nil
	for i := range horarios {
		// Ajustar FECHA al timezone de Bogotá
		horarios[i].FECHA = horarios[i].FECHA.In(database.BogotaZone)

		// Ajustar HORA_APERTURA al timezone de Bogotá si no es nil
		if horarios[i].HORA_APERTURA != nil && !horarios[i].HORA_APERTURA.IsZero() {
			adjustedHoraApertura := horarios[i].HORA_APERTURA.In(database.BogotaZone)
			horarios[i].HORA_APERTURA = &adjustedHoraApertura
		} else {
			horarios[i].HORA_APERTURA = nil
		}

		// Ajustar HORA_CIERRE al timezone de Bogotá si no es nil
		if horarios[i].HORA_CIERRE != nil && !horarios[i].HORA_CIERRE.IsZero() {
			adjustedHoraCierre := horarios[i].HORA_CIERRE.In(database.BogotaZone)
			horarios[i].HORA_CIERRE = &adjustedHoraCierre
		} else {
			horarios[i].HORA_CIERRE = nil
		}
	}

	// Responder con los datos ajustados
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Cambios de horario obtenidos correctamente",
		Data:    horarios,
	}
	c.ServeJSON()
}

// @Title GetByCurrentDate
// @Summary Consultar cambios de horario para la fecha actual
// @Description Obtiene el cambio de horario que aplica para la fecha actual, si existe.
// @Tags cambios_horario
// @Accept json
// @Produce json
// @Success 200 {object} models.CambiosHorario "Cambio de horario para la fecha actual"
// @Failure 404 {object} models.ApiResponse "No hay cambios de horario para la fecha actual"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /cambios_horario/actual [get]
func (c *CambiosHorarioController) GetByCurrentDate() {
	o := orm.NewOrm()
	var cambioHorario models.CambiosHorario

	// Obtener la fecha actual
	currentDate := time.Now().In(database.BogotaZone).Format("2006-01-02")

	// Consultar si hay un cambio de horario para la fecha actual
	err := o.QueryTable(new(models.CambiosHorario)).
		Filter("FECHA", currentDate).
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

	// Ajustar fechas y horas a la zona horaria de Bogotá
	cambioHorario.FECHA = cambioHorario.FECHA.In(database.BogotaZone)
	if cambioHorario.HORA_APERTURA != nil {
		adjustedHoraApertura := cambioHorario.HORA_APERTURA.In(database.BogotaZone)
		cambioHorario.HORA_APERTURA = &adjustedHoraApertura
	}
	if cambioHorario.HORA_CIERRE != nil {
		adjustedHoraCierre := cambioHorario.HORA_CIERRE.In(database.BogotaZone)
		cambioHorario.HORA_CIERRE = &adjustedHoraCierre
	}

	// Responder con el cambio de horario encontrado
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Cambio de horario encontrado para la fecha actual",
		Data:    cambioHorario,
	}
	c.ServeJSON()
}
