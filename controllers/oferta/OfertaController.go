package oferta

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"restaurante/logging"
	"restaurante/models"
	"restaurante/services"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

// Interfaces para testing
type ofertaQuerySeter interface {
	All(interface{}, ...string) (int64, error)
	Filter(string, ...interface{}) ofertaQuerySeter
	OrderBy(...string) ofertaQuerySeter
	Limit(int) ofertaQuerySeter
	Offset(int64) ofertaQuerySeter
	Count() (int64, error)
	One(interface{}) error
}

type ofertaOrmer interface {
	QueryTable(interface{}) ofertaQuerySeter
	Insert(interface{}) (int64, error)
	Read(interface{}, ...string) error
	Update(interface{}, ...string) (int64, error)
	Delete(interface{}, ...string) (int64, error)
}

type ofertQSAdapter struct{ qs orm.QuerySeter }

func (a ofertQSAdapter) All(res interface{}, cols ...string) (int64, error) {
	return a.qs.All(res, cols...)
}
func (a ofertQSAdapter) Filter(expr string, args ...interface{}) ofertaQuerySeter {
	return ofertQSAdapter{qs: a.qs.Filter(expr, args...)}
}
func (a ofertQSAdapter) OrderBy(exprs ...string) ofertaQuerySeter {
	return ofertQSAdapter{qs: a.qs.OrderBy(exprs...)}
}
func (a ofertQSAdapter) Limit(limit int) ofertaQuerySeter {
	return ofertQSAdapter{qs: a.qs.Limit(limit)}
}
func (a ofertQSAdapter) Offset(offset int64) ofertaQuerySeter {
	return ofertQSAdapter{qs: a.qs.Offset(offset)}
}
func (a ofertQSAdapter) Count() (int64, error) {
	return a.qs.Count()
}
func (a ofertQSAdapter) One(container interface{}) error {
	return a.qs.One(container)
}

type ofertOrmAdapter struct{ o orm.Ormer }

func (a ofertOrmAdapter) QueryTable(i interface{}) ofertaQuerySeter {
	return ofertQSAdapter{qs: a.o.QueryTable(i)}
}
func (a ofertOrmAdapter) Insert(v interface{}) (int64, error)      { return a.o.Insert(v) }
func (a ofertOrmAdapter) Read(v interface{}, cols ...string) error { return a.o.Read(v, cols...) }
func (a ofertOrmAdapter) Update(v interface{}, cols ...string) (int64, error) {
	return a.o.Update(v, cols...)
}
func (a ofertOrmAdapter) Delete(v interface{}, cols ...string) (int64, error) {
	return a.o.Delete(v, cols...)
}

var ofertOrmNew = func() ofertaOrmer { return ofertOrmAdapter{o: orm.NewOrm()} }

// Variable mockeable para crear el servicio
var newOfertaService = func(o orm.Ormer) *services.OfertaService {
	return services.NewOfertaService(o)
}

type OfertaController struct {
	web.Controller
}

// @Title GetAll
// @Summary Obtener todas las ofertas
// @Tags ofertas
// @Accept json
// @Produce json
// @Param activo query bool false "Filtrar por estado activo"
// @Param restaurante_id query int false "ID del restaurante"
// @Param titulo query string false "Filtrar por título"
// @Param limit query int false "Límite de resultados (default: 20)"
// @Param offset query int false "Offset para paginación (default: 0)"
// @Success 200 {object} models.ApiResponse{data=models.PaginatedResponse}
// @Failure 400 {object} models.ApiResponse
// @Failure 500 {object} models.ApiResponse
// @Router /ofertas [get]
func (c *OfertaController) GetAll() {
	o := ofertOrmNew()
	qs := o.QueryTable("oferta")

	// Aplicar filtros
	if activo := c.GetString("activo"); activo != "" {
		if activoBool, err := strconv.ParseBool(activo); err == nil {
			qs = qs.Filter("activo", activoBool)
		}
	}

	if restauranteId := c.GetString("restaurante_id"); restauranteId != "" {
		if id, err := strconv.ParseInt(restauranteId, 10, 64); err == nil {
			qs = qs.Filter("pk_id_restaurante", id)
		}
	}

	if titulo := c.GetString("titulo"); titulo != "" {
		qs = qs.Filter("titulo__icontains", titulo)
	}

	// Paginación
	limit, _ := c.GetInt("limit", 20)
	offset, _ := c.GetInt("offset", 0)

	if limit > 100 {
		limit = 100
	}

	// Contar total
	total, err := qs.Count()
	if err != nil {
		logging.LogControllerError(c.Ctx, "ofertas.getall.count_error", err, nil)
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener ofertas",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	// Obtener datos
	var ofertas []*models.Oferta
	_, err = qs.OrderBy("-pk_id_oferta").Limit(limit).Offset(int64(offset)).All(&ofertas)
	if err != nil {
		logging.LogControllerError(c.Ctx, "ofertas.getall.query_error", err, nil)
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener ofertas",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))
	page := (offset / limit) + 1

	response := models.PaginatedResponse{
		Data:       ofertas,
		Total:      total,
		Page:       page,
		PageSize:   limit,
		TotalPages: totalPages,
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Ofertas obtenidas exitosamente",
		Data:    response,
	}
	_ = c.ServeJSON()
}

// @Title Post
// @Summary Crear oferta
// @Tags ofertas
// @Accept json
// @Produce json
// @Param body body models.CrearOfertaRequest true "Datos de la oferta"
// @Success 201 {object} models.ApiResponse{data=models.Oferta}
// @Failure 400 {object} models.ApiResponse
// @Failure 422 {object} models.ApiResponse
// @Failure 500 {object} models.ApiResponse
// @Router /ofertas [post]
func (c *OfertaController) Post() {
	var req models.CrearOfertaRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		logging.LogControllerError(c.Ctx, "ofertas.post.bad_json", err, nil)
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "JSON inválido",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	// Validar tipo de descuento
	if !req.TipoDescuento.IsValid() {
		logging.LogControllerError(c.Ctx, "ofertas.post.invalid_tipo_descuento", nil, map[string]interface{}{"tipoDescuento": req.TipoDescuento})
		c.Ctx.Output.SetStatus(http.StatusUnprocessableEntity)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusUnprocessableEntity,
			Message: "Tipo de descuento no válido - debe ser PORCENTAJE o MONTO",
		}
		_ = c.ServeJSON()
		return
	}

	// Parsear fechas
	fechaInicio, err := time.Parse("2006-01-02", req.FechaInicio)
	if err != nil {
		logging.LogControllerError(c.Ctx, "ofertas.post.invalid_fecha_inicio", err, map[string]interface{}{"fechaInicio": req.FechaInicio})
		c.Ctx.Output.SetStatus(http.StatusUnprocessableEntity)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusUnprocessableEntity,
			Message: "Fecha de inicio inválida - debe tener formato YYYY-MM-DD",
		}
		_ = c.ServeJSON()
		return
	}

	fechaFin, err := time.Parse("2006-01-02", req.FechaFin)
	if err != nil {
		logging.LogControllerError(c.Ctx, "ofertas.post.invalid_fecha_fin", err, map[string]interface{}{"fechaFin": req.FechaFin})
		c.Ctx.Output.SetStatus(http.StatusUnprocessableEntity)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusUnprocessableEntity,
			Message: "Fecha de fin inválida - debe tener formato YYYY-MM-DD",
		}
		_ = c.ServeJSON()
		return
	}

	// Parsear horarios si están especificados
	var horaInicio, horaFin *time.Time
	if req.HoraInicio != nil {
		if hora, err := time.Parse("15:04", *req.HoraInicio); err == nil {
			horaInicio = &hora
		} else {
			logging.LogControllerError(c.Ctx, "ofertas.post.invalid_hora_inicio", err, map[string]interface{}{"horaInicio": *req.HoraInicio})
			c.Ctx.Output.SetStatus(http.StatusUnprocessableEntity)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusUnprocessableEntity,
				Message: "Hora de inicio inválida - debe tener formato HH:MM",
			}
			_ = c.ServeJSON()
			return
		}
	}

	if req.HoraFin != nil {
		if hora, err := time.Parse("15:04", *req.HoraFin); err == nil {
			horaFin = &hora
		} else {
			logging.LogControllerError(c.Ctx, "ofertas.post.invalid_hora_fin", err, map[string]interface{}{"horaFin": *req.HoraFin})
			c.Ctx.Output.SetStatus(http.StatusUnprocessableEntity)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusUnprocessableEntity,
				Message: "Hora de fin inválida - debe tener formato HH:MM",
			}
			_ = c.ServeJSON()
			return
		}
	}

	// Crear modelo
	oferta := &models.Oferta{
		Titulo:          req.Titulo,
		TipoDescuento:   req.TipoDescuento,
		ValorDescuento:  req.ValorDescuento,
		FechaInicio:     fechaInicio,
		FechaFin:        fechaFin,
		DiasSemanaArray: req.DiasSemana,
		HoraInicio:      horaInicio,
		HoraFin:         horaFin,
		Activo:          true,
		PkIdRestaurante: &models.Restaurante{PK_ID_RESTAURANTE: req.PkIdRestaurante},
	}

	// Validar reglas de negocio
	o := ofertOrmNew()
	ofertaService := newOfertaService(nil) // ValidarReglasNegocioOferta no usa el ORM

	if err := ofertaService.ValidarReglasNegocioOferta(oferta); err != nil {
		logging.LogControllerError(c.Ctx, "ofertas.post.validation_error", err, map[string]interface{}{"titulo": req.Titulo})
		c.Ctx.Output.SetStatus(http.StatusUnprocessableEntity)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusUnprocessableEntity,
			Message: "Error de validación",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	// Insertar en base de datos
	_, err = o.Insert(oferta)
	if err != nil {
		logging.LogControllerError(c.Ctx, "ofertas.post.insert_error", err, map[string]interface{}{"titulo": req.Titulo})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al crear oferta",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusCreated,
		Message: "Oferta creada exitosamente",
		Data:    oferta,
	}
	_ = c.ServeJSON()
}

// @Title GetById
// @Summary Obtener oferta por ID
// @Tags ofertas
// @Accept json
// @Produce json
// @Param id query int true "ID de la oferta"
// @Success 200 {object} models.ApiResponse{data=models.Oferta}
// @Failure 404 {object} models.ApiResponse
// @Router /ofertas/search [get]
func (c *OfertaController) GetById() {
	id, err := c.GetInt64("id")
	if err != nil || id == 0 {
		logging.LogControllerError(c.Ctx, "ofertas.getbyid.bad_request", err, map[string]interface{}{"id": c.GetString("id")})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "ID inválido o ausente",
		}
		_ = c.ServeJSON()
		return
	}

	o := ofertOrmNew()
	oferta := &models.Oferta{PkIdOferta: id}
	err = o.Read(oferta)
	if err != nil {
		if err == orm.ErrNoRows {
			c.Ctx.Output.SetStatus(http.StatusNotFound)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusNotFound,
				Message: "Oferta no encontrada",
			}
			_ = c.ServeJSON()
			return
		}
		logging.LogControllerError(c.Ctx, "ofertas.getbyid.read_error", err, map[string]interface{}{"id": id})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error interno del servidor",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Oferta encontrada",
		Data:    oferta,
	}
	_ = c.ServeJSON()
}

// @Title Put
// @Summary Actualizar oferta
// @Tags ofertas
// @Accept json
// @Produce json
// @Param id query int true "ID de la oferta"
// @Param body body models.CrearOfertaRequest true "Datos actualizados de la oferta"
// @Success 200 {object} models.ApiResponse{data=models.Oferta}
// @Failure 400 {object} models.ApiResponse
// @Failure 404 {object} models.ApiResponse
// @Failure 422 {object} models.ApiResponse
// @Router /ofertas [put]
func (c *OfertaController) Put() {
	id, err := c.GetInt64("id")
	if err != nil || id == 0 {
		logging.LogControllerError(c.Ctx, "ofertas.put.bad_request", err, map[string]interface{}{"id": c.GetString("id")})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "ID inválido o ausente",
		}
		_ = c.ServeJSON()
		return
	}

	var req models.CrearOfertaRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		logging.LogControllerError(c.Ctx, "ofertas.put.bad_json", err, map[string]interface{}{"id": id})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "JSON inválido",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	o := ofertOrmNew()

	// Verificar que la oferta existe
	oferta := &models.Oferta{PkIdOferta: id}
	err = o.Read(oferta)
	if err != nil {
		if err == orm.ErrNoRows {
			c.Ctx.Output.SetStatus(http.StatusNotFound)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusNotFound,
				Message: "Oferta no encontrada",
			}
			_ = c.ServeJSON()
			return
		}
		logging.LogControllerError(c.Ctx, "ofertas.put.read_error", err, map[string]interface{}{"id": id})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error interno del servidor",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	// Validar tipo de descuento
	if !req.TipoDescuento.IsValid() {
		logging.LogControllerError(c.Ctx, "ofertas.put.invalid_tipo_descuento", nil, map[string]interface{}{"tipoDescuento": req.TipoDescuento, "id": id})
		c.Ctx.Output.SetStatus(http.StatusUnprocessableEntity)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusUnprocessableEntity,
			Message: "Tipo de descuento no válido - debe ser PORCENTAJE o MONTO",
		}
		_ = c.ServeJSON()
		return
	}

	// Parsear fechas
	fechaInicio, err := time.Parse("2006-01-02", req.FechaInicio)
	if err != nil {
		logging.LogControllerError(c.Ctx, "ofertas.put.invalid_fecha_inicio", err, map[string]interface{}{"fechaInicio": req.FechaInicio, "id": id})
		c.Ctx.Output.SetStatus(http.StatusUnprocessableEntity)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusUnprocessableEntity,
			Message: "Fecha de inicio inválida - debe tener formato YYYY-MM-DD",
		}
		_ = c.ServeJSON()
		return
	}

	fechaFin, err := time.Parse("2006-01-02", req.FechaFin)
	if err != nil {
		logging.LogControllerError(c.Ctx, "ofertas.put.invalid_fecha_fin", err, map[string]interface{}{"fechaFin": req.FechaFin, "id": id})
		c.Ctx.Output.SetStatus(http.StatusUnprocessableEntity)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusUnprocessableEntity,
			Message: "Fecha de fin inválida - debe tener formato YYYY-MM-DD",
		}
		_ = c.ServeJSON()
		return
	}

	// Parsear horarios si están especificados
	var horaInicio, horaFin *time.Time
	if req.HoraInicio != nil {
		if hora, err := time.Parse("15:04", *req.HoraInicio); err == nil {
			horaInicio = &hora
		} else {
			logging.LogControllerError(c.Ctx, "ofertas.put.invalid_hora_inicio", err, map[string]interface{}{"horaInicio": *req.HoraInicio, "id": id})
			c.Ctx.Output.SetStatus(http.StatusUnprocessableEntity)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusUnprocessableEntity,
				Message: "Hora de inicio inválida - debe tener formato HH:MM",
			}
			_ = c.ServeJSON()
			return
		}
	}

	if req.HoraFin != nil {
		if hora, err := time.Parse("15:04", *req.HoraFin); err == nil {
			horaFin = &hora
		} else {
			logging.LogControllerError(c.Ctx, "ofertas.put.invalid_hora_fin", err, map[string]interface{}{"horaFin": *req.HoraFin, "id": id})
			c.Ctx.Output.SetStatus(http.StatusUnprocessableEntity)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusUnprocessableEntity,
				Message: "Hora de fin inválida - debe tener formato HH:MM",
			}
			_ = c.ServeJSON()
			return
		}
	}

	// Actualizar campos
	oferta.Titulo = req.Titulo
	oferta.TipoDescuento = req.TipoDescuento
	oferta.ValorDescuento = req.ValorDescuento
	oferta.FechaInicio = fechaInicio
	oferta.FechaFin = fechaFin
	oferta.DiasSemanaArray = req.DiasSemana
	oferta.HoraInicio = horaInicio
	oferta.HoraFin = horaFin
	oferta.PkIdRestaurante = &models.Restaurante{PK_ID_RESTAURANTE: req.PkIdRestaurante}

	// Validar reglas de negocio
	ofertaService := newOfertaService(nil) // ValidarReglasNegocioOferta no usa el ORM
	if err := ofertaService.ValidarReglasNegocioOferta(oferta); err != nil {
		logging.LogControllerError(c.Ctx, "ofertas.put.validation_error", err, map[string]interface{}{"titulo": req.Titulo, "id": id})
		c.Ctx.Output.SetStatus(http.StatusUnprocessableEntity)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusUnprocessableEntity,
			Message: "Error de validación",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	// Actualizar en base de datos
	_, err = o.Update(oferta)
	if err != nil {
		logging.LogControllerError(c.Ctx, "ofertas.put.update_error", err, map[string]interface{}{"id": id})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al actualizar oferta",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Oferta actualizada exitosamente",
		Data:    oferta,
	}
	_ = c.ServeJSON()
}

// @Title Delete
// @Summary Desactivar oferta
// @Tags ofertas
// @Accept json
// @Produce json
// @Param id query int true "ID de la oferta"
// @Success 200 {object} models.ApiResponse
// @Failure 400 {object} models.ApiResponse
// @Failure 404 {object} models.ApiResponse
// @Router /ofertas [delete]
func (c *OfertaController) Delete() {
	id, err := c.GetInt64("id")
	if err != nil || id == 0 {
		logging.LogControllerError(c.Ctx, "ofertas.delete.bad_request", err, map[string]interface{}{"id": c.GetString("id")})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "ID inválido o ausente",
		}
		_ = c.ServeJSON()
		return
	}

	o := ofertOrmNew()
	oferta := &models.Oferta{PkIdOferta: id}
	err = o.Read(oferta)
	if err != nil {
		if err == orm.ErrNoRows {
			c.Ctx.Output.SetStatus(http.StatusNotFound)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusNotFound,
				Message: "Oferta no encontrada",
			}
			_ = c.ServeJSON()
			return
		}
		logging.LogControllerError(c.Ctx, "ofertas.delete.read_error", err, map[string]interface{}{"id": id})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error interno del servidor",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	if !oferta.Activo {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "La oferta ya está desactivada",
		}
		_ = c.ServeJSON()
		return
	}

	// Desactivar oferta (borrado lógico)
	oferta.Activo = false
	_, err = o.Update(oferta, "Activo")
	if err != nil {
		logging.LogControllerError(c.Ctx, "ofertas.delete.update_error", err, map[string]interface{}{"id": id})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al desactivar oferta",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Oferta desactivada exitosamente",
	}
	_ = c.ServeJSON()
}

// Métodos adicionales específicos de ofertas

// @Title ObtenerOfertasActivas
// @Summary Obtener ofertas activas
// @Tags ofertas
// @Accept json
// @Produce json
// @Param restaurante_id query int true "ID del restaurante"
// @Param fecha query string false "Fecha a consultar (YYYY-MM-DD, default: hoy)"
// @Param hora query string false "Hora a consultar (HH:MM, default: ahora)"
// @Param producto_id query int false "ID del producto específico"
// @Success 200 {object} models.ApiResponse{data=[]models.OfertaActivaResponse}
// @Failure 400 {object} models.ApiResponse
// @Failure 500 {object} models.ApiResponse
// @Router /ofertas/activas [get]
func (c *OfertaController) ObtenerOfertasActivas() {
	restauranteId, err := c.GetInt64("restaurante_id")
	if err != nil || restauranteId == 0 {
		logging.LogControllerError(c.Ctx, "ofertas.activas.bad_request", err, map[string]interface{}{"restaurante_id": c.GetString("restaurante_id")})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "restaurante_id es requerido y debe ser un número entero válido",
		}
		_ = c.ServeJSON()
		return
	}

	var fecha *time.Time
	if fechaStr := c.GetString("fecha"); fechaStr != "" {
		if f, err := time.Parse("2006-01-02", fechaStr); err == nil {
			fecha = &f
		} else {
			logging.LogControllerError(c.Ctx, "ofertas.activas.invalid_fecha", err, map[string]interface{}{"fecha": fechaStr})
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusBadRequest,
				Message: "Fecha inválida - debe tener formato YYYY-MM-DD",
			}
			_ = c.ServeJSON()
			return
		}
	}

	var hora *time.Time
	if horaStr := c.GetString("hora"); horaStr != "" {
		if h, err := time.Parse("15:04", horaStr); err == nil {
			hora = &h
		} else {
			logging.LogControllerError(c.Ctx, "ofertas.activas.invalid_hora", err, map[string]interface{}{"hora": horaStr})
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusBadRequest,
				Message: "Hora inválida - debe tener formato HH:MM",
			}
			_ = c.ServeJSON()
			return
		}
	}

	var productoId *int64
	if productoIdStr := c.GetString("producto_id"); productoIdStr != "" {
		if id, err := strconv.ParseInt(productoIdStr, 10, 64); err == nil {
			productoId = &id
		}
	}

	ofertaService := newOfertaService(orm.NewOrm())
	ofertas, err := ofertaService.ObtenerOfertasActivas(c.Ctx.Request.Context(), restauranteId, fecha, hora, productoId)
	if err != nil {
		logging.LogControllerError(c.Ctx, "ofertas.activas.service_error", err, map[string]interface{}{"restaurante_id": restauranteId})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener ofertas activas",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Ofertas activas obtenidas exitosamente",
		Data:    ofertas,
	}
	_ = c.ServeJSON()
}

// @Title AsociarProducto
// @Summary Asociar producto a oferta
// @Tags ofertas
// @Accept json
// @Produce json
// @Param id query int true "ID de la oferta"
// @Param body body models.AsociarProductoOfertaRequest true "ID del producto"
// @Success 201 {object} models.ApiResponse
// @Failure 400 {object} models.ApiResponse
// @Failure 404 {object} models.ApiResponse
// @Failure 409 {object} models.ApiResponse
// @Router /ofertas/productos [post]
func (c *OfertaController) AsociarProducto() {
	ofertaId, err := c.GetInt64("id")
	if err != nil || ofertaId == 0 {
		logging.LogControllerError(c.Ctx, "ofertas.asociar.bad_request", err, map[string]interface{}{"id": c.GetString("id")})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "ID inválido o ausente",
		}
		_ = c.ServeJSON()
		return
	}

	var req models.AsociarProductoOfertaRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		logging.LogControllerError(c.Ctx, "ofertas.asociar.bad_json", err, map[string]interface{}{"id": ofertaId})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "JSON inválido",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	o := ofertOrmNew()

	// Verificar que la oferta existe
	oferta := &models.Oferta{PkIdOferta: ofertaId}
	err = o.Read(oferta)
	if err != nil {
		if err == orm.ErrNoRows {
			c.Ctx.Output.SetStatus(http.StatusOK)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusNotFound,
				Message: "Oferta no encontrada",
			}
			_ = c.ServeJSON()
			return
		}
		logging.LogControllerError(c.Ctx, "ofertas.asociar.read_oferta_error", err, map[string]interface{}{"id": ofertaId})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error interno del servidor",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	// Verificar que el producto existe
	producto := &models.Producto{PK_ID_PRODUCTO: req.ProductoId}
	err = o.Read(producto)
	if err != nil {
		if err == orm.ErrNoRows {
			c.Ctx.Output.SetStatus(http.StatusOK)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusNotFound,
				Message: "Producto no encontrado",
			}
			_ = c.ServeJSON()
			return
		}
		logging.LogControllerError(c.Ctx, "ofertas.asociar.read_producto_error", err, map[string]interface{}{"producto_id": req.ProductoId})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error interno del servidor",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	// Crear la asociación
	ofertaProducto := &models.OfertaProducto{
		PkIdOferta:   oferta,
		PkIdProducto: producto,
	}

	_, err = o.Insert(ofertaProducto)
	if err != nil {
		// Si ya existe la asociación, devolver conflicto
		if fmt.Sprintf("%v", err) == "UNIQUE constraint failed" ||
			fmt.Sprintf("%v", err) == "duplicate key value violates unique constraint" {
			c.Ctx.Output.SetStatus(http.StatusConflict)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusConflict,
				Message: "El producto ya está asociado a esta oferta",
			}
		} else {
			logging.LogControllerError(c.Ctx, "ofertas.asociar.insert_error", err, map[string]interface{}{"oferta_id": ofertaId, "producto_id": req.ProductoId})
			c.Ctx.Output.SetStatus(http.StatusInternalServerError)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusInternalServerError,
				Message: "Error al asociar producto",
				Cause:   err.Error(),
			}
		}
		_ = c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusCreated,
		Message: "Producto asociado correctamente",
	}
	_ = c.ServeJSON()
}

// @Title DesasociarProducto
// @Summary Desasociar producto de oferta
// @Tags ofertas
// @Accept json
// @Produce json
// @Param id query int true "ID de la oferta"
// @Param producto_id query int true "ID del producto"
// @Success 200 {object} models.ApiResponse
// @Failure 400 {object} models.ApiResponse
// @Failure 404 {object} models.ApiResponse
// @Router /ofertas/productos [delete]
func (c *OfertaController) DesasociarProducto() {
	ofertaId, err := c.GetInt64("id")
	if err != nil || ofertaId == 0 {
		logging.LogControllerError(c.Ctx, "ofertas.desasociar.bad_oferta_id", err, map[string]interface{}{"id": c.GetString("id")})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "ID de oferta inválido o ausente",
		}
		_ = c.ServeJSON()
		return
	}

	productoId, err := c.GetInt64("producto_id")
	if err != nil || productoId == 0 {
		logging.LogControllerError(c.Ctx, "ofertas.desasociar.bad_producto_id", err, map[string]interface{}{"producto_id": c.GetString("producto_id")})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "ID de producto inválido o ausente",
		}
		_ = c.ServeJSON()
		return
	}

	o := ofertOrmNew()

	// Buscar la asociación
	ofertaProducto := &models.OfertaProducto{}
	err = o.QueryTable("oferta_producto").
		Filter("pk_id_oferta", ofertaId).
		Filter("pk_id_producto", productoId).
		One(ofertaProducto)

	if err != nil {
		if err == orm.ErrNoRows {
			c.Ctx.Output.SetStatus(http.StatusOK)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusNotFound,
				Message: "Asociación no encontrada",
			}
			_ = c.ServeJSON()
			return
		}
		logging.LogControllerError(c.Ctx, "ofertas.desasociar.query_error", err, map[string]interface{}{"oferta_id": ofertaId, "producto_id": productoId})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error interno del servidor",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	// Eliminar la asociación
	_, err = o.Delete(ofertaProducto)
	if err != nil {
		logging.LogControllerError(c.Ctx, "ofertas.desasociar.delete_error", err, map[string]interface{}{"oferta_id": ofertaId, "producto_id": productoId})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al desasociar producto",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Producto desasociado correctamente",
	}
	_ = c.ServeJSON()
}
