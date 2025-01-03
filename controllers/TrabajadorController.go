package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"restaurante/database"
	"restaurante/models"
	"strconv"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
	"golang.org/x/crypto/bcrypt"
)

type TrabajadorController struct {
	web.Controller
}

// Función para hash de contraseñas
func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// Validar fechas relacionadas con el trabajador
func validateDates(fechaIngreso, fechaRetiro *time.Time) error {
	if fechaIngreso != nil && fechaRetiro != nil {
		if fechaRetiro.Before(*fechaIngreso) {
			return fmt.Errorf("la fecha de retiro no puede ser anterior a la fecha de ingreso")
		}
	}
	return nil
}

// @Title GetAll
// @Summary Obtener todos los trabajadores con filtros
// @Description Devuelve todos los trabajadores registrados en la base de datos, con opción de filtrar por fecha de ingreso, rol, estado de retiro, o solo retirados.
// @Tags trabajadores
// @Accept json
// @Produce json
// @Param   fecha_ingreso    query   string   false   "Filtrar por fecha exacta de ingreso (YYYY-MM-DD)"
// @Param   rol              query   string   false   "Filtrar por rol del trabajador"
// @Param   incluir_retirados query  bool     false   "Incluir trabajadores retirados (true/false)"
// @Param   solo_retirados    query  bool     false   "Ver solo trabajadores retirados (true/false)"
// @Success 200 {array} models.Trabajador "Lista de trabajadores"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /trabajadores [get]
func (c *TrabajadorController) GetAll() {
	o := orm.NewOrm()
	var trabajadores []models.Trabajador

	// Leer parámetros de la URL
	fechaIngreso := c.GetString("fecha_ingreso")
	rol := c.GetString("rol")
	incluirRetirados, _ := c.GetBool("incluir_retirados", false) // Por defecto, no incluir retirados
	soloRetirados, _ := c.GetBool("solo_retirados", false)       // Por defecto, no mostrar solo retirados

	// Construir consulta inicial
	query := o.QueryTable(new(models.Trabajador))

	if soloRetirados {
		// Solo mostrar trabajadores con FECHA_RETIRO no nula
		query = query.Filter("FECHA_RETIRO__isnull", false)
	} else if !incluirRetirados {
		// Excluir trabajadores retirados si no se solicita incluirlos
		query = query.Filter("FECHA_RETIRO__isnull", true)
	}

	// Ejecutar la consulta
	_, err := query.All(&trabajadores)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener trabajadores de la base de datos",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Ajustar fechas a zona horaria Bogotá y excluir contraseñas
	for i := range trabajadores {
		trabajadores[i].FECHA_INGRESO = trabajadores[i].FECHA_INGRESO.In(database.BogotaZone)

		if trabajadores[i].FECHA_RETIRO != nil {
			fechaRetiro := trabajadores[i].FECHA_RETIRO.In(database.BogotaZone)
			trabajadores[i].FECHA_RETIRO = &fechaRetiro
		}

		if trabajadores[i].FECHA_NACIMIENTO != nil {
			fechaNacimiento := trabajadores[i].FECHA_NACIMIENTO.In(database.BogotaZone)
			trabajadores[i].FECHA_NACIMIENTO = &fechaNacimiento
		}

		trabajadores[i].PASSWORD = "" // Excluir contraseña
	}

	// Filtrar trabajadores según los parámetros proporcionados
	var filteredTrabajadores []models.Trabajador
	for _, trabajador := range trabajadores {
		if fechaIngreso != "" && trabajador.FECHA_INGRESO.Format("2006-01-02") != fechaIngreso {
			continue
		}
		if rol != "" && trabajador.ROL != rol {
			continue
		}
		filteredTrabajadores = append(filteredTrabajadores, trabajador)
	}

	// Si no hay resultados
	if len(filteredTrabajadores) == 0 {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "No se encontraron trabajadores que coincidan con los filtros proporcionados",
		}
		c.ServeJSON()
		return
	}

	// Respuesta con los trabajadores filtrados
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Trabajadores obtenidos exitosamente",
		Data:    filteredTrabajadores,
	}
	c.ServeJSON()
}

// @Title GetById
// @Summary Obtener trabajador por ID
// @Description Devuelve un trabajador específico por ID utilizando query parameters.
// @Tags trabajadores
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID del Trabajador"
// @Success 200 {object} models.Trabajador "Trabajador encontrado"
// @Failure 404 {object} models.ApiResponse "Trabajador no encontrado"
// @Security BearerAuth
// @Router /trabajadores/search [get]
func (c *TrabajadorController) GetById() {
	o := orm.NewOrm()
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

	trabajador := models.Trabajador{PK_DOCUMENTO_TRABAJADOR: id}
	err = o.Read(&trabajador)
	if err == orm.ErrNoRows {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Trabajador no encontrado",
		}
		c.ServeJSON()
		return
	}

	trabajador.PASSWORD = ""
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Trabajador encontrado",
		Data:    trabajador,
	}
	c.ServeJSON()
}

// @Title Create
// @Summary Crear un nuevo trabajador
// @Description Crea un nuevo trabajador en la base de datos.
// @Tags trabajadores
// @Accept json
// @Produce json
// @Param   body  body   models.Trabajador true  "Datos del trabajador a crear"
// @Success 201 {object} models.Trabajador "Trabajador creado"
// @Failure 400 {object} models.ApiResponse "Error en la solicitud"
// @Security BearerAuth
// @Router /trabajadores [post]
func (c *TrabajadorController) Post() {
	o := orm.NewOrm()
	var input map[string]interface{}

	// Decodificar la solicitud
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &input); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Error al decodificar la solicitud",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Crear instancia del modelo Trabajador
	var trabajador models.Trabajador

	// Procesar FECHA_INGRESO
	if fechaIngresoStr, ok := input["FECHA_INGRESO"].(string); ok && fechaIngresoStr != "" {
		parsedDate, err := time.Parse("2006-01-02", fechaIngresoStr)
		if err != nil {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusBadRequest,
				Message: "Formato de fecha inválido para FECHA_INGRESO",
				Cause:   err.Error(),
			}
			c.ServeJSON()
			return
		}
		trabajador.FECHA_INGRESO = parsedDate
	} else {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El campo FECHA_INGRESO es obligatorio",
		}
		c.ServeJSON()
		return
	}

	// Otros campos...

	// Insertar en la base de datos
	_, err := o.Insert(&trabajador)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al crear el trabajador",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Excluir contraseña de la respuesta
	trabajador.PASSWORD = ""

	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusCreated,
		Message: "Trabajador creado correctamente",
		Data:    trabajador,
	}
	c.ServeJSON()
}

// @Title Update
// @Summary Actualizar un trabajador
// @Description Actualiza los datos de un trabajador existente.
// @Tags trabajadores
// @Accept json
// @Produce json
// @Param   id    query    int  true   "ID del Trabajador"
// @Param   body  body   models.Trabajador true  "Datos del trabajador a actualizar"
// @Success 200 {object} models.Trabajador "Trabajador actualizado"
// @Failure 404 {object} models.ApiResponse "Trabajador no encontrado"
// @Security BearerAuth
// @Router /trabajadores [put]
func (c *TrabajadorController) Put() {
	o := orm.NewOrm()
	id, err := c.GetInt64("id")

	if err != nil || id == 0 {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "ID inválido o ausente",
		}
		c.ServeJSON()
		return
	}

	// Buscar trabajador existente
	trabajador := models.Trabajador{PK_DOCUMENTO_TRABAJADOR: id}
	if err := o.Read(&trabajador); err != nil {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Trabajador no encontrado",
		}
		c.ServeJSON()
		return
	}

	// Decodificar el cuerpo de la solicitud
	var input map[string]interface{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &input); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Error al decodificar los datos",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Actualizar campos proporcionados
	if nombre, ok := input["NOMBRE"].(string); ok && nombre != "" {
		trabajador.NOMBRE = nombre
	}

	if apellido, ok := input["APELLIDO"].(string); ok && apellido != "" {
		trabajador.APELLIDO = apellido
	}

	if rol, ok := input["ROL"].(string); ok && rol != "" {
		trabajador.ROL = rol
	}

	if sueldo, ok := input["SUELDO"].(float64); ok {
		trabajador.SUELDO = int64(sueldo)
	}

	if nuevo, ok := input["NUEVO"].(bool); ok {
		trabajador.NUEVO = nuevo
	}

	if telefono, ok := input["TELEFONO"].(string); ok && telefono != "" {
		trabajador.TELEFONO = &telefono
	}

	if horario, ok := input["HORARIO"].(string); ok && horario != "" {
		trabajador.HORARIO = &horario
	}

	if fechaIngresoStr, ok := input["FECHA_INGRESO"].(string); ok && fechaIngresoStr != "" {
		parsedDate, err := time.Parse("2006-01-02", fechaIngresoStr)
		if err != nil {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusBadRequest,
				Message: "Formato de fecha inválido para FECHA_INGRESO",
				Cause:   err.Error(),
			}
			c.ServeJSON()
			return
		}
		trabajador.FECHA_INGRESO = parsedDate
	}

	if fechaRetiroStr, ok := input["FECHA_RETIRO"].(string); ok && fechaRetiroStr != "" {
		parsedDate, err := time.Parse("2006-01-02", fechaRetiroStr)
		if err != nil {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusBadRequest,
				Message: "Formato de fecha inválido para FECHA_RETIRO",
				Cause:   err.Error(),
			}
			c.ServeJSON()
			return
		}
		fechaRetiro := parsedDate
		trabajador.FECHA_RETIRO = &fechaRetiro
	}

	if fechaNacimientoStr, ok := input["FECHA_NACIMIENTO"].(string); ok && fechaNacimientoStr != "" {
		parsedDate, err := time.Parse("2006-01-02", fechaNacimientoStr)
		if err != nil {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusBadRequest,
				Message: "Formato de fecha inválido para FECHA_NACIMIENTO",
				Cause:   err.Error(),
			}
			c.ServeJSON()
			return
		}
		fechaNacimiento := parsedDate
		trabajador.FECHA_NACIMIENTO = &fechaNacimiento
	}

	if password, ok := input["PASSWORD"].(string); ok && password != "" {
		hashedPassword, err := hashPassword(password)
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
		trabajador.PASSWORD = hashedPassword
	}

	// Validar fechas (FECHA_INGRESO y FECHA_RETIRO)
	if err := validateDates(&trabajador.FECHA_INGRESO, trabajador.FECHA_RETIRO); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Actualizar en la base de datos
	if _, err := o.Update(&trabajador); err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al actualizar el trabajador",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Excluir contraseña de la respuesta
	trabajador.PASSWORD = ""

	// Responder con éxito
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Trabajador actualizado correctamente",
		Data:    trabajador,
	}
	c.ServeJSON()
}

// @Title Delete
// @Summary Eliminar un trabajador
// @Description Elimina un trabajador de la base de datos.
// @Tags trabajadores
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID del Trabajador"
// @Success 200 {object} models.ApiResponse "Trabajador eliminado"
// @Failure 404 {object} models.ApiResponse "Trabajador no encontrado"
// @Security BearerAuth
// @Router /trabajadores [delete]
func (c *TrabajadorController) Delete() {
	o := orm.NewOrm()

	// Obtener el ID del query parameter
	idStr := c.GetString("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id == 0 {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El parámetro 'id' es inválido o está ausente",
			Cause:   "Se requiere un ID numérico válido en el parámetro 'id'",
		}
		c.ServeJSON()
		return
	}

	// Buscar al trabajador
	trabajador := models.Trabajador{PK_DOCUMENTO_TRABAJADOR: int64(id)}
	if err := o.Read(&trabajador); err != nil {
		if err == orm.ErrNoRows {
			c.Ctx.Output.SetStatus(http.StatusNotFound)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusNotFound,
				Message: "Trabajador no encontrado",
			}
		} else {
			c.Ctx.Output.SetStatus(http.StatusInternalServerError)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusInternalServerError,
				Message: "Error al buscar el trabajador",
				Cause:   err.Error(),
			}
		}
		c.ServeJSON()
		return
	}

	// Actualizar la fecha de retiro a la fecha actual en zona horaria de Bogotá
	fechaRetiro := time.Now().In(database.BogotaZone)
	trabajador.FECHA_RETIRO = &fechaRetiro

	// Actualizar el registro en la base de datos
	if _, err := o.Update(&trabajador, "FECHA_RETIRO"); err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al actualizar la fecha de retiro del trabajador",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Responder con éxito
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Fecha de retiro del trabajador actualizada correctamente",
		Data:    trabajador,
	}
	c.ServeJSON()
}
