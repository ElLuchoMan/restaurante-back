package controllers

import (
	"encoding/json"
	"net/http"
	"restaurante/models"
	"strconv"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

type DomicilioController struct {
	web.Controller
}

// @Title GetAll
// @Summary Obtener todos los domicilios con posibilidad de filtrar
// @Description Devuelve todos los domicilios registrados en la base de datos, con opción de filtrar por dirección, teléfono y actualizado por.
// @Tags domicilios
// @Accept json
// @Produce json
// @Param   direccion    query   string   false   "Filtrar por dirección"
// @Param   telefono     query   string   false   "Filtrar por teléfono"
// @Param   fecha     query   string   false   "Filtrar por fecha"
// @Param   updated_by   query   string   false   "Filtrar por usuario que realizó la última actualización"
// @Success 200 {array} models.Domicilio "Lista de domicilios"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /domicilios [get]
func (c *DomicilioController) GetAll() {
	o := orm.NewOrm()
	qs := o.QueryTable(new(models.Domicilio))

	// Leer parámetros de la URL
	direccion := c.GetString("direccion")
	telefono := c.GetString("telefono")
	updatedBy := c.GetString("updated_by")
	fecha := c.GetString("fecha")

	// Aplicar filtros opcionales
	if direccion != "" {
		qs = qs.Filter("DIRECCION__icontains", direccion) // Búsqueda parcial
	}
	if telefono != "" {
		qs = qs.Filter("TELEFONO", telefono)
	}
	if updatedBy != "" {
		qs = qs.Filter("UPDATED_BY__icontains", updatedBy)
	}
	if fecha != "" {
		qs = qs.Filter("FECHA", fecha)
	}

	var domicilios []models.Domicilio
	count, err := qs.All(&domicilios)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener domicilios de la base de datos",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Si no se encuentran resultados
	if count == 0 {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "No se encontraron domicilios que coincidan con los filtros proporcionados",
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Domicilios obtenidos exitosamente",
		Data:    domicilios,
	}
	c.ServeJSON()
}

// @Title GetById
// @Summary Obtener domicilio por ID
// @Description Devuelve un domicilio específico por ID utilizando query parameters.
// @Tags domicilios
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID del Domicilio"
// @Success 200 {object} models.Domicilio "Domicilio encontrado"
// @Failure 404 {object} models.ApiResponse "Domicilio no encontrado"
// @Security BearerAuth
// @Router /domicilios/search [get]
func (c *DomicilioController) GetById() {
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

	domicilio := models.Domicilio{PK_ID_DOMICILIO: id}

	err = o.Read(&domicilio)
	if err == orm.ErrNoRows {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Domicilio no encontrado",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Domicilio encontrado",
		Data:    domicilio,
	}
	c.ServeJSON()
}

// @Title Create
// @Summary Crear un nuevo domicilio
// @Description Crea un nuevo domicilio en la base de datos.
// @Tags domicilios
// @Accept json
// @Produce json
// @Param   body  body   models.Domicilio true  "Datos del domicilio a crear"
// @Success 201 {object} models.Domicilio "Domicilio creado"
// @Failure 400 {object} models.ApiResponse "Error en la solicitud"
// @Security BearerAuth
// @Router /domicilios [post]
func (c *DomicilioController) Post() {
	o := orm.NewOrm()
	var input map[string]interface{}
	var domicilio models.Domicilio

	// Decodificar la solicitud
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &input); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Error al procesar la solicitud",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Validar y establecer los campos obligatorios
	if direccion, ok := input["DIRECCION"].(string); ok && direccion != "" {
		domicilio.DIRECCION = direccion
	} else {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El campo 'DIRECCION' es obligatorio",
		}
		c.ServeJSON()
		return
	}
	// Validar fecha:
	if fechaStr, ok := input["FECHA"].(string); ok && fechaStr != "" {
		parsedDate, err := time.Parse("2006-01-02", fechaStr)
		if err != nil {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusBadRequest,
				Message: "Formato de fecha inválido",
				Cause:   err.Error(),
			}
			c.ServeJSON()
			return
		}
		domicilio.FECHA = parsedDate
	} else {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El campo FECHA no puede estar vacío",
		}
		c.ServeJSON()
		return
	}

	if telefono, ok := input["TELEFONO"].(string); ok && telefono != "" {
		domicilio.TELEFONO = telefono
	} else {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El campo 'TELEFONO' es obligatorio",
		}
		c.ServeJSON()
		return
	}

	// Procesar campos opcionales
	if estadoPago, ok := input["ESTADO_PAGO"].(string); ok {
		domicilio.ESTADO_PAGO = estadoPago
	}
	if entregado, ok := input["ENTREGADO"].(bool); ok {
		domicilio.ENTREGADO = entregado
	}
	if observaciones, ok := input["OBSERVACIONES"].(string); ok {
		domicilio.OBSERVACIONES = observaciones
	}
	if createdBy, ok := input["CREATED_BY"].(string); ok {
		domicilio.CREATED_BY = &createdBy
	}

	// Establecer valores automáticos
	domicilio.CREATED_AT = time.Now().UTC()
	domicilio.UPDATED_AT = time.Time{} // Inicializa vacío

	// Insertar en la base de datos
	_, err := o.Insert(&domicilio)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al crear el domicilio",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Responder con éxito
	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusCreated,
		Message: "Domicilio creado correctamente",
		Data:    domicilio,
	}
	c.ServeJSON()
}

// @Title Update
// @Summary Actualizar un domicilio
// @Description Actualiza los datos de un domicilio existente.
// @Tags domicilios
// @Accept json
// @Produce json
// @Param   id    query    int  true   "ID del Domicilio"
// @Param   body  body   models.Domicilio true  "Datos del domicilio a actualizar"
// @Success 200 {object} models.Domicilio "Domicilio actualizado"
// @Failure 404 {object} models.ApiResponse "Domicilio no encontrado"
// @Security BearerAuth
// @Router /domicilios [put]
func (c *DomicilioController) Put() {
	o := orm.NewOrm()

	// Obtener el ID del domicilio
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

	// Buscar el domicilio por ID
	domicilio := models.Domicilio{PK_ID_DOMICILIO: id}
	if err := o.Read(&domicilio); err == orm.ErrNoRows {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Domicilio no encontrado",
		}
		c.ServeJSON()
		return
	}

	// Deserializar datos actualizados
	var input map[string]interface{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &input); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Error al procesar la solicitud",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Actualizar campos
	if direccion, ok := input["DIRECCION"].(string); ok {
		domicilio.DIRECCION = direccion
	}
	if telefono, ok := input["TELEFONO"].(string); ok {
		domicilio.TELEFONO = telefono
	}
	if estadoPago, ok := input["ESTADO_PAGO"].(string); ok {
		domicilio.ESTADO_PAGO = estadoPago
	}
	if entregado, ok := input["ENTREGADO"].(bool); ok {
		domicilio.ENTREGADO = entregado
	}
	if updatedBy, ok := input["UPDATED_BY"].(string); ok {
		domicilio.UPDATED_BY = &updatedBy
	}

	// Actualizar la fecha de modificación
	domicilio.UPDATED_AT = time.Now().UTC()

	// Guardar cambios
	if _, err := o.Update(&domicilio); err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al actualizar el domicilio",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Responder con éxito
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Domicilio actualizado correctamente",
		Data:    domicilio,
	}
	c.ServeJSON()
}

// @Title Delete
// @Summary Eliminar un domicilio
// @Description Elimina un domicilio de la base de datos.
// @Tags domicilios
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID del Domicilio"
// @Success 204 {object} nil "Domicilio eliminado"
// @Failure 404 {object} models.ApiResponse "Domicilio no encontrado"
// @Security BearerAuth
// @Router /domicilios [delete]
func (c *DomicilioController) Delete() {
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

	domicilio := models.Domicilio{PK_ID_DOMICILIO: id}

	if _, err := o.Delete(&domicilio); err == nil {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusOK,
			Message: "Domicilio eliminado",
		}
		c.ServeJSON()
	} else {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Domicilio no encontrado",
			Cause:   err.Error(),
		}
		c.ServeJSON()
	}
}
