package cupon

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
type cuponQuerySeter interface {
	All(interface{}, ...string) (int64, error)
	Filter(string, ...interface{}) cuponQuerySeter
	OrderBy(...string) cuponQuerySeter
	Limit(int) cuponQuerySeter
	Offset(int64) cuponQuerySeter
	Count() (int64, error)
	One(interface{}) error
}

type cuponOrmer interface {
	QueryTable(interface{}) cuponQuerySeter
	Insert(interface{}) (int64, error)
	Read(interface{}, ...string) error
	Update(interface{}, ...string) (int64, error)
	Delete(interface{}, ...string) (int64, error)
}

type cupQSAdapter struct{ qs orm.QuerySeter }

func (a cupQSAdapter) All(res interface{}, cols ...string) (int64, error) {
	return a.qs.All(res, cols...)
}
func (a cupQSAdapter) Filter(expr string, args ...interface{}) cuponQuerySeter {
	return cupQSAdapter{qs: a.qs.Filter(expr, args...)}
}
func (a cupQSAdapter) OrderBy(exprs ...string) cuponQuerySeter {
	return cupQSAdapter{qs: a.qs.OrderBy(exprs...)}
}
func (a cupQSAdapter) Limit(limit int) cuponQuerySeter {
	return cupQSAdapter{qs: a.qs.Limit(limit)}
}
func (a cupQSAdapter) Offset(offset int64) cuponQuerySeter {
	return cupQSAdapter{qs: a.qs.Offset(offset)}
}
func (a cupQSAdapter) Count() (int64, error) {
	return a.qs.Count()
}
func (a cupQSAdapter) One(container interface{}) error {
	return a.qs.One(container)
}

type cupOrmAdapter struct{ o orm.Ormer }

func (a cupOrmAdapter) QueryTable(i interface{}) cuponQuerySeter {
	return cupQSAdapter{qs: a.o.QueryTable(i)}
}
func (a cupOrmAdapter) Insert(v interface{}) (int64, error)      { return a.o.Insert(v) }
func (a cupOrmAdapter) Read(v interface{}, cols ...string) error { return a.o.Read(v, cols...) }
func (a cupOrmAdapter) Update(v interface{}, cols ...string) (int64, error) {
	return a.o.Update(v, cols...)
}
func (a cupOrmAdapter) Delete(v interface{}, cols ...string) (int64, error) {
	return a.o.Delete(v, cols...)
}

var cupOrmNew = func() cuponOrmer { return cupOrmAdapter{o: orm.NewOrm()} }

// Variable mockeable para tests
var newCuponService = func(o orm.Ormer) *services.CuponService {
	return services.NewCuponService(o)
}

type CuponController struct {
	web.Controller
}

// @Title GetAll
// @Summary Obtener todos los cupones
// @Tags cupones
// @Accept json
// @Produce json
// @Param activo query bool false "Filtrar por estado activo"
// @Param codigo query string false "Filtrar por código"
// @Param scope query string false "Filtrar por scope (GLOBAL, PRODUCTO, CATEGORIA, CLIENTE)"
// @Param fecha_desde query string false "Fecha desde (YYYY-MM-DD)"
// @Param fecha_hasta query string false "Fecha hasta (YYYY-MM-DD)"
// @Param limit query int false "Límite de resultados (default: 20)"
// @Param offset query int false "Offset para paginación (default: 0)"
// @Success 200 {object} models.ApiResponse{data=models.PaginatedResponse}
// @Failure 400 {object} models.ApiResponse
// @Failure 500 {object} models.ApiResponse
// @Router /cupones [get]
func (c *CuponController) GetAll() {
	o := cupOrmNew()
	qs := o.QueryTable("cupon")

	// Aplicar filtros
	if activo := c.GetString("activo"); activo != "" {
		if activoBool, err := strconv.ParseBool(activo); err == nil {
			qs = qs.Filter("activo", activoBool)
		}
	}

	if codigo := c.GetString("codigo"); codigo != "" {
		qs = qs.Filter("codigo__icontains", codigo)
	}

	if scope := c.GetString("scope"); scope != "" {
		qs = qs.Filter("scope", scope)
	}

	if fechaDesde := c.GetString("fecha_desde"); fechaDesde != "" {
		if fecha, err := time.Parse("2006-01-02", fechaDesde); err == nil {
			qs = qs.Filter("fecha_inicio__gte", fecha)
		}
	}

	if fechaHasta := c.GetString("fecha_hasta"); fechaHasta != "" {
		if fecha, err := time.Parse("2006-01-02", fechaHasta); err == nil {
			qs = qs.Filter("fecha_fin__lte", fecha)
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
		logging.LogControllerError(c.Ctx, "cupones.getall.count_error", err, nil)
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener cupones",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	// Obtener datos
	var cupones []*models.Cupon
	_, err = qs.OrderBy("-pk_id_cupon").Limit(limit).Offset(int64(offset)).All(&cupones)
	if err != nil {
		logging.LogControllerError(c.Ctx, "cupones.getall.query_error", err, nil)
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener cupones",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))
	page := (offset / limit) + 1

	response := models.PaginatedResponse{
		Data:       cupones,
		Total:      total,
		Page:       page,
		PageSize:   limit,
		TotalPages: totalPages,
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Cupones obtenidos exitosamente",
		Data:    response,
	}
	_ = c.ServeJSON()
}

// @Title Post
// @Summary Crear cupón
// @Tags cupones
// @Accept json
// @Produce json
// @Param body body models.CrearCuponRequest true "Datos del cupón"
// @Success 201 {object} models.ApiResponse{data=models.Cupon}
// @Failure 400 {object} models.ApiResponse
// @Failure 422 {object} models.ApiResponse
// @Failure 500 {object} models.ApiResponse
// @Router /cupones [post]
func (c *CuponController) Post() {
	var req models.CrearCuponRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		logging.LogControllerError(c.Ctx, "cupones.post.bad_json", err, nil)
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "JSON inválido",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	// Validar enums
	if !req.Scope.IsValid() {
		logging.LogControllerError(c.Ctx, "cupones.post.invalid_scope", nil, map[string]interface{}{"scope": req.Scope})
		c.Ctx.Output.SetStatus(http.StatusUnprocessableEntity)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusUnprocessableEntity,
			Message: "Scope no válido - debe ser GLOBAL, PRODUCTO, CATEGORIA o CLIENTE",
		}
		_ = c.ServeJSON()
		return
	}

	if !req.TipoDescuento.IsValid() {
		logging.LogControllerError(c.Ctx, "cupones.post.invalid_tipo_descuento", nil, map[string]interface{}{"tipoDescuento": req.TipoDescuento})
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
		logging.LogControllerError(c.Ctx, "cupones.post.invalid_fecha_inicio", err, map[string]interface{}{"fechaInicio": req.FechaInicio})
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
		logging.LogControllerError(c.Ctx, "cupones.post.invalid_fecha_fin", err, map[string]interface{}{"fechaFin": req.FechaFin})
		c.Ctx.Output.SetStatus(http.StatusUnprocessableEntity)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusUnprocessableEntity,
			Message: "Fecha de fin inválida - debe tener formato YYYY-MM-DD",
		}
		_ = c.ServeJSON()
		return
	}

	// Crear modelo
	cupon := &models.Cupon{
		Codigo:           req.Codigo,
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

	// Asignar relaciones según el scope
	if req.PkIdProducto != nil {
		cupon.PkIdProducto = &models.Producto{PK_ID_PRODUCTO: *req.PkIdProducto}
	}
	if req.PkIdCategoria != nil {
		cupon.PkIdCategoria = &models.Categoria{PK_ID_CATEGORIA: *req.PkIdCategoria}
	}
	if req.PkDocumentoCliente != nil {
		cupon.PkDocumentoCliente = &models.Cliente{PK_DOCUMENTO_CLIENTE: *req.PkDocumentoCliente}
	}

	// Validar reglas de negocio
	o := cupOrmNew()
	// ValidarReglasNegocioCupon no usa el ORM, así que podemos pasar nil
	cuponService := newCuponService(nil)

	if err := cuponService.ValidarReglasNegocioCupon(cupon); err != nil {
		logging.LogControllerError(c.Ctx, "cupones.post.validation_error", err, map[string]interface{}{"codigo": req.Codigo})
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
	_, err = o.Insert(cupon)
	if err != nil {
		logging.LogControllerError(c.Ctx, "cupones.post.insert_error", err, map[string]interface{}{"codigo": req.Codigo})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al crear cupón",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusCreated,
		Message: "Cupón creado exitosamente",
		Data:    cupon,
	}
	_ = c.ServeJSON()
}

// @Title GetById
// @Summary Obtener cupón por ID
// @Tags cupones
// @Accept json
// @Produce json
// @Param id query string true "ID o código del cupón"
// @Success 200 {object} models.ApiResponse{data=models.Cupon}
// @Failure 404 {object} models.ApiResponse
// @Router /cupones/search [get]
func (c *CuponController) GetById() {
	idOrCodigo := c.GetString("id")
	if idOrCodigo == "" {
		logging.LogControllerError(c.Ctx, "cupones.getbyid.bad_request", nil, map[string]interface{}{"id": idOrCodigo})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "ID o código es requerido",
		}
		_ = c.ServeJSON()
		return
	}

	o := cupOrmNew()
	cupon := &models.Cupon{}

	// Intentar buscar por ID primero
	if id, err := strconv.ParseInt(idOrCodigo, 10, 64); err == nil {
		cupon.PkIdCupon = id
		err = o.Read(cupon)
		if err != nil {
			if err == orm.ErrNoRows {
				// No encontrado por ID, resetear para buscar por código
				cupon.PkIdCupon = 0
			} else {
				logging.LogControllerError(c.Ctx, "cupones.getbyid.read_error", err, map[string]interface{}{"id": id})
				c.Ctx.Output.SetStatus(http.StatusInternalServerError)
				c.Data["json"] = models.ApiResponse{
					Code:    http.StatusInternalServerError,
					Message: "Error interno del servidor",
					Cause:   err.Error(),
				}
				_ = c.ServeJSON()
				return
			}
		}
	}

	// Si no se encontró por ID, buscar por código
	if cupon.PkIdCupon == 0 {
		err := o.QueryTable("cupon").Filter("codigo", idOrCodigo).One(cupon)
		if err != nil {
			if err == orm.ErrNoRows {
				c.Ctx.Output.SetStatus(http.StatusNotFound)
				c.Data["json"] = models.ApiResponse{
					Code:    http.StatusNotFound,
					Message: "Cupón no encontrado",
				}
				_ = c.ServeJSON()
				return
			}
			logging.LogControllerError(c.Ctx, "cupones.getbyid.query_error", err, map[string]interface{}{"codigo": idOrCodigo})
			c.Ctx.Output.SetStatus(http.StatusInternalServerError)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusInternalServerError,
				Message: "Error interno del servidor",
				Cause:   err.Error(),
			}
			_ = c.ServeJSON()
			return
		}
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Cupón encontrado",
		Data:    cupon,
	}
	_ = c.ServeJSON()
}

// @Title Put
// @Summary Actualizar cupón
// @Tags cupones
// @Accept json
// @Produce json
// @Param id query int true "ID del cupón"
// @Param body body models.CrearCuponRequest true "Datos actualizados del cupón"
// @Success 200 {object} models.ApiResponse{data=models.Cupon}
// @Failure 400 {object} models.ApiResponse
// @Failure 404 {object} models.ApiResponse
// @Failure 422 {object} models.ApiResponse
// @Router /cupones [put]
func (c *CuponController) Put() {
	id, err := c.GetInt64("id")
	if err != nil || id == 0 {
		logging.LogControllerError(c.Ctx, "cupones.put.bad_request", err, map[string]interface{}{"id": c.GetString("id")})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "ID inválido o ausente",
		}
		_ = c.ServeJSON()
		return
	}

	var req models.CrearCuponRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		logging.LogControllerError(c.Ctx, "cupones.put.bad_json", err, map[string]interface{}{"id": id})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "JSON inválido",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	o := cupOrmNew()

	// Verificar que el cupón existe
	cupon := &models.Cupon{PkIdCupon: id}
	err = o.Read(cupon)
	if err != nil {
		if err == orm.ErrNoRows {
			c.Ctx.Output.SetStatus(http.StatusNotFound)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusNotFound,
				Message: "Cupón no encontrado",
			}
			_ = c.ServeJSON()
			return
		}
		logging.LogControllerError(c.Ctx, "cupones.put.read_error", err, map[string]interface{}{"id": id})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error interno del servidor",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	// Validar enums
	if !req.Scope.IsValid() {
		logging.LogControllerError(c.Ctx, "cupones.put.invalid_scope", nil, map[string]interface{}{"scope": req.Scope, "id": id})
		c.Ctx.Output.SetStatus(http.StatusUnprocessableEntity)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusUnprocessableEntity,
			Message: "Scope no válido - debe ser GLOBAL, PRODUCTO, CATEGORIA o CLIENTE",
		}
		_ = c.ServeJSON()
		return
	}

	if !req.TipoDescuento.IsValid() {
		logging.LogControllerError(c.Ctx, "cupones.put.invalid_tipo_descuento", nil, map[string]interface{}{"tipoDescuento": req.TipoDescuento, "id": id})
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
		logging.LogControllerError(c.Ctx, "cupones.put.invalid_fecha_inicio", err, map[string]interface{}{"fechaInicio": req.FechaInicio, "id": id})
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
		logging.LogControllerError(c.Ctx, "cupones.put.invalid_fecha_fin", err, map[string]interface{}{"fechaFin": req.FechaFin, "id": id})
		c.Ctx.Output.SetStatus(http.StatusUnprocessableEntity)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusUnprocessableEntity,
			Message: "Fecha de fin inválida - debe tener formato YYYY-MM-DD",
		}
		_ = c.ServeJSON()
		return
	}

	// Actualizar campos
	cupon.Codigo = req.Codigo
	cupon.Scope = req.Scope
	cupon.TipoDescuento = req.TipoDescuento
	cupon.ValorDescuento = req.ValorDescuento
	cupon.MaxUsos = req.MaxUsos
	cupon.LimitePorCliente = req.LimitePorCliente
	cupon.MontoMinimo = req.MontoMinimo
	cupon.FechaInicio = fechaInicio
	cupon.FechaFin = fechaFin

	// Limpiar relaciones anteriores
	cupon.PkIdProducto = nil
	cupon.PkIdCategoria = nil
	cupon.PkDocumentoCliente = nil

	// Asignar nuevas relaciones según el scope
	if req.PkIdProducto != nil {
		cupon.PkIdProducto = &models.Producto{PK_ID_PRODUCTO: *req.PkIdProducto}
	}
	if req.PkIdCategoria != nil {
		cupon.PkIdCategoria = &models.Categoria{PK_ID_CATEGORIA: *req.PkIdCategoria}
	}
	if req.PkDocumentoCliente != nil {
		cupon.PkDocumentoCliente = &models.Cliente{PK_DOCUMENTO_CLIENTE: *req.PkDocumentoCliente}
	}

	// Validar reglas de negocio
	cuponService := newCuponService(nil)
	if err := cuponService.ValidarReglasNegocioCupon(cupon); err != nil {
		logging.LogControllerError(c.Ctx, "cupones.put.validation_error", err, map[string]interface{}{"codigo": req.Codigo, "id": id})
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
	_, err = o.Update(cupon)
	if err != nil {
		logging.LogControllerError(c.Ctx, "cupones.put.update_error", err, map[string]interface{}{"id": id})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al actualizar cupón",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Cupón actualizado exitosamente",
		Data:    cupon,
	}
	_ = c.ServeJSON()
}

// @Title Delete
// @Summary Desactivar cupón
// @Tags cupones
// @Accept json
// @Produce json
// @Param id query int true "ID del cupón"
// @Success 200 {object} models.ApiResponse
// @Failure 400 {object} models.ApiResponse
// @Failure 404 {object} models.ApiResponse
// @Router /cupones [delete]
func (c *CuponController) Delete() {
	id, err := c.GetInt64("id")
	if err != nil || id == 0 {
		logging.LogControllerError(c.Ctx, "cupones.delete.bad_request", err, map[string]interface{}{"id": c.GetString("id")})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "ID inválido o ausente",
		}
		_ = c.ServeJSON()
		return
	}

	o := cupOrmNew()
	cupon := &models.Cupon{PkIdCupon: id}
	err = o.Read(cupon)
	if err != nil {
		if err == orm.ErrNoRows {
			c.Ctx.Output.SetStatus(http.StatusNotFound)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusNotFound,
				Message: "Cupón no encontrado",
			}
			_ = c.ServeJSON()
			return
		}
		logging.LogControllerError(c.Ctx, "cupones.delete.read_error", err, map[string]interface{}{"id": id})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error interno del servidor",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	if !cupon.Activo {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "El cupón ya está desactivado",
		}
		_ = c.ServeJSON()
		return
	}

	// Desactivar cupón (borrado lógico)
	cupon.Activo = false
	_, err = o.Update(cupon, "Activo")
	if err != nil {
		logging.LogControllerError(c.Ctx, "cupones.delete.update_error", err, map[string]interface{}{"id": id})
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al desactivar cupón",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Cupón desactivado exitosamente",
	}
	_ = c.ServeJSON()
}

// Métodos adicionales específicos de cupones

// @Title ValidarCupon
// @Summary Validar cupón
// @Tags cupones
// @Accept json
// @Produce json
// @Param body body models.ValidarCuponRequest true "Datos para validación"
// @Success 200 {object} models.ApiResponse{data=models.ValidarCuponResponse}
// @Failure 400 {object} models.ApiResponse
// @Failure 422 {object} models.ApiResponse
// @Router /cupones/validar [post]
func (c *CuponController) ValidarCupon() {
	var req models.ValidarCuponRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		logging.LogControllerError(c.Ctx, "cupones.validar.bad_json", err, nil)
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "JSON inválido",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	cuponService := newCuponService(orm.NewOrm())
	response, err := cuponService.ValidarCupon(c.Ctx.Request.Context(), &req)
	if err != nil {
		logging.LogControllerError(c.Ctx, "cupones.validar.service_error", err, map[string]interface{}{"codigo": "validacion"})
		c.Ctx.Output.SetStatus(http.StatusUnprocessableEntity)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusUnprocessableEntity,
			Message: "Error al validar cupón",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Cupón validado exitosamente",
		Data:    response,
	}
	_ = c.ServeJSON()
}

// @Title RedimirCupon
// @Summary Redimir cupón
// @Tags cupones
// @Accept json
// @Produce json
// @Param codigo path string true "Código del cupón"
// @Param body body models.RedimirCuponRequest true "Datos de redención"
// @Success 201 {object} models.ApiResponse{data=models.CuponRedencion}
// @Failure 400 {object} models.ApiResponse
// @Failure 404 {object} models.ApiResponse
// @Failure 409 {object} models.ApiResponse
// @Failure 422 {object} models.ApiResponse
// @Router /cupones/{codigo}/redimir [post]
func (c *CuponController) RedimirCupon() {
	codigo := c.Ctx.Input.Param(":codigo")

	var req models.RedimirCuponRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		logging.LogControllerError(c.Ctx, "cupones.redimir.bad_json", err, map[string]interface{}{"codigo": codigo})
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "JSON inválido",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	cuponService := newCuponService(nil)
	redencion, err := cuponService.RedimirCupon(c.Ctx.Request.Context(), codigo, &req)
	if err != nil {
		logging.LogControllerError(c.Ctx, "cupones.redimir.service_error", err, map[string]interface{}{"codigo": codigo})

		// Determinar el tipo de error
		errorMsg := err.Error()
		switch errorMsg {
		case "cupón no encontrado":
			c.Ctx.Output.SetStatus(http.StatusOK)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusNotFound,
				Message: "Cupón no encontrado",
			}
		case "cupón no aplicable", "Cliente ha alcanzado el límite de usos para este cupón", "Cupón ha alcanzado el límite máximo de usos":
			c.Ctx.Output.SetStatus(http.StatusConflict)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusConflict,
				Message: errorMsg,
			}
		default:
			c.Ctx.Output.SetStatus(http.StatusUnprocessableEntity)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusUnprocessableEntity,
				Message: "Error al redimir cupón",
				Cause:   errorMsg,
			}
		}
		_ = c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(http.StatusCreated)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusCreated,
		Message: "Cupón redimido exitosamente",
		Data:    redencion,
	}
	_ = c.ServeJSON()
}

// @Title ListarRedenciones
// @Summary Listar redenciones de cupones
// @Tags cupones
// @Accept json
// @Produce json
// @Param cupon_codigo query string false "Código del cupón"
// @Param cupon_id query int false "ID del cupón"
// @Param cliente_id query int false "ID del cliente"
// @Param limit query int false "Límite de resultados (default: 20)"
// @Param offset query int false "Offset para paginación (default: 0)"
// @Success 200 {object} models.ApiResponse{data=models.PaginatedResponse}
// @Failure 400 {object} models.ApiResponse
// @Failure 500 {object} models.ApiResponse
// @Router /cupones/redenciones [get]
func (c *CuponController) ListarRedenciones() {
	o := cupOrmNew()
	qs := o.QueryTable("cupon_redencion")

	// Aplicar filtros
	if cuponCodigo := c.GetString("cupon_codigo"); cuponCodigo != "" {
		// Buscar el cupón por código primero
		cupon := &models.Cupon{}
		err := o.QueryTable("cupon").Filter("codigo", cuponCodigo).One(cupon)
		if err == nil {
			qs = qs.Filter("pk_id_cupon", cupon.PkIdCupon)
		} else {
			// Si no se encuentra el cupón, no hay redenciones
			response := models.PaginatedResponse{
				Data:       []*models.CuponRedencion{},
				Total:      0,
				Page:       1,
				PageSize:   20,
				TotalPages: 0,
			}
			c.Ctx.Output.SetStatus(http.StatusOK)
			c.Data["json"] = models.ApiResponse{
				Code:    http.StatusOK,
				Message: "Redenciones obtenidas exitosamente",
				Data:    response,
			}
			_ = c.ServeJSON()
			return
		}
	}

	if cuponId := c.GetString("cupon_id"); cuponId != "" {
		if id, err := strconv.ParseInt(cuponId, 10, 64); err == nil {
			qs = qs.Filter("pk_id_cupon", id)
		}
	}

	if clienteId := c.GetString("cliente_id"); clienteId != "" {
		if id, err := strconv.ParseInt(clienteId, 10, 64); err == nil {
			qs = qs.Filter("pk_documento_cliente", id)
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
		logging.LogControllerError(c.Ctx, "cupones.redenciones.count_error", err, nil)
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener redenciones",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	// Obtener datos
	var redenciones []*models.CuponRedencion
	_, err = qs.OrderBy("-created_at").Limit(limit).Offset(int64(offset)).All(&redenciones)
	if err != nil {
		logging.LogControllerError(c.Ctx, "cupones.redenciones.query_error", err, nil)
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = models.ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Error al obtener redenciones",
			Cause:   err.Error(),
		}
		_ = c.ServeJSON()
		return
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))
	page := (offset / limit) + 1

	response := models.PaginatedResponse{
		Data:       redenciones,
		Total:      total,
		Page:       page,
		PageSize:   limit,
		TotalPages: totalPages,
	}

	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = models.ApiResponse{
		Code:    http.StatusOK,
		Message: "Redenciones obtenidas exitosamente",
		Data:    response,
	}
	_ = c.ServeJSON()
}
