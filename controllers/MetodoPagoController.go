package controllers

import (
	"encoding/json"
	"net/http"
	"restaurante/models"
	"strconv"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

type MetodoPagoController struct {
	web.Controller
}

// @Title GetAll
// @Summary Obtener todos los métodos de pago
// @Description Devuelve todos los métodos de pago registrados en la base de datos.
// @Tags metodos_pago
// @Accept json
// @Produce json
// @Success 200 {array} models.MetodoPago "Lista de métodos de pago"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Router /metodos_pago [get]
func (c *MetodoPagoController) GetAll() {
	o := orm.NewOrm()
	var metodos []models.MetodoPago

	_, err := o.QueryTable(new(models.MetodoPago)).All(&metodos)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener métodos de pago de la base de datos",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Métodos de pago obtenidos exitosamente",
		Data:    metodos,
	}
	c.ServeJSON()
}

// @Title GetById
// @Summary Obtener método de pago por ID
// @Description Devuelve un método de pago específico por ID utilizando query parameters.
// @Tags metodos_pago
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID del Método de Pago"
// @Success 200 {object} models.MetodoPago "Método de pago encontrado"
// @Failure 404 {object} models.ApiResponse "Método de pago no encontrado"
// @Router /metodos_pago/search [get]
func (c *MetodoPagoController) GetById() {
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

	metodo := models.MetodoPago{PK_ID_METODO_PAGO: id}

	err = o.Read(&metodo)
	if err == orm.ErrNoRows {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Método de pago no encontrado",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Método de pago encontrado",
		Data:    metodo,
	}
	c.ServeJSON()
}

// @Title Create
// @Summary Crear un nuevo método de pago
// @Description Crea un nuevo método de pago en la base de datos.
// @Tags metodos_pago
// @Accept json
// @Produce json
// @Param   body  body   models.MetodoPago true  "Datos del método de pago a crear"
// @Success 201 {object} models.MetodoPago "Método de pago creado"
// @Failure 400 {object} models.ApiResponse "Error en la solicitud"
// @Router /metodos_pago [post]
func (c *MetodoPagoController) Post() {
	o := orm.NewOrm()
	var metodo models.MetodoPago

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &metodo); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Error en la solicitud",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	_, err := o.Insert(&metodo)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al crear el método de pago",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusCreated,
		Message: "Método de pago creado correctamente",
		Data:    metodo,
	}
	c.ServeJSON()
}

// @Title Update
// @Summary Actualizar un método de pago
// @Description Actualiza los datos de un método de pago existente.
// @Tags metodos_pago
// @Accept json
// @Produce json
// @Param   id    query    int  true   "ID del Método de Pago"
// @Param   body  body   models.MetodoPago true  "Datos del método de pago a actualizar"
// @Success 200 {object} models.MetodoPago "Método de pago actualizado"
// @Failure 404 {object} models.ApiResponse "Método de pago no encontrado"
// @Router /metodos_pago [put]
func (c *MetodoPagoController) Put() {
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

	metodo := models.MetodoPago{PK_ID_METODO_PAGO: id}

	if o.Read(&metodo) == nil {
		var updatedMetodo models.MetodoPago
		if err := json.Unmarshal(c.Ctx.Input.RequestBody, &updatedMetodo); err != nil {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusBadRequest,
				Message: "Error en la solicitud",
				Cause:   err.Error(),
			}
			c.ServeJSON()
			return
		}

		updatedMetodo.PK_ID_METODO_PAGO = id
		_, err := o.Update(&updatedMetodo)
		if err != nil {
			c.Ctx.Output.SetStatus(http.StatusInternalServerError)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusInternalServerError,
				Message: "Error al actualizar el método de pago",
				Cause:   err.Error(),
			}
			c.ServeJSON()
			return
		}

		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusOK,
			Message: "Método de pago actualizado",
			Data:    updatedMetodo,
		}
		c.ServeJSON()
	} else {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Método de pago no encontrado",
			Cause:   err.Error(),
		}
		c.ServeJSON()
	}
}

// @Title Delete
// @Summary Eliminar un método de pago
// @Description Elimina un método de pago de la base de datos.
// @Tags metodos_pago
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID del Método de Pago"
// @Success 200 {object} models.ApiResponse "Método de pago eliminado"
// @Failure 404 {object} models.ApiResponse "Método de pago no encontrado"
// @Router /metodos_pago [delete]
func (c *MetodoPagoController) Delete() {
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

	metodo := models.MetodoPago{PK_ID_METODO_PAGO: id}

	if _, err := o.Delete(&metodo); err == nil {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusOK,
			Message: "Método de pago eliminado",
		}
		c.ServeJSON()
	} else {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Método de pago no encontrado",
			Cause:   err.Error(),
		}
		c.ServeJSON()
	}
}
