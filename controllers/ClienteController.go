package controllers

import (
	"encoding/json"
	"net/http"
	"restaurante/models"
	"strings"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
	"golang.org/x/crypto/bcrypt"
)

var ormNew = orm.NewOrm

var queryAllClientes = func(o orm.Ormer, clientes *[]models.Cliente) (int64, error) {
	return o.QueryTable(new(models.Cliente)).All(clientes)
}

var readCliente = func(o orm.Ormer, c *models.Cliente) error {
	return o.Read(c)
}

var insertCliente = func(o orm.Ormer, c *models.Cliente) (int64, error) {
	return o.Insert(c)
}

var updateCliente = func(o orm.Ormer, c *models.Cliente) (int64, error) {
	return o.Update(c)
}

var deleteCliente = func(o orm.Ormer, c *models.Cliente) (int64, error) {
	return o.Delete(c)
}

var bcryptGenerate = bcrypt.GenerateFromPassword

type ClienteController struct {
	web.Controller
}

func normalizeEmail(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func isUniqueEmailErr(err error) bool {
	if err == nil {
		return false
	}
	// Detecta violaciones de unicidad por nombre de índice o mensaje
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "uq_cliente_correo") ||
		(strings.Contains(msg, "unique") && strings.Contains(msg, "correo"))
}

// @Title GetAll
// @Summary Obtener todos los clientes con opción de filtrar campos
// @Description Devuelve todos los clientes registrados en la base de datos, con opción de retornar solo nombre completo y teléfono.
// @Tags clientes
// @Accept json
// @Produce json
// @Param   limit  query    int     false  "Cantidad de resultados por página" minimum(1) maximum(100) default(10)
// @Param   offset query    int     false  "Número de registros a omitir" minimum(0) default(0)
// @Param   fields  query    string  false  "Campos a incluir en la respuesta (opciones: 'nombre_completo_telefono')"
// @Success 200 {object} models.ApiResponse{data=[]models.Cliente} "Lista de clientes"
// @Failure 401 {object} models.ApiResponse "No autorizado"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /clientes [get]
func (c *ClienteController) GetAll() {
	o := ormNew()
	var clientes []models.Cliente

	// Obtener el valor del parámetro fields
	fields := c.GetString("fields")

	// Si se especificó `limit` o `offset` en la query, aplícalos.
	// Si no se especifican, usamos el helper `queryAllClientes` para mantener
	// la capacidad de mock en tests.
	limitStr := c.GetString("limit")
	offsetStr := c.GetString("offset")

	if limitStr != "" || offsetStr != "" {
		// Parsear valores; GetInt devuelve 0 en error, tomamos 10 como default solo
		// cuando el parámetro `limit` fue provisto pero igual a 0.
		limit, _ := c.GetInt("limit")
		offset, _ := c.GetInt("offset")
		if limit == 0 {
			limit = 10
		}

		qs := o.QueryTable(new(models.Cliente))
		if _, err := qs.Limit(limit, offset).All(&clientes); err != nil {
			c.Ctx.Output.SetStatus(http.StatusInternalServerError)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusInternalServerError,
				Message: "Error al obtener clientes de la base de datos",
				Cause:   err.Error(),
			}
			c.ServeJSON()
			return
		}
	} else {
		// Modo por defecto (sin paginado): delegar al helper (útil para tests)
		_, err := queryAllClientes(o, &clientes)
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
// @Success 200 {object} models.ApiResponse{data=models.Cliente} "Cliente encontrado"
// @Failure 401 {object} models.ApiResponse "No autorizado"
// @Failure 404 {object} models.ApiResponse "Cliente no encontrado"
// @Security BearerAuth
// @Router /clientes/search [get]
func (c *ClienteController) GetById() {
	o := ormNew()
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

	cliente := models.Cliente{PK_DOCUMENTO_CLIENTE: id}

	err = readCliente(o, &cliente)
	if err == orm.ErrNoRows {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Cliente no encontrado",
		}
		c.ServeJSON()
		return
	} else if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al consultar el cliente",
			Cause:   err.Error(),
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
// @Param   body  body   models.ClienteCreateRequest true  "Datos del cliente a crear"
// @Success 201 {object} models.ApiResponse{data=models.Cliente} "Cliente creado"
// @Failure 400 {object} models.ApiResponse "Error en la solicitud"
// @Failure 409 {object} models.ApiResponse "Correo ya registrado"
// @Router /clientes [post]
func (c *ClienteController) Post() {
	o := ormNew()
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

	// Validar y normalizar correo
	correoTrim := strings.TrimSpace(cliente.CORREO)
	if correoTrim == "" {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El campo correo es obligatorio",
		}
		c.ServeJSON()
		return
	}
	cliente.CORREO = normalizeEmail(correoTrim)

	// Hash de la contraseña antes de insertar
	hashedPassword, err := bcryptGenerate([]byte(cliente.PASSWORD), bcrypt.DefaultCost)
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
	if _, err = insertCliente(o, &cliente); err != nil {
		if isUniqueEmailErr(err) {
			c.Ctx.Output.SetStatus(http.StatusConflict)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusConflict,
				Message: "El correo ya está registrado",
				Cause:   err.Error(),
			}
			c.ServeJSON()
			return
		}
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
// @Param   body  body   models.ClienteUpdateRequest true  "Datos del cliente a actualizar (sólo campos a modificar)"
// @Success 200 {object} models.ApiResponse{data=models.Cliente} "Cliente actualizado"
// @Failure 401 {object} models.ApiResponse "No autorizado"
// @Failure 404 {object} models.ApiResponse "Cliente no encontrado"
// @Failure 409 {object} models.ApiResponse "Correo ya registrado"
// @Security BearerAuth
// @Router /clientes [put]
func (c *ClienteController) Put() {
	o := ormNew()

	// Obtener el ID del query parameter
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

	// Verificar si el cliente existe
	cliente := models.Cliente{PK_DOCUMENTO_CLIENTE: id}
	if err := readCliente(o, &cliente); err != nil {
		if err == orm.ErrNoRows {
			c.Ctx.Output.SetStatus(http.StatusOK)
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

	// Normalizar/Conservar correo
	correoTrim := strings.TrimSpace(updatedCliente.CORREO)
	if correoTrim == "" {
		updatedCliente.CORREO = cliente.CORREO // no sobreescribir con vacío
	} else {
		updatedCliente.CORREO = normalizeEmail(correoTrim)
	}

	// Si se proporciona una nueva contraseña, hashéala; de lo contrario conserva la actual
	if updatedCliente.PASSWORD != "" {
		hashedPassword, err := bcryptGenerate([]byte(updatedCliente.PASSWORD), bcrypt.DefaultCost)
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
	if _, err = updateCliente(o, &updatedCliente); err != nil {
		if isUniqueEmailErr(err) {
			c.Ctx.Output.SetStatus(http.StatusConflict)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusConflict,
				Message: "El correo ya está registrado",
				Cause:   err.Error(),
			}
			c.ServeJSON()
			return
		}
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
// @Failure 401 {object} models.ApiResponse "No autorizado"
// @Failure 404 {object} models.ApiResponse "Cliente no encontrado"
// @Security BearerAuth
// @Router /clientes [delete]
func (c *ClienteController) Delete() {
	o := ormNew()

	// Obtener el ID del query parameter
	id, err := c.GetInt64("id")
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

	if _, err := deleteCliente(o, &cliente); err == nil {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusOK,
			Message: "Cliente eliminado",
		}
		c.ServeJSON()
	} else {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Cliente no encontrado",
			Cause:   err.Error(),
		}
		c.ServeJSON()
	}
}
