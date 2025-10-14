package reserva

import (
	"encoding/json"
	"fmt"
	"net/http"
	"restaurante/logging"
	"restaurante/models"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

type ReservaController struct {
	web.Controller
}

var ormNew = func() orm.Ormer { return orm.NewOrm() }

var estadosPermitidos = map[models.EstadoReserva]bool{
	models.EstadoReservaPendiente:  true,
	models.EstadoReservaConfirmada: true,
	models.EstadoReservaCancelada:  true,
	models.EstadoReservaCumplida:   true,
}

var queryAllReservas = func(o orm.Ormer, reservas *[]models.Reserva) (int64, error) {
	return o.QueryTable(new(models.Reserva)).
		RelatedSel("PK_ID_CONTACTO").
		RelatedSel("PK_ID_RESTAURANTE").
		All(reservas)
}

var readReserva = func(o orm.Ormer, r *models.Reserva) error {
	return o.Read(r)
}

var insertReserva = func(o orm.Ormer, r *models.Reserva) (int64, error) {
	return o.Insert(r)
}

var updateReserva = func(o orm.Ormer, r *models.Reserva, cols ...string) (int64, error) {
	return o.Update(r, cols...)
}

var queryReservasByParam = func(o orm.Ormer, contactoID int64, fecha time.Time, useContacto, useFecha bool, reservas *[]models.Reserva) (int64, error) {
	qs := o.QueryTable(new(models.Reserva))
	if useContacto {
		qs = qs.Filter("PK_ID_CONTACTO", contactoID)
	}
	if useFecha {

		dateStr := fecha.Format("2006-01-02")
		qs = qs.Filter("FECHA__exact", dateStr)
	}
	return qs.All(reservas)
}

var queryReservasByDocumentoCliente = func(o orm.Ormer, documentoCliente int64, fecha time.Time, useFecha bool, reservas *[]models.Reserva) (int64, error) {
	qs := o.QueryTable(new(models.Reserva)).
		RelatedSel("PK_ID_CONTACTO").
		RelatedSel("PK_ID_RESTAURANTE").
		Filter("PK_ID_CONTACTO__PKDocumentoCliente", documentoCliente)

	if useFecha {
		qs = qs.Filter("FECHA", fecha)
	}

	return qs.All(reservas)
}

var queryReservasByDocumentoContacto = func(o orm.Ormer, documentoContacto int64, fecha time.Time, useFecha bool, reservas *[]models.Reserva) (int64, error) {
	qs := o.QueryTable(new(models.Reserva)).
		RelatedSel("PK_ID_CONTACTO").
		RelatedSel("PK_ID_RESTAURANTE").
		Filter("PK_ID_CONTACTO__DocumentoContacto", documentoContacto)

	if useFecha {
		qs = qs.Filter("FECHA", fecha)
	}

	return qs.All(reservas)
}

var insertReservaContacto = func(o orm.Ormer, rc *models.ReservaContacto) (int64, error) {
	return o.Insert(rc)
}

var queryReservaContactoByDocumento = func(o orm.Ormer, documento int64, rc *models.ReservaContacto) error {
	return o.QueryTable(new(models.ReservaContacto)).Filter("DocumentoContacto", documento).One(rc)
}

var queryReservaContactoByCliente = func(o orm.Ormer, clienteDoc int64, rc *models.ReservaContacto) error {
	return o.QueryTable(new(models.ReservaContacto)).Filter("PKDocumentoCliente", clienteDoc).One(rc)
}

var readCliente = func(o orm.Ormer, c *models.Cliente) error {
	return o.Read(c)
}

func createOrFindReservaContacto(o orm.Ormer, input map[string]interface{}) (*models.ReservaContacto, error) {
	var contacto models.ReservaContacto

	if docContacto, ok := input["documentoContacto"].(float64); ok {
		documento := int64(docContacto)

		err := queryReservaContactoByDocumento(o, documento, &contacto)
		if err == nil {

			return &contacto, nil
		}
		if err != orm.ErrNoRows {

			return nil, err
		}

		contacto = models.ReservaContacto{
			DocumentoContacto: &documento,
		}

		if nombre, ok := input["nombreCompleto"].(string); ok && nombre != "" {
			contacto.NombreCompleto = nombre
		} else {
			return nil, fmt.Errorf("nombreCompleto es requerido para usuarios no registrados")
		}

		if telefono, ok := input["telefono"].(string); ok && telefono != "" {
			contacto.Telefono = &telefono
		}

		id, err := insertReservaContacto(o, &contacto)
		if err != nil {
			return nil, err
		}
		contacto.PKIDContacto = id
		return &contacto, nil
	}

	if docCliente, ok := input["documentoCliente"].(float64); ok {
		clienteDoc := int64(docCliente)

		err := queryReservaContactoByCliente(o, clienteDoc, &contacto)
		if err == nil {

			return &contacto, nil
		}
		if err != orm.ErrNoRows {

			return nil, err
		}

		cliente := models.Cliente{PK_DOCUMENTO_CLIENTE: clienteDoc}
		if err := readCliente(o, &cliente); err != nil {
			return nil, fmt.Errorf("cliente no encontrado: %w", err)
		}

		contacto = models.ReservaContacto{
			PKDocumentoCliente: &cliente,
			NombreCompleto:     cliente.NOMBRE + " " + cliente.APELLIDO,
		}
		if cliente.TELEFONO != "" {
			contacto.Telefono = &cliente.TELEFONO
		}

		id, err := insertReservaContacto(o, &contacto)
		if err != nil {
			return nil, err
		}
		contacto.PKIDContacto = id
		return &contacto, nil
	}

	return nil, fmt.Errorf("debe proporcionar documentoContacto o documentoCliente")
}

// @Title GetAll
// @Summary Obtener todas las reservas
// @Description Devuelve todas las reservas registradas en la base de datos.
// @Tags reservas
// @Accept json
// @Produce json
// @Success 200 {array} models.Reserva "Lista de reservas"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Router /reservas [get]
func (c *ReservaController) GetAll() {
	o := ormNew()
	var reservas []models.Reserva

	_, err := queryAllReservas(o, &reservas)
	if err != nil {
		logging.LogControllerError(c.Ctx, "reservas.getall.db_error", err, nil)
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener reservas de la base de datos",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Reservas obtenidas exitosamente",
		Data:    reservas,
	}
	_ = c.ServeJSON()
}

// @Title GetById
// @Summary Obtener reserva por ID
// @Description Devuelve una reserva específica por ID utilizando query parameters.
// @Tags reservas
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID de la Reserva"
// @Success 200 {object} models.Reserva "Reserva encontrada"
// @Failure 404 {object} models.ApiResponse "Reserva no encontrada"
// @Router /reservas/search [get]
func (c *ReservaController) GetById() {
	o := ormNew()
	id, err := c.GetInt64("id")

	if err != nil || id == 0 {
		logging.LogControllerError(c.Ctx, "reservas.getbyid.bad_request", err, map[string]interface{}{"id": c.GetString("id")})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El parámetro 'id' es inválido o está ausente",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	reserva := models.Reserva{PK_ID_RESERVA: id}

	err = readReserva(o, &reserva)
	if err == orm.ErrNoRows {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Reserva no encontrada",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Reserva encontrada",
		Data:    reserva,
	}
	_ = c.ServeJSON()
}

// @Title Create
// @Summary Crear una nueva reserva
// @Description Crea una nueva reserva en la base de datos.
// @Tags reservas
// @Accept json
// @Produce json
// @Param   body  body   models.ReservaCreateRequest true  "Datos de la reserva a crear (fecha YYYY-MM-DD, hora HH:MM:SS)"
// @Success 201 {object} models.Reserva "Reserva creada"
// @Failure 400 {object} models.ApiResponse "Error en la solicitud"
// @Router /reservas [post]
func (c *ReservaController) Post() {
	o := ormNew()
	var input map[string]interface{}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &input); err != nil {
		logging.LogControllerError(c.Ctx, "reservas.post.bad_json", err, map[string]interface{}{"body": string(c.Ctx.Input.RequestBody)})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Error al decodificar la solicitud", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}

	contacto, err := createOrFindReservaContacto(o, input)
	if err != nil {
		logging.LogControllerError(c.Ctx, "reservas.post.contacto_error", err, map[string]interface{}{"body": string(c.Ctx.Input.RequestBody)})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Error al procesar contacto", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}

	var reserva models.Reserva

	if fechaStr, ok := input["fechaReserva"].(string); ok && fechaStr != "" {

		parsedDateUTC, err := time.Parse("2006-01-02", fechaStr)
		if err != nil {
			logging.LogControllerError(c.Ctx, "reservas.post.validation_error", err, map[string]interface{}{"fechaReserva": fechaStr, "body": string(c.Ctx.Input.RequestBody)})
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Formato de fecha inválido (use YYYY-MM-DD)", Cause: err.Error()}
			_ = c.ServeJSON()
			return
		}

		fechaMidUTC := time.Date(parsedDateUTC.Year(), parsedDateUTC.Month(), parsedDateUTC.Day(), 12, 0, 0, 0, time.UTC)

		reserva.FECHA = fechaMidUTC
	} else {
		logging.LogControllerError(c.Ctx, "reservas.post.validation_error", nil, map[string]interface{}{"missing": "fechaReserva", "body": string(c.Ctx.Input.RequestBody)})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "El campo fechaReserva es requerido"}
		_ = c.ServeJSON()
		return
	}

	if horaStr, ok := input["horaReserva"].(string); ok && horaStr != "" {

		parsedHoraUTC, err := time.Parse("15:04:05", horaStr)
		if err != nil {
			logging.LogControllerError(c.Ctx, "reservas.post.validation_error", err, map[string]interface{}{"horaReserva": horaStr, "body": string(c.Ctx.Input.RequestBody)})
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Formato de hora inválido (use HH:MM:SS)", Cause: err.Error()}
			_ = c.ServeJSON()
			return
		}

		reserva.HORA = parsedHoraUTC
	} else {
		logging.LogControllerError(c.Ctx, "reservas.post.validation_error", nil, map[string]interface{}{"missing": "horaReserva", "body": string(c.Ctx.Input.RequestBody)})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "El campo horaReserva es requerido"}
		_ = c.ServeJSON()
		return
	}

	if personas, ok := input["personas"].(float64); ok && personas > 0 {
		reserva.PERSONAS = int(personas)
	} else {
		logging.LogControllerError(c.Ctx, "reservas.post.validation_error", nil, map[string]interface{}{"personas": input["personas"], "body": string(c.Ctx.Input.RequestBody)})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "El campo personas debe ser un número mayor a 0"}
		_ = c.ServeJSON()
		return
	}

	if estadoStr, ok := input["estadoReserva"].(string); ok && estadoStr != "" {
		estado := models.EstadoReserva(estadoStr)
		if !estadosPermitidos[estado] {
			logging.LogControllerError(c.Ctx, "reservas.post.validation_error", nil, map[string]interface{}{"estadoReserva": estadoStr, "body": string(c.Ctx.Input.RequestBody)})
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Estado de reserva inválido", Cause: "El estado debe ser uno de: PENDIENTE, CONFIRMADA, CANCELADA, CUMPLIDA"}
			_ = c.ServeJSON()
			return
		}
		reserva.ESTADO_RESERVA = &estado
	} else {

		estadoDefault := models.EstadoReservaPendiente
		reserva.ESTADO_RESERVA = &estadoDefault
	}

	if indicaciones, ok := input["indicaciones"].(string); ok && indicaciones != "" {
		reserva.INDICACIONES = &indicaciones
	}
	if createdBy, ok := input["createdBy"].(string); ok && createdBy != "" {
		reserva.CREATED_BY = &createdBy
	}

	if restauranteID, ok := input["restauranteId"].(float64); ok {
		val := int64(restauranteID)
		reserva.PK_ID_RESTAURANTE = &models.Restaurante{PK_ID_RESTAURANTE: val}
	} else {
		logging.LogControllerError(c.Ctx, "reservas.post.validation_error", nil, map[string]interface{}{"missing": "restauranteId", "body": string(c.Ctx.Input.RequestBody)})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "El campo restauranteId es requerido"}
		_ = c.ServeJSON()
		return
	}

	reserva.PK_ID_CONTACTO = contacto

	reserva.CREATED_AT = time.Now().UTC()
	reserva.UPDATED_AT = time.Time{}

	if _, err := insertReserva(o, &reserva); err != nil {
		logging.LogControllerError(c.Ctx, "reservas.post.insert_error", err, map[string]interface{}{"contactoId": contacto.PKIDContacto, "restauranteId": reserva.PK_ID_RESTAURANTE, "body": string(c.Ctx.Input.RequestBody)})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al crear la reserva", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = models.ApiResponse{Code: http.StatusCreated, Message: "Reserva creada correctamente", Data: reserva}
	_ = c.ServeJSON()
}

// @Title Update
// @Summary Actualizar una reserva
// @Description Actualiza los datos de una reserva existente.
// @Tags reservas
// @Accept json
// @Produce json
// @Param   id    query    int  true   "ID de la Reserva"
// @Param   body  body   models.ReservaUpdateRequest true  "Datos de la reserva a actualizar (sólo campos a modificar)"
// @Success 200 {object} models.Reserva "Reserva actualizada"
// @Failure 404 {object} models.ApiResponse "Reserva no encontrada"
// @Security BearerAuth
// @Router /reservas [put]
func (c *ReservaController) Put() {
	o := ormNew()
	id, err := c.GetInt64("id")
	if err != nil || id == 0 {
		logging.LogControllerError(c.Ctx, "reservas.put.bad_request", err, map[string]interface{}{"id": c.GetString("id")})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "El parámetro 'id' es inválido o está ausente", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}

	reserva := models.Reserva{PK_ID_RESERVA: id}
	if err := readReserva(o, &reserva); err != nil {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{Code: http.StatusNotFound, Message: "Reserva no encontrada"}
		_ = c.ServeJSON()
		return
	}

	var input map[string]interface{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &input); err != nil {
		logging.LogControllerError(c.Ctx, "reservas.put.bad_json", err, map[string]interface{}{"id": id, "body": string(c.Ctx.Input.RequestBody)})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Error al decodificar la solicitud", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}

	if _, hasDocContacto := input["documentoContacto"]; hasDocContacto {
		contacto, err := createOrFindReservaContacto(o, input)
		if err != nil {
			logging.LogControllerError(c.Ctx, "reservas.put.contacto_error", err, map[string]interface{}{"id": id, "body": string(c.Ctx.Input.RequestBody)})
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Error al procesar contacto", Cause: err.Error()}
			_ = c.ServeJSON()
			return
		}
		reserva.PK_ID_CONTACTO = contacto
	} else if _, hasDocCliente := input["documentoCliente"]; hasDocCliente {
		contacto, err := createOrFindReservaContacto(o, input)
		if err != nil {
			logging.LogControllerError(c.Ctx, "reservas.put.contacto_error", err, map[string]interface{}{"id": id, "body": string(c.Ctx.Input.RequestBody)})
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Error al procesar contacto", Cause: err.Error()}
			_ = c.ServeJSON()
			return
		}
		reserva.PK_ID_CONTACTO = contacto
	}

	if fechaStr, ok := input["fechaReserva"].(string); ok && fechaStr != "" {
		if parsed, err := time.Parse("2006-01-02", fechaStr); err == nil {

			reserva.FECHA = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 12, 0, 0, 0, time.UTC)
		} else {
			logging.LogControllerError(c.Ctx, "reservas.put.validation_error", err, map[string]interface{}{"id": id, "fechaReserva": fechaStr, "body": string(c.Ctx.Input.RequestBody)})
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Formato de fecha inválido (use YYYY-MM-DD)", Cause: err.Error()}
			_ = c.ServeJSON()
			return
		}
	}

	if horaStr, ok := input["horaReserva"].(string); ok && horaStr != "" {
		if parsed, err := time.Parse("15:04:05", horaStr); err == nil {
			reserva.HORA = parsed
		} else {
			logging.LogControllerError(c.Ctx, "reservas.put.validation_error", err, map[string]interface{}{"id": id, "horaReserva": horaStr, "body": string(c.Ctx.Input.RequestBody)})
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Formato de hora inválido (use HH:MM:SS)", Cause: err.Error()}
			_ = c.ServeJSON()
			return
		}
	}

	if personas, ok := input["personas"].(float64); ok && personas > 0 {
		reserva.PERSONAS = int(personas)
	}

	if estadoStr, ok := input["estadoReserva"].(string); ok && estadoStr != "" {
		estado := models.EstadoReserva(estadoStr)
		if estadosPermitidos[estado] {
			reserva.ESTADO_RESERVA = &estado
		} else {
			logging.LogControllerError(c.Ctx, "reservas.put.validation_error", nil, map[string]interface{}{"id": id, "estadoReserva": estadoStr, "body": string(c.Ctx.Input.RequestBody)})
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Estado de reserva inválido", Cause: "El estado debe ser uno de: PENDIENTE, CONFIRMADA, CANCELADA, CUMPLIDA"}
			_ = c.ServeJSON()
			return
		}
	}

	if indicaciones, ok := input["indicaciones"].(string); ok {
		reserva.INDICACIONES = &indicaciones
	}

	if updatedBy, ok := input["updatedBy"].(string); ok && updatedBy != "" {
		reserva.UPDATED_BY = &updatedBy
	}

	if restauranteID, ok := input["restauranteId"].(float64); ok {
		val := int64(restauranteID)
		reserva.PK_ID_RESTAURANTE = &models.Restaurante{PK_ID_RESTAURANTE: val}
	}

	reserva.UPDATED_AT = time.Now().UTC()

	if _, err := updateReserva(o, &reserva); err != nil {
		logging.LogControllerError(c.Ctx, "reservas.put.update_error", err, map[string]interface{}{"id": id, "body": string(c.Ctx.Input.RequestBody)})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al actualizar la reserva", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{Code: http.StatusOK, Message: "Reserva actualizada correctamente", Data: reserva}
	_ = c.ServeJSON()
}

// @Title GetByCliente
// @Summary Obtener reservas por contacto y/o fecha
// @Description Devuelve las reservas asociadas a un contacto en una fecha específica, todas sus reservas si no se especifica la fecha, o todas las reservas en una fecha específica si no se especifica el contacto.
// @Tags reservas
// @Accept json
// @Produce json
// @Param contactoId query int false "ID del Contacto (Opcional)"
// @Param fecha query string false "Fecha de la reserva (YYYY-MM-DD) (Opcional)"
// @Success 200 {array} models.Reserva "Lista de reservas encontradas"
// @Failure 400 {object} models.ApiResponse "Error en los parámetros"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Router /reservas/parameter [get]
func (c *ReservaController) GetByParameter() {
	o := ormNew()
	var reservas []models.Reserva

	contactoID, errContacto := c.GetInt64("contactoId")
	fechaReserva := c.GetString("fecha")

	useContacto := errContacto == nil && contactoID != 0
	var parsedDate time.Time
	useFecha := false
	if fechaReserva != "" {
		var err error

		parsedDate, err = time.Parse("2006-01-02", fechaReserva)
		if err != nil {
			logging.LogControllerError(c.Ctx, "reservas.parameter.validation_error", err, map[string]interface{}{"fecha": fechaReserva})
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusBadRequest,
				Message: "El parámetro 'fecha' debe tener el formato YYYY-MM-DD",
			}
			_ = c.ServeJSON()
			return
		}
		useFecha = true
	}

	_, err := queryReservasByParam(o, contactoID, parsedDate, useContacto, useFecha, &reservas)
	if err != nil {
		logging.LogControllerError(c.Ctx, "reservas.parameter.db_error", err, map[string]interface{}{"contactoId": contactoID, "fecha": fechaReserva})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener reservas",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	if len(reservas) == 0 {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusOK,
			Message: "No se encontraron reservas",
			Data:    reservas,
		}
		_ = c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Reservas obtenidas exitosamente",
		Data:    reservas,
	}
	_ = c.ServeJSON()
}

// @Title GetByDocumento
// @Summary Obtener reservas por documento (cliente loggeado o no loggeado)
// @Description Busca reservas por documento de cliente registrado o documento de contacto. Intenta primero como cliente registrado, luego como contacto.
// @Tags reservas
// @Accept json
// @Produce json
// @Param documento query int true "Documento del Cliente o Contacto"
// @Param fecha query string false "Fecha de la reserva (YYYY-MM-DD) (Opcional)"
// @Success 200 {array} models.Reserva "Lista de reservas encontradas"
// @Failure 400 {object} models.ApiResponse "Error en los parámetros"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Router /reservas/documento [get]
func (c *ReservaController) GetByDocumento() {
	o := ormNew()
	var reservas []models.Reserva

	documento, err := c.GetInt64("documento")
	if err != nil || documento == 0 {
		logging.LogControllerError(c.Ctx, "reservas.documento.bad_request", err, map[string]interface{}{"documento": c.GetString("documento")})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El parámetro 'documento' es requerido y debe ser un número válido",
		}
		_ = c.ServeJSON()
		return
	}

	fechaReserva := c.GetString("fecha")
	var parsedDate time.Time
	useFecha := false

	if fechaReserva != "" {
		var err error
		parsedDate, err = time.Parse("2006-01-02", fechaReserva)
		if err != nil {
			logging.LogControllerError(c.Ctx, "reservas.documento.validation_error", err, map[string]interface{}{"fecha": fechaReserva, "documento": documento})
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusBadRequest,
				Message: "El parámetro 'fecha' debe tener el formato YYYY-MM-DD",
			}
			_ = c.ServeJSON()
			return
		}
		useFecha = true
	}

	count, err := queryReservasByDocumentoCliente(o, documento, parsedDate, useFecha, &reservas)
	if err != nil {
		logging.LogControllerError(c.Ctx, "reservas.documento.db_error_cliente", err, map[string]interface{}{"documento": documento, "fecha": fechaReserva})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener reservas",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	if count == 0 {
		_, err = queryReservasByDocumentoContacto(o, documento, parsedDate, useFecha, &reservas)
		if err != nil {
			logging.LogControllerError(c.Ctx, "reservas.documento.db_error_contacto", err, map[string]interface{}{"documento": documento, "fecha": fechaReserva})
			c.Ctx.Output.SetStatus(http.StatusInternalServerError)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusInternalServerError,
				Message: "Error al obtener reservas",
				Cause:   err.Error(),
			}
			_ = c.ServeJSON()
			return
		}
	}

	if len(reservas) == 0 {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusOK,
			Message: "No se encontraron reservas para este documento",
			Data:    reservas,
		}
		_ = c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Reservas obtenidas exitosamente",
		Data:    reservas,
	}
	_ = c.ServeJSON()
}

// @Title GetByDocumentoCliente
// @Summary Obtener reservas por documento de cliente registrado
// @Description Devuelve las reservas asociadas a un documento de cliente registrado, opcionalmente filtradas por fecha.
// @Tags reservas
// @Accept json
// @Produce json
// @Param documentoCliente query int true "Documento del Cliente Registrado"
// @Param fecha query string false "Fecha de la reserva (YYYY-MM-DD) (Opcional)"
// @Success 200 {array} models.Reserva "Lista de reservas encontradas"
// @Failure 400 {object} models.ApiResponse "Error en los parámetros"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Router /reservas/cliente [get]
func (c *ReservaController) GetByDocumentoCliente() {
	o := ormNew()
	var reservas []models.Reserva

	documentoCliente, err := c.GetInt64("documentoCliente")
	if err != nil || documentoCliente == 0 {
		logging.LogControllerError(c.Ctx, "reservas.cliente.bad_request", err, map[string]interface{}{"documentoCliente": c.GetString("documentoCliente")})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El parámetro 'documentoCliente' es requerido y debe ser un número válido",
		}
		_ = c.ServeJSON()
		return
	}

	fechaReserva := c.GetString("fecha")
	var parsedDate time.Time
	useFecha := false

	if fechaReserva != "" {
		var err error
		parsedDate, err = time.Parse("2006-01-02", fechaReserva)
		if err != nil {
			logging.LogControllerError(c.Ctx, "reservas.cliente.validation_error", err, map[string]interface{}{"fecha": fechaReserva, "documentoCliente": documentoCliente})
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusBadRequest,
				Message: "El parámetro 'fecha' debe tener el formato YYYY-MM-DD",
			}
			_ = c.ServeJSON()
			return
		}
		useFecha = true
	}

	_, err = queryReservasByDocumentoCliente(o, documentoCliente, parsedDate, useFecha, &reservas)
	if err != nil {
		logging.LogControllerError(c.Ctx, "reservas.cliente.db_error", err, map[string]interface{}{"documentoCliente": documentoCliente, "fecha": fechaReserva})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener reservas",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	if len(reservas) == 0 {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusOK,
			Message: "No se encontraron reservas para este cliente",
			Data:    reservas,
		}
		_ = c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Reservas del cliente obtenidas exitosamente",
		Data:    reservas,
	}
	_ = c.ServeJSON()
}

// @Title Delete
// @Summary Cancelar una reserva
// @Description Actualiza el estado de una reserva a "CANCELADA".
// @Tags reservas
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID de la Reserva"
// @Success 200 {object} models.ApiResponse "Reserva cancelada"
// @Failure 404 {object} models.ApiResponse "Reserva no encontrada"
// @Security BearerAuth
// @Router /reservas [delete]
func (c *ReservaController) Delete() {
	o := ormNew()
	id, err := c.GetInt64("id")
	if err != nil || id == 0 {
		logging.LogControllerError(c.Ctx, "reservas.delete.bad_request", err, map[string]interface{}{"id": c.GetString("id")})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El parámetro 'id' es inválido o está ausente",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}
	reserva := models.Reserva{PK_ID_RESERVA: id}
	if err := readReserva(o, &reserva); err != nil {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{Code: http.StatusNotFound, Message: "Reserva no encontrada"}
		_ = c.ServeJSON()
		return
	}
	estadoCancelada := models.EstadoReservaCancelada
	reserva.ESTADO_RESERVA = &estadoCancelada
	reserva.UPDATED_AT = time.Now()
	if _, err := updateReserva(o, &reserva, "estadoReserva", "updatedAt"); err != nil {
		logging.LogControllerError(c.Ctx, "reservas.delete.update_error", err, map[string]interface{}{"id": id})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al cancelar la reserva", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{Code: http.StatusOK, Message: "Reserva cancelada correctamente", Data: reserva}
	_ = c.ServeJSON()
}
