package controllers

import (
	"encoding/json"
	"net/http"
	"restaurante/models"
	"strconv"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

type PlatoIngredienteController struct {
	web.Controller
}

// @Title GetAll
// @Summary Obtener todas las relaciones plato-ingrediente
// @Description Devuelve todas las relaciones entre platos e ingredientes.
// @Tags plato_ingredientes
// @Accept json
// @Produce json
// @Success 200 {array} models.PlatoIngrediente "Lista de relaciones"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Router /plato_ingredientes [get]
func (c *PlatoIngredienteController) GetAll() {
	o := orm.NewOrm()
	var relaciones []models.PlatoIngrediente

	_, err := o.QueryTable(new(models.PlatoIngrediente)).All(&relaciones)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener las relaciones de la base de datos",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Relaciones obtenidas exitosamente",
		Data:    relaciones,
	}
	c.ServeJSON()
}

// @Title GetById
// @Summary Obtener relación por ID
// @Description Devuelve una relación específica por ID.
// @Tags plato_ingredientes
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID de la Relación"
// @Success 200 {object} models.PlatoIngrediente "Relación encontrada"
// @Failure 404 {object} models.ApiResponse "Relación no encontrada"
// @Router /plato_ingredientes/search [get]
func (c *PlatoIngredienteController) GetById() {
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

	relacion := models.PlatoIngrediente{PK_ID_PLATO_INGREDIENTE: id}

	err = o.Read(&relacion)
	if err == orm.ErrNoRows {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Relación no encontrada",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Relación encontrada",
		Data:    relacion,
	}
	c.ServeJSON()
}

// @Title Create
// @Summary Crear una nueva relación plato-ingrediente
// @Description Crea una nueva relación entre un plato y un ingrediente.
// @Tags plato_ingredientes
// @Accept json
// @Produce json
// @Param   body  body   models.PlatoIngrediente true  "Datos de la relación a crear"
// @Success 201 {object} models.PlatoIngrediente "Relación creada"
// @Failure 400 {object} models.ApiResponse "Error en la solicitud"
// @Router /plato_ingredientes [post]
func (c *PlatoIngredienteController) Post() {
	o := orm.NewOrm()
	var relacion models.PlatoIngrediente

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &relacion); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Error en la solicitud",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	id, err := o.Insert(&relacion)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al crear la relación",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	relacion.PK_ID_PLATO_INGREDIENTE = id

	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusCreated,
		Message: "Relación creada correctamente",
		Data:    relacion,
	}
	c.ServeJSON()
}

// @Title Update
// @Summary Actualizar una relación plato-ingrediente
// @Description Actualiza los datos de una relación existente.
// @Tags plato_ingredientes
// @Accept json
// @Produce json
// @Param   id    query    int  true   "ID de la Relación"
// @Param   body  body   models.PlatoIngrediente true  "Datos de la relación a actualizar"
// @Success 200 {object} models.PlatoIngrediente "Relación actualizada"
// @Failure 404 {object} models.ApiResponse "Relación no encontrada"
// @Router /plato_ingredientes [put]
func (c *PlatoIngredienteController) Put() {
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

	relacion := models.PlatoIngrediente{PK_ID_PLATO_INGREDIENTE: id}

	if o.Read(&relacion) == nil {
		var updatedRelacion models.PlatoIngrediente
		if err := json.Unmarshal(c.Ctx.Input.RequestBody, &updatedRelacion); err != nil {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusBadRequest,
				Message: "Error en la solicitud",
				Cause:   err.Error(),
			}
			c.ServeJSON()
			return
		}

		updatedRelacion.PK_ID_PLATO_INGREDIENTE = id
		_, err := o.Update(&updatedRelacion)
		if err != nil {
			c.Ctx.Output.SetStatus(http.StatusInternalServerError)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusInternalServerError,
				Message: "Error al actualizar la relación",
				Cause:   err.Error(),
			}
			c.ServeJSON()
			return
		}

		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusOK,
			Message: "Relación actualizada",
			Data:    updatedRelacion,
		}
		c.ServeJSON()
	} else {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Relación no encontrada",
		}
		c.ServeJSON()
	}
}

// @Title Delete
// @Summary Eliminar una relación plato-ingrediente
// @Description Elimina una relación de la base de datos.
// @Tags plato_ingredientes
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID de la Relación"
// @Success 200 {object} models.ApiResponse "Relación eliminada"
// @Failure 404 {object} models.ApiResponse "Relación no encontrada"
// @Router /plato_ingredientes [delete]
func (c *PlatoIngredienteController) Delete() {
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

	relacion := models.PlatoIngrediente{PK_ID_PLATO_INGREDIENTE: id}

	if _, err := o.Delete(&relacion); err == nil {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusOK,
			Message: "Relación eliminada",
		}
		c.ServeJSON()
	} else {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Relación no encontrada",
			Cause:   err.Error(),
		}
		c.ServeJSON()
	}
}
