package cambioshorario

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

// CambiosHorarioController gestiona /cambios_horario. GET /actual es público;
// el resto exige token (filtro loginc.ValidateToken en el router).
type CambiosHorarioController struct {
	web.Controller
}

const (
	horaCierreCerrado   = "23:59:59"
	horaAperturaCerrado = "00:00:00"
)

// colombia es la zona del restaurante (UTC-5, sin horario de verano): "hoy" se
// calcula con la fecha local, no con la fecha UTC (que cambia a las 19:00).
var colombia = time.FixedZone("COT", -5*60*60)

// now permite fijar la hora en los tests.
var now = time.Now

// wallClock devuelve la hora de pared de un valor leído de la BD (que el driver
// entrega con el desfase LMT que corrige models.FormatTimeWithLMT).
func wallClock(t time.Time) time.Time {
	// FormatTimeWithLMT siempre produce HH:MM:SS, por lo que no puede fallar.
	w, _ := models.ParseTimeToUTC(models.FormatTimeWithLMT(t))
	return w
}

// response arma la respuesta de un cambio de horario cuyas horas ya son de pared.
func response(h models.CambiosHorario) models.CambiosHorarioResponse {
	r := models.CambiosHorarioResponse{
		CambioHorarioId:    h.PK_ID_CAMBIO_HORARIO,
		FechaCambioHorario: models.FormatDateUTC(h.FECHA),
		HoraCierre:         h.HORA_CIERRE.Format("15:04:05"),
		Abierto:            h.ABIERTO,
	}
	if h.HORA_APERTURA != nil {
		s := h.HORA_APERTURA.Format("15:04:05")
		r.HoraApertura = &s
	}
	return r
}

// cerrado fija las horas de un día cerrado (00:00:00 - 23:59:59).
func cerrado(h *models.CambiosHorario) {
	// constantes válidas: ParseTimeToUTC no puede fallar.
	ha, _ := models.ParseTimeToUTC(horaAperturaCerrado)
	hc, _ := models.ParseTimeToUTC(horaCierreCerrado)
	h.HORA_APERTURA, h.HORA_CIERRE = &ha, hc
}

// fail registra el error y responde con el status indicado (code == status HTTP).
func (c *CambiosHorarioController) fail(status int, event, message string, err error, fields map[string]interface{}) {
	logging.LogControllerError(c.Ctx, event, err, fields)
	httpx.Fail(&c.Controller, status, message, err)
}

// @Title GetAll
// @Summary Listar cambios de horario
// @Description Devuelve todos los cambios de horario ordenados por fecha (fechas DD-MM-YYYY, horas HH:MM:SS). Sin resultados responde 200 con `data: []`. Requiere token.
// @Tags cambios_horario
// @Accept json
// @Produce json
// @Success 200 {object} models.ApiResponse{data=[]models.CambiosHorarioResponse} "Listado de cambios de horario (puede ser vacío)"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /cambios_horario [get]
func (c *CambiosHorarioController) GetAll() {
	var horarios []models.CambiosHorario
	if _, err := orm.NewOrm().QueryTable(new(models.CambiosHorario)).OrderBy("FECHA").All(&horarios); err != nil {
		c.fail(http.StatusInternalServerError, "cambios_horario.getall.db_error", "Error al obtener cambios de horario", err, nil)
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, "Cambios de horario obtenidos correctamente", httpx.List(horarios))
}

// @Title GetByCurrentDate
// @Summary Consultar el cambio de horario de hoy
// @Description Endpoint público (sin token). Devuelve el cambio de horario que aplica a la fecha actual en Colombia (UTC-5). Si no hay ninguno para hoy responde 404.
// @Tags cambios_horario
// @Accept json
// @Produce json
// @Success 200 {object} models.ApiResponse{data=models.CambiosHorarioResponse} "Cambio de horario de hoy"
// @Failure 404 {object} models.ApiResponse "No hay cambios de horario para la fecha actual"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Router /cambios_horario/actual [get]
func (c *CambiosHorarioController) GetByCurrentDate() {
	local := now().In(colombia)
	dateNoon := time.Date(local.Year(), local.Month(), local.Day(), 12, 0, 0, 0, time.UTC)

	var cambio models.CambiosHorario
	if err := orm.NewOrm().QueryTable(new(models.CambiosHorario)).Filter("FECHA", dateNoon).One(&cambio); err != nil {
		if errors.Is(err, orm.ErrNoRows) {
			httpx.Fail(&c.Controller, http.StatusNotFound, "No hay cambios de horario para la fecha actual", nil)
			return
		}
		c.fail(http.StatusInternalServerError, "cambios_horario.actual.db_error", "Error al consultar cambios de horario", err, map[string]interface{}{"date": dateNoon})
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, "Cambio de horario encontrado para la fecha actual", cambio)
}

// @Title Post
// @Summary Crear un cambio de horario
// @Description Crea un cambio de horario. Obligatorios: fechaCambioHorario (YYYY-MM-DD) y abierto. Si abierto=true, horaApertura y horaCierre (HH:MM o HH:MM:SS) son obligatorias; si abierto=false se ignoran y se fijan 00:00:00 - 23:59:59. Requiere token.
// @Tags cambios_horario
// @Accept json
// @Produce json
// @Param   body body models.CambiosHorarioCreateRequest true "Datos del cambio de horario"
// @Success 201 {object} models.ApiResponse{data=models.CambiosHorarioResponse} "Cambio de horario creado"
// @Failure 400 {object} models.ApiResponse "Solicitud inválida (JSON, campos obligatorios, fecha u hora)"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 409 {object} models.ApiResponse "Ya existe un cambio de horario para esa fecha"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /cambios_horario [post]
func (c *CambiosHorarioController) Post() {
	body := c.Ctx.Input.RequestBody
	var in models.CambiosHorarioCreateRequest
	if err := json.Unmarshal(body, &in); err != nil {
		c.fail(http.StatusBadRequest, "cambios_horario.post.bad_json", "Error al procesar la solicitud", err, nil)
		return
	}
	present, _ := httpx.Present(body)
	bad := func(message string, err error) {
		c.fail(http.StatusBadRequest, "cambios_horario.post.validation_error", message, err, nil)
	}

	if strings.TrimSpace(in.FechaCambioHorario) == "" {
		bad("El campo 'fechaCambioHorario' es obligatorio", nil)
		return
	}
	fecha, err := models.ParseDateToNoonUTC(in.FechaCambioHorario)
	if err != nil {
		bad("Formato de fecha inválido para 'fechaCambioHorario', use YYYY-MM-DD", err)
		return
	}
	if !present["abierto"] {
		bad("El campo 'abierto' es obligatorio", nil)
		return
	}

	horario := models.CambiosHorario{FECHA: fecha, ABIERTO: in.Abierto}
	if !in.Abierto {
		cerrado(&horario)
	} else {
		if in.HoraApertura == nil || strings.TrimSpace(*in.HoraApertura) == "" {
			bad("El campo 'horaApertura' es obligatorio cuando 'abierto' es true", nil)
			return
		}
		if in.HoraCierre == nil || strings.TrimSpace(*in.HoraCierre) == "" {
			bad("El campo 'horaCierre' es obligatorio cuando 'abierto' es true", nil)
			return
		}
		apertura, err := models.ParseTimeToUTC(strings.TrimSpace(*in.HoraApertura))
		if err != nil {
			bad("Formato de hora inválido para 'horaApertura' (use HH:MM:SS o HH:MM)", err)
			return
		}
		cierre, err := models.ParseTimeToUTC(strings.TrimSpace(*in.HoraCierre))
		if err != nil {
			bad("Formato de hora inválido para 'horaCierre' (use HH:MM:SS o HH:MM)", err)
			return
		}
		horario.HORA_APERTURA, horario.HORA_CIERRE = &apertura, cierre
	}

	if _, err := orm.NewOrm().Insert(&horario); err != nil {
		if dberr.IsUnique(err) {
			c.fail(http.StatusConflict, "cambios_horario.post.unique_conflict", "Ya existe un cambio de horario para esa fecha", err, nil)
			return
		}
		c.fail(http.StatusInternalServerError, "cambios_horario.post.db_error", "Error al crear el cambio de horario", err, nil)
		return
	}

	httpx.Send(&c.Controller, http.StatusCreated, "Cambio de horario creado correctamente", response(horario))
}

// @Title Update
// @Summary Actualizar un cambio de horario
// @Description Actualización parcial con merge: los campos ausentes se CONSERVAN y null en cualquiera responde 400 (no hay campos anulables). Si el resultado es abierto=false las horas se fuerzan a 00:00:00 - 23:59:59. Si pasa de cerrado a abierto, horaApertura y horaCierre son obligatorias en la petición. Requiere token.
// @Tags cambios_horario
// @Accept json
// @Produce json
// @Param   id   query int true "ID del cambio de horario (entero positivo)"
// @Param   body body models.CambiosHorarioUpdateRequest true "Campos a modificar (todos opcionales)"
// @Success 200 {object} models.ApiResponse{data=models.CambiosHorarioResponse} "Cambio de horario actualizado"
// @Failure 400 {object} models.ApiResponse "Solicitud inválida (id, JSON, null, fecha u hora)"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 404 {object} models.ApiResponse "Cambio de horario no encontrado"
// @Failure 409 {object} models.ApiResponse "Ya existe un cambio de horario para esa fecha"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /cambios_horario [put]
func (c *CambiosHorarioController) Put() {
	id, err := httpx.PositiveInt64Param(&c.Controller, "id")
	if err != nil {
		c.fail(http.StatusBadRequest, "cambios_horario.put.bad_request", "ID inválido o ausente", err, map[string]interface{}{"id": c.GetString("id")})
		return
	}

	o := orm.NewOrm()
	var horario models.CambiosHorario
	if err := o.QueryTable(new(models.CambiosHorario)).Filter("PK_ID_CAMBIO_HORARIO", id).One(&horario); err != nil {
		if errors.Is(err, orm.ErrNoRows) {
			httpx.Fail(&c.Controller, http.StatusNotFound, "Cambio de horario no encontrado", nil)
			return
		}
		c.fail(http.StatusInternalServerError, "cambios_horario.put.db_error", "Error al buscar el cambio de horario", err, map[string]interface{}{"id": id})
		return
	}

	var in models.CambiosHorarioUpdateRequest
	if err := httpx.DecodeMerge(c.Ctx.Input.RequestBody, &in); err != nil {
		c.fail(http.StatusBadRequest, "cambios_horario.put.bad_json", "Error al procesar la solicitud", err, map[string]interface{}{"id": id})
		return
	}
	bad := func(message string, err error) {
		c.fail(http.StatusBadRequest, "cambios_horario.put.validation_error", message, err, map[string]interface{}{"id": id})
	}

	// Horas de pared (las leídas de la BD traen desfase LMT) para no re-guardarlas corridas.
	if horario.HORA_APERTURA != nil {
		w := wallClock(*horario.HORA_APERTURA)
		horario.HORA_APERTURA = &w
	}
	horario.HORA_CIERRE = wallClock(horario.HORA_CIERRE)
	reabre := in.Abierto != nil && *in.Abierto && !horario.ABIERTO

	if in.FechaCambioHorario != nil {
		fecha, err := models.ParseDateToNoonUTC(strings.TrimSpace(*in.FechaCambioHorario))
		if err != nil {
			bad("Formato de fecha inválido para 'fechaCambioHorario', use YYYY-MM-DD", err)
			return
		}
		horario.FECHA = fecha
	}
	if in.Abierto != nil {
		horario.ABIERTO = *in.Abierto
	}

	if !horario.ABIERTO {
		cerrado(&horario)
	} else {
		if reabre && (in.HoraApertura == nil || in.HoraCierre == nil) {
			bad("Al pasar a abierto=true, 'horaApertura' y 'horaCierre' son obligatorias", nil)
			return
		}
		if in.HoraApertura != nil {
			t, err := models.ParseTimeToUTC(strings.TrimSpace(*in.HoraApertura))
			if err != nil {
				bad("Formato de hora inválido para 'horaApertura' (use HH:MM:SS o HH:MM)", err)
				return
			}
			horario.HORA_APERTURA = &t
		}
		if in.HoraCierre != nil {
			t, err := models.ParseTimeToUTC(strings.TrimSpace(*in.HoraCierre))
			if err != nil {
				bad("Formato de hora inválido para 'horaCierre' (use HH:MM:SS o HH:MM)", err)
				return
			}
			horario.HORA_CIERRE = t
		}
	}

	if _, err := o.Update(&horario); err != nil {
		if dberr.IsUnique(err) {
			c.fail(http.StatusConflict, "cambios_horario.put.unique_conflict", "Ya existe un cambio de horario para esa fecha", err, map[string]interface{}{"id": id})
			return
		}
		c.fail(http.StatusInternalServerError, "cambios_horario.put.db_error", "Error al actualizar el cambio de horario", err, map[string]interface{}{"id": id})
		return
	}

	httpx.Send(&c.Controller, http.StatusOK, "Cambio de horario actualizado correctamente", response(horario))
}

// @Title Delete
// @Summary Eliminar un cambio de horario
// @Description Elimina un cambio de horario. Si algún restaurante lo referencia responde 409. Requiere token.
// @Tags cambios_horario
// @Accept json
// @Produce json
// @Param   id query int true "ID del cambio de horario (entero positivo)"
// @Success 200 {object} models.ApiResponse "Cambio de horario eliminado"
// @Failure 400 {object} models.ApiResponse "ID inválido o ausente"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 404 {object} models.ApiResponse "Cambio de horario no encontrado"
// @Failure 409 {object} models.ApiResponse "El cambio de horario está en uso"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /cambios_horario [delete]
func (c *CambiosHorarioController) Delete() {
	id, err := httpx.PositiveInt64Param(&c.Controller, "id")
	if err != nil {
		c.fail(http.StatusBadRequest, "cambios_horario.delete.bad_request", "ID inválido o ausente", err, map[string]interface{}{"id": c.GetString("id")})
		return
	}

	// DELETE directo (no QueryTable.Delete): el ORM borraría en cascada los
	// restaurantes que referencian el cambio; así lo decide la FK de la BD (409).
	res, err := orm.NewOrm().Raw("DELETE FROM cambios_horario WHERE pk_id_cambio_horario = ?", id).Exec()
	if err != nil {
		if dberr.IsForeignKey(err) {
			c.fail(http.StatusConflict, "cambios_horario.delete.fk_conflict", "No se puede eliminar: el cambio de horario está en uso", err, map[string]interface{}{"id": id})
			return
		}
		c.fail(http.StatusInternalServerError, "cambios_horario.delete.db_error", "Error al eliminar el cambio de horario", err, map[string]interface{}{"id": id})
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		httpx.Fail(&c.Controller, http.StatusNotFound, "Cambio de horario no encontrado", nil)
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, "Cambio de horario eliminado correctamente", nil)
}
