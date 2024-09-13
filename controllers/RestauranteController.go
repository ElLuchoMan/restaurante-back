package controllers

import (
	"encoding/json"
	"net/http"
	"restaurante/models"
	"strconv"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

type RestauranteController struct {
	web.Controller
}

// @Title GetAll
// @Summary Obtener todos los restaurantes
// @Description Devuelve todos los restaurantes registrados en la base de datos.
// @Tags restaurantes
// @Accept json
// @Produce json
// @Success 200 {array} models.Restaurante "Lista de restaurantes"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Router /restaurantes [get]
func (c *RestauranteController) GetAll() {
	o := orm.NewOrm()
	var restaurantes []models.Restaurante

	_, err := o.QueryTable(new(models.Restaurante)).All(&restaurantes)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener restaurantes de la base de datos",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Restaurantes obtenidos exitosamente",
		Data:    restaurantes,
	}
	c.ServeJSON()
}

// @Title GetById
// @Summary Obtener restaurante por ID
// @Description Devuelve un restaurante específico por ID utilizando query parameters.
// @Tags restaurantes
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID del Restaurante"
// @Success 200 {object} models.Restaurante "Restaurante encontrado"
// @Failure 404 {object} models.ApiResponse "Restaurante no encontrado"
// @Router /restaurantes/search [get]
func (c *RestauranteController) GetById() {
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

	restaurante := models.Restaurante{PK_ID_RESTAURANTE: id}

	err = o.Read(&restaurante)
	if err == orm.ErrNoRows {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Restaurante no encontrado",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Restaurante encontrado",
		Data:    restaurante,
	}
	c.ServeJSON()
}

// @Title Create
// @Summary Crear un nuevo restaurante
// @Description Crea un nuevo restaurante en la base de datos.
// @Tags restaurantes
// @Accept json
// @Produce json
// @Param   body  body   models.Restaurante true  "Datos del restaurante a crear"
// @Success 201 {object} models.Restaurante "Restaurante creado"
// @Failure 400 {object} models.ApiResponse "Error en la solicitud"
// @Router /restaurantes [post]
func (c *RestauranteController) Post() {
	o := orm.NewOrm()
	var restaurante models.Restaurante

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &restaurante); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Error en la solicitud",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	_, err := o.Insert(&restaurante)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al crear el restaurante",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusCreated,
		Message: "Restaurante creado correctamente",
		Data:    restaurante,
	}
	c.ServeJSON()
}

// @Title Update
// @Summary Actualizar un restaurante
// @Description Actualiza los datos de un restaurante existente.
// @Tags restaurantes
// @Accept json
// @Produce json
// @Param   id    query    int  true   "ID del Restaurante"
// @Param   body  body   models.Restaurante true  "Datos del restaurante a actualizar"
// @Success 200 {object} models.Restaurante "Restaurante actualizado"
// @Failure 404 {object} models.ApiResponse "Restaurante no encontrado"
// @Router /restaurantes [put]
func (c *RestauranteController) Put() {
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

	restaurante := models.Restaurante{PK_ID_RESTAURANTE: id}

	if o.Read(&restaurante) == nil {
		var updatedRestaurante models.Restaurante
		if err := json.Unmarshal(c.Ctx.Input.RequestBody, &updatedRestaurante); err != nil {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusBadRequest,
				Message: "Error en la solicitud",
				Cause:   err.Error(),
			}
			c.ServeJSON()
			return
		}

		updatedRestaurante.PK_ID_RESTAURANTE = id
		_, err := o.Update(&updatedRestaurante)
		if err != nil {
			c.Ctx.Output.SetStatus(http.StatusInternalServerError)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusInternalServerError,
				Message: "Error al actualizar el restaurante",
				Cause:   err.Error(),
			}
			c.ServeJSON()
			return
		}

		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusOK,
			Message: "Restaurante actualizado",
			Data:    updatedRestaurante,
		}
		c.ServeJSON()
	} else {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Restaurante no encontrado",
		}
		c.ServeJSON()
	}
}

// @Title Delete
// @Summary Eliminar un restaurante
// @Description Elimina un restaurante de la base de datos.
// @Tags restaurantes
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID del Restaurante"
// @Success 204 {object} nil "Restaurante eliminado"
// @Failure 404 {object} models.ApiResponse "Restaurante no encontrado"
// @Router /restaurantes [delete]
func (c *RestauranteController) Delete() {
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

	restaurante := models.Restaurante{PK_ID_RESTAURANTE: id}

	if _, err := o.Delete(&restaurante); err == nil {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusOK,
			Message: "Restaurante eliminado",
		}
		c.ServeJSON()
	} else {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Restaurante no encontrado",
			Cause:   err.Error(),
		}
		c.ServeJSON()
	}
}
