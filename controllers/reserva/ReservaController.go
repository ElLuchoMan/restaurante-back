package reserva

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"restaurante/internal/httpx"
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

	msgEstadoInvalido = "El estado debe ser uno de: PENDIENTE, CONFIRMADA, CANCELADA, CUMPLIDA"
)

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
		return nil, newErr(http.StatusNotFound, "Reserva no encontrada", err)
	}
	if err != nil {
		return nil, newErr(http.StatusInternalServerError, "Error al obtener la reserva", err)
	}
	return &r, nil
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
		return "", newErr(http.StatusBadRequest, "Estado de reserva inválido", errors.New(msgEstadoInvalido))
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
// @Description Devuelve todas las reservas con su contacto (nombreCompleto, teléfono, documentos; nunca contraseñas) y su restaurante ya cargados. Lista vacía: `data` es `[]`. Fechas de respuesta: fechaReserva DD-MM-YYYY, horaReserva HH:MM:SS, createdAt/updatedAt DD-MM-YYYY HH:MM:SS. Público (no exige token).
// @Tags reservas
// @Accept json
// @Produce json
// @Success 200 {object} models.ApiResponse{data=[]models.ReservaResponse} "Lista de reservas (puede ser [])"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
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
// @Description Devuelve una reserva por ID con contacto y restaurante cargados. Fechas de respuesta: fechaReserva DD-MM-YYYY, horaReserva HH:MM:SS. Público (no exige token).
// @Tags reservas
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID de la reserva (entero positivo)"
// @Success 200 {object} models.ApiResponse{data=models.ReservaResponse} "Reserva encontrada"
// @Failure 400 {object} models.ApiResponse "id ausente o inválido"
// @Failure 404 {object} models.ApiResponse "Reserva no encontrada"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Router /reservas/search [get]
func (c *ReservaController) GetById() {
	id, err := httpx.PositiveInt64Param(&c.Controller, "id")
	if err != nil {
		httpx.Fail(&c.Controller, http.StatusBadRequest, "El parámetro 'id' es inválido o está ausente", err)
		return
	}
	reserva, apiErr := loadReserva(orm.NewOrm(), id)
	if apiErr != nil {
		c.fail("reservas.getbyid.db_error", apiErr)
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, "Reserva encontrada", reserva)
}

// @Title Create
// @Summary Crear una nueva reserva
// @Description Crea una reserva. El contacto se resuelve con `documentoContacto` (invitado; si no existe se crea y exige `nombreCompleto`) o con `documentoCliente` (cliente registrado); si se envían ambos prevalece `documentoContacto`. `contactoId` NO se acepta. Peticiones: fechaReserva YYYY-MM-DD, horaReserva HH:MM:SS, personas >= 1, estadoReserva opcional (por defecto PENDIENTE). La respuesta devuelve la reserva con contacto y restaurante (sin contraseñas); fechas de respuesta en DD-MM-YYYY. Público (no exige token).
// @Tags reservas
// @Accept json
// @Produce json
// @Param   body  body   models.ReservaCreateRequest true  "Datos de la reserva a crear"
// @Success 201 {object} models.ApiResponse{data=models.ReservaResponse} "Reserva creada"
// @Failure 400 {object} models.ApiResponse "JSON, campos obligatorios, fecha, hora, personas, estado o contacto inválidos"
// @Failure 404 {object} models.ApiResponse "Restaurante o cliente no encontrado"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Router /reservas [post]
func (c *ReservaController) Post() {
	var in models.ReservaCreateRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &in); err != nil {
		c.fail("reservas.post.bad_json", newErr(http.StatusBadRequest, "Error al decodificar la solicitud", err))
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
	httpx.Send(&c.Controller, http.StatusCreated, "Reserva creada correctamente", creada)
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
// @Description Actualiza solo los campos enviados; los ausentes se conservan. `indicaciones` y `updatedBy` admiten `null` (limpian el campo); `null` en cualquier otro campo devuelve 400. Para cambiar el contacto envíe `documentoContacto` o `documentoCliente` (se busca o crea el contacto; `contactoId` NO se acepta). Peticiones: fechaReserva YYYY-MM-DD, horaReserva HH:MM:SS, personas >= 1. La respuesta devuelve la reserva completa (fechas DD-MM-YYYY).
// @Tags reservas
// @Accept json
// @Produce json
// @Param   id    query    int  true   "ID de la reserva (entero positivo)"
// @Param   body  body   models.ReservaUpdateRequest true  "Campos a modificar"
// @Success 200 {object} models.ApiResponse{data=models.ReservaResponse} "Reserva actualizada"
// @Failure 400 {object} models.ApiResponse "id, JSON, null no permitido, fecha, hora, personas, estado o contacto inválidos"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 404 {object} models.ApiResponse "Reserva, restaurante o cliente no encontrado"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /reservas [put]
func (c *ReservaController) Put() {
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
	reserva, apiErr := loadReserva(o, id)
	if apiErr != nil {
		c.fail("reservas.put.load_error", apiErr)
		return
	}

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
	httpx.Send(&c.Controller, http.StatusOK, "Reserva actualizada correctamente", actualizada)
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
// @Description Devuelve las reservas de un contacto en una fecha, todas las de un contacto, o todas las de una fecha. Sin ningún filtro devuelve todas. Cada reserva trae contacto y restaurante cargados. Filtro `fecha` en YYYY-MM-DD; fechas de respuesta en DD-MM-YYYY. Sin resultados: 200 con `data: []`. Público (no exige token).
// @Tags reservas
// @Accept json
// @Produce json
// @Param contactoId query int false "ID del contacto (entero positivo)"
// @Param fecha query string false "Fecha de la reserva, formato YYYY-MM-DD"
// @Success 200 {object} models.ApiResponse{data=[]models.ReservaResponse} "Lista de reservas (puede ser [])"
// @Failure 400 {object} models.ApiResponse "contactoId o fecha inválidos"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
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
// @Description Busca reservas por documento: primero como cliente registrado y, si no hay resultados, como documento de contacto (invitado). Filtro `fecha` en YYYY-MM-DD; fechas de respuesta en DD-MM-YYYY. Sin resultados: 200 con `data: []`. Público (no exige token).
// @Tags reservas
// @Accept json
// @Produce json
// @Param documento query int true "Documento del cliente o del contacto (entero positivo)"
// @Param fecha query string false "Fecha de la reserva, formato YYYY-MM-DD"
// @Success 200 {object} models.ApiResponse{data=[]models.ReservaResponse} "Lista de reservas (puede ser [])"
// @Failure 400 {object} models.ApiResponse "documento o fecha inválidos"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
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
// @Description Devuelve las reservas de un cliente registrado, opcionalmente filtradas por fecha (YYYY-MM-DD). Fechas de respuesta en DD-MM-YYYY. Sin resultados: 200 con `data: []`. Público (no exige token).
// @Tags reservas
// @Accept json
// @Produce json
// @Param documentoCliente query int true "Documento del cliente registrado (entero positivo)"
// @Param fecha query string false "Fecha de la reserva, formato YYYY-MM-DD"
// @Success 200 {object} models.ApiResponse{data=[]models.ReservaResponse} "Lista de reservas (puede ser [])"
// @Failure 400 {object} models.ApiResponse "documentoCliente o fecha inválidos"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Router /reservas/cliente [get]
func (c *ReservaController) GetByDocumentoCliente() {
	documento, err := httpx.PositiveInt64Param(&c.Controller, "documentoCliente")
	if err != nil {
		httpx.Fail(&c.Controller, http.StatusBadRequest, "El parámetro 'documentoCliente' es requerido y debe ser un número válido", err)
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

func mensajeCliente(n int) string {
	if n == 0 {
		return "No se encontraron reservas para este cliente"
	}
	return "Reservas del cliente obtenidas exitosamente"
}

// @Title Delete
// @Summary Cancelar una reserva
// @Description No borra la reserva: cambia su estado a CANCELADA y devuelve la reserva actualizada. Si ya estaba cancelada responde 409.
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
	id, err := httpx.PositiveInt64Param(&c.Controller, "id")
	if err != nil {
		httpx.Fail(&c.Controller, http.StatusBadRequest, "El parámetro 'id' es inválido o está ausente", err)
		return
	}
	o := orm.NewOrm()
	reserva, apiErr := loadReserva(o, id)
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
