package horariotrabajador

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

// HorarioTrabajadorController gestiona /horario_trabajador (requiere token).
type HorarioTrabajadorController struct {
	web.Controller
}

// diaToDB normaliza un día ("lunes", "LUNES") a su forma canónica ("Lunes") y
// devuelve false si no es un día de la semana.
func diaToDB(d string) (string, bool) {
	d = strings.TrimSpace(strings.ToLower(d))
	if d == "" {
		return "", false
	}
	canon := strings.ToUpper(d[:1]) + d[1:]
	switch models.DiaSemana(canon) {
	case models.DiaLunes, models.DiaMartes, models.DiaMiercoles,
		models.DiaJueves, models.DiaViernes, models.DiaSabado, models.DiaDomingo:
		return canon, true
	}
	return "", false
}

// wallClock devuelve la hora de pared de un valor leído de la BD (que el driver
// entrega con el desfase LMT que corrige models.FormatTimeWithLMT).
func wallClock(t time.Time) time.Time {
	// FormatTimeWithLMT siempre produce HH:MM:SS, por lo que no puede fallar.
	w, _ := models.ParseTimeToUTC(models.FormatTimeWithLMT(t))
	return w
}

func horarioResponse(doc int64, dia string, inicio, fin time.Time) models.HorarioTrabajadorResponse {
	return models.HorarioTrabajadorResponse{
		DocumentoTrabajador: doc,
		Dia:                 dia,
		HoraInicio:          inicio.Format("15:04:05"),
		HoraFin:             fin.Format("15:04:05"),
	}
}

// fail registra el error y responde con el status indicado (code == status HTTP).
func (c *HorarioTrabajadorController) fail(status int, event, message string, err error, fields map[string]interface{}) {
	logging.LogControllerError(c.Ctx, event, err, fields)
	httpx.Fail(&c.Controller, status, message, err)
}

// requireKey lee y valida los parámetros `documento` y `dia` (obligatorios).
func (c *HorarioTrabajadorController) requireKey(op string) (int64, string, bool) {
	doc, err := httpx.PositiveInt64Param(&c.Controller, "documento")
	if err != nil {
		c.fail(http.StatusBadRequest, "horario_trabajador."+op+".bad_request", "Parámetro 'documento' inválido o ausente", err, map[string]interface{}{"documento": c.GetString("documento")})
		return 0, "", false
	}
	dia, ok := diaToDB(c.GetString("dia"))
	if !ok {
		c.fail(http.StatusBadRequest, "horario_trabajador."+op+".bad_request", "Parámetro 'dia' inválido o ausente", nil, map[string]interface{}{"dia": c.GetString("dia")})
		return 0, "", false
	}
	return doc, dia, true
}

// writeError traduce errores de escritura: unicidad -> 409, FK -> 400, resto -> 500.
func (c *HorarioTrabajadorController) writeError(op, message string, err error, doc int64, dia string) {
	fields := map[string]interface{}{"documento": doc, "dia": dia}
	switch {
	case dberr.IsUnique(err):
		c.fail(http.StatusConflict, "horario_trabajador."+op+".unique_conflict", "El trabajador ya tiene un horario para ese día", err, fields)
	case dberr.IsForeignKey(err):
		c.fail(http.StatusBadRequest, "horario_trabajador."+op+".fk_error", "El trabajador indicado no existe", err, fields)
	default:
		c.fail(http.StatusInternalServerError, "horario_trabajador."+op+".db_error", message, err, fields)
	}
}

// @Title GetAll
// @Summary Listar horarios de trabajadores
// @Description Lista los horarios semanales (documentoTrabajador, dia, horaInicio, horaFin en HH:MM:SS), opcionalmente filtrados por documento del trabajador y/o día. Un `documento` o `dia` inválido responde 400. Sin resultados responde 200 con `data: []`. Requiere token.
// @Tags horarios_trabajador
// @Accept json
// @Produce json
// @Param   documento  query int    false "Documento del trabajador (entero positivo)"
// @Param   dia        query string false "Día a filtrar (no distingue mayúsculas)" Enums(Lunes, Martes, Miércoles, Jueves, Viernes, Sábado, Domingo)
// @Success 200 {object} models.ApiResponse{data=[]models.HorarioTrabajadorResponse} "Lista de horarios (puede ser vacía)"
// @Failure 400 {object} models.ApiResponse "Parámetro 'documento' o 'dia' inválido"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /horario_trabajador [get]
func (c *HorarioTrabajadorController) GetAll() {
	query := "SELECT pk_documento_trabajador, dia, hora_inicio, hora_fin FROM horario_trabajador"
	var args []interface{}
	var conds []string

	if raw := strings.TrimSpace(c.GetString("documento")); raw != "" {
		doc, err := httpx.PositiveInt64Param(&c.Controller, "documento")
		if err != nil {
			c.fail(http.StatusBadRequest, "horario_trabajador.getall.bad_request", "Parámetro 'documento' inválido", err, map[string]interface{}{"documento": raw})
			return
		}
		conds = append(conds, "pk_documento_trabajador = ?")
		args = append(args, doc)
	}
	if raw := strings.TrimSpace(c.GetString("dia")); raw != "" {
		dia, ok := diaToDB(raw)
		if !ok {
			c.fail(http.StatusBadRequest, "horario_trabajador.getall.bad_request", "Parámetro 'dia' inválido", nil, map[string]interface{}{"dia": raw})
			return
		}
		conds = append(conds, "dia = ?")
		args = append(args, dia)
	}
	if len(conds) > 0 {
		query += " WHERE " + strings.Join(conds, " AND ")
	}
	query += " ORDER BY pk_documento_trabajador, dia"

	var horarios []models.HorarioTrabajador
	if _, err := orm.NewOrm().Raw(query, args...).QueryRows(&horarios); err != nil {
		c.fail(http.StatusInternalServerError, "horario_trabajador.getall.db_error", "Error al obtener horarios", err, map[string]interface{}{"documento": c.GetString("documento"), "dia": c.GetString("dia")})
		return
	}

	httpx.Send(&c.Controller, http.StatusOK, "Horarios obtenidos correctamente", httpx.List(horarios))
}

// @Title Post
// @Summary Crear horario para un trabajador
// @Description Crea el horario de un trabajador para un día (un trabajador solo puede tener un horario por día). Obligatorios: documentoTrabajador, dia, horaInicio y horaFin (HH:MM o HH:MM:SS; horaFin debe ser mayor que horaInicio). Requiere token.
// @Tags horarios_trabajador
// @Accept json
// @Produce json
// @Param   body  body   models.HorarioTrabajadorCreateRequest true  "Datos del horario"
// @Success 201 {object} models.ApiResponse{data=models.HorarioTrabajadorResponse} "Horario creado"
// @Failure 400 {object} models.ApiResponse "Solicitud inválida (JSON, día, horas, o el trabajador no existe)"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 409 {object} models.ApiResponse "El trabajador ya tiene horario para ese día"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /horario_trabajador [post]
func (c *HorarioTrabajadorController) Post() {
	var in models.HorarioTrabajadorCreateRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &in); err != nil {
		c.fail(http.StatusBadRequest, "horario_trabajador.post.bad_json", "Error al decodificar la solicitud", err, nil)
		return
	}
	bad := func(message string, err error, fields map[string]interface{}) {
		c.fail(http.StatusBadRequest, "horario_trabajador.post.validation_error", message, err, fields)
	}
	if in.DocumentoTrabajador <= 0 {
		bad("El campo 'documentoTrabajador' es obligatorio y debe ser un número positivo", nil, nil)
		return
	}
	dia, ok := diaToDB(in.Dia)
	if !ok {
		bad("Día inválido", nil, map[string]interface{}{"dia": in.Dia})
		return
	}
	inicio, err1 := models.ParseTimeToUTC(strings.TrimSpace(in.HoraInicio))
	fin, err2 := models.ParseTimeToUTC(strings.TrimSpace(in.HoraFin))
	if err := errors.Join(err1, err2); err != nil {
		bad("Formato de hora inválido (use HH:MM:SS o HH:MM)", err, map[string]interface{}{"horaInicio": in.HoraInicio, "horaFin": in.HoraFin})
		return
	}
	if !fin.After(inicio) {
		bad("horaFin debe ser mayor que horaInicio", nil, map[string]interface{}{"horaInicio": in.HoraInicio, "horaFin": in.HoraFin})
		return
	}

	if _, err := orm.NewOrm().Raw("INSERT INTO horario_trabajador (pk_documento_trabajador, dia, hora_inicio, hora_fin) VALUES (?, ?, ?, ?)", in.DocumentoTrabajador, dia, inicio, fin).Exec(); err != nil {
		c.writeError("post", "Error al crear horario", err, in.DocumentoTrabajador, dia)
		return
	}

	httpx.Send(&c.Controller, http.StatusCreated, "Horario creado correctamente", horarioResponse(in.DocumentoTrabajador, dia, inicio, fin))
}

// @Title Put
// @Summary Actualizar horario de un trabajador
// @Description Actualización parcial con merge de un horario (identificado por `documento` y `dia`): `horaInicio` y `horaFin` ausentes se CONSERVAN; null en cualquiera responde 400 (no hay campos anulables). Formato HH:MM o HH:MM:SS; horaFin debe ser mayor que horaInicio. Requiere token.
// @Tags horarios_trabajador
// @Accept json
// @Produce json
// @Param   documento query int    true "Documento del trabajador (entero positivo)"
// @Param   dia       query string true "Día del horario" Enums(Lunes, Martes, Miércoles, Jueves, Viernes, Sábado, Domingo)
// @Param   body      body  models.HorarioTrabajadorUpdateRequest true "Horas a modificar (todas opcionales)"
// @Success 200 {object} models.ApiResponse{data=models.HorarioTrabajadorResponse} "Horario actualizado"
// @Failure 400 {object} models.ApiResponse "Solicitud inválida (parámetros, JSON, null, formato de hora u horas incoherentes)"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 404 {object} models.ApiResponse "Horario no encontrado"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /horario_trabajador [put]
func (c *HorarioTrabajadorController) Put() {
	doc, dia, ok := c.requireKey("put")
	if !ok {
		return
	}

	o := orm.NewOrm()
	var horario models.HorarioTrabajador
	if err := o.Raw("SELECT pk_documento_trabajador, dia, hora_inicio, hora_fin FROM horario_trabajador WHERE pk_documento_trabajador = ? AND dia = ?", doc, dia).QueryRow(&horario); err != nil {
		if errors.Is(err, orm.ErrNoRows) {
			httpx.Fail(&c.Controller, http.StatusNotFound, "Horario no encontrado", nil)
			return
		}
		c.fail(http.StatusInternalServerError, "horario_trabajador.put.db_error", "Error al consultar horario", err, map[string]interface{}{"documento": doc, "dia": dia})
		return
	}

	var in models.HorarioTrabajadorUpdateRequest
	if err := httpx.DecodeMerge(c.Ctx.Input.RequestBody, &in); err != nil {
		c.fail(http.StatusBadRequest, "horario_trabajador.put.bad_json", "Error al decodificar la solicitud", err, map[string]interface{}{"documento": doc, "dia": dia})
		return
	}

	inicio, fin := wallClock(horario.HORA_INICIO), wallClock(horario.HORA_FIN)
	bad := func(message string, err error) {
		c.fail(http.StatusBadRequest, "horario_trabajador.put.validation_error", message, err, map[string]interface{}{"documento": doc, "dia": dia})
	}
	if in.HoraInicio != nil {
		t, err := models.ParseTimeToUTC(strings.TrimSpace(*in.HoraInicio))
		if err != nil {
			bad("Formato de horaInicio inválido (use HH:MM:SS o HH:MM)", err)
			return
		}
		inicio = t
	}
	if in.HoraFin != nil {
		t, err := models.ParseTimeToUTC(strings.TrimSpace(*in.HoraFin))
		if err != nil {
			bad("Formato de horaFin inválido (use HH:MM:SS o HH:MM)", err)
			return
		}
		fin = t
	}
	if !fin.After(inicio) {
		bad("horaFin debe ser mayor que horaInicio", nil)
		return
	}

	if _, err := o.Raw("UPDATE horario_trabajador SET hora_inicio = ?, hora_fin = ? WHERE pk_documento_trabajador = ? AND dia = ?", inicio, fin, doc, dia).Exec(); err != nil {
		c.fail(http.StatusInternalServerError, "horario_trabajador.put.update_error", "Error al actualizar horario", err, map[string]interface{}{"documento": doc, "dia": dia})
		return
	}

	httpx.Send(&c.Controller, http.StatusOK, "Horario actualizado correctamente", horarioResponse(doc, dia, inicio, fin))
}

// @Title Delete
// @Summary Eliminar horario de un trabajador
// @Description Elimina el horario de un trabajador para un día. Requiere token.
// @Tags horarios_trabajador
// @Accept json
// @Produce json
// @Param   documento query int    true "Documento del trabajador (entero positivo)"
// @Param   dia       query string true "Día del horario" Enums(Lunes, Martes, Miércoles, Jueves, Viernes, Sábado, Domingo)
// @Success 200 {object} models.ApiResponse "Horario eliminado"
// @Failure 400 {object} models.ApiResponse "Parámetros inválidos"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 404 {object} models.ApiResponse "Horario no encontrado"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /horario_trabajador [delete]
func (c *HorarioTrabajadorController) Delete() {
	doc, dia, ok := c.requireKey("delete")
	if !ok {
		return
	}

	res, err := orm.NewOrm().Raw("DELETE FROM horario_trabajador WHERE pk_documento_trabajador = ? AND dia = ?", doc, dia).Exec()
	if err != nil {
		c.fail(http.StatusInternalServerError, "horario_trabajador.delete.delete_error", "Error al eliminar horario", err, map[string]interface{}{"documento": doc, "dia": dia})
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		httpx.Fail(&c.Controller, http.StatusNotFound, "Horario no encontrado", nil)
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, "Horario eliminado correctamente", nil)
}
