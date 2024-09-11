package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"restaurante/database"
	"restaurante/models"

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
// @Router /clientes [get]
func (c *ClienteController) GetAll() {
	var clientes []models.Cliente
	query := `SELECT "PK_DOCUMENTO_CLIENTE", "NOMBRE", "APELLIDO", "DIRECCION", "TELEFONO", "OBSERVACIONES", "PASSWORD" FROM "CLIENTE"`
	rows, err := database.DB.Query(query)
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
	defer rows.Close()

	for rows.Next() {
		var cliente models.Cliente
		err := rows.Scan(&cliente.PK_DOCUMENTO_CLIENTE, &cliente.NOMBRE, &cliente.APELLIDO, &cliente.DIRECCION, &cliente.TELEFONO, &cliente.OBSERVACIONES, &cliente.PASSWORD)
		if err != nil {
			c.Ctx.Output.SetStatus(http.StatusInternalServerError)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusInternalServerError,
				Message: "Error al escanear los datos: ",
				Cause:   err.Error(),
			}
			c.ServeJSON()
			return
		}
		clientes = append(clientes, cliente)
	}

	if len(clientes) == 0 {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "No se encontraron clientes",
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
// @Router /clientes/{id} [get]
func (c *ClienteController) GetById() {
	id := c.Ctx.Input.Param(":id")
	query := `SELECT "PK_DOCUMENTO_CLIENTE", "NOMBRE", "APELLIDO", "DIRECCION", "TELEFONO", "OBSERVACIONES", "PASSWORD" FROM "CLIENTE" WHERE "PK_DOCUMENTO_CLIENTE" = $1`
	var cliente models.Cliente
	err := database.DB.QueryRow(query, id).Scan(&cliente.PK_DOCUMENTO_CLIENTE, &cliente.NOMBRE, &cliente.APELLIDO, &cliente.DIRECCION, &cliente.TELEFONO, &cliente.OBSERVACIONES, &cliente.PASSWORD)

	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Cliente no encontrado",
			Cause:   err.Error(),
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
// @Router /clientes [post]
func (c *ClienteController) Post() {
	var cliente models.Cliente
	body := c.Ctx.Input.RequestBody
	fmt.Println("Cuerpo recibido:", string(body))

	if len(body) == 0 {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El cuerpo de la solicitud está vacío",
		}
		c.ServeJSON()
		return
	}

	err := json.Unmarshal(body, &cliente)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Error parsing input data",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	query := `INSERT INTO "CLIENTE" ("PK_DOCUMENTO_CLIENTE", "NOMBRE", "APELLIDO", "DIRECCION", "TELEFONO", "OBSERVACIONES", "PASSWORD")
			  VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err = database.DB.Exec(query, cliente.PK_DOCUMENTO_CLIENTE, cliente.NOMBRE, cliente.APELLIDO, cliente.DIRECCION, cliente.TELEFONO, cliente.OBSERVACIONES, cliente.PASSWORD)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error creando cliente",
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
// @Router /clientes/{id} [put]
func (c *ClienteController) Put() {
	id := c.Ctx.Input.Param(":id")
	var cliente models.Cliente
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &cliente)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Error parsing input data",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	query := `UPDATE "CLIENTE" SET "NOMBRE"=$1, "APELLIDO"=$2, "DIRECCION"=$3, "TELEFONO"=$4, "OBSERVACIONES"=$5, "PASSWORD"=$6 WHERE "PK_DOCUMENTO_CLIENTE"=$7`
	_, err = database.DB.Exec(query, cliente.NOMBRE, cliente.APELLIDO, cliente.DIRECCION, cliente.TELEFONO, cliente.OBSERVACIONES, cliente.PASSWORD, id)
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
		Data:    cliente,
	}
	c.ServeJSON()
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
// @Router /clientes/{id} [delete]
func (c *ClienteController) Delete() {
	id := c.Ctx.Input.Param(":id")

	query := `DELETE FROM "CLIENTE" WHERE "PK_DOCUMENTO_CLIENTE"=$1`
	_, err := database.DB.Exec(query, id)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al eliminar el cliente",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Cliente eliminado",
	}
	c.ServeJSON()
}
