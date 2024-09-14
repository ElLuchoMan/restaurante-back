package controllers

import (
	"encoding/json"
	"net/http"
	"restaurante/models"
	"strconv"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

type DomicilioController struct {
	web.Controller
}

// @Title GetAll
// @Summary Obtener todos los domicilios
// @Description Devuelve todos los domicilios registrados en la base de datos.
// @Tags domicilios
// @Accept json
// @Produce json
// @Success 200 {array} models.Domicilio "Lista de domicilios"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Router /domicilios [get]
func (c *DomicilioController) GetAll() {
	o := orm.NewOrm()
	var domicilios []models.Domicilio

	_, err := o.QueryTable(new(models.Domicilio)).All(&domicilios)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener domicilios de la base de datos",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Domicilios obtenidos exitosamente",
		Data:    domicilios,
	}
	c.ServeJSON()
}

// @Title GetById
// @Summary Obtener domicilio por ID
// @Description Devuelve un domicilio específico por ID utilizando query parameters.
// @Tags domicilios
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID del Domicilio"
// @Success 200 {object} models.Domicilio "Domicilio encontrado"
// @Failure 404 {object} models.ApiResponse "Domicilio no encontrado"
// @Router /domicilios/search [get]
func (c *DomicilioController) GetById() {
	o := orm.NewOrm()
	id, err := c.GetInt("id")

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

	domicilio := models.Domicilio{PK_ID_DOMICILIO: id}

	err = o.Read(&domicilio)
	if err == orm.ErrNoRows {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Domicilio no encontrado",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Domicilio encontrado",
		Data:    domicilio,
	}
	c.ServeJSON()
}

// @Title Create
// @Summary Crear un nuevo domicilio
// @Description Crea un nuevo domicilio en la base de datos.
// @Tags domicilios
// @Accept json
// @Produce json
// @Param   body  body   models.Domicilio true  "Datos del domicilio a crear"
// @Success 201 {object} models.Domicilio "Domicilio creado"
// @Failure 400 {object} models.ApiResponse "Error en la solicitud"
// @Router /domicilios [post]
func (c *DomicilioController) Post() {
	o := orm.NewOrm()
	var domicilio models.Domicilio

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &domicilio); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Error en la solicitud",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	_, err := o.Insert(&domicilio)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al crear el domicilio",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusCreated,
		Message: "Domicilio creado correctamente",
		Data:    domicilio,
	}
	c.ServeJSON()
}

// @Title Update
// @Summary Actualizar un domicilio
// @Description Actualiza los datos de un domicilio existente.
// @Tags domicilios
// @Accept json
// @Produce json
// @Param   id    query    int  true   "ID del Domicilio"
// @Param   body  body   models.Domicilio true  "Datos del domicilio a actualizar"
// @Success 200 {object} models.Domicilio "Domicilio actualizado"
// @Failure 404 {object} models.ApiResponse "Domicilio no encontrado"
// @Router /domicilios [put]
func (c *DomicilioController) Put() {
	o := orm.NewOrm()

	idStr := c.GetString("id")
	id, err := strconv.Atoi(idStr)
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

	domicilio := models.Domicilio{PK_ID_DOMICILIO: id}

	if o.Read(&domicilio) == nil {
		var updatedDomicilio models.Domicilio
		if err := json.Unmarshal(c.Ctx.Input.RequestBody, &updatedDomicilio); err != nil {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusBadRequest,
				Message: "Error en la solicitud",
				Cause:   err.Error(),
			}
			c.ServeJSON()
			return
		}

		updatedDomicilio.PK_ID_DOMICILIO = id
		_, err := o.Update(&updatedDomicilio)
		if err != nil {
			c.Ctx.Output.SetStatus(http.StatusInternalServerError)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusInternalServerError,
				Message: "Error al actualizar el domicilio",
				Cause:   err.Error(),
			}
			c.ServeJSON()
			return
		}

		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusOK,
			Message: "Domicilio actualizado",
			Data:    updatedDomicilio,
		}
		c.ServeJSON()
	} else {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Domicilio no encontrado",
			Cause:   err.Error(),
		}
		c.ServeJSON()
	}
}

// @Title Delete
// @Summary Eliminar un domicilio
// @Description Elimina un domicilio de la base de datos.
// @Tags domicilios
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID del Domicilio"
// @Success 204 {object} nil "Domicilio eliminado"
// @Failure 404 {object} models.ApiResponse "Domicilio no encontrado"
// @Router /domicilios [delete]
func (c *DomicilioController) Delete() {
	o := orm.NewOrm()

	idStr := c.GetString("id")
	id, err := strconv.Atoi(idStr)
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

	domicilio := models.Domicilio{PK_ID_DOMICILIO: id}

	if _, err := o.Delete(&domicilio); err == nil {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusOK,
			Message: "Domicilio eliminado",
		}
		c.ServeJSON()
	} else {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Domicilio no encontrado",
			Cause:   err.Error(),
		}
		c.ServeJSON()
	}
}
