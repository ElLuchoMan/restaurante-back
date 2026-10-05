package oferta

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"restaurante/internal/authz"
	"restaurante/internal/dberr"
	"restaurante/internal/httpx"
	"restaurante/logging"
	"restaurante/models"
	"restaurante/services"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

// OfertaController gestiona /ofertas. En las respuestas `restauranteId` es el
// objeto restaurante completo (ver models.OfertaDoc).
type OfertaController struct {
	web.Controller
}

const (
	msgIDInvalido   = "El parámetro 'id' es inválido o está ausente"
	msgBadJSON      = "JSON inválido"
	msgValidacion   = "Error de validación"
	msgNoEncontrada = "Oferta no encontrada"
)

// nullableUpdate son los campos de PUT que admiten null explícito (se limpian).
var nullableUpdate = []string{"horaInicio", "horaFin"}

// porID lee la oferta con su restaurante (RelatedSel) y deserializa los días.
func porID(id int64) (*models.Oferta, error) {
	o := &models.Oferta{}
	err := orm.NewOrm().QueryTable(new(models.Oferta)).Filter("pk_id_oferta", id).RelatedSel("PkIdRestaurante").One(o)
	if err != nil {
		return nil, err
	}
	o.AfterLoad()
	return o, nil
}

// load valida `id` y lee la oferta (400 / 404 / 500).
func (c *OfertaController) load(op string) (*models.Oferta, bool) {
	id, err := httpx.PositiveInt64Param(&c.Controller, "id")
	if err != nil {
		logging.LogControllerError(c.Ctx, "ofertas."+op+".bad_request", err, map[string]interface{}{"id": c.GetString("id")})
		httpx.Fail(&c.Controller, http.StatusBadRequest, msgIDInvalido, err)
		return nil, false
	}
	o, err := porID(id)
	if err != nil {
		c.readError(op, id, err)
		return nil, false
	}
	return o, true
}

func (c *OfertaController) readError(op string, id int64, err error) {
	if errors.Is(err, orm.ErrNoRows) {
		httpx.Fail(&c.Controller, http.StatusNotFound, msgNoEncontrada, nil)
		return
	}
	logging.LogControllerError(c.Ctx, "ofertas."+op+".read_error", err, map[string]interface{}{"id": id})
	httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error interno del servidor", err)
}

// writeError: unicidad -> 409, FK (restaurante inexistente) -> 400, otro -> 500.
func (c *OfertaController) writeError(op, msg string, err error) {
	logging.LogControllerError(c.Ctx, "ofertas."+op, err, nil)
	switch {
	case dberr.IsUnique(err):
		httpx.Fail(&c.Controller, http.StatusConflict, "Ya existe una oferta con ese título", err)
	case dberr.IsForeignKey(err):
		httpx.Fail(&c.Controller, http.StatusBadRequest, "El restaurante indicado no existe", err)
	default:
		httpx.Fail(&c.Controller, http.StatusInternalServerError, msg, err)
	}
}

// invalid responde 422 con el motivo.
func (c *OfertaController) invalid(op, msg string, err error) {
	logging.LogControllerError(c.Ctx, "ofertas."+op, err, map[string]interface{}{"motivo": msg})
	httpx.Fail(&c.Controller, http.StatusUnprocessableEntity, msg, err)
}

const (
	msgTipo       = "Tipo de descuento no válido - debe ser PORCENTAJE o MONTO"
	msgFechaIni   = "Fecha de inicio inválida - debe tener formato YYYY-MM-DD"
	msgFechaFin   = "Fecha de fin inválida - debe tener formato YYYY-MM-DD"
	msgHoraIni    = "Hora de inicio inválida - debe tener formato HH:MM o HH:MM:SS"
	msgHoraFin    = "Hora de fin inválida - debe tener formato HH:MM o HH:MM:SS"
	msgTituloVac  = "El título es obligatorio"
	msgRestaurant = "restauranteId debe ser un entero positivo"
)

// @Title GetAll
// @Summary Obtener todas las ofertas
// @Description Lista paginada (más recientes primero). `data.data` es la lista de ofertas (`[]` si no hay) y cada una trae `restauranteId` como objeto restaurante. `limit` por defecto 20 (máximo 100) y `offset` por defecto 0.
// @Tags ofertas
// @Accept json
// @Produce json
// @Param activo query bool false "Filtrar por estado activo (true/false)"
// @Param restaurante_id query int false "ID del restaurante (entero positivo)"
// @Param titulo query string false "Filtrar por título (contiene, sin distinguir mayúsculas)"
// @Param limit query int false "Límite de resultados (1-100, por defecto 20; valores mayores se limitan a 100)"
// @Param offset query int false "Offset para paginación (>= 0, por defecto 0)"
// @Success 200 {object} models.ApiResponse{data=models.OfertaPaginadaDoc} "Ofertas obtenidas"
// @Failure 400 {object} models.ApiResponse "Parámetros de filtro o paginación inválidos"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /ofertas [get]
func (c *OfertaController) GetAll() {
	qs := orm.NewOrm().QueryTable(new(models.Oferta))

	if v := strings.TrimSpace(c.GetString("activo")); v != "" {
		activo, err := strconv.ParseBool(v)
		if err != nil {
			httpx.Fail(&c.Controller, http.StatusBadRequest, "El parámetro 'activo' debe ser true o false", err)
			return
		}
		qs = qs.Filter("activo", activo)
	}
	if strings.TrimSpace(c.GetString("restaurante_id")) != "" {
		id, err := httpx.PositiveInt64Param(&c.Controller, "restaurante_id")
		if err != nil {
			httpx.Fail(&c.Controller, http.StatusBadRequest, "El parámetro 'restaurante_id' es inválido", err)
			return
		}
		qs = qs.Filter("pk_id_restaurante", id)
	}
	if titulo := strings.TrimSpace(c.GetString("titulo")); titulo != "" {
		qs = qs.Filter("titulo__icontains", titulo)
	}

	limit, err := c.GetInt("limit", 20)
	if err != nil || limit < 1 {
		httpx.Fail(&c.Controller, http.StatusBadRequest, "El parámetro 'limit' debe ser un entero positivo", err)
		return
	}
	offset, err := c.GetInt("offset", 0)
	if err != nil || offset < 0 {
		httpx.Fail(&c.Controller, http.StatusBadRequest, "El parámetro 'offset' debe ser un entero mayor o igual a 0", err)
		return
	}
	if limit > 100 {
		limit = 100
	}

	total, err := qs.Count()
	if err != nil {
		logging.LogControllerError(c.Ctx, "ofertas.getall.count_error", err, nil)
		httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error al obtener ofertas", err)
		return
	}

	var ofertas []*models.Oferta
	if _, err = qs.RelatedSel("PkIdRestaurante").OrderBy("-pk_id_oferta").Limit(limit).Offset(int64(offset)).All(&ofertas); err != nil {
		logging.LogControllerError(c.Ctx, "ofertas.getall.query_error", err, nil)
		httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error al obtener ofertas", err)
		return
	}
	for _, o := range ofertas {
		o.AfterLoad()
	}

	httpx.Send(&c.Controller, http.StatusOK, "Ofertas obtenidas exitosamente", models.PaginatedResponse{
		Data:       httpx.List(ofertas),
		Total:      total,
		Page:       offset/limit + 1,
		PageSize:   limit,
		TotalPages: int((total + int64(limit) - 1) / int64(limit)),
	})
}

// parseHora interpreta una hora HH:MM[:SS].
func parseHora(s string) (*time.Time, error) {
	h, err := models.ParseTimeToUTC(s)
	if err != nil {
		return nil, err
	}
	return &h, nil
}

// validar aplica las reglas de negocio de la oferta y responde 422 si fallan.
func (c *OfertaController) validar(op string, o *models.Oferta) bool {
	if strings.TrimSpace(o.Titulo) == "" {
		c.invalid(op+".titulo", msgTituloVac, nil)
		return false
	}
	if o.PkIdRestaurante == nil || o.PkIdRestaurante.PK_ID_RESTAURANTE <= 0 {
		c.invalid(op+".restaurante", msgRestaurant, nil)
		return false
	}
	if err := services.NewOfertaService(nil).ValidarReglasNegocioOferta(o); err != nil {
		c.invalid(op+".validation_error", msgValidacion, err)
		return false
	}
	return true
}

// @Title Post
// @Summary Crear oferta
// @Description Solo Administrador. Crea una oferta activa. Errores de validación de negocio (tipo, fechas, horas, porcentaje 1-100, días válidos, título, restauranteId) responden 422; un `restauranteId` inexistente responde 400 y un título repetido 409. Fechas YYYY-MM-DD, horas HH:MM o HH:MM:SS; `diasSemana` vacío significa todos los días. Devuelve la oferta con `restauranteId` como objeto restaurante.
// @Tags ofertas
// @Accept json
// @Produce json
// @Param body body models.CrearOfertaRequest true "Datos de la oferta"
// @Success 201 {object} models.ApiResponse{data=models.OfertaDoc} "Oferta creada"
// @Failure 400 {object} models.ApiResponse "JSON inválido o restaurante inexistente"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 403 {object} models.ApiResponse "Se requiere rol Administrador"
// @Failure 409 {object} models.ApiResponse "Ya existe una oferta con ese título"
// @Failure 422 {object} models.ApiResponse "Error de validación"
// @Failure 500 {object} models.ApiResponse "Error al crear la oferta"
// @Security BearerAuth
// @Router /ofertas [post]
func (c *OfertaController) Post() {
	if _, ok := authz.RequireAdmin(&c.Controller); !ok {
		return
	}
	var req models.CrearOfertaRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		logging.LogControllerError(c.Ctx, "ofertas.post.bad_json", err, nil)
		httpx.Fail(&c.Controller, http.StatusBadRequest, msgBadJSON, err)
		return
	}
	if !req.TipoDescuento.IsValid() {
		c.invalid("post.invalid_tipo_descuento", msgTipo, nil)
		return
	}
	fechaInicio, err := models.ParseDateToNoonUTC(req.FechaInicio)
	if err != nil {
		c.invalid("post.invalid_fecha_inicio", msgFechaIni, err)
		return
	}
	fechaFin, err := models.ParseDateToNoonUTC(req.FechaFin)
	if err != nil {
		c.invalid("post.invalid_fecha_fin", msgFechaFin, err)
		return
	}
	oferta := &models.Oferta{
		Titulo:          strings.TrimSpace(req.Titulo),
		TipoDescuento:   req.TipoDescuento,
		ValorDescuento:  req.ValorDescuento,
		FechaInicio:     fechaInicio,
		FechaFin:        fechaFin,
		DiasSemanaArray: req.DiasSemana,
		Activo:          true,
		PkIdRestaurante: &models.Restaurante{PK_ID_RESTAURANTE: req.PkIdRestaurante},
	}
	if req.HoraInicio != nil {
		if oferta.HoraInicio, err = parseHora(*req.HoraInicio); err != nil {
			c.invalid("post.invalid_hora_inicio", msgHoraIni, err)
			return
		}
	}
	if req.HoraFin != nil {
		if oferta.HoraFin, err = parseHora(*req.HoraFin); err != nil {
			c.invalid("post.invalid_hora_fin", msgHoraFin, err)
			return
		}
	}
	if !c.validar("post", oferta) {
		return
	}

	oferta.BeforeInsert()
	if _, err := orm.NewOrm().Insert(oferta); err != nil {
		c.writeError("post.insert_error", "Error al crear oferta", err)
		return
	}
	creada, err := porID(oferta.PkIdOferta)
	if err != nil {
		c.readError("post", oferta.PkIdOferta, err)
		return
	}
	httpx.Send(&c.Controller, http.StatusCreated, "Oferta creada exitosamente", creada)
}

// @Title GetById
// @Summary Obtener oferta por ID
// @Tags ofertas
// @Accept json
// @Produce json
// @Param id query int true "ID de la oferta (entero positivo)"
// @Success 200 {object} models.ApiResponse{data=models.OfertaDoc} "Oferta encontrada"
// @Failure 400 {object} models.ApiResponse "Parámetro 'id' inválido o ausente"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 404 {object} models.ApiResponse "Oferta no encontrada"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /ofertas/search [get]
func (c *OfertaController) GetById() {
	oferta, ok := c.load("getbyid")
	if !ok {
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, "Oferta encontrada", oferta)
}

// aplicar copia sobre o los campos presentes de req. Devuelve false (ya
// respondido) si algún valor es inválido.
func (c *OfertaController) aplicar(body []byte, req *models.ActualizarOfertaRequest, o *models.Oferta) bool {
	var err error
	if req.Titulo != nil {
		o.Titulo = strings.TrimSpace(*req.Titulo)
	}
	if req.TipoDescuento != nil {
		if !req.TipoDescuento.IsValid() {
			c.invalid("put.invalid_tipo_descuento", msgTipo, nil)
			return false
		}
		o.TipoDescuento = *req.TipoDescuento
	}
	if req.ValorDescuento != nil {
		o.ValorDescuento = *req.ValorDescuento
	}
	if req.FechaInicio != nil {
		if o.FechaInicio, err = models.ParseDateToNoonUTC(*req.FechaInicio); err != nil {
			c.invalid("put.invalid_fecha_inicio", msgFechaIni, err)
			return false
		}
	}
	if req.FechaFin != nil {
		if o.FechaFin, err = models.ParseDateToNoonUTC(*req.FechaFin); err != nil {
			c.invalid("put.invalid_fecha_fin", msgFechaFin, err)
			return false
		}
	}
	if req.DiasSemana != nil {
		o.DiasSemanaArray = req.DiasSemana
	}
	if req.HoraInicio != nil {
		if o.HoraInicio, err = parseHora(*req.HoraInicio); err != nil {
			c.invalid("put.invalid_hora_inicio", msgHoraIni, err)
			return false
		}
	} else if httpx.IsNull(body, "horaInicio") {
		o.HoraInicio = nil
	}
	if req.HoraFin != nil {
		if o.HoraFin, err = parseHora(*req.HoraFin); err != nil {
			c.invalid("put.invalid_hora_fin", msgHoraFin, err)
			return false
		}
	} else if httpx.IsNull(body, "horaFin") {
		o.HoraFin = nil
	}
	if req.PkIdRestaurante != nil {
		o.PkIdRestaurante = &models.Restaurante{PK_ID_RESTAURANTE: *req.PkIdRestaurante}
	}
	if req.Activo != nil {
		o.Activo = *req.Activo
	}
	return true
}

// @Title Put
// @Summary Actualizar oferta
// @Description Solo Administrador. Actualización parcial (merge): los campos ausentes se conservan (cuerpo `models.ActualizarOfertaRequest`). `horaInicio` y `horaFin` admiten null explícito para quitar el horario (deben limpiarse juntos); null en cualquier otro campo responde 400. `diasSemana: []` significa todos los días. `activo` permite reactivar una oferta desactivada. Un cuerpo sin cambios responde 200. Validación de negocio incumplida: 422.
// @Tags ofertas
// @Accept json
// @Produce json
// @Param id query int true "ID de la oferta (entero positivo)"
// @Param body body models.ActualizarOfertaRequest true "Campos a modificar"
// @Success 200 {object} models.ApiResponse{data=models.OfertaDoc} "Oferta actualizada"
// @Failure 400 {object} models.ApiResponse "id o JSON inválido, null en campo no anulable, restaurante inexistente"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 403 {object} models.ApiResponse "Se requiere rol Administrador"
// @Failure 404 {object} models.ApiResponse "Oferta no encontrada"
// @Failure 409 {object} models.ApiResponse "Ya existe una oferta con ese título"
// @Failure 422 {object} models.ApiResponse "Error de validación"
// @Failure 500 {object} models.ApiResponse "Error al actualizar la oferta"
// @Security BearerAuth
// @Router /ofertas [put]
func (c *OfertaController) Put() {
	if _, ok := authz.RequireAdmin(&c.Controller); !ok {
		return
	}
	oferta, ok := c.load("put")
	if !ok {
		return
	}
	body := c.Ctx.Input.RequestBody
	var req models.ActualizarOfertaRequest
	if err := httpx.DecodeMerge(body, &req, nullableUpdate...); err != nil {
		logging.LogControllerError(c.Ctx, "ofertas.put.bad_json", err, map[string]interface{}{"id": oferta.PkIdOferta})
		httpx.Fail(&c.Controller, http.StatusBadRequest, msgBadJSON, err)
		return
	}
	if !c.aplicar(body, &req, oferta) || !c.validar("put", oferta) {
		return
	}

	oferta.BeforeUpdate()
	if _, err := orm.NewOrm().Update(oferta); err != nil {
		c.writeError("put.update_error", "Error al actualizar oferta", err)
		return
	}
	actualizada, err := porID(oferta.PkIdOferta)
	if err != nil {
		c.readError("put", oferta.PkIdOferta, err)
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, "Oferta actualizada exitosamente", actualizada)
}

// @Title Delete
// @Summary Desactivar oferta
// @Description Solo Administrador. No elimina la fila: desactiva la oferta (`activo = false`; se reactiva con PUT `activo: true`). Si ya estaba desactivada responde 400.
// @Tags ofertas
// @Accept json
// @Produce json
// @Param id query int true "ID de la oferta (entero positivo)"
// @Success 200 {object} models.ApiResponse "Oferta desactivada"
// @Failure 400 {object} models.ApiResponse "Parámetro 'id' inválido o oferta ya desactivada"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 403 {object} models.ApiResponse "Se requiere rol Administrador"
// @Failure 404 {object} models.ApiResponse "Oferta no encontrada"
// @Failure 500 {object} models.ApiResponse "Error al desactivar la oferta"
// @Security BearerAuth
// @Router /ofertas [delete]
func (c *OfertaController) Delete() {
	if _, ok := authz.RequireAdmin(&c.Controller); !ok {
		return
	}
	oferta, ok := c.load("delete")
	if !ok {
		return
	}
	if !oferta.Activo {
		httpx.Fail(&c.Controller, http.StatusBadRequest, "La oferta ya está desactivada", nil)
		return
	}
	oferta.Activo = false
	if _, err := orm.NewOrm().Update(oferta, "Activo"); err != nil {
		logging.LogControllerError(c.Ctx, "ofertas.delete.update_error", err, map[string]interface{}{"id": oferta.PkIdOferta})
		httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error al desactivar oferta", err)
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, "Oferta desactivada exitosamente", nil)
}

// @Title ObtenerOfertasActivas
// @Summary Obtener ofertas activas
// @Description Ruta pública (sin token). Ofertas activas del restaurante vigentes en la fecha y hora indicadas (por defecto, ahora en hora de Bogotá), con los ids de sus productos (`productosIds`, `[]` si no tiene). Con `producto_id` solo las que incluyen ese producto. Sin resultados, `data` es `[]`.
// @Tags ofertas
// @Accept json
// @Produce json
// @Param restaurante_id query int true "ID del restaurante (entero positivo)"
// @Param fecha query string false "Fecha a consultar (YYYY-MM-DD, por defecto hoy)"
// @Param hora query string false "Hora a consultar (HH:MM o HH:MM:SS, por defecto ahora)"
// @Param producto_id query int false "ID del producto específico (entero positivo)"
// @Success 200 {object} models.ApiResponse{data=[]models.OfertaActivaResponse} "Ofertas activas (puede ser vacío)"
// @Failure 400 {object} models.ApiResponse "Parámetros inválidos"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Router /ofertas/activas [get]
func (c *OfertaController) ObtenerOfertasActivas() {
	restauranteID, err := httpx.PositiveInt64Param(&c.Controller, "restaurante_id")
	if err != nil {
		logging.LogControllerError(c.Ctx, "ofertas.activas.bad_request", err, map[string]interface{}{"restaurante_id": c.GetString("restaurante_id")})
		httpx.Fail(&c.Controller, http.StatusBadRequest, "restaurante_id es requerido y debe ser un entero positivo", err)
		return
	}

	var fecha, hora *time.Time
	if s := strings.TrimSpace(c.GetString("fecha")); s != "" {
		f, err := models.ParseDateToNoonUTC(s)
		if err != nil {
			logging.LogControllerError(c.Ctx, "ofertas.activas.invalid_fecha", err, map[string]interface{}{"fecha": s})
			httpx.Fail(&c.Controller, http.StatusBadRequest, "Fecha inválida - debe tener formato YYYY-MM-DD", err)
			return
		}
		fecha = &f
	}
	if s := strings.TrimSpace(c.GetString("hora")); s != "" {
		h, err := models.ParseTimeToUTC(s)
		if err != nil {
			logging.LogControllerError(c.Ctx, "ofertas.activas.invalid_hora", err, map[string]interface{}{"hora": s})
			httpx.Fail(&c.Controller, http.StatusBadRequest, "Hora inválida - debe tener formato HH:MM o HH:MM:SS", err)
			return
		}
		hora = &h
	}
	var productoID *int64
	if strings.TrimSpace(c.GetString("producto_id")) != "" {
		id, err := httpx.PositiveInt64Param(&c.Controller, "producto_id")
		if err != nil {
			httpx.Fail(&c.Controller, http.StatusBadRequest, "El parámetro 'producto_id' es inválido", err)
			return
		}
		productoID = &id
	}

	ofertas, err := services.NewOfertaService(orm.NewOrm()).ObtenerOfertasActivas(c.Ctx.Request.Context(), restauranteID, fecha, hora, productoID)
	if err != nil {
		logging.LogControllerError(c.Ctx, "ofertas.activas.service_error", err, map[string]interface{}{"restaurante_id": restauranteID})
		httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error al obtener ofertas activas", err)
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, "Ofertas activas obtenidas exitosamente", httpx.List(ofertas))
}

// @Title AsociarProducto
// @Summary Asociar producto a oferta
// @Description Solo Administrador. Asocia un producto existente a una oferta existente. 404 si no existe la oferta o el producto; 409 si ya estaban asociados. `data` devuelve `{ofertaId, productoId}`.
// @Tags ofertas
// @Accept json
// @Produce json
// @Param id query int true "ID de la oferta (entero positivo)"
// @Param body body models.AsociarProductoOfertaRequest true "ID del producto"
// @Success 201 {object} models.ApiResponse{data=models.OfertaProductoAsociacionDoc} "Producto asociado"
// @Failure 400 {object} models.ApiResponse "id o JSON inválido, productoId no positivo"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 403 {object} models.ApiResponse "Se requiere rol Administrador"
// @Failure 404 {object} models.ApiResponse "Oferta o producto no encontrado"
// @Failure 409 {object} models.ApiResponse "El producto ya está asociado a la oferta"
// @Failure 500 {object} models.ApiResponse "Error al asociar el producto"
// @Security BearerAuth
// @Router /ofertas/productos [post]
func (c *OfertaController) AsociarProducto() {
	if _, ok := authz.RequireAdmin(&c.Controller); !ok {
		return
	}
	ofertaID, err := httpx.PositiveInt64Param(&c.Controller, "id")
	if err != nil {
		logging.LogControllerError(c.Ctx, "ofertas.asociar.bad_request", err, map[string]interface{}{"id": c.GetString("id")})
		httpx.Fail(&c.Controller, http.StatusBadRequest, msgIDInvalido, err)
		return
	}
	var req models.AsociarProductoOfertaRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		logging.LogControllerError(c.Ctx, "ofertas.asociar.bad_json", err, map[string]interface{}{"id": ofertaID})
		httpx.Fail(&c.Controller, http.StatusBadRequest, msgBadJSON, err)
		return
	}
	if req.ProductoId <= 0 {
		httpx.Fail(&c.Controller, http.StatusBadRequest, "productoId debe ser un entero positivo", nil)
		return
	}

	o := orm.NewOrm()
	if err := o.Read(&models.Oferta{PkIdOferta: ofertaID}); err != nil {
		c.assocReadError("oferta", msgNoEncontrada, ofertaID, err)
		return
	}
	if err := o.Read(&models.Producto{PK_ID_PRODUCTO: req.ProductoId}); err != nil {
		c.assocReadError("producto", "Producto no encontrado", req.ProductoId, err)
		return
	}

	// Raw: OfertaProducto no tiene llave primaria propia y orm.Insert no la soporta.
	if _, err := o.Raw("INSERT INTO oferta_producto (pk_id_oferta, pk_id_producto) VALUES (?, ?)", ofertaID, req.ProductoId).Exec(); err != nil {
		logging.LogControllerError(c.Ctx, "ofertas.asociar.insert_error", err, map[string]interface{}{"oferta_id": ofertaID, "producto_id": req.ProductoId})
		if dberr.IsUnique(err) {
			httpx.Fail(&c.Controller, http.StatusConflict, "El producto ya está asociado a esta oferta", err)
			return
		}
		httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error al asociar producto", err)
		return
	}
	httpx.Send(&c.Controller, http.StatusCreated, "Producto asociado correctamente", models.OfertaProductoAsociacionDoc{OfertaId: ofertaID, ProductoId: req.ProductoId})
}

func (c *OfertaController) assocReadError(what, notFound string, id int64, err error) {
	if errors.Is(err, orm.ErrNoRows) {
		httpx.Fail(&c.Controller, http.StatusNotFound, notFound, nil)
		return
	}
	logging.LogControllerError(c.Ctx, "ofertas.asociar.read_"+what+"_error", err, map[string]interface{}{"id": id})
	httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error interno del servidor", err)
}

// @Title DesasociarProducto
// @Summary Desasociar producto de oferta
// @Description Solo Administrador. Elimina la asociación entre la oferta y el producto. 404 si la asociación no existe.
// @Tags ofertas
// @Accept json
// @Produce json
// @Param id query int true "ID de la oferta (entero positivo)"
// @Param producto_id query int true "ID del producto (entero positivo)"
// @Success 200 {object} models.ApiResponse "Producto desasociado"
// @Failure 400 {object} models.ApiResponse "id o producto_id inválido o ausente"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 403 {object} models.ApiResponse "Se requiere rol Administrador"
// @Failure 404 {object} models.ApiResponse "Asociación no encontrada"
// @Failure 500 {object} models.ApiResponse "Error al desasociar el producto"
// @Security BearerAuth
// @Router /ofertas/productos [delete]
func (c *OfertaController) DesasociarProducto() {
	if _, ok := authz.RequireAdmin(&c.Controller); !ok {
		return
	}
	ofertaID, err := httpx.PositiveInt64Param(&c.Controller, "id")
	if err != nil {
		logging.LogControllerError(c.Ctx, "ofertas.desasociar.bad_oferta_id", err, map[string]interface{}{"id": c.GetString("id")})
		httpx.Fail(&c.Controller, http.StatusBadRequest, "ID de oferta inválido o ausente", err)
		return
	}
	productoID, err := httpx.PositiveInt64Param(&c.Controller, "producto_id")
	if err != nil {
		logging.LogControllerError(c.Ctx, "ofertas.desasociar.bad_producto_id", err, map[string]interface{}{"producto_id": c.GetString("producto_id")})
		httpx.Fail(&c.Controller, http.StatusBadRequest, "ID de producto inválido o ausente", err)
		return
	}

	res, err := orm.NewOrm().Raw("DELETE FROM oferta_producto WHERE pk_id_oferta = ? AND pk_id_producto = ?", ofertaID, productoID).Exec()
	if err != nil {
		logging.LogControllerError(c.Ctx, "ofertas.desasociar.delete_error", err, map[string]interface{}{"oferta_id": ofertaID, "producto_id": productoID})
		httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error al desasociar producto", err)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		httpx.Fail(&c.Controller, http.StatusNotFound, "Asociación no encontrada", nil)
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, "Producto desasociado correctamente", nil)
}
