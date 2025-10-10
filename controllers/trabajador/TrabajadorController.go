package trabajador

import (
	"encoding/json"
	"fmt"
	"net/http"
	"restaurante/database"
	"restaurante/logging"
	"restaurante/models"
	"strconv"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
	"golang.org/x/crypto/bcrypt"
)

var newTrabajadorOrm = orm.NewOrm

type TrabajadorController struct {
	web.Controller
}

var generateFromPassword = bcrypt.GenerateFromPassword

var hashPassword = func(password string) (string, error) {
	hashedPassword, err := generateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

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

	fechaIngreso := c.GetString("fecha_ingreso")
	if fechaIngreso == "" {
		fechaIngreso = c.GetString("fechaIngreso")
	}
	rol := c.GetString("rol")
	incluirRetirados, _ := c.GetBool("incluir_retirados", false)
	soloRetirados, _ := c.GetBool("solo_retirados", false)

	query := o.QueryTable(new(models.Trabajador))
	if soloRetirados {
		query = query.Filter("FECHA_RETIRO__isnull", false)
	} else if !incluirRetirados {
		query = query.Filter("FECHA_RETIRO__isnull", true)
	}

	if fechaIngreso != "" {
		parsed, err := time.Parse("2006-01-02", fechaIngreso)
		if err != nil {
			logging.LogControllerError(c.Ctx, "trabajadores.getall.bad_fecha_ingreso", err, map[string]interface{}{"fecha_ingreso": fechaIngreso})
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

	_, err := query.All(&trabajadores)
	if err != nil {
		logging.LogControllerError(c.Ctx, "trabajadores.getall.db_error", err, map[string]interface{}{"rol": rol, "solo_retirados": soloRetirados, "incluir_retirados": incluirRetirados, "fecha_ingreso": fechaIngreso})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener trabajadores de la base de datos",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	for i := range trabajadores {
		trabajadores[i].PASSWORD = ""
		// Los datos de fecha ya están en zona horaria de Bogotá - no aplicar conversiones
		var horarios []models.HorarioTrabajador
		if _, err := o.QueryTable(new(models.HorarioTrabajador)).
			Filter("pk_documento_trabajador", trabajadores[i].PK_DOCUMENTO_TRABAJADOR).
			All(&horarios); err == nil {
			trabajadores[i].HORARIOS = horarios
		}
	}

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
		logging.LogControllerError(c.Ctx, "trabajadores.getbyid.bad_request", err, map[string]interface{}{"id": c.GetString("id")})
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
	} else if err != nil {
		logging.LogControllerError(c.Ctx, "trabajadores.getbyid.db_error", err, map[string]interface{}{"id": id})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al buscar el trabajador",
			Cause:   err.Error(),
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

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &input); err != nil {
		logging.LogControllerError(c.Ctx, "trabajadores.post.bad_json", err, nil)
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Error al decodificar la solicitud",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	var trabajador models.Trabajador

	if doc, ok := input["documentoTrabajador"].(float64); ok {
		trabajador.PK_DOCUMENTO_TRABAJADOR = int64(doc)
	} else {
		logging.LogControllerError(c.Ctx, "trabajadores.post.validation_error", nil, map[string]interface{}{"missing": "documentoTrabajador"})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El campo 'documentoTrabajador' es obligatorio y debe ser un número válido",
		}
		_ = c.ServeJSON()
		return
	}

	if nombre, ok := input["nombre"].(string); ok && nombre != "" {
		trabajador.NOMBRE = nombre
	} else {
		logging.LogControllerError(c.Ctx, "trabajadores.post.validation_error", nil, map[string]interface{}{"missing": "nombre"})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El campo 'nombre' es obligatorio",
		}
		_ = c.ServeJSON()
		return
	}

	if apellido, ok := input["apellido"].(string); ok && apellido != "" {
		trabajador.APELLIDO = apellido
	} else {
		logging.LogControllerError(c.Ctx, "trabajadores.post.validation_error", nil, map[string]interface{}{"missing": "apellido"})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El campo 'apellido' es obligatorio",
		}
		_ = c.ServeJSON()
		return
	}

	if rol, ok := input["rol"].(string); ok && rol != "" {
		trabajador.ROL = models.RolTrabajador(rol)
	} else {
		logging.LogControllerError(c.Ctx, "trabajadores.post.validation_error", nil, map[string]interface{}{"missing": "rol"})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El campo 'rol' es obligatorio",
		}
		_ = c.ServeJSON()
		return
	}

	if fechaIngresoStr, ok := input["fechaIngreso"].(string); ok && fechaIngresoStr != "" {
		parsedDate, err := time.Parse("2006-01-02", fechaIngresoStr)
		if err != nil {
			logging.LogControllerError(c.Ctx, "trabajadores.post.bad_fecha_ingreso", err, map[string]interface{}{"fechaIngreso": fechaIngresoStr})
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
		logging.LogControllerError(c.Ctx, "trabajadores.post.validation_error", nil, map[string]interface{}{"missing": "fechaIngreso"})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El campo 'fechaIngreso' es obligatorio",
		}
		_ = c.ServeJSON()
		return
	}

	if sueldo, ok := input["sueldo"].(float64); ok {
		trabajador.SUELDO = int64(sueldo)
	} else {
		logging.LogControllerError(c.Ctx, "trabajadores.post.validation_error", nil, map[string]interface{}{"missing": "sueldo"})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El campo 'sueldo' es obligatorio",
		}
		_ = c.ServeJSON()
		return
	}

	if password, ok := input["password"].(string); ok && password != "" {
		hashedPassword, err := hashPassword(password)
		if err != nil {
			logging.LogControllerError(c.Ctx, "trabajadores.post.hash_error", err, nil)
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
		logging.LogControllerError(c.Ctx, "trabajadores.post.validation_error", nil, map[string]interface{}{"missing": "password"})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El campo 'password' es obligatorio",
		}
		_ = c.ServeJSON()
		return
	}

	if telefono, ok := input["telefono"].(string); ok {
		trabajador.TELEFONO = &telefono
	}

	if pkRestaurante, ok := input["restauranteId"].(float64); ok {
		valor := int64(pkRestaurante)
		trabajador.PK_ID_RESTAURANTE = &models.Restaurante{PK_ID_RESTAURANTE: valor}
	}

	if fechaNacimientoStr, ok := input["fechaNacimiento"].(string); ok && fechaNacimientoStr != "" {
		parsedDate, err := time.Parse("2006-01-02", fechaNacimientoStr)
		if err != nil {
			logging.LogControllerError(c.Ctx, "trabajadores.post.bad_fecha_nacimiento", err, map[string]interface{}{"fechaNacimiento": fechaNacimientoStr})
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

	_, err := o.Insert(&trabajador)
	if err != nil {
		logging.LogControllerError(c.Ctx, "trabajadores.post.insert_error", err, map[string]interface{}{"documento": trabajador.PK_DOCUMENTO_TRABAJADOR})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al crear el trabajador",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

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
		logging.LogControllerError(c.Ctx, "trabajadores.put.bad_request", err, map[string]interface{}{"id": c.GetString("id")})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "ID inválido o ausente",
		}
		_ = c.ServeJSON()
		return
	}

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

	var input map[string]interface{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &input); err != nil {
		logging.LogControllerError(c.Ctx, "trabajadores.put.bad_json", err, map[string]interface{}{"id": id})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Error al decodificar los datos",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

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
			logging.LogControllerError(c.Ctx, "trabajadores.put.bad_fecha_ingreso", err, map[string]interface{}{"id": id, "fecha_ingreso": v})
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
			logging.LogControllerError(c.Ctx, "trabajadores.put.bad_fecha_retiro", err, map[string]interface{}{"id": id, "fecha_retiro": v})
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
			logging.LogControllerError(c.Ctx, "trabajadores.put.bad_fecha_nacimiento", err, map[string]interface{}{"id": id, "fecha_nacimiento": v})
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
			logging.LogControllerError(c.Ctx, "trabajadores.put.hash_error", err, map[string]interface{}{"id": id})
			c.Ctx.Output.SetStatus(http.StatusInternalServerError)
			c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al procesar la contraseña", Cause: err.Error()}
			_ = c.ServeJSON()
			return
		}
		trabajador.PASSWORD = hashedPassword
	}

	if err := validateDates(&trabajador.FECHA_INGRESO, trabajador.FECHA_RETIRO); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: err.Error()}
		_ = c.ServeJSON()
		return
	}

	if _, err := o.Update(&trabajador); err != nil {
		logging.LogControllerError(c.Ctx, "trabajadores.put.update_error", err, map[string]interface{}{"id": id})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al actualizar el trabajador", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}

	trabajador.PASSWORD = ""

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

	idStr := c.GetString("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id == 0 {
		logging.LogControllerError(c.Ctx, "trabajadores.delete.bad_request", err, map[string]interface{}{"id": idStr})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El parámetro 'id' es inválido o está ausente",
			Cause:   "Se requiere un ID numérico válido en el parámetro 'id'",
		}
		_ = c.ServeJSON()
		return
	}

	trabajador := models.Trabajador{PK_DOCUMENTO_TRABAJADOR: int64(id)}
	if err := o.Read(&trabajador); err != nil {
		if err == orm.ErrNoRows {
			c.Ctx.Output.SetStatus(http.StatusOK)
			c.Data["json"] = models.ApiResponse{Code: http.StatusNotFound, Message: "Trabajador no encontrado"}
		} else {
			logging.LogControllerError(c.Ctx, "trabajadores.delete.db_read_error", err, map[string]interface{}{"id": id})
			c.Ctx.Output.SetStatus(http.StatusInternalServerError)
			c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al buscar el trabajador", Cause: err.Error()}
		}
		_ = c.ServeJSON()
		return
	}

	fechaRetiro := time.Now().In(database.BogotaZone)
	trabajador.FECHA_RETIRO = &fechaRetiro

	if _, err := o.Update(&trabajador, "FECHA_RETIRO"); err != nil {
		logging.LogControllerError(c.Ctx, "trabajadores.delete.update_error", err, map[string]interface{}{"id": id})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al actualizar la fecha de retiro del trabajador", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}
	trabajador.PASSWORD = ""
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{Code: http.StatusOK, Message: "Fecha de retiro del trabajador actualizada correctamente", Data: trabajador}
	_ = c.ServeJSON()
}
