package cliente

import (
	"encoding/json"
	"net/http"
	"restaurante/logging"
	"restaurante/models"
	"strings"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
	"golang.org/x/crypto/bcrypt"
)

var ormNew func() orm.Ormer

var queryAllClientes func(o orm.Ormer, clientes *[]models.Cliente) (int64, error)

var readCliente func(o orm.Ormer, c *models.Cliente) error

var insertCliente func(o orm.Ormer, c *models.Cliente) (int64, error)

var updateCliente func(o orm.Ormer, c *models.Cliente) (int64, error)

var deleteCliente func(o orm.Ormer, c *models.Cliente) (int64, error)

var bcryptGenerate func([]byte, int) ([]byte, error)

func useDefaultClienteWrappers() {
	ormNew = orm.NewOrm
	queryAllClientes = func(o orm.Ormer, clientes *[]models.Cliente) (int64, error) {
		return o.QueryTable(new(models.Cliente)).All(clientes)
	}
	readCliente = func(o orm.Ormer, c *models.Cliente) error { return o.Read(c) }
	insertCliente = func(o orm.Ormer, c *models.Cliente) (int64, error) { return o.Insert(c) }
	updateCliente = func(o orm.Ormer, c *models.Cliente) (int64, error) { return o.Update(c) }
	deleteCliente = func(o orm.Ormer, c *models.Cliente) (int64, error) { return o.Delete(c) }
	bcryptGenerate = bcrypt.GenerateFromPassword
}

func init() {
	useDefaultClienteWrappers()
}

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

	fields := c.GetString("fields")

	limitStr := c.GetString("limit")
	offsetStr := c.GetString("offset")

	if limitStr != "" || offsetStr != "" {
		limit, errL := c.GetInt("limit")
		offset, errO := c.GetInt("offset")
		if limitStr != "" && errL != nil {
			logging.LogControllerError(c.Ctx, "clientes.getall.bad_request", errL, map[string]interface{}{"limit": limitStr})
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Parámetro 'limit' inválido", Cause: errL.Error()}
			_ = c.ServeJSON()
			return
		}
		if offsetStr != "" && errO != nil {
			logging.LogControllerError(c.Ctx, "clientes.getall.bad_request", errO, map[string]interface{}{"offset": offsetStr})
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Parámetro 'offset' inválido", Cause: errO.Error()}
			_ = c.ServeJSON()
			return
		}
		if limitStr != "" && limit == 0 {
			limit = 10
		}

		qs := o.QueryTable(new(models.Cliente))
		if _, err := qs.Limit(limit, offset).All(&clientes); err != nil {
			logging.LogControllerError(c.Ctx, "clientes.getall.db_error", err, nil)
			c.Ctx.Output.SetStatus(http.StatusInternalServerError)
			c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al obtener clientes", Cause: err.Error()}
			_ = c.ServeJSON()
			return
		}
	} else {
		if _, err := queryAllClientes(o, &clientes); err != nil {
			logging.LogControllerError(c.Ctx, "clientes.getall.db_error", err, nil)
			c.Ctx.Output.SetStatus(http.StatusInternalServerError)
			c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al obtener clientes", Cause: err.Error()}
			_ = c.ServeJSON()
			return
		}
	}

	if fields == "nombre_completo_telefono" {
		var reduced []map[string]interface{}
		for _, cli := range clientes {
			fullName := strings.TrimSpace(cli.NOMBRE + " " + cli.APELLIDO)
			reduced = append(reduced, map[string]interface{}{
				"documentoCliente": cli.PK_DOCUMENTO_CLIENTE,
				"nombre_completo":  fullName,
				"telefono":         cli.TELEFONO,
			})
		}
		c.Data["json"] = models.ApiResponse{Code: http.StatusOK, Message: "Clientes obtenidos", Data: reduced}
		_ = c.ServeJSON()
		return
	}

	for i := range clientes {
		clientes[i].PASSWORD = ""
	}
	c.Data["json"] = models.ApiResponse{Code: http.StatusOK, Message: "Clientes obtenidos", Data: clientes}
	_ = c.ServeJSON()
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
		logging.LogControllerError(c.Ctx, "clientes.getbyid.bad_request", err, map[string]interface{}{"id": c.GetString("id")})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El parámetro 'id' es inválido o está ausente",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
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
		_ = c.ServeJSON()
		return
	} else if err != nil {
		logging.LogControllerError(c.Ctx, "clientes.getbyid.db_error", err, map[string]interface{}{"id": id})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al consultar el cliente",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	cliente.PASSWORD = ""

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Cliente encontrado",
		Data:    cliente,
	}
	_ = c.ServeJSON()
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
		logging.LogControllerError(c.Ctx, "clientes.post.bad_json", err, nil)
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Error al decodificar la solicitud",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	correoTrim := strings.TrimSpace(cliente.CORREO)
	if correoTrim == "" {
		logging.LogControllerError(c.Ctx, "clientes.post.validation_error", nil, map[string]interface{}{"missing": "correo"})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El campo correo es obligatorio",
		}
		_ = c.ServeJSON()
		return
	}
	cliente.CORREO = normalizeEmail(correoTrim)

	hashedPassword, err := bcryptGenerate([]byte(cliente.PASSWORD), bcrypt.DefaultCost)
	if err != nil {
		logging.LogControllerError(c.Ctx, "clientes.post.hash_error", err, nil)
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al procesar la contraseña",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}
	cliente.PASSWORD = string(hashedPassword)

	if _, err = insertCliente(o, &cliente); err != nil {
		if isUniqueEmailErr(err) {
			logging.LogControllerError(c.Ctx, "clientes.post.unique_conflict", err, map[string]interface{}{"correo": cliente.CORREO})
			c.Ctx.Output.SetStatus(http.StatusConflict)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusConflict,
				Message: "El correo ya está registrado",
				Cause:   err.Error(),
			}
			_ = c.ServeJSON()
			return
		}
		logging.LogControllerError(c.Ctx, "clientes.post.insert_error", err, map[string]interface{}{"correo": cliente.CORREO})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al crear el cliente",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	cliente.PASSWORD = ""

	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusCreated,
		Message: "Cliente creado correctamente",
		Data:    cliente,
	}
	_ = c.ServeJSON()
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

	id, err := c.GetInt64("id")
	if err != nil || id == 0 {
		logging.LogControllerError(c.Ctx, "clientes.put.bad_request", err, map[string]interface{}{"id": c.GetString("id")})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El parámetro 'id' es inválido o está ausente",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	cliente := models.Cliente{PK_DOCUMENTO_CLIENTE: id}
	if err := readCliente(o, &cliente); err != nil {
		if err == orm.ErrNoRows {
			c.Ctx.Output.SetStatus(http.StatusOK)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusNotFound,
				Message: "Cliente no encontrado",
			}
			_ = c.ServeJSON()
		} else {
			logging.LogControllerError(c.Ctx, "clientes.put.db_error", err, map[string]interface{}{"id": id})
			c.Ctx.Output.SetStatus(http.StatusInternalServerError)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusInternalServerError,
				Message: "Error al buscar el cliente",
				Cause:   err.Error(),
			}
			_ = c.ServeJSON()
		}
		return
	}

	var updatedCliente models.Cliente
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &updatedCliente); err != nil {
		logging.LogControllerError(c.Ctx, "clientes.put.bad_json", err, map[string]interface{}{"id": id})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Error al decodificar la solicitud",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	updatedCliente.PK_DOCUMENTO_CLIENTE = cliente.PK_DOCUMENTO_CLIENTE

	correoTrim := strings.TrimSpace(updatedCliente.CORREO)
	if correoTrim == "" {
		updatedCliente.CORREO = cliente.CORREO
	} else {
		updatedCliente.CORREO = normalizeEmail(correoTrim)
	}

	if updatedCliente.PASSWORD != "" {
		hashedPassword, err := bcryptGenerate([]byte(updatedCliente.PASSWORD), bcrypt.DefaultCost)
		if err != nil {
			logging.LogControllerError(c.Ctx, "clientes.put.hash_error", err, map[string]interface{}{"id": id})
			c.Ctx.Output.SetStatus(http.StatusInternalServerError)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusInternalServerError,
				Message: "Error al procesar la contraseña",
				Cause:   err.Error(),
			}
			_ = c.ServeJSON()
			return
		}
		updatedCliente.PASSWORD = string(hashedPassword)
	} else {
		updatedCliente.PASSWORD = cliente.PASSWORD
	}

	if _, err = updateCliente(o, &updatedCliente); err != nil {
		if isUniqueEmailErr(err) {
			logging.LogControllerError(c.Ctx, "clientes.put.unique_conflict", err, map[string]interface{}{"id": id, "correo": updatedCliente.CORREO})
			c.Ctx.Output.SetStatus(http.StatusConflict)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusConflict,
				Message: "El correo ya está registrado",
				Cause:   err.Error(),
			}
			_ = c.ServeJSON()
			return
		}
		logging.LogControllerError(c.Ctx, "clientes.put.update_error", err, map[string]interface{}{"id": id})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al actualizar el cliente",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	updatedCliente.PASSWORD = ""

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Cliente actualizado",
		Data:    updatedCliente,
	}
	_ = c.ServeJSON()
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

	id, err := c.GetInt64("id")
	if err != nil || id == 0 {
		logging.LogControllerError(c.Ctx, "clientes.delete.bad_request", err, map[string]interface{}{"id": c.GetString("id")})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El parámetro 'id' es inválido o está ausente",
		}
		_ = c.ServeJSON()
		return
	}

	cliente := models.Cliente{PK_DOCUMENTO_CLIENTE: id}

	if _, err := deleteCliente(o, &cliente); err == nil {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusOK,
			Message: "Cliente eliminado",
		}
		_ = c.ServeJSON()
	} else {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Cliente no encontrado",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
	}
}
