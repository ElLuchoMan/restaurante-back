package pago

import (
	"encoding/json"
	"net/http"
	"restaurante/database"
	"restaurante/logging"
	"restaurante/models"
	"strconv"
	"strings"
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
	"PAGADO":    true,
	"PENDIENTE": true,
	"NO_PAGO":   true,
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
// @Param   estado   query   string   false   "Filtrar por estado del pago (PAGADO, PENDIENTE, NO_PAGO)"
// @Param   metodo_pago     query   int      false   "Filtrar por metodo de pago"
// @Success 200 {object} models.ApiResponse{data=[]models.Pago} "Lista de pagos"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /pagos [get]
func (c *PagoController) GetAll() {
	o := pagoNewOrm()
	var pagos []models.Pago

	_, err := o.QueryTable(new(models.Pago)).All(&pagos)
	if err != nil {
		logging.LogControllerError(c.Ctx, "pagos.getall.db_error", err, nil)
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener pagos de la base de datos",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
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
	estado := strings.ToUpper(c.GetString("estado"))
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
		if metodo_pago > 0 && (pago.PK_ID_METODO_PAGO == nil || pago.PK_ID_METODO_PAGO.PK_ID_METODO_PAGO != int64(metodo_pago)) {
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
		_ = c.ServeJSON()
		return
	}

	// Respuesta con los pagos filtrados
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Pagos obtenidos exitosamente",
		Data:    filteredPagos,
	}
	_ = c.ServeJSON()
}

// @Title GetById
// @Summary Obtener pago por ID
// @Description Devuelve un pago específico por ID.
// @Tags pagos
// @Accept json
// @Produce json
// @Param   id     query    int     true        "ID del Pago"
// @Success 200 {object} models.ApiResponse{data=models.Pago} "Pago encontrado"
// @Failure 404 {object} models.ApiResponse "Pago no encontrado"
// @Security BearerAuth
// @Router /pagos/search [get]
func (c *PagoController) GetById() {
	o := pagoNewOrm()
	id, err := c.GetInt("id")

	if err != nil || id == 0 {
		logging.LogControllerError(c.Ctx, "pagos.getbyid.bad_request", err, map[string]interface{}{"id": c.GetString("id")})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El parámetro 'id' es inválido o está ausente",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
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
		_ = c.ServeJSON()
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
	_ = c.ServeJSON()
}

// @Title Create
// @Summary Crear un nuevo pago
// @Description Crea un nuevo pago en la base de datos.
// @Tags pagos
// @Accept json
// @Produce json
// @Param   body  body   models.PagoCreateRequest true  "Datos del pago a crear (fecha YYYY-MM-DD, hora HH:MM:SS)"
// @Success 201 {object} models.ApiResponse{data=models.Pago} "Pago creado"
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
		logging.LogControllerError(c.Ctx, "pagos.post.bad_json", err, map[string]interface{}{"body": string(c.Ctx.Input.RequestBody)})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Error al decodificar la solicitud",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	// Validaciones y parseo
	if in.FechaPago == "" {
		logging.LogControllerError(c.Ctx, "pagos.post.validation_error", nil, map[string]interface{}{"missing": "fechaPago", "body": string(c.Ctx.Input.RequestBody)})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "El campo fechaPago no puede estar vacío"}
		_ = c.ServeJSON()
		return
	}
	fecha, err := time.Parse("2006-01-02", in.FechaPago)
	if err != nil {
		logging.LogControllerError(c.Ctx, "pagos.post.validation_error", err, map[string]interface{}{"fechaPago": in.FechaPago, "body": string(c.Ctx.Input.RequestBody)})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Formato de fecha inválido (use YYYY-MM-DD)", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}

	if in.HoraPago == "" {
		logging.LogControllerError(c.Ctx, "pagos.post.validation_error", nil, map[string]interface{}{"missing": "horaPago", "body": string(c.Ctx.Input.RequestBody)})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "El campo horaPago no puede estar vacío"}
		_ = c.ServeJSON()
		return
	}
	hora, err := time.Parse("15:04:05", in.HoraPago)
	if err != nil {
		logging.LogControllerError(c.Ctx, "pagos.post.validation_error", err, map[string]interface{}{"horaPago": in.HoraPago, "body": string(c.Ctx.Input.RequestBody)})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Formato de hora inválido, debe ser HH:mm:ss", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}

	if in.Monto == 0 {
		logging.LogControllerError(c.Ctx, "pagos.post.validation_error", nil, map[string]interface{}{"missing": "monto", "body": string(c.Ctx.Input.RequestBody)})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "El campo monto es obligatorio y debe ser un número"}
		_ = c.ServeJSON()
		return
	}

	if in.EstadoPago == "" {
		logging.LogControllerError(c.Ctx, "pagos.post.validation_error", nil, map[string]interface{}{"missing": "estadoPago", "body": string(c.Ctx.Input.RequestBody)})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "El campo estadoPago es obligatorio"}
		_ = c.ServeJSON()
		return
	}
	in.EstadoPago = strings.ToUpper(in.EstadoPago)
	if !estadosPagoPermitidos[in.EstadoPago] {
		logging.LogControllerError(c.Ctx, "pagos.post.validation_error", nil, map[string]interface{}{"estadoPago": in.EstadoPago, "body": string(c.Ctx.Input.RequestBody)})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "Estado de pago inválido",
			Cause:   "El estado debe ser 'PAGADO', 'PENDIENTE' o 'NO_PAGO'",
		}
		_ = c.ServeJSON()
		return
	}

	if in.MetodoPagoId == 0 {
		logging.LogControllerError(c.Ctx, "pagos.post.validation_error", nil, map[string]interface{}{"missing": "metodoPagoId", "body": string(c.Ctx.Input.RequestBody)})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "El campo metodoPagoId es obligatorio y debe ser un número válido"}
		_ = c.ServeJSON()
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
		ESTADO_PAGO:       in.EstadoPago,
		PK_ID_METODO_PAGO: &models.MetodoPago{PK_ID_METODO_PAGO: in.MetodoPagoId},
		UPDATED_BY:        updatedBy,
	}

	// Insertar
	if _, err := o.Insert(&pago); err != nil {
		logging.LogControllerError(c.Ctx, "pagos.post.insert_error", err, map[string]interface{}{"body": string(c.Ctx.Input.RequestBody)})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al crear el pago", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusCreated,
		Message: "Pago creado correctamente",
		Data:    pago,
	}
	_ = c.ServeJSON()
}

// @Title Update
// @Summary Actualizar un pago
// @Description Actualiza los datos de un pago existente.
// @Tags pagos
// @Accept json
// @Produce json
// @Param   id    query    int  true   "ID del Pago"
// @Param   body  body   models.PagoUpdateRequest true  "Datos del pago a actualizar (sólo campos a modificar, formatos: fecha YYYY-MM-DD, hora HH:MM:SS)"
// @Success 200 {object} models.ApiResponse{data=models.Pago} "Pago actualizado"
// @Failure 404 {object} models.ApiResponse "Pago no encontrado"
// @Security BearerAuth
// @Router /pagos [put]
func (c *PagoController) Put() {
	o := pagoNewOrm()

	idStr := c.GetString("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id == 0 {
		logging.LogControllerError(c.Ctx, "pagos.put.bad_request", err, map[string]interface{}{"id": idStr, "body": string(c.Ctx.Input.RequestBody)})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El parámetro 'id' es inválido o está ausente",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	pago := models.Pago{PK_ID_PAGO: int64(id)}
	if err := o.Read(&pago); err == orm.ErrNoRows {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Pago no encontrado",
		}
		_ = c.ServeJSON()
		return
	}

	var input map[string]interface{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &input); err != nil {
		logging.LogControllerError(c.Ctx, "pagos.put.bad_json", err, map[string]interface{}{"id": id, "body": string(c.Ctx.Input.RequestBody)})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Error al decodificar la solicitud", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}

	getStr := func(keys ...string) (string, bool) {
		for _, k := range keys {
			if v, ok := input[k].(string); ok && v != "" {
				return v, true
			}
		}
		return "", false
	}
	getFloat := func(keys ...string) (float64, bool) {
		for _, k := range keys {
			if v, ok := input[k].(float64); ok {
				return v, true
			}
		}
		return 0, false
	}

	if fechaStr, ok := getStr("fecha", "FECHA"); ok {
		parsedDate, err := time.Parse("2006-01-02", fechaStr)
		if err != nil {
			logging.LogControllerError(c.Ctx, "pagos.put.validation_error", err, map[string]interface{}{"id": id, "fecha": fechaStr, "body": string(c.Ctx.Input.RequestBody)})
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Formato de fecha inválido (YYYY-MM-DD)", Cause: err.Error()}
			_ = c.ServeJSON()
			return
		}
		pago.FECHA = parsedDate
	}

	if horaStr, ok := getStr("hora", "HORA"); ok {
		parsedHora, err := time.Parse("15:04:05", horaStr)
		if err != nil {
			logging.LogControllerError(c.Ctx, "pagos.put.validation_error", err, map[string]interface{}{"id": id, "hora": horaStr, "body": string(c.Ctx.Input.RequestBody)})
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Formato de hora inválido (HH:MM:SS)", Cause: err.Error()}
			_ = c.ServeJSON()
			return
		}
		pago.HORA = parsedHora
	} else {
		logging.LogControllerError(c.Ctx, "pagos.put.validation_error", nil, map[string]interface{}{"id": id, "missing": "hora", "body": string(c.Ctx.Input.RequestBody)})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "El campo hora no puede estar vacío"}
		_ = c.ServeJSON()
		return
	}

	if monto, ok := getFloat("monto", "MONTO"); ok {
		pago.MONTO = int64(monto)
	}

	if estado, ok := getStr("estadoPago", "ESTADO_PAGO"); ok {
		estado = strings.ToUpper(estado)
		if !estadosPagoPermitidos[estado] {
			logging.LogControllerError(c.Ctx, "pagos.put.validation_error", nil, map[string]interface{}{"id": id, "estado": estado, "body": string(c.Ctx.Input.RequestBody)})
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "Estado de pago inválido. Debe ser 'PAGADO', 'PENDIENTE' o 'NO_PAGO'"}
			_ = c.ServeJSON()
			return
		}
		pago.ESTADO_PAGO = estado
	}

	if updatedBy, ok := getStr("updatedBy", "UPDATED_BY"); ok {
		pago.UPDATED_BY = &updatedBy
	}

	pago.UPDATED_AT = time.Now().UTC()

	if v, ok := getFloat("metodoPagoId", "PK_ID_METODO_PAGO"); ok && v != 0 {
		pago.PK_ID_METODO_PAGO = &models.MetodoPago{PK_ID_METODO_PAGO: int64(v)}
	} else {
		logging.LogControllerError(c.Ctx, "pagos.put.validation_error", nil, map[string]interface{}{"id": id, "missing": "metodoPagoId", "body": string(c.Ctx.Input.RequestBody)})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{Code: http.StatusBadRequest, Message: "El campo metodoPagoId es obligatorio y debe ser un número válido"}
		_ = c.ServeJSON()
		return
	}

	if _, err := o.Update(&pago); err != nil {
		logging.LogControllerError(c.Ctx, "pagos.put.update_error", err, map[string]interface{}{"id": id, "body": string(c.Ctx.Input.RequestBody)})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{Code: http.StatusInternalServerError, Message: "Error al actualizar el pago", Cause: err.Error()}
		_ = c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{Code: http.StatusOK, Message: "Pago actualizado correctamente", Data: pago}
	_ = c.ServeJSON()
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
		logging.LogControllerError(c.Ctx, "pagos.delete.bad_request", err, map[string]interface{}{"id": idStr})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El parámetro 'id' es inválido o está ausente",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	pago := models.Pago{PK_ID_PAGO: int64(id)}

	if _, err := o.Delete(&pago); err == nil {
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusOK,
			Message: "Pago eliminado",
		}
		_ = c.ServeJSON()
	} else {
		logging.LogControllerError(c.Ctx, "pagos.delete.delete_error", err, map[string]interface{}{"id": id})
		c.Ctx.Output.SetStatus(http.StatusOK)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusNotFound,
			Message: "Pago no encontrado",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
	}
}
