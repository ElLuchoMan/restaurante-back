package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"restaurante/models"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

type NominaTrabajadorController struct {
	web.Controller
}

// Interfaces y adaptadores para inyectar el ORM en tests
type ntQuerySeter interface {
	Filter(string, ...interface{}) ntQuerySeter
	All(interface{}, ...string) (int64, error)
	One(interface{}, ...string) error
	OrderBy(...string) ntQuerySeter
	Exist() bool
}

type ntOrmer interface {
	QueryTable(interface{}) ntQuerySeter
	Insert(interface{}) (int64, error)
}

type ntQSAdapter struct{ qs orm.QuerySeter }

func (a ntQSAdapter) Filter(expr string, args ...interface{}) ntQuerySeter {
	return ntQSAdapter{qs: a.qs.Filter(expr, args...)}
}
func (a ntQSAdapter) All(res interface{}, cols ...string) (int64, error) {
	return a.qs.All(res, cols...)
}
func (a ntQSAdapter) One(res interface{}, cols ...string) error { return a.qs.One(res, cols...) }

//go:noinline
func (a ntQSAdapter) OrderBy(expr ...string) ntQuerySeter {
	return ntQSAdapter{qs: a.qs.OrderBy(expr...)}
}

//go:noinline
func (a ntQSAdapter) Exist() bool { return a.qs.Exist() }

type ntOrmAdapter struct{ o orm.Ormer }

func (a ntOrmAdapter) QueryTable(i interface{}) ntQuerySeter {
	return ntQSAdapter{qs: a.o.QueryTable(i)}
}

//go:noinline
func (a ntOrmAdapter) Insert(v interface{}) (int64, error) { return a.o.Insert(v) }

var nomtraOrmNew = func() ntOrmer { return ntOrmAdapter{o: orm.NewOrm()} }

// @Title GetAll
// @Summary Obtener todas las relaciones nómina-trabajador
// @Description Obtiene un listado de todas las relaciones nómina-trabajador registradas en la base de datos
// @Tags nomina_trabajador
// @Accept json
// @Produce json
// @Success 200 {object} models.ApiResponse{data=[]models.NominaTrabajador} "Listado de relaciones nómina-trabajador"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /nomina_trabajador [get]
func (c *NominaTrabajadorController) GetAll() {
	o := nomtraOrmNew()
	var relaciones []models.NominaTrabajador

	// Obtener las relaciones desde la base de datos
	_, err := o.QueryTable(new(models.NominaTrabajador)).All(&relaciones)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener las relaciones nómina-trabajador",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	// Responder con éxito
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Relaciones nómina-trabajador obtenidas correctamente",
		Data:    relaciones,
	}
	_ = c.ServeJSON()
}

// @Title Post
// @Summary Crear una nómina-trabajador con cálculo automático
// @Description Crea una nueva relación nómina-trabajador, calculando incidencias y total a pagar basado en el sueldo y las incidencias del trabajador.
// @Tags nomina_trabajador
// @Accept json
// @Produce json
// @Param body body models.NominaTrabajadorRequest true "Datos de la nómina-trabajador"
// @Success 201 {object} models.ApiResponse{data=models.NominaTrabajadorResponse} "Nómina-trabajador creada"
// @Failure 400 {object} models.ApiResponse "Error en la solicitud"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /nomina_trabajador [post]
func (c *NominaTrabajadorController) Post() {
	o := nomtraOrmNew()
	var input models.NominaTrabajadorRequest
	var nominaTrabajador models.NominaTrabajador

	// Decodificar la solicitud
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &input); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Error al procesar la solicitud",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	// Validar documento del trabajador
	if input.PK_DOCUMENTO_TRABAJADOR == 0 {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El campo documentoTrabajador es obligatorio y debe ser válido",
		}
		_ = c.ServeJSON()
		return
	}
	nominaTrabajador.PK_DOCUMENTO_TRABAJADOR = &models.Trabajador{PK_DOCUMENTO_TRABAJADOR: input.PK_DOCUMENTO_TRABAJADOR}

	// Calcular el rango de fechas
	now := time.Now()
	startDate := time.Date(now.Year(), now.Month()-1, 20, 0, 0, 0, 0, now.Location())
	endDate := time.Date(now.Year(), now.Month(), 20, 23, 59, 59, 999, now.Location())

	// Consultar incidencias
	var incidencias []models.Incidencia
	_, err := o.QueryTable(new(models.Incidencia)).
		Filter("pk_documento_trabajador", input.PK_DOCUMENTO_TRABAJADOR).
		Filter("fecha__gte", startDate).
		Filter("fecha__lte", endDate).
		All(&incidencias)

	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al consultar incidencias del trabajador",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	// Calcular el monto de incidencias
	var montoIncidencias int64
	for _, incidencia := range incidencias {
		if incidencia.RESTA {
			montoIncidencias -= incidencia.MONTO
		} else {
			montoIncidencias += incidencia.MONTO
		}
	}
	nominaTrabajador.MONTO_INCIDENCIAS = &montoIncidencias

	// Consultar el sueldo del trabajador
	var trabajador models.Trabajador
	err = o.QueryTable(new(models.Trabajador)).
		Filter("pk_documento_trabajador", input.PK_DOCUMENTO_TRABAJADOR).
		One(&trabajador)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al consultar el sueldo del trabajador",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}
	nominaTrabajador.SUELDO_BASE = trabajador.SUELDO

	// Registrar en la base de datos
	// Asignar la nómina correspondiente (usar la nómina más reciente si no se especifica)
	var ultimaNomina models.Nomina
	err = o.QueryTable(new(models.Nomina)).OrderBy("-fecha").One(&ultimaNomina)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener la nómina activa",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}
	nominaTrabajador.PK_ID_NOMINA = &ultimaNomina

	// Generar descripción dinámica con mes y año reales de la nómina
	descripcion := fmt.Sprintf("Nómina del mes de %s de %d más incidencias si aplica", obtenerMesEnEspañol(ultimaNomina.FECHA.Month()), ultimaNomina.FECHA.Year())
	nominaTrabajador.DETALLES = &descripcion

	// Verificar si ya existe la relación (mismo trabajador y misma nómina)
	exists := o.QueryTable(new(models.NominaTrabajador)).
		Filter("pk_documento_trabajador", input.PK_DOCUMENTO_TRABAJADOR).
		Filter("pk_id_nomina", ultimaNomina.PK_ID_NOMINA).
		Exist()
	if exists {
		// Idempotente: devolver 200 con la relación existente
		var existente models.NominaTrabajador
		if err := o.QueryTable(new(models.NominaTrabajador)).
			Filter("pk_documento_trabajador", input.PK_DOCUMENTO_TRABAJADOR).
			Filter("pk_id_nomina", ultimaNomina.PK_ID_NOMINA).
			One(&existente); err == nil {
			c.Ctx.Output.SetStatus(http.StatusOK)
			c.Data["json"] = models.ApiResponse{Code: http.StatusOK, Message: "Relación nómina-trabajador ya existía", Data: existente}
			_ = c.ServeJSON()
			return
		}
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{Code: http.StatusOK, Message: "Relación nómina-trabajador ya existía"}
		_ = c.ServeJSON()
		return
	}

	_, err = o.Insert(&nominaTrabajador)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al registrar la nómina-trabajador",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	// Preparar la respuesta
	response := models.NominaTrabajadorResponse{
		SUELDO_BASE:             trabajador.SUELDO,
		MONTO_INCIDENCIAS:       montoIncidencias,
		DETALLES:                descripcion,
		PK_DOCUMENTO_TRABAJADOR: input.PK_DOCUMENTO_TRABAJADOR,
	}

	// Responder con éxito
	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusCreated,
		Message: "Nómina-trabajador creada correctamente",
		Data:    response,
	}
	_ = c.ServeJSON()
}

// @Title GetByTrabajador
// @Summary Obtener relaciones nómina-trabajador según filtros
// @Description Obtiene las relaciones nómina-trabajador según los filtros aplicados (nómina actual, nóminas pagas, nóminas no pagas, nómina por mes y año, todas las nóminas).
// @Tags nomina_trabajador
// @Accept json
// @Produce json
// @Param documento query int true "Documento del trabajador"
// @Param actual query bool false "Consultar solo la nómina actual"
// @Param pagas query bool false "Consultar solo nóminas pagadas"
// @Param no_pagas query bool false "Consultar solo nóminas no pagadas"
// @Param mes query int false "Mes (1-12) para filtrar nóminas"
// @Param anio query int false "Año (YYYY) para filtrar nóminas"
// @Success 200 {object} models.ApiResponse{data=[]models.NominaTrabajador} "Relaciones nómina-trabajador encontradas"
// @Failure 404 {object} models.ApiResponse "Relación nómina-trabajador no encontrada"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /nomina_trabajador/search [get]
func (c *NominaTrabajadorController) GetByTrabajador() {
	o := orm.NewOrm()
	documento, errDoc := c.GetInt64("documento")
	actual, errAct := c.GetBool("actual")
	pagas, errPag := c.GetBool("pagas")
	noPagas, errNoPag := c.GetBool("no_pagas")
	mes, errMes := c.GetInt("mes")
	anio, errAnio := c.GetInt("anio")

	if c.GetString("documento") == "" || errDoc != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Parámetro 'documento' es obligatorio y debe ser válido"}
		_ = c.ServeJSON()
		return
	}
	if c.GetString("actual") != "" && errAct != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Parámetro 'actual' inválido", Cause: errAct.Error()}
		_ = c.ServeJSON()
		return
	}
	if c.GetString("pagas") != "" && errPag != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Parámetro 'pagas' inválido", Cause: errPag.Error()}
		_ = c.ServeJSON()
		return
	}
	if c.GetString("no_pagas") != "" && errNoPag != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Parámetro 'no_pagas' inválido", Cause: errNoPag.Error()}
		_ = c.ServeJSON()
		return
	}
	if c.GetString("mes") != "" && errMes != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Parámetro 'mes' inválido", Cause: errMes.Error()}
		_ = c.ServeJSON()
		return
	}
	if c.GetString("anio") != "" && errAnio != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Parámetro 'anio' inválido", Cause: errAnio.Error()}
		_ = c.ServeJSON()
		return
	}

	// Validar el documento del trabajador
	if documento == 0 {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El parámetro 'documento' es obligatorio.",
		}
		_ = c.ServeJSON()
		return
	}

	// Base de la consulta
	var relaciones []models.NominaTrabajador
	sql := `
       SELECT nt.* FROM "nomina_trabajador" nt
       JOIN "nomina" n ON nt."pk_id_nomina" = n."pk_id_nomina"
       WHERE nt."pk_documento_trabajador" = ?
   `
	params := []interface{}{documento}

	// Filtrar por nómina actual
	if actual {
		sql += ` AND n."fecha" = (SELECT MAX("fecha") FROM "nomina")`
	}

	// Filtrar por nóminas pagas o no pagas
	if pagas {
		sql += ` AND n."estado_nomina" = 'PAGO'`
	} else if noPagas {
		sql += ` AND n."estado_nomina" = 'NO_PAGO'`
	}

	// Filtrar por mes y año
	if mes > 0 && anio > 0 {
		sql += ` AND EXTRACT(MONTH FROM n."fecha") = ? AND EXTRACT(YEAR FROM n."fecha") = ?`
		params = append(params, mes, anio)
	}

	// Ejecutar la consulta
	_, err := o.Raw(sql, params...).QueryRows(&relaciones)

	// Validar si hay resultados
	if err == orm.ErrNoRows || len(relaciones) == 0 {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "No se encontraron relaciones nómina-trabajador para los filtros aplicados.",
		}
		_ = c.ServeJSON()
		return
	} else if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al buscar las relaciones nómina-trabajador.",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	// Responder con éxito
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Relaciones nómina-trabajador encontradas.",
		Data:    relaciones,
	}
	_ = c.ServeJSON()
}

func obtenerMesEnEspañol(mes time.Month) string {
	meses := map[time.Month]string{
		time.January:   "Enero",
		time.February:  "Febrero",
		time.March:     "Marzo",
		time.April:     "Abril",
		time.May:       "Mayo",
		time.June:      "Junio",
		time.July:      "Julio",
		time.August:    "Agosto",
		time.September: "Septiembre",
		time.October:   "Octubre",
		time.November:  "Noviembre",
		time.December:  "Diciembre",
	}
	return meses[mes]
}

// @Title GetNominasByMes
// @Summary Consultar nóminas del mes actual o de un mes/año específico
// @Description Obtiene todas las relaciones nómina-trabajador del mes actual o de un mes/año específico, incluyendo el nombre y apellido del trabajador.
// @Tags nomina_trabajador
// @Accept json
// @Produce json
// @Param mes query int false "Mes (1-12) para filtrar nóminas"
// @Param anio query int false "Año (YYYY) para filtrar nóminas"
// @Success 200 {object} models.ApiResponse{data=[]map[string]interface{}} "Relaciones nómina-trabajador encontradas"
// @Failure 404 {object} models.ApiResponse "No se encontraron relaciones nómina-trabajador"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /nomina_trabajador/mes [get]
func (c *NominaTrabajadorController) GetNominasByMes() {
	o := orm.NewOrm()
	mes, errMes := c.GetInt("mes")
	anio, errAnio := c.GetInt("anio")
	if c.GetString("mes") == "" || c.GetString("anio") == "" || errMes != nil || errAnio != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Parámetros 'mes' y 'anio' obligatorios y válidos"}
		_ = c.ServeJSON()
		return
	}

	// Validar parámetros
	if mes < 1 || mes > 12 || anio < 1 {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Los parámetros 'mes' y 'anio' deben ser válidos.",
		}
		_ = c.ServeJSON()
		return
	}

	// Consulta SQL
	var resultados []models.NominaTrabajadorDetalle
	sql := `
       SELECT
               nt."sueldo_base",
               nt."monto_incidencias",
               nt."detalles",
               nt."pk_documento_trabajador",
               nt."pk_id_nomina",
               t."nombre",
               t."apellido"
       FROM "nomina_trabajador" nt
       JOIN "trabajador" t ON nt."pk_documento_trabajador" = t."pk_documento_trabajador"
       JOIN "nomina" n ON nt."pk_id_nomina" = n."pk_id_nomina"
       WHERE EXTRACT(MONTH FROM n."fecha") = ?
       AND EXTRACT(YEAR FROM n."fecha") = ?
`
	// Ejecutar la consulta
	num, err := o.Raw(sql, mes, anio).QueryRows(&resultados)
	fmt.Printf("Número de filas recuperadas: %d\n", num)
	fmt.Printf("Resultados: %+v\n", resultados)

	// Validar resultados
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al buscar las nóminas.",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	if len(resultados) == 0 {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "No se encontraron nóminas para el mes y año especificados.",
		}
		_ = c.ServeJSON()
		return
	}

	// Responder con éxito
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Nóminas encontradas.",
		Data:    resultados,
	}
	_ = c.ServeJSON()
}
