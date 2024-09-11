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
// @Description Obtener todos los clientes
// @Success 200 {object} []Cliente
// @Failure 500 Error en la base de datos
// @router /clientes [get]
func (c *ClienteController) GetAll() {
	var clientes []models.Cliente

	// Consulta para obtener todos los clientes
	query := `SELECT "PK_DOCUMENTO_CLIENTE", "NOMBRE", "APELLIDO", "DIRECCION", "TELEFONO", "OBSERVACIONES", "PASSWORD" FROM "CLIENTE"`

	// Ejecutamos la consulta
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

	// Iteramos sobre los resultados
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

	// Si no hubo clientes, devolvemos un código 404
	if len(clientes) == 0 {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "No se encontraron clientes",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Si todo sale bien, devolvemos la lista de clientes
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Clientes obtenidos exitosamente",
		Data:    clientes,
	}
	c.ServeJSON()
}

// @Title GetById
// @Description Obtener cliente por ID
// @Param   id     path    int     true        "ID del Cliente"
// @Success 200 {object} models.Cliente
// @Failure 404 Cliente no encontrado
// @router /clientes/:id [get]
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
// @Description Crear un nuevo cliente
// @Success 201 {object} models.Cliente
// @Failure 400 Error en la solicitud
// @router /clientes [post]
func (c *ClienteController) Post() {
	var cliente models.Cliente
	// Mostrar el cuerpo de la solicitud recibido
	body := c.Ctx.Input.RequestBody
	fmt.Println("Cuerpo recibido:", string(body)) // Esto imprimirá el cuerpo recibido

	if len(body) == 0 {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El cuerpo de la solicitud está vacío",
		}
		c.ServeJSON()
		return
	}
	// Imprimir el cuerpo recibido para verificar
	fmt.Println("Cuerpo recibido:", string(body))

	// Intentar deserializar el cuerpo de la solicitud
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

	// Crear el cliente en la base de datos
	query := `INSERT INTO "CLIENTE" ("PK_DOCUMENTO_CLIENTE", "NOMBRE", "APELLIDO", "DIRECCION", "TELEFONO", "OBSERVACIONES", "PASSWORD")
			  VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err = database.DB.Exec(query, cliente.PK_DOCUMENTO_CLIENTE, cliente.NOMBRE, cliente.APELLIDO, cliente.DIRECCION, cliente.TELEFONO, cliente.OBSERVACIONES, cliente.PASSWORD)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error creating cliente",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Respuesta exitosa
	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusCreated,
		Message: "Cliente creado correctamente",
		Data:    cliente,
	}
	c.ServeJSON()
}

// @Title Update
// @Description Actualizar un cliente
// @Param   id     path    int     true        "ID del Cliente"
// @Success 200 {object} models.Cliente
// @Failure 404 Cliente no encontrado
// @router /clientes/:id [put]
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
// @Description Eliminar un cliente
// @Param   id     path    int     true        "ID del Cliente"
// @Success 204 {object} nil
// @Failure 404 Cliente no encontrado
// @router /clientes/:id [delete]
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
