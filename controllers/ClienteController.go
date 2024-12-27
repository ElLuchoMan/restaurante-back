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

type ClienteController struct {
	web.Controller
}

// @Title GetAll
// @Summary Obtener todos los clientes con opción de filtrar campos
// @Description Devuelve todos los clientes registrados en la base de datos, con opción de retornar solo nombre completo y teléfono.
// @Tags clientes
// @Accept json
// @Produce json
// @Param   limit  query    int     false  "Cantidad de resultados por página (por defecto es 10)"
// @Param   offset query    int     false  "Número de registros a omitir desde el inicio (por defecto es 0)"
// @Param   fields  query    string  false  "Especifica los campos a incluir en la respuesta (opciones: 'nombre_completo_telefono')"
// @Success 200 {array} interface{} "Lista de clientes con los campos especificados"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /clientes [get]
func (c *ClienteController) GetAll() {
	o := orm.NewOrm()
	var clientes []models.Cliente

	// Obtener el valor del parámetro fields
	fields := c.GetString("fields")

	_, err := o.QueryTable(new(models.Cliente)).All(&clientes)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener clientes de la base de datos",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Manejar la respuesta basada en el parámetro fields
	if fields == "nombre_completo_telefono" {
		var filteredClientes []map[string]string
		for _, cliente := range clientes {
			filteredClientes = append(filteredClientes, map[string]string{
				"nombre_completo": cliente.NOMBRE + " " + cliente.APELLIDO,
				"telefono":        cliente.TELEFONO,
			})
		}

		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusOK,
			Message: "Clientes obtenidos exitosamente",
			Data:    filteredClientes,
		}
		c.ServeJSON()
		return
	}

	// Respuesta completa por defecto, excluyendo las contraseñas
	for i := range clientes {
		clientes[i].PASSWORD = ""
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Clientes obtenidos exitosamente",
		Data:    clientes,
	}
	c.ServeJSON()
}

// @Title GetById
// @Summary Obtener cliente por ID
// @Description Devuelve un cliente específico por ID utilizando query parameters.
// @Tags clientes
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID del Cliente"
// @Success 200 {object} models.Cliente "Cliente encontrado"
// @Failure 404 {object} models.ApiResponse "Cliente no encontrado"
// @Security BearerAuth
// @Router /clientes/search [get]
func (c *ClienteController) GetById() {
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

	cliente := models.Cliente{PK_DOCUMENTO_CLIENTE: id}

	err = o.Read(&cliente)
	if err == orm.ErrNoRows {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Cliente no encontrado",
		}
		c.ServeJSON()
		return
	}

	// Excluir la contraseña de la respuesta
	cliente.PASSWORD = ""

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Cliente encontrado",
		Data:    cliente,
	}
	c.ServeJSON()
}

// @Title Create
// @Summary Crear un nuevo cliente
// @Description Crea un nuevo cliente en la base de datos.
// @Tags clientes
// @Accept json
// @Produce json
// @Param   body  body   models.Cliente true  "Datos del cliente a crear"
// @Success 201 {object} models.Cliente "Cliente creado"
// @Failure 400 {object} models.ApiResponse "Error en la solicitud"
// @Security BearerAuth
// @Router /clientes [post]
func (c *ClienteController) Post() {
	o := orm.NewOrm()
	var cliente models.Cliente

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &cliente); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Error al decodificar la solicitud",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Hash de la contraseña antes de insertar
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(cliente.PASSWORD), bcrypt.DefaultCost)
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
	cliente.PASSWORD = string(hashedPassword)

	// Inserción en la base de datos
	_, err = o.Insert(&cliente)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al crear el cliente",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Excluir la contraseña de la respuesta
	cliente.PASSWORD = ""

	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusCreated,
		Message: "Cliente creado correctamente",
		Data:    cliente,
	}
	c.ServeJSON()
}

// @Title Update
// @Summary Actualizar un cliente
// @Description Actualiza los datos de un cliente existente.
// @Tags clientes
// @Accept json
// @Produce json
// @Param   id    query    int  true   "ID del Cliente"
// @Param   body  body   models.Cliente true  "Datos del cliente a actualizar"
// @Success 200 {object} models.Cliente "Cliente actualizado"
// @Failure 404 {object} models.ApiResponse "Cliente no encontrado"
// @Security BearerAuth
// @Router /clientes [put]
func (c *ClienteController) Put() {
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

	// Verificar si el cliente existe
	cliente := models.Cliente{PK_DOCUMENTO_CLIENTE: id}
	if err := o.Read(&cliente); err != nil {
		if err == orm.ErrNoRows {
			c.Ctx.Output.SetStatus(http.StatusNotFound)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusNotFound,
				Message: "Cliente no encontrado",
			}
			c.ServeJSON()
		} else {
			c.Ctx.Output.SetStatus(http.StatusInternalServerError)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusInternalServerError,
				Message: "Error al buscar el cliente",
				Cause:   err.Error(),
			}
			c.ServeJSON()
		}
		return
	}

	// Decodificar los datos actualizados
	var updatedCliente models.Cliente
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &updatedCliente); err != nil {
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
	updatedCliente.PK_DOCUMENTO_CLIENTE = cliente.PK_DOCUMENTO_CLIENTE

	// Si se proporciona una nueva contraseña, hashéala
	if updatedCliente.PASSWORD != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(updatedCliente.PASSWORD), bcrypt.DefaultCost)
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
		updatedCliente.PASSWORD = string(hashedPassword)
	} else {
		// Mantener la contraseña existente
		updatedCliente.PASSWORD = cliente.PASSWORD
	}

	// Actualizar en la base de datos
	_, err = o.Update(&updatedCliente)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al actualizar el cliente",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Excluir la contraseña de la respuesta
	updatedCliente.PASSWORD = ""

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Cliente actualizado",
		Data:    updatedCliente,
	}
	c.ServeJSON()
}

// @Title Delete
// @Summary Eliminar un cliente
// @Description Elimina un cliente de la base de datos.
// @Tags clientes
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID del Cliente"
// @Success 200 {object} models.ApiResponse "Cliente eliminado"
// @Failure 404 {object} models.ApiResponse "Cliente no encontrado"
// @Security BearerAuth
// @Router /clientes [delete]
func (c *ClienteController) Delete() {
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

	cliente := models.Cliente{PK_DOCUMENTO_CLIENTE: id}

	if _, err := o.Delete(&cliente); err == nil {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusOK,
			Message: "Cliente eliminado",
		}
		c.ServeJSON()
	} else {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Cliente no encontrado",
			Cause:   err.Error(),
		}
		c.ServeJSON()
	}
}
