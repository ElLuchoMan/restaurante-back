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

// newTrabajadorOrm allows tests to replace orm.NewOrm.
var newTrabajadorOrm = orm.NewOrm

type TrabajadorController struct {
	web.Controller
}

// hashPassword allows tests to stub the hash function.
var hashPassword = func(password string) (string, error) {
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
// @Success 200 {object} models.ApiResponse{data=[]models.Trabajador} "Lista de trabajadores"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /trabajadores [get]
func (c *TrabajadorController) GetAll() {
	o := newTrabajadorOrm()
	var trabajadores []models.Trabajador

	// Leer parámetros de la URL
	fechaIngreso := c.GetString("fecha_ingreso")
	if fechaIngreso == "" {
		fechaIngreso = c.GetString("fechaIngreso")
	}
	rol := c.GetString("rol")
	incluirRetirados, _ := c.GetBool("incluir_retirados", false) // Por defecto, no incluir retirados
	soloRetirados, _ := c.GetBool("solo_retirados", false)       // Por defecto, no mostrar solo retirados

	// Priorizar "solo retirados" sobre "incluir retirados"
	query := o.QueryTable(new(models.Trabajador))
	if soloRetirados {
		query = query.Filter("FECHA_RETIRO__isnull", false)
	} else if !incluirRetirados {
		query = query.Filter("FECHA_RETIRO__isnull", true)
	}

	// Aplicar filtros adicionales
	if fechaIngreso != "" {
		parsed, err := time.Parse("2006-01-02", fechaIngreso)
		if err != nil {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Formato de fecha inválido para 'fecha_ingreso', use YYYY-MM-DD", Cause: err.Error()}
			_ = c.ServeJSON()
			return
		}
		query = query.Filter("FECHA_INGRESO", parsed)
	}
	if rol != "" {
		query = query.Filter("ROL__exact", rol)
	}

	// Ejecutar consulta
	_, err := query.All(&trabajadores)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener trabajadores de la base de datos",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	// Excluir contraseñas y ajustar fechas
	for i := range trabajadores {
		trabajadores[i].PASSWORD = "" // Excluir contraseña
		if trabajadores[i].FECHA_RETIRO != nil {
			fechaRetiroUTC := trabajadores[i].FECHA_RETIRO.In(database.BogotaZone)
			trabajadores[i].FECHA_RETIRO = &fechaRetiroUTC // UTC sin ajuste
		}
		if trabajadores[i].FECHA_NACIMIENTO != nil {
			fechaNacimientoUTC := trabajadores[i].FECHA_NACIMIENTO.In(database.BogotaZone)
			trabajadores[i].FECHA_NACIMIENTO = &fechaNacimientoUTC // UTC sin ajuste
		}
		trabajadores[i].FECHA_INGRESO = trabajadores[i].FECHA_INGRESO.In(database.BogotaZone)
		var horarios []models.HorarioTrabajador
		if _, err := o.QueryTable(new(models.HorarioTrabajador)).
			Filter("pk_documento_trabajador", trabajadores[i].PK_DOCUMENTO_TRABAJADOR).
			All(&horarios); err == nil {
			trabajadores[i].HORARIOS = horarios
		}
	}

	// Si no hay resultados
	if len(trabajadores) == 0 {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusOK,
			Message: "No se encontraron trabajadores que coincidan con los filtros proporcionados",
			Data:    trabajadores,
		}
		_ = c.ServeJSON()
		return
	}

	// Respuesta exitosa
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Trabajadores obtenidos exitosamente",
		Data:    trabajadores,
	}
	_ = c.ServeJSON()
}

// @Title GetById
// @Summary Obtener trabajador por ID
// @Description Devuelve un trabajador específico por ID utilizando query parameters.
// @Tags trabajadores
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID del Trabajador"
// @Success 200 {object} models.ApiResponse{data=models.Trabajador} "Trabajador encontrado"
// @Failure 404 {object} models.ApiResponse "Trabajador no encontrado"
// @Security BearerAuth
// @Router /trabajadores/search [get]
func (c *TrabajadorController) GetById() {
	o := newTrabajadorOrm()
	id, err := c.GetInt64("id")

	if err != nil || id == 0 {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El parámetro 'id' es inválido o está ausente",
		}
		_ = c.ServeJSON()
		return
	}

	trabajador := models.Trabajador{PK_DOCUMENTO_TRABAJADOR: id}
	err = o.Read(&trabajador)
	if err == orm.ErrNoRows {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Trabajador no encontrado",
		}
		_ = c.ServeJSON()
		return
	}

	trabajador.PASSWORD = ""
	var horarios []models.HorarioTrabajador
	if _, err := o.QueryTable(new(models.HorarioTrabajador)).
		Filter("pk_documento_trabajador", trabajador.PK_DOCUMENTO_TRABAJADOR).
		All(&horarios); err == nil {
		trabajador.HORARIOS = horarios
	}
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Trabajador encontrado",
		Data:    trabajador,
	}
	_ = c.ServeJSON()
}

// @Title Create
// @Summary Crear un nuevo trabajador
// @Description Crea un nuevo trabajador en la base de datos.
// @Tags trabajadores
// @Accept json
// @Produce json
// @Param   body  body   models.TrabajadorCreateRequest true  "Datos del trabajador a crear (fecha YYYY-MM-DD)"
// @Success 201 {object} models.ApiResponse{data=models.Trabajador} "Trabajador creado"
// @Failure 400 {object} models.ApiResponse "Error en la solicitud"
// @Security BearerAuth
// @Router /trabajadores [post]
func (c *TrabajadorController) Post() {
	o := newTrabajadorOrm()
	var input map[string]interface{}

	// Decodificar la solicitud
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &input); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Error al decodificar la solicitud",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}
	fmt.Println("Este es el input recibido:", input)
	// Crear instancia del modelo Trabajador
	var trabajador models.Trabajador

	// Procesar PK_DOCUMENTO_TRABAJADOR
	if doc, ok := input["documentoTrabajador"].(float64); ok {
		trabajador.PK_DOCUMENTO_TRABAJADOR = int64(doc)
	} else {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El campo 'documentoTrabajador' es obligatorio y debe ser un número válido",
		}
		_ = c.ServeJSON()
		return
	}

	// Procesar NOMBRE
	if nombre, ok := input["nombre"].(string); ok && nombre != "" {
		trabajador.NOMBRE = nombre
	} else {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El campo 'nombre' es obligatorio",
		}
		_ = c.ServeJSON()
		return
	}

	// Procesar APELLIDO
	if apellido, ok := input["apellido"].(string); ok && apellido != "" {
		trabajador.APELLIDO = apellido
	} else {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El campo 'apellido' es obligatorio",
		}
		_ = c.ServeJSON()
		return
	}

	// Procesar ROL (aceptar cualquier string en creación)
	if rol, ok := input["rol"].(string); ok && rol != "" {
		trabajador.ROL = models.RolTrabajador(rol)
	} else {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El campo 'rol' es obligatorio",
		}
		_ = c.ServeJSON()
		return
	}

	// Procesar FECHA_INGRESO
	if fechaIngresoStr, ok := input["fechaIngreso"].(string); ok && fechaIngresoStr != "" {
		parsedDate, err := time.Parse("2006-01-02", fechaIngresoStr)
		if err != nil {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusBadRequest,
				Message: "Formato de fecha inválido para 'fechaIngreso'",
				Cause:   err.Error(),
			}
			_ = c.ServeJSON()
			return
		}
		trabajador.FECHA_INGRESO = parsedDate
	} else {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El campo 'fechaIngreso' es obligatorio",
		}
		_ = c.ServeJSON()
		return
	}

	// Procesar SUELDO
	if sueldo, ok := input["sueldo"].(float64); ok {
		trabajador.SUELDO = int64(sueldo)
	} else {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El campo 'sueldo' es obligatorio",
		}
		_ = c.ServeJSON()
		return
	}

	// Procesar PASSWORD
	if password, ok := input["password"].(string); ok && password != "" {
		hashedPassword, err := hashPassword(password)
		if err != nil {
			c.Ctx.Output.SetStatus(http.StatusInternalServerError)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusInternalServerError,
				Message: "Error al procesar la contraseña",
				Cause:   err.Error(),
			}
			_ = c.ServeJSON()
			return
		}
		trabajador.PASSWORD = hashedPassword
	} else {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El campo 'password' es obligatorio",
		}
		_ = c.ServeJSON()
		return
	}

	// Procesar otros campos opcionales (ejemplo TELEFONO)
	if telefono, ok := input["telefono"].(string); ok {
		trabajador.TELEFONO = &telefono
	}

	// Procesar PK_ID_RESTAURANTE
	if pkRestaurante, ok := input["restauranteId"].(float64); ok {
		valor := int64(pkRestaurante)
		trabajador.PK_ID_RESTAURANTE = &models.Restaurante{PK_ID_RESTAURANTE: valor}
	}

	// Procesar FECHA_NACIMIENTO
	if fechaNacimientoStr, ok := input["fechaNacimiento"].(string); ok && fechaNacimientoStr != "" {
		parsedDate, err := time.Parse("2006-01-02", fechaNacimientoStr)
		if err != nil {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusBadRequest,
				Message: "Formato de fecha inválido para 'fechaNacimiento'",
				Cause:   err.Error(),
			}
			_ = c.ServeJSON()
			return
		}
		trabajador.FECHA_NACIMIENTO = &parsedDate
	}

	// Insertar en la base de datos
	_, err := o.Insert(&trabajador)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al crear el trabajador",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
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
	_ = c.ServeJSON()
}

// @Title Update
// @Summary Actualizar un trabajador
// @Description Actualiza los datos de un trabajador existente.
// @Tags trabajadores
// @Accept json
// @Produce json
// @Param   id    query    int  true   "ID del Trabajador"
// @Param   body  body   models.TrabajadorUpdateRequest true  "Datos del trabajador a actualizar (sólo campos a modificar)"
// @Success 200 {object} models.ApiResponse "Trabajador actualizado"
// @Failure 404 {object} models.ApiResponse "Trabajador no encontrado"
// @Security BearerAuth
// @Router /trabajadores [put]
func (c *TrabajadorController) Put() {
	o := newTrabajadorOrm()
	id, err := c.GetInt64("id")

	if err != nil || id == 0 {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "ID inválido o ausente",
		}
		_ = c.ServeJSON()
		return
	}

	// Buscar trabajador existente
	trabajador := models.Trabajador{PK_DOCUMENTO_TRABAJADOR: id}
	if err := o.Read(&trabajador); err != nil {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Trabajador no encontrado",
		}
		_ = c.ServeJSON()
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
		_ = c.ServeJSON()
		return
	}

	// helper to read case-insensitive fields
	getStr := func(keys ...string) (string, bool) {
		for _, k := range keys {
			if v, ok := input[k].(string); ok && v != "" {
				return v, true
			}
		}
		return "", false
	}
	getFloat := func(keys ...string) (float64, bool) {
		for _, k := range keys {
			if v, ok := input[k].(float64); ok {
				return v, true
			}
		}
		return 0, false
	}
	getBool := func(keys ...string) (bool, bool) {
		for _, k := range keys {
			if v, ok := input[k].(bool); ok {
				return v, true
			}
		}
		return false, false
	}

	if v, ok := getStr("NOMBRE", "nombre"); ok {
		trabajador.NOMBRE = v
	}
	if v, ok := getStr("APELLIDO", "apellido"); ok {
		trabajador.APELLIDO = v
	}
	if v, ok := getStr("ROL", "rol"); ok {
		trabajador.ROL = models.RolTrabajador(v)
	}
	if v, ok := getFloat("SUELDO", "sueldo"); ok {
		trabajador.SUELDO = int64(v)
	}
	if v, ok := getBool("NUEVO", "nuevo"); ok {
		trabajador.NUEVO = v
	}
	if v, ok := getStr("TELEFONO", "telefono"); ok {
		trabajador.TELEFONO = &v
	}

	if v, ok := getStr("FECHA_INGRESO", "fechaIngreso"); ok {
		parsedDate, err := time.Parse("2006-01-02", v)
		if err != nil {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Formato de fecha inválido para FECHA_INGRESO", Cause: err.Error()}
			_ = c.ServeJSON()
			return
		}
		trabajador.FECHA_INGRESO = parsedDate
	}
	if v, ok := getStr("FECHA_RETIRO", "fechaRetiro"); ok {
		parsedDate, err := time.Parse("2006-01-02", v)
		if err != nil {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Formato de fecha inválido para FECHA_RETIRO", Cause: err.Error()}
			_ = c.ServeJSON()
			return
		}
		fechaRetiro := parsedDate
		trabajador.FECHA_RETIRO = &fechaRetiro
	}
	if v, ok := getStr("FECHA_NACIMIENTO", "fechaNacimiento"); ok {
		parsedDate, err := time.Parse("2006-01-02", v)
		if err != nil {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Formato de fecha inválido para FECHA_NACIMIENTO", Cause: err.Error()}
			_ = c.ServeJSON()
			return
		}
		fechaNacimiento := parsedDate
		trabajador.FECHA_NACIMIENTO = &fechaNacimiento
	}
	if v, ok := getStr("PASSWORD", "password"); ok {
		hashedPassword, err := hashPassword(v)
		if err != nil {
			c.Ctx.Output.SetStatus(http.StatusInternalServerError)
			c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al procesar la contraseña", Cause: err.Error()}
			_ = c.ServeJSON()
			return
		}
		trabajador.PASSWORD = hashedPassword
	}

	// Validar fechas (FECHA_INGRESO y FECHA_RETIRO)
	if err := validateDates(&trabajador.FECHA_INGRESO, trabajador.FECHA_RETIRO); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: err.Error()}
		_ = c.ServeJSON()
		return
	}

	// Actualizar en la base de datos
	if _, err := o.Update(&trabajador); err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al actualizar el trabajador", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}

	// Excluir contraseña de la respuesta
	trabajador.PASSWORD = ""

	// Responder con éxito
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{Code: http.StatusOK, Message: "Trabajador actualizado correctamente", Data: trabajador}
	_ = c.ServeJSON()
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
	o := newTrabajadorOrm()

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
		_ = c.ServeJSON()
		return
	}

	// Buscar al trabajador
	trabajador := models.Trabajador{PK_DOCUMENTO_TRABAJADOR: int64(id)}
	if err := o.Read(&trabajador); err != nil {
		if err == orm.ErrNoRows {
			c.Ctx.Output.SetStatus(http.StatusOK)
			c.Data["json"] = models.ApiResponse{Code: http.StatusNotFound, Message: "Trabajador no encontrado"}
		} else {
			c.Ctx.Output.SetStatus(http.StatusInternalServerError)
			c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al buscar el trabajador", Cause: err.Error()}
		}
		_ = c.ServeJSON()
		return
	}

	// Actualizar la fecha de retiro a la fecha actual en zona horaria de Bogotá
	fechaRetiro := time.Now().In(database.BogotaZone)
	trabajador.FECHA_RETIRO = &fechaRetiro

	// Actualizar el registro en la base de datos
	if _, err := o.Update(&trabajador, "FECHA_RETIRO"); err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al actualizar la fecha de retiro del trabajador", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}
	trabajador.PASSWORD = ""
	// Responder con éxito
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{Code: http.StatusOK, Message: "Fecha de retiro del trabajador actualizada correctamente", Data: trabajador}
	_ = c.ServeJSON()
}
