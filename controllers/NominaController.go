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

type NominaController struct {
	web.Controller
}

// Estados permitidos para la nómina
var estadosNominaPermitidos = map[string]bool{
	"PAGO":    true,
	"NO PAGO": true,
}

// @Title GetAll
// @Summary Obtener todas las nóminas con filtros
// @Description Devuelve todas las nóminas registradas en la base de datos, con opción de filtrar por fecha exacta, mes y año.
// @Tags nominas
// @Accept json
// @Produce json
// @Param   fecha    query   string   false   "Filtrar por fecha exacta (YYYY-MM-DD)"
// @Param   mes      query   int      false   "Filtrar por mes (1-12)"
// @Param   anio     query   int      false   "Filtrar por año (YYYY)"
// @Success 200 {array} models.Nomina "Lista de nóminas"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /nominas [get]
func (c *NominaController) GetAll() {
	o := orm.NewOrm()
	var nominas []models.Nomina

	// Traer todos los registros de la base de datos
	_, err := o.QueryTable(new(models.Nomina)).All(&nominas)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener nóminas de la base de datos",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Leer parámetros de la URL
	fecha := c.GetString("fecha")
	mes, _ := c.GetInt("mes")
	anio, _ := c.GetInt("anio")

	// Aplicar filtros en memoria
	var filteredNominas []models.Nomina
	for _, nomina := range nominas {
		if fecha != "" && nomina.FECHA.Format("2006-01-02") != fecha {
			continue
		}
		if mes > 0 && mes <= 12 && int(nomina.FECHA.Month()) != mes {
			continue
		}
		if anio > 0 && nomina.FECHA.Year() != anio {
			continue
		}
		filteredNominas = append(filteredNominas, nomina)
	}

	// Si no hay resultados
	if len(filteredNominas) == 0 {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "No se encontraron nóminas que coincidan con los filtros proporcionados",
		}
		c.ServeJSON()
		return
	}

	// Responder con las nóminas filtradas
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Nóminas obtenidas exitosamente",
		Data:    filteredNominas,
	}
	c.ServeJSON()
}

// @Title GetById
// @Summary Obtener nómina por ID
// @Description Devuelve una nómina específica por ID utilizando query parameters.
// @Tags nominas
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID de la Nómina"
// @Success 200 {object} models.Nomina "Nómina encontrada"
// @Failure 404 {object} models.ApiResponse "Nómina no encontrada"
// @Security BearerAuth
// @Router /nominas/search [get]
func (c *NominaController) GetById() {
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

	// Convertir el id a int64
	nomina := models.Nomina{PK_ID_NOMINA: int64(id)}

	err = o.Read(&nomina)
	if err == orm.ErrNoRows {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Nómina no encontrada",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Nómina encontrada",
		Data:    nomina,
	}
	c.ServeJSON()
}

// @Title Create
// @Summary Crear una nueva nómina
// @Description Crea una nueva nómina en la base de datos.
// @Tags nominas
// @Accept json
// @Produce json
// @Param   body  body   models.Nomina true  "Datos de la nómina a crear"
// @Success 201 {object} models.Nomina "Nómina creada"
// @Failure 400 {object} models.ApiResponse "Error en la solicitud"
// @Security BearerAuth
// @Router /nominas [post]
func (c *NominaController) Post() {
	o := orm.NewOrm()
	var input map[string]interface{}
	var nomina models.Nomina

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

	// Validar y procesar FECHA
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
		nomina.FECHA = parsedDate
	} else {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El campo 'FECHA' no puede estar vacío",
		}
		c.ServeJSON()
		return
	}

	// Validar y procesar MONTO
	if monto, ok := input["MONTO"].(float64); ok {
		nomina.MONTO = int64(monto)
	} else {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El campo 'MONTO' es obligatorio y debe ser un número",
		}
		c.ServeJSON()
		return
	}

	// Validar y procesar ESTADO_NOMINA
	if estado, ok := input["ESTADO_NOMINA"].(string); ok {
		if !estadosNominaPermitidos[estado] {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusBadRequest,
				Message: "El estado de la nómina debe ser 'PAGO' o 'NO PAGO'",
			}
			c.ServeJSON()
			return
		}
		nomina.ESTADO_NOMINA = estado
	} else {
		nomina.ESTADO_NOMINA = "NO PAGO" // Valor por defecto
	}

	// Insertar en la base de datos
	_, err := o.Insert(&nomina)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al crear la nómina",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Responder con éxito
	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusCreated,
		Message: "Nómina creada correctamente",
		Data:    nomina,
	}
	c.ServeJSON()
}

// @Title Update
// @Summary Actualizar una nómina
// @Description Actualiza los datos de una nómina existente.
// @Tags nominas
// @Accept json
// @Produce json
// @Param   id    query    int  true   "ID de la Nómina"
// @Param   body  body   models.Nomina true  "Datos de la nómina a actualizar"
// @Success 200 {object} models.Nomina "Nómina actualizada"
// @Failure 404 {object} models.ApiResponse "Nómina no encontrada"
// @Security BearerAuth
// @Router /nominas [put]
func (c *NominaController) Put() {
	o := orm.NewOrm()

	// Obtener el ID
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

	// Buscar la nómina
	nomina := models.Nomina{PK_ID_NOMINA: int64(id)}
	if err := o.Read(&nomina); err == orm.ErrNoRows {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Nómina no encontrada",
		}
		c.ServeJSON()
		return
	}

	// Actualizar los datos
	var updatedNomina models.Nomina
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &updatedNomina); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Error al procesar la solicitud",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	if updatedNomina.ESTADO_NOMINA != "" && !estadosNominaPermitidos[updatedNomina.ESTADO_NOMINA] {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El estado de la nómina debe ser 'PAGO' o 'NO PAGO'",
		}
		c.ServeJSON()
		return
	}

	// Aplicar cambios
	nomina.FECHA = updatedNomina.FECHA
	nomina.MONTO = updatedNomina.MONTO
	nomina.ESTADO_NOMINA = updatedNomina.ESTADO_NOMINA

	if _, err := o.Update(&nomina); err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al actualizar la nómina",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Nómina actualizada correctamente",
		Data:    nomina,
	}
	c.ServeJSON()
}

// @Title Delete
// @Summary Eliminar una nómina (lógica)
// @Description Marca una nómina como "NO PAGO" en lugar de eliminarla físicamente.
// @Tags nominas
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID de la Nómina"
// @Success 200 {object} models.ApiResponse "Nómina eliminada lógicamente"
// @Failure 404 {object} models.ApiResponse "Nómina no encontrada"
// @Security BearerAuth
// @Router /nominas [delete]
func (c *NominaController) Delete() {
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

	nomina := models.Nomina{PK_ID_NOMINA: int64(id)}
	if err := o.Read(&nomina); err != nil {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Nómina no encontrada",
		}
		c.ServeJSON()
		return
	}

	nomina.ESTADO_NOMINA = "NO PAGO"
	if _, err := o.Update(&nomina, "ESTADO_NOMINA"); err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al eliminar lógicamente la nómina",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Nómina eliminada lógicamente",
	}
	c.ServeJSON()
}
