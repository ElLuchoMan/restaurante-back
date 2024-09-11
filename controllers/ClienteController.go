package controllers

import (
	"encoding/json"
	"net/http"
	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

type ClienteController struct {
	web.Controller
}

// @Title GetAll
// @Summary Obtener todos los clientes
// @Description Devuelve todos los clientes registrados en la base de datos.
// @Tags clientes
// @Accept json
// @Produce json
// @Success 200 {array} models.Cliente "Lista de clientes"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Router /restaurante/v1/clientes [get]
func (c *ClienteController) GetAll() {
	o := orm.NewOrm()
	var clientes []models.Cliente

	// Consulta todos los clientes
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
// @Description Devuelve un cliente específico por ID.
// @Tags clientes
// @Accept json
// @Produce json
// @Param   id     path    int     true        "ID del Cliente"
// @Success 200 {object} models.Cliente "Cliente encontrado"
// @Failure 404 {object} models.ApiResponse "Cliente no encontrado"
// @Router /restaurante/v1/clientes/{id} [get]
func (c *ClienteController) GetById() {
	o := orm.NewOrm()
	id, _ := c.GetInt(":id")
	cliente := models.Cliente{PK_DOCUMENTO_CLIENTE: id}

	err := o.Read(&cliente)
	if err == orm.ErrNoRows {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Cliente no encontrado",
		}
		c.ServeJSON()
		return
	}

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
// @Router /restaurante/v1/clientes [post]
func (c *ClienteController) Post() {
	o := orm.NewOrm()
	var cliente models.Cliente

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &cliente); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Error en la solicitud",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	_, err := o.Insert(&cliente)
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
// @Param   id    path    int  true   "ID del Cliente"
// @Param   body  body   models.Cliente true  "Datos del cliente a actualizar"
// @Success 200 {object} models.Cliente "Cliente actualizado"
// @Failure 404 {object} models.ApiResponse "Cliente no encontrado"
// @Router /restaurante/v1/clientes/{id} [put]
func (c *ClienteController) Put() {
	o := orm.NewOrm()
	id, _ := c.GetInt(":id")
	cliente := models.Cliente{PK_DOCUMENTO_CLIENTE: id}

	if o.Read(&cliente) == nil {
		var updatedCliente models.Cliente
		if err := json.Unmarshal(c.Ctx.Input.RequestBody, &updatedCliente); err != nil {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusBadRequest,
				Message: "Error en la solicitud",
				Cause:   err.Error(),
			}
			c.ServeJSON()
			return
		}

		updatedCliente.PK_DOCUMENTO_CLIENTE = id
		_, err := o.Update(&updatedCliente)
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

		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusOK,
			Message: "Cliente actualizado",
			Data:    updatedCliente,
		}
		c.ServeJSON()
	} else {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Cliente no encontrado",
		}
		c.ServeJSON()
	}
}

// @Title Delete
// @Summary Eliminar un cliente
// @Description Elimina un cliente de la base de datos.
// @Tags clientes
// @Accept json
// @Produce json
// @Param   id     path    int     true        "ID del Cliente"
// @Success 204 {object} nil "Cliente eliminado"
// @Failure 404 {object} models.ApiResponse "Cliente no encontrado"
// @Router /restaurante/v1/clientes/{id} [delete]
func (c *ClienteController) Delete() {
	o := orm.NewOrm()
	id, _ := c.GetInt(":id")
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
		}
		c.ServeJSON()
	}
}
