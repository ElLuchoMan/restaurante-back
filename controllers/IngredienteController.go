package controllers

import (
	"encoding/base64"
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
// @Description Devuelve todos los ingredientes registrados en la base de datos, independientemente de su estado.
// @Tags ingredientes
// @Accept json
// @Produce json
// @Success 200 {array} models.Ingrediente "Lista de todos los ingredientes"
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

	// Excluir la foto (FOTO) de cada ingrediente
	for i := range ingredientes {
		ingredientes[i].FOTO = ""
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Todos los ingredientes obtenidos exitosamente",
		Data:    ingredientes,
	}
	c.ServeJSON()
}

// @Title GetAllActive
// @Summary Obtener todos los ingredientes activos
// @Description Devuelve solo los ingredientes que están activos (ACTIVO = TRUE).
// @Tags ingredientes
// @Accept json
// @Produce json
// @Success 200 {array} models.Ingrediente "Lista de ingredientes activos"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Router /ingredientes/active [get]
func (c *IngredienteController) GetAllActive() {
	o := orm.NewOrm()
	var ingredientes []models.Ingrediente

	_, err := o.QueryTable(new(models.Ingrediente)).Filter("ACTIVO", true).All(&ingredientes)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener ingredientes activos de la base de datos",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Excluir la foto (FOTO) de cada ingrediente
	for i := range ingredientes {
		ingredientes[i].FOTO = ""
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Ingredientes activos obtenidos exitosamente",
		Data:    ingredientes,
	}
	c.ServeJSON()
}

// @Title GetById
// @Summary Obtener ingrediente por ID
// @Description Devuelve un ingrediente específico por ID, incluyendo la imagen en formato Base64.
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
// @Description Crea un nuevo ingrediente en la base de datos, incluyendo una imagen en formato Base64.
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

	// Validar campos requeridos
	ingrediente.NOMBRE = c.GetString("NOMBRE")
	if ingrediente.NOMBRE == "" {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El campo 'NOMBRE' es requerido.",
		}
		c.ServeJSON()
		return
	}

	ingrediente.TIPO = c.GetString("TIPO")
	ingrediente.PESO, _ = c.GetInt64("PESO")
	ingrediente.CALORIAS, _ = c.GetInt64("CALORIAS")

	// Manejar la imagen opcional
	file, fileHeader, err := c.GetFile("FOTO")
	if err == nil {
		defer file.Close()

		// Validar el tamaño del archivo (máximo 1MB)
		if fileHeader.Size > 1024*1024 { // 1MB en bytes
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusBadRequest,
				Message: "La imagen no debe superar los 1MB.",
			}
			c.ServeJSON()
			return
		}

		// Leer y convertir a Base64
		fileBytes, err := ioutil.ReadAll(file)
		if err != nil {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusBadRequest,
				Message: "Error al leer la imagen.",
				Cause:   err.Error(),
			}
			c.ServeJSON()
			return
		}
		ingrediente.FOTO = base64.StdEncoding.EncodeToString(fileBytes)
	}

	// Insertar el nuevo ingrediente en la base de datos
	_, err = o.Insert(&ingrediente)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al crear el ingrediente.",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

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
// @Description Actualiza los datos de un ingrediente existente, incluyendo una imagen en formato Base64.
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
			Message: "El parámetro 'id' es inválido o está ausente.",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	ingrediente := models.Ingrediente{PK_ID_INGREDIENTE: id}

	if o.Read(&ingrediente) == nil {
		ingrediente.NOMBRE = c.GetString("NOMBRE")
		ingrediente.TIPO = c.GetString("TIPO")
		ingrediente.PESO, _ = c.GetInt64("PESO")
		ingrediente.CALORIAS, _ = c.GetInt64("CALORIAS")

		// Manejar la imagen opcional
		file, fileHeader, err := c.GetFile("FOTO")
		if err == nil {
			defer file.Close()

			// Validar el tamaño del archivo (máximo 1MB)
			if fileHeader.Size > 1024*1024 { // 1MB en bytes
				c.Ctx.Output.SetStatus(http.StatusBadRequest)
				c.Data["json"] = models.ApiResponse{
					Code:    http.StatusBadRequest,
					Message: "La imagen no debe superar los 1MB.",
				}
				c.ServeJSON()
				return
			}

			fileBytes, err := ioutil.ReadAll(file)
			if err != nil {
				c.Ctx.Output.SetStatus(http.StatusBadRequest)
				c.Data["json"] = models.ApiResponse{
					Code:    http.StatusBadRequest,
					Message: "Error al leer la imagen.",
					Cause:   err.Error(),
				}
				c.ServeJSON()
				return
			}
			ingrediente.FOTO = base64.StdEncoding.EncodeToString(fileBytes)
		}

		// Actualizar el ingrediente en la base de datos
		_, err = o.Update(&ingrediente)
		if err != nil {
			c.Ctx.Output.SetStatus(http.StatusInternalServerError)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusInternalServerError,
				Message: "Error al actualizar el ingrediente.",
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
			Message: "Ingrediente no encontrado.",
		}
		c.ServeJSON()
	}
}

// @Title Delete
// @Summary Desactivar un ingrediente
// @Description Marca un ingrediente como inactivo (borrado lógico).
// @Tags ingredientes
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID del Ingrediente"
// @Success 200 {object} models.ApiResponse "Ingrediente desactivado"
// @Failure 404 {object} models.ApiResponse "Ingrediente no encontrado"
// @Router /ingredientes [delete]
func (c *IngredienteController) Delete() {
	o := orm.NewOrm()

	// Obtener el ID del query parameter
	idStr := c.GetString("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id == 0 {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El parámetro 'id' es inválido o está ausente",
		}
		c.ServeJSON()
		return
	}

	ingrediente := models.Ingrediente{PK_ID_INGREDIENTE: id}

	if o.Read(&ingrediente) == nil {
		ingrediente.ACTIVO = false
		_, err := o.Update(&ingrediente)
		if err != nil {
			c.Ctx.Output.SetStatus(http.StatusInternalServerError)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusInternalServerError,
				Message: "Error al desactivar el ingrediente.",
				Cause:   err.Error(),
			}
			c.ServeJSON()
			return
		}

		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusOK,
			Message: "Ingrediente desactivado",
		}
		c.ServeJSON()
	} else {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Ingrediente no encontrado.",
		}
		c.ServeJSON()
	}
}
