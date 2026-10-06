package nominatrabajador

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"restaurante/internal/dberr"
	"restaurante/internal/httpx"
	"restaurante/logging"
	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

// NominaTrabajadorController expone la relación nómina-trabajador. Las
// respuestas siempre llevan el envoltorio models.ApiResponse y el status HTTP
// coincide con su `code`. No existen PUT ni DELETE: la relación se crea con el
// cálculo automático y no se edita.
type NominaTrabajadorController struct {
	web.Controller
}

var meses = [...]string{"Enero", "Febrero", "Marzo", "Abril", "Mayo", "Junio", "Julio", "Agosto", "Septiembre", "Octubre", "Noviembre", "Diciembre"}

func obtenerMesEnEspañol(mes time.Month) string { return meses[mes-1] }

// itemDe convierte el modelo en la forma pública única (FK como ids; nulos como 0 y "").
func itemDe(n models.NominaTrabajador) models.NominaTrabajadorItem {
	item := models.NominaTrabajadorItem{
		PK_ID_NOMINA_TRABAJADOR: n.PK_ID_NOMINA_TRABAJADOR,
		SUELDO_BASE:             n.SUELDO_BASE,
	}
	if n.MONTO_INCIDENCIAS != nil {
		item.MONTO_INCIDENCIAS = *n.MONTO_INCIDENCIAS
	}
	if n.DETALLES != nil {
		item.DETALLES = *n.DETALLES
	}
	if n.PK_DOCUMENTO_TRABAJADOR != nil {
		item.PK_DOCUMENTO_TRABAJADOR = n.PK_DOCUMENTO_TRABAJADOR.PK_DOCUMENTO_TRABAJADOR
	}
	if n.PK_ID_NOMINA != nil {
		item.PK_ID_NOMINA = n.PK_ID_NOMINA.PK_ID_NOMINA
	}
	return item
}

// boolParam lee un query param booleano opcional (false si no viene).
func (c *NominaTrabajadorController) boolParam(key string) (bool, error) {
	if c.GetString(key) == "" {
		return false, nil
	}
	v, err := strconv.ParseBool(c.GetString(key))
	if err != nil {
		return false, fmt.Errorf("el parámetro '%s' debe ser true o false: %w", key, err)
	}
	return v, nil
}

// periodoParams lee mes (1-12) y anio (>0) opcionales; 0 si no vienen.
func (c *NominaTrabajadorController) periodoParams() (mes, anio int, err error) {
	if c.GetString("mes") != "" {
		if mes, err = c.GetInt("mes"); err != nil || mes < 1 || mes > 12 {
			return 0, 0, errors.New("el parámetro 'mes' debe ser un entero entre 1 y 12")
		}
	}
	if c.GetString("anio") != "" {
		if anio, err = c.GetInt("anio"); err != nil || anio < 1 {
			return 0, 0, errors.New("el parámetro 'anio' debe ser un entero positivo")
		}
	}
	return mes, anio, nil
}

// @Title GetAll
// @Summary Obtener todas las relaciones nómina-trabajador
// @Description Lista todas las relaciones nómina-trabajador. Cada elemento tiene la forma única `NominaTrabajadorItem`: las FK se responden como ids numéricos (`documentoTrabajador`, `nominaId`) y `montoIncidencias`/`detalles` nulos como 0 y "". Sin resultados: 200 con `data: []`.
// @Tags nomina_trabajador
// @Accept json
// @Produce json
// @Success 200 {object} models.ApiResponse{data=[]models.NominaTrabajadorItem} "Relaciones nómina-trabajador (puede ser [])"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /nomina_trabajador [get]
func (c *NominaTrabajadorController) GetAll() {
	var relaciones []models.NominaTrabajador
	if _, err := orm.NewOrm().QueryTable(new(models.NominaTrabajador)).All(&relaciones); err != nil {
		logging.LogControllerError(c.Ctx, "nomina_trabajador.getall.db_error", err, nil)
		httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error al obtener las relaciones nómina-trabajador", err)
		return
	}
	items := make([]models.NominaTrabajadorItem, 0, len(relaciones))
	for _, r := range relaciones {
		items = append(items, itemDe(r))
	}
	httpx.Send(&c.Controller, http.StatusOK, "Relaciones nómina-trabajador obtenidas correctamente", items)
}

// montoIncidencias suma las incidencias del trabajador entre el día 20 del mes
// anterior y el día 20 del mes actual (las que restan, restan).
func montoIncidencias(o orm.Ormer, documento int64, ahora time.Time) (int64, error) {
	desde := time.Date(ahora.Year(), ahora.Month()-1, 20, 0, 0, 0, 0, ahora.Location())
	hasta := time.Date(ahora.Year(), ahora.Month(), 20, 23, 59, 59, 999, ahora.Location())
	var incidencias []models.Incidencia
	if _, err := o.QueryTable(new(models.Incidencia)).
		Filter("PK_DOCUMENTO_TRABAJADOR", documento).
		Filter("FECHA__gte", desde).
		Filter("FECHA__lte", hasta).
		All(&incidencias); err != nil {
		return 0, err
	}
	var total int64
	for _, i := range incidencias {
		if i.RESTA {
			total -= i.MONTO
		} else {
			total += i.MONTO
		}
	}
	return total, nil
}

// @Title Post
// @Summary Crear la nómina-trabajador de la última nómina (cálculo automático)
// @Description Crea la relación entre el trabajador y la última nómina: sueldo base del trabajador, suma de incidencias (día 20 del mes anterior al día 20 del actual) y detalle los calcula el backend; cualquier otro campo del cuerpo se ignora. Respuesta de forma única (`NominaTrabajadorItem`, con `nominaTrabajadorId`): 201 si se creó o 200 si ya existía para esa nómina.
// @Tags nomina_trabajador
// @Accept json
// @Produce json
// @Param body body models.NominaTrabajadorRequest true "Documento del trabajador"
// @Success 201 {object} models.ApiResponse{data=models.NominaTrabajadorItem} "Relación creada"
// @Success 200 {object} models.ApiResponse{data=models.NominaTrabajadorItem} "La relación ya existía"
// @Failure 400 {object} models.ApiResponse "JSON inválido o documentoTrabajador ausente/inválido"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 404 {object} models.ApiResponse "Trabajador no encontrado"
// @Failure 409 {object} models.ApiResponse "La relación ya existe (concurrencia)"
// @Failure 422 {object} models.ApiResponse "No hay ninguna nómina generada"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /nomina_trabajador [post]
func (c *NominaTrabajadorController) Post() {
	var input models.NominaTrabajadorRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &input); err != nil {
		httpx.Fail(&c.Controller, http.StatusBadRequest, "Error al procesar la solicitud", err)
		return
	}
	if input.PK_DOCUMENTO_TRABAJADOR <= 0 {
		httpx.Fail(&c.Controller, http.StatusBadRequest, "El campo documentoTrabajador es obligatorio y debe ser un entero positivo", nil)
		return
	}
	documento := input.PK_DOCUMENTO_TRABAJADOR

	o := orm.NewOrm()
	var trabajador models.Trabajador
	if err := o.QueryTable(new(models.Trabajador)).Filter("PK_DOCUMENTO_TRABAJADOR", documento).One(&trabajador); err != nil {
		if errors.Is(err, orm.ErrNoRows) {
			httpx.Fail(&c.Controller, http.StatusNotFound, "Trabajador no encontrado", err)
			return
		}
		logging.LogControllerError(c.Ctx, "nomina_trabajador.post.trabajador_error", err, nil)
		httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error al consultar el trabajador", err)
		return
	}

	var ultima models.Nomina
	if err := o.QueryTable(new(models.Nomina)).OrderBy("-FECHA").One(&ultima); err != nil {
		if errors.Is(err, orm.ErrNoRows) {
			httpx.Fail(&c.Controller, http.StatusUnprocessableEntity, "No hay ninguna nómina generada", err)
			return
		}
		logging.LogControllerError(c.Ctx, "nomina_trabajador.post.nomina_error", err, nil)
		httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error al obtener la nómina activa", err)
		return
	}

	var existente models.NominaTrabajador
	err := o.QueryTable(new(models.NominaTrabajador)).
		Filter("PK_DOCUMENTO_TRABAJADOR", documento).
		Filter("PK_ID_NOMINA", ultima.PK_ID_NOMINA).
		One(&existente)
	if err == nil {
		httpx.Send(&c.Controller, http.StatusOK, "Relación nómina-trabajador ya existía", itemDe(existente))
		return
	}
	if !errors.Is(err, orm.ErrNoRows) {
		logging.LogControllerError(c.Ctx, "nomina_trabajador.post.existente_error", err, nil)
		httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error al consultar la relación nómina-trabajador", err)
		return
	}

	incidencias, err := montoIncidencias(o, documento, time.Now())
	if err != nil {
		logging.LogControllerError(c.Ctx, "nomina_trabajador.post.incidencias_error", err, nil)
		httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error al consultar incidencias del trabajador", err)
		return
	}
	detalle := fmt.Sprintf("Nómina del mes de %s de %d más incidencias si aplica", obtenerMesEnEspañol(ultima.FECHA.Month()), ultima.FECHA.Year())
	nueva := models.NominaTrabajador{
		SUELDO_BASE:             trabajador.SUELDO,
		MONTO_INCIDENCIAS:       &incidencias,
		DETALLES:                &detalle,
		PK_DOCUMENTO_TRABAJADOR: &models.Trabajador{PK_DOCUMENTO_TRABAJADOR: documento},
		PK_ID_NOMINA:            &models.Nomina{PK_ID_NOMINA: ultima.PK_ID_NOMINA},
	}
	id, err := o.Insert(&nueva)
	if err != nil {
		if dberr.IsUnique(err) {
			httpx.Fail(&c.Controller, http.StatusConflict, "La relación nómina-trabajador ya existe", err)
			return
		}
		logging.LogControllerError(c.Ctx, "nomina_trabajador.post.insert_error", err, nil)
		httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error al registrar la nómina-trabajador", err)
		return
	}
	nueva.PK_ID_NOMINA_TRABAJADOR = id
	httpx.Send(&c.Controller, http.StatusCreated, "Nómina-trabajador creada correctamente", itemDe(nueva))
}

// @Title GetByTrabajador
// @Summary Obtener relaciones nómina-trabajador de un trabajador
// @Description Devuelve las relaciones de un trabajador (forma `NominaTrabajadorItem`). Filtros combinables: `actual` (solo la última nómina), `pagas` / `no_pagas` (estado de la nómina; no pueden ir ambos en true), y `mes` y/o `anio` (por la fecha de la nómina). Sin resultados: 200 con `data: []`.
// @Tags nomina_trabajador
// @Accept json
// @Produce json
// @Param documento query int true "Documento del trabajador (entero positivo)"
// @Param actual query bool false "Solo la nómina actual (la más reciente)"
// @Param pagas query bool false "Solo nóminas pagadas"
// @Param no_pagas query bool false "Solo nóminas no pagadas"
// @Param mes query int false "Mes (1-12)"
// @Param anio query int false "Año (YYYY)"
// @Success 200 {object} models.ApiResponse{data=[]models.NominaTrabajadorItem} "Relaciones encontradas (puede ser [])"
// @Failure 400 {object} models.ApiResponse "Parámetros ausentes o inválidos"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /nomina_trabajador/search [get]
func (c *NominaTrabajadorController) GetByTrabajador() {
	documento, err := httpx.PositiveInt64Param(&c.Controller, "documento")
	if err != nil {
		httpx.Fail(&c.Controller, http.StatusBadRequest, "El parámetro 'documento' es obligatorio y debe ser un entero positivo", err)
		return
	}
	actual, errAct := c.boolParam("actual")
	pagas, errPag := c.boolParam("pagas")
	noPagas, errNoPag := c.boolParam("no_pagas")
	if err := errors.Join(errAct, errPag, errNoPag); err != nil {
		httpx.Fail(&c.Controller, http.StatusBadRequest, "Los parámetros 'actual', 'pagas' y 'no_pagas' deben ser true o false", err)
		return
	}
	if pagas && noPagas {
		httpx.Fail(&c.Controller, http.StatusBadRequest, "Los parámetros 'pagas' y 'no_pagas' no pueden usarse a la vez", nil)
		return
	}
	mes, anio, err := c.periodoParams()
	if err != nil {
		httpx.Fail(&c.Controller, http.StatusBadRequest, "Los parámetros 'mes' y 'anio' son inválidos", err)
		return
	}

	query := `
       SELECT nt."pk_id_nomina_trabajador", nt."sueldo_base",
              COALESCE(nt."monto_incidencias", 0) AS "monto_incidencias",
              COALESCE(nt."detalles", '') AS "detalles",
              nt."pk_documento_trabajador", nt."pk_id_nomina"
       FROM "nomina_trabajador" nt
       JOIN "nomina" n ON nt."pk_id_nomina" = n."pk_id_nomina"
       WHERE nt."pk_documento_trabajador" = ?`
	params := []interface{}{documento}
	if actual {
		query += ` AND n."fecha" = (SELECT MAX("fecha") FROM "nomina")`
	}
	if pagas {
		query += ` AND n."estado_nomina" = 'PAGO'`
	}
	if noPagas {
		query += ` AND n."estado_nomina" = 'NO_PAGO'`
	}
	if mes > 0 {
		query += ` AND EXTRACT(MONTH FROM n."fecha") = ?`
		params = append(params, mes)
	}
	if anio > 0 {
		query += ` AND EXTRACT(YEAR FROM n."fecha") = ?`
		params = append(params, anio)
	}
	query += ` ORDER BY n."fecha", nt."pk_id_nomina_trabajador"`

	var items []models.NominaTrabajadorItem
	if _, err := orm.NewOrm().Raw(query, params...).QueryRows(&items); err != nil {
		logging.LogControllerError(c.Ctx, "nomina_trabajador.search.db_error", err, map[string]interface{}{"documento": documento})
		httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error al buscar las relaciones nómina-trabajador", err)
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, "Relaciones nómina-trabajador encontradas", httpx.List(items))
}

// @Title GetNominasByMes
// @Summary Consultar las nóminas de los trabajadores de un mes
// @Description Devuelve las relaciones nómina-trabajador del mes/año indicados (por defecto, el mes y año actuales), con el nombre y apellido del trabajador y el `nominaTrabajadorId`. Sin resultados: 200 con `data: []`.
// @Tags nomina_trabajador
// @Accept json
// @Produce json
// @Param mes query int false "Mes (1-12); por defecto el actual"
// @Param anio query int false "Año (YYYY); por defecto el actual"
// @Success 200 {object} models.ApiResponse{data=[]models.NominaTrabajadorDetalle} "Relaciones encontradas (puede ser [])"
// @Failure 400 {object} models.ApiResponse "mes o anio inválidos"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /nomina_trabajador/mes [get]
func (c *NominaTrabajadorController) GetNominasByMes() {
	mes, anio, err := c.periodoParams()
	if err != nil {
		httpx.Fail(&c.Controller, http.StatusBadRequest, "Los parámetros 'mes' y 'anio' son inválidos", err)
		return
	}
	ahora := time.Now()
	if mes == 0 {
		mes = int(ahora.Month())
	}
	if anio == 0 {
		anio = ahora.Year()
	}

	var resultados []models.NominaTrabajadorDetalle
	const query = `
       SELECT
               nt."pk_id_nomina_trabajador",
               nt."sueldo_base",
               COALESCE(nt."monto_incidencias", 0) AS "monto_incidencias",
               COALESCE(nt."detalles", '') AS "detalles",
               nt."pk_documento_trabajador",
               nt."pk_id_nomina",
               t."nombre",
               t."apellido"
       FROM "nomina_trabajador" nt
       JOIN "trabajador" t ON nt."pk_documento_trabajador" = t."pk_documento_trabajador"
       JOIN "nomina" n ON nt."pk_id_nomina" = n."pk_id_nomina"
       WHERE EXTRACT(MONTH FROM n."fecha") = ?
       AND EXTRACT(YEAR FROM n."fecha") = ?
       ORDER BY nt."pk_id_nomina_trabajador"`
	if _, err := orm.NewOrm().Raw(query, mes, anio).QueryRows(&resultados); err != nil {
		logging.LogControllerError(c.Ctx, "nomina_trabajador.mes.db_error", err, nil)
		httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error al buscar las nóminas", err)
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, "Nóminas encontradas", httpx.List(resultados))
}
