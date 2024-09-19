package controllers

import (
	"encoding/json"
	"net/http"
	"restaurante/models"
	"strconv"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

type InventarioController struct {
	web.Controller
}

// @Title GetAll
// @Summary Obtener todos los inventarios
// @Description Devuelve todos los registros de inventario.
// @Tags inventarios
// @Accept json
// @Produce json
// @Success 200 {array} models.Inventario "Lista de inventarios"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /inventarios [get]
func (c *InventarioController) GetAll() {
	o := orm.NewOrm()
	var inventarios []models.Inventario

	_, err := o.QueryTable(new(models.Inventario)).All(&inventarios)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener inventarios de la base de datos",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Inventarios obtenidos exitosamente",
		Data:    inventarios,
	}
	c.ServeJSON()
}

// @Title GetById
// @Summary Obtener inventario por ID
// @Description Devuelve un registro de inventario específico por ID.
// @Tags inventarios
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID del Inventario"
// @Success 200 {object} models.Inventario "Inventario encontrado"
// @Failure 404 {object} models.ApiResponse "Inventario no encontrado"
// @Security BearerAuth
// @Router /inventarios/search [get]
func (c *InventarioController) GetById() {
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

	inventario := models.Inventario{PK_ID_INVENTARIO: id}

	err = o.Read(&inventario)
	if err == orm.ErrNoRows {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Inventario no encontrado",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Inventario encontrado",
		Data:    inventario,
	}
	c.ServeJSON()
}

// @Title Create
// @Summary Crear un nuevo inventario
// @Description Crea un nuevo registro de inventario.
// @Tags inventarios
// @Accept json
// @Produce json
// @Param   body  body   models.Inventario true  "Datos del inventario a crear"
// @Success 201 {object} models.Inventario "Inventario creado"
// @Failure 400 {object} models.ApiResponse "Error en la solicitud"
// @Security BearerAuth
// @Router /inventarios [post]
func (c *InventarioController) Post() {
	o := orm.NewOrm()
	var inventario models.Inventario

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &inventario); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Error en la solicitud",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	id, err := o.Insert(&inventario)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al crear el inventario",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	inventario.PK_ID_INVENTARIO = id

	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusCreated,
		Message: "Inventario creado correctamente",
		Data:    inventario,
	}
	c.ServeJSON()
}

// @Title Update
// @Summary Actualizar un inventario
// @Description Actualiza los datos de un inventario existente.
// @Tags inventarios
// @Accept json
// @Produce json
// @Param   id    query    int  true   "ID del Inventario"
// @Param   body  body   models.Inventario true  "Datos del inventario a actualizar"
// @Success 200 {object} models.Inventario "Inventario actualizado"
// @Failure 404 {object} models.ApiResponse "Inventario no encontrado"
// @Security BearerAuth
// @Router /inventarios [put]
func (c *InventarioController) Put() {
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

	inventario := models.Inventario{PK_ID_INVENTARIO: id}

	if o.Read(&inventario) == nil {
		var updatedInventario models.Inventario
		if err := json.Unmarshal(c.Ctx.Input.RequestBody, &updatedInventario); err != nil {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusBadRequest,
				Message: "Error en la solicitud",
				Cause:   err.Error(),
			}
			c.ServeJSON()
			return
		}

		updatedInventario.PK_ID_INVENTARIO = id
		_, err := o.Update(&updatedInventario)
		if err != nil {
			c.Ctx.Output.SetStatus(http.StatusInternalServerError)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusInternalServerError,
				Message: "Error al actualizar el inventario",
				Cause:   err.Error(),
			}
			c.ServeJSON()
			return
		}

		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusOK,
			Message: "Inventario actualizado",
			Data:    updatedInventario,
		}
		c.ServeJSON()
	} else {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Inventario no encontrado",
		}
		c.ServeJSON()
	}
}

// @Title Delete
// @Summary Eliminar un inventario
// @Description Elimina un registro de inventario de la base de datos.
// @Tags inventarios
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID del Inventario"
// @Success 200 {object} models.ApiResponse "Inventario eliminado"
// @Failure 404 {object} models.ApiResponse "Inventario no encontrado"
// @Security BearerAuth
// @Router /inventarios [delete]
func (c *InventarioController) Delete() {
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

	inventario := models.Inventario{PK_ID_INVENTARIO: id}

	if _, err := o.Delete(&inventario); err == nil {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusOK,
			Message: "Inventario eliminado",
		}
		c.ServeJSON()
	} else {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Inventario no encontrado",
			Cause:   err.Error(),
		}
		c.ServeJSON()
	}
}
