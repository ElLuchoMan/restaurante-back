package reserva

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	loginc "restaurante/controllers/login"
	"restaurante/internal/clientip"
	"restaurante/internal/httpx"
	"restaurante/internal/notify"
	"restaurante/internal/ratelimit"
	"restaurante/logging"
	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

// ReservaController expone las reservas. Las respuestas siempre llevan el
// envoltorio models.ApiResponse y el status HTTP coincide con su `code`.
type ReservaController struct {
	web.Controller
}

const (
	layoutFecha = "2006-01-02"
	layoutHora  = "15:04:05"

	msgNoEncontrada = "Reserva no encontrada"
	msgSinToken     = "Token ausente o inválido" //nolint:gosec // mensaje de error, no una credencial
)

// consultaRL limita por IP la consulta pública de invitado (anti-enumeración).
var consultaRL = ratelimit.New(ratelimit.EnvInt("RESERVA_CONSULTA_MAX_REQ_PER_MIN", 10), time.Minute, 10000)

var estadosPermitidos = map[models.EstadoReserva]bool{
	models.EstadoReservaPendiente:  true,
	models.EstadoReservaConfirmada: true,
	models.EstadoReservaCancelada:  true,
	models.EstadoReservaCumplida:   true,
}

// apiError es un error de negocio ya traducido a status HTTP.
type apiError struct {
	status  int
	message string
	cause   error
}

func newErr(status int, message string, cause error) *apiError {
	return &apiError{status: status, message: message, cause: cause}
}

func (c *ReservaController) fail(event string, e *apiError) {
	if e.status >= http.StatusInternalServerError {
		logging.LogControllerError(c.Ctx, event, e.cause, nil)
	}
	httpx.Fail(&c.Controller, e.status, e.message, e.cause)
}

// reservasConRelaciones devuelve la consulta base con contacto y restaurante
// cargados (así contactoId y restauranteId llegan completos, sin N+1).
func reservasConRelaciones(o orm.Ormer) orm.QuerySeter {
	return o.QueryTable(new(models.Reserva)).RelatedSel("PK_ID_CONTACTO", "PK_ID_RESTAURANTE")
}

// loadReserva carga una reserva con sus relaciones: 404 si no existe.
func loadReserva(o orm.Ormer, id int64) (*models.Reserva, *apiError) {
	var r models.Reserva
	err := reservasConRelaciones(o).Filter("PK_ID_RESERVA", id).One(&r)
	if errors.Is(err, orm.ErrNoRows) {
		return nil, noEncontrada()
	}
	if err != nil {
		return nil, newErr(http.StatusInternalServerError, "Error al obtener la reserva", err)
	}
	return &r, nil
}

// claimsOrFail devuelve los claims del access token o responde 401 y devuelve nil.
func (c *ReservaController) claimsOrFail() *loginc.Claims {
	claims := loginc.ClaimsFromContext(c.Ctx)
	if claims == nil {
		httpx.Fail(&c.Controller, http.StatusUnauthorized, msgSinToken, nil)
	}
	return claims
}

// poseeReserva indica si el documento del token coincide con el del contacto de
// la reserva (documento de invitado o de cliente registrado).
func poseeReserva(claims *loginc.Claims, r *models.Reserva) bool {
	ct := r.PK_ID_CONTACTO
	if ct == nil {
		return false
	}
	if ct.DocumentoContacto != nil && *ct.DocumentoContacto == claims.Documento {
		return true
	}
	return ct.PKDocumentoCliente != nil && ct.PKDocumentoCliente.PK_DOCUMENTO_CLIENTE == claims.Documento
}

// clienteNotificable devuelve el documento del cliente registrado al que se
// avisa de su reserva (0 si es un invitado sin sesión de cliente): el cliente
// del contacto o, si no lo hay, el cliente que actúa con su token.
func clienteNotificable(claims *loginc.Claims, r *models.Reserva) int64 {
	if ct := r.PK_ID_CONTACTO; ct != nil && ct.PKDocumentoCliente != nil {
		return ct.PKDocumentoCliente.PK_DOCUMENTO_CLIENTE
	}
	if claims != nil && !claims.IsStaff() && poseeReserva(claims, r) {
		return claims.Documento
	}
	return 0
}

// eventoReserva arma el evento de notificación de una reserva ya guardada.
func eventoReserva(tipo notify.Tipo, claims *loginc.Claims, r *models.Reserva) notify.Evento {
	ev := notify.Evento{
		Tipo:          tipo,
		ReservaID:     r.PK_ID_RESERVA,
		Cliente:       clienteNotificable(claims, r),
		Fecha:         r.FECHA.Format(layoutFecha),
		Hora:          r.HORA.Format(layoutHora),
		DesdePersonal: claims.IsStaff(),
	}
	if r.ESTADO_RESERVA != nil {
		ev.Estado = *r.ESTADO_RESERVA
	}
	return ev
}

// loadAutorizada carga la reserva y exige que el llamador sea trabajador o su
// dueño; si no, responde igual que si no existiera (404) para no revelar ids.
func loadAutorizada(o orm.Ormer, id int64, claims *loginc.Claims) (*models.Reserva, *apiError) {
	r, apiErr := loadReserva(o, id)
	if apiErr != nil {
		return nil, apiErr
	}
	if !claims.IsStaff() && !poseeReserva(claims, r) {
		return nil, noEncontrada()
	}
	return r, nil
}

// noEncontrada es el 404 uniforme (mismo cuerpo y causa) para "no existe" y
// "no es suya", de modo que no sirva para enumerar reservas.
func noEncontrada() *apiError {
	return newErr(http.StatusNotFound, msgNoEncontrada, orm.ErrNoRows)
}

func forbidden(msg string) *apiError {
	return newErr(http.StatusForbidden, msg, nil)
}

// checkClienteUpdate limita lo que un Cliente (no trabajador) puede cambiar en
// su reserva: no puede reasignar el contacto a otro documento ni cambiar el
// estado salvo para cancelar.
func checkClienteUpdate(claims *loginc.Claims, in *models.ReservaUpdateRequest) *apiError {
	if in.DocumentoContacto != nil && *in.DocumentoContacto != claims.Documento {
		return forbidden("No puede asignar la reserva a otro documento")
	}
	if in.DocumentoCliente != nil && *in.DocumentoCliente != claims.Documento {
		return forbidden("No puede asignar la reserva a otro documento")
	}
	if in.EstadoReserva != nil && *in.EstadoReserva != string(models.EstadoReservaCancelada) {
		return forbidden("Solo puede cancelar su reserva; el estado lo gestiona el personal")
	}
	return nil
}

// checkCreateAutorizacion limita la creación pública: sin ser trabajador solo
// se crea en estado PENDIENTE y asociada a un cliente registrado únicamente si
// el token es de ese cliente.
func checkCreateAutorizacion(claims *loginc.Claims, in *models.ReservaCreateRequest) *apiError {
	if claims.IsStaff() {
		return nil
	}
	if in.EstadoReserva != nil && *in.EstadoReserva != "" && *in.EstadoReserva != string(models.EstadoReservaPendiente) {
		return forbidden("El estado inicial lo define el personal")
	}
	if in.DocumentoCliente != nil && in.DocumentoContacto == nil {
		if claims == nil {
			return newErr(http.StatusUnauthorized, msgSinToken, nil)
		}
		if claims.Documento != *in.DocumentoCliente {
			return forbidden("No puede reservar a nombre de otro cliente")
		}
	}
	return nil
}

func parseFecha(s string) (time.Time, *apiError) {
	parsed, err := time.Parse(layoutFecha, s)
	if err != nil {
		return time.Time{}, newErr(http.StatusBadRequest, "Formato de fecha inválido (use YYYY-MM-DD)", err)
	}
	return time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 12, 0, 0, 0, time.UTC), nil
}

func parseHora(s string) (time.Time, *apiError) {
	parsed, err := time.Parse(layoutHora, s)
	if err != nil {
		return time.Time{}, newErr(http.StatusBadRequest, "Formato de hora inválido (use HH:MM:SS)", err)
	}
	return parsed, nil
}

func parseEstado(s string) (models.EstadoReserva, *apiError) {
	estado := models.EstadoReserva(s)
	if !estadosPermitidos[estado] {
		return "", newErr(http.StatusBadRequest, "Estado de reserva inválido", errors.New("el estado debe ser uno de: PENDIENTE, CONFIRMADA, CANCELADA, CUMPLIDA"))
	}
	return estado, nil
}

func checkPersonas(n int) *apiError {
	if n <= 0 {
		return newErr(http.StatusBadRequest, "El campo personas debe ser un número mayor a 0", nil)
	}
	return nil
}

// checkRestaurante valida que el restaurante exista (404 si no).
func checkRestaurante(o orm.Ormer, id int64) (*models.Restaurante, *apiError) {
	if id <= 0 {
		return nil, newErr(http.StatusBadRequest, "El campo restauranteId debe ser un entero positivo", nil)
	}
	rest := models.Restaurante{PK_ID_RESTAURANTE: id}
	err := o.Read(&rest)
	if errors.Is(err, orm.ErrNoRows) {
		return nil, newErr(http.StatusNotFound, "Restaurante no encontrado", err)
	}
	if err != nil {
		return nil, newErr(http.StatusInternalServerError, "Error al consultar el restaurante", err)
	}
	return &rest, nil
}

// contactoInput agrupa los datos de contacto aceptados en POST y PUT.
type contactoInput struct {
	documentoContacto *int64
	documentoCliente  *int64
	nombreCompleto    *string
	telefono          *string
}

func (in contactoInput) provided() bool {
	return in.documentoContacto != nil || in.documentoCliente != nil
}

// resolveContacto busca (o crea) el contacto: por documentoContacto (invitado)
// o, si no viene, por documentoCliente (cliente registrado).
func resolveContacto(o orm.Ormer, in contactoInput) (*models.ReservaContacto, *apiError) {
	if in.documentoContacto != nil {
		return contactoInvitado(o, *in.documentoContacto, in)
	}
	if in.documentoCliente != nil {
		return contactoCliente(o, *in.documentoCliente)
	}
	return nil, newErr(http.StatusBadRequest, "Error al procesar contacto", errors.New("debe proporcionar documentoContacto o documentoCliente"))
}

func contactoInvitado(o orm.Ormer, documento int64, in contactoInput) (*models.ReservaContacto, *apiError) {
	if documento <= 0 {
		return nil, newErr(http.StatusBadRequest, "Error al procesar contacto", errors.New("documentoContacto debe ser un entero positivo"))
	}
	var contacto models.ReservaContacto
	err := o.QueryTable(new(models.ReservaContacto)).Filter("DocumentoContacto", documento).One(&contacto)
	if err == nil {
		return &contacto, nil
	}
	if !errors.Is(err, orm.ErrNoRows) {
		return nil, newErr(http.StatusInternalServerError, "Error al consultar el contacto", err)
	}
	if in.nombreCompleto == nil || strings.TrimSpace(*in.nombreCompleto) == "" {
		return nil, newErr(http.StatusBadRequest, "Error al procesar contacto", errors.New("nombreCompleto es requerido para usuarios no registrados"))
	}
	contacto = models.ReservaContacto{DocumentoContacto: &documento, NombreCompleto: strings.TrimSpace(*in.nombreCompleto)}
	if in.telefono != nil && *in.telefono != "" {
		contacto.Telefono = in.telefono
	}
	return insertContacto(o, &contacto)
}

func contactoCliente(o orm.Ormer, documento int64) (*models.ReservaContacto, *apiError) {
	if documento <= 0 {
		return nil, newErr(http.StatusBadRequest, "Error al procesar contacto", errors.New("documentoCliente debe ser un entero positivo"))
	}
	var contacto models.ReservaContacto
	err := o.QueryTable(new(models.ReservaContacto)).Filter("PKDocumentoCliente", documento).One(&contacto)
	if err == nil {
		return &contacto, nil
	}
	if !errors.Is(err, orm.ErrNoRows) {
		return nil, newErr(http.StatusInternalServerError, "Error al consultar el contacto", err)
	}
	cliente := models.Cliente{PK_DOCUMENTO_CLIENTE: documento}
	if err := o.Read(&cliente); err != nil {
		if errors.Is(err, orm.ErrNoRows) {
			return nil, newErr(http.StatusNotFound, "Cliente no encontrado", err)
		}
		return nil, newErr(http.StatusInternalServerError, "Error al consultar el cliente", err)
	}
	contacto = models.ReservaContacto{
		PKDocumentoCliente: &models.Cliente{PK_DOCUMENTO_CLIENTE: documento},
		NombreCompleto:     strings.TrimSpace(cliente.NOMBRE + " " + cliente.APELLIDO),
	}
	if cliente.TELEFONO != "" {
		contacto.Telefono = &cliente.TELEFONO
	}
	return insertContacto(o, &contacto)
}

func insertContacto(o orm.Ormer, contacto *models.ReservaContacto) (*models.ReservaContacto, *apiError) {
	id, err := o.Insert(contacto)
	if err != nil {
		return nil, newErr(http.StatusInternalServerError, "Error al crear el contacto", err)
	}
	contacto.PKIDContacto = id
	return contacto, nil
}

// optionalDate lee el query param opcional `fecha` (YYYY-MM-DD).
func optionalDate(c *ReservaController) (string, *apiError) {
	fecha := c.GetString("fecha")
	if fecha == "" {
		return "", nil
	}
	if _, err := time.Parse(layoutFecha, fecha); err != nil {
		return "", newErr(http.StatusBadRequest, "El parámetro 'fecha' debe tener el formato YYYY-MM-DD", err)
	}
	return fecha, nil
}

func (c *ReservaController) sendList(reservas []models.Reserva, vacio, lleno string) {
	msg := lleno
	if len(reservas) == 0 {
		msg = vacio
	}
	httpx.Send(&c.Controller, http.StatusOK, msg, httpx.List(reservas))
}

// @Title GetAll
// @Summary Obtener todas las reservas
// @Description Devuelve todas las reservas con su contacto (nombreCompleto, teléfono, documentos; nunca contraseñas) y su restaurante ya cargados. Lista vacía: `data` es `[]`. Fechas de respuesta: fechaReserva DD-MM-YYYY, horaReserva HH:MM:SS, createdAt/updatedAt DD-MM-YYYY HH:MM:SS. Solo personal (trabajadores/administrador): contiene datos personales de los contactos.
// @Tags reservas
// @Accept json
// @Produce json
// @Success 200 {object} models.ApiResponse{data=[]models.ReservaResponse} "Lista de reservas (puede ser [])"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 403 {object} models.ApiResponse "El token no es de un trabajador"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /reservas [get]
func (c *ReservaController) GetAll() {
	o := orm.NewOrm()
	var reservas []models.Reserva
	if _, err := reservasConRelaciones(o).All(&reservas); err != nil {
		c.fail("reservas.getall.db_error", newErr(http.StatusInternalServerError, "Error al obtener reservas de la base de datos", err))
		return
	}
	c.sendList(reservas, "No se encontraron reservas", "Reservas obtenidas exitosamente")
}

// @Title GetById
// @Summary Obtener reserva por ID
// @Description Devuelve una reserva por ID con contacto y restaurante cargados. Fechas de respuesta: fechaReserva DD-MM-YYYY, horaReserva HH:MM:SS. Requiere token: el personal ve cualquier reserva; un Cliente solo las suyas (documento del token = documento del contacto); en otro caso responde 404. Los invitados usan `GET /reservas/consulta`.
// @Tags reservas
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID de la reserva (entero positivo)"
// @Success 200 {object} models.ApiResponse{data=models.ReservaResponse} "Reserva encontrada"
// @Failure 400 {object} models.ApiResponse "id ausente o inválido"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 404 {object} models.ApiResponse "Reserva no encontrada (o no pertenece al cliente del token)"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /reservas/search [get]
func (c *ReservaController) GetById() {
	claims := c.claimsOrFail()
	if claims == nil {
		return
	}
	id, err := httpx.PositiveInt64Param(&c.Controller, "id")
	if err != nil {
		httpx.Fail(&c.Controller, http.StatusBadRequest, "El parámetro 'id' es inválido o está ausente", err)
		return
	}
	reserva, apiErr := loadAutorizada(orm.NewOrm(), id, claims)
	if apiErr != nil {
		c.fail("reservas.getbyid.db_error", apiErr)
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, "Reserva encontrada", reserva)
}

// @Title Create
// @Summary Crear una nueva reserva
// @Description Crea una reserva. El contacto se resuelve con `documentoContacto` (invitado; si no existe se crea y exige `nombreCompleto`) o con `documentoCliente` (cliente registrado); si se envían ambos prevalece `documentoContacto`. `contactoId` NO se acepta. Peticiones: fechaReserva YYYY-MM-DD, horaReserva HH:MM:SS, personas >= 1, estadoReserva opcional (por defecto PENDIENTE). Público: no exige token (invitado), pero si se envía uno se usa para autorizar. Sin ser trabajador: el estado solo puede ser PENDIENTE (403) y `documentoCliente` (sin `documentoContacto`) exige el token de ese mismo cliente (401/403). La respuesta es la reserva completa (contacto y restaurante, sin contraseñas) solo para el personal o el cliente dueño; para un invitado devuelve únicamente los datos mínimos (`ReservaConsultaResponse`: sin nombre, teléfono ni documento). Fechas de respuesta en DD-MM-YYYY. Envía en segundo plano (best-effort) un push al cliente registrado y, si la crea un cliente o un invitado, un aviso a los trabajadores.
// @Tags reservas
// @Accept json
// @Produce json
// @Param   body  body   models.ReservaCreateRequest true  "Datos de la reserva a crear"
// @Success 201 {object} models.ApiResponse{data=models.ReservaConsultaResponse} "Reserva creada (invitado: datos mínimos; personal o cliente dueño: models.ReservaResponse completa)"
// @Failure 400 {object} models.ApiResponse "JSON, campos obligatorios, fecha, hora, personas, estado o contacto inválidos"
// @Failure 401 {object} models.ApiResponse "documentoCliente sin token"
// @Failure 403 {object} models.ApiResponse "Estado distinto de PENDIENTE o documentoCliente de otro cliente"
// @Failure 404 {object} models.ApiResponse "Restaurante o cliente no encontrado"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Router /reservas [post]
func (c *ReservaController) Post() {
	var in models.ReservaCreateRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &in); err != nil {
		c.fail("reservas.post.bad_json", newErr(http.StatusBadRequest, "Error al decodificar la solicitud", err))
		return
	}

	claims := loginc.ClaimsFromContext(c.Ctx)
	if apiErr := checkCreateAutorizacion(claims, &in); apiErr != nil {
		c.fail("reservas.post.forbidden", apiErr)
		return
	}

	reserva, apiErr := validateCreate(&in)
	if apiErr != nil {
		c.fail("reservas.post.validation_error", apiErr)
		return
	}

	o := orm.NewOrm()
	rest, apiErr := checkRestaurante(o, *in.RestauranteId)
	if apiErr != nil {
		c.fail("reservas.post.restaurante_error", apiErr)
		return
	}
	reserva.PK_ID_RESTAURANTE = rest

	contacto, apiErr := resolveContacto(o, contactoInput{
		documentoContacto: in.DocumentoContacto,
		documentoCliente:  in.DocumentoCliente,
		nombreCompleto:    in.NombreCompleto,
		telefono:          in.Telefono,
	})
	if apiErr != nil {
		c.fail("reservas.post.contacto_error", apiErr)
		return
	}
	reserva.PK_ID_CONTACTO = contacto

	id, err := o.Insert(reserva)
	if err != nil {
		c.fail("reservas.post.insert_error", newErr(http.StatusInternalServerError, "Error al crear la reserva", err))
		return
	}

	creada, apiErr := loadReserva(o, id)
	if apiErr != nil {
		c.fail("reservas.post.reload_error", apiErr)
		return
	}
	notify.Enviar(eventoReserva(notify.ReservaCreada, claims, creada))
	if claims.IsStaff() || (claims != nil && poseeReserva(claims, creada)) {
		httpx.Send(&c.Controller, http.StatusCreated, "Reserva creada correctamente", creada)
		return
	}
	httpx.Send(&c.Controller, http.StatusCreated, "Reserva creada correctamente", consultaView(creada))
}

// validateCreate valida el cuerpo de POST y arma la reserva (sin relaciones).
func validateCreate(in *models.ReservaCreateRequest) (*models.Reserva, *apiError) {
	var reserva models.Reserva
	if in.FechaReserva == nil || *in.FechaReserva == "" {
		return nil, newErr(http.StatusBadRequest, "El campo fechaReserva es requerido", nil)
	}
	fecha, apiErr := parseFecha(*in.FechaReserva)
	if apiErr != nil {
		return nil, apiErr
	}
	reserva.FECHA = fecha

	if in.HoraReserva == nil || *in.HoraReserva == "" {
		return nil, newErr(http.StatusBadRequest, "El campo horaReserva es requerido", nil)
	}
	hora, apiErr := parseHora(*in.HoraReserva)
	if apiErr != nil {
		return nil, apiErr
	}
	reserva.HORA = hora

	if in.Personas == nil {
		return nil, newErr(http.StatusBadRequest, "El campo personas debe ser un número mayor a 0", nil)
	}
	if apiErr := checkPersonas(*in.Personas); apiErr != nil {
		return nil, apiErr
	}
	reserva.PERSONAS = *in.Personas

	estado := models.EstadoReservaPendiente
	if in.EstadoReserva != nil && *in.EstadoReserva != "" {
		estado, apiErr = parseEstado(*in.EstadoReserva)
		if apiErr != nil {
			return nil, apiErr
		}
	}
	reserva.ESTADO_RESERVA = &estado

	if in.Indicaciones != nil && *in.Indicaciones != "" {
		reserva.INDICACIONES = in.Indicaciones
	}
	if in.CreatedBy != nil && *in.CreatedBy != "" {
		reserva.CREATED_BY = in.CreatedBy
	}
	if in.RestauranteId == nil {
		return nil, newErr(http.StatusBadRequest, "El campo restauranteId es requerido", nil)
	}
	return &reserva, nil
}

// @Title Update
// @Summary Actualizar una reserva (merge parcial)
// @Description Actualiza solo los campos enviados; los ausentes se conservan. `indicaciones` y `updatedBy` admiten `null` (limpian el campo); `null` en cualquier otro campo devuelve 400. Para cambiar el contacto envíe `documentoContacto` o `documentoCliente` (se busca o crea el contacto; `contactoId` NO se acepta). Peticiones: fechaReserva YYYY-MM-DD, horaReserva HH:MM:SS, personas >= 1. La respuesta devuelve la reserva completa (fechas DD-MM-YYYY). Requiere token: el personal modifica cualquier reserva; un Cliente solo las suyas (404 si no son suyas) y no puede reasignar el contacto a otro documento ni cambiar el estado salvo a CANCELADA (403). Los invitados no pueden modificar. Si cambia el estado, avisa por push al cliente registrado de la reserva (y a los trabajadores cuando el propio cliente la cancela); best-effort, en segundo plano.
// @Tags reservas
// @Accept json
// @Produce json
// @Param   id    query    int  true   "ID de la reserva (entero positivo)"
// @Param   body  body   models.ReservaUpdateRequest true  "Campos a modificar"
// @Success 200 {object} models.ApiResponse{data=models.ReservaResponse} "Reserva actualizada"
// @Failure 400 {object} models.ApiResponse "id, JSON, null no permitido, fecha, hora, personas, estado o contacto inválidos"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 403 {object} models.ApiResponse "Un Cliente intenta reasignar el contacto o cambiar el estado (salvo cancelar)"
// @Failure 404 {object} models.ApiResponse "Reserva (o no pertenece al cliente del token), restaurante o cliente no encontrado"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /reservas [put]
func (c *ReservaController) Put() {
	claims := c.claimsOrFail()
	if claims == nil {
		return
	}
	id, err := httpx.PositiveInt64Param(&c.Controller, "id")
	if err != nil {
		httpx.Fail(&c.Controller, http.StatusBadRequest, "El parámetro 'id' es inválido o está ausente", err)
		return
	}
	body := c.Ctx.Input.RequestBody
	var in models.ReservaUpdateRequest
	if err := httpx.DecodeMerge(body, &in, "indicaciones", "updatedBy"); err != nil {
		c.fail("reservas.put.bad_json", newErr(http.StatusBadRequest, "Error al decodificar la solicitud", err))
		return
	}

	o := orm.NewOrm()
	reserva, apiErr := loadAutorizada(o, id, claims)
	if apiErr != nil {
		c.fail("reservas.put.load_error", apiErr)
		return
	}
	if !claims.IsStaff() {
		if apiErr := checkClienteUpdate(claims, &in); apiErr != nil {
			c.fail("reservas.put.forbidden", apiErr)
			return
		}
	}
	estadoAnterior := reserva.ESTADO_RESERVA // applyUpdate asigna un puntero nuevo, no escribe sobre este

	cols, apiErr := applyUpdate(o, reserva, &in, body)
	if apiErr != nil {
		c.fail("reservas.put.validation_error", apiErr)
		return
	}
	if _, err := o.Update(reserva, cols...); err != nil {
		c.fail("reservas.put.update_error", newErr(http.StatusInternalServerError, "Error al actualizar la reserva", err))
		return
	}

	actualizada, apiErr := loadReserva(o, id)
	if apiErr != nil {
		c.fail("reservas.put.reload_error", apiErr)
		return
	}
	if cambioEstado(estadoAnterior, actualizada.ESTADO_RESERVA) {
		notify.Enviar(eventoReserva(notify.ReservaEstado, claims, actualizada))
	}
	httpx.Send(&c.Controller, http.StatusOK, "Reserva actualizada correctamente", actualizada)
}

// cambioEstado indica si la reserva quedó con un estado distinto (y definido).
func cambioEstado(antes, despues *models.EstadoReserva) bool {
	return despues != nil && (antes == nil || *antes != *despues)
}

// applyUpdate aplica sobre reserva los campos presentes en in y devuelve las
// columnas a actualizar (siempre incluye UPDATED_AT).
func applyUpdate(o orm.Ormer, reserva *models.Reserva, in *models.ReservaUpdateRequest, body []byte) ([]string, *apiError) {
	cols := []string{"UPDATED_AT"}

	contacto := contactoInput{
		documentoContacto: in.DocumentoContacto,
		documentoCliente:  in.DocumentoCliente,
		nombreCompleto:    in.NombreCompleto,
		telefono:          in.Telefono,
	}
	if in.FechaReserva != nil {
		fecha, apiErr := parseFecha(*in.FechaReserva)
		if apiErr != nil {
			return nil, apiErr
		}
		reserva.FECHA = fecha
		cols = append(cols, "FECHA")
	}
	if in.HoraReserva != nil {
		hora, apiErr := parseHora(*in.HoraReserva)
		if apiErr != nil {
			return nil, apiErr
		}
		reserva.HORA = hora
		cols = append(cols, "HORA")
	}
	if in.Personas != nil {
		if apiErr := checkPersonas(*in.Personas); apiErr != nil {
			return nil, apiErr
		}
		reserva.PERSONAS = *in.Personas
		cols = append(cols, "PERSONAS")
	}
	if in.EstadoReserva != nil {
		estado, apiErr := parseEstado(*in.EstadoReserva)
		if apiErr != nil {
			return nil, apiErr
		}
		reserva.ESTADO_RESERVA = &estado
		cols = append(cols, "ESTADO_RESERVA")
	}
	if in.Indicaciones != nil || httpx.IsNull(body, "indicaciones") {
		reserva.INDICACIONES = in.Indicaciones
		cols = append(cols, "INDICACIONES")
	}
	if in.UpdatedBy != nil || httpx.IsNull(body, "updatedBy") {
		reserva.UPDATED_BY = in.UpdatedBy
		cols = append(cols, "UPDATED_BY")
	}
	if in.RestauranteId != nil {
		rest, apiErr := checkRestaurante(o, *in.RestauranteId)
		if apiErr != nil {
			return nil, apiErr
		}
		reserva.PK_ID_RESTAURANTE = rest
		cols = append(cols, "PK_ID_RESTAURANTE")
	}
	if contacto.provided() {
		nuevo, apiErr := resolveContacto(o, contacto)
		if apiErr != nil {
			return nil, apiErr
		}
		reserva.PK_ID_CONTACTO = nuevo
		cols = append(cols, "PK_ID_CONTACTO")
	}
	return cols, nil
}

// @Title GetByParameter
// @Summary Obtener reservas por contacto y/o fecha
// @Description Devuelve las reservas de un contacto en una fecha, todas las de un contacto, o todas las de una fecha. Sin ningún filtro devuelve todas. Cada reserva trae contacto y restaurante cargados. Filtro `fecha` en YYYY-MM-DD; fechas de respuesta en DD-MM-YYYY. Sin resultados: 200 con `data: []`. Solo personal (sirve también para las reservas del día con `fecha`).
// @Tags reservas
// @Accept json
// @Produce json
// @Param contactoId query int false "ID del contacto (entero positivo)"
// @Param fecha query string false "Fecha de la reserva, formato YYYY-MM-DD"
// @Success 200 {object} models.ApiResponse{data=[]models.ReservaResponse} "Lista de reservas (puede ser [])"
// @Failure 400 {object} models.ApiResponse "contactoId o fecha inválidos"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 403 {object} models.ApiResponse "El token no es de un trabajador"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /reservas/parameter [get]
func (c *ReservaController) GetByParameter() {
	qs := reservasConRelaciones(orm.NewOrm())

	if c.GetString("contactoId") != "" {
		contactoID, err := httpx.PositiveInt64Param(&c.Controller, "contactoId")
		if err != nil {
			httpx.Fail(&c.Controller, http.StatusBadRequest, "El parámetro 'contactoId' es inválido", err)
			return
		}
		qs = qs.Filter("PK_ID_CONTACTO", contactoID)
	}
	fecha, apiErr := optionalDate(c)
	if apiErr != nil {
		c.fail("reservas.parameter.validation_error", apiErr)
		return
	}
	if fecha != "" {
		qs = qs.Filter("FECHA__exact", fecha)
	}

	var reservas []models.Reserva
	if _, err := qs.All(&reservas); err != nil {
		c.fail("reservas.parameter.db_error", newErr(http.StatusInternalServerError, "Error al obtener reservas", err))
		return
	}
	c.sendList(reservas, "No se encontraron reservas", "Reservas obtenidas exitosamente")
}

// reservasPorFiltro ejecuta la consulta con relaciones y el filtro de relación dado.
func reservasPorFiltro(o orm.Ormer, expr string, documento int64, fecha string) ([]models.Reserva, error) {
	qs := reservasConRelaciones(o).Filter(expr, documento)
	if fecha != "" {
		qs = qs.Filter("FECHA__exact", fecha)
	}
	var reservas []models.Reserva
	_, err := qs.All(&reservas)
	return reservas, err
}

// @Title GetByDocumento
// @Summary Obtener reservas por documento (cliente registrado o invitado)
// @Description Busca reservas por documento: primero como cliente registrado y, si no hay resultados, como documento de contacto (invitado). Filtro `fecha` en YYYY-MM-DD; fechas de respuesta en DD-MM-YYYY. Sin resultados: 200 con `data: []`. Solo personal.
// @Tags reservas
// @Accept json
// @Produce json
// @Param documento query int true "Documento del cliente o del contacto (entero positivo)"
// @Param fecha query string false "Fecha de la reserva, formato YYYY-MM-DD"
// @Success 200 {object} models.ApiResponse{data=[]models.ReservaResponse} "Lista de reservas (puede ser [])"
// @Failure 400 {object} models.ApiResponse "documento o fecha inválidos"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 403 {object} models.ApiResponse "El token no es de un trabajador"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /reservas/documento [get]
func (c *ReservaController) GetByDocumento() {
	documento, err := httpx.PositiveInt64Param(&c.Controller, "documento")
	if err != nil {
		httpx.Fail(&c.Controller, http.StatusBadRequest, "El parámetro 'documento' es requerido y debe ser un número válido", err)
		return
	}
	fecha, apiErr := optionalDate(c)
	if apiErr != nil {
		c.fail("reservas.documento.validation_error", apiErr)
		return
	}

	o := orm.NewOrm()
	reservas, err := reservasPorFiltro(o, "PK_ID_CONTACTO__PKDocumentoCliente", documento, fecha)
	if err != nil {
		c.fail("reservas.documento.db_error_cliente", newErr(http.StatusInternalServerError, "Error al obtener reservas", err))
		return
	}
	if len(reservas) == 0 {
		reservas, err = reservasPorFiltro(o, "PK_ID_CONTACTO__DocumentoContacto", documento, fecha)
		if err != nil {
			c.fail("reservas.documento.db_error_contacto", newErr(http.StatusInternalServerError, "Error al obtener reservas", err))
			return
		}
	}
	c.sendList(reservas, "No se encontraron reservas para este documento", "Reservas obtenidas exitosamente")
}

// @Title GetByDocumentoCliente
// @Summary Obtener reservas por documento de cliente registrado
// @Description Devuelve las reservas de un cliente registrado, opcionalmente filtradas por fecha (YYYY-MM-DD). Fechas de respuesta en DD-MM-YYYY. Sin resultados: 200 con `data: []`. Requiere token. Un Cliente solo ve las suyas: el documento sale del token, `documentoCliente` es opcional y, si se envía y no coincide, responde 403. El personal debe enviar `documentoCliente` y puede consultar cualquiera.
// @Tags reservas
// @Accept json
// @Produce json
// @Param documentoCliente query int false "Documento del cliente registrado (entero positivo). Obligatorio para el personal; opcional para un Cliente (debe coincidir con el token)"
// @Param fecha query string false "Fecha de la reserva, formato YYYY-MM-DD"
// @Success 200 {object} models.ApiResponse{data=[]models.ReservaResponse} "Lista de reservas (puede ser [])"
// @Failure 400 {object} models.ApiResponse "documentoCliente o fecha inválidos"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 403 {object} models.ApiResponse "documentoCliente distinto del documento del token (Cliente)"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /reservas/cliente [get]
func (c *ReservaController) GetByDocumentoCliente() {
	claims := c.claimsOrFail()
	if claims == nil {
		return
	}
	documento, apiErr := documentoSolicitado(c, claims)
	if apiErr != nil {
		c.fail("reservas.cliente.documento_error", apiErr)
		return
	}
	fecha, apiErr := optionalDate(c)
	if apiErr != nil {
		c.fail("reservas.cliente.validation_error", apiErr)
		return
	}
	reservas, err := reservasPorFiltro(orm.NewOrm(), "PK_ID_CONTACTO__PKDocumentoCliente", documento, fecha)
	if err != nil {
		c.fail("reservas.cliente.db_error", newErr(http.StatusInternalServerError, "Error al obtener reservas", err))
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, mensajeCliente(len(reservas)), httpx.List(reservas))
}

// documentoSolicitado resuelve el documento a consultar: el personal lo envía
// siempre; un Cliente solo puede consultar el de su token.
func documentoSolicitado(c *ReservaController, claims *loginc.Claims) (int64, *apiError) {
	if !claims.IsStaff() && c.GetString("documentoCliente") == "" {
		return claims.Documento, nil
	}
	documento, err := httpx.PositiveInt64Param(&c.Controller, "documentoCliente")
	if err != nil {
		return 0, newErr(http.StatusBadRequest, "El parámetro 'documentoCliente' es requerido y debe ser un número válido", err)
	}
	if !claims.IsStaff() && documento != claims.Documento {
		return 0, forbidden("Solo puede consultar sus propias reservas")
	}
	return documento, nil
}

func mensajeCliente(n int) string {
	if n == 0 {
		return "No se encontraron reservas para este cliente"
	}
	return "Reservas del cliente obtenidas exitosamente"
}

// @Title Delete
// @Summary Cancelar una reserva
// @Description No borra la reserva: cambia su estado a CANCELADA y devuelve la reserva actualizada. Si ya estaba cancelada responde 409. Requiere token: el personal cancela cualquier reserva; un Cliente solo las suyas (404 si no son suyas). Los invitados no pueden cancelar.
// @Tags reservas
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID de la reserva (entero positivo)"
// @Success 200 {object} models.ApiResponse{data=models.ReservaResponse} "Reserva cancelada"
// @Failure 400 {object} models.ApiResponse "id ausente o inválido"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 404 {object} models.ApiResponse "Reserva no encontrada"
// @Failure 409 {object} models.ApiResponse "La reserva ya estaba cancelada"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /reservas [delete]
func (c *ReservaController) Delete() {
	claims := c.claimsOrFail()
	if claims == nil {
		return
	}
	id, err := httpx.PositiveInt64Param(&c.Controller, "id")
	if err != nil {
		httpx.Fail(&c.Controller, http.StatusBadRequest, "El parámetro 'id' es inválido o está ausente", err)
		return
	}
	o := orm.NewOrm()
	reserva, apiErr := loadAutorizada(o, id, claims)
	if apiErr != nil {
		c.fail("reservas.delete.load_error", apiErr)
		return
	}
	if reserva.ESTADO_RESERVA != nil && *reserva.ESTADO_RESERVA == models.EstadoReservaCancelada {
		c.fail("reservas.delete.conflict", newErr(http.StatusConflict, "La reserva ya está cancelada", fmt.Errorf("reserva %d en estado %s", id, models.EstadoReservaCancelada)))
		return
	}
	cancelada := models.EstadoReservaCancelada
	reserva.ESTADO_RESERVA = &cancelada
	if _, err := o.Update(reserva, "ESTADO_RESERVA", "UPDATED_AT"); err != nil {
		c.fail("reservas.delete.update_error", newErr(http.StatusInternalServerError, "Error al cancelar la reserva", err))
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, "Reserva cancelada correctamente", reserva)
}

// consultaView arma la vista mínima de una reserva para invitados: sin nombre,
// teléfono ni documento del contacto.
func consultaView(r *models.Reserva) models.ReservaConsultaResponse {
	full := r.Response()
	out := models.ReservaConsultaResponse{
		ReservaID:     full.ReservaID,
		FechaReserva:  full.FechaReserva,
		HoraReserva:   full.HoraReserva,
		Personas:      full.Personas,
		EstadoReserva: full.EstadoReserva,
	}
	if full.RestauranteID != nil {
		out.Restaurante = &models.RestauranteConsultaResponse{
			RestauranteID:     full.RestauranteID.RestauranteID,
			NombreRestaurante: full.RestauranteID.NombreRestaurante,
		}
	}
	return out
}

func soloDigitos(s string) string {
	return strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, s)
}

// normalizaTelefono deja solo dígitos y quita el indicativo de Colombia (57).
func normalizaTelefono(s string) string {
	d := soloDigitos(s)
	if len(d) > 10 && strings.HasPrefix(d, "57") {
		return d[2:]
	}
	return d
}

// contactoCoincide indica si el valor dado es el teléfono o el documento
// (de invitado o de cliente registrado) del contacto de la reserva.
func contactoCoincide(ct *models.ReservaContacto, valor string) bool {
	if ct == nil {
		return false
	}
	tel := normalizaTelefono(valor)
	if tel == "" {
		return false
	}
	if ct.Telefono != nil && normalizaTelefono(*ct.Telefono) == tel {
		return true
	}
	doc, err := strconv.ParseInt(soloDigitos(valor), 10, 64)
	if err != nil {
		return false
	}
	if ct.DocumentoContacto != nil && *ct.DocumentoContacto == doc {
		return true
	}
	return ct.PKDocumentoCliente != nil && ct.PKDocumentoCliente.PK_DOCUMENTO_CLIENTE == doc
}

// @Title Consulta
// @Summary Consultar una reserva como invitado
// @Description Consulta pública (sin token) de una reserva con su id y el teléfono o documento del contacto. Devuelve solo los datos mínimos (reservaId, fecha, hora, personas, estado y restaurante): nunca nombre, teléfono ni documento. Para no permitir enumeración, un id inexistente y un contacto que no coincide responden el mismo 404. Límite por IP: 10 peticiones por minuto (`RESERVA_CONSULTA_MAX_REQ_PER_MIN`), 429 al excederlo. Fechas de respuesta en DD-MM-YYYY.
// @Tags reservas
// @Accept json
// @Produce json
// @Param reservaId query int true "ID de la reserva (entero positivo)" example(12)
// @Param contacto query string true "Teléfono o documento del contacto de la reserva" example(3001234567)
// @Success 200 {object} models.ApiResponse{data=models.ReservaConsultaResponse} "Reserva encontrada"
// @Failure 400 {object} models.ApiResponse "reservaId o contacto ausentes o inválidos"
// @Failure 404 {object} models.ApiResponse "Reserva no encontrada (id inexistente o contacto no coincide)"
// @Failure 429 {object} models.ApiResponse "Demasiadas consultas desde esta IP"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Router /reservas/consulta [get]
func (c *ReservaController) Consulta() {
	if !consultaRL.Allow(clientip.FromRequest(c.Ctx.Request, clientip.HopsFromEnv())) {
		httpx.Fail(&c.Controller, http.StatusTooManyRequests, "Demasiadas consultas, intente más tarde", nil)
		return
	}
	id, err := httpx.PositiveInt64Param(&c.Controller, "reservaId")
	if err != nil {
		httpx.Fail(&c.Controller, http.StatusBadRequest, "El parámetro 'reservaId' es inválido o está ausente", err)
		return
	}
	contacto := strings.TrimSpace(c.GetString("contacto"))
	if normalizaTelefono(contacto) == "" {
		httpx.Fail(&c.Controller, http.StatusBadRequest, "El parámetro 'contacto' (teléfono o documento) es requerido", nil)
		return
	}
	reserva, apiErr := loadReserva(orm.NewOrm(), id)
	if apiErr != nil {
		c.fail("reservas.consulta.db_error", apiErr)
		return
	}
	if !contactoCoincide(reserva.PK_ID_CONTACTO, contacto) {
		c.fail("reservas.consulta.no_coincide", noEncontrada())
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, "Reserva encontrada", consultaView(reserva))
}
