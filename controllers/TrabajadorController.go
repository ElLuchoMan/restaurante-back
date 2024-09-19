package controllers

import (
	"encoding/json"
	"net/http"
	"restaurante/models"
	"strconv"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
	"golang.org/x/crypto/bcrypt"
)

type TrabajadorController struct {
	web.Controller
}

// @Title GetAll
// @Summary Obtener todos los trabajadores
// @Description Devuelve todos los trabajadores registrados en la base de datos.
// @Tags trabajadores
// @Accept json
// @Produce json
// @Success 200 {array} models.Trabajador "Lista de trabajadores"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /trabajadores [get]
func (c *TrabajadorController) GetAll() {
	o := orm.NewOrm()
	var trabajadores []models.Trabajador

	_, err := o.QueryTable(new(models.Trabajador)).All(&trabajadores)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener trabajadores de la base de datos",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Excluir la contraseña de la respuesta
	for i := range trabajadores {
		trabajadores[i].PASSWORD = ""
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Trabajadores obtenidos exitosamente",
		Data:    trabajadores,
	}
	c.ServeJSON()
}

// @Title GetById
// @Summary Obtener trabajador por ID
// @Description Devuelve un trabajador específico por ID utilizando query parameters.
// @Tags trabajadores
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID del Trabajador"
// @Success 200 {object} models.Trabajador "Trabajador encontrado"
// @Failure 404 {object} models.ApiResponse "Trabajador no encontrado"
// @Security BearerAuth
// @Router /trabajadores/search [get]
func (c *TrabajadorController) GetById() {
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

	trabajador := models.Trabajador{PK_DOCUMENTO_TRABAJADOR: int64(id)}

	err = o.Read(&trabajador)
	if err == orm.ErrNoRows || err == orm.ErrMissPK {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Trabajador no encontrado",
			Cause:   "No existe un trabajador con el ID proporcionado",
		}
		c.ServeJSON()
		return
	} else if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al buscar el trabajador",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Excluir la contraseña de la respuesta
	trabajador.PASSWORD = ""

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Trabajador encontrado",
		Data:    trabajador,
	}
	c.ServeJSON()
}

// @Title Create
// @Summary Crear un nuevo trabajador
// @Description Crea un nuevo trabajador en la base de datos.
// @Tags trabajadores
// @Accept json
// @Produce json
// @Param   body  body   models.Trabajador true  "Datos del trabajador a crear"
// @Success 201 {object} models.Trabajador "Trabajador creado"
// @Failure 400 {object} models.ApiResponse "Error en la solicitud"
// @Security BearerAuth
// @Router /trabajadores [post]
func (c *TrabajadorController) Post() {
	o := orm.NewOrm()
	var trabajador models.Trabajador

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &trabajador); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Error al decodificar la solicitud",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Validación de campos obligatorios
	if trabajador.PK_DOCUMENTO_TRABAJADOR == 0 || trabajador.NOMBRE == "" || trabajador.APELLIDO == "" {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Faltan campos obligatorios",
			Cause:   " Los campos: Documento, Nombre y Apellido son obligatorios",
		}
		c.ServeJSON()
		return
	}

	// Hash de la contraseña antes de insertar
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(trabajador.PASSWORD), bcrypt.DefaultCost)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al procesar la contraseña",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}
	trabajador.PASSWORD = string(hashedPassword)

	// Inserción en la base de datos
	_, err = o.Insert(&trabajador)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al crear el trabajador",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Excluir la contraseña de la respuesta
	trabajador.PASSWORD = ""

	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusCreated,
		Message: "Trabajador creado correctamente",
		Data:    trabajador,
	}
	c.ServeJSON()
}

// @Title Update
// @Summary Actualizar un trabajador
// @Description Actualiza los datos de un trabajador existente.
// @Tags trabajadores
// @Accept json
// @Produce json
// @Param   id    query    int  true   "ID del Trabajador"
// @Param   body  body   models.Trabajador true  "Datos del trabajador a actualizar"
// @Success 200 {object} models.Trabajador "Trabajador actualizado"
// @Failure 404 {object} models.ApiResponse "Trabajador no encontrado"
// @Security BearerAuth
// @Router /trabajadores [put]
func (c *TrabajadorController) Put() {
	o := orm.NewOrm()

	// Obtener el ID del query parameter
	idStr := c.GetString("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id == 0 {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El parámetro 'id' es inválido o está ausente",
			Cause:   "Se requiere un ID numérico válido en el parámetro 'id'",
		}
		c.ServeJSON()
		return
	}

	// Verificar si el trabajador existe
	trabajador := models.Trabajador{PK_DOCUMENTO_TRABAJADOR: int64(id)}
	if err := o.Read(&trabajador); err != nil {
		if err == orm.ErrNoRows {
			c.Ctx.Output.SetStatus(http.StatusNotFound)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusNotFound,
				Message: "Trabajador no encontrado",
			}
			c.ServeJSON()
		} else {
			c.Ctx.Output.SetStatus(http.StatusInternalServerError)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusInternalServerError,
				Message: "Error al buscar el trabajador",
				Cause:   err.Error(),
			}
			c.ServeJSON()
		}
		return
	}

	// Decodificar los datos actualizados
	var updatedTrabajador models.Trabajador
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &updatedTrabajador); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Error al decodificar la solicitud",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Mantener el ID original
	updatedTrabajador.PK_DOCUMENTO_TRABAJADOR = trabajador.PK_DOCUMENTO_TRABAJADOR

	// Si se proporciona una nueva contraseña, hashéala
	if updatedTrabajador.PASSWORD != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(updatedTrabajador.PASSWORD), bcrypt.DefaultCost)
		if err != nil {
			c.Ctx.Output.SetStatus(http.StatusInternalServerError)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusInternalServerError,
				Message: "Error al procesar la contraseña",
				Cause:   err.Error(),
			}
			c.ServeJSON()
			return
		}
		updatedTrabajador.PASSWORD = string(hashedPassword)
	} else {
		// Mantener la contraseña existente
		updatedTrabajador.PASSWORD = trabajador.PASSWORD
	}

	// Actualizar en la base de datos
	_, err = o.Update(&updatedTrabajador)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al actualizar el trabajador",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Excluir la contraseña de la respuesta
	updatedTrabajador.PASSWORD = ""

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Trabajador actualizado",
		Data:    updatedTrabajador,
	}
	c.ServeJSON()
}

// @Title Delete
// @Summary Eliminar un trabajador
// @Description Elimina un trabajador de la base de datos.
// @Tags trabajadores
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID del Trabajador"
// @Success 200 {object} models.ApiResponse "Trabajador eliminado"
// @Failure 404 {object} models.ApiResponse "Trabajador no encontrado"
// @Security BearerAuth
// @Router /trabajadores [delete]
func (c *TrabajadorController) Delete() {
	o := orm.NewOrm()

	// Obtener el ID del query parameter
	idStr := c.GetString("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id == 0 {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El parámetro 'id' es inválido o está ausente",
			Cause:   "Se requiere un ID numérico válido en el parámetro 'id'",
		}
		c.ServeJSON()
		return
	}

	trabajador := models.Trabajador{PK_DOCUMENTO_TRABAJADOR: int64(id)}

	if num, err := o.Delete(&trabajador); err == nil {
		if num > 0 {
			c.Ctx.Output.SetStatus(http.StatusOK)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusOK,
				Message: "Trabajador eliminado",
			}
			c.ServeJSON()
		} else {
			c.Ctx.Output.SetStatus(http.StatusNotFound)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusNotFound,
				Message: "Trabajador no encontrado",
			}
			c.ServeJSON()
		}
	} else {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al eliminar el trabajador",
			Cause:   err.Error(),
		}
		c.ServeJSON()
	}
}
