package controllers

import (
	"io/ioutil"
	"net/http"
	"restaurante/models"
	"strconv"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

type IngredienteController struct {
	web.Controller
}

// @Title GetAll
// @Summary Obtener todos los ingredientes
// @Description Devuelve todos los ingredientes registrados en la base de datos.
// @Tags ingredientes
// @Accept json
// @Produce json
// @Success 200 {array} models.Ingrediente "Lista de ingredientes"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Router /ingredientes [get]
func (c *IngredienteController) GetAll() {
	o := orm.NewOrm()
	var ingredientes []models.Ingrediente

	_, err := o.QueryTable(new(models.Ingrediente)).All(&ingredientes)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener ingredientes de la base de datos",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Ingredientes obtenidos exitosamente",
		Data:    ingredientes,
	}
	c.ServeJSON()
}

// @Title GetById
// @Summary Obtener ingrediente por ID
// @Description Devuelve un ingrediente específico por ID utilizando query parameters.
// @Tags ingredientes
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID del Ingrediente"
// @Success 200 {object} models.Ingrediente "Ingrediente encontrado"
// @Failure 404 {object} models.ApiResponse "Ingrediente no encontrado"
// @Router /ingredientes/search [get]
func (c *IngredienteController) GetById() {
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

	ingrediente := models.Ingrediente{PK_ID_INGREDIENTE: id}

	err = o.Read(&ingrediente)
	if err == orm.ErrNoRows {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Ingrediente no encontrado",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Ingrediente encontrado",
		Data:    ingrediente,
	}
	c.ServeJSON()
}

// @Title Create
// @Summary Crear un nuevo ingrediente
// @Description Crea un nuevo ingrediente en la base de datos.
// @Tags ingredientes
// @Accept multipart/form-data
// @Produce json
// @Param   NOMBRE        formData  string  true   "Nombre del ingrediente"
// @Param   TIPO          formData  string  true   "Tipo del ingrediente"
// @Param   PESO          formData  int     true   "Peso del ingrediente"
// @Param   CALORIAS      formData  int     true   "Calorías del ingrediente"
// @Param   FOTO          formData  file    false  "Imagen del ingrediente (opcional)"
// @Success 201 {object} models.Ingrediente "Ingrediente creado"
// @Failure 400 {object} models.ApiResponse "Error en la solicitud"
// @Router /ingredientes [post]
func (c *IngredienteController) Post() {
	o := orm.NewOrm()
	var ingrediente models.Ingrediente

	// Obtener los campos del formulario
	ingrediente.NOMBRE = c.GetString("NOMBRE")
	ingrediente.TIPO = c.GetString("TIPO")
	ingrediente.PESO, _ = c.GetInt64("PESO")
	ingrediente.CALORIAS, _ = c.GetInt64("CALORIAS")

	// Obtener el archivo de imagen
	file, _, err := c.GetFile("FOTO")
	if err == nil {
		defer file.Close()
		// Leer el contenido del archivo
		fileBytes, err := ioutil.ReadAll(file)
		if err == nil {
			ingrediente.FOTO = string(fileBytes)
		}
	}

	// Insertar el nuevo ingrediente en la base de datos
	id, err := o.Insert(&ingrediente)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al crear el ingrediente",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	ingrediente.PK_ID_INGREDIENTE = id

	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusCreated,
		Message: "Ingrediente creado correctamente",
		Data:    ingrediente,
	}
	c.ServeJSON()
}

// @Title Update
// @Summary Actualizar un ingrediente
// @Description Actualiza los datos de un ingrediente existente.
// @Tags ingredientes
// @Accept multipart/form-data
// @Produce json
// @Param   id            query    int     true   "ID del Ingrediente"
// @Param   NOMBRE        formData  string  true   "Nombre del ingrediente"
// @Param   TIPO          formData  string  true   "Tipo del ingrediente"
// @Param   PESO          formData  int     true   "Peso del ingrediente"
// @Param   CALORIAS      formData  int     true   "Calorías del ingrediente"
// @Param   FOTO          formData  file    false  "Imagen del ingrediente (opcional)"
// @Success 200 {object} models.Ingrediente "Ingrediente actualizado"
// @Failure 404 {object} models.ApiResponse "Ingrediente no encontrado"
// @Router /ingredientes [put]
func (c *IngredienteController) Put() {
	o := orm.NewOrm()

	// Obtener el ID del query parameter
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

	ingrediente := models.Ingrediente{PK_ID_INGREDIENTE: id}

	if o.Read(&ingrediente) == nil {
		// Actualizar los campos del formulario
		ingrediente.NOMBRE = c.GetString("NOMBRE")
		ingrediente.TIPO = c.GetString("TIPO")
		ingrediente.PESO, _ = c.GetInt64("PESO")
		ingrediente.CALORIAS, _ = c.GetInt64("CALORIAS")

		// Obtener el archivo de imagen si fue enviado
		file, _, err := c.GetFile("FOTO")
		if err == nil {
			defer file.Close()
			// Leer el contenido del archivo
			fileBytes, err := ioutil.ReadAll(file)
			if err == nil {
				ingrediente.FOTO = string(fileBytes)
			}
		}

		// Actualizar el ingrediente en la base de datos
		_, err = o.Update(&ingrediente)
		if err != nil {
			c.Ctx.Output.SetStatus(http.StatusInternalServerError)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusInternalServerError,
				Message: "Error al actualizar el ingrediente",
				Cause:   err.Error(),
			}
			c.ServeJSON()
			return
		}

		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusOK,
			Message: "Ingrediente actualizado",
			Data:    ingrediente,
		}
		c.ServeJSON()
	} else {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Ingrediente no encontrado",
		}
		c.ServeJSON()
	}
}

// @Title Delete
// @Summary Eliminar un ingrediente
// @Description Elimina un ingrediente de la base de datos.
// @Tags ingredientes
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID del Ingrediente"
// @Success 200 {object} models.ApiResponse "Ingrediente eliminado"
// @Failure 404 {object} models.ApiResponse "Ingrediente no encontrado"
// @Router /ingredientes [delete]
func (c *IngredienteController) Delete() {
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

	ingrediente := models.Ingrediente{PK_ID_INGREDIENTE: id}

	if _, err := o.Delete(&ingrediente); err == nil {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusOK,
			Message: "Ingrediente eliminado",
		}
		c.ServeJSON()
	} else {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Ingrediente no encontrado",
			Cause:   err.Error(),
		}
		c.ServeJSON()
	}
}
