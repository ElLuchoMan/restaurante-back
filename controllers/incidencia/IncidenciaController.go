package incidencia

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"restaurante/internal/dberr"
	"restaurante/internal/httpx"
	"restaurante/logging"
	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

// IncidenciaController gestiona /incidencias (requiere token).
type IncidenciaController struct {
	web.Controller
}

// fail registra el error y responde con el status indicado (code == status HTTP).
func (c *IncidenciaController) fail(status int, event, message string, err error, fields map[string]interface{}) {
	logging.LogControllerError(c.Ctx, event, err, fields)
	httpx.Fail(&c.Controller, status, message, err)
}

// writeError traduce errores de escritura: FK inexistente -> 400, resto -> 500.
func (c *IncidenciaController) writeError(op, message string, err error, fields map[string]interface{}) {
	if dberr.IsForeignKey(err) {
		c.fail(http.StatusBadRequest, "incidencias."+op+".fk_error", "El trabajador indicado no existe", err, fields)
		return
	}
	c.fail(http.StatusInternalServerError, "incidencias."+op+".db_error", message, err, fields)
}

// @Title GetAll
// @Summary Listar incidencias
// @Description Devuelve todas las incidencias (fechas DD-MM-YYYY). Sin resultados responde 200 con `data: []`. Requiere token.
// @Tags incidencias
// @Accept json
// @Produce json
// @Success 200 {object} models.ApiResponse{data=[]models.IncidenciaResponse} "Lista de incidencias (puede ser vacía)"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /incidencias [get]
func (c *IncidenciaController) GetAll() {
	var incidencias []models.Incidencia
	if _, err := orm.NewOrm().QueryTable(new(models.Incidencia)).OrderBy("PK_ID_INCIDENCIA").All(&incidencias); err != nil {
		c.fail(http.StatusInternalServerError, "incidencias.getall.db_error", "Error al obtener incidencias", err, nil)
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, "Incidencias obtenidas correctamente", httpx.List(incidencias))
}

// @Title GetByDocumentAndDate
// @Summary Buscar incidencias de un trabajador en un mes
// @Description Devuelve las incidencias de un trabajador en el mes y año indicados. Todos los parámetros son obligatorios. Si no hay incidencias responde 200 con `data: []`. Requiere token.
// @Tags incidencias
// @Accept json
// @Produce json
// @Param   documento     query    int     true   "Documento del trabajador (entero positivo)"
// @Param   mes           query    int     true   "Mes de la incidencia (1-12)" minimum(1) maximum(12)
// @Param   anio          query    int     true   "Año de la incidencia (1900 hasta el año actual)"
// @Success 200 {object} models.ApiResponse{data=[]models.IncidenciaResponse} "Incidencias encontradas (puede ser vacía)"
// @Failure 400 {object} models.ApiResponse "Parámetro 'documento', 'mes' o 'anio' inválido o ausente"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /incidencias/search [get]
func (c *IncidenciaController) GetByDocumentAndDate() {
	documento, err := httpx.PositiveInt64Param(&c.Controller, "documento")
	if err != nil {
		c.fail(http.StatusBadRequest, "incidencias.search.bad_request", "El parámetro 'documento' es inválido o ausente", err, map[string]interface{}{"documento": c.GetString("documento")})
		return
	}

	mes, err := c.GetInt("mes")
	if err != nil || mes < 1 || mes > 12 {
		c.fail(http.StatusBadRequest, "incidencias.search.bad_request", "El parámetro 'mes' es inválido. Debe estar entre 1 y 12", err, map[string]interface{}{"mes": c.GetString("mes")})
		return
	}

	anio, err := c.GetInt("anio")
	if err != nil || anio < 1900 || anio > time.Now().Year() {
		c.fail(http.StatusBadRequest, "incidencias.search.bad_request", "El parámetro 'anio' es inválido o ausente", err, map[string]interface{}{"anio": c.GetString("anio")})
		return
	}

	fechaInicio := time.Date(anio, time.Month(mes), 1, 12, 0, 0, 0, time.UTC)
	fechaFin := fechaInicio.AddDate(0, 1, -1)

	var incidencias []models.Incidencia
	if _, err := orm.NewOrm().QueryTable(new(models.Incidencia)).
		Filter("PK_DOCUMENTO_TRABAJADOR", documento).
		Filter("FECHA__gte", fechaInicio).
		Filter("FECHA__lte", fechaFin).
		OrderBy("FECHA", "PK_ID_INCIDENCIA").
		All(&incidencias); err != nil {
		c.fail(http.StatusInternalServerError, "incidencias.search.db_error", "Error al buscar incidencias", err, map[string]interface{}{"documento": documento, "mes": mes, "anio": anio})
		return
	}

	httpx.Send(&c.Controller, http.StatusOK, "Incidencias encontradas", httpx.List(incidencias))
}

// @Title Post
// @Summary Crear una incidencia
// @Description Crea una incidencia para un trabajador existente. Obligatorios: documentoTrabajador (> 0), fechaIncidencia (YYYY-MM-DD), monto (>= 0), resta (booleano) y motivo. Requiere token.
// @Tags incidencias
// @Accept json
// @Produce json
// @Param body body models.IncidenciaCreateRequest true "Datos de la incidencia"
// @Success 201 {object} models.ApiResponse{data=models.IncidenciaResponse} "Incidencia creada"
// @Failure 400 {object} models.ApiResponse "Solicitud inválida (JSON, campos obligatorios, fecha o trabajador inexistente)"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /incidencias [post]
func (c *IncidenciaController) Post() {
	body := c.Ctx.Input.RequestBody
	var in models.IncidenciaCreateRequest
	if err := json.Unmarshal(body, &in); err != nil {
		c.fail(http.StatusBadRequest, "incidencias.post.bad_json", "Error al procesar la solicitud", err, nil)
		return
	}
	present, _ := httpx.Present(body)
	bad := func(field, message string, err error) {
		c.fail(http.StatusBadRequest, "incidencias.post.validation_error", message, err, map[string]interface{}{"field": field})
	}

	if strings.TrimSpace(in.FechaIncidencia) == "" {
		bad("fechaIncidencia", "El campo fechaIncidencia es obligatorio", nil)
		return
	}
	fecha, err := models.ParseDateToNoonUTC(in.FechaIncidencia)
	if err != nil {
		bad("fechaIncidencia", "Formato de fecha inválido para fechaIncidencia, use YYYY-MM-DD", err)
		return
	}
	if !present["monto"] {
		bad("monto", "El campo monto es obligatorio", nil)
		return
	}
	if in.Monto < 0 {
		bad("monto", "El campo monto no puede ser negativo", nil)
		return
	}
	if !present["resta"] {
		bad("resta", "El campo resta es obligatorio", nil)
		return
	}
	motivo := strings.TrimSpace(in.Motivo)
	if motivo == "" {
		bad("motivo", "El campo motivo es obligatorio", nil)
		return
	}
	if in.DocumentoTrabajador <= 0 {
		bad("documentoTrabajador", "El campo documentoTrabajador es obligatorio y debe ser un número positivo", nil)
		return
	}

	incidencia := models.Incidencia{
		FECHA:                   fecha,
		MONTO:                   in.Monto,
		RESTA:                   in.Resta,
		MOTIVO:                  motivo,
		PK_DOCUMENTO_TRABAJADOR: &models.Trabajador{PK_DOCUMENTO_TRABAJADOR: in.DocumentoTrabajador},
	}
	if _, err := orm.NewOrm().Insert(&incidencia); err != nil {
		c.writeError("post", "Error al crear la incidencia", err, map[string]interface{}{"fecha": in.FechaIncidencia, "doc": in.DocumentoTrabajador})
		return
	}

	httpx.Send(&c.Controller, http.StatusCreated, "Incidencia creada correctamente", incidencia)
}

// @Title Update
// @Summary Actualizar una incidencia
// @Description Actualización parcial con merge: los campos ausentes se CONSERVAN y null en cualquiera responde 400 (no hay campos anulables). Fecha en YYYY-MM-DD; monto >= 0; motivo no vacío; documentoTrabajador debe existir. Requiere token.
// @Tags incidencias
// @Accept json
// @Produce json
// @Param id query int true "ID de la incidencia (entero positivo)"
// @Param body body models.IncidenciaUpdateRequest true "Campos a modificar (todos opcionales)"
// @Success 200 {object} models.ApiResponse{data=models.IncidenciaResponse} "Incidencia actualizada"
// @Failure 400 {object} models.ApiResponse "Solicitud inválida (id, JSON, null, fecha, monto, motivo o trabajador inexistente)"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 404 {object} models.ApiResponse "Incidencia no encontrada"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /incidencias [put]
func (c *IncidenciaController) Put() {
	id, err := httpx.PositiveInt64Param(&c.Controller, "id")
	if err != nil {
		c.fail(http.StatusBadRequest, "incidencias.put.bad_request", "El parámetro 'id' es inválido o está ausente", err, map[string]interface{}{"id": c.GetString("id")})
		return
	}

	o := orm.NewOrm()
	incidencia := models.Incidencia{PK_ID_INCIDENCIA: id}
	if err := o.Read(&incidencia); err != nil {
		if errors.Is(err, orm.ErrNoRows) {
			httpx.Fail(&c.Controller, http.StatusNotFound, "Incidencia no encontrada", nil)
			return
		}
		c.fail(http.StatusInternalServerError, "incidencias.put.db_read_error", "Error al buscar la incidencia", err, map[string]interface{}{"id": id})
		return
	}

	var in models.IncidenciaUpdateRequest
	if err := httpx.DecodeMerge(c.Ctx.Input.RequestBody, &in); err != nil {
		c.fail(http.StatusBadRequest, "incidencias.put.bad_json", "Error al decodificar la solicitud", err, map[string]interface{}{"id": id})
		return
	}
	bad := func(message string, err error) {
		c.fail(http.StatusBadRequest, "incidencias.put.validation_error", message, err, map[string]interface{}{"id": id})
	}

	if in.FechaIncidencia != nil {
		fecha, err := models.ParseDateToNoonUTC(strings.TrimSpace(*in.FechaIncidencia))
		if err != nil {
			bad("Formato de fecha inválido para fechaIncidencia, use YYYY-MM-DD", err)
			return
		}
		incidencia.FECHA = fecha
	}
	if in.Monto != nil {
		if *in.Monto < 0 {
			bad("El campo monto no puede ser negativo", nil)
			return
		}
		incidencia.MONTO = *in.Monto
	}
	if in.Resta != nil {
		incidencia.RESTA = *in.Resta
	}
	if in.Motivo != nil {
		motivo := strings.TrimSpace(*in.Motivo)
		if motivo == "" {
			bad("El campo motivo no puede estar vacío", nil)
			return
		}
		incidencia.MOTIVO = motivo
	}
	if in.DocumentoTrabajador != nil {
		if *in.DocumentoTrabajador <= 0 {
			bad("El campo documentoTrabajador debe ser un número positivo", nil)
			return
		}
		incidencia.PK_DOCUMENTO_TRABAJADOR = &models.Trabajador{PK_DOCUMENTO_TRABAJADOR: *in.DocumentoTrabajador}
	}

	if _, err := o.Update(&incidencia); err != nil {
		c.writeError("put", "Error al actualizar la incidencia", err, map[string]interface{}{"id": id})
		return
	}

	httpx.Send(&c.Controller, http.StatusOK, "Incidencia actualizada correctamente", incidencia)
}

// @Title Delete
// @Summary Eliminar una incidencia
// @Description Elimina una incidencia. Si no existe responde 404. Requiere token.
// @Tags incidencias
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID de la incidencia (entero positivo)"
// @Success 200 {object} models.ApiResponse "Incidencia eliminada"
// @Failure 400 {object} models.ApiResponse "ID inválido o ausente"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 404 {object} models.ApiResponse "Incidencia no encontrada"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /incidencias [delete]
func (c *IncidenciaController) Delete() {
	id, err := httpx.PositiveInt64Param(&c.Controller, "id")
	if err != nil {
		c.fail(http.StatusBadRequest, "incidencias.delete.bad_request", "ID inválido o ausente", err, map[string]interface{}{"id": c.GetString("id")})
		return
	}

	n, err := orm.NewOrm().Delete(&models.Incidencia{PK_ID_INCIDENCIA: id})
	if err != nil {
		c.fail(http.StatusInternalServerError, "incidencias.delete.db_error", "Error al eliminar la incidencia", err, map[string]interface{}{"id": id})
		return
	}
	if n == 0 {
		httpx.Fail(&c.Controller, http.StatusNotFound, "Incidencia no encontrada", nil)
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, "Incidencia eliminada correctamente", nil)
}
