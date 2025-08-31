package controllers

import (
	"encoding/json"
	"net/http"
	"restaurante/database"
	"restaurante/models"
	"strconv"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

type PagoController struct {
	web.Controller
}

type ormer interface {
	QueryTable(interface{}) orm.QuerySeter
	Read(interface{}, ...string) error
	Insert(interface{}) (int64, error)
	Update(interface{}, ...string) (int64, error)
	Delete(interface{}, ...string) (int64, error)
}

var pagoNewOrm = func() ormer { return orm.NewOrm() }

// Estados permitidos para los pagos
var estadosPagoPermitidos = map[string]bool{
	"pagado":    true,
	"pendiente": true,
	"no pago":   true,
}

// @Title GetAll
// @Summary Obtener todos los pagos con filtros
// @Description Devuelve todos los pagos registrados en la base de datos, con opción de filtrar por fecha exacta, mes, año y estado.
// @Tags pagos
// @Accept json
// @Produce json
// @Param   fecha    query   string   false   "Filtrar por fecha exacta (YYYY-MM-DD)"
// @Param   dia      query   int      false   "Filtrar por dia (1-31)"
// @Param   mes      query   int      false   "Filtrar por mes (1-12)"
// @Param   anio     query   int      false   "Filtrar por año (YYYY)"
// @Param   estado   query   string   false   "Filtrar por estado del pago (pagado, pendiente, no pago)"
// @Param   metodo_pago     query   int      false   "Filtrar por metodo de pago"
// @Success 200 {array} models.Pago "Lista de pagos"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /pagos [get]
func (c *PagoController) GetAll() {
	o := pagoNewOrm()
	var pagos []models.Pago

	_, err := o.QueryTable(new(models.Pago)).All(&pagos)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener pagos de la base de datos",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Ajustar fechas y hora al formato correcto
	for i := range pagos {
		pagos[i].UPDATED_AT = pagos[i].UPDATED_AT.UTC()
		pagos[i].FECHA = pagos[i].FECHA.UTC()
		pagos[i].HORA = pagos[i].HORA.UTC()
	}

	// Leer parámetros de la URL
	fecha := c.GetString("fecha")
	dia, _ := c.GetInt("dia")
	mes, _ := c.GetInt("mes")
	anio, _ := c.GetInt("anio")
	estado := c.GetString("estado")
	metodo_pago, _ := c.GetInt("metodo_pago")

	// Filtrar los pagos según los parámetros proporcionados
	var filteredPagos []models.Pago
	for _, pago := range pagos {
		if fecha != "" && pago.FECHA.Format("2006-01-02") != fecha {
			continue
		}
		if dia > 0 && dia <= 31 && pago.FECHA.Day() != dia {
			continue
		}
		if mes > 0 && mes <= 12 && int(pago.FECHA.Month()) != mes {
			continue
		}
		if anio > 0 && pago.FECHA.Year() != anio {
			continue
		}
		if estado != "" && pago.ESTADO_PAGO != estado {
			continue
		}
		if metodo_pago > 0 && pago.PK_ID_METODO_PAGO != int64(metodo_pago) {
			continue
		}

		filteredPagos = append(filteredPagos, pago)
	}

	// Si no hay resultados
	if len(filteredPagos) == 0 {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "No se encontraron pagos que coincidan con los filtros proporcionados",
		}
		c.ServeJSON()
		return
	}

	// Respuesta con los pagos filtrados
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Pagos obtenidos exitosamente",
		Data:    filteredPagos,
	}
	c.ServeJSON()
}

// @Title GetById
// @Summary Obtener pago por ID
// @Description Devuelve un pago específico por ID.
// @Tags pagos
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID del Pago"
// @Success 200 {object} models.Pago "Pago encontrado"
// @Failure 404 {object} models.ApiResponse "Pago no encontrado"
// @Security BearerAuth
// @Router /pagos/search [get]
func (c *PagoController) GetById() {
	o := pagoNewOrm()
	id, err := c.GetInt("id")

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

	pago := models.Pago{PK_ID_PAGO: int64(id)}
	err = o.Read(&pago)
	if err == orm.ErrNoRows {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Pago no encontrado",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Ajustar fechas y hora
	pago.FECHA = pago.FECHA.In(database.BogotaZone)
	pago.UPDATED_AT = pago.UPDATED_AT.In(database.BogotaZone)
	pago.HORA = pago.HORA.In(database.BogotaZone)

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Pago encontrado",
		Data:    pago,
	}
	c.ServeJSON()
}

// @Title Create
// @Summary Crear un nuevo pago
// @Description Crea un nuevo pago en la base de datos.
// @Tags pagos
// @Accept json
// @Produce json
// @Param   body  body   models.Pago true  "Datos del pago a crear"
// @Success 201 {object} models.Pago "Pago creado"
// @Failure 400 {object} models.ApiResponse "Error en la solicitud"
// @Security BearerAuth
// @Router /pagos [post]
func (c *PagoController) Post() {
	o := pagoNewOrm()

	// Estructura de entrada *según swagger*
	type pagoIn struct {
		EstadoPago   string `json:"estadoPago"`
		FechaPago    string `json:"fechaPago"`    // YYYY-MM-DD
		HoraPago     string `json:"horaPago"`     // HH:mm:ss
		MetodoPagoId int64  `json:"metodoPagoId"` // entero
		Monto        int64  `json:"monto"`        // entero
		UpdatedAt    string `json:"updatedAt"`    // ignorado al insertar
		UpdatedBy    string `json:"updatedBy"`
	}

	var in pagoIn
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &in); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Error al decodificar la solicitud",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Validaciones y parseo
	if in.FechaPago == "" {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "El campo fechaPago no puede estar vacío"}
		c.ServeJSON()
		return
	}
	fecha, err := time.Parse("2006-01-02", in.FechaPago)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Formato de fecha inválido (use YYYY-MM-DD)", Cause: err.Error()}
		c.ServeJSON()
		return
	}

	if in.HoraPago == "" {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "El campo horaPago no puede estar vacío"}
		c.ServeJSON()
		return
	}
	hora, err := time.Parse("15:04:05", in.HoraPago)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Formato de hora inválido, debe ser HH:mm:ss", Cause: err.Error()}
		c.ServeJSON()
		return
	}

	if in.Monto == 0 {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "El campo monto es obligatorio y debe ser un número"}
		c.ServeJSON()
		return
	}

	if in.EstadoPago != "" {
		if !estadosPagoPermitidos[in.EstadoPago] {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusBadRequest,
				Message: "Estado de pago inválido",
				Cause:   "El estado debe ser 'pagado', 'pendiente' o 'no pago'",
			}
			c.ServeJSON()
			return
		}
	}

	if in.MetodoPagoId == 0 {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "El campo metodoPagoId es obligatorio y debe ser un número válido"}
		c.ServeJSON()
		return
	}

	// Mapear a tu entidad de BD
	var updatedBy *string
	if in.UpdatedBy != "" {
		updatedBy = &in.UpdatedBy
	}

	pago := models.Pago{
		FECHA:             fecha,
		HORA:              hora,
		MONTO:             in.Monto,
		ESTADO_PAGO:       in.EstadoPago,   // e.g. "pagado"
		PK_ID_METODO_PAGO: in.MetodoPagoId, // FK
		UPDATED_BY:        updatedBy,       // opcional
		// UPDATED_AT se maneja con auto_now en el modelo
	}

	// Insertar
	if _, err := o.Insert(&pago); err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al crear el pago", Cause: err.Error()}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusCreated,
		Message: "Pago creado correctamente",
		Data:    pago, // tu MarshalJSON ya lo devuelve con fecha DD-MM-YYYY y updatedAt DD-MM-YYYY HH:mm:ss
	}
	c.ServeJSON()
}

// @Title Update
// @Summary Actualizar un pago
// @Description Actualiza los datos de un pago existente.
// @Tags pagos
// @Accept json
// @Produce json
// @Param   id    query    int  true   "ID del Pago"
// @Param   body  body   models.Pago true  "Datos del pago a actualizar"
// @Success 200 {object} models.Pago "Pago actualizado"
// @Failure 404 {object} models.ApiResponse "Pago no encontrado"
// @Security BearerAuth
// @Router /pagos [put]
func (c *PagoController) Put() {
	o := pagoNewOrm()

	// Obtener el ID del pago desde los parámetros
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

	// Buscar el pago por ID
	pago := models.Pago{PK_ID_PAGO: int64(id)}
	if err := o.Read(&pago); err == orm.ErrNoRows {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Pago no encontrado",
		}
		c.ServeJSON()
		return
	}

	// Deserializar los datos actualizados desde el cuerpo de la solicitud
	var input map[string]interface{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &input); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Error al decodificar la solicitud",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Validar y actualizar los campos
	if fechaStr, ok := input["FECHA"].(string); ok && fechaStr != "" {
		parsedDate, err := time.Parse("2006-01-02", fechaStr)
		if err != nil {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusBadRequest,
				Message: "Formato de fecha inválido",
				Cause:   err.Error(),
			}
			c.ServeJSON()
			return
		}
		pago.FECHA = parsedDate
	}

	// Procesar HORA
	if horaStr, ok := input["HORA"].(string); ok && horaStr != "" {
		// Validar el formato de HORA
		parsedHora, err := time.Parse("15:04:05", horaStr)
		if err != nil {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusBadRequest,
				Message: "Formato de hora inválido, debe ser HH:mm:ss",
				Cause:   err.Error(),
			}
			c.ServeJSON()
			return
		}
		pago.HORA = parsedHora
	} else {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El campo HORA no puede estar vacío",
		}
		c.ServeJSON()
		return
	}

	if monto, ok := input["MONTO"].(float64); ok {
		pago.MONTO = int64(monto)
	}

	if estado, ok := input["ESTADO_PAGO"].(string); ok && estado != "" {
		if !estadosPagoPermitidos[estado] {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusBadRequest,
				Message: "Estado de pago inválido. Debe ser 'pagado', 'pendiente' o 'no pago'",
			}
			c.ServeJSON()
			return
		}
		pago.ESTADO_PAGO = estado
	}

	if updatedBy, ok := input["UPDATED_BY"].(string); ok {
		pago.UPDATED_BY = &updatedBy
	}

	// Actualizar la fecha de modificación
	pago.UPDATED_AT = time.Now().UTC()

	if pkMetodoPago, ok := input["PK_ID_METODO_PAGO"].(float64); ok {
		valorMetodoPago := int64(pkMetodoPago)
		pago.PK_ID_METODO_PAGO = valorMetodoPago
	} else {
		// Opcional: Manejo de errores o acciones si el campo es obligatorio
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El campo PK_ID_METODO_PAGO es obligatorio y debe ser un número válido",
		}
		c.ServeJSON()
		return
	}

	// Guardar los cambios en la base de datos
	if _, err := o.Update(&pago); err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al actualizar el pago",
			Cause:   err.Error(),
		}
		c.ServeJSON()
		return
	}

	// Responder con los datos actualizados
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Pago actualizado correctamente",
		Data:    pago,
	}
	c.ServeJSON()
}

// @Title Delete
// @Summary Eliminar un pago
// @Description Elimina un pago de la base de datos.
// @Tags pagos
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID del Pago"
// @Success 200 {object} models.ApiResponse "Pago eliminado"
// @Failure 404 {object} models.ApiResponse "Pago no encontrado"
// @Security BearerAuth
// @Router /pagos [delete]
func (c *PagoController) Delete() {
	o := pagoNewOrm()

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

	pago := models.Pago{PK_ID_PAGO: int64(id)}

	if _, err := o.Delete(&pago); err == nil {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusOK,
			Message: "Pago eliminado",
		}
		c.ServeJSON()
	} else {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Pago no encontrado",
			Cause:   err.Error(),
		}
		c.ServeJSON()
	}
}
