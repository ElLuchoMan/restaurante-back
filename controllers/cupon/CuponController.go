package cupon

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"restaurante/internal/dberr"
	"restaurante/internal/httpx"
	"restaurante/logging"
	"restaurante/models"
	"restaurante/services"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

// CuponController gestiona /cupones. En las respuestas `productoId`,
// `categoriaId` y `documentoCliente` son los objetos relacionados completos
// (nunca incluyen contraseñas ni la imagen del producto) y se omiten si el
// cupón no los tiene.
type CuponController struct {
	web.Controller
}

const (
	msgIDInvalido  = "El parámetro 'id' es inválido o está ausente"
	msgBadJSON     = "JSON inválido"
	msgValidacion  = "Error de validación"
	msgNoEncontrao = "Cupón no encontrado"
	msgScope       = "Scope no válido - debe ser GLOBAL, PRODUCTO, CATEGORIA o CLIENTE"
	msgTipo        = "Tipo de descuento no válido - debe ser PORCENTAJE o MONTO"
	msgFechaIni    = "Fecha de inicio inválida - debe tener formato YYYY-MM-DD"
	msgFechaFin    = "Fecha de fin inválida - debe tener formato YYYY-MM-DD"
	msgCodigo      = "El código es obligatorio y debe tener entre 3 y 50 caracteres"
)

// nullableUpdate son los campos de PUT que admiten null explícito (se limpian).
var nullableUpdate = []string{"maxUsos", "limitePorCliente", "montoMinimo", "productoId", "categoriaId", "documentoCliente"}

// relaciones son las relaciones que se cargan en las respuestas.
var relaciones = []interface{}{"PkIdProducto", "PkIdCategoria", "PkDocumentoCliente"}

// limpiar quita de las relaciones cargadas lo que no debe viajar en la respuesta.
func limpiar(cp *models.Cupon) {
	if cp.PkIdProducto != nil {
		cp.PkIdProducto.IMAGEN = ""
	}
	if cp.PkDocumentoCliente != nil {
		cp.PkDocumentoCliente.PASSWORD = ""
	}
}

// buscar lee un cupón (con sus relaciones) por campo = valor.
func buscar(campo string, valor interface{}) (*models.Cupon, error) {
	cp := &models.Cupon{}
	if err := orm.NewOrm().QueryTable(new(models.Cupon)).Filter(campo, valor).RelatedSel(relaciones...).One(cp); err != nil {
		return nil, err
	}
	limpiar(cp)
	return cp, nil
}

func porID(id int64) (*models.Cupon, error) { return buscar("pk_id_cupon", id) }

// load valida `id` y lee el cupón (400 / 404 / 500).
func (c *CuponController) load(op string) (*models.Cupon, bool) {
	id, err := httpx.PositiveInt64Param(&c.Controller, "id")
	if err != nil {
		logging.LogControllerError(c.Ctx, "cupones."+op+".bad_request", err, map[string]interface{}{"id": c.GetString("id")})
		httpx.Fail(&c.Controller, http.StatusBadRequest, msgIDInvalido, err)
		return nil, false
	}
	cp, err := porID(id)
	if err != nil {
		c.readError(op, id, err)
		return nil, false
	}
	return cp, true
}

func (c *CuponController) readError(op string, id interface{}, err error) {
	if errors.Is(err, orm.ErrNoRows) {
		httpx.Fail(&c.Controller, http.StatusNotFound, msgNoEncontrao, nil)
		return
	}
	logging.LogControllerError(c.Ctx, "cupones."+op+".read_error", err, map[string]interface{}{"id": id})
	httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error interno del servidor", err)
}

// writeError: unicidad (código repetido) -> 409, FK (producto, categoría o
// cliente inexistente) -> 400, otro -> 500.
func (c *CuponController) writeError(op, msg string, err error) {
	logging.LogControllerError(c.Ctx, "cupones."+op, err, nil)
	switch {
	case dberr.IsUnique(err):
		httpx.Fail(&c.Controller, http.StatusConflict, "Ya existe un cupón con ese código", err)
	case dberr.IsForeignKey(err):
		httpx.Fail(&c.Controller, http.StatusBadRequest, "El producto, la categoría o el cliente indicado no existe", err)
	default:
		httpx.Fail(&c.Controller, http.StatusInternalServerError, msg, err)
	}
}

func (c *CuponController) invalid(op, msg string, err error) {
	logging.LogControllerError(c.Ctx, "cupones."+op, err, map[string]interface{}{"motivo": msg})
	httpx.Fail(&c.Controller, http.StatusUnprocessableEntity, msg, err)
}

// paginacion lee limit (1-100, por defecto 20) y offset (>= 0).
func (c *CuponController) paginacion() (limit, offset int, ok bool) {
	limit, err := c.GetInt("limit", 20)
	if err != nil || limit < 1 {
		httpx.Fail(&c.Controller, http.StatusBadRequest, "El parámetro 'limit' debe ser un entero positivo", err)
		return 0, 0, false
	}
	offset, err = c.GetInt("offset", 0)
	if err != nil || offset < 0 {
		httpx.Fail(&c.Controller, http.StatusBadRequest, "El parámetro 'offset' debe ser un entero mayor o igual a 0", err)
		return 0, 0, false
	}
	if limit > 100 {
		limit = 100
	}
	return limit, offset, true
}

func pagina[T any](items []T, total int64, limit, offset int) models.PaginatedResponse {
	return models.PaginatedResponse{
		Data:       httpx.List(items),
		Total:      total,
		Page:       offset/limit + 1,
		PageSize:   limit,
		TotalPages: int((total + int64(limit) - 1) / int64(limit)),
	}
}

// @Title GetAll
// @Summary Obtener todos los cupones
// @Description Lista paginada (más recientes primero). `data.data` es la lista de cupones (`[]` si no hay). `limit` por defecto 20 (máximo 100) y `offset` por defecto 0. `fecha_desde` filtra por `fechaInicio >=` y `fecha_hasta` por `fechaFin <=`.
// @Tags cupones
// @Accept json
// @Produce json
// @Param activo query bool false "Filtrar por estado activo (true/false)"
// @Param codigo query string false "Filtrar por código (contiene, sin distinguir mayúsculas)"
// @Param scope query string false "Filtrar por scope" Enums(GLOBAL, PRODUCTO, CATEGORIA, CLIENTE)
// @Param fecha_desde query string false "Fecha de inicio desde (YYYY-MM-DD)"
// @Param fecha_hasta query string false "Fecha de fin hasta (YYYY-MM-DD)"
// @Param limit query int false "Límite de resultados (1-100, por defecto 20)"
// @Param offset query int false "Offset para paginación (>= 0, por defecto 0)"
// @Success 200 {object} models.ApiResponse{data=models.CuponPaginadoDoc} "Cupones obtenidos"
// @Failure 400 {object} models.ApiResponse "Parámetros de filtro o paginación inválidos"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /cupones [get]
func (c *CuponController) GetAll() {
	qs := orm.NewOrm().QueryTable(new(models.Cupon))

	if v := strings.TrimSpace(c.GetString("activo")); v != "" {
		activo, err := strconv.ParseBool(v)
		if err != nil {
			httpx.Fail(&c.Controller, http.StatusBadRequest, "El parámetro 'activo' debe ser true o false", err)
			return
		}
		qs = qs.Filter("activo", activo)
	}
	if codigo := strings.TrimSpace(c.GetString("codigo")); codigo != "" {
		qs = qs.Filter("codigo__icontains", codigo)
	}
	if scope := strings.TrimSpace(c.GetString("scope")); scope != "" {
		if !models.CuponScope(scope).IsValid() {
			httpx.Fail(&c.Controller, http.StatusBadRequest, msgScope, nil)
			return
		}
		qs = qs.Filter("scope", scope)
	}
	if v := strings.TrimSpace(c.GetString("fecha_desde")); v != "" {
		fecha, err := models.ParseDateToNoonUTC(v)
		if err != nil {
			httpx.Fail(&c.Controller, http.StatusBadRequest, "El parámetro 'fecha_desde' debe tener formato YYYY-MM-DD", err)
			return
		}
		qs = qs.Filter("fecha_inicio__gte", fecha)
	}
	if v := strings.TrimSpace(c.GetString("fecha_hasta")); v != "" {
		fecha, err := models.ParseDateToNoonUTC(v)
		if err != nil {
			httpx.Fail(&c.Controller, http.StatusBadRequest, "El parámetro 'fecha_hasta' debe tener formato YYYY-MM-DD", err)
			return
		}
		qs = qs.Filter("fecha_fin__lte", fecha)
	}
	limit, offset, ok := c.paginacion()
	if !ok {
		return
	}

	total, err := qs.Count()
	if err != nil {
		logging.LogControllerError(c.Ctx, "cupones.getall.count_error", err, nil)
		httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error al obtener cupones", err)
		return
	}
	var cupones []*models.Cupon
	if _, err = qs.RelatedSel(relaciones...).OrderBy("-pk_id_cupon").Limit(limit).Offset(int64(offset)).All(&cupones); err != nil {
		logging.LogControllerError(c.Ctx, "cupones.getall.query_error", err, nil)
		httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error al obtener cupones", err)
		return
	}
	for _, cp := range cupones {
		limpiar(cp)
	}
	httpx.Send(&c.Controller, http.StatusOK, "Cupones obtenidos exitosamente", pagina(cupones, total, limit, offset))
}

// validar aplica las reglas del cupón y responde 422 si fallan.
func (c *CuponController) validar(op string, cp *models.Cupon) bool {
	if n := len(cp.Codigo); n < 3 || n > 50 {
		c.invalid(op+".codigo", msgCodigo, nil)
		return false
	}
	if err := services.NewCuponService(nil).ValidarReglasNegocioCupon(cp); err != nil {
		c.invalid(op+".validation_error", msgValidacion, err)
		return false
	}
	return true
}

// @Title Post
// @Summary Crear cupón
// @Description Crea un cupón activo. Validación de negocio incumplida (scope, tipo, fechas, código de 3 a 50 caracteres, porcentaje 1-100, combinación de producto/categoría/cliente según el scope) responde 422; un producto, categoría o cliente inexistente responde 400 y un código repetido 409. Fechas YYYY-MM-DD. Devuelve el cupón con sus relaciones como objetos.
// @Tags cupones
// @Accept json
// @Produce json
// @Param body body models.CrearCuponRequest true "Datos del cupón"
// @Success 201 {object} models.ApiResponse{data=models.CuponDoc} "Cupón creado"
// @Failure 400 {object} models.ApiResponse "JSON inválido o referencia inexistente"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 409 {object} models.ApiResponse "Ya existe un cupón con ese código"
// @Failure 422 {object} models.ApiResponse "Error de validación"
// @Failure 500 {object} models.ApiResponse "Error al crear el cupón"
// @Security BearerAuth
// @Router /cupones [post]
func (c *CuponController) Post() {
	var req models.CrearCuponRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		logging.LogControllerError(c.Ctx, "cupones.post.bad_json", err, nil)
		httpx.Fail(&c.Controller, http.StatusBadRequest, msgBadJSON, err)
		return
	}
	if !req.Scope.IsValid() {
		c.invalid("post.invalid_scope", msgScope, nil)
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
	cupon := &models.Cupon{
		Codigo:           strings.TrimSpace(req.Codigo),
		Scope:            req.Scope,
		TipoDescuento:    req.TipoDescuento,
		ValorDescuento:   req.ValorDescuento,
		MaxUsos:          req.MaxUsos,
		LimitePorCliente: req.LimitePorCliente,
		MontoMinimo:      req.MontoMinimo,
		FechaInicio:      fechaInicio,
		FechaFin:         fechaFin,
		Activo:           true,
	}
	if req.PkIdProducto != nil {
		cupon.PkIdProducto = &models.Producto{PK_ID_PRODUCTO: *req.PkIdProducto}
	}
	if req.PkIdCategoria != nil {
		cupon.PkIdCategoria = &models.Categoria{PK_ID_CATEGORIA: *req.PkIdCategoria}
	}
	if req.PkDocumentoCliente != nil {
		cupon.PkDocumentoCliente = &models.Cliente{PK_DOCUMENTO_CLIENTE: *req.PkDocumentoCliente}
	}
	if !c.validar("post", cupon) {
		return
	}

	if _, err := orm.NewOrm().Insert(cupon); err != nil {
		c.writeError("post.insert_error", "Error al crear cupón", err)
		return
	}
	creado, err := porID(cupon.PkIdCupon)
	if err != nil {
		c.readError("post", cupon.PkIdCupon, err)
		return
	}
	httpx.Send(&c.Controller, http.StatusCreated, "Cupón creado exitosamente", creado)
}

// @Title GetById
// @Summary Obtener cupón por ID o código
// @Description `id` puede ser el id numérico del cupón o su código; se busca primero por id (si es numérico) y luego por código.
// @Tags cupones
// @Accept json
// @Produce json
// @Param id query string true "ID numérico o código del cupón"
// @Success 200 {object} models.ApiResponse{data=models.CuponDoc} "Cupón encontrado"
// @Failure 400 {object} models.ApiResponse "Parámetro 'id' ausente"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 404 {object} models.ApiResponse "Cupón no encontrado"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /cupones/search [get]
func (c *CuponController) GetById() {
	idOrCodigo := strings.TrimSpace(c.GetString("id"))
	if idOrCodigo == "" {
		logging.LogControllerError(c.Ctx, "cupones.getbyid.bad_request", nil, map[string]interface{}{"id": idOrCodigo})
		httpx.Fail(&c.Controller, http.StatusBadRequest, "El parámetro 'id' (ID o código) es obligatorio", nil)
		return
	}

	var (
		cupon *models.Cupon
		err   = orm.ErrNoRows
	)
	if id, perr := strconv.ParseInt(idOrCodigo, 10, 64); perr == nil && id > 0 {
		cupon, err = porID(id)
	}
	if errors.Is(err, orm.ErrNoRows) {
		cupon, err = buscar("codigo", idOrCodigo)
	}
	if err != nil {
		c.readError("getbyid", idOrCodigo, err)
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, "Cupón encontrado", cupon)
}

// aplicar copia sobre cp los campos presentes de req (merge). Devuelve false
// (ya respondido) si algún valor es inválido.
func (c *CuponController) aplicar(body []byte, req *models.ActualizarCuponRequest, cp *models.Cupon) bool {
	var err error
	if req.Codigo != nil {
		cp.Codigo = strings.TrimSpace(*req.Codigo)
	}
	if req.Scope != nil {
		if !req.Scope.IsValid() {
			c.invalid("put.invalid_scope", msgScope, nil)
			return false
		}
		cp.Scope = *req.Scope
	}
	if req.TipoDescuento != nil {
		if !req.TipoDescuento.IsValid() {
			c.invalid("put.invalid_tipo_descuento", msgTipo, nil)
			return false
		}
		cp.TipoDescuento = *req.TipoDescuento
	}
	if req.ValorDescuento != nil {
		cp.ValorDescuento = *req.ValorDescuento
	}
	if req.FechaInicio != nil {
		if cp.FechaInicio, err = models.ParseDateToNoonUTC(*req.FechaInicio); err != nil {
			c.invalid("put.invalid_fecha_inicio", msgFechaIni, err)
			return false
		}
	}
	if req.FechaFin != nil {
		if cp.FechaFin, err = models.ParseDateToNoonUTC(*req.FechaFin); err != nil {
			c.invalid("put.invalid_fecha_fin", msgFechaFin, err)
			return false
		}
	}
	if req.MaxUsos != nil {
		cp.MaxUsos = req.MaxUsos
	} else if httpx.IsNull(body, "maxUsos") {
		cp.MaxUsos = nil
	}
	if req.LimitePorCliente != nil {
		cp.LimitePorCliente = req.LimitePorCliente
	} else if httpx.IsNull(body, "limitePorCliente") {
		cp.LimitePorCliente = nil
	}
	if req.MontoMinimo != nil {
		cp.MontoMinimo = req.MontoMinimo
	} else if httpx.IsNull(body, "montoMinimo") {
		cp.MontoMinimo = nil
	}
	if req.PkIdProducto != nil {
		cp.PkIdProducto = &models.Producto{PK_ID_PRODUCTO: *req.PkIdProducto}
	} else if httpx.IsNull(body, "productoId") {
		cp.PkIdProducto = nil
	}
	if req.PkIdCategoria != nil {
		cp.PkIdCategoria = &models.Categoria{PK_ID_CATEGORIA: *req.PkIdCategoria}
	} else if httpx.IsNull(body, "categoriaId") {
		cp.PkIdCategoria = nil
	}
	if req.PkDocumentoCliente != nil {
		cp.PkDocumentoCliente = &models.Cliente{PK_DOCUMENTO_CLIENTE: *req.PkDocumentoCliente}
	} else if httpx.IsNull(body, "documentoCliente") {
		cp.PkDocumentoCliente = nil
	}
	if req.Activo != nil {
		cp.Activo = *req.Activo
	}
	return true
}

// @Title Put
// @Summary Actualizar cupón
// @Description Actualización parcial (merge): los campos ausentes se conservan (cuerpo `models.ActualizarCuponRequest`). `maxUsos`, `limitePorCliente`, `montoMinimo`, `productoId`, `categoriaId` y `documentoCliente` admiten null explícito (se limpian); null en cualquier otro campo responde 400. Al cambiar de `scope` debe ajustarse también la relación correspondiente (p. ej. pasar a GLOBAL exige `productoId`, `categoriaId` y `documentoCliente` en null) o la validación responde 422. `activo` permite reactivar un cupón desactivado. Un cuerpo sin cambios responde 200.
// @Tags cupones
// @Accept json
// @Produce json
// @Param id query int true "ID del cupón (entero positivo)"
// @Param body body models.ActualizarCuponRequest true "Campos a modificar"
// @Success 200 {object} models.ApiResponse{data=models.CuponDoc} "Cupón actualizado"
// @Failure 400 {object} models.ApiResponse "id o JSON inválido, null en campo no anulable o referencia inexistente"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 404 {object} models.ApiResponse "Cupón no encontrado"
// @Failure 409 {object} models.ApiResponse "Ya existe un cupón con ese código"
// @Failure 422 {object} models.ApiResponse "Error de validación"
// @Failure 500 {object} models.ApiResponse "Error al actualizar el cupón"
// @Security BearerAuth
// @Router /cupones [put]
func (c *CuponController) Put() {
	cupon, ok := c.load("put")
	if !ok {
		return
	}
	body := c.Ctx.Input.RequestBody
	var req models.ActualizarCuponRequest
	if err := httpx.DecodeMerge(body, &req, nullableUpdate...); err != nil {
		logging.LogControllerError(c.Ctx, "cupones.put.bad_json", err, map[string]interface{}{"id": cupon.PkIdCupon})
		httpx.Fail(&c.Controller, http.StatusBadRequest, msgBadJSON, err)
		return
	}
	if !c.aplicar(body, &req, cupon) || !c.validar("put", cupon) {
		return
	}

	if _, err := orm.NewOrm().Update(cupon); err != nil {
		c.writeError("put.update_error", "Error al actualizar cupón", err)
		return
	}
	actualizado, err := porID(cupon.PkIdCupon)
	if err != nil {
		c.readError("put", cupon.PkIdCupon, err)
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, "Cupón actualizado exitosamente", actualizado)
}

// @Title Delete
// @Summary Desactivar cupón
// @Description No elimina la fila: desactiva el cupón (`activo = false`; se reactiva con PUT `activo: true`). Si ya estaba desactivado responde 400.
// @Tags cupones
// @Accept json
// @Produce json
// @Param id query int true "ID del cupón (entero positivo)"
// @Success 200 {object} models.ApiResponse "Cupón desactivado"
// @Failure 400 {object} models.ApiResponse "Parámetro 'id' inválido o cupón ya desactivado"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 404 {object} models.ApiResponse "Cupón no encontrado"
// @Failure 500 {object} models.ApiResponse "Error al desactivar el cupón"
// @Security BearerAuth
// @Router /cupones [delete]
func (c *CuponController) Delete() {
	cupon, ok := c.load("delete")
	if !ok {
		return
	}
	if !cupon.Activo {
		httpx.Fail(&c.Controller, http.StatusBadRequest, "El cupón ya está desactivado", nil)
		return
	}
	cupon.Activo = false
	if _, err := orm.NewOrm().Update(cupon, "Activo"); err != nil {
		logging.LogControllerError(c.Ctx, "cupones.delete.update_error", err, map[string]interface{}{"id": cupon.PkIdCupon})
		httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error al desactivar cupón", err)
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, "Cupón desactivado exitosamente", nil)
}

// @Title ValidarCupon
// @Summary Validar cupón
// @Description Evalúa si un cupón es aplicable a un cliente y a unos ítems. Siempre responde 200 con `aplicable` true/false: si es false, `motivo` explica por qué (cupón inexistente, inactivo, fuera de vigencia, usos agotados, cliente no permitido, monto mínimo o sin productos aplicables). Requiere `codigo`, `clienteId` positivo y al menos un ítem con `productoId` > 0, `cantidad` >= 1 y `precio` >= 0.
// @Tags cupones
// @Accept json
// @Produce json
// @Param body body models.ValidarCuponRequest true "Datos para validación"
// @Success 200 {object} models.ApiResponse{data=models.ValidarCuponResponse} "Resultado de la validación"
// @Failure 400 {object} models.ApiResponse "JSON inválido o campos requeridos incorrectos"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 500 {object} models.ApiResponse "Error al validar el cupón"
// @Security BearerAuth
// @Router /cupones/validar [post]
func (c *CuponController) ValidarCupon() {
	var req models.ValidarCuponRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		logging.LogControllerError(c.Ctx, "cupones.validar.bad_json", err, nil)
		httpx.Fail(&c.Controller, http.StatusBadRequest, msgBadJSON, err)
		return
	}
	req.Codigo = strings.TrimSpace(req.Codigo)
	if req.Codigo == "" || req.ClienteId <= 0 || len(req.Items) == 0 {
		httpx.Fail(&c.Controller, http.StatusBadRequest, "codigo, clienteId (positivo) e items (al menos uno) son obligatorios", nil)
		return
	}
	for _, it := range req.Items {
		if it.ProductoId <= 0 || it.Cantidad < 1 || it.Precio < 0 {
			httpx.Fail(&c.Controller, http.StatusBadRequest, "Cada ítem requiere productoId > 0, cantidad >= 1 y precio >= 0", nil)
			return
		}
	}

	resp, err := services.NewCuponServiceFromOrm(orm.NewOrm()).ValidarCupon(c.Ctx.Request.Context(), &req)
	if err != nil {
		logging.LogControllerError(c.Ctx, "cupones.validar.service_error", err, map[string]interface{}{"codigo": req.Codigo})
		httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error al validar cupón", err)
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, "Cupón validado exitosamente", resp)
}

// @Title RedimirCupon
// @Summary Redimir cupón
// @Description Registra la redención de un cupón para un cliente y un pedido existentes; el descuento se calcula con el detalle del pedido (`detalle_pedido`). Un cupón no puede redimirse dos veces en el mismo pedido. Errores: 400 (JSON inválido, `clienteId`/`pedidoId` ausentes o no positivos), 404 (cupón, cliente o pedido inexistente), 409 (cupón agotado, límite por cliente alcanzado o ya redimido en el pedido), 422 (cupón no aplicable: inactivo, fuera de vigencia, cliente no permitido, monto mínimo, sin productos aplicables). La respuesta es el registro de redención con montoDescuento.
// @Tags cupones
// @Accept json
// @Produce json
// @Param codigo path string true "Código del cupón"
// @Param body body models.RedimirCuponRequest true "Datos de redención (clienteId y pedidoId obligatorios)"
// @Success 201 {object} models.ApiResponse{data=models.CuponRedencionDoc} "Cupón redimido"
// @Failure 400 {object} models.ApiResponse "JSON inválido o ids ausentes/no positivos"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 404 {object} models.ApiResponse "Cupón, cliente o pedido no encontrado"
// @Failure 409 {object} models.ApiResponse "Cupón agotado o ya redimido"
// @Failure 422 {object} models.ApiResponse "Cupón no aplicable"
// @Failure 500 {object} models.ApiResponse "Error interno"
// @Security BearerAuth
// @Router /cupones/{codigo}/redimir [post]
func (c *CuponController) RedimirCupon() {
	codigo := strings.TrimSpace(c.Ctx.Input.Param(":codigo"))

	var req models.RedimirCuponRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		logging.LogControllerError(c.Ctx, "cupones.redimir.bad_json", err, map[string]interface{}{"codigo": codigo})
		httpx.Fail(&c.Controller, http.StatusBadRequest, msgBadJSON, err)
		return
	}
	if codigo == "" || req.ClienteId <= 0 || (req.PedidoId != nil && *req.PedidoId <= 0) {
		httpx.Fail(&c.Controller, http.StatusBadRequest, "El código, clienteId y pedidoId deben ser válidos (enteros positivos)", nil)
		return
	}

	redencion, err := services.NewCuponServiceFromOrm(orm.NewOrm()).RedimirCupon(c.Ctx.Request.Context(), codigo, &req)
	if err != nil {
		c.redimirError(codigo, err)
		return
	}
	httpx.Send(&c.Controller, http.StatusCreated, "Cupón redimido exitosamente", redencion)
}

// redimirError traduce los errores tipados del servicio a su código HTTP.
func (c *CuponController) redimirError(codigo string, err error) {
	logging.LogControllerError(c.Ctx, "cupones.redimir.service_error", err, map[string]interface{}{"codigo": codigo})
	switch {
	case errors.Is(err, services.ErrPedidoRequerido):
		httpx.Fail(&c.Controller, http.StatusBadRequest, "Error al redimir cupón", err)
	case errors.Is(err, services.ErrCuponNoEncontrado), errors.Is(err, services.ErrClienteNoEncontrado), errors.Is(err, services.ErrPedidoNoEncontrado):
		httpx.Fail(&c.Controller, http.StatusNotFound, "Error al redimir cupón", err)
	case errors.Is(err, services.ErrCuponConflicto), dberr.IsUnique(err):
		httpx.Fail(&c.Controller, http.StatusConflict, "Error al redimir cupón", err)
	case errors.Is(err, services.ErrCuponNoAplicable):
		httpx.Fail(&c.Controller, http.StatusUnprocessableEntity, "Error al redimir cupón", err)
	default:
		httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error al redimir cupón", err)
	}
}

// @Title ListarRedenciones
// @Summary Listar redenciones de cupones
// @Description Lista paginada (más recientes primero). `cupon_codigo` desconocido devuelve una página vacía. `data.data` es `[]` si no hay resultados. Cada redención trae `cuponId`, `documentoCliente` y `pedidoId` como objetos (sin contraseñas).
// @Tags cupones
// @Accept json
// @Produce json
// @Param cupon_codigo query string false "Código del cupón"
// @Param cupon_id query int false "ID del cupón (entero positivo)"
// @Param cliente_id query int false "Documento del cliente (entero positivo)"
// @Param limit query int false "Límite de resultados (1-100, por defecto 20)"
// @Param offset query int false "Offset para paginación (>= 0, por defecto 0)"
// @Success 200 {object} models.ApiResponse{data=models.CuponRedencionPaginadaDoc} "Redenciones obtenidas"
// @Failure 400 {object} models.ApiResponse "Parámetros de filtro o paginación inválidos"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /cupones/redenciones [get]
func (c *CuponController) ListarRedenciones() {
	o := orm.NewOrm()
	qs := o.QueryTable(new(models.CuponRedencion))

	limit, offset, ok := c.paginacion()
	if !ok {
		return
	}
	if codigo := strings.TrimSpace(c.GetString("cupon_codigo")); codigo != "" {
		cupon := &models.Cupon{}
		if err := o.QueryTable(new(models.Cupon)).Filter("codigo", codigo).One(cupon); err != nil {
			if !errors.Is(err, orm.ErrNoRows) {
				c.readError("redenciones", codigo, err)
				return
			}
			httpx.Send(&c.Controller, http.StatusOK, "Redenciones obtenidas exitosamente", pagina([]*models.CuponRedencion{}, 0, limit, offset))
			return
		}
		qs = qs.Filter("pk_id_cupon", cupon.PkIdCupon)
	}
	if strings.TrimSpace(c.GetString("cupon_id")) != "" {
		id, err := httpx.PositiveInt64Param(&c.Controller, "cupon_id")
		if err != nil {
			httpx.Fail(&c.Controller, http.StatusBadRequest, "El parámetro 'cupon_id' es inválido", err)
			return
		}
		qs = qs.Filter("pk_id_cupon", id)
	}
	if strings.TrimSpace(c.GetString("cliente_id")) != "" {
		id, err := httpx.PositiveInt64Param(&c.Controller, "cliente_id")
		if err != nil {
			httpx.Fail(&c.Controller, http.StatusBadRequest, "El parámetro 'cliente_id' es inválido", err)
			return
		}
		qs = qs.Filter("pk_documento_cliente", id)
	}

	total, err := qs.Count()
	if err != nil {
		logging.LogControllerError(c.Ctx, "cupones.redenciones.count_error", err, nil)
		httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error al obtener redenciones", err)
		return
	}
	var redenciones []*models.CuponRedencion
	if _, err = qs.RelatedSel("PkIdCupon", "PkDocumentoCliente", "PkIdPedido").OrderBy("-created_at").Limit(limit).Offset(int64(offset)).All(&redenciones); err != nil {
		logging.LogControllerError(c.Ctx, "cupones.redenciones.query_error", err, nil)
		httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error al obtener redenciones", err)
		return
	}
	for _, r := range redenciones {
		if r.PkDocumentoCliente != nil {
			r.PkDocumentoCliente.PASSWORD = ""
		}
	}
	httpx.Send(&c.Controller, http.StatusOK, "Redenciones obtenidas exitosamente", pagina(redenciones, total, limit, offset))
}
