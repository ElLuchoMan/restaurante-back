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

// @Title GetAll
// @Summary Obtener todas las relaciones nómina-trabajador
// @Description Obtiene un listado de todas las relaciones nómina-trabajador registradas en la base de datos
// @Tags nomina_trabajador
// @Accept json
// @Produce json
// @Success 200 {array} models.NominaTrabajador "Listado de relaciones nómina-trabajador"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /nomina_trabajador [get]
func (c *NominaTrabajadorController) GetAll() {
	o := orm.NewOrm()
	var relaciones []models.NominaTrabajador

	_, err := o.QueryTable(new(models.NominaTrabajador)).All(&relaciones)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener las relaciones nómina-trabajador",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Responder con éxito
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Relaciones nómina-trabajador obtenidas correctamente",
		Data:    relaciones,
	}
	c.ServeJSON()
}

// @Title Post
// @Summary Crear una nómina-trabajador con cálculo automático
// @Description Crea una nueva relación nómina-trabajador, calculando incidencias y total a pagar basado en el sueldo y las incidencias del trabajador.
// @Tags nomina_trabajador
// @Accept json
// @Produce json
// @Param body body models.NominaTrabajadorRequest true "Datos de la nómina-trabajador"
// @Success 201 {object} models.NominaTrabajadorResponse "Nómina-trabajador creada"
// @Failure 400 {object} models.ApiResponse "Error en la solicitud"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /nomina_trabajador [post]
func (c *NominaTrabajadorController) Post() {
	o := orm.NewOrm()
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
		c.ServeJSON()
		return
	}

	// Validar documento del trabajador
	if input.PK_DOCUMENTO_TRABAJADOR == 0 {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El campo PK_DOCUMENTO_TRABAJADOR es obligatorio y debe ser válido",
		}
		c.ServeJSON()
		return
	}
	nominaTrabajador.PK_DOCUMENTO_TRABAJADOR = input.PK_DOCUMENTO_TRABAJADOR

	// Calcular el rango de fechas
	now := time.Now()
	startDate := time.Date(now.Year(), now.Month()-1, 20, 0, 0, 0, 0, now.Location())
	endDate := time.Date(now.Year(), now.Month(), 20, 23, 59, 59, 999, now.Location())

	// Consultar incidencias
	var incidencias []models.Incidencia
	_, err := o.QueryTable(new(models.Incidencia)).
		Filter("PK_DOCUMENTO_TRABAJADOR", input.PK_DOCUMENTO_TRABAJADOR).
		Filter("FECHA__gte", startDate).
		Filter("FECHA__lte", endDate).
		All(&incidencias)

	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al consultar incidencias del trabajador",
			Cause:   err.Error(),
		}
		c.ServeJSON()
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
		Filter("PK_DOCUMENTO_TRABAJADOR", input.PK_DOCUMENTO_TRABAJADOR).
		One(&trabajador)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al consultar el sueldo del trabajador",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}
	nominaTrabajador.SUELDO_BASE = trabajador.SUELDO

	// Calcular el total a pagar
	total := trabajador.SUELDO + montoIncidencias
	nominaTrabajador.TOTAL = &total

	// Generar descripción dinámica
	descripcion := fmt.Sprintf("Nómina del mes de %s más incidencias si aplica", obtenerMesEnEspañol(now.Month()))
	nominaTrabajador.DETALLES = &descripcion

	// Registrar en la base de datos
	_, err = o.Insert(&nominaTrabajador)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al registrar la nómina-trabajador",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Preparar la respuesta
	response := models.NominaTrabajadorResponse{
		PK_ID_NOMINA_TRABAJADOR: nominaTrabajador.PK_ID_NOMINA_TRABAJADOR,
		SUELDO_BASE:             trabajador.SUELDO,
		MONTO_INCIDENCIAS:       montoIncidencias,
		TOTAL:                   total,
		DETALLES:                descripcion,
	}

	// Responder con éxito
	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusCreated,
		Message: "Nómina-trabajador creada correctamente",
		Data:    response,
	}
	c.ServeJSON()
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
// @Success 200 {array} models.NominaTrabajador "Relaciones nómina-trabajador encontradas"
// @Failure 404 {object} models.ApiResponse "Relación nómina-trabajador no encontrada"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /nomina_trabajador/search [get]
func (c *NominaTrabajadorController) GetByTrabajador() {
	o := orm.NewOrm()
	documento, _ := c.GetInt64("documento")
	actual, _ := c.GetBool("actual")
	pagas, _ := c.GetBool("pagas")
	noPagas, _ := c.GetBool("no_pagas")
	mes, _ := c.GetInt("mes")
	anio, _ := c.GetInt("anio")

	// Validar el documento del trabajador
	if documento == 0 {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El parámetro 'documento' es obligatorio.",
		}
		c.ServeJSON()
		return
	}

	// Base de la consulta
	var relaciones []models.NominaTrabajador
	sql := `
        SELECT nt.* FROM "NOMINA_TRABAJADOR" nt
        JOIN "NOMINA" n ON nt."PK_ID_NOMINA" = n."PK_ID_NOMINA"
        WHERE nt."PK_DOCUMENTO_TRABAJADOR" = ?
    `
	params := []interface{}{documento}

	// Filtrar por nómina actual
	if actual {
		sql += ` AND n."FECHA" = (SELECT MAX("FECHA") FROM "NOMINA")`
	}

	// Filtrar por nóminas pagas o no pagas
	if pagas {
		sql += ` AND n."ESTADO_NOMINA" = 'PAGO'`
	} else if noPagas {
		sql += ` AND n."ESTADO_NOMINA" = 'NO PAGO'`
	}

	// Filtrar por mes y año
	if mes > 0 && anio > 0 {
		sql += ` AND EXTRACT(MONTH FROM n."FECHA") = ? AND EXTRACT(YEAR FROM n."FECHA") = ?`
		params = append(params, mes, anio)
	}

	// Ejecutar la consulta
	_, err := o.Raw(sql, params...).QueryRows(&relaciones)

	// Validar si hay resultados
	if err == orm.ErrNoRows || len(relaciones) == 0 {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "No se encontraron relaciones nómina-trabajador para los filtros aplicados.",
		}
		c.ServeJSON()
		return
	} else if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al buscar las relaciones nómina-trabajador.",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Responder con éxito
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Relaciones nómina-trabajador encontradas.",
		Data:    relaciones,
	}
	c.ServeJSON()
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
// @Success 200 {array} map[string]interface{} "Relaciones nómina-trabajador encontradas"
// @Failure 404 {object} models.ApiResponse "No se encontraron relaciones nómina-trabajador"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /nomina_trabajador/mes [get]
func (c *NominaTrabajadorController) GetNominasByMes() {
	o := orm.NewOrm()
	mes, _ := c.GetInt("mes")
	anio, _ := c.GetInt("anio")

	// Validar parámetros
	if mes < 1 || mes > 12 || anio < 1 {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Los parámetros 'mes' y 'anio' deben ser válidos.",
		}
		c.ServeJSON()
		return
	}

	// Consulta SQL
	var resultados []models.NominaTrabajadorDetalle
	sql := `
	SELECT 
		nt."PK_ID_NOMINA_TRABAJADOR", 
		nt."SUELDO_BASE", 
		nt."MONTO_INCIDENCIAS", 
		nt."TOTAL", 
		nt."DETALLES", 
		nt."PK_DOCUMENTO_TRABAJADOR", 
		nt."PK_ID_NOMINA", 
		t."NOMBRE", 
		t."APELLIDO"
	FROM "NOMINA_TRABAJADOR" nt
	JOIN "TRABAJADOR" t ON nt."PK_DOCUMENTO_TRABAJADOR" = t."PK_DOCUMENTO_TRABAJADOR"
	JOIN "NOMINA" n ON nt."PK_ID_NOMINA" = n."PK_ID_NOMINA"
	WHERE EXTRACT(MONTH FROM n."FECHA") = ? 
	AND EXTRACT(YEAR FROM n."FECHA") = ?
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
		c.ServeJSON()
		return
	}

	if len(resultados) == 0 {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "No se encontraron nóminas para el mes y año especificados.",
		}
		c.ServeJSON()
		return
	}

	// Responder con éxito
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Nóminas encontradas.",
		Data:    resultados,
	}
	c.ServeJSON()
}
