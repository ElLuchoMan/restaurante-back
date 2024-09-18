package controllers

import (
	"encoding/json"
	"net/http"
	"restaurante/models"
	"strconv"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

type PagoController struct {
	web.Controller
}

// @Title GetAll
// @Summary Obtener todos los pagos
// @Description Devuelve todos los pagos registrados en la base de datos.
// @Tags pagos
// @Accept json
// @Produce json
// @Success 200 {array} models.Pago "Lista de pagos"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Router /pagos [get]
func (c *PagoController) GetAll() {
	o := orm.NewOrm()
	var pagos []models.Pago

	_, err := o.QueryTable(new(models.Pago)).All(&pagos)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener pagos de la base de datos",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Pagos obtenidos exitosamente",
		Data:    pagos,
	}
	c.ServeJSON()
}

// @Title GetById
// @Summary Obtener pago por ID
// @Description Devuelve un pago específico por ID utilizando query parameters.
// @Tags pagos
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID del Pago"
// @Success 200 {object} models.Pago "Pago encontrado"
// @Failure 404 {object} models.ApiResponse "Pago no encontrado"
// @Router /pagos/search [get]
func (c *PagoController) GetById() {
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

	pago := models.Pago{PK_ID_PAGO: id}

	err = o.Read(&pago)
	if err == orm.ErrNoRows {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Pago no encontrado",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Pago encontrado",
		Data:    pago,
	}
	c.ServeJSON()
}

// @Title Create
// @Summary Crear un nuevo pago
// @Description Crea un nuevo pago en la base de datos.
// @Tags pagos
// @Accept json
// @Produce json
// @Param   body  body   models.Pago true  "Datos del pago a crear"
// @Success 201 {object} models.Pago "Pago creado"
// @Failure 400 {object} models.ApiResponse "Error en la solicitud"
// @Router /pagos [post]
func (c *PagoController) Post() {
	o := orm.NewOrm()
	var pago models.Pago

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &pago); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Error en la solicitud",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	_, err := o.Insert(&pago)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al crear el pago",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusCreated,
		Message: "Pago creado correctamente",
		Data:    pago,
	}
	c.ServeJSON()
}

// @Title Update
// @Summary Actualizar un pago
// @Description Actualiza los datos de un pago existente.
// @Tags pagos
// @Accept json
// @Produce json
// @Param   id    query    int  true   "ID del Pago"
// @Param   body  body   models.Pago true  "Datos del pago a actualizar"
// @Success 200 {object} models.Pago "Pago actualizado"
// @Failure 404 {object} models.ApiResponse "Pago no encontrado"
// @Router /pagos [put]
func (c *PagoController) Put() {
	o := orm.NewOrm()

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

	pago := models.Pago{PK_ID_PAGO: id}

	if o.Read(&pago) == nil {
		var updatedPago models.Pago
		if err := json.Unmarshal(c.Ctx.Input.RequestBody, &updatedPago); err != nil {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusBadRequest,
				Message: "Error en la solicitud",
				Cause:   err.Error(),
			}
			c.ServeJSON()
			return
		}

		updatedPago.PK_ID_PAGO = id
		_, err := o.Update(&updatedPago)
		if err != nil {
			c.Ctx.Output.SetStatus(http.StatusInternalServerError)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusInternalServerError,
				Message: "Error al actualizar el pago",
				Cause:   err.Error(),
			}
			c.ServeJSON()
			return
		}

		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusOK,
			Message: "Pago actualizado",
			Data:    updatedPago,
		}
		c.ServeJSON()
	} else {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Pago no encontrado",
			Cause:   err.Error(),
		}
		c.ServeJSON()
	}
}

// @Title Delete
// @Summary Eliminar un pago
// @Description Elimina un pago de la base de datos.
// @Tags pagos
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID del Pago"
// @Success 200 {object} models.ApiResponse "Pago eliminado"
// @Failure 404 {object} models.ApiResponse "Pago no encontrado"
// @Router /pagos [delete]
func (c *PagoController) Delete() {
	o := orm.NewOrm()

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

	pago := models.Pago{PK_ID_PAGO: id}

	if _, err := o.Delete(&pago); err == nil {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusOK,
			Message: "Pago eliminado",
		}
		c.ServeJSON()
	} else {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Pago no encontrado",
			Cause:   err.Error(),
		}
		c.ServeJSON()
	}
}
