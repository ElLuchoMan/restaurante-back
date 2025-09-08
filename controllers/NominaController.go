package controllers

import (
	"encoding/json"
	"net/http"
	"restaurante/models"
	"strconv"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

type NominaController struct {
	web.Controller
}

// puntos de inyección
var (
	ormNewNomina    = orm.NewOrm
	queryAllNominas = func(o orm.Ormer, out *[]models.Nomina) (int64, error) {
		return o.QueryTable(new(models.Nomina)).All(out)
	}
	readNominaFn         = func(o orm.Ormer, n *models.Nomina) error { return o.Read(n) }
	updateNominaFn       = func(o orm.Ormer, n *models.Nomina, cols ...string) (int64, error) { return o.Update(n, cols...) }
	findExistingNominaFn = func(o orm.Ormer, fecha time.Time) (*models.Nomina, error) {
		var existing models.Nomina
		err := o.Raw(
			"SELECT pk_id_nomina, fecha, monto, estado_nomina FROM nomina WHERE EXTRACT(YEAR FROM fecha) = ? AND EXTRACT(MONTH FROM fecha) = ? LIMIT 1",
			fecha.Year(), int(fecha.Month()),
		).QueryRow(&existing)
		if err != nil {
			return nil, err
		}
		return &existing, nil
	}
)

// Estados permitidos para la nómina
var estadosNominaPermitidos = map[models.EstadoNomina]bool{
	models.EstadoNominaPago:   true,
	models.EstadoNominaNoPago: true,
}

// @Title GetAll
// @Summary Obtener todas las nóminas con filtros
// @Description Devuelve todas las nóminas registradas en la base de datos, con opción de filtrar por fecha exacta, mes y año.
// @Tags nominas
// @Accept json
// @Produce json
// @Param   fecha    query   string   false   "Filtrar por fecha exacta (YYYY-MM-DD)"
// @Param   mes      query   int      false   "Filtrar por mes (1-12)"
// @Param   anio     query   int      false   "Filtrar por año (YYYY)"
// @Success 200 {object} models.ApiResponse{data=[]models.Nomina} "Lista de nóminas"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /nominas [get]
func (c *NominaController) GetAll() {
	o := ormNewNomina()
	var nominas []models.Nomina

	_, err := queryAllNominas(o, &nominas)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener nóminas de la base de datos",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	fecha := c.GetString("fecha")
	mes, _ := c.GetInt("mes")
	anio, _ := c.GetInt("anio")

	var filteredNominas []models.Nomina
	for _, nomina := range nominas {
		if fecha != "" && nomina.FECHA.Format("2006-01-02") != fecha {
			continue
		}
		if mes > 0 && mes <= 12 && int(nomina.FECHA.Month()) != mes {
			continue
		}
		if anio > 0 && nomina.FECHA.Year() != anio {
			continue
		}
		filteredNominas = append(filteredNominas, nomina)
	}

	if len(filteredNominas) == 0 {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "No se encontraron nóminas que coincidan con los filtros proporcionados",
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Nóminas obtenidas exitosamente",
		Data:    filteredNominas,
	}
	c.ServeJSON()
}

// @Title Post
// @Summary Crear una nueva nómina
// @Description Inserta un registro en la tabla "NOMINA"; el trigger genera automáticamente los cálculos.
// @Tags nominas
// @Accept json
// @Produce json
// @Param   body  body   models.NominaCreateRequest true  "Datos de la nómina a crear (sin 'MONTO'; fecha YYYY-MM-DD)"
// @Success 201 {object} models.ApiResponse{data=models.Nomina} "Nómina creada"
// @Failure 400 {object} models.ApiResponse "Error en la solicitud"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /nominas [post]
func (c *NominaController) Post() {
	o := ormNewNomina()
	var input models.Nomina

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
	if input.FECHA.IsZero() {
		now := time.Now()
		input.FECHA = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	}

	// Reglas de negocio
	// 1) No generar nómina antes del día 20 del mes
	if input.FECHA.Day() < 20 {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "No se puede generar una nómina antes del día 20 del mes",
		}
		c.ServeJSON()
		return
	}
	// 2) Si ya existe una nómina del mes, marcar control_nomina como REGENERADA y retornar la existente
	existing, getErr := findExistingNominaFn(o, input.FECHA)
	if getErr == nil && existing != nil && existing.PK_ID_NOMINA != 0 {
		// Upsert en control_nomina
		if _, err := o.Raw(
			"INSERT INTO control_nomina (fecha, estado) VALUES ($1, 'REGENERADA') ON CONFLICT (fecha) DO UPDATE SET estado = 'REGENERADA'",
			existing.FECHA,
		).Exec(); err != nil {
			c.Ctx.Output.SetStatus(http.StatusInternalServerError)
			c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al marcar nómina como REGENERADA", Cause: err.Error()}
			c.ServeJSON()
			return
		}
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{Code: http.StatusOK, Message: "Nómina ya existía; marcada como REGENERADA", Data: *existing}
		c.ServeJSON()
		return
	}
	if getErr != nil && getErr != orm.ErrNoRows {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al validar nóminas del mes", Cause: getErr.Error()}
		c.ServeJSON()
		return
	}

	if !estadosNominaPermitidos[input.ESTADO_NOMINA] {
		input.ESTADO_NOMINA = models.EstadoNominaNoPago
	}

	input.MONTO = 0 // lo calcula el trigger tras el insert

	if _, err := o.Insert(&input); err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al crear la nómina",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	var updatedNomina models.Nomina
	if err := o.QueryTable(new(models.Nomina)).
		Filter("PK_ID_NOMINA", input.PK_ID_NOMINA).
		One(&updatedNomina); err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al verificar la nómina generada",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusCreated,
		Message: "Nómina creada correctamente",
		Data:    updatedNomina,
	}
	c.ServeJSON()
}

// @Title Update
// @Summary Actualizar el estado de una nómina
// @Description Cambia el estado de una nómina existente a "PAGO".
// @Tags nominas
// @Accept json
// @Produce json
// @Param   id    query    int  true   "ID de la Nómina"
// @Success 200 {object} models.ApiResponse{data=models.Nomina} "Nómina actualizada"
// @Failure 404 {object} models.ApiResponse "Nómina no encontrada"
// @Failure 400 {object} models.ApiResponse "Error en la solicitud"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /nominas [put]
func (c *NominaController) Put() {
	o := ormNewNomina()

	// Obtener el ID
	idStr := c.GetString("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id == 0 {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El parámetro 'id' es inválido o está ausente",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Buscar la nómina
	nomina := models.Nomina{PK_ID_NOMINA: int64(id)}
	if err := readNominaFn(o, &nomina); err == orm.ErrNoRows {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Nómina no encontrada",
		}
		c.ServeJSON()
		return
	}

	// Cambiar el estado a "PAGO" si no lo está ya
	if nomina.ESTADO_NOMINA == models.EstadoNominaPago {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "La nómina ya está en estado 'PAGO'",
		}
		c.ServeJSON()
		return
	}
	nomina.ESTADO_NOMINA = models.EstadoNominaPago

	// Guardar los cambios
	if _, err := updateNominaFn(o, &nomina, "ESTADO_NOMINA"); err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al actualizar el estado de la nómina",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Responder con éxito
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Estado de la nómina actualizado a 'PAGO' correctamente",
		Data:    nomina,
	}
	c.ServeJSON()
}

// @Title Delete
// @Summary Eliminar una nómina (lógica)
// @Description Marca una nómina como "NO_PAGO" en lugar de eliminarla físicamente.
// @Tags nominas
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID de la Nómina"
// @Success 200 {object} models.ApiResponse "Nómina eliminada lógicamente"
// @Failure 404 {object} models.ApiResponse "Nómina no encontrada"
// @Security BearerAuth
// @Router /nominas [delete]
func (c *NominaController) Delete() {
	o := ormNewNomina()

	idStr := c.GetString("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id == 0 {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El parámetro 'id' es inválido o está ausente",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	nomina := models.Nomina{PK_ID_NOMINA: int64(id)}
	if err := readNominaFn(o, &nomina); err != nil {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Nómina no encontrada",
		}
		c.ServeJSON()
		return
	}

	nomina.ESTADO_NOMINA = models.EstadoNominaNoPago
	if _, err := updateNominaFn(o, &nomina, "ESTADO_NOMINA"); err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al eliminar lógicamente la nómina",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Nómina eliminada lógicamente",
	}
	c.ServeJSON()
}
