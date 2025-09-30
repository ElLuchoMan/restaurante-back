package push

import (
	"encoding/json"
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
type pushQuerySeter interface {
	All(interface{}, ...string) (int64, error)
	Filter(string, ...interface{}) pushQuerySeter
	OrderBy(...string) pushQuerySeter
	Limit(int) pushQuerySeter
	Offset(int64) pushQuerySeter
	Count() (int64, error)
	One(interface{}) error
}

type pushOrmer interface {
	QueryTable(interface{}) pushQuerySeter
	Insert(interface{}) (int64, error)
	Read(interface{}, ...string) error
	Update(interface{}, ...string) (int64, error)
	Delete(interface{}, ...string) (int64, error)
}

type pushQSAdapter struct{ qs orm.QuerySeter }

func (a pushQSAdapter) All(res interface{}, cols ...string) (int64, error) {
	return a.qs.All(res, cols...)
}
func (a pushQSAdapter) Filter(expr string, args ...interface{}) pushQuerySeter {
	return pushQSAdapter{qs: a.qs.Filter(expr, args...)}
}
func (a pushQSAdapter) OrderBy(exprs ...string) pushQuerySeter {
	return pushQSAdapter{qs: a.qs.OrderBy(exprs...)}
}
func (a pushQSAdapter) Limit(limit int) pushQuerySeter {
	return pushQSAdapter{qs: a.qs.Limit(limit)}
}
func (a pushQSAdapter) Offset(offset int64) pushQuerySeter {
	return pushQSAdapter{qs: a.qs.Offset(offset)}
}
func (a pushQSAdapter) Count() (int64, error) {
	return a.qs.Count()
}
func (a pushQSAdapter) One(container interface{}) error {
	return a.qs.One(container)
}

type pushOrmAdapter struct{ o orm.Ormer }

func (a pushOrmAdapter) QueryTable(i interface{}) pushQuerySeter {
	return pushQSAdapter{qs: a.o.QueryTable(i)}
}
func (a pushOrmAdapter) Insert(v interface{}) (int64, error)      { return a.o.Insert(v) }
func (a pushOrmAdapter) Read(v interface{}, cols ...string) error { return a.o.Read(v, cols...) }
func (a pushOrmAdapter) Update(v interface{}, cols ...string) (int64, error) {
	return a.o.Update(v, cols...)
}
func (a pushOrmAdapter) Delete(v interface{}, cols ...string) (int64, error) {
	return a.o.Delete(v, cols...)
}

var pushOrmNew = func() pushOrmer { return pushOrmAdapter{o: orm.NewOrm()} }

type PushController struct {
	web.Controller
}

// @Title GetAll
// @Summary Obtener dispositivos push
// @Tags push_notifications
// @Accept json
// @Produce json
// @Param cliente_id query int false "ID del cliente"
// @Param trabajador_id query int false "ID del trabajador"
// @Param plataforma query string false "Plataforma (WEB, ANDROID, IOS)"
// @Param limit query int false "Límite de resultados (default: 20)"
// @Param offset query int false "Offset para paginación (default: 0)"
// @Success 200 {object} models.ApiResponse{data=models.PaginatedResponse}
// @Failure 400 {object} models.ApiResponse
// @Failure 500 {object} models.ApiResponse
// @Router /push/dispositivos [get]
func (c *PushController) GetAll() {
	o := pushOrmNew()
	qs := o.QueryTable("push_dispositivo")

	// Aplicar filtros
	if clienteId := c.GetString("cliente_id"); clienteId != "" {
		if id, err := strconv.ParseInt(clienteId, 10, 64); err == nil {
			qs = qs.Filter("pk_documento_cliente", id)
		}
	}

	if trabajadorId := c.GetString("trabajador_id"); trabajadorId != "" {
		if id, err := strconv.ParseInt(trabajadorId, 10, 64); err == nil {
			qs = qs.Filter("pk_documento_trabajador", id)
		}
	}

	if plataforma := c.GetString("plataforma"); plataforma != "" {
		qs = qs.Filter("plataforma", plataforma)
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
		logging.LogControllerError(c.Ctx, "push.getall.count_error", err, nil)
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener dispositivos",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	// Obtener datos
	var dispositivos []*models.PushDispositivo
	_, err = qs.OrderBy("-created_at").Limit(limit).Offset(int64(offset)).All(&dispositivos)
	if err != nil {
		logging.LogControllerError(c.Ctx, "push.getall.query_error", err, nil)
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener dispositivos",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))
	page := (offset / limit) + 1

	response := models.PaginatedResponse{
		Data:       dispositivos,
		Total:      total,
		Page:       page,
		PageSize:   limit,
		TotalPages: totalPages,
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Dispositivos obtenidos exitosamente",
		Data:    response,
	}
	_ = c.ServeJSON()
}

// @Title Post
// @Summary Registrar dispositivo push
// @Tags push_notifications
// @Accept json
// @Produce json
// @Param body body models.RegistrarDispositivoRequest true "Datos del dispositivo"
// @Success 201 {object} models.ApiResponse{data=models.PushDispositivo}
// @Failure 400 {object} models.ApiResponse
// @Failure 422 {object} models.ApiResponse
// @Failure 500 {object} models.ApiResponse
// @Router /push/dispositivos [post]
func (c *PushController) Post() {
	var req models.RegistrarDispositivoRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		logging.LogControllerError(c.Ctx, "push.post.bad_json", err, nil)
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "JSON inválido",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	// Validar plataforma
	if !req.Plataforma.IsValid() {
		logging.LogControllerError(c.Ctx, "push.post.invalid_plataforma", nil, map[string]interface{}{"plataforma": req.Plataforma})
		c.Ctx.Output.SetStatus(http.StatusUnprocessableEntity)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusUnprocessableEntity,
			Message: "Plataforma no válida - debe ser WEB, ANDROID o IOS",
		}
		_ = c.ServeJSON()
		return
	}

	pushService := services.NewPushService(orm.NewOrm())
	dispositivo, err := pushService.RegistrarDispositivo(c.Ctx.Request.Context(), &req)
	if err != nil {
		logging.LogControllerError(c.Ctx, "push.post.service_error", err, map[string]interface{}{"dispositivo": "registro"})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al registrar dispositivo",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusCreated,
		Message: "Dispositivo registrado exitosamente",
		Data:    dispositivo,
	}
	_ = c.ServeJSON()
}

// @Title GetById
// @Summary Obtener dispositivo por ID
// @Tags push_notifications
// @Accept json
// @Produce json
// @Param id query int true "ID del dispositivo"
// @Success 200 {object} models.ApiResponse{data=models.PushDispositivo}
// @Failure 404 {object} models.ApiResponse
// @Router /push/dispositivos/search [get]
func (c *PushController) GetById() {
	id, err := c.GetInt64("id")
	if err != nil || id == 0 {
		logging.LogControllerError(c.Ctx, "push.getbyid.bad_request", err, map[string]interface{}{"id": c.GetString("id")})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "ID inválido o ausente",
		}
		_ = c.ServeJSON()
		return
	}

	o := pushOrmNew()
	dispositivo := &models.PushDispositivo{PkIdPushDispositivo: id}
	err = o.Read(dispositivo)
	if err != nil {
		if err == orm.ErrNoRows {
			c.Ctx.Output.SetStatus(http.StatusOK)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusNotFound,
				Message: "Dispositivo no encontrado",
			}
			_ = c.ServeJSON()
			return
		}
		logging.LogControllerError(c.Ctx, "push.getbyid.read_error", err, map[string]interface{}{"id": id})
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
		Message: "Dispositivo encontrado",
		Data:    dispositivo,
	}
	_ = c.ServeJSON()
}

// @Title Put
// @Summary Actualizar dispositivo push
// @Tags push_notifications
// @Accept json
// @Produce json
// @Param id query int true "ID del dispositivo"
// @Param body body models.ActualizarEstadoDispositivoRequest true "Nuevo estado"
// @Success 200 {object} models.ApiResponse
// @Failure 400 {object} models.ApiResponse
// @Failure 404 {object} models.ApiResponse
// @Router /push/dispositivos [put]
func (c *PushController) Put() {
	id, err := c.GetInt64("id")
	if err != nil || id == 0 {
		logging.LogControllerError(c.Ctx, "push.put.bad_request", err, map[string]interface{}{"id": c.GetString("id")})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "ID inválido o ausente",
		}
		_ = c.ServeJSON()
		return
	}

	var req models.ActualizarEstadoDispositivoRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		logging.LogControllerError(c.Ctx, "push.put.bad_json", err, map[string]interface{}{"id": id})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "JSON inválido",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	pushService := services.NewPushService(orm.NewOrm())
	err = pushService.ActualizarEstadoDispositivo(c.Ctx.Request.Context(), id, req.Enabled)
	if err != nil {
		logging.LogControllerError(c.Ctx, "push.put.service_error", err, map[string]interface{}{"id": id})
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Dispositivo no encontrado",
		}
		_ = c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Estado actualizado correctamente",
	}
	_ = c.ServeJSON()
}

// @Title Delete
// @Summary Eliminar dispositivo push
// @Tags push_notifications
// @Accept json
// @Produce json
// @Param id query int true "ID del dispositivo"
// @Success 200 {object} models.ApiResponse
// @Failure 400 {object} models.ApiResponse
// @Failure 404 {object} models.ApiResponse
// @Router /push/dispositivos [delete]
func (c *PushController) Delete() {
	id, err := c.GetInt64("id")
	if err != nil || id == 0 {
		logging.LogControllerError(c.Ctx, "push.delete.bad_request", err, map[string]interface{}{"id": c.GetString("id")})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "ID inválido o ausente",
		}
		_ = c.ServeJSON()
		return
	}

	o := pushOrmNew()
	dispositivo := &models.PushDispositivo{PkIdPushDispositivo: id}
	err = o.Read(dispositivo)
	if err != nil {
		if err == orm.ErrNoRows {
			c.Ctx.Output.SetStatus(http.StatusOK)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusNotFound,
				Message: "Dispositivo no encontrado",
			}
			_ = c.ServeJSON()
			return
		}
		logging.LogControllerError(c.Ctx, "push.delete.read_error", err, map[string]interface{}{"id": id})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error interno del servidor",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	// Eliminar dispositivo
	_, err = o.Delete(dispositivo)
	if err != nil {
		logging.LogControllerError(c.Ctx, "push.delete.delete_error", err, map[string]interface{}{"id": id})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al eliminar dispositivo",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Dispositivo eliminado exitosamente",
	}
	_ = c.ServeJSON()
}

// Métodos adicionales específicos de push

// @Title ActualizarUltimaVista
// @Summary Actualizar última vista del dispositivo
// @Tags push_notifications
// @Accept json
// @Produce json
// @Param id query int true "ID del dispositivo"
// @Success 200 {object} models.ApiResponse
// @Failure 400 {object} models.ApiResponse
// @Failure 404 {object} models.ApiResponse
// @Router /push/dispositivos/visto [patch]
func (c *PushController) ActualizarUltimaVista() {
	id, err := c.GetInt64("id")
	if err != nil || id == 0 {
		logging.LogControllerError(c.Ctx, "push.visto.bad_request", err, map[string]interface{}{"id": c.GetString("id")})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "ID inválido o ausente",
		}
		_ = c.ServeJSON()
		return
	}

	pushService := services.NewPushService(orm.NewOrm())
	err = pushService.ActualizarUltimaVista(c.Ctx.Request.Context(), id)
	if err != nil {
		logging.LogControllerError(c.Ctx, "push.visto.service_error", err, map[string]interface{}{"id": id})
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Dispositivo no encontrado",
		}
		_ = c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Última vista actualizada correctamente",
	}
	_ = c.ServeJSON()
}

// @Title ActualizarTopics
// @Summary Actualizar topics suscritos del dispositivo
// @Tags push_notifications
// @Accept json
// @Produce json
// @Param id query int true "ID del dispositivo"
// @Param body body models.ActualizarTopicsRequest true "Nuevos topics"
// @Success 200 {object} models.ApiResponse
// @Failure 400 {object} models.ApiResponse
// @Failure 404 {object} models.ApiResponse
// @Router /push/dispositivos/topics [patch]
func (c *PushController) ActualizarTopics() {
	id, err := c.GetInt64("id")
	if err != nil || id == 0 {
		logging.LogControllerError(c.Ctx, "push.topics.bad_request", err, map[string]interface{}{"id": c.GetString("id")})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "ID inválido o ausente",
		}
		_ = c.ServeJSON()
		return
	}

	var req models.ActualizarTopicsRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		logging.LogControllerError(c.Ctx, "push.topics.bad_json", err, map[string]interface{}{"id": id})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "JSON inválido",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	pushService := services.NewPushService(orm.NewOrm())
	err = pushService.ActualizarTopicsDispositivo(c.Ctx.Request.Context(), id, req.SubscribedTopics)
	if err != nil {
		logging.LogControllerError(c.Ctx, "push.topics.service_error", err, map[string]interface{}{"id": id})
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Dispositivo no encontrado",
		}
		_ = c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Topics actualizados correctamente",
	}
	_ = c.ServeJSON()
}

// @Title EnviarNotificacion
// @Summary Enviar notificación push
// @Tags push_notifications
// @Accept json
// @Produce json
// @Param body body models.EnviarNotificacionRequest true "Datos de la notificación a enviar"
// @Success 200 {object} models.ApiResponse{data=models.EnviarNotificacionResponse}
// @Failure 400 {object} models.ApiResponse
// @Failure 422 {object} models.ApiResponse
// @Router /push/enviar [post]
func (c *PushController) EnviarNotificacion() {
	var req models.EnviarNotificacionRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		logging.LogControllerError(c.Ctx, "push.enviar.bad_json", err, nil)
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "JSON inválido",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	// Validar campos requeridos
	if req.Notificacion.Titulo == "" {
		logging.LogControllerError(c.Ctx, "push.enviar.missing_titulo", nil, nil)
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El título es requerido",
		}
		_ = c.ServeJSON()
		return
	}

	if req.Notificacion.Mensaje == "" {
		logging.LogControllerError(c.Ctx, "push.enviar.missing_mensaje", nil, nil)
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El mensaje es requerido",
		}
		_ = c.ServeJSON()
		return
	}

	// Crear servicio y enviar notificación
	service := services.NewPushService(orm.NewOrm())
	response, err := service.EnviarNotificacion(&req)
	if err != nil {
		logging.LogControllerError(c.Ctx, "push.enviar.service_error", err, map[string]interface{}{"titulo": req.Notificacion.Titulo})
		c.Ctx.Output.SetStatus(http.StatusUnprocessableEntity)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusUnprocessableEntity,
			Message: "Error al enviar notificación",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Notificación enviada exitosamente",
		Data:    response,
	}
	_ = c.ServeJSON()
}

// @Title ListarEnvios
// @Summary Listar envíos push
// @Tags push_notifications
// @Accept json
// @Produce json
// @Param dispositivo_id query int false "ID del dispositivo"
// @Param fecha_desde query string false "Fecha desde (YYYY-MM-DD)"
// @Param fecha_hasta query string false "Fecha hasta (YYYY-MM-DD)"
// @Param limit query int false "Límite de resultados (default: 20)"
// @Param offset query int false "Offset para paginación (default: 0)"
// @Success 200 {object} models.ApiResponse{data=models.PaginatedResponse}
// @Failure 400 {object} models.ApiResponse
// @Failure 500 {object} models.ApiResponse
// @Router /push/envios [get]
func (c *PushController) ListarEnvios() {
	o := pushOrmNew()
	qs := o.QueryTable("push_envio")

	// Aplicar filtros
	if dispositivoId := c.GetString("dispositivo_id"); dispositivoId != "" {
		if id, err := strconv.ParseInt(dispositivoId, 10, 64); err == nil {
			qs = qs.Filter("pk_id_push_dispositivo", id)
		}
	}

	if fechaDesde := c.GetString("fecha_desde"); fechaDesde != "" {
		if fecha, err := time.Parse("2006-01-02", fechaDesde); err == nil {
			qs = qs.Filter("sent_at__gte", fecha)
		}
	}

	if fechaHasta := c.GetString("fecha_hasta"); fechaHasta != "" {
		if fecha, err := time.Parse("2006-01-02", fechaHasta); err == nil {
			// Agregar 23:59:59 para incluir todo el día
			fechaFin := fecha.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			qs = qs.Filter("sent_at__lte", fechaFin)
		}
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
		logging.LogControllerError(c.Ctx, "push.envios.count_error", err, nil)
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener envíos",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	// Obtener datos
	var envios []*models.PushEnvio
	_, err = qs.OrderBy("-sent_at").Limit(limit).Offset(int64(offset)).All(&envios)
	if err != nil {
		logging.LogControllerError(c.Ctx, "push.envios.query_error", err, nil)
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener envíos",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))
	page := (offset / limit) + 1

	response := models.PaginatedResponse{
		Data:       envios,
		Total:      total,
		Page:       page,
		PageSize:   limit,
		TotalPages: totalPages,
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Envíos obtenidos exitosamente",
		Data:    response,
	}
	_ = c.ServeJSON()
}

// @Title RegistrarEnvio
// @Summary Registrar envío push
// @Tags push_notifications
// @Accept json
// @Produce json
// @Param body body models.RegistrarEnvioRequest true "Datos del envío"
// @Success 201 {object} models.ApiResponse{data=models.PushEnvio}
// @Failure 400 {object} models.ApiResponse
// @Failure 422 {object} models.ApiResponse
// @Router /push/envios [post]
func (c *PushController) RegistrarEnvio() {
	var req models.RegistrarEnvioRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		logging.LogControllerError(c.Ctx, "push.registrar_envio.bad_json", err, nil)
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "JSON inválido",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	// Validar proveedor
	if !req.Proveedor.IsValid() {
		logging.LogControllerError(c.Ctx, "push.registrar_envio.invalid_proveedor", nil, map[string]interface{}{"proveedor": req.Proveedor})
		c.Ctx.Output.SetStatus(http.StatusUnprocessableEntity)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusUnprocessableEntity,
			Message: "Proveedor no válido - debe ser WEB_PUSH o FCM",
		}
		_ = c.ServeJSON()
		return
	}

	pushService := services.NewPushService(orm.NewOrm())
	envio, err := pushService.RegistrarEnvio(c.Ctx.Request.Context(), &req)
	if err != nil {
		logging.LogControllerError(c.Ctx, "push.registrar_envio.service_error", err, map[string]interface{}{"proveedor": req.Proveedor})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al registrar envío",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusCreated,
		Message: "Envío registrado exitosamente",
		Data:    envio,
	}
	_ = c.ServeJSON()
}
