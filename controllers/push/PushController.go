package push

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
	"unicode/utf8"

	"restaurante/internal/httpx"
	"restaurante/logging"
	"restaurante/models"
	"restaurante/services"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

var newPushService = func(o orm.Ormer) services.PushServiceInterface {
	return services.NewPushService(o)
}

var newServiceOrm = func() orm.Ormer { return orm.NewOrm() }

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

const (
	defaultPageSize = 20
	maxPageSize     = 100
	maxTituloLen    = 100
	maxMensajeLen   = 500
)

// idParam lee el query param obligatorio `id` (entero positivo). Si es
// inválido responde 400 y devuelve ok=false.
func (c *PushController) idParam(event string) (id int64, ok bool) {
	id, err := c.GetInt64("id")
	if err != nil || id <= 0 {
		logging.LogControllerError(c.Ctx, event, err, map[string]interface{}{"id": c.GetString("id")})
		httpx.Fail(&c.Controller, http.StatusBadRequest, "ID inválido o ausente", nil)
		return 0, false
	}
	return id, true
}

// positiveIntParam lee un query param opcional entero positivo (0 si no viene).
func (c *PushController) positiveIntParam(name, event string) (v int64, ok bool) {
	raw := c.GetString(name)
	if raw == "" {
		return 0, true
	}
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || v <= 0 {
		logging.LogControllerError(c.Ctx, event, err, map[string]interface{}{name: raw})
		httpx.Fail(&c.Controller, http.StatusBadRequest, name+" inválido: debe ser un entero positivo", nil)
		return 0, false
	}
	return v, true
}

// pagination lee limit (1..100, por defecto 20; los mayores a 100 se reducen a
// 100) y offset (>= 0, por defecto 0). Si son inválidos responde 400.
func (c *PushController) pagination(event string) (limit int, offset int, ok bool) {
	limit, offset = defaultPageSize, 0
	if raw := c.GetString("limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 {
			logging.LogControllerError(c.Ctx, event, err, map[string]interface{}{"limit": raw})
			httpx.Fail(&c.Controller, http.StatusBadRequest, "limit inválido: debe ser un entero mayor o igual a 1", nil)
			return 0, 0, false
		}
		limit = min(n, maxPageSize)
	}
	if raw := c.GetString("offset"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 0 {
			logging.LogControllerError(c.Ctx, event, err, map[string]interface{}{"offset": raw})
			httpx.Fail(&c.Controller, http.StatusBadRequest, "offset inválido: debe ser un entero mayor o igual a 0", nil)
			return 0, 0, false
		}
		offset = n
	}
	return limit, offset, true
}

// sendPage responde 200 con un PaginatedResponse cuyo data nunca es null.
func sendPage[T any](c *PushController, message string, items []T, total int64, limit, offset int) {
	httpx.Send(&c.Controller, http.StatusOK, message, models.PaginatedResponse{
		Data:       httpx.List(items),
		Total:      total,
		Page:       offset/limit + 1,
		PageSize:   limit,
		TotalPages: int((total + int64(limit) - 1) / int64(limit)),
	})
}

// fail traduce un error del servicio: ValidationError -> 400, NotFoundError ->
// 404, ConflictError -> 409 y cualquier otro -> 500 con el mensaje genérico indicado.
func (c *PushController) fail(event, internalMessage string, err error, fields map[string]interface{}) {
	switch {
	case services.IsValidationError(err):
		httpx.Fail(&c.Controller, http.StatusBadRequest, err.Error(), nil)
	case services.IsNotFoundError(err):
		httpx.Fail(&c.Controller, http.StatusNotFound, err.Error(), nil)
	case services.IsConflictError(err):
		httpx.Fail(&c.Controller, http.StatusConflict, err.Error(), nil)
	default:
		logging.LogControllerError(c.Ctx, event, err, fields)
		httpx.Fail(&c.Controller, http.StatusInternalServerError, internalMessage, err)
	}
}

// failBadJSON responde 400 por un cuerpo JSON inválido.
func (c *PushController) failBadJSON(event string, err error) {
	logging.LogControllerError(c.Ctx, event, err, nil)
	httpx.Fail(&c.Controller, http.StatusBadRequest, "JSON inválido", err)
}

// @Title GetAll
// @Summary Listar dispositivos push
// @Description Lista paginada de dispositivos push registrados, del más reciente al más antiguo. `data.data` es `[]` cuando no hay resultados. `documentoCliente` y `documentoTrabajador` son el número de documento (no el objeto completo).
// @Tags push_notifications
// @Accept json
// @Produce json
// @Param cliente_id query int false "Filtra por documento del cliente"
// @Param trabajador_id query int false "Filtra por documento del trabajador"
// @Param plataforma query string false "Filtra por plataforma" Enums(WEB, ANDROID, IOS)
// @Param limit query int false "Tamaño de página (1-100; los valores mayores se reducen a 100)" minimum(1) maximum(100) default(20)
// @Param offset query int false "Desplazamiento para paginación" minimum(0) default(0)
// @Success 200 {object} models.ApiResponse{data=models.PushDispositivosPage} "Dispositivos obtenidos exitosamente"
// @Failure 400 {object} models.ApiResponse "Parámetros inválidos (cliente_id, trabajador_id, plataforma, limit u offset)"
// @Failure 401 {object} models.ApiResponse "Token no proporcionado o inválido"
// @Failure 500 {object} models.ApiResponse "Error interno"
// @Security BearerAuth
// @Router /push/dispositivos [get]
func (c *PushController) GetAll() {
	o := pushOrmNew()
	qs := o.QueryTable("push_dispositivo")

	clienteID, ok := c.positiveIntParam("cliente_id", "push.getall.bad_cliente_id")
	if !ok {
		return
	}
	if clienteID != 0 {
		qs = qs.Filter("pk_documento_cliente", clienteID)
	}

	trabajadorID, ok := c.positiveIntParam("trabajador_id", "push.getall.bad_trabajador_id")
	if !ok {
		return
	}
	if trabajadorID != 0 {
		qs = qs.Filter("pk_documento_trabajador", trabajadorID)
	}

	if plataforma := c.GetString("plataforma"); plataforma != "" {
		if !models.PlataformaNotificacion(plataforma).IsValid() {
			logging.LogControllerError(c.Ctx, "push.getall.bad_plataforma", nil, map[string]interface{}{"plataforma": plataforma})
			httpx.Fail(&c.Controller, http.StatusBadRequest, "plataforma no válida - debe ser WEB, ANDROID o IOS", nil)
			return
		}
		qs = qs.Filter("plataforma", plataforma)
	}

	limit, offset, ok := c.pagination("push.getall.bad_pagination")
	if !ok {
		return
	}

	total, err := qs.Count()
	if err != nil {
		logging.LogControllerError(c.Ctx, "push.getall.count_error", err, nil)
		httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error al obtener dispositivos", err)
		return
	}

	var dispositivos []*models.PushDispositivo
	if _, err = qs.OrderBy("-created_at").Limit(limit).Offset(int64(offset)).All(&dispositivos); err != nil {
		logging.LogControllerError(c.Ctx, "push.getall.query_error", err, nil)
		httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error al obtener dispositivos", err)
		return
	}
	for _, d := range dispositivos {
		d.AfterLoad()
	}

	sendPage(c, "Dispositivos obtenidos exitosamente", dispositivos, total, limit, offset)
}

// @Title Post
// @Summary Registrar dispositivo push
// @Description Registra un dispositivo (upsert por `fcmToken`/`endpoint`). Si el token ya existe se reactiva, se actualizan sus datos y se reasigna al propietario indicado (responde 200 en vez de 201). Debe indicarse exactamente uno entre `documentoCliente` y `documentoTrabajador`. WEB exige `endpoint`, `p256dh` y `auth` y no admite `fcmToken`; ANDROID/IOS exigen `fcmToken` y no admiten los campos web.
// @Tags push_notifications
// @Accept json
// @Produce json
// @Param body body models.RegistrarDispositivoRequest true "Datos del dispositivo"
// @Success 201 {object} models.ApiResponse{data=models.PushDispositivo} "Dispositivo nuevo registrado"
// @Success 200 {object} models.ApiResponse{data=models.PushDispositivo} "Dispositivo ya existente: re-registrado y reactivado"
// @Failure 400 {object} models.ApiResponse "JSON o datos inválidos (plataforma, cliente/trabajador, token o endpoint)"
// @Failure 401 {object} models.ApiResponse "Token no proporcionado o inválido"
// @Failure 404 {object} models.ApiResponse "El cliente o trabajador indicado no existe"
// @Failure 409 {object} models.ApiResponse "Conflicto de unicidad detectado por la base de datos (fcmToken o endpoint duplicado en una carrera)"
// @Failure 500 {object} models.ApiResponse "Error interno"
// @Security BearerAuth
// @Router /push/dispositivos [post]
func (c *PushController) Post() {
	var req models.RegistrarDispositivoRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.failBadJSON("push.post.bad_json", err)
		return
	}

	dispositivo, created, err := newPushService(newServiceOrm()).RegistrarDispositivo(c.Ctx.Request.Context(), &req)
	if err != nil {
		c.fail("push.post.service_error", "Error al registrar dispositivo", err, map[string]interface{}{"plataforma": req.Plataforma})
		return
	}

	if created {
		httpx.Send(&c.Controller, http.StatusCreated, "Dispositivo registrado exitosamente", dispositivo)
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, "Dispositivo actualizado exitosamente", dispositivo)
}

// @Title GetById
// @Summary Obtener dispositivo push por ID
// @Tags push_notifications
// @Accept json
// @Produce json
// @Param id query int true "ID del dispositivo"
// @Success 200 {object} models.ApiResponse{data=models.PushDispositivo} "Dispositivo encontrado"
// @Failure 400 {object} models.ApiResponse "ID inválido o ausente"
// @Failure 401 {object} models.ApiResponse "Token no proporcionado o inválido"
// @Failure 404 {object} models.ApiResponse "Dispositivo no encontrado"
// @Failure 500 {object} models.ApiResponse "Error interno"
// @Security BearerAuth
// @Router /push/dispositivos/search [get]
func (c *PushController) GetById() {
	id, ok := c.idParam("push.getbyid.bad_request")
	if !ok {
		return
	}

	dispositivo := &models.PushDispositivo{PkIdPushDispositivo: id}
	if err := pushOrmNew().Read(dispositivo); err != nil {
		if err == orm.ErrNoRows {
			httpx.Fail(&c.Controller, http.StatusNotFound, "Dispositivo no encontrado", nil)
			return
		}
		logging.LogControllerError(c.Ctx, "push.getbyid.read_error", err, map[string]interface{}{"id": id})
		httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error interno del servidor", err)
		return
	}
	dispositivo.AfterLoad()

	httpx.Send(&c.Controller, http.StatusOK, "Dispositivo encontrado", dispositivo)
}

// @Title Put
// @Summary Actualizar dispositivo push
// @Description Actualización parcial (merge): los campos ausentes se conservan (p. ej. `{"enabled":false}` solo cambia `enabled`). `locale`, `timeZone`, `appVersion` y `userAgent` aceptan `null` para limpiarse; `enabled` y `subscribedTopics` no admiten `null` (400). Devuelve el dispositivo actualizado.
// @Tags push_notifications
// @Accept json
// @Produce json
// @Param id query int true "ID del dispositivo"
// @Param body body models.ActualizarDispositivoRequest true "Campos a actualizar"
// @Success 200 {object} models.ApiResponse{data=models.PushDispositivo} "Dispositivo actualizado"
// @Failure 400 {object} models.ApiResponse "ID o JSON inválido, o null en un campo no anulable"
// @Failure 401 {object} models.ApiResponse "Token no proporcionado o inválido"
// @Failure 404 {object} models.ApiResponse "Dispositivo no encontrado"
// @Failure 500 {object} models.ApiResponse "Error interno"
// @Security BearerAuth
// @Router /push/dispositivos [put]
func (c *PushController) Put() {
	id, ok := c.idParam("push.put.bad_request")
	if !ok {
		return
	}

	dispositivo, err := newPushService(newServiceOrm()).ActualizarDispositivo(c.Ctx.Request.Context(), id, c.Ctx.Input.RequestBody)
	if err != nil {
		c.fail("push.put.service_error", "Error al actualizar dispositivo", err, map[string]interface{}{"id": id})
		return
	}

	httpx.Send(&c.Controller, http.StatusOK, "Dispositivo actualizado correctamente", dispositivo)
}

// @Title Delete
// @Summary Eliminar dispositivo push
// @Description Elimina el dispositivo y, en cascada, su historial de envíos.
// @Tags push_notifications
// @Accept json
// @Produce json
// @Param id query int true "ID del dispositivo"
// @Success 200 {object} models.ApiResponse "Dispositivo eliminado"
// @Failure 400 {object} models.ApiResponse "ID inválido o ausente"
// @Failure 401 {object} models.ApiResponse "Token no proporcionado o inválido"
// @Failure 404 {object} models.ApiResponse "Dispositivo no encontrado"
// @Failure 500 {object} models.ApiResponse "Error interno"
// @Security BearerAuth
// @Router /push/dispositivos [delete]
func (c *PushController) Delete() {
	id, ok := c.idParam("push.delete.bad_request")
	if !ok {
		return
	}

	o := pushOrmNew()
	dispositivo := &models.PushDispositivo{PkIdPushDispositivo: id}
	if err := o.Read(dispositivo); err != nil {
		if err == orm.ErrNoRows {
			httpx.Fail(&c.Controller, http.StatusNotFound, "Dispositivo no encontrado", nil)
			return
		}
		logging.LogControllerError(c.Ctx, "push.delete.read_error", err, map[string]interface{}{"id": id})
		httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error interno del servidor", err)
		return
	}

	if _, err := o.Delete(dispositivo); err != nil {
		logging.LogControllerError(c.Ctx, "push.delete.delete_error", err, map[string]interface{}{"id": id})
		httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error al eliminar dispositivo", err)
		return
	}

	httpx.Send(&c.Controller, http.StatusOK, "Dispositivo eliminado exitosamente", nil)
}

// @Title ActualizarUltimaVista
// @Summary Actualizar última vista del dispositivo
// @Description Marca `lastSeenAt` del dispositivo con la hora actual.
// @Tags push_notifications
// @Accept json
// @Produce json
// @Param id query int true "ID del dispositivo"
// @Success 200 {object} models.ApiResponse "Última vista actualizada"
// @Failure 400 {object} models.ApiResponse "ID inválido o ausente"
// @Failure 401 {object} models.ApiResponse "Token no proporcionado o inválido"
// @Failure 404 {object} models.ApiResponse "Dispositivo no encontrado"
// @Failure 500 {object} models.ApiResponse "Error interno"
// @Security BearerAuth
// @Router /push/dispositivos/visto [patch]
func (c *PushController) ActualizarUltimaVista() {
	id, ok := c.idParam("push.visto.bad_request")
	if !ok {
		return
	}

	if err := newPushService(newServiceOrm()).ActualizarUltimaVista(c.Ctx.Request.Context(), id); err != nil {
		c.fail("push.visto.service_error", "Error al actualizar la última vista", err, map[string]interface{}{"id": id})
		return
	}

	httpx.Send(&c.Controller, http.StatusOK, "Última vista actualizada correctamente", nil)
}

// @Title ActualizarTopics
// @Summary Reemplazar topics suscritos del dispositivo
// @Description Reemplaza la lista de topics del dispositivo (`[]` los elimina todos). `subscribedTopics` es obligatorio y no admite null.
// @Tags push_notifications
// @Accept json
// @Produce json
// @Param id query int true "ID del dispositivo"
// @Param body body models.ActualizarTopicsRequest true "Nuevos topics"
// @Success 200 {object} models.ApiResponse "Topics actualizados"
// @Failure 400 {object} models.ApiResponse "ID o JSON inválido, o subscribedTopics ausente/null"
// @Failure 401 {object} models.ApiResponse "Token no proporcionado o inválido"
// @Failure 404 {object} models.ApiResponse "Dispositivo no encontrado"
// @Failure 500 {object} models.ApiResponse "Error interno"
// @Security BearerAuth
// @Router /push/dispositivos/topics [patch]
func (c *PushController) ActualizarTopics() {
	id, ok := c.idParam("push.topics.bad_request")
	if !ok {
		return
	}

	var req models.ActualizarTopicsRequest
	if err := httpx.DecodeMerge(c.Ctx.Input.RequestBody, &req); err != nil {
		c.failBadJSON("push.topics.bad_json", err)
		return
	}
	if req.SubscribedTopics == nil {
		httpx.Fail(&c.Controller, http.StatusBadRequest, "subscribedTopics es requerido", nil)
		return
	}

	if err := newPushService(newServiceOrm()).ActualizarTopicsDispositivo(c.Ctx.Request.Context(), id, req.SubscribedTopics); err != nil {
		c.fail("push.topics.service_error", "Error al actualizar topics", err, map[string]interface{}{"id": id})
		return
	}

	httpx.Send(&c.Controller, http.StatusOK, "Topics actualizados correctamente", nil)
}

// @Title EnviarNotificacion
// @Summary Enviar notificación push
// @Description Envía la notificación a los dispositivos habilitados que coincidan con `destinatarios` y registra cada envío. Responde 200 aun cuando algún dispositivo falle (ver `enviosFallidos` y `detalleEnvios`).
// @Tags push_notifications
// @Accept json
// @Produce json
// @Param body body models.EnviarNotificacionRequest true "Datos de la notificación a enviar"
// @Success 200 {object} models.ApiResponse{data=models.EnviarNotificacionResponse} "Notificación procesada"
// @Failure 400 {object} models.ApiResponse "JSON o datos inválidos (título/mensaje, remitente o destinatarios)"
// @Failure 401 {object} models.ApiResponse "Token no proporcionado o inválido"
// @Failure 404 {object} models.ApiResponse "El trabajador remitente no existe"
// @Failure 500 {object} models.ApiResponse "Error interno"
// @Security BearerAuth
// @Router /push/enviar [post]
func (c *PushController) EnviarNotificacion() {
	var req models.EnviarNotificacionRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.failBadJSON("push.enviar.bad_json", err)
		return
	}

	if n := utf8.RuneCountInString(req.Notificacion.Titulo); n < 1 || n > maxTituloLen {
		logging.LogControllerError(c.Ctx, "push.enviar.invalid_titulo", nil, nil)
		httpx.Fail(&c.Controller, http.StatusBadRequest, "El título es requerido (máximo 100 caracteres)", nil)
		return
	}

	if n := utf8.RuneCountInString(req.Notificacion.Mensaje); n < 1 || n > maxMensajeLen {
		logging.LogControllerError(c.Ctx, "push.enviar.invalid_mensaje", nil, nil)
		httpx.Fail(&c.Controller, http.StatusBadRequest, "El mensaje es requerido (máximo 500 caracteres)", nil)
		return
	}

	response, err := newPushService(newServiceOrm()).EnviarNotificacion(&req)
	if err != nil {
		c.fail("push.enviar.service_error", "Error al enviar notificación", err, map[string]interface{}{"titulo": req.Notificacion.Titulo})
		return
	}

	httpx.Send(&c.Controller, http.StatusOK, "Notificación enviada exitosamente", response)
}

// @Title ListarEnvios
// @Summary Listar envíos push
// @Description Lista paginada de envíos, del más reciente al más antiguo. `data.data` es `[]` cuando no hay resultados. `pushDispositivoId` es el id numérico del dispositivo y `sentAt` tiene formato DD-MM-YYYY HH:MM:SS (hora de Bogotá).
// @Tags push_notifications
// @Accept json
// @Produce json
// @Param dispositivo_id query int false "Filtra por ID del dispositivo"
// @Param fecha_desde query string false "Envíos desde esta fecha, inclusive (YYYY-MM-DD)" format(date)
// @Param fecha_hasta query string false "Envíos hasta esta fecha, inclusive (YYYY-MM-DD)" format(date)
// @Param limit query int false "Tamaño de página (1-100; los valores mayores se reducen a 100)" minimum(1) maximum(100) default(20)
// @Param offset query int false "Desplazamiento para paginación" minimum(0) default(0)
// @Success 200 {object} models.ApiResponse{data=models.PushEnviosPage} "Envíos obtenidos exitosamente"
// @Failure 400 {object} models.ApiResponse "Parámetros inválidos (dispositivo_id, fechas, limit u offset)"
// @Failure 401 {object} models.ApiResponse "Token no proporcionado o inválido"
// @Failure 500 {object} models.ApiResponse "Error interno"
// @Security BearerAuth
// @Router /push/envios [get]
func (c *PushController) ListarEnvios() {
	o := pushOrmNew()
	qs := o.QueryTable("push_envio")

	dispositivoID, ok := c.positiveIntParam("dispositivo_id", "push.envios.bad_dispositivo_id")
	if !ok {
		return
	}
	if dispositivoID != 0 {
		qs = qs.Filter("pk_id_push_dispositivo", dispositivoID)
	}

	var fechaDesde, fechaHasta time.Time
	for _, p := range []struct {
		name string
		dst  *time.Time
	}{{"fecha_desde", &fechaDesde}, {"fecha_hasta", &fechaHasta}} {
		raw := c.GetString(p.name)
		if raw == "" {
			continue
		}
		fecha, err := models.ParseDateToNoonUTC(raw)
		if err != nil {
			logging.LogControllerError(c.Ctx, "push.envios.bad_fecha", err, map[string]interface{}{p.name: raw})
			httpx.Fail(&c.Controller, http.StatusBadRequest, p.name+" inválida: use el formato YYYY-MM-DD", nil)
			return
		}
		*p.dst = fecha
	}
	if !fechaDesde.IsZero() {
		qs = qs.Filter("sent_at__gte", fechaDesde)
	}
	if !fechaHasta.IsZero() {
		qs = qs.Filter("sent_at__lte", fechaHasta.Add(23*time.Hour+59*time.Minute+59*time.Second))
	}

	limit, offset, ok := c.pagination("push.envios.bad_pagination")
	if !ok {
		return
	}

	total, err := qs.Count()
	if err != nil {
		logging.LogControllerError(c.Ctx, "push.envios.count_error", err, nil)
		httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error al obtener envíos", err)
		return
	}

	var envios []*models.PushEnvio
	if _, err = qs.OrderBy("-sent_at").Limit(limit).Offset(int64(offset)).All(&envios); err != nil {
		logging.LogControllerError(c.Ctx, "push.envios.query_error", err, nil)
		httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error al obtener envíos", err)
		return
	}
	for _, e := range envios {
		e.AfterLoad()
	}

	sendPage(c, "Envíos obtenidos exitosamente", envios, total, limit, offset)
}

// @Title RegistrarEnvio
// @Summary Registrar envío push
// @Description Registra manualmente el resultado de un envío ya realizado a un dispositivo.
// @Tags push_notifications
// @Accept json
// @Produce json
// @Param body body models.RegistrarEnvioRequest true "Datos del envío"
// @Success 201 {object} models.ApiResponse{data=models.PushEnvio} "Envío registrado"
// @Failure 400 {object} models.ApiResponse "JSON o datos inválidos (pushDispositivoId, proveedor)"
// @Failure 401 {object} models.ApiResponse "Token no proporcionado o inválido"
// @Failure 404 {object} models.ApiResponse "El dispositivo indicado no existe"
// @Failure 500 {object} models.ApiResponse "Error interno"
// @Security BearerAuth
// @Router /push/envios [post]
func (c *PushController) RegistrarEnvio() {
	var req models.RegistrarEnvioRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.failBadJSON("push.registrar_envio.bad_json", err)
		return
	}

	envio, err := newPushService(newServiceOrm()).RegistrarEnvio(c.Ctx.Request.Context(), &req)
	if err != nil {
		c.fail("push.registrar_envio.service_error", "Error al registrar envío", err, map[string]interface{}{"proveedor": req.Proveedor})
		return
	}

	httpx.Send(&c.Controller, http.StatusCreated, "Envío registrado exitosamente", envio)
}
