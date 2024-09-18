package controllers

import (
	"encoding/json"
	"net/http"
	"restaurante/models"
	"strconv"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

type NominaController struct {
	web.Controller
}

// @Title GetAll
// @Summary Obtener todas las nóminas
// @Description Devuelve todas las nóminas registradas en la base de datos.
// @Tags nominas
// @Accept json
// @Produce json
// @Success 200 {array} models.Nomina "Lista de nóminas"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Router /nominas [get]
func (c *NominaController) GetAll() {
	o := orm.NewOrm()
	var nominas []models.Nomina

	_, err := o.QueryTable(new(models.Nomina)).All(&nominas)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener las nóminas de la base de datos",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Nóminas obtenidas exitosamente",
		Data:    nominas,
	}
	c.ServeJSON()
}

// @Title GetById
// @Summary Obtener nómina por ID
// @Description Devuelve una nómina específica por ID.
// @Tags nominas
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID de la Nómina"
// @Success 200 {object} models.Nomina "Nómina encontrada"
// @Failure 404 {object} models.ApiResponse "Nómina no encontrada"
// @Router /nominas/search [get]
func (c *NominaController) GetById() {
	o := orm.NewOrm()
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

	nomina := models.Nomina{PK_ID_NOMINA: id}

	err = o.Read(&nomina)
	if err == orm.ErrNoRows {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Nómina no encontrada",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Nómina encontrada",
		Data:    nomina,
	}
	c.ServeJSON()
}

// @Title Create
// @Summary Crear una nueva nómina
// @Description Crea una nueva nómina en la base de datos.
// @Tags nominas
// @Accept json
// @Produce json
// @Param   body  body   models.Nomina true  "Datos de la nómina a crear"
// @Success 201 {object} models.Nomina "Nómina creada"
// @Failure 400 {object} models.ApiResponse "Error en la solicitud"
// @Router /nominas [post]
func (c *NominaController) Post() {
	o := orm.NewOrm()
	var nomina models.Nomina

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &nomina); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Error en la solicitud",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	id, err := o.Insert(&nomina)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al crear la nómina",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	nomina.PK_ID_NOMINA = id

	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusCreated,
		Message: "Nómina creada correctamente",
		Data:    nomina,
	}
	c.ServeJSON()
}

// @Title Update
// @Summary Actualizar una nómina
// @Description Actualiza los datos de una nómina existente.
// @Tags nominas
// @Accept json
// @Produce json
// @Param   id    query    int  true   "ID de la Nómina"
// @Param   body  body   models.Nomina true  "Datos de la nómina a actualizar"
// @Success 200 {object} models.Nomina "Nómina actualizada"
// @Failure 404 {object} models.ApiResponse "Nómina no encontrada"
// @Router /nominas [put]
func (c *NominaController) Put() {
	o := orm.NewOrm()

	idStr := c.GetString("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
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

	nomina := models.Nomina{PK_ID_NOMINA: id}

	if o.Read(&nomina) == nil {
		var updatedNomina models.Nomina
		if err := json.Unmarshal(c.Ctx.Input.RequestBody, &updatedNomina); err != nil {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusBadRequest,
				Message: "Error en la solicitud",
				Cause:   err.Error(),
			}
			c.ServeJSON()
			return
		}

		updatedNomina.PK_ID_NOMINA = id
		_, err := o.Update(&updatedNomina)
		if err != nil {
			c.Ctx.Output.SetStatus(http.StatusInternalServerError)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusInternalServerError,
				Message: "Error al actualizar la nómina",
				Cause:   err.Error(),
			}
			c.ServeJSON()
			return
		}

		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusOK,
			Message: "Nómina actualizada",
			Data:    updatedNomina,
		}
		c.ServeJSON()
	} else {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Nómina no encontrada",
		}
		c.ServeJSON()
	}
}

// @Title Delete
// @Summary Eliminar una nómina
// @Description Elimina una nómina de la base de datos.
// @Tags nominas
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID de la Nómina"
// @Success 200 {object} models.ApiResponse "Nómina eliminada"
// @Failure 404 {object} models.ApiResponse "Nómina no encontrada"
// @Router /nominas [delete]
func (c *NominaController) Delete() {
	o := orm.NewOrm()

	idStr := c.GetString("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
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

	nomina := models.Nomina{PK_ID_NOMINA: id}

	if _, err := o.Delete(&nomina); err == nil {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusOK,
			Message: "Nómina eliminada",
		}
		c.ServeJSON()
	} else {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Nómina no encontrada",
			Cause:   err.Error(),
		}
		c.ServeJSON()
	}
}
