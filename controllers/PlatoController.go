package controllers

import (
	"encoding/json"
	"net/http"
	"restaurante/models"
	"strconv"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

type PlatoController struct {
	web.Controller
}

// @Title GetAll
// @Summary Obtener todos los platos
// @Description Devuelve todos los platos registrados en la base de datos.
// @Tags platos
// @Accept json
// @Produce json
// @Success 200 {array} models.Plato "Lista de platos"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Router /platos [get]
func (c *PlatoController) GetAll() {
	o := orm.NewOrm()
	var platos []models.Plato

	_, err := o.QueryTable(new(models.Plato)).All(&platos)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener platos de la base de datos",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Platos obtenidos exitosamente",
		Data:    platos,
	}
	c.ServeJSON()
}

// @Title GetById
// @Summary Obtener plato por ID
// @Description Devuelve un plato específico por ID utilizando query parameters.
// @Tags platos
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID del Plato"
// @Success 200 {object} models.Plato "Plato encontrado"
// @Failure 404 {object} models.ApiResponse "Plato no encontrado"
// @Router /platos/search [get]
func (c *PlatoController) GetById() {
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

	plato := models.Plato{PK_ID_PLATO: int64(id)}

	err = o.Read(&plato)
	if err == orm.ErrNoRows {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Plato no encontrado",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Plato encontrado",
		Data:    plato,
	}
	c.ServeJSON()
}

// @Title Create
// @Summary Crear un nuevo plato
// @Description Crea un nuevo plato en la base de datos.
// @Tags platos
// @Accept json
// @Produce json
// @Param   body  body   models.Plato true  "Datos del plato a crear"
// @Success 201 {object} models.Plato "Plato creado"
// @Failure 400 {object} models.ApiResponse "Error en la solicitud"
// @Router /platos [post]
func (c *PlatoController) Post() {
	o := orm.NewOrm()
	var plato models.Plato

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &plato); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Error en la solicitud",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	_, err := o.Insert(&plato)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al crear el plato",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusCreated,
		Message: "Plato creado correctamente",
		Data:    plato,
	}
	c.ServeJSON()
}

// @Title Update
// @Summary Actualizar un plato
// @Description Actualiza los datos de un plato existente.
// @Tags platos
// @Accept json
// @Produce json
// @Param   id    query    int  true   "ID del Plato"
// @Param   body  body   models.Plato true  "Datos del plato a actualizar"
// @Success 200 {object} models.Plato "Plato actualizado"
// @Failure 404 {object} models.ApiResponse "Plato no encontrado"
// @Router /platos [put]
func (c *PlatoController) Put() {
	o := orm.NewOrm()

	// Obtener el ID del query parameter
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

	plato := models.Plato{PK_ID_PLATO: int64(id)}

	if o.Read(&plato) == nil {
		var updatedPlato models.Plato
		if err := json.Unmarshal(c.Ctx.Input.RequestBody, &updatedPlato); err != nil {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusBadRequest,
				Message: "Error en la solicitud",
				Cause:   err.Error(),
			}
			c.ServeJSON()
			return
		}

		updatedPlato.PK_ID_PLATO = int64(id)
		_, err := o.Update(&updatedPlato)
		if err != nil {
			c.Ctx.Output.SetStatus(http.StatusInternalServerError)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusInternalServerError,
				Message: "Error al actualizar el plato",
				Cause:   err.Error(),
			}
			c.ServeJSON()
			return
		}

		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusOK,
			Message: "Plato actualizado",
			Data:    updatedPlato,
		}
		c.ServeJSON()
	} else {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Plato no encontrado",
		}
		c.ServeJSON()
	}
}

// @Title Delete
// @Summary Eliminar un plato
// @Description Elimina un plato de la base de datos.
// @Tags platos
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID del Plato"
// @Success 204 {object} nil "Plato eliminado"
// @Failure 404 {object} models.ApiResponse "Plato no encontrado"
// @Router /platos [delete]
func (c *PlatoController) Delete() {
	o := orm.NewOrm()

	// Obtener el ID del query parameter
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

	plato := models.Plato{PK_ID_PLATO: int64(id)}

	if _, err := o.Delete(&plato); err == nil {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusOK,
			Message: "Plato eliminado",
		}
		c.ServeJSON()
	} else {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Plato no encontrado",
			Cause:   err.Error(),
		}
		c.ServeJSON()
	}
}
