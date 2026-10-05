package controlnomina

import (
	"errors"
	"net/http"

	"restaurante/internal/httpx"
	"restaurante/logging"
	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

// ControlNominaController expone el control de generación de nóminas. Las
// respuestas siempre llevan el envoltorio models.ApiResponse y el status HTTP
// coincide con su `code`.
type ControlNominaController struct{ web.Controller }

// @Title GetAll
// @Summary Listar control de nómina
// @Description Devuelve los registros de control de nómina (estado de generación por fecha), opcionalmente filtrados por fecha. Petición: `fecha` en YYYY-MM-DD; respuesta: `fecha` en DD-MM-YYYY. Sin resultados: 200 con `data: []`.
// @Tags control_nomina
// @Accept json
// @Produce json
// @Param fecha query string false "Fecha exacta, formato YYYY-MM-DD"
// @Success 200 {object} models.ApiResponse{data=[]models.ControlNominaResponse} "Registros de control (puede ser [])"
// @Failure 400 {object} models.ApiResponse "fecha con formato inválido"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /control_nomina [get]
func (c *ControlNominaController) GetAll() {
	qs := orm.NewOrm().QueryTable(new(models.ControlNomina))
	if f := c.GetString("fecha"); f != "" {
		d, err := models.ParseDateToNoonUTC(f)
		if err != nil {
			httpx.Fail(&c.Controller, http.StatusBadRequest, "El parámetro 'fecha' debe tener el formato YYYY-MM-DD", err)
			return
		}
		qs = qs.Filter("Fecha", d)
	}
	var list []models.ControlNomina
	if _, err := qs.All(&list); err != nil {
		logging.LogControllerError(c.Ctx, "control_nomina.getall.db_error", err, nil)
		httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error al obtener control de nómina", err)
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, "Control de nómina", httpx.List(list))
}

// @Title GetById
// @Summary Obtener control de nómina por ID
// @Description Devuelve un registro de control de nómina por su ID. `fecha` se responde como DD-MM-YYYY.
// @Tags control_nomina
// @Accept json
// @Produce json
// @Param id query int true "ID del control (entero positivo)"
// @Success 200 {object} models.ApiResponse{data=models.ControlNominaResponse} "Registro encontrado"
// @Failure 400 {object} models.ApiResponse "id ausente o inválido"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 404 {object} models.ApiResponse "Registro no encontrado"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /control_nomina/search [get]
func (c *ControlNominaController) GetById() {
	id, err := httpx.PositiveInt64Param(&c.Controller, "id")
	if err != nil {
		httpx.Fail(&c.Controller, http.StatusBadRequest, "El parámetro 'id' es inválido o está ausente", err)
		return
	}
	var row models.ControlNomina
	if err := orm.NewOrm().QueryTable(new(models.ControlNomina)).Filter("PK_ID_CONTROL_NOMINA", id).One(&row); err != nil {
		if errors.Is(err, orm.ErrNoRows) {
			httpx.Fail(&c.Controller, http.StatusNotFound, "Registro no encontrado", err)
			return
		}
		logging.LogControllerError(c.Ctx, "control_nomina.getbyid.db_error", err, map[string]interface{}{"id": id})
		httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error al obtener el registro", err)
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, "Registro encontrado", row)
}
