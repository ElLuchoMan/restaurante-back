package incidencia

import (
	"encoding/json"
	"net/http"
	"restaurante/logging"
	"restaurante/models"
	"strconv"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

type IncidenciaController struct {
	web.Controller
}

// Hooks de ORM para pruebas
type incidenciaOrmer interface {
	QueryTable(interface{}) orm.QuerySeter
	Insert(interface{}) (int64, error)
	Read(interface{}, ...string) error
	Update(interface{}, ...string) (int64, error)
	Delete(interface{}, ...string) (int64, error)
}
type incidenciaOrmAdapter struct{ o orm.Ormer }

func (a incidenciaOrmAdapter) QueryTable(i interface{}) orm.QuerySeter  { return a.o.QueryTable(i) }
func (a incidenciaOrmAdapter) Insert(v interface{}) (int64, error)      { return a.o.Insert(v) }
func (a incidenciaOrmAdapter) Read(v interface{}, cols ...string) error { return a.o.Read(v, cols...) }
func (a incidenciaOrmAdapter) Update(v interface{}, cols ...string) (int64, error) {
	return a.o.Update(v, cols...)
}
func (a incidenciaOrmAdapter) Delete(v interface{}, cols ...string) (int64, error) {
	return a.o.Delete(v, cols...)
}

var incidenciaOrmNew = func() incidenciaOrmer { return incidenciaOrmAdapter{o: orm.NewOrm()} }

// @Title GetAll
// @Summary Obtener todas las incidencias
// @Description Devuelve una lista de todas las incidencias registradas en la base de datos.
// @Tags incidencias
// @Accept json
// @Produce json
// @Success 200 {object} models.ApiResponse{data=[]models.Incidencia} "Lista de incidencias"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /incidencias [get]
func (c *IncidenciaController) GetAll() {
	o := incidenciaOrmNew()
	var incidencias []models.Incidencia

	_, err := o.QueryTable(new(models.Incidencia)).All(&incidencias)
	if err != nil {
		logging.LogControllerError(c.Ctx, "incidencias.getall.db_error", err, nil)
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener incidencias",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Incidencias obtenidas correctamente",
		Data:    incidencias,
	}
	_ = c.ServeJSON()
}

// @Title GetByDocumentAndDate
// @Summary Obtener incidencias por documento y/o fecha
// @Description Devuelve una lista de incidencias según los filtros proporcionados.
// @Tags incidencias
// @Accept json
// @Produce json
// @Param   documento     query    int     true   "Documento del Trabajador"
// @Param   mes           query    int     true   "Mes de la Incidencia (1-12)"
// @Param   anio          query    int     true   "Año de la Incidencia"
// @Success 200 {object} models.ApiResponse{data=[]models.Incidencia} "Lista de incidencias encontradas"
// @Failure 400 {object} models.ApiResponse "Error en la solicitud"
// @Failure 404 {object} models.ApiResponse "No se encontraron incidencias"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /incidencias/search [get]
func (c *IncidenciaController) GetByDocumentAndDate() {
	o := incidenciaOrmNew()

	// Obtener parámetros de la consulta
	documento, err := c.GetInt64("documento")
	if err != nil || documento == 0 {
		logging.LogControllerError(c.Ctx, "incidencias.search.bad_request", err, map[string]interface{}{"documento": c.GetString("documento")})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El parámetro 'documento' es inválido o ausente",
		}
		_ = c.ServeJSON()
		return
	}

	mes, err := c.GetInt("mes")
	if err != nil || mes < 1 || mes > 12 {
		logging.LogControllerError(c.Ctx, "incidencias.search.bad_request", err, map[string]interface{}{"mes": c.GetString("mes")})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El parámetro 'mes' es inválido. Debe estar entre 1 y 12",
		}
		_ = c.ServeJSON()
		return
	}

	anio, err := c.GetInt("anio")
	if err != nil || anio < 1900 || anio > time.Now().Year() {
		logging.LogControllerError(c.Ctx, "incidencias.search.bad_request", err, map[string]interface{}{"anio": c.GetString("anio")})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El parámetro 'anio' es inválido o ausente",
		}
		_ = c.ServeJSON()
		return
	}

	// Calcular rango de fechas
	fechaInicio := time.Date(anio, time.Month(mes), 1, 0, 0, 0, 0, time.UTC)
	fechaFin := fechaInicio.AddDate(0, 1, 0).Add(-time.Second) // Fin del mes

	// Filtrar las incidencias
	var incidencias []models.Incidencia
	_, err = o.QueryTable(new(models.Incidencia)).
		Filter("PK_DOCUMENTO_TRABAJADOR", documento).
		Filter("FECHA__gte", fechaInicio.Format("2006-01-02")).
		Filter("FECHA__lte", fechaFin.Format("2006-01-02")).
		All(&incidencias)

	if err == orm.ErrNoRows || len(incidencias) == 0 {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "No se encontraron incidencias para los parámetros proporcionados",
		}
		_ = c.ServeJSON()
		return
	} else if err != nil {
		logging.LogControllerError(c.Ctx, "incidencias.search.db_error", err, map[string]interface{}{"documento": documento, "mes": mes, "anio": anio})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al buscar incidencias",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	// Responder con las incidencias encontradas
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Incidencias encontradas",
		Data:    incidencias,
	}
	_ = c.ServeJSON()
}

// @Title Post
// @Summary Crear una nueva incidencia
// @Description Crea una nueva incidencia en la base de datos.
// @Tags incidencias
// @Accept json
// @Produce json
// @Param body body models.IncidenciaCreateRequest true "Datos de la incidencia (fecha YYYY-MM-DD)"
// @Success 201 {object} map[string]interface{} "Incidencia creada"
// @Failure 400 {object} models.ApiResponse "Error en la solicitud"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /incidencias [post]
func (c *IncidenciaController) Post() {
	o := incidenciaOrmNew()
	var input map[string]interface{}
	var incidencia models.Incidencia

	// Decodificar la solicitud
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &input); err != nil {
		logging.LogControllerError(c.Ctx, "incidencias.post.bad_json", err, nil)
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Error al procesar la solicitud",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	// Validar y procesar fechaIncidencia
	if fechaStr, ok := input["fechaIncidencia"].(string); ok && fechaStr != "" {
		parsedDate, err := time.Parse("2006-01-02", fechaStr)
		if err != nil {
			logging.LogControllerError(c.Ctx, "incidencias.post.bad_fecha", err, map[string]interface{}{"fechaIncidencia": fechaStr})
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusBadRequest,
				Message: "Formato de fecha inválido para fechaIncidencia",
				Cause:   err.Error(),
			}
			_ = c.ServeJSON()
			return
		}
		incidencia.FECHA = parsedDate
	} else {
		logging.LogControllerError(c.Ctx, "incidencias.post.validation_error", nil, map[string]interface{}{"missing": "fechaIncidencia"})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El campo fechaIncidencia es obligatorio",
		}
		_ = c.ServeJSON()
		return
	}

	// Validar y procesar monto
	if monto, ok := input["monto"].(float64); ok {
		incidencia.MONTO = int64(monto)
	} else {
		logging.LogControllerError(c.Ctx, "incidencias.post.validation_error", nil, map[string]interface{}{"missing": "monto"})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El campo monto es obligatorio y debe ser un número",
		}
		_ = c.ServeJSON()
		return
	}

	// Validar y procesar resta
	if resta, ok := input["resta"].(bool); ok {
		incidencia.RESTA = resta
	} else {
		logging.LogControllerError(c.Ctx, "incidencias.post.validation_error", nil, map[string]interface{}{"missing": "resta"})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El campo resta es obligatorio",
		}
		_ = c.ServeJSON()
		return
	}

	// Validar y procesar motivo
	if motivo, ok := input["motivo"].(string); ok && motivo != "" {
		incidencia.MOTIVO = motivo
	} else {
		logging.LogControllerError(c.Ctx, "incidencias.post.validation_error", nil, map[string]interface{}{"missing": "motivo"})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El campo motivo es obligatorio",
		}
		_ = c.ServeJSON()
		return
	}

	// Procesar documentoTrabajador (obligatorio)
	if documento, ok := input["documentoTrabajador"].(float64); ok && documento != 0 {
		doc := int64(documento)
		incidencia.PK_DOCUMENTO_TRABAJADOR = &models.Trabajador{PK_DOCUMENTO_TRABAJADOR: doc}
	} else {
		logging.LogControllerError(c.Ctx, "incidencias.post.validation_error", nil, map[string]interface{}{"missing": "documentoTrabajador"})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El campo documentoTrabajador es obligatorio",
		}
		_ = c.ServeJSON()
		return
	}

	// Insertar en la base de datos
	_, err := o.Insert(&incidencia)
	if err != nil {
		logging.LogControllerError(c.Ctx, "incidencias.post.insert_error", err, map[string]interface{}{"fecha": incidencia.FECHA, "doc": incidencia.PK_DOCUMENTO_TRABAJADOR.PK_DOCUMENTO_TRABAJADOR})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al crear la incidencia",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	// Preparar la respuesta con el formato deseado
	response := map[string]interface{}{
		"incidenciaId":    incidencia.PK_ID_INCIDENCIA,
		"fechaIncidencia": incidencia.FECHA.Format("2006-01-02"),
		"monto":           incidencia.MONTO,
		"resta":           incidencia.RESTA,
		"motivo":          incidencia.MOTIVO,
		"documentoTrabajador": func() int64 {
			if incidencia.PK_DOCUMENTO_TRABAJADOR != nil {
				return incidencia.PK_DOCUMENTO_TRABAJADOR.PK_DOCUMENTO_TRABAJADOR
			}
			return 0
		}(),
	}

	// Responder con éxito
	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusCreated,
		Message: "Incidencia creada correctamente",
		Data:    response,
	}
	_ = c.ServeJSON()
}

// @Title Update
// @Summary Actualizar una incidencia
// @Description Actualiza los datos de una incidencia existente en la base de datos.
// @Tags incidencias
// @Accept json
// @Produce json
// @Param id query int true "ID de la Incidencia"
// @Param body body models.IncidenciaUpdateRequest true "Datos de la incidencia a actualizar (sólo campos a modificar)"
// @Success 200 {object} map[string]interface{} "Incidencia actualizada"
// @Failure 400 {object} models.ApiResponse "Error en la solicitud"
// @Failure 404 {object} models.ApiResponse "Incidencia no encontrada"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /incidencias [put]
func (c *IncidenciaController) Put() {
	o := incidenciaOrmNew()

	// Obtener el ID de la incidencia desde los parámetros
	idStr := c.GetString("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id == 0 {
		logging.LogControllerError(c.Ctx, "incidencias.put.bad_request", err, map[string]interface{}{"id": idStr})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El parámetro 'id' es inválido o está ausente",
		}
		_ = c.ServeJSON()
		return
	}

	// Buscar la incidencia por ID
	incidencia := models.Incidencia{PK_ID_INCIDENCIA: int64(id)}
	if err := o.Read(&incidencia); err == orm.ErrNoRows {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Incidencia no encontrada",
		}
		_ = c.ServeJSON()
		return
	}

	// Deserializar los datos actualizados desde el cuerpo de la solicitud
	var input map[string]interface{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &input); err != nil {
		logging.LogControllerError(c.Ctx, "incidencias.put.bad_json", err, map[string]interface{}{"id": id})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Error al decodificar la solicitud",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	// Validar y actualizar los campos
	if fechaStr, ok := input["fechaIncidencia"].(string); ok && fechaStr != "" {
		parsedDate, err := time.Parse("2006-01-02", fechaStr)
		if err != nil {
			logging.LogControllerError(c.Ctx, "incidencias.put.bad_fecha", err, map[string]interface{}{"id": id, "fechaIncidencia": fechaStr})
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusBadRequest,
				Message: "Formato de fecha inválido para fechaIncidencia",
				Cause:   err.Error(),
			}
			_ = c.ServeJSON()
			return
		}
		incidencia.FECHA = parsedDate
	}

	if monto, ok := input["monto"].(float64); ok {
		incidencia.MONTO = int64(monto)
	}
	if resta, ok := input["resta"].(bool); ok {
		incidencia.RESTA = resta
	}
	if motivo, ok := input["motivo"].(string); ok && motivo != "" {
		incidencia.MOTIVO = motivo
	}
	if documento, ok := input["documentoTrabajador"].(float64); ok && documento != 0 {
		doc := int64(documento)
		incidencia.PK_DOCUMENTO_TRABAJADOR = &models.Trabajador{PK_DOCUMENTO_TRABAJADOR: doc}
	}

	// Guardar los cambios en la base de datos
	if _, err := o.Update(&incidencia); err != nil {
		logging.LogControllerError(c.Ctx, "incidencias.put.update_error", err, map[string]interface{}{"id": id})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al actualizar la incidencia",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	// Preparar la respuesta con el formato deseado
	response := map[string]interface{}{
		"incidenciaId":    incidencia.PK_ID_INCIDENCIA,
		"fechaIncidencia": incidencia.FECHA.Format("2006-01-02"),
		"monto":           incidencia.MONTO,
		"resta":           incidencia.RESTA,
		"motivo":          incidencia.MOTIVO,
		"documentoTrabajador": func() int64 {
			if incidencia.PK_DOCUMENTO_TRABAJADOR != nil {
				return incidencia.PK_DOCUMENTO_TRABAJADOR.PK_DOCUMENTO_TRABAJADOR
			}
			return 0
		}(),
	}

	// Responder con éxito
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Incidencia actualizada correctamente",
		Data:    response,
	}
	_ = c.ServeJSON()
}

// @Title Delete
// @Summary Eliminar una incidencia
// @Description Elimina una incidencia de la base de datos.
// @Tags incidencias
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID de la incidencia"
// @Success 200 {object} models.ApiResponse "Incidencia eliminada"
// @Failure 404 {object} models.ApiResponse "Incidencia no encontrada"
// @Security BearerAuth
// @Router /incidencias [delete]
func (c *IncidenciaController) Delete() {
	o := incidenciaOrmNew()
	id, err := c.GetInt64("id")
	if err != nil {
		logging.LogControllerError(c.Ctx, "incidencias.delete.bad_request", err, map[string]interface{}{"id": c.GetString("id")})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "ID inválido o ausente",
		}
		_ = c.ServeJSON()
		return
	}

	_, err = o.Delete(&models.Incidencia{PK_ID_INCIDENCIA: id})
	if err == orm.ErrNoRows {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Incidencia no encontrada",
		}
		_ = c.ServeJSON()
		return
	} else if err != nil {
		logging.LogControllerError(c.Ctx, "incidencias.delete.db_error", err, map[string]interface{}{"id": id})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al eliminar la incidencia",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Incidencia eliminada correctamente",
	}
	_ = c.ServeJSON()
}
