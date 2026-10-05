package nomina

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"restaurante/internal/dberr"
	"restaurante/internal/httpx"
	"restaurante/logging"
	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

// NominaController expone las nóminas. Las respuestas siempre llevan el
// envoltorio models.ApiResponse y el status HTTP coincide con su `code`.
type NominaController struct {
	web.Controller
}

const layoutFecha = "2006-01-02"

var estadosNominaPermitidos = map[models.EstadoNomina]bool{
	models.EstadoNominaPago:   true,
	models.EstadoNominaNoPago: true,
}

// bodyVacio indica si el cuerpo de la petición no trae contenido.
func bodyVacio(b []byte) bool { return len(bytes.TrimSpace(b)) == 0 }

// optionalInt lee un query param entero opcional dentro de [min, max].
func (c *NominaController) optionalInt(key string, min, max int) (int, error) {
	raw := c.GetString(key)
	if raw == "" {
		return 0, nil
	}
	v, err := c.GetInt(key)
	if err != nil || v < min || v > max {
		return 0, errors.New("el parámetro '" + key + "' debe ser un entero válido")
	}
	return v, nil
}

// @Title GetAll
// @Summary Obtener todas las nóminas con filtros
// @Description Devuelve las nóminas, opcionalmente filtradas por fecha exacta, mes y/o año (los filtros se combinan). Petición: `fecha` en YYYY-MM-DD; respuesta: `fechaNomina` en DD-MM-YYYY. `monto` lo calcula la base de datos. Sin resultados: 200 con `data: []`.
// @Tags nominas
// @Accept json
// @Produce json
// @Param   fecha    query   string   false   "Fecha exacta, formato YYYY-MM-DD"
// @Param   mes      query   int      false   "Mes (1-12)"
// @Param   anio     query   int      false   "Año (YYYY)"
// @Success 200 {object} models.ApiResponse{data=[]models.NominaResponse} "Lista de nóminas (puede ser [])"
// @Failure 400 {object} models.ApiResponse "fecha, mes o anio inválidos"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /nominas [get]
func (c *NominaController) GetAll() {
	fecha := c.GetString("fecha")
	if fecha != "" {
		if _, err := time.Parse(layoutFecha, fecha); err != nil {
			httpx.Fail(&c.Controller, http.StatusBadRequest, "El parámetro 'fecha' debe tener el formato YYYY-MM-DD", err)
			return
		}
	}
	mes, err := c.optionalInt("mes", 1, 12)
	if err != nil {
		httpx.Fail(&c.Controller, http.StatusBadRequest, "El parámetro 'mes' debe estar entre 1 y 12", err)
		return
	}
	anio, err := c.optionalInt("anio", 1, 9999)
	if err != nil {
		httpx.Fail(&c.Controller, http.StatusBadRequest, "El parámetro 'anio' es inválido", err)
		return
	}

	var nominas []models.Nomina
	if _, err := orm.NewOrm().QueryTable(new(models.Nomina)).All(&nominas); err != nil {
		logging.LogControllerError(c.Ctx, "nominas.getall.db_error", err, nil)
		httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error al obtener nóminas de la base de datos", err)
		return
	}

	filtradas := make([]models.Nomina, 0, len(nominas))
	for _, n := range nominas {
		f := n.FECHA.UTC()
		if fecha != "" && f.Format(layoutFecha) != fecha {
			continue
		}
		if mes > 0 && int(f.Month()) != mes {
			continue
		}
		if anio > 0 && f.Year() != anio {
			continue
		}
		filtradas = append(filtradas, n)
	}
	httpx.Send(&c.Controller, http.StatusOK, "Nóminas obtenidas exitosamente", filtradas)
}

// nominaDelMes busca una nómina existente en el mismo mes que fecha.
func nominaDelMes(o orm.Ormer, fecha time.Time) (*models.Nomina, error) {
	inicio := time.Date(fecha.Year(), fecha.Month(), 1, 0, 0, 0, 0, time.UTC)
	var existente models.Nomina
	err := o.QueryTable(new(models.Nomina)).
		Filter("FECHA__gte", inicio).
		Filter("FECHA__lt", inicio.AddDate(0, 1, 0)).
		One(&existente)
	if err != nil {
		return nil, err
	}
	return &existente, nil
}

// @Title Post
// @Summary Crear una nueva nómina
// @Description Inserta una nómina; el trigger de la base de datos calcula `monto` (el cliente no puede enviarlo; `nominaId` y `monto` en el cuerpo se ignoran). El cuerpo es opcional: por defecto fechaNomina es hoy y estadoNomina NO_PAGO. Petición: `fechaNomina` en YYYY-MM-DD (el día debe ser >= 20); respuesta: `fechaNomina` en DD-MM-YYYY. Si ya existe una nómina en ese mes no se crea otra: se marca el control como REGENERADA y se devuelve la existente con 200.
// @Tags nominas
// @Accept json
// @Produce json
// @Param   body  body   models.NominaCreateRequest false  "Datos de la nómina (opcional)"
// @Success 201 {object} models.ApiResponse{data=models.NominaResponse} "Nómina creada"
// @Success 200 {object} models.ApiResponse{data=models.NominaResponse} "Ya existía una nómina en el mes; marcada como REGENERADA"
// @Failure 400 {object} models.ApiResponse "JSON, fecha o estado inválidos, o día anterior al 20"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 409 {object} models.ApiResponse "Ya existe una nómina con esa fecha"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /nominas [post]
func (c *NominaController) Post() {
	var in models.NominaCreateRequest
	if body := c.Ctx.Input.RequestBody; !bodyVacio(body) {
		if err := json.Unmarshal(body, &in); err != nil {
			httpx.Fail(&c.Controller, http.StatusBadRequest, "Error al procesar la solicitud", err)
			return
		}
	}

	fecha := time.Now()
	fecha = time.Date(fecha.Year(), fecha.Month(), fecha.Day(), 12, 0, 0, 0, time.UTC)
	if in.FechaNomina != nil {
		parsed, err := models.ParseDateToNoonUTC(*in.FechaNomina)
		if err != nil {
			httpx.Fail(&c.Controller, http.StatusBadRequest, "El campo fechaNomina debe tener el formato YYYY-MM-DD", err)
			return
		}
		fecha = parsed
	}
	estado := models.EstadoNominaNoPago
	if in.EstadoNomina != nil {
		if !estadosNominaPermitidos[*in.EstadoNomina] {
			httpx.Fail(&c.Controller, http.StatusBadRequest, "El campo estadoNomina debe ser PAGO o NO_PAGO", nil)
			return
		}
		estado = *in.EstadoNomina
	}
	if fecha.Day() < 20 {
		httpx.Fail(&c.Controller, http.StatusBadRequest, "No se puede generar una nómina antes del día 20 del mes", nil)
		return
	}

	o := orm.NewOrm()
	existente, err := nominaDelMes(o, fecha)
	if err == nil {
		if _, err := o.Raw(
			"INSERT INTO control_nomina (fecha, estado) VALUES ($1, 'REGENERADA') ON CONFLICT (fecha) DO UPDATE SET estado = 'REGENERADA'",
			existente.FECHA,
		).Exec(); err != nil {
			logging.LogControllerError(c.Ctx, "nominas.post.control_nomina_error", err, nil)
			httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error al marcar nómina como REGENERADA", err)
			return
		}
		httpx.Send(&c.Controller, http.StatusOK, "Nómina ya existía; marcada como REGENERADA", existente)
		return
	}
	if !errors.Is(err, orm.ErrNoRows) {
		logging.LogControllerError(c.Ctx, "nominas.post.validate_month_error", err, nil)
		httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error al validar nóminas del mes", err)
		return
	}

	nueva := models.Nomina{FECHA: fecha, ESTADO_NOMINA: estado}
	id, err := o.Insert(&nueva)
	if err != nil {
		if dberr.IsUnique(err) {
			httpx.Fail(&c.Controller, http.StatusConflict, "Ya existe una nómina con esa fecha", err)
			return
		}
		logging.LogControllerError(c.Ctx, "nominas.post.insert_error", err, nil)
		httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error al crear la nómina", err)
		return
	}

	var creada models.Nomina
	if err := o.QueryTable(new(models.Nomina)).Filter("PK_ID_NOMINA", id).One(&creada); err != nil {
		logging.LogControllerError(c.Ctx, "nominas.post.verify_error", err, map[string]interface{}{"id": id})
		httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error al verificar la nómina generada", err)
		return
	}
	httpx.Send(&c.Controller, http.StatusCreated, "Nómina creada correctamente", creada)
}

// cargarNomina valida el query param id y carga la nómina (400/404/500 ya enviados).
func (c *NominaController) cargarNomina(o orm.Ormer, evento string) (*models.Nomina, bool) {
	id, err := httpx.PositiveInt64Param(&c.Controller, "id")
	if err != nil {
		httpx.Fail(&c.Controller, http.StatusBadRequest, "El parámetro 'id' es inválido o está ausente", err)
		return nil, false
	}
	nomina := models.Nomina{PK_ID_NOMINA: id}
	if err := o.Read(&nomina); err != nil {
		if errors.Is(err, orm.ErrNoRows) {
			httpx.Fail(&c.Controller, http.StatusNotFound, "Nómina no encontrada", err)
			return nil, false
		}
		logging.LogControllerError(c.Ctx, evento, err, map[string]interface{}{"id": id})
		httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error al obtener la nómina", err)
		return nil, false
	}
	return &nomina, true
}

// cambiarEstado persiste el nuevo estado: 409 si la nómina ya lo tenía.
func (c *NominaController) cambiarEstado(o orm.Ormer, nomina *models.Nomina, estado models.EstadoNomina, evento, mensaje string) {
	if nomina.ESTADO_NOMINA == estado {
		httpx.Fail(&c.Controller, http.StatusConflict, "La nómina ya está en estado '"+estado+"'", nil)
		return
	}
	nomina.ESTADO_NOMINA = estado
	if _, err := o.Update(nomina, "ESTADO_NOMINA"); err != nil {
		logging.LogControllerError(c.Ctx, evento, err, map[string]interface{}{"id": nomina.PK_ID_NOMINA})
		httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error al actualizar el estado de la nómina", err)
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, mensaje, nomina)
}

// @Title Update
// @Summary Actualizar el estado de una nómina
// @Description Cambia el estado de una nómina. El cuerpo es opcional: sin cuerpo (o sin `estadoNomina`) la nómina se marca PAGO. `estadoNomina` es el único campo editable y no admite null (400); `fechaNomina`, `monto` y `nominaId` se ignoran. Si la nómina ya tenía ese estado responde 409.
// @Tags nominas
// @Accept json
// @Produce json
// @Param   id    query    int  true   "ID de la nómina (entero positivo)"
// @Param   body  body   models.NominaUpdateRequest false  "Campos a modificar (opcional)"
// @Success 200 {object} models.ApiResponse{data=models.NominaResponse} "Nómina actualizada"
// @Failure 400 {object} models.ApiResponse "id, JSON, null o estado inválidos"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 404 {object} models.ApiResponse "Nómina no encontrada"
// @Failure 409 {object} models.ApiResponse "La nómina ya tenía ese estado"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /nominas [put]
func (c *NominaController) Put() {
	o := orm.NewOrm()
	nomina, ok := c.cargarNomina(o, "nominas.put.db_error")
	if !ok {
		return
	}

	estado := models.EstadoNominaPago
	if body := c.Ctx.Input.RequestBody; !bodyVacio(body) {
		var in models.NominaUpdateRequest
		if err := httpx.DecodeMerge(body, &in); err != nil {
			httpx.Fail(&c.Controller, http.StatusBadRequest, "Error al procesar la solicitud", err)
			return
		}
		if in.EstadoNomina != nil {
			if !estadosNominaPermitidos[*in.EstadoNomina] {
				httpx.Fail(&c.Controller, http.StatusBadRequest, "El campo estadoNomina debe ser PAGO o NO_PAGO", nil)
				return
			}
			estado = *in.EstadoNomina
		}
	}
	c.cambiarEstado(o, nomina, estado, "nominas.put.update_error", "Estado de la nómina actualizado a '"+estado+"' correctamente")
}

// @Title Delete
// @Summary Eliminar una nómina (lógica)
// @Description No borra la nómina: la marca como NO_PAGO y devuelve la nómina actualizada. Si ya estaba en NO_PAGO responde 409.
// @Tags nominas
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID de la nómina (entero positivo)"
// @Success 200 {object} models.ApiResponse{data=models.NominaResponse} "Nómina marcada como NO_PAGO"
// @Failure 400 {object} models.ApiResponse "id ausente o inválido"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 404 {object} models.ApiResponse "Nómina no encontrada"
// @Failure 409 {object} models.ApiResponse "La nómina ya estaba en NO_PAGO"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /nominas [delete]
func (c *NominaController) Delete() {
	o := orm.NewOrm()
	nomina, ok := c.cargarNomina(o, "nominas.delete.db_error")
	if !ok {
		return
	}
	c.cambiarEstado(o, nomina, models.EstadoNominaNoPago, "nominas.delete.update_error", "Nómina eliminada lógicamente")
}
